/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// whisperCmd represents the whisper command
var whisperCmd = &cobra.Command{
	Use:   "whisper [flags] [conversation-name]",
	Short: "Start a whisper session",
	Long:  `Starts whispering on the named conversation (or your default, if the name is omitted.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting to whisper...NOT")
	},
}

func init() {
	rootCmd.AddCommand(whisperCmd)
	whisperCmd.Args = cobra.MaximumNArgs(1)
}
