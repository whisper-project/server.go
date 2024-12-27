/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package cmd

import (
	"context"
	"fmt"
	"maps"
	"os"
	"slices"
	"time"

	storage2 "github.com/whisper-project/server.golang/common/platform"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/spf13/cobra"

	"github.com/whisper-project/server.golang/legacy/client"
	"github.com/whisper-project/server.golang/legacy/profile"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Compute statistics about database content",
	Long: `Compute statistics about the profiles, clients, and conversations
to be found in the database in the specified environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var from string
		from, err = cmd.Flags().GetString("from")
		if err != nil {
			panic(err)
		}
		stats(from)
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)

	statsCmd.Flags().StringP("from", "f", "production", "source environment")
	statsCmd.Flags().BoolP("dump", "d", false, "dump the database content")
	statsCmd.Flags().StringP("path", "p", "/tmp", "directory for dumped content")
}

var millis30days int64 = 30 * 24 * 60 * 60 * 1000

func stats(from string) {
	if err := storage2.PushConfig(from); err != nil {
		panic(err)
	}
	defer storage2.PopConfig()

	cs := analyzeClients()
	ps := analyzeProfiles(cs)
	printClientStats(cs)
	printProfileStats(ps)
}

type clientStatistics struct {
	clients          map[string]client.Data
	lastLaunched     map[string]int64 // profile ID to last launch time
	recentlyLaunched int64
	builds           map[string]int64 // build info to count of clients
}

func newClientStatistics() clientStatistics {
	return clientStatistics{
		clients:      make(map[string]client.Data),
		lastLaunched: make(map[string]int64),
		builds:       make(map[string]int64),
	}
}

func analyzeClients() clientStatistics {
	cs := newClientStatistics()
	c := client.Data{}
	processed := 0
	now := time.Now().UnixMilli()
	classify := func() {
		cs.clients[c.Id] = c
		if processed++; processed%10 == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "\nProcessed %d clients...", processed)
		}
		if now-c.LastLaunch <= millis30days {
			cs.recentlyLaunched++
		}
		cs.lastLaunched[c.ProfileId] = max(cs.lastLaunched[c.ProfileId], c.LastLaunch)
		cs.builds[c.AppInfo] += 1
	}

	// collect the client data
	_, _ = fmt.Fprintf(os.Stderr, "Starting to process clients...")
	if err := storage2.MapFields(context.Background(), classify, &c); err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintf(os.Stderr, "\nProcessed %d clients.\n", processed)

	return cs
}

type profileStatistics struct {
	anonymous         map[string]profile.UserProfile
	abandoned         map[string]profile.UserProfile
	inactive          map[string]profile.UserProfile
	active            map[string]profile.UserProfile
	priorWhisperers   []string
	currentWhisperers []string
	priorListeners    []string
	currentListeners  []string
	webListeners      []string
}

func newProfileStatistics() profileStatistics {
	return profileStatistics{
		anonymous: make(map[string]profile.UserProfile),
		abandoned: make(map[string]profile.UserProfile),
		inactive:  make(map[string]profile.UserProfile),
		active:    make(map[string]profile.UserProfile),
	}
}

func analyzeProfiles(cs clientStatistics) profileStatistics {
	// profile classification logic: anonymous and abandoned profiles don't get statistics
	ps := newProfileStatistics()
	allIds := mapset.NewSet[string]()
	whisperers := mapset.NewSet[string]() // profile ids that people have listened to
	listeners := mapset.NewSet[string]()  // profiles ids that have listened to people
	processed := 0
	p := profile.UserProfile{}
	classify := func() {
		allIds.Add(p.Id)
		if processed++; processed%10 == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "\nProcessed %d profiles...", processed)
		}
		if p.Name == "" {
			ps.anonymous[p.Id] = p
			return
		}
		if p.LastUsed == 0 {
			p.LastUsed = max(p.WhisperProfile.Timestamp*1000, p.ListenProfile.Timestamp*1000, cs.lastLaunched[p.Id])
		}
		if cs.lastLaunched[p.Id] == 0 {
			if p.Password == "" {
				ps.abandoned[p.Id] = p
				return
			}
			if time.Now().UnixMilli()-p.LastUsed > millis30days {
				// shared, but not used in 30 days
				ps.abandoned[p.Id] = p
				return
			}
			ps.inactive[p.Id] = p
		} else {
			ps.active[p.Id] = p
		}
		listeners.Append(allowedListeners(p.WhisperProfile)...)
		whisperers.Append(pastWhisperers(p.ListenProfile)...)
	}

	// collect the profile data
	_, _ = fmt.Fprintf(os.Stderr, "Starting to process profiles...")
	if err := storage2.MapFields(context.Background(), classify, &p); err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintf(os.Stderr, "\nProcessed %d profiles.\n", processed)

	// prune whisperer and listener profiles
	for _, id := range whisperers.ToSlice() {
		if _, isActive := ps.active[id]; isActive {
			ps.currentWhisperers = append(ps.currentWhisperers, id)
		} else {
			ps.priorListeners = append(ps.priorListeners, id)
		}
	}
	for _, id := range listeners.ToSlice() {
		if !allIds.Contains(id) {
			ps.webListeners = append(ps.webListeners, id)
		} else if _, isActive := ps.active[id]; isActive {
			ps.currentListeners = append(ps.currentListeners, id)
		} else {
			ps.priorListeners = append(ps.priorListeners, id)
		}
	}
	return ps
}

func printProfileStats(ps profileStatistics) {
	// print the stats
	fmt.Printf("There are %d anonymous profiles.\n", len(ps.anonymous))
	fmt.Printf("There are %d abandoned profiles.\n", len(ps.abandoned))
	fmt.Printf("There are %d inactive, recently-used, shared profiles.\n", len(ps.inactive))
	fmt.Printf("There are %d active profiles:\n", len(ps.active))
	fmt.Printf("    %d current whisperers and %d prior whisperers.\n",
		len(ps.currentWhisperers), len(ps.priorWhisperers))
	fmt.Printf("    %d current listeners, %d prior listeners, and %d web listeners.\n",
		len(ps.currentListeners), len(ps.priorListeners), len(ps.webListeners))
}

func printClientStats(cs clientStatistics) {
	fmt.Printf("There are %d clients.\n", len(cs.clients))
	builds := slices.Collect(maps.Keys(cs.builds))
	slices.Sort(builds)
	fmt.Printf("Client build distribution:\n")
	for _, b := range builds {
		fmt.Printf("    %s: %d\n", b, cs.builds[b])
	}
}

func allowedListeners(w profile.WhisperProfile) []string {
	var ls []string
	for _, v := range w.Table {
		for id := range v.Allowed {
			ls = append(ls, id)
		}
	}
	return ls
}

func pastWhisperers(w profile.ListenProfile) []string {
	var ws []string
	for _, l := range w.Table {
		ws = append(ws, l.Owner)
	}
	return ws
}
