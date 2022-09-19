package config

import (
	"os"
	"strings"
	"sync"
)

const (
	EnvKeyStashGraphQLUrl        = "STASH_GRAPHQL_URL"
	EnvKeyStashApiKey            = "STASH_API_KEY"
	EnvKeyFavoriteTag            = "FAVORITE_TAG"
	EnvKeyFilters                = "FILTERS"
	EnvKeyHeresphereSyncMarkers  = "HERESPHERE_SYNC_MARKERS"
	EnvKeyHeresphereQuickMarkers = "HERESPHERE_QUICK_MARKERS"
	EnvKeyLogLevel               = "LOG_LEVEL"
	EnvKeyDisableRedact          = "DISABLE_REDACT"
)

type Application struct {
	StashGraphQLUrl        string
	StashApiKey            string
	FavoriteTag            string
	Filters                string
	HeresphereSyncMarkers  bool
	HeresphereQuickMarkers bool
	LogLevel               string
	IsRedactDisabled       bool
}

var cfg Application

var once sync.Once

func Get() Application {
	once.Do(func() {
		cfg = Application{
			StashGraphQLUrl:        getEnvOrDefault(EnvKeyStashGraphQLUrl, "http://localhost:9999/graphql"),
			StashApiKey:            getEnvOrDefault(EnvKeyStashApiKey, ""),
			FavoriteTag:            getEnvOrDefault(EnvKeyFavoriteTag, "FAVORITE"),
			Filters:                getEnvOrDefault(EnvKeyFilters, ""),
			HeresphereSyncMarkers:  getEnvOrDefault(EnvKeyHeresphereSyncMarkers, "false") == "true",
			HeresphereQuickMarkers: getEnvOrDefault(EnvKeyHeresphereQuickMarkers, "false") == "true",
			LogLevel:               strings.ToLower(getEnvOrDefault(EnvKeyLogLevel, "info")),
			IsRedactDisabled:       getEnvOrDefault(EnvKeyDisableRedact, "false") == "true",
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
