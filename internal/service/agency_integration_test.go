package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/agency"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

const testAuthorityToken = "test-authority-token-that-is-at-least-32-bytes"

func TestAgencyIssueClaimResolveAndCausalBudget(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	embedder, _ := embed.NewHashEmbedder(8)
	signer, err := agency.NewSignerForTest()
	if err != nil {
		t.Fatal(err)
	}
	policy := agency.DefaultPolicy(true)
	policy.MaxProposalsPerChain = 2
	runtime, err := service.New(memory, embedder, service.Config{DefaultRecallK: 10, DefaultPackK: 10, DefaultTokenBudget: 1000, AgencyPolicy: policy, AgencySigner: signer, AgencyIssuerToken: "test-issuer-token-that-is-at-least-32-bytes", AgencyAuthorityToken: testAuthorityToken})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	observe(t, runtime, testutil.Event("event-a", "agency evidence", now.Add(-time.Second)))
	root := agencyRequest("proposal-root", "chain-a", "", 0, now)
	badToken := root
	badToken.IssuerToken = "wrong-token-that-is-still-at-least-32-bytes"
	if _, err := runtime.IssueAgencyProposal(ctx, badToken); err == nil {
		t.Fatal("proposal with an invalid issuer token was accepted")
	}
	missingEvidence := agencyRequest("proposal-missing", "chain-missing", "", 0, now)
	missingEvidence.Proposal.EvidenceIDs = []string{"event-missing"}
	if _, err := runtime.IssueAgencyProposal(ctx, missingEvidence); !errors.Is(err, store.ErrAgencyEvidence) {
		t.Fatalf("missing evidence error = %v", err)
	}
	issued, err := runtime.IssueAgencyProposal(ctx, root)
	if err != nil || issued.Duplicate || issued.Record.Status != model.AgencyPending || issued.Snapshot.AgencyVersion != 2 {
		t.Fatalf("issued = %+v, %v", issued, err)
	}
	if _, err := runtime.ClaimAgencyProposals(ctx, model.ClaimAgencyProposalsRequest{ProtocolVersion: model.ProtocolVersion, AuthorityToken: testAuthorityToken, TenantID: "tenant-a", ConsumerID: strings.Repeat("x", 257), Limit: 1}); err == nil {
		t.Fatal("oversized authority consumer id was accepted")
	}
	duplicate, err := runtime.IssueAgencyProposal(ctx, root)
	if err != nil || !duplicate.Duplicate || duplicate.Record.Signed != issued.Record.Signed {
		t.Fatalf("duplicate = %+v, %v", duplicate, err)
	}
	child := agencyRequest("proposal-child", "chain-a", "proposal-root", 1, now)
	child.Proposal.Priority = .6
	if _, err := runtime.IssueAgencyProposal(ctx, child); err != nil {
		t.Fatal(err)
	}
	grandchild := agencyRequest("proposal-grandchild", "chain-a", "proposal-child", 2, now)
	if _, err := runtime.IssueAgencyProposal(ctx, grandchild); !errors.Is(err, store.ErrAgencyChainBudget) {
		t.Fatalf("chain budget error = %v", err)
	}

	if _, err := runtime.ClaimAgencyProposals(ctx, model.ClaimAgencyProposalsRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", ConsumerID: "authority-a", Limit: 1}); err == nil {
		t.Fatal("claim without authority token was accepted")
	}
	claimed, err := runtime.ClaimAgencyProposals(ctx, model.ClaimAgencyProposalsRequest{ProtocolVersion: model.ProtocolVersion, AuthorityToken: testAuthorityToken, TenantID: "tenant-a", ConsumerID: "authority-a", Limit: 1})
	if err != nil || len(claimed.Records) != 1 || claimed.Records[0].Proposal.ID != "proposal-root" || claimed.Records[0].Status != model.AgencyClaimed {
		t.Fatalf("claimed = %+v, %v", claimed, err)
	}
	secondClaim, err := runtime.ClaimAgencyProposals(ctx, model.ClaimAgencyProposalsRequest{ProtocolVersion: model.ProtocolVersion, AuthorityToken: testAuthorityToken, TenantID: "tenant-a", ConsumerID: "authority-b", Limit: 1})
	if err != nil || len(secondClaim.Records) != 1 || secondClaim.Records[0].Proposal.ID != "proposal-child" {
		t.Fatalf("second claim = %+v, %v", secondClaim, err)
	}
	resolution := model.ResolveAgencyProposalRequest{ProtocolVersion: model.ProtocolVersion, AuthorityToken: testAuthorityToken, TenantID: "tenant-a", ProposalID: "proposal-root", ConsumerID: "authority-b", Decision: model.AgencyApproved, Reason: "authorized", ExecutionRef: "job-1"}
	unauthenticatedResolution := resolution
	unauthenticatedResolution.AuthorityToken = "wrong-authority-token-that-is-at-least-32-bytes"
	if _, err := runtime.ResolveAgencyProposal(ctx, unauthenticatedResolution); err == nil {
		t.Fatal("resolution with wrong authority token was accepted")
	}
	if _, err := runtime.ResolveAgencyProposal(ctx, resolution); !errors.Is(err, store.ErrAgencyLease) {
		t.Fatalf("rival resolution error = %v", err)
	}
	resolution.ConsumerID = "authority-a"
	withoutExecution := resolution
	withoutExecution.ExecutionRef = ""
	if _, err := runtime.ResolveAgencyProposal(ctx, withoutExecution); err == nil {
		t.Fatal("approval without execution_ref was accepted")
	}
	approved, err := runtime.ResolveAgencyProposal(ctx, resolution)
	if err != nil || approved.Record.Status != model.AgencyApproved || approved.Record.ExecutionRef != "job-1" {
		t.Fatalf("approved = %+v, %v", approved, err)
	}
	again, err := runtime.ResolveAgencyProposal(ctx, resolution)
	if err != nil || !again.Duplicate {
		t.Fatalf("duplicate resolution = %+v, %v", again, err)
	}
	resolution.Decision = model.AgencyRejected
	resolution.ExecutionRef = ""
	if _, err := runtime.ResolveAgencyProposal(ctx, resolution); !errors.Is(err, store.ErrAgencyConflict) {
		t.Fatalf("terminal conflict = %v", err)
	}
	beforeDelete := memory.Snapshot(ctx)
	deleted, err := runtime.Delete(ctx, model.DeleteRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", EventID: "event-a"})
	if err != nil || !deleted.Deleted || deleted.Snapshot.AgencyVersion != beforeDelete.AgencyVersion+1 {
		t.Fatalf("agency-aware delete = %+v, %v", deleted, err)
	}
	afterDelete, err := runtime.ClaimAgencyProposals(ctx, model.ClaimAgencyProposalsRequest{ProtocolVersion: model.ProtocolVersion, AuthorityToken: testAuthorityToken, TenantID: "tenant-a", ConsumerID: "authority-c", Limit: 10})
	if err != nil || len(afterDelete.Records) != 0 {
		t.Fatalf("deleted evidence left claimable proposals: %+v, %v", afterDelete, err)
	}
}

func agencyRequest(id, chainID, parentID string, depth int, now time.Time) model.IssueAgencyProposalRequest {
	return model.IssueAgencyProposalRequest{ProtocolVersion: model.ProtocolVersion, IssuerToken: "test-issuer-token-that-is-at-least-32-bytes", Proposal: model.AgencyProposalDraft{
		ID: id, TenantID: "tenant-a", SessionID: "openclaw:session-a", Action: model.AgencyNotify,
		Reason: "A useful follow-up became timely.", EvidenceIDs: []string{"event-a"}, ExpectedUtility: .8, Priority: .7,
		NotBefore: now, ExpiresAt: now.Add(time.Hour), IdempotencyKey: id, CausalChainID: chainID, ParentProposalID: parentID, CausalChainDepth: depth,
	}}
}
