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
	envKeyLogLevel         = "LOG_LEVEL"
	envKeyDisableLogColor  = "DISABLE_LOG_COLOR"
	envKeyDisableRedact    = "DISABLE_REDACT"
	envKeyForceHTTPS       = "FORCE_HTTPS"
	envKeyHeatmapHeightPx  = "HEATMAP_HEIGHT_PX"
	envKeyDisablePlayCount = "DISABLE_PLAY_COUNT"
	envKeyExcludeSortName  = "EXCLUDE_SORT_NAME"
)

type ApplicationConfig struct {
	ListenAddress       string
	StashGraphQLUrl     string
	StashApiKey         string
	FavoriteTag         string
	LogLevel            string
	DisableLogColor     bool
	IsRedactDisabled    bool
	ForceHTTPS          bool
	HeatmapHeightPx     int
	IsPlayCountDisabled bool
	ExcludeSortName     string
}

var applicationConfig ApplicationConfig

func Init() {
	pflag.String(envKeyListenAddress, ":9666", "Local address for Stash-VR to listen on")
	_ = viper.BindPFlag(envKeyListenAddress, pflag.Lookup(envKeyListenAddress))

	pflag.String(envKeyStashGraphQLUrl, "http://localhost:9999/graphql", "Url to Stash graphql")
	_ = viper.BindPFlag(envKeyStashGraphQLUrl, pflag.Lookup(envKeyStashGraphQLUrl))

	pflag.String(envKeyStashApiKey, "", "Stash API key")
	_ = viper.BindPFlag(envKeyStashApiKey, pflag.Lookup(envKeyStashApiKey))

	pflag.String(envKeyFavoriteTag, "FAVORITE", "Name of tag in Stash to hold scenes marked as favorites")
	_ = viper.BindPFlag(envKeyFavoriteTag, pflag.Lookup(envKeyFavoriteTag))

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

	pflag.String(envKeyExcludeSortName, "hidden", "Exclude tags with this sort name")
	_ = viper.BindPFlag(envKeyExcludeSortName, pflag.Lookup(envKeyExcludeSortName))

	pflag.BoolP("help", "h", false, "Display usage information")
	_ = viper.BindPFlag("help", pflag.Lookup("help"))

	pflag.Parse()

	if viper.GetBool("help") {
		pflag.Usage()
		os.Exit(1)
	}

	viper.AutomaticEnv()

	applicationConfig.ListenAddress = viper.GetString(envKeyListenAddress)
	applicationConfig.StashGraphQLUrl = viper.GetString(envKeyStashGraphQLUrl)
	applicationConfig.StashApiKey = viper.GetString(envKeyStashApiKey)
	applicationConfig.FavoriteTag = viper.GetString(envKeyFavoriteTag)
	applicationConfig.LogLevel = strings.ToLower(viper.GetString(envKeyLogLevel))
	applicationConfig.DisableLogColor = viper.GetBool(envKeyDisableLogColor)
	applicationConfig.IsRedactDisabled = viper.GetBool(envKeyDisableRedact)
	applicationConfig.ForceHTTPS = viper.GetBool(envKeyForceHTTPS)
	applicationConfig.HeatmapHeightPx = viper.GetInt(envKeyHeatmapHeightPx)
	applicationConfig.IsPlayCountDisabled = viper.GetBool(envKeyDisablePlayCount)
	applicationConfig.ExcludeSortName = viper.GetString(envKeyExcludeSortName)

}

func Application() ApplicationConfig {
	return applicationConfig
}

func (a ApplicationConfig) Redacted() ApplicationConfig {
	a.StashGraphQLUrl = Redacted(a.StashGraphQLUrl)
	a.StashApiKey = Redacted(a.StashApiKey)
	return a
}
