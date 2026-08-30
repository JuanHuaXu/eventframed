package claimsexperiment

import (
	"context"
	"fmt"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func RunResidualUtility(ctx context.Context) (ResidualReport, error) {
	const training, evaluation = 60, 40
	policies := []residual.Policy{
		{Clip: .15, MinSupport: 1e9, MinConfidence: 1, ConfidenceDelta: .5, MotionLimit: 0, MaxAge: 24 * time.Hour, ImprovementDelta: .001},
		{Clip: .15, MinSupport: 8, MinConfidence: .1, ConfidenceDelta: .5, MotionLimit: .05, MaxAge: 24 * time.Hour, ImprovementDelta: .001},
	}
	runtimes := make([]*service.Service, 0, len(policies))
	for index, policy := range policies {
		memory := memorystore.New()
		embedder, _ := embed.NewHashEmbedder(8)
		runtime, err := service.New(memory, embedder, service.Config{
			DefaultRecallK: 1, DefaultPackK: 1, DefaultTokenBudget: 100,
			BayesianPolicy:       bayes.Policy{VectorWeight: 1, Threshold: 0, CriticalThreshold: 0, AuditProbability: 1, MaxActive: 1, AuditSeed: fmt.Sprintf("residual-%d", index)},
			BayesianChangePolicy: bayes.ChangePolicy{Hazard: .001, Threshold: 1, MaxRun: 64}, ResidualPolicy: policy,
		})
		if err != nil {
			return ResidualReport{}, err
		}
		event := testutil.Event("biased-memory", "repeated systematic distractor", experimentAnchor.Add(-time.Hour))
		event.TenantID, event.SessionID, event.Embedding, event.EmbeddingModel = "residual-experiment", "source", []float32{1, 0, 0, 0, 0, 0, 0, 0}, embedder.ModelKey()
		if err := observeEvent(ctx, runtime, event); err != nil {
			return ResidualReport{}, err
		}
		runtimes = append(runtimes, runtime)
	}
	var losses [2]float64
	applied := 0
	for turn := 0; turn < training+evaluation; turn++ {
		asOf := experimentAnchor.Add(time.Duration(turn) * time.Minute)
		for index, runtime := range runtimes {
			packet, err := runtime.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "residual-experiment", SessionID: "trajectory", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: "feature-hash-v1:d8", AsOf: asOf, RecallK: 1, PackK: 1, TokenBudget: 100})
			if err != nil {
				return ResidualReport{}, err
			}
			candidate := packet.Candidates[0]
			if turn >= training {
				losses[index] += square(candidate.Forecast.CorrectedLaw.Useful)
				if index == 1 && candidate.Forecast.ResidualApplied {
					applied++
				}
			}
			if err := submitOutcome(ctx, runtime, packet, candidate.Event.ID, false, fmt.Sprintf("residual-%d-%03d", index, turn), asOf.Add(time.Second)); err != nil {
				return ResidualReport{}, err
			}
		}
	}
	base, corrected := losses[0]/evaluation, losses[1]/evaluation
	return ResidualReport{TrainingCases: training, EvaluationCases: evaluation, BaselineBrier: base, ResidualBrier: corrected, AbsoluteGain: base - corrected, RelativeReduction: (base - corrected) / base, ResidualAppliedCases: applied, ResidualAppliedRate: float64(applied) / evaluation, SystematicOutcomeRate: 0}, nil
}
