package service_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/epistemic"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func evidenceConfig(t testing.TB, enabled bool) (service.Config, ed25519.PrivateKey) {
	t.Helper()
	config := service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2000, OverfetchMultiplier: 2}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if enabled {
		now := time.Now().UTC()
		config.EvidenceVerifier, err = epistemic.NewVerifier([]epistemic.IssuerKey{{Issuer: "test-observer", KeyID: "v1", TenantID: "tenant-a", Source: model.OutcomeFullStream, PublicKey: pub, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour)}}, 64)
		if err != nil {
			t.Fatal(err)
		}
		config.WorkingBelief = bayes.DefaultWorkingPolicy()
	}
	return config, priv
}

func evidenceRuntime(t testing.TB, backend store.EventStore, cfg service.Config) *service.Service {
	t.Helper()
	emb, err := embed.NewHashEmbedder(32)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := service.New(backend, emb, cfg)
	if err != nil {
		t.Fatal(err)
	}
	return runtime
}

func evidenceRecall(t testing.TB, runtime *service.Service) model.ContextPacket {
	t.Helper()
	packet, err := runtime.Recall(context.Background(), model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "test", Query: "public evidence", AsOf: time.Now().UTC(), RecallK: 50, PackK: 10, TokenBudget: 2000})
	if err != nil {
		t.Fatal(err)
	}
	return packet
}

func seedEvidence(t testing.TB, runtime *service.Service, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		e := testutil.Event(fmt.Sprintf("public-%d", i), fmt.Sprintf("public evidence topic %d", i), time.Now().Add(-time.Minute))
		if _, err := runtime.Observe(context.Background(), model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: e.ID, Event: e}); err != nil {
			t.Fatal(err)
		}
	}
}

func certifyEvidenceFixture(t testing.TB, runtime *service.Service, backend store.EventStore) {
	t.Helper()
	ctx := context.Background()
	s := backend.Snapshot(ctx)
	now := time.Now().UTC()
	// Synthetic exhaustive fixture authority, not an empirical certificate.
	selection := model.SelectionSupportCertificate{ID: "fixture-selection", TenantID: "tenant-a", PolicyVersion: s.PolicyVersion, EvidenceEpoch: s.EvidenceEpoch, MinSelectionProbability: 1, SimultaneousCoverage: .95, Procedure: "exhaustive synthetic fixture", Issuer: "fixture", ExternalAudit: true, ValidFrom: now.Add(-time.Minute), ValidUntil: now.Add(time.Hour)}
	if _, err := runtime.PublishSelectionCertificate(ctx, model.PublishSelectionCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: selection}); err != nil {
		t.Fatal(err)
	}
	omitted := model.OmittedInfluenceCertificate{ID: "fixture-omitted", TenantID: "tenant-a", PolicyVersion: s.PolicyVersion, EvidenceEpoch: s.EvidenceEpoch, DivergenceUCB: 0, DivergenceLimit: .05, AuditProbability: .02, SimultaneousCoverage: .95, Procedure: "exhaustive synthetic fixture", Issuer: "fixture", ExternalAudit: true, ValidUntil: now.Add(time.Hour)}
	if _, err := runtime.PublishOmittedInfluenceCertificate(ctx, model.PublishOmittedInfluenceCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: omitted}); err != nil {
		t.Fatal(err)
	}
}

func TestAuthenticatedBeliefScoredAndTrustChange(t *testing.T) {
	cfg, key := evidenceConfig(t, true)
	backend := memorystore.New()
	runtime := evidenceRuntime(t, backend, cfg)
	seedEvidence(t, runtime, 1)
	certifyEvidenceFixture(t, runtime, backend)
	packet := evidenceRecall(t, runtime)
	r := signedFeedback(t, packet, "first", key)
	result, err := runtime.ObserveBayesianOutcome(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	packet = evidenceRecall(t, runtime)
	if packet.Candidates[0].Forecast.BeliefLaw == nil {
		t.Fatal("working belief did not reach scored bundle")
	}
	if packet.Candidates[0].Forecast.BeliefLaw.Useful != result.Posterior.WorkingBelief.PredictiveUseful {
		t.Fatal("scored wrong predictive map")
	}
	// A new enrolled registry must not inherit old trust even after new audit
	// certificates are published. A fresh observation starts a fresh belief.
	next, nextKey := evidenceConfig(t, true)
	runtime = evidenceRuntime(t, backend, next)
	defer runtime.Close()
	certifyEvidenceFixture(t, runtime, backend)
	packet = evidenceRecall(t, runtime)
	if packet.Candidates[0].Forecast.BeliefLaw != nil {
		t.Fatal("old trust belief served")
	}
	comparison, err := runtime.CompareBayesianGroup(context.Background(), model.BayesianGroupComparisonRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", MemberEventIDs: []string{r.EventID, "unobserved"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range comparison.Members {
		if member.UsefulWeight != 0 || member.NotUsefulWeight != 0 {
			t.Fatal("old trust evidence entered group comparison")
		}
	}
	fresh := signedFeedback(t, packet, "second", nextKey)
	changed, err := runtime.ObserveBayesianOutcome(context.Background(), fresh)
	if err != nil {
		t.Fatal(err)
	}
	if changed.Posterior.Alpha != 2 || changed.Posterior.EffectiveSupport != 1 {
		t.Fatal("old trust statistics survived")
	}
}

func signedFeedback(t testing.TB, packet model.ContextPacket, id string, key ed25519.PrivateKey) model.BayesianOutcomeRequest {
	t.Helper()
	now := time.Now().UTC()
	r := model.BayesianOutcomeRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, TenantID: "tenant-a", JournalID: packet.BayesianShadow.JournalID, EventID: packet.Candidates[0].Event.ID, Useful: true, ObservedAt: now, AvailableAt: now, Source: model.OutcomeFullStream, InclusionProbability: 1}
	if key != nil {
		r.Attestation = &model.EvidenceAttestation{Issuer: "test-observer", KeyID: "v1", ObservationID: id}
		signFeedback(t, &r, key)
	}
	return r
}

