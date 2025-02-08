/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/whisper-project/server.golang/api/console"
	"github.com/whisper-project/server.golang/api/saywhat"
	"github.com/whisper-project/server.golang/lifecycle"
	"github.com/whisper-project/server.golang/platform"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the whisper server",
	Long:  `Runs the whisper server until it's killed by signal.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(0)
		env, _ := cmd.Flags().GetString("env")
		address, _ := cmd.Flags().GetString("address")
		port, _ := cmd.Flags().GetString("port")
		err := platform.PushConfig(env)
		if err != nil {
			panic(fmt.Sprintf("Can't load configuration: %v", err))
		}
		defer platform.PopConfig()
		serve(address, port)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Args = cobra.NoArgs
	serveCmd.Flags().StringP("env", "e", "development", "The environment to run in")
	serveCmd.Flags().StringP("address", "a", "127.0.0.1", "The IP address to listen on")
	serveCmd.Flags().StringP("port", "p", "8080", "The port to listen on")
}

func serve(address, port string) {
	r, err := lifecycle.CreateEngine()
	if err != nil {
		panic(err)
	}
	r.Static("/say-what", "./saywhat.js/dist")
	sayWhat := r.Group("/api/say-what/v1")
	saywhat.AddRoutes(sayWhat)
	consoleClient := r.Group("/api/console/v0")
	console.AddRoutes(consoleClient)
	lifecycle.Startup(r, fmt.Sprintf("%s:%s", address, port))
}
