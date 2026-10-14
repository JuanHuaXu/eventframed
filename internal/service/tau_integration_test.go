package service_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/agency"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestTauPlanLifecycleBothBackends(t *testing.T) {
	for _, backend := range []string{"memory", "libravdb"} {
		t.Run(backend, func(t *testing.T) {
			ctx := context.Background()
			var db store.EventStore
			emb, _ := embed.NewHashEmbedder(8)
			if backend == "memory" {
				db = memorystore.New()
			} else {
				disk, err := libravdbstore.Open(libravdbstore.Config{Path: filepath.Join(t.TempDir(), "tau.db"), Dimension: 8, Quantization: "none", EmbeddingModel: emb.ModelKey()})
				if err != nil {
					t.Fatal(err)
				}
				defer disk.Close()
				db = disk
			}
			signer, _ := agency.NewSignerForTest()
			runtime, err := service.New(db, emb, service.Config{DefaultRecallK: 10, DefaultPackK: 10, DefaultTokenBudget: 1000, AgencyPolicy: agency.DefaultPolicy(true), AgencySigner: signer, AgencyIssuerToken: "test-issuer-token-that-is-at-least-32-bytes", AgencyAuthorityToken: testAuthorityToken})
			if err != nil {
				t.Fatal(err)
			}
			now := time.Now().UTC()
			origin := testutil.Event("tau-origin", "Public synthetic task start", now.Add(-time.Second))
			observe(t, runtime, origin)
			request := model.PlanTauTurnRequest{ProtocolVersion: model.ProtocolVersion, IssuerToken: "test-issuer-token-that-is-at-least-32-bytes", TenantID: "tenant-a", SessionID: "session-a", PlanID: "opportunity-1", OriginEventID: origin.ID, EvidenceIDs: []string{origin.ID}, Base: db.Snapshot(ctx), Priority: .5, Reason: "Public synthetic follow-up", Forecast: model.TauForecast{ID: "law-1", Origin: origin.OccurredAt, IssuedAt: now, ValidUntil: origin.OccurredAt.Add(time.Minute), HorizonMS: 60000, Points: []model.TauPoint{{Mark: "complete", DelayMS: 30000, Probability: .9}}, NoEvent: .1}, Policy: model.TauTurnPolicy{TargetKind: "complete", Benefit: 1, FalseCost: .5, TurnCost: .1, MinProbability: .7, Margin: .01, LeadMS: 2000, WindowMS: 10000, SlackMS: 5000, MinDelayMS: 1000, CooldownMS: 60000}}
			request.Forecast.Base = request.Base
			bad := request
			bad.IssuerToken = "bad"
			if _, err := runtime.PlanTauTurn(ctx, bad); err == nil {
				t.Fatal("unauthenticated plan accepted")
			}
			result, err := runtime.PlanTauTurn(ctx, request)
			if err != nil || result.Proposal == nil {
				t.Fatalf("plan=%+v err=%v", result, err)
			}
			// Only agency counters changed: a stable retry returns the same proposal.
			retry, err := runtime.PlanTauTurn(ctx, request)
			if err != nil || retry.Proposal == nil || !retry.Proposal.Duplicate {
				t.Fatalf("retry=%+v %v", retry, err)
			}
			if !store.TauSnapshotCompatible(result.Proposal.Record.Proposal.Tau, db.Snapshot(ctx)) {
				t.Fatal("proposal invalidated itself")
			}
			reader := db.(store.AgencyProposalReader)
			cloned, err := reader.GetAgencyProposal(ctx, "tenant-a", result.Proposal.Record.Proposal.ID)
			if err != nil {
				t.Fatal(err)
			}
			cloned.Proposal.Tau.Forecast.Points[0].Probability = 0
			cloned.Proposal.Tau.Base.RuntimeVersion = 0
			unchanged, err := reader.GetAgencyProposal(ctx, "tenant-a", result.Proposal.Record.Proposal.ID)
			if err != nil || unchanged.Proposal.Tau.Forecast.Points[0].Probability != .9 || unchanged.Proposal.Tau.Base.RuntimeVersion == 0 {
				t.Fatal("returned proposal aliases stored state")
			}
			// A second opportunity for the same target cannot reserve a second turn.
			other := request
			other.PlanID = "opportunity-2"
			if _, err := runtime.PlanTauTurn(ctx, other); !errors.Is(err, store.ErrAgencyChainBudget) {
				t.Fatalf("duplicate opportunity: %v", err)
			}
			p := result.Proposal.Record.Proposal
			rows, _, err := db.ClaimAgencyProposals(ctx, "tenant-a", "authority", p.NotBefore.Add(-time.Nanosecond), 10, time.Minute)
			if err != nil || len(rows) != 0 {
				t.Fatal("tau leased before nominal time")
			}
			rows, _, err = db.ClaimAgencyProposals(ctx, "tenant-a", "authority", p.NotBefore, 10, time.Minute)
			if err != nil || len(rows) != 1 || rows[0].Proposal.Tau.ForecastID != "law-1" {
				t.Fatalf("claim %v %v", rows, err)
			}
			// An observation between lease and approval forces scheduler rollback.
			observe(t, runtime, testutil.Event("tau-new", "New relevant observation", now))
			rebased := request
			rebased.PlanID = "rebased"
			rebased.Base = db.Snapshot(ctx)
			rebased.Forecast.Base = rebased.Base
			if _, err := runtime.PlanTauTurn(ctx, rebased); !errors.Is(err, store.ErrAgencyConflict) {
				t.Fatalf("stale forecast rebase: %v", err)
			}
			_, err = db.ResolveAgencyProposal(ctx, model.ResolveAgencyProposalRequest{TenantID: "tenant-a", ProposalID: p.ID, ConsumerID: "authority", Decision: model.AgencyApproved, Reason: "test", ExecutionRef: "job"}, p.NotBefore.Add(time.Second))
			if !errors.Is(err, store.ErrStaleSnapshot) {
				t.Fatalf("stale approval=%v", err)
			}
			rows, _, err = db.ClaimAgencyProposals(ctx, "tenant-a", "authority", p.NotBefore.Add(time.Second), 10, time.Minute)
			if err != nil || len(rows) != 0 {
				t.Fatal("stale work remained claimable")
			}
			replacement := request
			replacement.PlanID = "replacement"
			replacement.Base = db.Snapshot(ctx)
			replacement.Forecast.Base = replacement.Base
			replacement.Forecast.ID = "new-law"
			replacement.Forecast.HorizonMS = 3600000
			replacement.Forecast.ValidUntil = origin.OccurredAt.Add(time.Hour)
			replacement.Forecast.Points = []model.TauPoint{{Mark: "complete", DelayMS: 1800000, Probability: .9}}
			next, err := runtime.PlanTauTurn(ctx, replacement)
			if err != nil || next.Proposal == nil {
				t.Fatalf("fresh replacement: %+v %v", next, err)
			}
			nextP := next.Proposal.Record.Proposal
			rows, _, err = db.ClaimAgencyProposals(ctx, "tenant-a", "next-worker", nextP.NotBefore, 10, time.Minute)
			if err != nil || len(rows) != 1 {
				t.Fatalf("replacement claim: %v %v", rows, err)
			}
			late := model.ResolveAgencyProposalRequest{TenantID: "tenant-a", ProposalID: p.ID, ConsumerID: "authority", Decision: model.AgencyUncertain, Reason: "cancellation failed after expiry", ExecutionRef: "live-handle"}
			reported, err := db.ResolveAgencyProposal(ctx, late, nextP.NotBefore.Add(time.Second))
			if err != nil || reported.Record.Status != model.AgencyUncertain || reported.Record.ExecutionRef != "live-handle" {
				t.Fatalf("late uncertainty lost: %+v %v", reported, err)
			}
			_, err = db.ResolveAgencyProposal(ctx, model.ResolveAgencyProposalRequest{TenantID: "tenant-a", ProposalID: nextP.ID, ConsumerID: "next-worker", Decision: model.AgencyApproved, Reason: "new approval", ExecutionRef: "new-job"}, nextP.NotBefore.Add(2*time.Second))
			if !errors.Is(err, store.ErrAgencyConflict) {
				t.Fatalf("late uncertainty did not fence approval: %v", err)
			}
			rows, _, err = db.ClaimAgencyProposals(ctx, "tenant-a", "next-worker", nextP.NotBefore.Add(2*time.Second), 10, time.Minute)
			if err != nil || len(rows) != 0 {
				t.Fatal("late uncertainty did not fence claims")
			}
			duplicate, err := db.ResolveAgencyProposal(ctx, late, p.ExpiresAt.Add(time.Minute))
			if err != nil || !duplicate.Duplicate {
				t.Fatal("uncertainty retry is not idempotent")
			}
			approve := late
			approve.Decision = model.AgencyApproved
			if _, err := db.ResolveAgencyProposal(ctx, approve, p.ExpiresAt.Add(time.Minute)); !errors.Is(err, store.ErrAgencyConflict) {
				t.Fatalf("uncertain work reapproved: %v", err)
			}
			late.ExecutionRef = "second-handle"
			reported, err = db.ResolveAgencyProposal(ctx, late, p.ExpiresAt.Add(2*time.Minute))
			if err != nil || len(reported.Record.UncertainExecutionRefs) != 2 {
				t.Fatalf("uncertain handle history: %+v %v", reported, err)
			}
			if !store.TauOpportunityBlocked(result.Proposal.Record, reported.Record, db.Snapshot(ctx), p.ExpiresAt.Add(24*time.Hour)) {
				t.Fatal("uncertainty released reservation")
			}
			clear := late
			clear.Decision = model.AgencyRejected
			clear.ExecutionRef = ""
			clear.Reason = "cancellation acknowledged"
			if _, err := db.ResolveAgencyProposal(ctx, clear, nextP.NotBefore.Add(3*time.Minute)); !errors.Is(err, store.ErrAgencyConflict) {
				t.Fatal("ordinary rejection cleared uncertainty")
			}
			clear.CancelledExecutionRefs = []string{"live-handle"}
			if _, err := db.ResolveAgencyProposal(ctx, clear, nextP.NotBefore.Add(3*time.Minute)); !errors.Is(err, store.ErrAgencyConflict) {
				t.Fatal("stale cancellation set cleared uncertainty")
			}
			clear.CancelledExecutionRefs = []string{"live-handle", "second-handle"}
			if r, err := db.ResolveAgencyProposal(ctx, clear, nextP.NotBefore.Add(3*time.Minute)); err != nil || r.Record.Status != model.AgencyRejected {
				t.Fatalf("complete cancellation acknowledgment: %+v %v", r, err)
			}

			expiry := request
			expiry.PlanID = "expiry-clone"
			expiry.Forecast.ID = "expiry-law"
			expiry.Base = db.Snapshot(ctx)
			expiry.Forecast.Base = expiry.Base
			expiry.Policy.TargetKind = "other-completion"
			expiry.Forecast.Points = []model.TauPoint{{Mark: "other-completion", DelayMS: 30000, Probability: .9}}
			exp, err := runtime.PlanTauTurn(ctx, expiry)
			if err != nil || exp.Proposal == nil {
				t.Fatalf("expiry plan: %+v %v", exp, err)
			}
			expP := exp.Proposal.Record.Proposal
			if rows, _, err := db.ClaimAgencyProposals(ctx, "tenant-a", "expiry-worker", expP.NotBefore, 10, time.Minute); err != nil || len(rows) != 1 {
				t.Fatal("expiry fixture claim failed")
			}
			expResult, err := db.ResolveAgencyProposal(ctx, model.ResolveAgencyProposalRequest{TenantID: "tenant-a", ProposalID: expP.ID, ConsumerID: "expiry-worker", Decision: model.AgencyApproved, Reason: "late", ExecutionRef: "expiry-job"}, expP.ExpiresAt.Add(time.Second))
			if !errors.Is(err, store.ErrAgencyExpired) {
				t.Fatalf("expiry error: %v", err)
			}
			expResult.Record.Proposal.Tau.Forecast.Points[0].Probability = 0
			expResult.Record.Proposal.EvidenceIDs[0] = "tampered"
			persisted, err := reader.GetAgencyProposal(ctx, "tenant-a", expP.ID)
			if err != nil || persisted.Proposal.Tau.Forecast.Points[0].Probability != .9 || persisted.Proposal.EvidenceIDs[0] != origin.ID {
				t.Fatal("expiry error result aliases storage")
			}
		})
	}
}