func signFeedback(t testing.TB, r *model.BayesianOutcomeRequest, key ed25519.PrivateKey) {
	t.Helper()
	payload, err := epistemic.OutcomeSigningBytes(*r)
	if err != nil {
		t.Fatal(err)
	}
	r.Attestation.Signature = ed25519.Sign(key, payload)
}

func TestAuthenticatedOutcomeReplayAndRestart(t *testing.T) {
	for _, persistent := range []bool{false, true} {
		t.Run(fmt.Sprint(persistent), func(t *testing.T) {
			cfg, key := evidenceConfig(t, true)
			dbcfg := libravdbstore.Config{Path: t.TempDir() + "/evidence.libravdb", Dimension: 32, Quantization: "none", MemoryMapping: true, EmbeddingModel: "test-hash:d32"}
			var backend store.EventStore = memorystore.New()
			if persistent {
				db, err := libravdbstore.Open(dbcfg)
				if err != nil {
					t.Fatal(err)
				}
				backend = db
			}
			runtime := evidenceRuntime(t, backend, cfg)
			defer func() { _ = runtime.Close() }()
			seedEvidence(t, runtime, 2)
			packet := evidenceRecall(t, runtime)
			request := signedFeedback(t, packet, "unique-observation", key)
			unsigned := request
			unsigned.Attestation = nil
			if _, err := runtime.ObserveBayesianOutcome(context.Background(), unsigned); err == nil {
				t.Fatal("unsigned admitted")
			}
			var accepted atomic.Int32
			var wg sync.WaitGroup
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					r, err := runtime.ObserveBayesianOutcome(context.Background(), request)
					if err != nil {
						t.Error(err)
					} else if !r.Duplicate {
						accepted.Add(1)
					}
				}()
			}
			wg.Wait()
			if accepted.Load() != 1 {
				t.Fatalf("counted %d times", accepted.Load())
			}
			result, err := runtime.ObserveBayesianOutcome(context.Background(), request)
			if err != nil || !result.Duplicate || result.Posterior.WorkingBelief == nil {
				t.Fatalf("retry: %+v %v", result, err)
			}
			want := *result.Posterior.WorkingBelief
			result.Posterior.WorkingBelief.LogOdds = 999
			result, err = runtime.ObserveBayesianOutcome(context.Background(), request)
			if err != nil || *result.Posterior.WorkingBelief != want {
				t.Fatal("response aliases stored belief")
			}
			other := evidenceRecall(t, runtime)
			replay := request
			copyAttestation := *request.Attestation
			replay.Attestation = &copyAttestation
			replay.JournalID = other.BayesianShadow.JournalID
			replay.ObservedAt = time.Now().UTC()
			replay.AvailableAt = replay.ObservedAt
			signFeedback(t, &replay, key)
			if _, err := runtime.ObserveBayesianOutcome(context.Background(), replay); err == nil {
				t.Fatal("new journal laundered observation")
			}
			if persistent {
				if err := runtime.Close(); err != nil {
					t.Fatal(err)
				}
				db, err := libravdbstore.Open(dbcfg)
				if err != nil {
					t.Fatal(err)
				}
				runtime = evidenceRuntime(t, db, cfg)
				r, err := runtime.ObserveBayesianOutcome(context.Background(), request)
				if err != nil || !r.Duplicate || r.Posterior.WorkingBelief == nil || *r.Posterior.WorkingBelief != want {
					t.Fatalf("restart: %+v %v", r, err)
				}
			}
		})
	}
}

