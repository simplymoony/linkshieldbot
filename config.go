package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

//go:embed config.toml.dist
var cfgDist []byte

type config struct {
	Verbose bool `toml:"verbose"`

	PollerTimeout  int `toml:"poller_timeout,omitempty"`
	HandlerTimeout int `toml:"handler_timeout,omitempty"`

	Directives map[string]int64 `toml:"directives"`
}

// Loads config file at path into cfg, creating it if necessary.
func loadConfig(path string, cfg *config) (notfound bool, err error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			notfound = true
			if err = os.WriteFile(path, cfgDist, 0755); err != nil {
				return notfound, err
			}
		} else {
			return notfound, err
		}
	} else {
		if info.IsDir() {
			return notfound, fmt.Errorf("path must be a file, not a directory")
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return notfound, err
	}

	if err = toml.Unmarshal(data, cfg); err != nil {
		return notfound, err
	}

	return notfound, nil
}
