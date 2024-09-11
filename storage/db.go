package storage

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	clientUrl string
	client    *redis.Client
	keyPrefix string
)

func GetDb() (*redis.Client, string) {
	config := GetConfig()
	if client != nil && clientUrl == config.DbUrl {
		return client, keyPrefix
	}
	opts, err := redis.ParseURL(config.DbUrl)
	if err != nil {
		panic(fmt.Sprintf("invalid Redis url: %v", err))
	}
	clientUrl = config.DbUrl
	client = redis.NewClient(opts)
	keyPrefix = config.DbKeyPrefix
	return client, keyPrefix
}
