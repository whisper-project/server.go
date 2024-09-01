package storage

import (
	"github.com/redis/go-redis/v9"
	"log"
)

var client *redis.Client
var keyPrefix string

func GetDb() (*redis.Client, string) {
	if client != nil {
		return client, keyPrefix
	}
	config := GetConfig()
	opts, err := redis.ParseURL(config.DbUrl)
	if err != nil {
		log.Fatal("invalid Redis url:", err)
	}
	return redis.NewClient(opts), config.DbKeyPrefix
}
