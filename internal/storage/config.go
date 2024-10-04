/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package storage

import (
	"fmt"
	"os"
	"strings"

	"github.com/dotenv-org/godotenvvault"
)

type Config struct {
	AblyPublishKey   string
	AblySubscribeKey string
	ApnsUrl          string
	ApnsCredSecret   string
	ApnsCredId       string
	ApnsTeamId       string
	DbUrl            string
	DbKeyPrefix      string
}

var (
	testConfig = Config{
		AblyPublishKey:   "xVLyHw.DGYdkQ:FtPUNIourpYSoZAIbeon0p_rJGtb5vO1j2OIzP3GMX8",
		AblySubscribeKey: "xVLyHw.DGYdkQ:FtPUNIourpYSoZAIbeon0p_rJGtb5vO1j2OIzP3GMX8",
		ApnsUrl:          "http://localhost:2197",
		ApnsCredSecret:   "-----BEGIN PRIVATE KEY-----\nMIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgGSZi+0fnzC8bbBbI\nD5wyNIgqnl7dFLN+FlUD/mOAG+ShRANCAASZU2wXczRjmlkcHJp4yHTl3KlAXoB8\nozM8I6bJBZPUGlTdIpvV2u2mLhKBZNZIUDaqdHKkfukSn+hgdZspMtaA\n-----END PRIVATE KEY-----",
		ApnsCredId:       "89AB98CD89",
		ApnsTeamId:       "8CD8989AB9",
		DbUrl:            "redis://",
		DbKeyPrefix:      "t:",
	}
	loadedConfig = testConfig
	configStack  []Config
)

func GetConfig() *Config {
	return &loadedConfig
}

func PushConfig(env string) error {
	if env == "" {
		return pushEnvConfig("")
	}
	if strings.HasPrefix(env, "t") {
		return pushTestConfig()
	}
	if strings.HasPrefix(env, "d") {
		return pushEnvConfig(".env")
	}
	if strings.HasPrefix(env, "s") {
		return pushEnvConfig(".env.staging")
	}
	if strings.HasPrefix(env, "p") {
		return pushEnvConfig(".env.production")
	}
	return fmt.Errorf("unknown environment: %s", env)
}

func pushTestConfig() error {
	configStack = append(configStack, loadedConfig)
	loadedConfig = testConfig
	return nil
}

func pushEnvConfig(filename string) error {
	var err error
	if filename == "" {
		err = godotenvvault.Load()
	} else {
		for _, f := range []string{filename, "../" + filename, "../../" + filename, "../../../" + filename} {
			if _, err = os.Stat(f); err == nil {
				err = godotenvvault.Load(f)
				break
			}
		}
	}
	if err != nil {
		return fmt.Errorf("error loading .env vars: %v", err)
	}
	configStack = append(configStack, loadedConfig)
	loadedConfig = Config{
		AblyPublishKey:   os.Getenv("ABLY_PUBLISH_KEY"),
		AblySubscribeKey: os.Getenv("ABLY_SUBSCRIBE_KEY"),
		ApnsUrl:          os.Getenv("APNS_SERVER"),
		ApnsCredSecret:   os.Getenv("APNS_CRED_SECRET_PKCS8"),
		ApnsCredId:       os.Getenv("APNS_CRED_ID"),
		ApnsTeamId:       os.Getenv("APNS_TEAM_ID"),
		DbUrl:            os.Getenv("REDISCLOUD_URL"),
		DbKeyPrefix:      os.Getenv("DB_KEY_PREFIX"),
	}
	return nil
}

func PopConfig() {
	if len(configStack) == 0 {
		return
	}
	loadedConfig = configStack[len(configStack)-1]
	configStack = configStack[:len(configStack)-1]
	return
}
