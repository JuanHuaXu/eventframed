package service_test

import (
	"context"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"math"
	"testing"
)

func TestForecastRescueJournalRestartAndPolicy(t *testing.T) {
	for _, disk := range []bool{false, true} {
		t.Run(fmt.Sprint(disk), func(t *testing.T) {
			cfg, key := evidenceConfig(t, true)
			cfg.WorkingBelief = bayes.GridWorkingPolicy()
			cfg.ForecastRescue = true
			cfg.ResidualMode = service.ResidualModeDisabled
			cfg.BayesianChangePolicy = bayes.ChangePolicy{Hazard: .000001, Threshold: 1, MaxRun: 64}
			path := t.TempDir() + "/rescue.libravdb"
			open := func() store.EventStore {
				if !disk {
					return memorystore.New()
				}
				db, err := libravdbstore.Open(libravdbstore.Config{Path: path, Dimension: 32, EmbeddingModel: "test-hash:d32", Quantization: "none", MemoryMapping: true})
				if err != nil {
					t.Fatal(err)
				}
				return db
			}
			backend := open()
			runtime := evidenceRuntime(t, backend, cfg)
			seedEvidence(t, runtime, 1)
			certifyEvidenceFixture(t, runtime, backend)
			var mix bayes.ForecastMix
			var last model.BayesianOutcomeRequest
			var weights [4]float64
			for i := 0; i < 16; i++ {
				packet := evidenceRecall(t, runtime)
				f := packet.Candidates[0].Forecast
				if !f.ExpertMixture.Enabled || f.ModelKind != "expert-mixture-bernoulli-retrieval-usefulness" || f.CorrectedLaw != f.PreResidualLaw {
					t.Fatal("rescue not final law")
				}
				if math.Abs(f.CorrectedLaw.Useful-mix.Forecast(f.ExpertMixture.Probabilities)) > 1e-14 {
					t.Fatal("double calibration or wrong weights")
				}
				r := signedFeedback(t, packet, fmt.Sprint(i), key)
				r.Useful = false
				signFeedback(t, &r, key)
				// Mutating a client packet must not rewrite durable expert commitments.
				packet.Candidates[0].Forecast.ExpertMixture.Probabilities[0] = .999
				result, err := runtime.ObserveBayesianOutcome(context.Background(), r)
				if err != nil {
					t.Fatal(err)
				}
				mix = mix.Observe(f.ExpertMixture.Probabilities, false, 1)
				weights = result.Posterior.ForecastWeights
				if weights != mix.Weights {
					t.Fatal("did not learn from pre-outcome journal", weights, mix.Weights)
				}
				result.Posterior.ForecastWeights[0] = 999
				dup, err := runtime.ObserveBayesianOutcome(context.Background(), r)
				if err != nil || !dup.Duplicate || dup.Posterior.ForecastWeights != weights {
					t.Fatal("replay changed selector", err)
				}
				last = r
			}
			if disk {
				if err := runtime.Close(); err != nil {
					t.Fatal(err)
				}
				backend = open()
				runtime = evidenceRuntime(t, backend, cfg)
			}
			packet := evidenceRecall(t, runtime)
			f := packet.Candidates[0].Forecast
			if math.Abs(f.CorrectedLaw.Useful-mix.Forecast(f.ExpertMixture.Probabilities)) > 1e-14 {
				t.Fatal("restart lost weights")
			}
			// Changing composition changes policy: the old expert history cannot train
			// the new selector, even though full-stream late feedback is otherwise valid.
			cfg.BayesianScoreWeight = .2
			runtime = evidenceRuntime(t, backend, cfg)
			certifyEvidenceFixture(t, runtime, backend)
			packet = evidenceRecall(t, runtime)
			f = packet.Candidates[0].Forecast
			if math.Abs(f.CorrectedLaw.Useful-(bayes.ForecastMix{}).Forecast(f.ExpertMixture.Probabilities)) > 1e-14 {
				t.Fatal("stale selector served")
			}
			last.Attestation.ObservationID = "late-new-id"
			last.IdempotencyKey = "late-new-id"
			signFeedback(t, &last, key)
			late, err := runtime.ObserveBayesianOutcome(context.Background(), last)
			if err != nil {
				t.Fatal(err)
			}
			if late.Posterior.ForecastWeights != ([4]float64{}) {
				t.Fatal("stale journal trained new selector")
			}
			if err := runtime.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
