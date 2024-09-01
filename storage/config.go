package storage

import (
	"fmt"
	"github.com/dotenv-org/godotenvvault"
	"os"
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
		ApnsCredSecret:   "-----BEGIN PRIVATE KEY----- MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQg5TL3GlhuHCFZe0L/ g+rt2ibfrgaGaiYl1/N2FAms0yehRANCAAT6nm9Bs5+HXOI2DRm9h1LtQxofxa1e lMN+WP8KFt9KQ/yKYohq4ZLtvdxfjoPobxPNm+VGkycP8zQMK3RAwJSu -----END PRIVATE KEY-----",
		ApnsCredId:       "89AB98CD89",
		ApnsTeamId:       "8CD8989AB9",
		DbUrl:            "redis://",
		DbKeyPrefix:      "t:",
	}
	loadedConfig = testConfig
	configStack  = []Config{}
)

func GetConfig() *Config {
	return &loadedConfig
}

func PushConfig(from ...string) error {
	err := godotenvvault.Load(from...)
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

func PopConfig() error {
	if len(configStack) == 0 {
		return fmt.Errorf("no configs to pop")
	}
	loadedConfig = configStack[len(configStack)-1]
	configStack = configStack[:len(configStack)-1]
	return nil
}
