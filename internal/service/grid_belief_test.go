package service_test

import (
	"context"
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

func TestGridAuthenticatedServingAndRestart(t *testing.T) {
	for _, disk := range []bool{false, true} {
		name := "memory"
		if disk {
			name = "disk"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			cfg, key := evidenceConfig(t, true)
			cfg.WorkingBelief = bayes.GridWorkingPolicy()
			path := t.TempDir() + "/grid.libravdb"
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
			packet := evidenceRecall(t, runtime)
			request := signedFeedback(t, packet, "grid-one", key)
			r, err := runtime.ObserveBayesianOutcome(ctx, request)
			if err != nil {
				t.Fatal(err)
			}
			want := *r.Posterior.WorkingBelief
			r.Posterior.WorkingBelief.GridWeights[0] = 999
			r, err = runtime.ObserveBayesianOutcome(ctx, request)
			if err != nil || !r.Duplicate || *r.Posterior.WorkingBelief != want {
				t.Fatal("replay/alias invariant", err)
			}
			packet = evidenceRecall(t, runtime)
			c := packet.Candidates[0]
			if c.Forecast.BeliefLaw == nil || c.Forecast.BeliefLaw.Useful != want.PredictiveUseful || c.Forecast.ModelKind != "working-fixed-share-grid-bernoulli-retrieval-usefulness" {
				t.Fatal("grid not served")
			}
			if math.Abs(c.PredictiveScore-(.9*c.BaselineScore+.1*want.PredictiveUseful)) > 1e-12 {
				t.Fatal("grid not scored", c.PredictiveScore)
			}
			if disk {
				if err = runtime.Close(); err != nil {
					t.Fatal(err)
				}
				backend = open()
				runtime = evidenceRuntime(t, backend, cfg)
				r, err = runtime.ObserveBayesianOutcome(ctx, request)
				if err != nil || !r.Duplicate || *r.Posterior.WorkingBelief != want {
					t.Fatal("restart lost state", err)
				}
				packet = evidenceRecall(t, runtime)
				if packet.Candidates[0].Forecast.BeliefLaw == nil || packet.Candidates[0].Forecast.BeliefLaw.Useful != want.PredictiveUseful {
					t.Fatal("restart changed predictive law")
				}
			}
			if err = runtime.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
