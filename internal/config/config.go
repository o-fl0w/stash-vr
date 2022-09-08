package config

import (
	"os"
	"sync"
)

const (
	envKeyStashGraphQLUrl = "STASH_GRAPHQL_URL"
	envKeyStashApiKey     = "STASH_API_KEY"
)

type Application struct {
	StashGraphQLUrl string
	StashApiKey     string
}

var cfg Application

var once sync.Once

func Get() Application {
	once.Do(func() {
		cfg = Application{
			StashGraphQLUrl: getEnvOrDefault(envKeyStashGraphQLUrl, "http://localhost:9999/graphql"),
			StashApiKey:     getEnvOrDefault(envKeyStashApiKey, ""),
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
