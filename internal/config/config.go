package config

import (
	"os"
	"strings"
	"sync"
)

const (
	envKeyStashGraphQLUrl = "STASH_GRAPHQL_URL"
	envKeyStashApiKey     = "STASH_API_KEY"
	envKeyFavoriteTag     = "FAVORITE_TAG"
	envKeyFilters         = "FILTERS"
	envKeyLogLevel        = "LOG_LEVEL"
	envKeyDisableRedact   = "DISABLE_REDACT"
	envKeyForceHTTPS      = "FORCE_HTTPS"
)

var envKeyEnableGlanceMarkers = []string{"ENABLE_GLANCE_MARKERS", "HERESPHERE_QUICK_MARKERS"}
var envKeyAllowSyncMarkers = []string{"ALLOW_SYNC_MARKERS", "HERESPHERE_SYNC_MARKERS"}

type Application struct {
	StashGraphQLUrl        string
	StashApiKey            string
	FavoriteTag            string
	Filters                string
	IsGlanceMarkersEnabled bool
	IsSyncMarkersAllowed   bool
	LogLevel               string
	IsRedactDisabled       bool
	ForceHTTPS             bool
}

var cfg Application

var once sync.Once

func Get() Application {
	once.Do(func() {
		cfg = Application{
			StashGraphQLUrl:        getEnvOrDefault(envKeyStashGraphQLUrl, "http://localhost:9999/graphql"),
			StashApiKey:            getEnvOrDefault(envKeyStashApiKey, ""),
			FavoriteTag:            getEnvOrDefault(envKeyFavoriteTag, "FAVORITE"),
			Filters:                getEnvOrDefault(envKeyFilters, ""),
			IsGlanceMarkersEnabled: findEnvOrDefault(envKeyEnableGlanceMarkers, "false") == "true",
			IsSyncMarkersAllowed:   findEnvOrDefault(envKeyAllowSyncMarkers, "false") == "true",
			LogLevel:               strings.ToLower(getEnvOrDefault(envKeyLogLevel, "info")),
			IsRedactDisabled:       getEnvOrDefault(envKeyDisableRedact, "false") == "true",
			ForceHTTPS:             getEnvOrDefault(envKeyForceHTTPS, "false") == "true",
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

func findEnvOrDefault(keys []string, defaultValue string) string {
	for _, key := range keys {
		v, ok := os.LookupEnv(key)
		if ok {
			return v
		}
	}
	return defaultValue
}

func (a Application) Redacted() Application {
	a.StashGraphQLUrl = Redacted(a.StashGraphQLUrl)
	a.StashApiKey = Redacted(a.StashApiKey)
	return a
}
