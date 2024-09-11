package main

import (
	"fmt"

	"clickonetwo.io/whisper/server/middleware"
	"clickonetwo.io/whisper/server/storage"
)

func main() {
	err := storage.PushConfig(".env")
	if err != nil {
		panic(fmt.Sprintf("Can't load configuration: %v", err))
	}
	defer storage.PopConfig()
	r := middleware.CreateCoreEngine()
	err = r.Run("localhost:8080")
	if err != nil {
		fmt.Printf("Server exited with error: %v", err)
	}
}
