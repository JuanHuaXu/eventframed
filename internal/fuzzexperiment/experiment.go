package fuzzexperiment

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/fuzzing"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

const (
	stabilityThreshold        = .05
	requiredStableProbability = .90
	confidenceLevel           = .95
	minTrials                 = 24
)

type Result struct {
	SchemaVersion             string                     `json:"schema_version"`
	Dataset                   string                     `json:"dataset"`
	Split                     string                     `json:"split"`
	PredictorKind             string                     `json:"predictor_kind"`
	OutputFunctional          string                     `json:"output_functional"`
	DistanceKind              string                     `json:"distance_kind"`
	StabilityThreshold        float64                    `json:"stability_threshold"`
	RequiredStableProbability float64                    `json:"required_stable_probability"`
	ConfidenceLevel           float64                    `json:"confidence_level"`
	Queries                   int                        `json:"queries"`
	Trials                    int                        `json:"trials"`
	Properties                []model.FuzzPropertyReport `json:"properties"`
	Interpretation            []string                   `json:"interpretation"`
}

type corpusRow struct {
	Split   string                   `json:"split"`
	Capture model.CaptureTurnRequest `json:"capture"`
	Oracle  *struct {
		QueryBeforeCapture string   `json:"query_before_capture"`
		RelevantPriorIDs   []string `json:"relevant_prior_ids"`
		ObsoletePriorIDs   []string `json:"obsolete_prior_ids"`
	} `json:"oracle"`
}

func Run(path, split string) (Result, error) {
	rows, err := load(path)
	if err != nil {
		return Result{}, err
	}
	embedder, err := embed.NewHashEmbedder(256)
	if err != nil {
		return Result{}, err
	}
	events := make(map[string]model.Event, len(rows))
	for _, row := range rows {
		events[row.Capture.Turn.ID] = frame.FromTurn(row.Capture.Turn)
	}
	allTrials := make([]model.FuzzTrialResult, 0)
	queries := 0
	predictorKind := ""
	for _, row := range rows {
		if row.Split != split || row.Oracle == nil || len(row.Oracle.RelevantPriorIDs) == 0 {
			continue
		}
		target, ok := events[row.Oracle.RelevantPriorIDs[0]]
		if !ok {
			return Result{}, fmt.Errorf("query %q references missing target", row.Capture.Turn.ID)
		}
		distractor, ok := chooseDistractor(rows, events, row, target)
		if !ok {
			continue
		}
		contextEvents := []model.Event{target, distractor}
		predictor, err := fuzzing.NewEmbeddingNominationPredictor(context.Background(), embedder, row.Oracle.QueryBeforeCapture)
		if err != nil {
			return Result{}, err
		}
		predictorKind = predictor.Kind()
		perturbations := []model.FuzzPerturbation{
			envelopePerturbation(row.Capture.Turn.ID+":envelope", target),
			bundlePerturbation(row.Capture.Turn.ID+":cross-domain", "cross-domain-semantic-bundle", target, distractor, "complete public distractor EventFrame"),
		}
		if alternative, found := chooseRelevantAlternative(events, row.Oracle.RelevantPriorIDs, target.ID, row.Capture.Turn.AvailableAt); found {
			contextEvents = append(contextEvents, alternative)
			perturbations = append(perturbations, bundlePerturbation(row.Capture.Turn.ID+":paraphrase", "same-answer-paraphrase-bundle", target, alternative, "oracle-linked public paraphrase EventFrame"))
		}
		eventIDs := make([]string, len(contextEvents))
		for index := range contextEvents {
			eventIDs[index] = contextEvents[index].ID
		}
		request := model.FuzzSensitivityRequest{
			TenantID: target.TenantID, Query: row.Oracle.QueryBeforeCapture, AsOf: row.Capture.Turn.AvailableAt,
			EventIDs: eventIDs, Perturbations: perturbations,
			StabilityThreshold: stabilityThreshold, RequiredStableProbability: requiredStableProbability,
			ConfidenceLevel: confidenceLevel, MinTrials: 1,
		}
		response, err := fuzzing.Evaluate(context.Background(), request, contextEvents, predictor)
		if err != nil {
			return Result{}, fmt.Errorf("query %q: %w", row.Capture.Turn.ID, err)
		}
		allTrials = append(allTrials, response.Trials...)
		queries++
	}
	if queries == 0 {
		return Result{}, errors.New("no eligible public-fact queries")
	}
	properties := fuzzing.Summarize(allTrials, requiredStableProbability, confidenceLevel, minTrials)
	return Result{
		SchemaVersion: "eventframe.fuzz-public-facts.v1", Dataset: path, Split: split,
		PredictorKind: predictorKind, OutputFunctional: "normalized-context-embedding-nomination-law", DistanceKind: "total-variation",
		StabilityThreshold: stabilityThreshold, RequiredStableProbability: requiredStableProbability, ConfidenceLevel: confidenceLevel,
		Queries: queries, Trials: len(allTrials), Properties: properties,
		Interpretation: []string{
			"The experiment measures local model sensitivity, not real-world causal effects.",
			"A cross-domain bundle is a complete 5W1H semantic bundle copied from a different sourced public-fact EventFrame.",
			"Conditional-invariant status requires Bonferroni-simultaneous one-sided Wilson lower bounds and the frozen minimum trial count.",
			"Evaluation-oracle links construct the valid paraphrase family; this does not demonstrate autonomous discovery of a domain translation.",
		},
	}, nil
}

