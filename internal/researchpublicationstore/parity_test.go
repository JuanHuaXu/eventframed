package researchpublicationstore

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/agency"
	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestCertificateCompositionAndMaintenanceParity(t *testing.T) {
	for _, persistent := range []bool{false, true} {
		name := "memory"
		if persistent {
			name = "persistent"
		}
		t.Run(name, func(t *testing.T) {
			makeBackend := func() store.EventStore {
				if !persistent {
					return memorystore.New()
				}
				db, e := libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/parity.libravdb", Dimension: 4, EmbeddingModel: "fixture", Quantization: "none", MemoryMapping: true})
				if e != nil {
					t.Fatal(e)
				}
				return db
			}
			plain := makeBackend()
			defer plain.Close()
			wrapped, e := New(makeBackend())
			if e != nil {
				t.Fatal(e)
			}
			defer wrapped.Close()
			ctx := context.Background()
			now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
			run := func(name string, operation func(store.EventStore) (any, error)) {
				t.Helper()
				before := wrapped.Snapshot(ctx)
				want, e := operation(plain)
				if e != nil {
					t.Fatalf("control %s: %v", name, e)
				}
				got, e := operation(wrapped)
				if e != nil {
					t.Fatalf("adapter %s: %v", name, e)
				}
				if !reflect.DeepEqual(want, got) || wrapped.Snapshot(ctx) != plain.Snapshot(ctx) {
					t.Fatalf("%s changed result or version: %#v / %#v", name, want, got)
				}
				after := wrapped.Snapshot(ctx)
				if !wrapped.ResearchPublicationCompatible(ctx, after, now) {
					t.Fatalf("%s did not publish current version", name)
				}
				if before != after && wrapped.ResearchPublicationCompatible(ctx, before, now) {
					t.Fatalf("%s retained invalid old proof", name)
				}
			}
			run("policy", func(db store.EventStore) (any, error) { return db.BindBayesianPolicy(ctx, "policy-v1") })
			run("policy-noop", func(db store.EventStore) (any, error) { return db.BindBayesianPolicy(ctx, "policy-v1") })
			for _, id := range []string{"a", "b"} {
				run("seed-"+id, func(db store.EventStore) (any, error) {
					return db.Put(ctx, testutil.Event(id, "public stage", now.Add(-time.Minute)), []float32{1, 0, 0, 0}, id)
				})
			}
			run("selection", func(db store.EventStore) (any, error) {
				s := db.Snapshot(ctx)
				return db.PublishSelectionCertificate(ctx, model.SelectionSupportCertificate{ID: "selection", TenantID: "tenant-a", PolicyVersion: s.PolicyVersion, EvidenceEpoch: s.EvidenceEpoch, MinSelectionProbability: .2, Procedure: "fixture", ValidUntil: now.Add(time.Hour)})
			})
			run("selection-read", func(db store.EventStore) (any, error) { return db.GetSelectionCertificate(ctx, "tenant-a") })
			run("omission", func(db store.EventStore) (any, error) {
				s := db.Snapshot(ctx)
				return db.PublishOmittedInfluenceCertificate(ctx, model.OmittedInfluenceCertificate{ID: "omission", TenantID: "tenant-a", PolicyVersion: s.PolicyVersion, EvidenceEpoch: s.EvidenceEpoch, DivergenceUCB: .1, DivergenceLimit: .2, Procedure: "fixture", ValidUntil: now.Add(time.Hour)})
			})
			run("omission-read", func(db store.EventStore) (any, error) { return db.GetOmittedInfluenceCertificate(ctx, "tenant-a") })
			ap := func(db store.EventStore) (any, error) {
				s := db.Snapshot(ctx)
				return db.PublishAntiPigeonCertificate(ctx, model.AntiPigeonCertificate{ID: "ap", TenantID: "tenant-a", MemberEventIDs: []string{"a", "b"}, GraphVersion: s.GraphVersion, EvidenceEpoch: s.EvidenceEpoch})
			}
			run("anti-pigeon", ap)
			run("anti-pigeon-noop", ap)
			run("anti-pigeon-read", func(db store.EventStore) (any, error) {
				return db.GetAntiPigeonCertificate(ctx, "tenant-a", []string{"a", "b"})
			})
			run("composition", func(db store.EventStore) (any, error) {
				s := db.Snapshot(ctx)
				ev := testutil.Event("macro", "public mission", now)
				ev.Kind = model.HigherOrderEventKind
				ev.Provenance.SourceEventIDs = []string{"a", "b"}
				ev.Composition = &model.Composition{MemberEventIDs: []string{"a", "b"}, RepresentativeEventID: "a", RuleID: "stages", Resolution: "mission", Confidence: .9, AntiPigeonCertificateID: "ap", EvidenceEpoch: s.EvidenceEpoch}
				return db.PutComposition(ctx, ev, []float32{1, 0, 0, 0}, "macro", s)
			})
			run("composition-read", func(db store.EventStore) (any, error) { return db.GetEvents(ctx, "tenant-a", []string{"macro"}, now) })
			run("decompose", func(db store.EventStore) (any, error) {
				return db.DeleteComposition(ctx, "tenant-a", "macro", "fixture", now.Add(time.Second))
			})
			run("tombstone", func(db store.EventStore) (any, error) { return db.GetCompositionTombstone(ctx, "tenant-a", "macro") })
			run("compact", func(db store.EventStore) (any, error) { return nil, db.Compact(ctx) })
			// One signed fixture is reused by both stores; no executor or agent is
			// connected. The test compares persistence and publication semantics.
			signer, e := agency.NewSignerForTest()
			if e != nil {
				t.Fatal(e)
			}
			proposal, e := agency.BuildProposal(model.AgencyProposalDraft{
				ID: "proposal", TenantID: "tenant-a", SessionID: "fixture-session", Action: model.AgencyWake,
				Reason: "public fixture follow-up", EvidenceIDs: []string{"a"}, ExpectedUtility: .8, Priority: .7,
				NotBefore: now, ExpiresAt: now.Add(time.Hour), IdempotencyKey: "proposal", CausalChainID: "fixture-chain",
			}, now, agency.DefaultPolicy(true))
			if e != nil {
				t.Fatal(e)
			}
			signed, e := signer.Sign(proposal)
			if e != nil {
				t.Fatal(e)
			}
			record := model.AgencyProposalRecord{Proposal: proposal, Signed: signed, Status: model.AgencyPending}
			expectDuplicate := false
			putProposal := func(db store.EventStore) (any, error) {
				r, e := db.PutAgencyProposal(ctx, record, "proposal-digest", 8, 1000, now)
				if e == nil && (r.Record.Status != model.AgencyPending || r.Duplicate != expectDuplicate) {
					t.Fatal("wrong put branch", r)
				}
				return r, e
			}
			run("agency-put", putProposal)
			expectDuplicate = true
			run("agency-put-duplicate", putProposal)
			expectClaim := true
			claim := func(db store.EventStore) (any, error) {
				records, snapshot, e := db.ClaimAgencyProposals(ctx, "tenant-a", "fixture-authority", now.Add(time.Second), 10, time.Minute)
				if e == nil {
					if expectClaim && (len(records) != 1 || records[0].Status != model.AgencyClaimed || records[0].ClaimedBy != "fixture-authority") {
						t.Fatal("did not claim fixture", records)
					}
					if !expectClaim && len(records) != 0 {
						t.Fatal("leased proposal reclaimed", records)
					}
				}
				return struct {
					Records  []model.AgencyProposalRecord
					Snapshot model.Snapshot
				}{records, snapshot}, e
			}
			run("agency-claim", claim)
			expectClaim = false
			run("agency-claim-leased-noop", claim)
			expectDuplicate = false
			resolve := func(db store.EventStore) (any, error) {
				r, e := db.ResolveAgencyProposal(ctx, model.ResolveAgencyProposalRequest{TenantID: "tenant-a", ProposalID: "proposal", ConsumerID: "fixture-authority", Decision: model.AgencyApproved, Reason: "fixture only", ExecutionRef: "no-execution"}, now.Add(2*time.Second))
				if e == nil && (r.Record.Status != model.AgencyApproved || r.Duplicate != expectDuplicate || r.Record.ExecutionRef != "no-execution") {
					t.Fatal("wrong resolution branch", r)
				}
				return r, e
			}
			run("agency-resolve", resolve)
			expectDuplicate = true
			run("agency-resolve-duplicate", resolve)
			outcome := func(db store.EventStore) (any, error) {
				request := model.BayesianOutcomeRequest{IdempotencyKey: "fixture-outcome", TenantID: "tenant-a", EventID: "a", Useful: true, AvailableAt: now}
				observation := model.ResidualObservation{ActionKey: "fixture-action", GeneralKey: "fixture-general", HorizonKey: model.RetrievalUsefulnessHorizon, BaseProbability: .5, CommittedProbability: .5, Useful: true, ValidationEligible: true, EventID: "a", JournalID: "fixture-journal", PosteriorKey: "fixture-posterior", AvailableAt: now}
				return db.ApplyBayesianOutcome(ctx, request, "fixture-posterior", "", "fixture-digest", 1, bayes.ChangePolicy{Hazard: .05, Threshold: .3, MaxRun: 64}, bayes.GroupPolicy{}, observation, residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .55, ConfidenceDelta: .05, MotionLimit: .1, MaxAge: time.Hour, ImprovementDelta: .001})
			}
			run("posterior-update", outcome)
			run("posterior-duplicate", outcome)
			run("posterior-read", func(db store.EventStore) (any, error) {
				return db.GetBayesianPosterior(ctx, "tenant-a", "fixture-posterior")
			})
			type graphResult struct {
				Graph    model.PredictiveGraph
				Snapshot model.Snapshot
			}
			run("snap-publish", func(db store.EventStore) (any, error) {
				previous, e := db.GetPredictiveGraph(ctx, "tenant-a")
				if e != nil {
					return nil, e
				}
				published := model.PredictiveGraph{TenantID: "tenant-a", SourceSnapID: "fixture-snap", PublishedAt: now, Nodes: []model.CompatibilityNode{{ID: "fixture-bucket", Kind: "bucket", MemberEventIDs: []string{"a"}, PosteriorKeys: []string{"fixture-posterior"}, LawSpace: model.RetrievalUsefulnessHorizon}}}
				record := model.PredictiveSnapRecord{ID: "fixture-snap", TenantID: "tenant-a", PreviousGraph: previous, PublishedGraph: published, Closure: model.DependencyClosure{NodeIDs: []string{"fixture-bucket"}, EventIDs: []string{"a"}, PosteriorKeys: []string{"fixture-posterior"}}, SimultaneousCoverage: .95, Procedure: "fixture", Issuer: "fixture-auditor", PublishedAt: now}
				g, s, e := db.PublishPredictiveSnap(ctx, record)
				if e == nil && (g.SourceSnapID != "fixture-snap" || len(g.Nodes) != 1 || s.GraphVersion != previous.Version+1) {
					t.Fatal("snap not published", g, s)
				}
				return graphResult{g, s}, e
			})
			run("snap-invalidates-residual", func(db store.EventStore) (any, error) {
				r, e := db.GetResidualCandidates(ctx, "tenant-a", "fixture-action", "fixture-general")
				if e == nil && (r.Exact == nil || r.Exact.Active || r.General == nil || r.General.Active) {
					t.Fatal("residual invalidation missing", r)
				}
				return r, e
			})
			run("snap-invalidates-posterior", func(db store.EventStore) (any, error) {
				p, e := db.GetBayesianPosterior(ctx, "tenant-a", "fixture-posterior")
				if e == nil && p.Certified {
					t.Fatal("posterior remains certified", p)
				}
				return p, e
			})
			run("snap-rollback", func(db store.EventStore) (any, error) {
				before := db.Snapshot(ctx)
				started := time.Now()
				g, s, e := db.RollbackPredictiveSnap(ctx, "tenant-a", "fixture-snap", "fixture rollback")
				finished := time.Now()
				if e == nil && (len(g.Nodes) != 0 || g.SourceSnapID != "rollback:fixture-snap" || s.GraphVersion != before.GraphVersion+1) {
					t.Fatal("rollback missing", g, s)
				}
				if e == nil {
					stored, readErr := db.GetPredictiveGraph(ctx, "tenant-a")
					if readErr != nil || !reflect.DeepEqual(stored, g) || g.PublishedAt.Before(started) || g.PublishedAt.After(finished) {
						t.Fatal("rollback timestamp or persistence invalid", g, readErr)
					}
					// Independent backend calls generate distinct wall-clock times.
					// Validate each above, then normalize only this nondeterministic field.
					g.PublishedAt = time.Time{}
				}
				return graphResult{g, s}, e
			})
		})
	}
}
