package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/evaluation"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

const tenant = "synthetic-claims"

type runner struct {
	name     string
	runtime  *service.Service
	feedback string
}

func main() {
	output := flag.String("output", "", "optional evaluation dataset output path")
	flag.Parse()
	dataset, err := runExperiment()
	if err != nil {
		fail(err)
	}
	if *output != "" {
		payload, marshalErr := json.MarshalIndent(dataset, "", "  ")
		if marshalErr != nil {
			fail(marshalErr)
		}
		if writeErr := os.WriteFile(*output, append(payload, '\n'), 0o644); writeErr != nil {
			fail(writeErr)
		}
	}
	report, err := evaluation.Evaluate(dataset)
	if err != nil {
		fail(err)
	}
	stableDataset, shiftedDataset := dataset, dataset
	stableDataset.Cases = stableDataset.Cases[:80]
	shiftedDataset.Cases = shiftedDataset.Cases[80:]
	stable, err := evaluation.Evaluate(stableDataset)
	if err != nil {
		fail(err)
	}
	shifted, err := evaluation.Evaluate(shiftedDataset)
	if err != nil {
		fail(err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(struct {
		Overall      evaluation.Report `json:"overall"`
		StableRegime evaluation.Report `json:"stable_regime"`
		HiddenShift  evaluation.Report `json:"hidden_shift"`
	}{Overall: report, StableRegime: stable, HiddenShift: shifted}); err != nil {
		fail(err)
	}
}

func runExperiment() (evaluation.Dataset, error) {
	ctx := context.Background()
	now := time.Now().UTC()
	frozen := now.Add(-30 * time.Hour)
	seedTime := frozen.Add(-time.Hour)
	runners := make([]runner, 0, 5)
	definitions := []struct {
		name, feedback string
		policy         bayes.Policy
		residual       residual.Policy
	}{
		{"baseline", "none", bayes.Policy{VectorWeight: .6, NoveltyWeight: .15, IndependenceWeight: .1, Threshold: 2, CriticalThreshold: 2, AuditProbability: .02, MaxActive: 20, AuditSeed: "baseline"}, disabledResidual()},
		{"update_all", "all", bayes.Policy{VectorWeight: .6, NoveltyWeight: .15, IndependenceWeight: .1, Threshold: 0, CriticalThreshold: 0, AuditProbability: 1, MaxActive: 20, AuditSeed: "all"}, disabledResidual()},
		{"frontier_all_deep_selective", "all", bayes.Policy{VectorWeight: .6, NoveltyWeight: .15, IndependenceWeight: .1, Threshold: .72, CriticalThreshold: .55, AuditProbability: 1, MaxActive: 8, AuditSeed: "frontier-all-deep-selective", CheapUpdateAll: true}, disabledResidual()},
		{"selective", "selected", bayes.Policy{VectorWeight: .6, NoveltyWeight: .15, IndependenceWeight: .1, Threshold: .72, CriticalThreshold: .55, AuditProbability: .02, MaxActive: 8, AuditSeed: "selective"}, disabledResidual()},
		{"eventframe", "selected", bayes.Policy{VectorWeight: .6, NoveltyWeight: .15, IndependenceWeight: .1, Threshold: .72, CriticalThreshold: .55, AuditProbability: .02, MaxActive: 8, AuditSeed: "eventframe"}, residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .1, ConfidenceDelta: .5, MotionLimit: .2, MaxAge: 7 * 24 * time.Hour, ImprovementDelta: .001}},
	}
	for _, definition := range definitions {
		memory := memorystore.New()
		embedder, _ := embed.NewHashEmbedder(24)
		runtime, err := service.New(memory, embedder, service.Config{
			DefaultRecallK: 20, DefaultPackK: 20, DefaultTokenBudget: 4_000, OverfetchMultiplier: 2,
			BayesianPolicy: definition.policy, BayesianChangePolicy: bayes.ChangePolicy{Hazard: .001, Threshold: .99, MaxRun: 64},
			ResidualPolicy: definition.residual,
		})
		if err != nil {
			return evaluation.Dataset{}, err
		}
		for group := 0; group < 4; group++ {
			for member := 0; member < 5; member++ {
				id := fmt.Sprintf("g%d-m%d", group, member)
				event := testutil.Event(id, fmt.Sprintf("group %d memory %d", group, member), seedTime)
				event.TenantID, event.SessionID, event.Priority = tenant, "source", float64(member)/5
				event.Embedding, event.EmbeddingModel = eventVector(group, member), embedder.ModelKey()
				if _, err := runtime.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: event}); err != nil {
					return evaluation.Dataset{}, err
				}
			}
		}
		if definition.feedback != "none" {
			if err := publishSyntheticCertificates(ctx, runtime, memory.Snapshot(ctx), definition.policy.AuditProbability, now); err != nil {
				return evaluation.Dataset{}, err
			}
		}
		runners = append(runners, runner{name: definition.name, runtime: runtime, feedback: definition.feedback})
	}

	dataset := evaluation.Dataset{
		SchemaVersion: evaluation.SchemaVersion, EvaluationBlock: "confirmation", BaselineVariant: "baseline",
		PolicyFrozenAt: frozen, EmbargoSeconds: 3600, PriorityWeightScale: 4, BootstrapSamples: 2_000, BootstrapSeed: 260828,
	}
	for turn := 0; turn < 120; turn++ {
		predicted := frozen.Add(time.Duration(turn+7) * 10 * time.Minute)
		queryGroup := turn % 4
		targetGroup := queryGroup
		if turn >= 80 {
			targetGroup = (queryGroup + 1) % 4
		}
		priority := .4
		if turn >= 80 {
			priority = .95
		}
		item := evaluation.Case{
			ID: fmt.Sprintf("turn-%03d", turn), TrajectoryID: fmt.Sprintf("trajectory-%d", turn%4),
			PredictedAt: predicted, OutcomeAvailableAt: predicted.Add(time.Minute), Priority: priority,
			UniverseEventIDs: groupIDs(-1), RelevantEventIDs: groupIDs(targetGroup), Variants: make(map[string]evaluation.VariantForecast),
		}
		for runnerIndex := range runners {
			current := &runners[runnerIndex]
			packet, err := current.runtime.Recall(ctx, model.RecallRequest{
				ProtocolVersion: model.ProtocolVersion, TenantID: tenant, SessionID: item.TrajectoryID,
				Embedding: queryVector(queryGroup), EmbeddingModel: "feature-hash-v1:d24", AsOf: predicted,
				RecallK: 20, PackK: 20, TokenBudget: 4_000,
			})
			if err != nil {
				return evaluation.Dataset{}, fmt.Errorf("%s turn %d recall: %w", current.name, turn, err)
			}
			decisions := make(map[string]model.BayesianDecision, len(packet.BayesianShadow.Decisions))
			for _, decision := range packet.BayesianShadow.Decisions {
				decisions[decision.EventID] = decision
			}
			forecast := evaluation.VariantForecast{StateAsOf: predicted}
			for _, candidate := range packet.Candidates {
				decision := decisions[candidate.Event.ID]
				forecast.Candidates = append(forecast.Candidates, evaluation.CandidateForecast{
					EventID: candidate.Event.ID, SourceAvailableAt: candidate.Event.AvailableAt,
					Probability: candidate.Forecast.CorrectedLaw.Useful, Nominated: true, Activated: decision.Activated,
				})
			}
			item.Variants[current.name] = forecast
			if current.feedback == "none" {
				continue
			}
			for _, decision := range packet.BayesianShadow.Decisions {
				if current.feedback == "selected" && !decision.Activated {
					continue
				}
				useful := eventGroup(decision.EventID) == targetGroup
				source, inclusion := model.OutcomeFullStream, 1.0
				if current.feedback == "selected" {
					source, inclusion = model.OutcomeSelected, .5
				}
				_, err := current.runtime.ObserveBayesianOutcome(ctx, model.BayesianOutcomeRequest{
					ProtocolVersion: model.ProtocolVersion, IdempotencyKey: fmt.Sprintf("%s-%03d-%s", current.name, turn, decision.EventID),
					TenantID: tenant, JournalID: packet.BayesianShadow.JournalID, EventID: decision.EventID, Useful: useful,
					ObservedAt: item.OutcomeAvailableAt, AvailableAt: item.OutcomeAvailableAt, Source: source, InclusionProbability: inclusion,
				})
				if err != nil {
					return evaluation.Dataset{}, fmt.Errorf("%s turn %d outcome %s: %w", current.name, turn, decision.EventID, err)
				}
			}
		}
		dataset.Cases = append(dataset.Cases, item)
	}
	return dataset, nil
}

