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

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump profiles",
	Long: `This command dumps the profiles last transferred to an environment.
By default, it dumps from the test environment, but you can
control the source environment with a flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		from, err := cmd.Flags().GetString("from")
		if err != nil {
			panic(err)
		}
		dump(from)
	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)

	dumpCmd.Flags().StringP("from", "f", "test", "source environment")
}

func dump(from string) {
	// get to the environment
	if err := storage.PushConfig(from); err != nil {
		panic(err)
	}

	// load the transferred ids
	var testList profile.List = "transferred-profile-list"
	ids, err := storage.FetchRange(context.Background(), testList, 0, -1)
	if err != nil {
		panic(err)
	}

	// assemble dumped form of the transferred profiles
	dps := make(map[string][]profile.UserProfile)
	for _, id := range ids {
		p := profile.UserProfile{Id: id}
		if err := storage.LoadFields(context.Background(), &p); err != nil {
			panic(err)
		}
		dps[p.Name] = append(dps[p.Name], p)
	}

	// dump them
	_, _ = fmt.Fprintf(os.Stderr, "Found %d profiles to dump, for %d users.\n", len(ids), len(dps))
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(dps); err != nil {
		panic(err)
	}
	fmt.Println()
}
