package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/calibration"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/productioneval"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fail(err)
	}
	output := flag.String("output", "docs/experiments/codex-replay-2026-08-28", "derived artifact directory")
	embedderKind := flag.String("embedder", "hash", "embedding provider: hash or openai-compatible")
	embeddingURL := flag.String("embedding-url", "http://127.0.0.1:11434/v1/embeddings", "OpenAI-compatible embedding endpoint")
	embeddingModel := flag.String("embedding-model", "nomic-embed-text", "embedding model")
	dimension := flag.Int("dimension", 768, "embedding dimension")
	documentPrefix := flag.String("embedding-document-prefix", "search_document: ", "document embedding prefix")
	queryPrefix := flag.String("embedding-query-prefix", "search_query: ", "query embedding prefix")
	designOnly := flag.Bool("design-only", false, "evaluate design sessions without scoring the frozen confirmation block")
	holdoutOnly := flag.Bool("holdout-only", false, "assign every accepted session to a cold-start temporal holdout block")
	ablations := flag.Bool("ablations", false, "run the frozen output-improvement ablation family")
	maxSessions := flag.Int("max-sessions", 0, "optional deterministic accepted-session cap for development contract retests")
	maxSegmentsPerSession := flag.Int("max-segments-per-session", 0, "optional per-session segment cap for development contract retests")
	rankerEndpoint := flag.String("libravdb-ranker-endpoint", "", "LibraVDB gRPC endpoint for contract-native RankCandidates")
	dataStart := flag.String("data-start", "", "optional inclusive RFC3339 lower bound for a cold-start temporal holdout")
	dataEnd := flag.String("data-end", "2026-08-28T00:00:00Z", "exclusive RFC3339 data boundary")
	calibrationScale := flag.Float64("calibration-scale", 1, "baseline logit calibration scale")
	calibrationBias := flag.Float64("calibration-bias", 0, "baseline logit calibration bias")
	calibrationFloor := flag.Float64("calibration-floor", 1e-6, "baseline calibration probability floor")
	predictiveCalibrationScale := flag.Float64("predictive-calibration-scale", 0, "belief-conditioned score calibration scale; defaults to baseline calibration")
	predictiveCalibrationBias := flag.Float64("predictive-calibration-bias", 0, "belief-conditioned score calibration bias")
	predictiveCalibrationFloor := flag.Float64("predictive-calibration-floor", 0, "belief-conditioned calibration floor; defaults to baseline calibration")
	flag.Parse()
	var activeEmbedder embed.Embedder
	if *embedderKind == "openai-compatible" {
		activeEmbedder, err = embed.NewOpenAICompatible(embed.OpenAICompatibleConfig{
			URL: *embeddingURL, Model: *embeddingModel, Dimension: *dimension, Timeout: 60 * time.Second,
			DocumentPrefix: *documentPrefix, QueryPrefix: *queryPrefix,
		})
		if err != nil {
			fail(err)
		}
	} else if *embedderKind != "hash" {
		fail(fmt.Errorf("unsupported embedder %q", *embedderKind))
	}
	baselineCalibration := calibration.Logit{Scale: *calibrationScale, Bias: *calibrationBias, Floor: *calibrationFloor}
	predictiveCalibration := calibration.Logit{Scale: *predictiveCalibrationScale, Bias: *predictiveCalibrationBias, Floor: *predictiveCalibrationFloor}
	if !predictiveCalibration.Valid() {
		predictiveCalibration = baselineCalibration
	}
	var parsedStart time.Time
	if *dataStart != "" {
		parsedStart = mustTime(*dataStart)
	}
	var candidateRanker retrieval.CandidateRanker = retrieval.PassthroughRanker{}
	var candidateRetriever retrieval.CandidateRetriever
	var textIndexer retrieval.TextIndexer
	if *rankerEndpoint != "" {
		libraRanker, openErr := retrieval.OpenLibraVDBRanker(*rankerEndpoint)
		if openErr != nil {
			fail(openErr)
		}
		defer libraRanker.Close()
		candidateRanker = libraRanker
		candidateRetriever = libraRanker
		textIndexer = libraRanker
	}
	result, err := productioneval.RunCodex(context.Background(), productioneval.CodexConfig{
		SessionDirs: []string{filepath.Join(home, ".codex", "sessions"), filepath.Join(home, ".codex", "archived_sessions")},
		DataStart:   parsedStart, DataEnd: mustTime(*dataEnd), RuleFrozenAt: time.Now().UTC(), Embedder: activeEmbedder,
		DesignOnly: *designOnly, HoldoutOnly: *holdoutOnly, Ablations: *ablations,
		MaxSessions:           *maxSessions,
		MaxSegmentsPerSession: *maxSegmentsPerSession,
		CandidateRanker:       candidateRanker,
		CandidateRetriever:    candidateRetriever, TextIndexer: textIndexer,
		BaselineCalibration: baselineCalibration, PredictiveCalibration: predictiveCalibration,
	})
	if err != nil {
		fail(err)
	}
	artifacts := map[string]any{
		"protocol.json": result.Protocol, "design-dataset.json": result.Design, "design-report.json": result.DesignReport,
		"confirmation-dataset.json": result.Confirmation, "confirmation-report.json": result.ConfirmReport,
		"design-control-report.json": result.DesignControlReport, "confirmation-control-report.json": result.ConfirmControlReport,
	}
	for name, value := range artifacts {
		if err := productioneval.WriteJSON(filepath.Join(*output, name), value); err != nil {
			fail(err)
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(struct {
		Output       string `json:"output"`
		DesignCases  int    `json:"design_cases"`
		ConfirmCases int    `json:"confirmation_cases"`
	}{Output: *output, DesignCases: len(result.Design.Cases), ConfirmCases: len(result.Confirmation.Cases)}); err != nil {
		fail(err)
	}
}

func mustTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "eventframe-codex-eval:", err)
	os.Exit(1)
}
