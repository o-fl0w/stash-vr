package config

import (
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	envKeyStashGraphQLUrl  = "STASH_GRAPHQL_URL"
	envKeyStashApiKey      = "STASH_API_KEY"
	envKeyFavoriteTag      = "FAVORITE_TAG"
	envKeyFilters          = "FILTERS"
	envKeyLogLevel         = "LOG_LEVEL"
	envKeyDisableRedact    = "DISABLE_REDACT"
	envKeyForceHTTPS       = "FORCE_HTTPS"
	envKeyDisableHeatmap   = "DISABLE_HEATMAP"
	envKeyHeatmapHeightPx  = "HEATMAP_HEIGHT_PX"
	envKeyAllowSyncMarkers = "ALLOW_SYNC_MARKERS"
	envKeyDisablePlayCount = "DISABLE_PLAY_COUNT"
)

var deprecatedEnvKeys = []string{"ENABLE_GLANCE_MARKERS", "HERESPHERE_QUICK_MARKERS", "HERESPHERE_SYNC_MARKERS", "ENABLE_HEATMAP_DISPLAY"}

type Application struct {
	StashGraphQLUrl      string
	StashApiKey          string
	FavoriteTag          string
	Filters              string
	IsSyncMarkersAllowed bool
	LogLevel             string
	IsRedactDisabled     bool
	ForceHTTPS           bool
	IsHeatmapDisabled    bool
	HeatmapHeightPx      int
	IsPlayCountDisabled  bool
}

var cfg Application

var once sync.Once

func Get() Application {
	once.Do(func() {
		logDeprecatedKeysInUse()
		cfg = Application{
			StashGraphQLUrl:      getEnvOrDefaultStr(envKeyStashGraphQLUrl, "http://localhost:9999/graphql"),
			StashApiKey:          getEnvOrDefaultStr(envKeyStashApiKey, ""),
			FavoriteTag:          getEnvOrDefaultStr(envKeyFavoriteTag, "FAVORITE"),
			Filters:              getEnvOrDefaultStr(envKeyFilters, ""),
			IsSyncMarkersAllowed: getEnvOrDefaultBool(envKeyAllowSyncMarkers, false),
			LogLevel:             strings.ToLower(getEnvOrDefaultStr(envKeyLogLevel, "info")),
			IsRedactDisabled:     getEnvOrDefaultBool(envKeyDisableRedact, false),
			ForceHTTPS:           getEnvOrDefaultBool(envKeyForceHTTPS, false),
			IsHeatmapDisabled:    getEnvOrDefaultBool(envKeyDisableHeatmap, false),
			HeatmapHeightPx:      getEnvOrDefaultInt(envKeyHeatmapHeightPx, 0),
			IsPlayCountDisabled:  getEnvOrDefaultBool(envKeyDisablePlayCount, false),
		}
	})
	return cfg
}

func logDeprecatedKeysInUse() {
	for _, key := range deprecatedEnvKeys {
		val, found := os.LookupEnv(key)
		if found {
			log.Warn().Str("key", key).Str("value", val).Msg("Deprecated/removed option found. Ignoring.")
		}
	}
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	s, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal().Err(err).Str("key", key).Str("value", s).Msg("Invalid value in environment arguments. Must be an integer.")
		return 0
	}
	return val
}

func getEnvOrDefaultStr(key string, defaultValue string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return val
}

func getEnvOrDefaultBool(key string, defaultValue bool) bool {
	s, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	val, err := strconv.ParseBool(s)
	if err != nil {
		log.Fatal().Err(err).Str("key", key).Str("value", s).Msg("Invalid value in environment arguments. Must be a valid boolean (1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False)")
		return false
	}
	return val
}

func findEnvOrDefault(keys []string, defaultValue string) string {
	for i, key := range keys {
		v, ok := os.LookupEnv(key)
		if ok {
			if i > 0 {
				log.Warn().Str("deprecated", key).Str("replace with", keys[0]).Msg("Deprecated env. var. found")
			}
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
