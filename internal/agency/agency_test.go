package agency

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestBuildSignAndVerifyProposal(t *testing.T) {
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	scheduled := now.Add(time.Hour)
	proposal, err := BuildProposal(model.AgencyProposalDraft{
		ID: "proposal-1", TenantID: "tenant-a", SessionID: "openclaw:session-a", Action: model.AgencySchedule,
		Reason: "A follow-up became timely.", EvidenceIDs: []string{"event-b", "event-a"}, ExpectedUtility: .8, Priority: .7,
		NotBefore: now, ScheduledFor: &scheduled, ExpiresAt: now.Add(2 * time.Hour), IdempotencyKey: "proposal-1", CausalChainID: "chain-1",
	}, now, DefaultPolicy(true))
	if err != nil {
		t.Fatal(err)
	}
	if proposal.RequiredCapability != CapabilitySchedule || proposal.EvidenceIDs[0] != "event-a" || proposal.ContractVersion != model.ContractVersion {
		t.Fatalf("proposal = %+v", proposal)
	}
	signer, err := NewSignerForTest()
	if err != nil {
		t.Fatal(err)
	}
	signed, err := signer.Sign(proposal)
	if err != nil {
		t.Fatal(err)
	}
	verified, err := signer.Verify(signed)
	if err != nil || verified.ID != proposal.ID || verified.RequiredCapability != CapabilitySchedule {
		t.Fatalf("verified = %+v, %v", verified, err)
	}
	signed.Signature = "invalid"
	if _, err := signer.Verify(signed); err == nil {
		t.Fatal("expected signature rejection")
	}
}

func TestProposalValidationRejectsUnsafeShapes(t *testing.T) {
	now := time.Now().UTC()
	base := model.AgencyProposalDraft{
		ID: "proposal-1", TenantID: "tenant-a", SessionID: "session-a", Action: model.AgencyWake,
		Reason: "reason", EvidenceIDs: []string{"event-a"}, ExpectedUtility: .5, Priority: .5,
		NotBefore: now, ExpiresAt: now.Add(time.Hour), IdempotencyKey: "proposal-1", CausalChainID: "chain-1",
	}
	for name, mutate := range map[string]func(*model.AgencyProposalDraft){
		"unsupported action": func(value *model.AgencyProposalDraft) { value.Action = model.AgencyRemember },
		"duplicate evidence": func(value *model.AgencyProposalDraft) { value.EvidenceIDs = []string{"event-a", "event-a"} },
		"missing parent":     func(value *model.AgencyProposalDraft) { value.CausalChainDepth = 1 },
		"schedule on wake": func(value *model.AgencyProposalDraft) {
			when := now.Add(time.Minute)
			value.ScheduledFor = &when
		},
		"schedule at expiry": func(value *model.AgencyProposalDraft) {
			value.Action = model.AgencySchedule
			when := value.ExpiresAt
			value.ScheduledFor = &when
		},
		"oversized session": func(value *model.AgencyProposalDraft) { value.SessionID = strings.Repeat("s", maxSessionIDBytes+1) },
		"oversized evidence": func(value *model.AgencyProposalDraft) {
			value.EvidenceIDs = []string{strings.Repeat("e", maxIdentifierBytes+1)}
		},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := base
			mutate(&candidate)
			if _, err := BuildProposal(candidate, now, DefaultPolicy(true)); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestProposalDraftRemainsValidForIdempotentRetryUntilExpiry(t *testing.T) {
	issuedAt := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	draft := model.AgencyProposalDraft{
		ID: "proposal-retry", TenantID: "tenant-a", SessionID: "session-a", Action: model.AgencyWake,
		Reason: "reason", EvidenceIDs: []string{"event-a"}, ExpectedUtility: .5, Priority: .5,
		NotBefore: issuedAt, ExpiresAt: issuedAt.Add(time.Hour), IdempotencyKey: "proposal-retry", CausalChainID: "chain-retry",
	}
	if _, err := BuildProposal(draft, issuedAt.Add(30*time.Minute), DefaultPolicy(true)); err != nil {
		t.Fatalf("valid retry was rejected: %v", err)
	}
	if _, err := BuildProposal(draft, draft.ExpiresAt, DefaultPolicy(true)); err == nil {
		t.Fatal("expired proposal was accepted")
	}
}

func TestSignerKeyFilesRoundTrip(t *testing.T) {
	root := t.TempDir()
	privatePath, publicPath := filepath.Join(root, "agency.key"), filepath.Join(root, "agency.pub")
	first, err := LoadOrCreateSigner(privatePath, publicPath)
	if err != nil {
		t.Fatal(err)
	}
	second, err := LoadOrCreateSigner(privatePath, publicPath)
	if err != nil || first.keyID != second.keyID {
		t.Fatalf("reloaded signer = %+v, %v", second, err)
	}
	info, err := os.Stat(privatePath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("private key mode = %v", info.Mode().Perm())
	}
}

func TestIssuerTokenFileRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "issuer.token")
	first, err := LoadOrCreateIssuerToken(path)
	if err != nil {
		t.Fatal(err)
	}
	second, err := LoadOrCreateIssuerToken(path)
	if err != nil || first != second || len(first) < 32 {
		t.Fatalf("issuer token round trip = %q, %v", second, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("issuer token mode = %v", info.Mode().Perm())
	}
}

func TestAuthorityTokenFileRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "authority.token")
	first, err := LoadOrCreateAuthorityToken(path)
	if err != nil {
		t.Fatal(err)
	}
	second, err := LoadOrCreateAuthorityToken(path)
	if err != nil || first != second || len(first) < 32 {
		t.Fatalf("authority token round trip = %q, %v", second, err)
	}
}

func TestAgencySecretsRejectSymlinksAndDoNotOverwritePublicTarget(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.WriteFile(target, []byte("do-not-overwrite"), 0o600); err != nil {
		t.Fatal(err)
	}
	publicLink := filepath.Join(root, "agency.pub")
	if err := os.Symlink(target, publicLink); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := LoadOrCreateSigner(filepath.Join(root, "agency.key"), publicLink); err == nil {
		t.Fatal("signer accepted a symlink public key path")
	}
	contents, err := os.ReadFile(target)
	if err != nil || string(contents) != "do-not-overwrite" {
		t.Fatalf("public-key target changed to %q, %v", contents, err)
	}
	if _, err := LoadOrCreateAuthorityToken(publicLink); err == nil {
		t.Fatal("authority token accepted a symlink path")
	}
}
