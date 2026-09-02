package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/agency"
	"github.com/JuanHuaXu/eventframed/internal/api"
	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/calibration"
	"github.com/JuanHuaXu/eventframed/internal/config"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"github.com/JuanHuaXu/eventframed/internal/rankdelta"
	"github.com/JuanHuaXu/eventframed/internal/ranking"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "eventframed:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	settings, err := config.Parse(args)
	if err != nil {
		return err
	}
	if err := settings.EnsureDirectories(); err != nil {
		return err
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: parseLogLevel(settings.LogLevel)}))
	activeEmbedder, err := buildEmbedder(settings)
	if err != nil {
		return err
	}
	storeConfig := libravdbstore.Config{
		Path:           settings.DatabasePath,
		Dimension:      settings.Dimension,
		Quantization:   settings.Quantization,
		MemoryMapping:  true,
		EmbeddingModel: activeEmbedder.ModelKey(),
	}
	if settings.MigrateV1 {
		if err := libravdbstore.MigrateLegacy(context.Background(), storeConfig, activeEmbedder, settings.MigrationBackup); err != nil {
			return err
		}
		fmt.Println("eventframed: migration completed; backup:", settings.MigrationBackup)
		return nil
	}
	if settings.MigrateEventFrameCorpus {
		if err := libravdbstore.MigrateEventFrameCorpus(context.Background(), storeConfig, activeEmbedder, settings.MigrationBackup); err != nil {
			return err
		}
		fmt.Println("eventframed: EventFrame corpus migration completed; backup:", settings.MigrationBackup)
		return nil
	}
	eventStore, err := libravdbstore.Open(storeConfig)
	if err != nil {
		return err
	}
	if settings.ReindexEventFrameContract {
		defer eventStore.Close()
		contracts, openErr := retrieval.OpenLibraVDBContractsWithConfig(retrieval.ContractClientConfig{
			Endpoint: settings.LibraVDBContractEndpoint, TLSMode: settings.LibraVDBContractTLSMode,
			CAFile: settings.LibraVDBContractTLSCA, ClientCertFile: settings.LibraVDBContractTLSCert,
			ClientKeyFile: settings.LibraVDBContractTLSKey,
		})
		if openErr != nil {
			return openErr
		}
		defer contracts.Close()
		events, listErr := eventStore.ListAllEvents(context.Background())
		if listErr != nil {
			return listErr
		}
		for _, event := range events {
			if indexErr := service.IndexEventFrame(context.Background(), contracts, "eventframe-", event); indexErr != nil {
				return fmt.Errorf("reindex EventFrame %q: %w", event.ID, indexErr)
			}
		}
		fmt.Println("eventframed: remote EventFrame corpus rebuilt; events:", len(events))
		return nil
	}
	rankDeltaStore, err := rankdelta.Open(settings.RankDeltaSQLitePath, settings.RankDeltaCacheEntries)
	if err != nil {
		_ = eventStore.Close()
		return err
	}
	var agencySigner *agency.Signer
	var agencyIssuerToken string
	var agencyAuthorityToken string
	if settings.AgencyEnabled {
		agencySigner, err = agency.LoadOrCreateSigner(settings.AgencyPrivateKey, settings.AgencyPublicKey)
		if err != nil {
			_ = rankDeltaStore.Close()
			_ = eventStore.Close()
			return fmt.Errorf("load agency signing key: %w", err)
		}
		agencyIssuerToken, err = agency.LoadOrCreateIssuerToken(settings.AgencyIssuerToken)
		if err != nil {
			_ = rankDeltaStore.Close()
			_ = eventStore.Close()
			return fmt.Errorf("load agency issuer token: %w", err)
		}
		agencyAuthorityToken, err = agency.LoadOrCreateAuthorityToken(settings.AgencyAuthorityToken)
		if err != nil {
			_ = rankDeltaStore.Close()
			_ = eventStore.Close()
			return fmt.Errorf("load agency authority token: %w", err)
		}
	}
	baselineCalibration := calibration.Logit{Scale: settings.CalibrationScale, Bias: settings.CalibrationBias, Floor: settings.CalibrationFloor}
	predictiveCalibration := calibration.Logit{Scale: settings.PredictiveCalibrationScale, Bias: settings.PredictiveCalibrationBias, Floor: settings.PredictiveCalibrationFloor}
	if !predictiveCalibration.Valid() {
		predictiveCalibration = baselineCalibration
	}
	rankingPolicy := ranking.DefaultPolicy()
	rankingPolicy.ContextualEnabled = settings.ContextualScoring
	rankingPolicy.HierarchicalEnabled = settings.HierarchicalPosterior
	var candidateRanker retrieval.CandidateRanker = retrieval.PassthroughRanker{}
	var candidateRetriever retrieval.CandidateRetriever
	var candidateIndex retrieval.CandidateIndex
	var externalReadiness retrieval.ReadinessProbe
	var closeContracts func() error
	if settings.LibraVDBContractEndpoint != "" {
		libraContracts, openErr := retrieval.OpenLibraVDBContractsWithConfig(retrieval.ContractClientConfig{
			Endpoint: settings.LibraVDBContractEndpoint, TLSMode: settings.LibraVDBContractTLSMode,
			CAFile: settings.LibraVDBContractTLSCA, ClientCertFile: settings.LibraVDBContractTLSCert,
			ClientKeyFile:  settings.LibraVDBContractTLSKey,
			MaxConcurrent:  settings.LibraVDBContractConcurrency,
			RequestTimeout: time.Duration(settings.LibraVDBContractTimeoutMS) * time.Millisecond,
			MaxAttempts:    settings.LibraVDBContractAttempts, FailureThreshold: settings.LibraVDBCircuitFailures,
			OpenDuration: time.Duration(settings.LibraVDBCircuitCooldownMS) * time.Millisecond,
		})
		if openErr != nil {
			_ = rankDeltaStore.Close()
			_ = eventStore.Close()
			return openErr
		}
		candidateRanker = libraContracts
		candidateRetriever = libraContracts
		candidateIndex = libraContracts
		externalReadiness = libraContracts
		closeContracts = libraContracts.Close
		defer closeContracts()
	}
	runtime, err := service.New(eventStore, activeEmbedder, service.Config{
		DefaultRecallK:      settings.RecallK,
		DefaultPackK:        settings.PackK,
		DefaultTokenBudget:  settings.TokenBudget,
		OverfetchMultiplier: 4,
		Quantization:        settings.Quantization,
		BaselineCalibration: baselineCalibration, PredictiveCalibration: predictiveCalibration,
		BayesianGroupPolicy: bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 8, MaxMembers: 64, EquivalenceMargin: .15, EquivalenceThreshold: .80, MaxUncertainBorrowing: .10, SharedEvidenceWeight: settings.SharedEvidenceWeight},
		PackingPolicy:       packing.Policy{EvidenceOccupancyLimit: settings.EvidenceOccupancyLimit, EvidenceSimilarity: settings.EvidenceSimilarity},
		RankingPolicy:       rankingPolicy, CandidateRanker: candidateRanker,
		CandidateRankerRequired: settings.LibraVDBContractEndpoint != "",
		CandidateRetriever:      candidateRetriever, CandidateRetrieverRequired: settings.LibraVDBContractEndpoint != "",
		CandidateIndex: candidateIndex, CandidateCollectionPrefix: "eventframe-",
		ExternalReadiness: externalReadiness,
		RankDeltaStore:    rankDeltaStore, RankDeltaStoreRequired: true,
		ElasticRankDelta: ranking.ElasticDeltaPolicy{
			Enabled: settings.ElasticRankDelta, MinScale: settings.ElasticRankDeltaMinScale, MaxScale: settings.ElasticRankDeltaMaxScale,
		},
		ResidualMode:         settings.ResidualMode,
		AgencyPolicy:         agency.DefaultPolicy(settings.AgencyEnabled),
		AgencySigner:         agencySigner,
		AgencyIssuerToken:    agencyIssuerToken,
		AgencyAuthorityToken: agencyAuthorityToken,
		BackgroundFuzz: service.BackgroundFuzzPolicy{
			Enabled: settings.BackgroundFuzz, AnswerCertaintyThreshold: settings.BackgroundFuzzCertainty,
			QueueCapacity: settings.BackgroundFuzzQueue, WorkerInterval: time.Duration(settings.BackgroundFuzzIntervalMS) * time.Millisecond,
			JobTimeout: time.Duration(settings.BackgroundFuzzTimeoutMS) * time.Millisecond,
			Cooldown:   time.Duration(settings.BackgroundFuzzCooldownSec) * time.Second,
			MaxEvents:  settings.BackgroundFuzzMaxEvents, MaxPerturbations: settings.BackgroundFuzzMaxTrials,
		},
	})
	if err != nil {
		_ = rankDeltaStore.Close()
		_ = eventStore.Close()
		return err
	}
	defer runtime.Close()

	listener, cleanup, err := listen(settings.Listen)
	if err != nil {
		return err
	}
	defer cleanup()

	server := &http.Server{
		Handler:           api.NewServer(runtime, logger).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	logger.Info("eventframed started", "version", version, "listen", settings.Listen, "database", settings.DatabasePath, "dimension", settings.Dimension, "quantization", settings.Quantization, "embedder", activeEmbedder.Name(), "embedding_model", activeEmbedder.ModelKey())

	shutdownContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-shutdownContext.Done():
		context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(context); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
	}
	logger.Info("eventframed stopped")
	return nil
}

