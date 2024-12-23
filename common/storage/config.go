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

type Environment struct {
	Name             string
	AblyPublishKey   string
	AblySubscribeKey string
	ApnsUrl          string
	ApnsCredSecret   string
	ApnsCredId       string
	ApnsTeamId       string
	DbUrl            string
	DbKeyPrefix      string
}

//goland:noinspection SpellCheckingInspection
var (
	ciConfig = Environment{
		Name:             "CI",
		AblyPublishKey:   "xVLyHw.DGYdkQ:FtPUNIourpYSoZAIbeon0p_rJGtb5vO1j2OIzP3GMX8",
		AblySubscribeKey: "xVLyHw.DGYdkQ:FtPUNIourpYSoZAIbeon0p_rJGtb5vO1j2OIzP3GMX8",
		ApnsUrl:          "http://localhost:2197",
		ApnsCredSecret:   "-----BEGIN PRIVATE KEY-----\nMIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgGSZi+0fnzC8bbBbI\nD5wyNIgqnl7dFLN+FlUD/mOAG+ShRANCAASZU2wXczRjmlkcHJp4yHTl3KlAXoB8\nozM8I6bJBZPUGlTdIpvV2u2mLhKBZNZIUDaqdHKkfukSn+hgdZspMtaA\n-----END PRIVATE KEY-----",
		ApnsCredId:       "89AB98CD89",
		ApnsTeamId:       "8CD8989AB9",
		DbUrl:            "redis://",
		DbKeyPrefix:      "c:",
	}
	loadedConfig = ciConfig
	configStack  []Environment
)

func GetConfig() Environment {
	return loadedConfig
}

func PushConfig(name string) error {
	if name == "" {
		return pushEnvConfig("")
	}
	if strings.HasPrefix(name, "c") {
		return pushCiConfig()
	}
	if strings.HasPrefix(name, "d") {
		return pushEnvConfig(".env")
	}
	if strings.HasPrefix(name, "s") {
		return pushEnvConfig(".env.staging")
	}
	if strings.HasPrefix(name, "p") {
		return pushEnvConfig(".env.production")
	}
	if strings.HasPrefix(name, "t") {
		return pushEnvConfig(".env.testing")
	}
	return fmt.Errorf("unknown environment: %s", name)
}

func PushAlteredConfig(env Environment) {
	configStack = append(configStack, loadedConfig)
	loadedConfig = env
}

func pushCiConfig() error {
	configStack = append(configStack, loadedConfig)
	loadedConfig = ciConfig
	return nil
}

func pushEnvConfig(filename string) error {
	var d string
	var err error
	if filename == "" {
		if d, err = FindEnvFile(".env.vault", true); err == nil {
			if d == "" {
				err = godotenvvault.Overload()
			} else {
				var c string
				if c, err = os.Getwd(); err == nil {
					if err = os.Chdir(d); err == nil {
						err = godotenvvault.Overload()
						// if we fail to change back to the prior working directory, so be it.
						_ = os.Chdir(c)
					}
				}
			}
		}
	} else {
		if d, err = FindEnvFile(filename, false); err == nil {
			err = godotenvvault.Overload(d + filename)
		}
	}
	if err != nil {
		return fmt.Errorf("error loading .env vars: %v", err)
	}
	configStack = append(configStack, loadedConfig)
	loadedConfig = Environment{
		Name:             os.Getenv("ENVIRONMENT_NAME"),
		AblyPublishKey:   os.Getenv("ABLY_PUBLISH_KEY"),
		AblySubscribeKey: os.Getenv("ABLY_SUBSCRIBE_KEY"),
		ApnsUrl:          os.Getenv("APNS_SERVER"),
		ApnsCredSecret:   os.Getenv("APNS_CRED_SECRET_PKCS8"),
		ApnsCredId:       os.Getenv("APNS_CRED_ID"),
		ApnsTeamId:       os.Getenv("APNS_TEAM_ID"),
		DbUrl:            os.Getenv("REDIS_URL"),
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

func FindEnvFile(name string, fallback bool) (string, error) {
	for i := range 5 {
		d := ""
		for range i {
			d += "../"
		}
		if _, err := os.Stat(d + name); err == nil {
			return d, nil
		}
		if fallback {
			if _, err := os.Stat(d + ".env"); err == nil {
				return d, nil
			}
		}
	}
	return "", fmt.Errorf("no file %q found in path", name)
}
