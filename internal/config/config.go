package config

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Listen                      string
	DatabasePath                string
	Dimension                   int
	Quantization                string
	RecallK                     int
	PackK                       int
	TokenBudget                 int
	EvidenceOccupancyLimit      int
	EvidenceSimilarity          float64
	LogLevel                    string
	Embedder                    string
	EmbeddingURL                string
	EmbeddingModel              string
	EmbeddingAPIKeyEnv          string
	EmbeddingTimeoutSeconds     int
	EmbeddingDocumentPrefix     string
	EmbeddingQueryPrefix        string
	CalibrationScale            float64
	CalibrationBias             float64
	CalibrationFloor            float64
	PredictiveCalibrationScale  float64
	PredictiveCalibrationBias   float64
	PredictiveCalibrationFloor  float64
	ContextualScoring           bool
	HierarchicalPosterior       bool
	SharedEvidenceWeight        float64
	LibraVDBContractEndpoint    string
	LibraVDBContractTLSMode     string
	LibraVDBContractTLSCA       string
	LibraVDBContractTLSCert     string
	LibraVDBContractTLSKey      string
	LibraVDBContractConcurrency int
	LibraVDBContractTimeoutMS   int
	LibraVDBContractAttempts    int
	LibraVDBCircuitFailures     int
	LibraVDBCircuitCooldownMS   int
	LibraVDBRankerEndpoint      string
	RankDeltaSQLitePath         string
	RankDeltaCacheEntries       int
	ElasticRankDelta            bool
	ElasticRankDeltaMinScale    float64
	ElasticRankDeltaMaxScale    float64
	ResidualMode                string
	MigrateV1                   bool
	MigrateEventFrameCorpus     bool
	ReindexEventFrameContract   bool
	MigrationBackup             string
	AgencyEnabled               bool
	AgencyPrivateKey            string
	AgencyPublicKey             string
	AgencyIssuerToken           string
	AgencyAuthorityToken        string
	BackgroundFuzz              bool
	BackgroundFuzzCertainty     float64
	BackgroundFuzzQueue         int
	BackgroundFuzzIntervalMS    int
	BackgroundFuzzTimeoutMS     int
	BackgroundFuzzCooldownSec   int
	BackgroundFuzzMaxEvents     int
	BackgroundFuzzMaxTrials     int
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
	set.IntVar(&config.EvidenceOccupancyLimit, "evidence-occupancy-limit", 1, "maximum correlated claim/lineage records in one context packet")
	set.Float64Var(&config.EvidenceSimilarity, "evidence-similarity", .85, "same-lineage 5W1H similarity threshold for correlated evidence")
	set.StringVar(&config.LogLevel, "log-level", "info", "log level: debug, info, warn, or error")
	set.StringVar(&config.Embedder, "embedder", "hash", "embedding provider: hash or openai-compatible")
	set.StringVar(&config.EmbeddingURL, "embedding-url", "", "OpenAI-compatible embeddings endpoint")
	set.StringVar(&config.EmbeddingModel, "embedding-model", "", "embedding model name")
	set.StringVar(&config.EmbeddingAPIKeyEnv, "embedding-api-key-env", "EVENTFRAMED_EMBEDDING_API_KEY", "environment variable containing the embedding API key")
	set.IntVar(&config.EmbeddingTimeoutSeconds, "embedding-timeout", 10, "embedding request timeout in seconds")
	set.StringVar(&config.EmbeddingDocumentPrefix, "embedding-document-prefix", "", "optional embedding-model document prefix")
	set.StringVar(&config.EmbeddingQueryPrefix, "embedding-query-prefix", "", "optional embedding-model query prefix")
	set.Float64Var(&config.CalibrationScale, "calibration-scale", 1, "monotonic baseline logit calibration scale")
	set.Float64Var(&config.CalibrationBias, "calibration-bias", 0, "baseline logit calibration bias")
	set.Float64Var(&config.CalibrationFloor, "calibration-floor", 1e-6, "probability floor used by baseline calibration")
	set.Float64Var(&config.PredictiveCalibrationScale, "predictive-calibration-scale", 0, "belief-conditioned score calibration scale; defaults to baseline calibration")
	set.Float64Var(&config.PredictiveCalibrationBias, "predictive-calibration-bias", 0, "belief-conditioned score calibration bias")
	set.Float64Var(&config.PredictiveCalibrationFloor, "predictive-calibration-floor", 0, "belief-conditioned calibration floor; defaults to baseline calibration")
	set.BoolVar(&config.ContextualScoring, "contextual-scoring", false, "enable frozen contextual Bayesian score composition")
	set.BoolVar(&config.HierarchicalPosterior, "hierarchical-posterior", false, "enable weak tenant/horizon posterior shrinkage")
	set.Float64Var(&config.SharedEvidenceWeight, "shared-evidence-weight", .5, "effective weight of each outcome in a certified shared Anti-Pigeon posterior")
	set.StringVar(&config.LibraVDBContractEndpoint, "libravdb-contract-endpoint", "", "optional LibraVDB gRPC endpoint for contract-native indexing, nomination, and ranking (unix:/path or tcp:host:port)")
	set.StringVar(&config.LibraVDBContractTLSMode, "libravdb-contract-tls-mode", "auto", "LibraVDB contract transport: auto, tls, or insecure")
	set.StringVar(&config.LibraVDBContractTLSCA, "libravdb-contract-tls-ca", "", "optional CA certificate for the LibraVDB contract endpoint")
	set.StringVar(&config.LibraVDBContractTLSCert, "libravdb-contract-tls-client-cert", "", "optional mTLS client certificate for the LibraVDB contract endpoint")
	set.StringVar(&config.LibraVDBContractTLSKey, "libravdb-contract-tls-client-key", "", "optional mTLS client key for the LibraVDB contract endpoint")
	set.IntVar(&config.LibraVDBContractConcurrency, "libravdb-contract-concurrency", 16, "maximum concurrent LibraVDB contract RPCs")
	set.IntVar(&config.LibraVDBContractTimeoutMS, "libravdb-contract-timeout-ms", 2000, "per-attempt LibraVDB contract timeout in milliseconds")
	set.IntVar(&config.LibraVDBContractAttempts, "libravdb-contract-attempts", 2, "maximum attempts for retryable LibraVDB contract RPCs")
	set.IntVar(&config.LibraVDBCircuitFailures, "libravdb-circuit-failures", 5, "consecutive retryable failures before opening the LibraVDB circuit")
	set.IntVar(&config.LibraVDBCircuitCooldownMS, "libravdb-circuit-cooldown-ms", 5000, "LibraVDB circuit-open cooldown in milliseconds")
	set.StringVar(&config.LibraVDBRankerEndpoint, "libravdb-ranker-endpoint", "", "deprecated alias for libravdb-contract-endpoint")
	set.StringVar(&config.RankDeltaSQLitePath, "rank-delta-sqlite", filepath.Join(defaultsRoot, "data", "rank-deltas.sqlite"), "SQLite path for durable EventFrame post-retrieval rank deltas")
	set.IntVar(&config.RankDeltaCacheEntries, "rank-delta-cache-entries", 100_000, "maximum in-memory rank-delta cache entries")
	set.BoolVar(&config.ElasticRankDelta, "elastic-rank-delta", true, "modulate bounded rank corrections by answer certainty and correction reliability")
	set.Float64Var(&config.ElasticRankDeltaMinScale, "elastic-rank-delta-min-scale", .5, "rank-delta scale at a certain packing boundary")
	set.Float64Var(&config.ElasticRankDeltaMaxScale, "elastic-rank-delta-max-scale", 2.5, "rank-delta scale at an uncertain packing boundary")
	set.StringVar(&config.ResidualMode, "residual-mode", "apply", "residual output mode: apply, shadow, or disabled")
	set.BoolVar(&config.MigrateV1, "migrate-v1", false, "migrate a Phase 1 database to the durable schema, then exit")
	set.BoolVar(&config.MigrateEventFrameCorpus, "migrate-eventframe-corpus", false, "re-embed a durable full-text corpus as canonical 5W1H EventFrames, then exit")
	set.BoolVar(&config.ReindexEventFrameContract, "reindex-eventframe-contract", false, "rebuild the remote LibraVDB candidate corpus from active local EventFrames, then exit")
	set.StringVar(&config.MigrationBackup, "migration-backup", "", "required absolute backup path for migration")
	set.BoolVar(&config.AgencyEnabled, "agency-enabled", false, "enable signed data-only agency proposal endpoints")
	set.StringVar(&config.AgencyPrivateKey, "agency-private-key", filepath.Join(defaultsRoot, "keys", "agency_ed25519"), "Ed25519 agency private key path")
	set.StringVar(&config.AgencyPublicKey, "agency-public-key", filepath.Join(defaultsRoot, "keys", "agency_ed25519.pub"), "Ed25519 agency public key path")
	set.StringVar(&config.AgencyIssuerToken, "agency-issuer-token", filepath.Join(defaultsRoot, "keys", "agency_issuer.token"), "private agency proposal issuer token path")
	set.StringVar(&config.AgencyAuthorityToken, "agency-authority-token", filepath.Join(defaultsRoot, "keys", "agency_authority.token"), "private OpenClaw authority token path")
	set.BoolVar(&config.BackgroundFuzz, "background-fuzz", true, "enqueue low-certainty recall fuzz audits for an idle background worker")
	set.Float64Var(&config.BackgroundFuzzCertainty, "background-fuzz-certainty", .20, "maximum packing-boundary answer certainty that nominates a background fuzz audit")
	set.IntVar(&config.BackgroundFuzzQueue, "background-fuzz-queue", 128, "maximum in-memory background fuzz jobs")
	set.IntVar(&config.BackgroundFuzzIntervalMS, "background-fuzz-interval-ms", 30_000, "minimum interval between idle background fuzz job starts")
	set.IntVar(&config.BackgroundFuzzTimeoutMS, "background-fuzz-timeout-ms", 30_000, "timeout for one background fuzz job")
	set.IntVar(&config.BackgroundFuzzCooldownSec, "background-fuzz-cooldown-seconds", 900, "deduplication cooldown after a background fuzz job")
	set.IntVar(&config.BackgroundFuzzMaxEvents, "background-fuzz-max-events", 8, "maximum bounded frontier events in one background fuzz job")
	set.IntVar(&config.BackgroundFuzzMaxTrials, "background-fuzz-max-trials", 8, "maximum source-bundle perturbations in one background fuzz job")
	if err := set.Parse(args); err != nil {
		return Config{}, err
	}
	if config.Dimension <= 0 || config.RecallK <= 0 || config.PackK <= 0 || config.TokenBudget <= 0 {
		return Config{}, errors.New("dimension and budgets must be positive")
	}
	if config.PackK > config.RecallK {
		return Config{}, errors.New("pack-k cannot exceed recall-k")
	}
	if config.EvidenceOccupancyLimit <= 0 || config.EvidenceOccupancyLimit > config.PackK {
		return Config{}, errors.New("evidence-occupancy-limit must be in [1,pack-k]")
	}
	if math.IsNaN(config.EvidenceSimilarity) || math.IsInf(config.EvidenceSimilarity, 0) || config.EvidenceSimilarity <= 0 || config.EvidenceSimilarity > 1 {
		return Config{}, errors.New("evidence-similarity must be in (0,1]")
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
	if config.CalibrationScale <= 0 || config.CalibrationFloor <= 0 || config.CalibrationFloor >= .5 {
		return Config{}, errors.New("calibration scale must be positive and floor must be in (0,0.5)")
	}
	if config.PredictiveCalibrationScale != 0 && (config.PredictiveCalibrationScale <= 0 || config.PredictiveCalibrationFloor <= 0 || config.PredictiveCalibrationFloor >= .5) {
		return Config{}, errors.New("predictive calibration scale must be positive and floor must be in (0,0.5)")
	}
	if config.SharedEvidenceWeight <= 0 || config.SharedEvidenceWeight > 1 {
		return Config{}, errors.New("shared-evidence-weight must be in (0,1]")
	}
	if config.ResidualMode != "apply" && config.ResidualMode != "shadow" && config.ResidualMode != "disabled" {
		return Config{}, errors.New("residual-mode must be apply, shadow, or disabled")
	}
	if !filepath.IsAbs(config.RankDeltaSQLitePath) || config.RankDeltaCacheEntries <= 0 {
		return Config{}, errors.New("rank-delta-sqlite must be absolute and rank-delta-cache-entries must be positive")
	}
	if config.ElasticRankDeltaMinScale < 0 || config.ElasticRankDeltaMaxScale < config.ElasticRankDeltaMinScale || config.ElasticRankDeltaMaxScale > 10 {
		return Config{}, errors.New("elastic rank-delta scales must satisfy 0 <= min <= max <= 10")
	}
	if filepath.Clean(config.RankDeltaSQLitePath) == filepath.Clean(config.DatabasePath) {
		return Config{}, errors.New("rank-delta SQLite and LibraVDB database paths must differ")
	}
	if config.LibraVDBContractEndpoint != "" && config.LibraVDBRankerEndpoint != "" && config.LibraVDBContractEndpoint != config.LibraVDBRankerEndpoint {
		return Config{}, errors.New("libravdb contract and deprecated ranker endpoints disagree")
	}
	if config.LibraVDBContractEndpoint == "" {
		config.LibraVDBContractEndpoint = config.LibraVDBRankerEndpoint
	}
	if config.LibraVDBContractEndpoint != "" && !strings.HasPrefix(config.LibraVDBContractEndpoint, "unix:") && !strings.HasPrefix(config.LibraVDBContractEndpoint, "tcp:") {
		return Config{}, errors.New("libravdb-contract-endpoint must begin with unix: or tcp:")
	}
	if config.LibraVDBContractTLSMode != "auto" && config.LibraVDBContractTLSMode != "tls" && config.LibraVDBContractTLSMode != "insecure" {
		return Config{}, errors.New("libravdb-contract-tls-mode must be auto, tls, or insecure")
	}
	if (config.LibraVDBContractTLSCert == "") != (config.LibraVDBContractTLSKey == "") {
		return Config{}, errors.New("LibraVDB contract TLS client certificate and key must be configured together")
	}
	if config.LibraVDBContractTLSMode == "insecure" && (config.LibraVDBContractTLSCA != "" || config.LibraVDBContractTLSCert != "") {
		return Config{}, errors.New("LibraVDB contract TLS files cannot be used in insecure mode")
	}
	if config.LibraVDBContractConcurrency <= 0 || config.LibraVDBContractConcurrency > 1024 {
		return Config{}, errors.New("libravdb-contract-concurrency must be in [1,1024]")
	}
	if config.LibraVDBContractTimeoutMS <= 0 || config.LibraVDBContractAttempts <= 0 || config.LibraVDBContractAttempts > 4 || config.LibraVDBCircuitFailures <= 0 || config.LibraVDBCircuitCooldownMS <= 0 {
		return Config{}, errors.New("LibraVDB resilience controls must be positive and attempts must not exceed 4")
	}
	if config.MigrateV1 && config.MigrateEventFrameCorpus {
		return Config{}, errors.New("migrate-v1 and migrate-eventframe-corpus are mutually exclusive")
	}
	if config.ReindexEventFrameContract && (config.MigrateV1 || config.MigrateEventFrameCorpus) {
		return Config{}, errors.New("reindex-eventframe-contract must run separately after migration")
	}
	if config.ReindexEventFrameContract && config.LibraVDBContractEndpoint == "" {
		return Config{}, errors.New("reindex-eventframe-contract requires libravdb-contract-endpoint")
	}
	if (config.MigrateV1 || config.MigrateEventFrameCorpus) && !filepath.IsAbs(config.MigrationBackup) {
		return Config{}, errors.New("migration requires an absolute migration-backup path")
	}
	if config.AgencyEnabled && !distinctAbsolutePaths(config.AgencyPrivateKey, config.AgencyPublicKey, config.AgencyIssuerToken, config.AgencyAuthorityToken) {
		return Config{}, errors.New("agency private key, public key, issuer token, and authority token paths must be distinct absolute paths")
	}
	if config.AgencyEnabled && !strings.HasPrefix(config.Listen, "unix://") {
		return Config{}, errors.New("agency mode requires a local Unix socket listener")
	}
	if config.BackgroundFuzz && (config.BackgroundFuzzCertainty <= 0 || config.BackgroundFuzzCertainty > 1 || config.BackgroundFuzzQueue <= 0 || config.BackgroundFuzzQueue > 4096 || config.BackgroundFuzzIntervalMS <= 0 || config.BackgroundFuzzTimeoutMS <= 0 || config.BackgroundFuzzCooldownSec < 0 || config.BackgroundFuzzMaxEvents < 2 || config.BackgroundFuzzMaxEvents > 64 || config.BackgroundFuzzMaxTrials <= 0 || config.BackgroundFuzzMaxTrials > 512) {
		return Config{}, errors.New("background fuzz trigger, queue, timing, and audit bounds are invalid")
	}
	return config, nil
}

func (c Config) EnsureDirectories() error {
	if err := os.MkdirAll(filepath.Dir(c.DatabasePath), 0o700); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(c.RankDeltaSQLitePath), 0o700); err != nil {
		return fmt.Errorf("create rank-delta directory: %w", err)
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
