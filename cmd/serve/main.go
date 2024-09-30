package main

import (
	"fmt"

	"clickonetwo.io/whisper/server/api/saywhat"
	"clickonetwo.io/whisper/server/internal/middleware"
	"clickonetwo.io/whisper/server/internal/storage"
)

func main() {
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
