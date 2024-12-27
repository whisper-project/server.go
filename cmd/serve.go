/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package cmd

import (
	"fmt"
	"log"

	"github.com/whisper-project/server.golang/common/middleware"

	"github.com/whisper-project/server.golang/common/platform"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/whisper-project/server.golang/api/console"
	"github.com/whisper-project/server.golang/api/saywhat"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the whisper server",
	Long: `Runs the whisper server until it's killed by signal.
Runs in the development environment by default.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(0)
		env, _ := cmd.Flags().GetString("env")
		err := platform.PushConfig(env)
		if err != nil {
			panic(fmt.Sprintf("Can't load configuration: %v", err))
		}
		defer platform.PopConfig()
		serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Args = cobra.NoArgs
	serveCmd.Flags().StringP("env", "e", "development", "The environment to run in")
}

func serve() {
	if platform.GetConfig().Name == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := middleware.CreateCoreEngine()
	_ = r.SetTrustedProxies(nil)
	r.Static("/say-what", "./saywhat.js/dist")
	sayWhat := r.Group("/api/say-what/v1")
	saywhat.AddRoutes(sayWhat)
	consoleClient := r.Group("/api/console/v0")
	console.AddRoutes(consoleClient)
	if err := r.Run("localhost:8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
