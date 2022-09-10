package config

import (
	"os"
	"strings"
	"sync"
)

const (
	envKeyStashGraphQLUrl = "STASH_GRAPHQL_URL"
	envKeyStashApiKey     = "STASH_API_KEY"
	envKeyLogLevel        = "LOG_LEVEL"
	envKeyDisableRedact   = "DISABLE_REDACT"
)

type Application struct {
	StashGraphQLUrl  string
	StashApiKey      string
	LogLevel         string
	IsRedactDisabled bool
}

var cfg Application

var once sync.Once

func Get() Application {
	once.Do(func() {
		cfg = Application{
			StashGraphQLUrl:  getEnvOrDefault(envKeyStashGraphQLUrl, "http://localhost:9999/graphql"),
			StashApiKey:      getEnvOrDefault(envKeyStashApiKey, ""),
			LogLevel:         strings.ToLower(getEnvOrDefault(envKeyLogLevel, "info")),
			IsRedactDisabled: getEnvOrDefault(envKeyDisableRedact, "false") == "true",
		}
	})
	return cfg
}

func getEnvOrDefault(key string, defaultValue string) string {
	if val, ok := os.LookupEnv(key); !ok {
		return defaultValue
	} else {
		return val
	}
}

func (a Application) Redacted() Application {
	a.StashGraphQLUrl = Redacted(a.StashGraphQLUrl)
	a.StashApiKey = Redacted(a.StashApiKey)
	return a
}
