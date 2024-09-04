package main

import (
	"log"
	"time"

	ginzap "github.com/gin-contrib/zap"

	"clickonetwo.io/whisper/server/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	err := storage.PushConfig(".env")
	if err != nil {
		log.Fatal("Can't load configuration: ", err)
	}
	defer storage.PopConfig()

	// set up main router with zap logging
	r := gin.New()
	logger, _ := zap.NewDevelopment()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))
}