func publishSyntheticCertificates(ctx context.Context, runtime *service.Service, snapshot model.Snapshot, auditProbability float64, now time.Time) error {
	selection := model.SelectionSupportCertificate{ID: "synthetic-selection", TenantID: tenant, PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch, MinSelectionProbability: .5, SimultaneousCoverage: .95, Procedure: "exhaustive synthetic frontier", Issuer: "eventframe-experiment", ExternalAudit: true, ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour)}
	if _, err := runtime.PublishSelectionCertificate(ctx, model.PublishSelectionCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: selection}); err != nil {
		return err
	}
	omitted := model.OmittedInfluenceCertificate{ID: "synthetic-omitted", TenantID: tenant, PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch, DivergenceUCB: .01, DivergenceLimit: .05, AuditProbability: auditProbability, SimultaneousCoverage: .95, Procedure: "exhaustive synthetic shadow", Issuer: "eventframe-experiment", ExternalAudit: true, ValidUntil: now.Add(time.Hour)}
	_, err := runtime.PublishOmittedInfluenceCertificate(ctx, model.PublishOmittedInfluenceCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: omitted})
	return err
}

func disabledResidual() residual.Policy {
	return residual.Policy{Clip: .15, MinSupport: 1e9, MinConfidence: 1, ConfidenceDelta: .05, MotionLimit: 0, MaxAge: 7 * 24 * time.Hour, ImprovementDelta: .001}
}
func queryVector(group int) []float32 {
	vector := make([]float32, 24)
	vector[group] = 1
	for index := 0; index < 20; index++ {
		vector[4+index] = float32(index+1) / 10000
	}
	return vector
}
func eventVector(group, member int) []float32 {
	vector := make([]float32, 24)
	vector[group] = 1
	vector[4+group*5+member] = .1
	return vector
}
func groupIDs(group int) []string {
	out := make([]string, 0, 20)
	for g := 0; g < 4; g++ {
		if group >= 0 && g != group {
			continue
		}
		for m := 0; m < 5; m++ {
			out = append(out, fmt.Sprintf("g%d-m%d", g, m))
		}
	}
	return out
}
func eventGroup(id string) int { var group int; _, _ = fmt.Sscanf(id, "g%d-", &group); return group }
func fail(err error)           { fmt.Fprintln(os.Stderr, "eventframe-experiment:", err); os.Exit(1) }
