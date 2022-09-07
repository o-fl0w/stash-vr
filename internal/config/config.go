package config

import (
	"os"
)

const (
	envKeyStashGraphQLUrl = "STASH_GRAPHQL_URL"
	envKeyStashApiKey     = "STASH_API_KEY"
)

type Application struct {
	StashGraphQLUrl string
	StashApiKey     string
}

func Load() Application {
	config := Application{
		StashGraphQLUrl: getEnvOrDefault(envKeyStashGraphQLUrl, "http://localhost:9999/graphql"),
		StashApiKey:     getEnvOrDefault(envKeyStashApiKey, ""),
	}
	return config
}

func getEnvOrDefault(key string, defaultValue string) string {
	if val, ok := os.LookupEnv(key); !ok {
		return defaultValue
	} else {
		return val
	}
}
