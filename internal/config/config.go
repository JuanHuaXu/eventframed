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
	Listen                  string
	DatabasePath            string
	Dimension               int
	Quantization            string
	RecallK                 int
	PackK                   int
	TokenBudget             int
	LogLevel                string
	Embedder                string
	EmbeddingURL            string
	EmbeddingModel          string
	EmbeddingAPIKeyEnv      string
	EmbeddingTimeoutSeconds int
	MigrateV1               bool
	MigrationBackup         string
	AgencyEnabled           bool
	AgencyPrivateKey        string
	AgencyPublicKey         string
	AgencyIssuerToken       string
	AgencyAuthorityToken    string
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
	set.StringVar(&config.Embedder, "embedder", "hash", "embedding provider: hash or openai-compatible")
	set.StringVar(&config.EmbeddingURL, "embedding-url", "", "OpenAI-compatible embeddings endpoint")
	set.StringVar(&config.EmbeddingModel, "embedding-model", "", "embedding model name")
	set.StringVar(&config.EmbeddingAPIKeyEnv, "embedding-api-key-env", "EVENTFRAMED_EMBEDDING_API_KEY", "environment variable containing the embedding API key")
	set.IntVar(&config.EmbeddingTimeoutSeconds, "embedding-timeout", 10, "embedding request timeout in seconds")
	set.BoolVar(&config.MigrateV1, "migrate-v1", false, "migrate a Phase 1 database to the durable schema, then exit")
	set.StringVar(&config.MigrationBackup, "migration-backup", "", "required absolute backup path for migration")
	set.BoolVar(&config.AgencyEnabled, "agency-enabled", false, "enable signed data-only agency proposal endpoints")
	set.StringVar(&config.AgencyPrivateKey, "agency-private-key", filepath.Join(defaultsRoot, "keys", "agency_ed25519"), "Ed25519 agency private key path")
	set.StringVar(&config.AgencyPublicKey, "agency-public-key", filepath.Join(defaultsRoot, "keys", "agency_ed25519.pub"), "Ed25519 agency public key path")
	set.StringVar(&config.AgencyIssuerToken, "agency-issuer-token", filepath.Join(defaultsRoot, "keys", "agency_issuer.token"), "private agency proposal issuer token path")
	set.StringVar(&config.AgencyAuthorityToken, "agency-authority-token", filepath.Join(defaultsRoot, "keys", "agency_authority.token"), "private OpenClaw authority token path")
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
	if config.Embedder != "hash" && config.Embedder != "openai-compatible" {
		return Config{}, fmt.Errorf("unsupported embedder %q", config.Embedder)
	}
	if config.Embedder == "openai-compatible" && (config.EmbeddingURL == "" || config.EmbeddingModel == "") {
		return Config{}, errors.New("openai-compatible embedder requires embedding-url and embedding-model")
	}
	if config.EmbeddingTimeoutSeconds <= 0 {
		return Config{}, errors.New("embedding-timeout must be positive")
	}
	if config.MigrateV1 && !filepath.IsAbs(config.MigrationBackup) {
		return Config{}, errors.New("migrate-v1 requires an absolute migration-backup path")
	}
	if config.AgencyEnabled && !distinctAbsolutePaths(config.AgencyPrivateKey, config.AgencyPublicKey, config.AgencyIssuerToken, config.AgencyAuthorityToken) {
		return Config{}, errors.New("agency private key, public key, issuer token, and authority token paths must be distinct absolute paths")
	}
	if config.AgencyEnabled && !strings.HasPrefix(config.Listen, "unix://") {
		return Config{}, errors.New("agency mode requires a local Unix socket listener")
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
	if c.AgencyEnabled {
		if err := os.MkdirAll(filepath.Dir(c.AgencyPrivateKey), 0o700); err != nil {
			return fmt.Errorf("create agency key directory: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(c.AgencyPublicKey), 0o700); err != nil {
			return fmt.Errorf("create agency public-key directory: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(c.AgencyIssuerToken), 0o700); err != nil {
			return fmt.Errorf("create agency issuer-token directory: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(c.AgencyAuthorityToken), 0o700); err != nil {
			return fmt.Errorf("create agency authority-token directory: %w", err)
		}
	}
	return nil
}

func distinctAbsolutePaths(paths ...string) bool {
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			return false
		}
		clean := filepath.Clean(path)
		if _, exists := seen[clean]; exists {
			return false
		}
		seen[clean] = struct{}{}
	}
	return true
}
