package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"strings"
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
	envKeyStimhubUrl       = "STIMHUB_URL"
)

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
	StimhubUrl           string
}

var cfg Application

func init() {
	pflag.String(envKeyStashGraphQLUrl, "http://localhost:9999/graphql", "Url to Stash graphql")
	_ = viper.BindPFlag(envKeyStashGraphQLUrl, pflag.Lookup(envKeyStashGraphQLUrl))

	pflag.String(envKeyStashApiKey, "", "Stash API key")
	_ = viper.BindPFlag(envKeyStashApiKey, pflag.Lookup(envKeyStashApiKey))

	pflag.String(envKeyFavoriteTag, "FAVORITE", "Name of tag in Stash to hold scenes marked as favorites")
	_ = viper.BindPFlag(envKeyFavoriteTag, pflag.Lookup(envKeyFavoriteTag))

	pflag.String(envKeyFilters, "", "Narrow the selection of filters to show. Either 'frontpage' or a comma seperated list of filter ids")
	_ = viper.BindPFlag(envKeyFilters, pflag.Lookup(envKeyFilters))

	pflag.Bool(envKeyAllowSyncMarkers, false, "Enable sync of Marker from HereSphere")
	_ = viper.BindPFlag(envKeyAllowSyncMarkers, pflag.Lookup(envKeyAllowSyncMarkers))

	pflag.String(envKeyLogLevel, "info", "Set log level - trace, debug, warn, info or error")
	_ = viper.BindPFlag(envKeyLogLevel, pflag.Lookup(envKeyLogLevel))

	pflag.Bool(envKeyDisableRedact, false, "Disable redacting sensitive information from logs")
	_ = viper.BindPFlag(envKeyDisableRedact, pflag.Lookup(envKeyDisableRedact))

	pflag.Bool(envKeyForceHTTPS, false, "Force Stash-VR to use HTTPS")
	_ = viper.BindPFlag(envKeyForceHTTPS, pflag.Lookup(envKeyForceHTTPS))

	pflag.Bool(envKeyDisableHeatmap, false, "Disable display of funscript heatmaps")
	_ = viper.BindPFlag(envKeyDisableHeatmap, pflag.Lookup(envKeyDisableHeatmap))

	pflag.Int(envKeyHeatmapHeightPx, 0, "Height of heatmaps")
	_ = viper.BindPFlag(envKeyHeatmapHeightPx, pflag.Lookup(envKeyHeatmapHeightPx))

	pflag.Bool(envKeyDisablePlayCount, false, "Disable incrementing Stash play count for scenes")
	_ = viper.BindPFlag(envKeyDisablePlayCount, pflag.Lookup(envKeyDisablePlayCount))

	pflag.String(envKeyStimhubUrl, "", "")
	_ = viper.BindPFlag(envKeyStimhubUrl, pflag.Lookup(envKeyStimhubUrl))

	pflag.Parse()

	viper.AutomaticEnv()

	cfg.StashGraphQLUrl = viper.GetString(envKeyStashGraphQLUrl)
	cfg.StashApiKey = viper.GetString(envKeyStashApiKey)
	cfg.FavoriteTag = viper.GetString(envKeyFavoriteTag)
	cfg.Filters = viper.GetString(envKeyFilters)
	cfg.IsSyncMarkersAllowed = viper.GetBool(envKeyAllowSyncMarkers)
	cfg.LogLevel = strings.ToLower(viper.GetString(envKeyLogLevel))
	cfg.IsRedactDisabled = viper.GetBool(envKeyDisableRedact)
	cfg.ForceHTTPS = viper.GetBool(envKeyForceHTTPS)
	cfg.IsHeatmapDisabled = viper.GetBool(envKeyDisableHeatmap)
	cfg.HeatmapHeightPx = viper.GetInt(envKeyHeatmapHeightPx)
	cfg.IsPlayCountDisabled = viper.GetBool(envKeyDisablePlayCount)
	cfg.StimhubUrl = viper.GetString(envKeyStimhubUrl)

}

func Get() Application {
	return cfg
}

func (a Application) Redacted() Application {
	a.StashGraphQLUrl = Redacted(a.StashGraphQLUrl)
	a.StashApiKey = Redacted(a.StashApiKey)
	return a
}