func load(path string) ([]corpusRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	rows := make([]corpusRow, 0, 384)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	for scanner.Scan() {
		var row corpusRow
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, scanner.Err()
}

func chooseDistractor(rows []corpusRow, events map[string]model.Event, query corpusRow, target model.Event) (model.Event, bool) {
	excluded := make(map[string]struct{}, len(query.Oracle.RelevantPriorIDs)+len(query.Oracle.ObsoletePriorIDs))
	for _, id := range append(append([]string(nil), query.Oracle.RelevantPriorIDs...), query.Oracle.ObsoletePriorIDs...) {
		excluded[id] = struct{}{}
	}
	candidates := make([]model.Event, 0)
	for _, row := range rows {
		event := events[row.Capture.Turn.ID]
		if row.Capture.Turn.SessionID != query.Capture.Turn.SessionID || !event.AvailableAt.Before(query.Capture.Turn.AvailableAt) || event.ID == target.ID {
			continue
		}
		if _, skip := excluded[event.ID]; !skip {
			candidates = append(candidates, event)
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
	if len(candidates) == 0 {
		return model.Event{}, false
	}
	return candidates[0], true
}

func chooseRelevantAlternative(events map[string]model.Event, ids []string, targetID string, availableBy time.Time) (model.Event, bool) {
	for _, id := range ids {
		if id != targetID {
			if event, ok := events[id]; ok && !event.AvailableAt.After(availableBy) {
				return event, true
			}
		}
	}
	return model.Event{}, false
}

func envelopePerturbation(id string, target model.Event) model.FuzzPerturbation {
	return model.FuzzPerturbation{
		ID: id, PropertyID: "context-envelope", EventID: target.ID, ValidityRuleID: "public-fact-valid-context-relocation-v1",
		ValidationKind: model.FuzzValidationDeclaredRelocation,
		Replacements: map[model.FuzzField]model.Field{
			model.FuzzWho:   synthetic("another evaluation participant", "participant substitution leaves the public proposition unchanged"),
			model.FuzzWhere: synthetic("session:public-fact-transfer", "session relocation leaves the public proposition unchanged"),
		},
	}
}

func bundlePerturbation(id, property string, target, source model.Event, evidence string) model.FuzzPerturbation {
	return model.FuzzPerturbation{
		ID: id, PropertyID: property, EventID: target.ID, ValidityRuleID: "public-fact-complete-semantic-bundle-v1",
		ValidationKind: model.FuzzValidationSourceEventBundle, SourceEventID: source.ID,
		Replacements: map[model.FuzzField]model.Field{
			model.FuzzWhat: synthetic(source.What.Value, evidence+": what"),
			model.FuzzWhy:  synthetic(source.Why.Value, evidence+": why"),
			model.FuzzHow:  synthetic(source.How.Value, evidence+": how"),
		},
	}
}

func synthetic(value, evidence string) model.Field {
	return model.Field{Value: value, Source: model.SourceSynthetic, Confidence: 1, Evidence: evidence}
}

func Write(path string, result Result) error {
	payload, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return os.WriteFile(path, payload, 0o644)
}

func DefaultOutputName(now time.Time) string {
	return "fuzz-public-facts-" + now.UTC().Format("20060102T150405Z") + ".json"
}