func buildEmbedder(settings config.Config) (embed.Embedder, error) {
	if settings.Embedder == "hash" {
		return embed.NewHashEmbedder(settings.Dimension)
	}
	return embed.NewOpenAICompatible(embed.OpenAICompatibleConfig{
		URL: settings.EmbeddingURL, Model: settings.EmbeddingModel, APIKey: os.Getenv(settings.EmbeddingAPIKeyEnv),
		Dimension: settings.Dimension, Timeout: time.Duration(settings.EmbeddingTimeoutSeconds) * time.Second,
		DocumentPrefix: settings.EmbeddingDocumentPrefix, QueryPrefix: settings.EmbeddingQueryPrefix,
	})
}

func listen(endpoint string) (net.Listener, func(), error) {
	if strings.HasPrefix(endpoint, "tcp://") {
		listener, err := net.Listen("tcp", strings.TrimPrefix(endpoint, "tcp://"))
		return listener, func() {}, err
	}
	path := strings.TrimPrefix(endpoint, "unix://")
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			return nil, func() {}, fmt.Errorf("refusing to replace non-socket path %s", path)
		}
		if err := os.Remove(path); err != nil {
			return nil, func() {}, fmt.Errorf("remove stale socket: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, func() {}, err
	}
	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, func() {}, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		_ = listener.Close()
		return nil, func() {}, err
	}
	return listener, func() { _ = os.Remove(path) }, nil
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
