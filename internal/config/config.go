package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Listen       string
	DatabasePath string
	Dimension    int
	Quantization string
	RecallK      int
	PackK        int
	TokenBudget  int
	LogLevel     string
}

func Parse(args []string) (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("resolve home directory: %w", err)
	}
	defaultsRoot := filepath.Join(home, ".eventframed")
	set := flag.NewFlagSet("eventframed", flag.ContinueOnError)
	var config Config
	set.StringVar(&config.Listen, "listen", "unix://"+filepath.Join(defaultsRoot, "run", "eventframed.sock"), "listen endpoint (unix://path or tcp://host:port)")
	set.StringVar(&config.DatabasePath, "database", filepath.Join(defaultsRoot, "data", "eventframe.libravdb"), "LibraVDB database path")
	set.IntVar(&config.Dimension, "dimension", 768, "embedding dimension")
	set.StringVar(&config.Quantization, "quantization", "sq8", "traversal quantization: none, sq8, fsq6, or pq8")
	set.IntVar(&config.RecallK, "recall-k", 50, "default candidate recall budget")
	set.IntVar(&config.PackK, "pack-k", 10, "default final packing budget")
	set.IntVar(&config.TokenBudget, "token-budget", 2000, "default memory token budget")
	set.StringVar(&config.LogLevel, "log-level", "info", "log level: debug, info, warn, or error")
	if err := set.Parse(args); err != nil {
		return Config{}, err
	}
	if config.Dimension <= 0 || config.RecallK <= 0 || config.PackK <= 0 || config.TokenBudget <= 0 {
		return Config{}, errors.New("dimension and budgets must be positive")
	}
	if config.PackK > config.RecallK {
		return Config{}, errors.New("pack-k cannot exceed recall-k")
	}
	switch config.Quantization {
	case "none", "sq8", "fsq6", "pq8":
	default:
		return Config{}, fmt.Errorf("unsupported quantization %q", config.Quantization)
	}
	if !strings.HasPrefix(config.Listen, "unix://") && !strings.HasPrefix(config.Listen, "tcp://") {
		return Config{}, errors.New("listen must begin with unix:// or tcp://")
	}
	return config, nil
}

func (c Config) EnsureDirectories() error {
	if err := os.MkdirAll(filepath.Dir(c.DatabasePath), 0o700); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}
	if strings.HasPrefix(c.Listen, "unix://") {
		if err := os.MkdirAll(filepath.Dir(strings.TrimPrefix(c.Listen, "unix://")), 0o700); err != nil {
			return fmt.Errorf("create socket directory: %w", err)
		}
	}
	return nil
}
