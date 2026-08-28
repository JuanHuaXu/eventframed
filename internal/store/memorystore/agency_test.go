package memorystore_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/agency"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestExpiredProposalCannotResolveAsApproved(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	runtime := memorystore.New()
	record := agencyRecord(t, now, "proposal-expired", now.Add(time.Minute))
	if _, err := runtime.Put(ctx, testutil.Event("event-a", "evidence", now), make([]float32, 8), "event-digest"); err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.PutAgencyProposal(ctx, record, "proposal-digest", 8, 1000, now); err != nil {
		t.Fatal(err)
	}
	claimed, _, err := runtime.ClaimAgencyProposals(ctx, "tenant-a", "authority-a", now, 1, 2*time.Minute)
	if err != nil || len(claimed) != 1 {
		t.Fatalf("claim = %+v, %v", claimed, err)
	}
	request := model.ResolveAgencyProposalRequest{TenantID: "tenant-a", ProposalID: record.Proposal.ID, ConsumerID: "authority-a", Decision: model.AgencyApproved, Reason: "authorized", ExecutionRef: "job-1"}
	if _, err := runtime.ResolveAgencyProposal(ctx, request, record.Proposal.ExpiresAt); !errors.Is(err, store.ErrAgencyExpired) {
		t.Fatalf("expired approval error = %v", err)
	}
	remaining, _, err := runtime.ClaimAgencyProposals(ctx, "tenant-a", "authority-b", record.Proposal.ExpiresAt, 1, time.Minute)
	if err != nil || len(remaining) != 0 {
		t.Fatalf("expired proposal became claimable: %+v, %v", remaining, err)
	}
}

func TestLeaseIsExpiredAtItsExactBoundary(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	runtime := memorystore.New()
	record := agencyRecord(t, now, "proposal-lease", now.Add(time.Hour))
	if _, err := runtime.Put(ctx, testutil.Event("event-a", "evidence", now), make([]float32, 8), "event-digest"); err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.PutAgencyProposal(ctx, record, "proposal-digest", 8, 1000, now); err != nil {
		t.Fatal(err)
	}
	lease := time.Minute
	if _, _, err := runtime.ClaimAgencyProposals(ctx, "tenant-a", "authority-a", now, 1, lease); err != nil {
		t.Fatal(err)
	}
	request := model.ResolveAgencyProposalRequest{TenantID: "tenant-a", ProposalID: record.Proposal.ID, ConsumerID: "authority-a", Decision: model.AgencyApproved, Reason: "authorized", ExecutionRef: "job-1"}
	if _, err := runtime.ResolveAgencyProposal(ctx, request, now.Add(lease)); !errors.Is(err, store.ErrAgencyLease) {
		t.Fatalf("lease-boundary resolution error = %v", err)
	}
}

func TestExpiredPendingProposalDoesNotConsumeTenantQueueBudget(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	runtime := memorystore.New()
	if _, err := runtime.Put(ctx, testutil.Event("event-a", "evidence", now), make([]float32, 8), "event-digest"); err != nil {
		t.Fatal(err)
	}
	first := agencyRecord(t, now, "proposal-old", now.Add(time.Minute))
	if _, err := runtime.PutAgencyProposal(ctx, first, "digest-old", 8, 1, now); err != nil {
		t.Fatal(err)
	}
	second := agencyRecord(t, now.Add(2*time.Minute), "proposal-new", now.Add(time.Hour))
	if _, err := runtime.PutAgencyProposal(ctx, second, "digest-new", 8, 1, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("expired proposal blocked queue admission: %v", err)
	}
}

func agencyRecord(t *testing.T, now time.Time, id string, expiresAt time.Time) model.AgencyProposalRecord {
	t.Helper()
	proposal, err := agency.BuildProposal(model.AgencyProposalDraft{
		ID: id, TenantID: "tenant-a", SessionID: "openclaw:session-a", Action: model.AgencyNotify,
		Reason: "A follow-up became timely.", EvidenceIDs: []string{"event-a"}, ExpectedUtility: .8, Priority: .7,
		NotBefore: now, ExpiresAt: expiresAt, IdempotencyKey: id, CausalChainID: "chain-" + id,
	}, now, agency.DefaultPolicy(true))
	if err != nil {
		t.Fatal(err)
	}
	signer, err := agency.NewSignerForTest()
	if err != nil {
		t.Fatal(err)
	}
	signed, err := signer.Sign(proposal)
	if err != nil {
		t.Fatal(err)
	}
	return model.AgencyProposalRecord{Proposal: proposal, Signed: signed, Status: model.AgencyPending}
}
