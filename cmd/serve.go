/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"clickonetwo.io/whisper/api/saywhat"
	"clickonetwo.io/whisper/server/middleware"
	"clickonetwo.io/whisper/server/storage"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the whisperapp server",
	Long: `Runs the whisperapp server process.
The whisperapp will never terminate unless it is interrupted/terminated.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func serve() {
	err := storage.PushConfig(".env")
	if err != nil {
		panic(fmt.Sprintf("Can't load configuration: %v", err))
	}
	defer storage.PopConfig()
	r := middleware.CreateCoreEngine()
	r.Static("/say-what", "./saywhat.js/dist")
	sayWhat := r.Group("/api/say-what/v1")
	saywhat.AddRoutes(sayWhat)
	err = r.Run("localhost:5000")
	if err != nil {
		fmt.Printf("Server exited with error: %v", err)
	}
}
