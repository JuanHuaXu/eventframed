package epistemic

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func authFixture(t testing.TB, capacity int) (*Verifier, model.BayesianOutcomeRequest, ed25519.PrivateKey, time.Time) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	v, err := NewVerifier([]IssuerKey{{Issuer: "sensor", KeyID: "v1", TenantID: "tenant", Source: model.OutcomeFullStream, PublicKey: pub, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour)}}, capacity)
	if err != nil {
		t.Fatal(err)
	}
	r := model.BayesianOutcomeRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: "retry", TenantID: "tenant", EventID: "event", JournalID: "journal", Useful: true, ObservedAt: now, AvailableAt: now, Source: model.OutcomeFullStream, InclusionProbability: 1, Attestation: &model.EvidenceAttestation{Issuer: "sensor", KeyID: "v1", ObservationID: "observation"}}
	signOutcome(t, &r, priv)
	return v, r, priv, now
}

func signOutcome(t testing.TB, r *model.BayesianOutcomeRequest, key ed25519.PrivateKey) {
	t.Helper()
	b, err := OutcomeSigningBytes(*r)
	if err != nil {
		t.Fatal(err)
	}
	r.Attestation.Signature = ed25519.Sign(key, b)
}

func TestAuthenticatedEvidenceBoundaries(t *testing.T) {
	v, r, priv, now := authFixture(t, 2)
	key, err := v.Verify(r, now)
	if err != nil {
		t.Fatal(err)
	}
	r.IdempotencyKey = "new-transport"
	if got, err := v.Verify(r, now); err != nil || got != key {
		t.Fatalf("retry: %q %v", got, err)
	}
	for name, mutate := range map[string]func(*model.BayesianOutcomeRequest){
		"outcome": func(r *model.BayesianOutcomeRequest) { r.Useful = false },
		"tenant":  func(r *model.BayesianOutcomeRequest) { r.TenantID = "other" },
		"source":  func(r *model.BayesianOutcomeRequest) { r.Source = model.OutcomeIndependentAudit },
		"journal": func(r *model.BayesianOutcomeRequest) { r.JournalID = "other" },
		"signals": func(r *model.BayesianOutcomeRequest) { r.Signals.Correction = true },
		"weight":  func(r *model.BayesianOutcomeRequest) { r.InclusionProbability = .1 },
	} {
		t.Run(name, func(t *testing.T) {
			copy := r
			mutate(&copy)
			if _, err := v.Verify(copy, now); err == nil {
				t.Fatal("accepted modified claim")
			}
		})
	}
	if _, err := v.Verify(r, now.Add(2*time.Hour)); err == nil {
		t.Fatal("cached verification bypassed expiry")
	}
	r.Attestation.Parents = []string{"parent"}
	signOutcome(t, &r, priv)
	if _, err := v.Verify(r, now); err == nil {
		t.Fatal("derived claim counted independently")
	}
	r.Attestation.Parents = nil
	for i := 0; i < 10; i++ {
		r.Attestation.ObservationID = fmt.Sprint(i)
		signOutcome(t, &r, priv)
		if _, err := v.Verify(r, now); err != nil {
			t.Fatal(err)
		}
	}
	if len(v.cache) != 2 || v.lru.Len() != 2 {
		t.Fatal("unbounded cache")
	}
	r.Signals.Correction = true
	r.Useful = true
	signOutcome(t, &r, priv)
	r.Useful = false
	if _, err := v.Verify(r, now); err != nil {
		t.Fatalf("normalized durable evidence lost signature: %v", err)
	}
}

func BenchmarkEvidenceVerify(b *testing.B) {
	for _, cached := range []bool{false, true} {
		b.Run(fmt.Sprintf("cached=%t", cached), func(b *testing.B) {
			v, r, _, now := authFixture(b, 16)
			if cached {
				if _, err := v.Verify(r, now); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if !cached {
					v.mu.Lock()
					clear(v.cache)
					v.lru.Init()
					v.mu.Unlock()
				}
				if _, err := v.Verify(r, now); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestIssuerRotationAndRevocation(t *testing.T) {
	v, r, _, now := authFixture(t, 4)
	old, err := v.Verify(r, now)
	if err != nil {
		t.Fatal(err)
	}
	var enrolled IssuerKey
	for _, k := range v.keys {
		enrolled = k
	}
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	rotated := enrolled
	rotated.KeyID = "v2"
	rotated.PublicKey = pub
	next, err := NewVerifier([]IssuerKey{enrolled, rotated}, 4)
	if err != nil {
		t.Fatal(err)
	}
	r.Attestation.KeyID = "v2"
	signOutcome(t, &r, priv)
	got, err := next.Verify(r, now)
	if err != nil || got != old {
		t.Fatalf("rotation minted new observation identity: %v", err)
	}
	rotated.Revoked = true
	revoked, err := NewVerifier([]IssuerKey{enrolled, rotated}, 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := revoked.Verify(r, now); err == nil {
		t.Fatal("revoked key accepted")
	}
	if revoked.Fingerprint() == next.Fingerprint() {
		t.Fatal("revocation did not change policy")
	}
}
