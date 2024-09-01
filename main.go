package main

import (
	"clickonetwo/whisper/server/storage"
	"log"
)

func main() {
	err := storage.LoadConfig()
	if err != nil {
		log.Fatal("Can't load configuration: ", err)
	}
}
