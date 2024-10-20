/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package cmd

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"clickonetwo.io/whisper/internal/client"
	"clickonetwo.io/whisper/internal/profile"
	"clickonetwo.io/whisper/internal/storage"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Compute statistics about database content",
	Long: `Compute statistics about the profiles, clients, and conversations
to be found in the database in the specified environment.
Optionally dumps the database content to a JSON file.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var from, path string
		var dump bool
		from, err = cmd.Flags().GetString("from")
		if err != nil {
			panic(err)
		}
		dump, err = cmd.Flags().GetBool("dump")
		if err != nil {
			panic(err)
		}
		if dump {
			path, err = cmd.Flags().GetString("path")
			if err != nil {
				panic(err)
			}
		}
		stats(from, path)
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)

	statsCmd.Flags().StringP("from", "f", "production", "source environment")
	statsCmd.Flags().BoolP("dump", "d", false, "dump the database content")
	statsCmd.Flags().StringP("path", "p", "/tmp", "directory for dumped content")
}

func stats(from string, path string) {
	if err := storage.PushConfig(from); err != nil {
		panic(err)
	}
	defer storage.PopConfig()

	profileStats(path)
}

type UserSummary struct {
	Username      string              `json:"username"`
	LastUpdated   time.Time           `json:"last_updated"`
	LastUsed      time.Time           `json:"last_used"`
	ProfileId     string              `json:"profile_id"`
	IsShared      bool                `json:"is_shared"`
	Clients       []client.Data       `json:"clients"`
	ElevenLabsKey string              `json:"eleven_labs_key"`
	UserProfile   profile.UserProfile `json:"user_profile"`
}

type UserStats struct {
	name         string
	profileCount int
	sharedCount  int
	keyCount     int
	clientCount  int
	unusedCount  int
}

func profileStats(path string) {
	// profile classification logic
	byName := make(map[string][]UserSummary)
	processed := 0
	p := profile.UserProfile{}
	classify := func() {
		ids, err := storage.FetchMembers(context.Background(), profile.Clients(p.Id))
		if err != nil {
			panic(err)
		}
		cs := make([]client.Data, 0, len(ids))
		lastUsed := time.UnixMilli(0)
		for _, clientId := range ids {
			c := client.Data{Id: clientId}
			if err := storage.LoadFields(context.Background(), &c); err != nil {
				panic(err)
			}
			cs = append(cs, c)
			last := time.UnixMilli(c.LastLaunch)
			if last.After(lastUsed) {
				lastUsed = last
			}
		}
		s := UserSummary{
			Username:      p.Name,
			LastUpdated:   time.Unix(max(p.WhisperProfile.Timestamp, p.ListenProfile.Timestamp), 0),
			LastUsed:      lastUsed,
			ProfileId:     p.Id,
			IsShared:      p.Password != "",
			Clients:       cs,
			ElevenLabsKey: p.SettingsProfile.Settings["elevenlabs_api_key_preference"],
			UserProfile:   p,
		}
		byName[p.Name] = append(byName[p.Name], s)
		processed++
		if (processed % 10) == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "\nProcessed %d profiles...", processed)
		}
	}

	// do the classification
	if err := storage.MapFields(context.Background(), classify, &p); err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintf(os.Stderr, "\nProcessed %d profiles.\n", processed)

	// order the profiles by most recent use descending
	for _, ss := range byName {
		slices.SortFunc(ss, func(a, b UserSummary) int {
			switch diff := a.LastUsed.Sub(b.LastUsed); {
			case diff > 0:
				return -1
			case diff < 0:
				return 1
			default:
				return 0
			}
		})
	}

	// compute the user stats
	userStats := make([]UserStats, 0, len(byName)-1)
	for n, ss := range byName {
		if n == "" {
			continue
		}
		user := UserStats{
			name:         n,
			profileCount: len(ss),
			sharedCount:  sharedLen(ss),
			keyCount:     keyLen(ss),
			clientCount:  clientLen(ss),
			unusedCount:  unusedLen(ss, 30),
		}
		userStats = append(userStats, user)
	}
	slices.SortFunc(userStats, func(a, b UserStats) int {
		if a.profileCount == b.profileCount {
			return strings.Compare(strings.ToLower(a.name), strings.ToLower(b.name))
		}
		return b.profileCount - a.profileCount
	})

	// print the stats
	fmt.Printf("There are %d anonymous profiles (shared: %d, unused 30 days: %d).\n",
		len(byName[""]), sharedLen(byName[""]), unusedLen(byName[""], 30))
	fmt.Printf("These are %d named profiles:\n", processed-len(byName[""]))
	for _, s := range userStats {
		fmt.Printf("    %s: %d profiles, %d shared, %d with keys, %d on devices, %d unused 30 days\n",
			s.name, s.profileCount, s.sharedCount, s.keyCount, s.clientCount, s.unusedCount)
	}

	// dump the content
	if path == "" {
		return
	}
	DumpObjects(path+"/profiles.json", byName, "Profiles")
}

func sharedLen(ss []UserSummary) int {
	count := 0
	for _, s := range ss {
		if s.IsShared {
			count++
		}
	}
	return count
}

func unusedLen(ss []UserSummary, days int) int {
	count := 0
	for _, s := range ss {
		if time.Now().Sub(s.LastUpdated) > time.Duration(days*24)*time.Hour {
			count++
		}
	}
	return count
}

func keyLen(ss []UserSummary) int {
	count := 0
	for _, s := range ss {
		if s.ElevenLabsKey != "" {
			count++
		}
	}
	return count
}

func clientLen(ss []UserSummary) int {
	count := 0
	for _, s := range ss {
		if len(s.Clients) > 0 {
			count++
		}
	}
	return count
}
