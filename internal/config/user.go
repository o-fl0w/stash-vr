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
	appName    = "stash-vr"
	configFile = "config.json"
)

var userConfig *UserConfig

func User(ctx context.Context) UserConfig {
	if userConfig != nil {
		return clone(*userConfig)
	}

	path := resolvePath()
	cfg, err := read(ctx, path)
	if err != nil {
		log.Ctx(ctx).Warn().Msg("failed to load user config, using defaults")
		ensureDefaults(&cfg)
	}

	return cfg
}

func read(ctx context.Context, path string) (UserConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Ctx(ctx).Debug().Err(err).Str("path", path).Msg("error opening file")
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

	path := resolvePath()
	err := write(ctx, path, cfg)
	if err != nil {
		log.Ctx(ctx).Warn().Msg("failed to save user config")
	}

	c := clone(cfg)
	userConfig = &c
}

func write(ctx context.Context, path string, cfg UserConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Ctx(ctx).Debug().Err(err).Str("path", path).Msg("error creating directory")
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		log.Ctx(ctx).Debug().Err(err).Str("path", path).Msg("error writing file")
		return err
	}
	return nil
}

func resolvePath() string {
	// Docker convention
	if fi, err := os.Stat("/config"); err == nil && fi.IsDir() {
		return filepath.Join("/config", configFile)
	}
	// UserConfig config dir
	if base, err := os.UserConfigDir(); err == nil && base != "" {
		return filepath.Join(base, appName, configFile)
	}
	// Fallback to cwd
	return filepath.Join(".", configFile)
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
