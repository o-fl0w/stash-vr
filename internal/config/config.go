package config

import (
	"os"
)

const (
	envKeyStashGraphQLUrl = "STASH_GRAPHQL_URL"
)

type Application struct {
	StashGraphQLUrl string
}

func Load() Application {
	config := Application{
		StashGraphQLUrl: getEnvOrDefault(envKeyStashGraphQLUrl, "http://localhost:9999/graphql"),
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
