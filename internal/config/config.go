package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strings"
)

const (
	envKeyListenAddress    = "LISTEN_ADDRESS"
	envKeyStashGraphQLUrl  = "STASH_GRAPHQL_URL"
	envKeyStashApiKey      = "STASH_API_KEY"
	envKeyFavoriteTag      = "FAVORITE_TAG"
	envKeyFilters          = "FILTERS"
	envKeyLogLevel         = "LOG_LEVEL"
	envKeyDisableLogColor  = "DISABLE_LOG_COLOR"
	envKeyDisableRedact    = "DISABLE_REDACT"
	envKeyForceHTTPS       = "FORCE_HTTPS"
	envKeyHeatmapHeightPx  = "HEATMAP_HEIGHT_PX"
	envKeyDisablePlayCount = "DISABLE_PLAY_COUNT"
)

type Application struct {
	ListenAddress       string
	StashGraphQLUrl     string
	StashApiKey         string
	FavoriteTag         string
	Filters             string
	LogLevel            string
	DisableLogColor     bool
	IsRedactDisabled    bool
	ForceHTTPS          bool
	HeatmapHeightPx     int
	IsPlayCountDisabled bool
}

var cfg Application

func Init() {
	pflag.String(envKeyListenAddress, ":9666", "Local address for Stash-VR to listen on")
	_ = viper.BindPFlag(envKeyListenAddress, pflag.Lookup(envKeyListenAddress))

	pflag.String(envKeyStashGraphQLUrl, "http://localhost:9999/graphql", "Url to Stash graphql")
	_ = viper.BindPFlag(envKeyStashGraphQLUrl, pflag.Lookup(envKeyStashGraphQLUrl))

	pflag.String(envKeyStashApiKey, "", "Stash API key")
	_ = viper.BindPFlag(envKeyStashApiKey, pflag.Lookup(envKeyStashApiKey))

	pflag.String(envKeyFavoriteTag, "FAVORITE", "Name of tag in Stash to hold scenes marked as favorites")
	_ = viper.BindPFlag(envKeyFavoriteTag, pflag.Lookup(envKeyFavoriteTag))

	pflag.String(envKeyFilters, "", "Narrow the selection of filters to show. Either 'frontpage' or a comma seperated list of filter ids")
	_ = viper.BindPFlag(envKeyFilters, pflag.Lookup(envKeyFilters))

	pflag.String(envKeyLogLevel, "info", "Set log level - trace, debug, warn, info or error")
	_ = viper.BindPFlag(envKeyLogLevel, pflag.Lookup(envKeyLogLevel))

	pflag.Bool(envKeyDisableLogColor, false, "Disable colors in log output")
	_ = viper.BindPFlag(envKeyDisableLogColor, pflag.Lookup(envKeyDisableLogColor))

	pflag.Bool(envKeyDisableRedact, false, "Disable redacting sensitive information from logs")
	_ = viper.BindPFlag(envKeyDisableRedact, pflag.Lookup(envKeyDisableRedact))

	pflag.Bool(envKeyForceHTTPS, false, "Force Stash-VR to use HTTPS")
	_ = viper.BindPFlag(envKeyForceHTTPS, pflag.Lookup(envKeyForceHTTPS))

	pflag.Int(envKeyHeatmapHeightPx, 0, "Height of heatmaps")
	_ = viper.BindPFlag(envKeyHeatmapHeightPx, pflag.Lookup(envKeyHeatmapHeightPx))

	pflag.Bool(envKeyDisablePlayCount, false, "Disable incrementing Stash play count for scenes")
	_ = viper.BindPFlag(envKeyDisablePlayCount, pflag.Lookup(envKeyDisablePlayCount))

	pflag.BoolP("help", "h", false, "Display usage information")
	_ = viper.BindPFlag("help", pflag.Lookup("help"))

	pflag.Parse()

	if viper.GetBool("help") {
		pflag.Usage()
		os.Exit(1)
	}

	viper.AutomaticEnv()

	cfg.ListenAddress = viper.GetString(envKeyListenAddress)
	cfg.StashGraphQLUrl = viper.GetString(envKeyStashGraphQLUrl)
	cfg.StashApiKey = viper.GetString(envKeyStashApiKey)
	cfg.FavoriteTag = viper.GetString(envKeyFavoriteTag)
	cfg.Filters = viper.GetString(envKeyFilters)
	cfg.LogLevel = strings.ToLower(viper.GetString(envKeyLogLevel))
	cfg.DisableLogColor = viper.GetBool(envKeyDisableLogColor)
	cfg.IsRedactDisabled = viper.GetBool(envKeyDisableRedact)
	cfg.ForceHTTPS = viper.GetBool(envKeyForceHTTPS)
	cfg.HeatmapHeightPx = viper.GetInt(envKeyHeatmapHeightPx)
	cfg.IsPlayCountDisabled = viper.GetBool(envKeyDisablePlayCount)

}

func Get() Application {
	return cfg
}

func (a Application) Redacted() Application {
	a.StashGraphQLUrl = Redacted(a.StashGraphQLUrl)
	a.StashApiKey = Redacted(a.StashApiKey)
	return a
}
