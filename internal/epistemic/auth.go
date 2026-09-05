package epistemic

import (
	"container/list"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const AuthenticatedOutcomePrefix = "authenticated-observation:"

// LoadVerifier reads only public keys from an explicitly configured local file.
// Registry replacement is a controlled restart, never a data-plane operation.
func LoadVerifier(path string) (*Verifier, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 4*1024*1024+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 4*1024*1024 {
		return nil, errors.New("evidence registry exceeds 4 MiB")
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var keys []IssuerKey
	if err := decoder.Decode(&keys); err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, errors.New("trailing evidence registry data")
	}
	return NewVerifier(keys, 4096)
}

// IssuerKey is operator-enrolled authority, scoped to one tenant and feedback
// class. A new key for the same issuer does not create a new observation domain.
type IssuerKey struct {
	Issuer    string                      `json:"issuer"`
	KeyID     string                      `json:"key_id"`
	TenantID  string                      `json:"tenant_id"`
	Source    model.BayesianOutcomeSource `json:"source"`
	PublicKey []byte                      `json:"public_key"`
	NotBefore time.Time                   `json:"not_before"`
	NotAfter  time.Time                   `json:"not_after"`
	Revoked   bool                        `json:"revoked"`
}

// Verifier holds an immutable enrolled-key snapshot. Replacing the startup
// registry creates a new policy fingerprint and invalidates dependent state.
// The bounded cache holds only successful signature checks, never admission.
type Verifier struct {
	keys        map[string]IssuerKey
	fingerprint string
	mu          sync.Mutex
	cache       map[[32]byte]*list.Element
	lru         *list.List
	capacity    int
}

func NewVerifier(keys []IssuerKey, capacity int) (*Verifier, error) {
	if len(keys) == 0 || len(keys) > 4096 || capacity < 1 || capacity > 65536 {
		return nil, errors.New("invalid evidence trust/cache capacity")
	}
	copyKeys := append([]IssuerKey(nil), keys...)
	v := &Verifier{keys: make(map[string]IssuerKey), cache: make(map[[32]byte]*list.Element), lru: list.New(), capacity: capacity}
	for i, k := range copyKeys {
		if !validIdentity(k.Issuer) || !validIdentity(k.KeyID) || !validIdentity(k.TenantID) || len(k.PublicKey) != ed25519.PublicKeySize || !k.NotAfter.After(k.NotBefore) || k.NotBefore.IsZero() {
			return nil, errors.New("invalid enrolled evidence key")
		}
		if k.Source != model.OutcomeFullStream && k.Source != model.OutcomeSelected && k.Source != model.OutcomeIndependentAudit {
			return nil, errors.New("invalid evidence source scope")
		}
		key := registryKey(k.Issuer, k.KeyID, k.TenantID, k.Source)
		if _, exists := v.keys[key]; exists {
			return nil, errors.New("duplicate enrolled evidence key")
		}
		k.PublicKey = append([]byte(nil), k.PublicKey...)
		copyKeys[i] = k
		v.keys[key] = k
	}
	sort.Slice(copyKeys, func(i, j int) bool {
		a, b := copyKeys[i], copyKeys[j]
		return registryKey(a.Issuer, a.KeyID, a.TenantID, a.Source) < registryKey(b.Issuer, b.KeyID, b.TenantID, b.Source)
	})
	payload, err := json.Marshal(copyKeys)
	if err != nil {
		return nil, err
	}
	v.fingerprint = digest("evidence-trust-v1\x00" + string(payload))
	return v, nil
}

func registryKey(issuer, key, tenant string, source model.BayesianOutcomeSource) string {
	return strings.Join([]string{issuer, key, tenant, string(source)}, "\x00")
}
func validIdentity(s string) bool {
	return len(s) > 0 && len(s) <= 256 && utf8.ValidString(s) && strings.TrimSpace(s) == s && !strings.ContainsAny(s, "\x00\r\n")
}
func (v *Verifier) Fingerprint() string {
	if v == nil {
		return "legacy-unsigned"
	}
	return v.fingerprint
}

// OutcomeSigningBytes is a versioned JSON array, not arbitrary JSON or JCS.
// Floats use IEEE-754 hex bits and times UTC RFC3339Nano. All fields that affect
// interpretation are bound; transport retry IDs are deliberately excluded.
func OutcomeSigningBytes(r model.BayesianOutcomeRequest) ([]byte, error) {
	a := r.Attestation
	if a == nil || !validIdentity(a.Issuer) || !validIdentity(a.KeyID) || !validIdentity(a.ObservationID) || !validIdentity(r.TenantID) || !validIdentity(r.EventID) || !validIdentity(r.JournalID) || len(a.Parents) > 8 {
		return nil, errors.New("invalid evidence envelope")
	}
	if math.IsNaN(r.InclusionProbability) || math.IsInf(r.InclusionProbability, 0) {
		return nil, errors.New("invalid evidence inclusion probability")
	}
	parents := append([]string{}, a.Parents...)
	sort.Strings(parents)
	for i, p := range parents {
		if !validIdentity(p) || (i > 0 && parents[i-1] == p) || p == a.ObservationID {
			return nil, errors.New("invalid evidence parents")
		}
	}
	// Bind the interpreted outcome so the normalized durable request still
	// verifies when explicit feedback overrides the fallback Useful field.
	useful, _ := r.Signals.Resolve(r.Useful)
	return json.Marshal([]any{"eventframe-outcome-attestation-v1", r.ProtocolVersion, a.Issuer, a.KeyID, a.ObservationID, parents, r.TenantID, r.EventID, r.JournalID, useful, r.Signals.ExplicitUseful, r.Signals.Packed, r.Signals.Cited, r.Signals.SuccessfulDownstream, r.Signals.Correction, r.Signals.Rejected, r.ObservedAt.UTC().Format(time.RFC3339Nano), r.AvailableAt.UTC().Format(time.RFC3339Nano), r.Source, strconv.FormatUint(math.Float64bits(r.InclusionProbability), 16)})
}

// Verify returns a stable durable ledger key. Key rotation and a new journal
// cannot make the same issuer/observation count twice. Derived assertions are
// withheld here; their existence may still be retained as ordinary chat data.
func (v *Verifier) Verify(r model.BayesianOutcomeRequest, now time.Time) (string, error) {
	if v == nil {
		return "", errors.New("evidence verification is not configured")
	}
	payload, err := OutcomeSigningBytes(r)
	if err != nil {
		return "", err
	}
	a := r.Attestation
	k, ok := v.keys[registryKey(a.Issuer, a.KeyID, r.TenantID, r.Source)]
	if !ok || k.Revoked || now.Before(k.NotBefore) || !now.Before(k.NotAfter) || r.ObservedAt.Before(k.NotBefore) || !r.ObservedAt.Before(k.NotAfter) || r.AvailableAt.After(now) || r.AvailableAt.Before(r.ObservedAt) {
		return "", errors.New("evidence issuer is unavailable, revoked, expired, or out of scope")
	}
	if len(a.Signature) != ed25519.SignatureSize {
		return "", errors.New("invalid evidence signature size")
	}
	material := append(append([]byte(nil), payload...), a.Signature...)
	cacheKey := sha256.Sum256(material)
	v.mu.Lock()
	entry, hit := v.cache[cacheKey]
	if hit {
		v.lru.MoveToFront(entry)
	}
	v.mu.Unlock()
	if !hit {
		if !ed25519.Verify(k.PublicKey, payload, a.Signature) {
			return "", errors.New("invalid evidence signature")
		}
		v.mu.Lock()
		if entry, exists := v.cache[cacheKey]; exists {
			v.lru.MoveToFront(entry)
		} else {
			v.cache[cacheKey] = v.lru.PushFront(cacheKey)
			if v.lru.Len() > v.capacity {
				old := v.lru.Back()
				delete(v.cache, old.Value.([32]byte))
				v.lru.Remove(old)
			}
		}
		v.mu.Unlock()
	}
	if len(a.Parents) > 0 {
		return "", errors.New("derived evidence cannot supply a new independent outcome")
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("evidence-observation-v1\x00%s\x00%s\x00%s", r.TenantID, a.Issuer, a.ObservationID)))
	return AuthenticatedOutcomePrefix + hex.EncodeToString(sum[:]), nil
}
