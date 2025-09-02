package config

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"path/filepath"
)

type Filter struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Disabled bool   `json:"disabled"`
}

type UserConfig struct {
	Filters []Filter `json:"filters"`
}

const (
	configFile = "config.json"
)

var userConfig *UserConfig

func User(ctx context.Context) UserConfig {
	if userConfig != nil {
		return clone(*userConfig)
	}

	if Application().ConfigPath == "" {
		cfg := UserConfig{}
		ensureDefaults(&cfg)
		return cfg
	}

	path := resolvePath()
	cfg, err := read(ctx, path)
	if err != nil {
		ensureDefaults(&cfg)
	}

	return cfg
}

func read(ctx context.Context, path string) (UserConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Str("path", path).Msg("error opening config file")
		return UserConfig{}, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	var cfg UserConfig
	if err := dec.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
		return UserConfig{}, err
	}
	return cfg, nil
}

func Save(ctx context.Context, cfg UserConfig) {
	ensureDefaults(&cfg)
	c := clone(cfg)
	userConfig = &c

	if Application().ConfigPath == "" {
		log.Ctx(ctx).Warn().Msg("attempt to save user config but CONFIG_PATH not specified, config will apply but not persist")
		return
	}

	path := resolvePath()
	err := write(ctx, path, cfg)
	if err != nil {
		log.Ctx(ctx).Warn().Msg("failed to save user config")
	}
}

func write(ctx context.Context, path string, cfg UserConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Ctx(ctx).Debug().Err(err).Str("path", path).Msg("error creating config directory")
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		log.Ctx(ctx).Debug().Err(err).Str("path", path).Msg("error writing config file")
		return err
	}
	return nil
}

func resolvePath() string {
	return filepath.Join(Application().ConfigPath, configFile)
}

func ensureDefaults(u *UserConfig) {
	if u.Filters == nil {
		u.Filters = []Filter{}
	}
}

func clone(u UserConfig) UserConfig {
	out := UserConfig{Filters: make([]Filter, len(u.Filters))}
	copy(out.Filters, u.Filters)
	return out
}
