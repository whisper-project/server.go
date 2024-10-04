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

	"github.com/spf13/cobra"

	"clickonetwo.io/whisper/server/internal/profile"
	"clickonetwo.io/whisper/server/internal/storage"
)

// transferCmd represents the transfer command
var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer profiles between environments",
	Long: `By default, this command transfers profiles from production to test,
but you can use flags to control the source and the target environments.`,
	Run: func(cmd *cobra.Command, args []string) {
		from, err := cmd.Flags().GetString("from")
		if err != nil {
			panic(err)
		}
		to, err := cmd.Flags().GetString("to")
		if err != nil {
			panic(err)
		}
		transfer(from, to)
	},
}

func init() {
	rootCmd.AddCommand(transferCmd)

	transferCmd.Flags().StringP("from", "f", "production", "source environment")
	transferCmd.Flags().StringP("to", "t", "test", "target environment")
}

func transfer(from, to string) {
	var ps []profile.Data
	p := &profile.Data{}
	count := 0
	pushProfile := func() {
		if p.SettingsProfile.Settings["elevenlabs_api_key_preference"] != "" {
			ps = append(ps, *p)
		}
		count++
		if (count % 10) == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "\nProcessed %d profiles...", count)
		}
	}

	if err := storage.PushConfig(from); err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintf(os.Stderr, "Starting to transfer profiles from %s to %s...", from, to)
	if err := storage.MapFields(context.Background(), pushProfile, p); err != nil {
		panic(err)
	}
	if (count % 10) != 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Processed %d profiles.", count)
	}
	_, _ = fmt.Fprintf(os.Stderr, "\nFound %d profiles to transfer...\n", len(ps))
	storage.PopConfig()

	if err := storage.PushConfig(to); err != nil {
		panic(err)
	}
	var testList profile.List = "transferred-profile-list"
	var ids []string
	for _, tp := range ps {
		if err := storage.SaveFields(context.Background(), &tp); err != nil {
			panic(err)
		}
		ids = append(ids, tp.Id)
	}
	if err := storage.PushRange(context.Background(), testList, false, ids...); err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintf(os.Stderr, "Transferred profile IDs are in %q with prefix %q", testList.StorageId(), testList.StoragePrefix())
}
