package config

import (
	"os"
	"strings"
	"sync"
)

const (
	EnvKeyStashGraphQLUrl      = "STASH_GRAPHQL_URL"
	EnvKeyStashApiKey          = "STASH_API_KEY"
	EnvKeyFavoriteTag          = "FAVORITE_TAG"
	EnvKeyFrontPageFiltersOnly = "FRONT_PAGE_FILTERS_ONLY"
	EnvKeyLogLevel             = "LOG_LEVEL"
	EnvKeyDisableRedact        = "DISABLE_REDACT"
)

type Application struct {
	StashGraphQLUrl      string
	StashApiKey          string
	FavoriteTag          string
	FrontPageFiltersOnly bool
	LogLevel             string
	IsRedactDisabled     bool
}

var cfg Application

var once sync.Once

func Get() Application {
	once.Do(func() {
		cfg = Application{
			StashGraphQLUrl:      getEnvOrDefault(EnvKeyStashGraphQLUrl, "http://localhost:9999/graphql"),
			StashApiKey:          getEnvOrDefault(EnvKeyStashApiKey, ""),
			FavoriteTag:          getEnvOrDefault(EnvKeyFavoriteTag, "FAVORITE"),
			FrontPageFiltersOnly: getEnvOrDefault(EnvKeyFrontPageFiltersOnly, "false") == "true",
			LogLevel:             strings.ToLower(getEnvOrDefault(EnvKeyLogLevel, "info")),
			IsRedactDisabled:     getEnvOrDefault(EnvKeyDisableRedact, "false") == "true",
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