func TestWorkingSplitResetMaterialization(t *testing.T) {
	for _, disk := range []bool{false, true} {
		t.Run(fmt.Sprint(disk), func(t *testing.T) {
			var backend store.EventStore = memorystore.New()
			if disk {
				db, err := libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/split.libravdb", Dimension: 32, EmbeddingModel: "test-hash:d32", Quantization: "none", MemoryMapping: true})
				if err != nil {
					t.Fatal(err)
				}
				backend = db
			}
			defer backend.Close()
			ctx := context.Background()
			snapshot := backend.Snapshot(ctx)
			cert := model.AntiPigeonCertificate{ID: "working-group", TenantID: "tenant-a", MemberEventIDs: []string{"a", "b"}, GraphVersion: snapshot.GraphVersion, EvidenceEpoch: snapshot.EvidenceEpoch}
			if _, err := backend.PublishAntiPigeonCertificate(ctx, cert); err != nil {
				t.Fatal(err)
			}
			change := bayes.ChangePolicy{Hazard: .000001, Threshold: 1, MaxRun: 64, Working: bayes.DefaultWorkingPolicy(), EvidenceTrust: "fixture"}
			group := bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 8, MaxMembers: 16, DisableAutoRevision: true}
			var result store.BayesianOutcomeResult
			for i := 0; i < 25; i++ {
				event, useful := "a", true
				if i%2 == 1 || i == 24 {
					event, useful = "b", false
				}
				if i == 24 {
					group.DisableAutoRevision = false
					change.Threshold = 1e-12
				}
				r := model.BayesianOutcomeRequest{IdempotencyKey: fmt.Sprint(i), TenantID: "tenant-a", EventID: event, Useful: useful, AvailableAt: time.Now().UTC(), Source: model.OutcomeFullStream}
				var err error
				result, err = backend.ApplyBayesianOutcome(ctx, r, "ap:working-group", "", r.IdempotencyKey, 1, change, group, model.ResidualObservation{ActionKey: r.IdempotencyKey, GeneralKey: "working", CommittedProbability: .5, Useful: useful, AvailableAt: r.AvailableAt}, residual.Policy{})
				if err != nil {
					t.Fatal(err)
				}
			}
			expected := bayes.UpdateWorking(nil, false, 1, true, change.Working)
			if result.Revision.Action != model.BayesianRevisionSplitReset || result.Posterior.WorkingBelief == nil || *result.Posterior.WorkingBelief != *expected || result.Posterior.EffectiveSupport != 1 {
				t.Fatalf("split reset lost or duplicated revealing outcome: %+v", result)
			}
			sibling, err := backend.GetBayesianPosterior(ctx, "tenant-a", "a")
			if err != nil || sibling.WorkingBelief != nil || sibling.EvidenceTrust != "fixture" || sibling.EffectiveSupport != 12 {
				t.Fatalf("sibling incorrectly inherited/reset evidence: %+v %v", sibling, err)
			}
		})
	}
}

// Complete internal outcome calls: signatures are produced by the external
// observer before timing; signature verification and persistence stay inside.
func BenchmarkAuthenticatedOutcome(b *testing.B) {
	for _, persistent := range []bool{false, true} {
		for _, enabled := range []bool{false, true} {
			b.Run(fmt.Sprintf("disk=%t/upgrade=%t", persistent, enabled), func(b *testing.B) {
				cfg, key := evidenceConfig(b, enabled)
				if !enabled {
					key = nil
				}
				var backend store.EventStore = memorystore.New()
				if persistent {
					db, err := libravdbstore.Open(libravdbstore.Config{Path: b.TempDir() + "/bench.libravdb", Dimension: 32, Quantization: "none", MemoryMapping: true, EmbeddingModel: "test-hash:d32"})
					if err != nil {
						b.Fatal(err)
					}
					backend = db
				}
				runtime := evidenceRuntime(b, backend, cfg)
				defer runtime.Close()
				seedEvidence(b, runtime, 50)
				packet := evidenceRecall(b, runtime)
				requests := make([]model.BayesianOutcomeRequest, b.N)
				for i := range requests {
					requests[i] = signedFeedback(b, packet, fmt.Sprint(i), key)
				}
				durations := make([]time.Duration, b.N)
				b.ReportAllocs()
				b.ResetTimer()
				for i, r := range requests {
					start := time.Now()
					if _, err := runtime.ObserveBayesianOutcome(context.Background(), r); err != nil {
						b.Fatal(err)
					}
					durations[i] = time.Since(start)
				}
				b.StopTimer()
				sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
				b.ReportMetric(float64(durations[(len(durations)-1)*99/100].Nanoseconds()), "p99-ns")
			})
		}
	}
}
