/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

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

func profileStats(path string) {
	// profile classification logic
	var all []*profile.UserProfile
	byId := make(map[string][]*profile.UserProfile)
	byName := make(map[string][]*profile.UserProfile)
	byClientCount := make(map[int][]*profile.UserProfile)
	byElevenLabsKey := make(map[string][]*profile.UserProfile)
	processed := 0
	p := profile.UserProfile{}
	classify := func() {
		pc := p
		all = append(all, &pc)
		byId[p.Id] = append(byId[p.Id], &pc)
		byName[p.Name] = append(byName[p.Name], &pc)
		clients, err := storage.FetchMembers(context.Background(), profile.Clients(p.Id))
		if err != nil {
			panic(err)
		}
		byClientCount[len(clients)] = append(byClientCount[len(clients)], &pc)
		if key := p.SettingsProfile.Settings["elevenlabs_api_key_preference"]; key != "" {
			byElevenLabsKey[key] = append(byElevenLabsKey[key], &pc)
		}
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

	// print the stats
	fmt.Printf("There are %d profiles, of which %d are shared.\n",
		len(all), sharedLen(all))
	fmt.Printf("There are %d anonymous profiles, %d of which are shared.\n",
		len(byName[""]), sharedLen(byName[""]))
	fmt.Printf("There are %d profiles with 0 clients, %d of which are shared, and %d of which are anonymous.\n",
		len(byClientCount[0]), sharedLen(byClientCount[0]), anonLen(byClientCount[0]))
	fmt.Printf("There are %d different ElevenLabs API keys, in %d different profiles.\n",
		len(byElevenLabsKey), mappedLen(byElevenLabsKey))

	// dump the content
	if path == "" {
		return
	}
	content := map[string]interface{}{
		"all":             any(all),
		"byUser":          any(byName),
		"byClientCount":   any(byClientCount),
		"byElevenLabsKey": any(byElevenLabsKey),
	}
	path = path + "/stats-profiles.json"
	stream, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer stream.Close()
	encoder := json.NewEncoder(stream)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(content); err != nil {
		panic(err)
	}
	fmt.Printf("Profiles dumped to %q\n", path)
}

func sharedLen(ps []*profile.UserProfile) int {
	count := 0
	for _, p := range ps {
		if p.Password != "" {
			count++
		}
	}
	return count
}

func anonLen(ps []*profile.UserProfile) int {
	count := 0
	for _, p := range ps {
		if p.Name == "" {
			count++
		}
	}
	return count
}

func mappedLen[T int | string](mps map[T][]*profile.UserProfile) int {
	count := 0
	for _, v := range mps {
		count += len(v)
	}
	return count
}
