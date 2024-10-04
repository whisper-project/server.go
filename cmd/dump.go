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

	"clickonetwo.io/whisper/server/internal/profile"
	"clickonetwo.io/whisper/server/internal/storage"
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
	dps := make(map[string]profile.Data, len(ids))
	for _, id := range ids {
		p := profile.Data{Id: id}
		if err := storage.LoadFields(context.Background(), &p); err != nil {
			panic(err)
		}
		dps[p.Id] = p
	}

	// dump them
	if err := json.NewEncoder(os.Stdout).Encode(dps); err != nil {
		panic(err)
	}
	fmt.Println()
}
