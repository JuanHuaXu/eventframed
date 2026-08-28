package agency

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const (
	CapabilityWake     = "eventframe.agency.wake"
	CapabilityNotify   = "eventframe.agency.notify"
	CapabilitySchedule = "eventframe.agency.schedule"
	maxIdentifierBytes = 256
	maxSessionIDBytes  = 1024
)

type Policy struct {
	Enabled              bool
	MaxEvidence          int
	MaxReasonBytes       int
	MaxClaims            int
	MaxChainDepth        int
	MaxProposalsPerChain int
	MaxPendingPerTenant  int
	MaxFuture            time.Duration
	MaxTTL               time.Duration
	LeaseDuration        time.Duration
}

func DefaultPolicy(enabled bool) Policy {
	return Policy{Enabled: enabled, MaxEvidence: 32, MaxReasonBytes: 4096, MaxClaims: 50, MaxChainDepth: 3, MaxProposalsPerChain: 8, MaxPendingPerTenant: 1000, MaxFuture: 30 * 24 * time.Hour, MaxTTL: 7 * 24 * time.Hour, LeaseDuration: 30 * time.Second}
}

func (policy Policy) Valid() bool {
	return policy.MaxEvidence > 0 && policy.MaxReasonBytes > 0 && policy.MaxClaims > 0 && policy.MaxChainDepth >= 0 && policy.MaxProposalsPerChain > 0 && policy.MaxPendingPerTenant > 0 && policy.MaxFuture > 0 && policy.MaxTTL > 0 && policy.LeaseDuration > 0
}

type Signer struct {
	private ed25519.PrivateKey
	public  ed25519.PublicKey
	keyID   string
}

func LoadOrCreateSigner(privatePath, publicPath string) (*Signer, error) {
	if privatePath == "" || publicPath == "" || privatePath == publicPath {
		return nil, errors.New("distinct agency private and public key paths are required")
	}
	if err := os.MkdirAll(filepath.Dir(privatePath), 0o700); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(publicPath), 0o700); err != nil {
		return nil, err
	}
	encoded, err := readRestrictedFile(privatePath, "agency private key")
	if errors.Is(err, os.ErrNotExist) {
		_, privateKey, generateErr := ed25519.GenerateKey(rand.Reader)
		if generateErr != nil {
			return nil, generateErr
		}
		encoded = []byte(base64.RawURLEncoding.EncodeToString(privateKey) + "\n")
		file, openErr := os.OpenFile(privatePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if openErr != nil {
			return nil, fmt.Errorf("create agency private key: %w", openErr)
		}
		if _, writeErr := file.Write(encoded); writeErr != nil {
			_ = file.Close()
			return nil, writeErr
		}
		if closeErr := file.Close(); closeErr != nil {
			return nil, closeErr
		}
	} else if err != nil {
		return nil, err
	}
	privateKey, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(string(encoded)))
	if err != nil || len(privateKey) != ed25519.PrivateKeySize {
		return nil, errors.New("agency private key is malformed")
	}
	publicKey := ed25519.PrivateKey(privateKey).Public().(ed25519.PublicKey)
	if err := ensurePublicKeyFile(publicPath, publicKey); err != nil {
		return nil, err
	}
	digest := sha256.Sum256(publicKey)
	return &Signer{private: ed25519.PrivateKey(privateKey), public: publicKey, keyID: hex.EncodeToString(digest[:12])}, nil
}

func NewSignerForTest() (*Signer, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(publicKey)
	return &Signer{private: privateKey, public: publicKey, keyID: hex.EncodeToString(digest[:12])}, nil
}

func LoadOrCreateIssuerToken(tokenPath string) (string, error) {
	return loadOrCreateToken(tokenPath, "issuer")
}

func LoadOrCreateAuthorityToken(tokenPath string) (string, error) {
	return loadOrCreateToken(tokenPath, "authority")
}

func loadOrCreateToken(tokenPath, label string) (string, error) {
	if tokenPath == "" {
		return "", fmt.Errorf("agency %s token path is required", label)
	}
	if err := os.MkdirAll(filepath.Dir(tokenPath), 0o700); err != nil {
		return "", err
	}
	encoded, err := readRestrictedFile(tokenPath, "agency "+label+" token")
	if errors.Is(err, os.ErrNotExist) {
		raw := make([]byte, 32)
		if _, readErr := rand.Read(raw); readErr != nil {
			return "", readErr
		}
		value := base64.RawURLEncoding.EncodeToString(raw)
		file, openErr := os.OpenFile(tokenPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if openErr != nil {
			return "", fmt.Errorf("create agency %s token: %w", label, openErr)
		}
		if _, writeErr := file.WriteString(value + "\n"); writeErr != nil {
			_ = file.Close()
			return "", writeErr
		}
		if closeErr := file.Close(); closeErr != nil {
			return "", closeErr
		}
		encoded = []byte(value)
	} else if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(encoded))
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(raw) != 32 {
		return "", fmt.Errorf("agency %s token is malformed", label)
	}
	return value, nil
}

func readRestrictedFile(path, label string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s must be a regular non-symlink file", label)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("%s must not be accessible by group or others", label)
	}
	return os.ReadFile(path)
}

func ensurePublicKeyFile(path string, publicKey ed25519.PublicKey) error {
	expected := base64.RawURLEncoding.EncodeToString(publicKey) + "\n"
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		file, openErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if openErr != nil {
			return fmt.Errorf("create agency public key: %w", openErr)
		}
		if _, writeErr := file.WriteString(expected); writeErr != nil {
			_ = file.Close()
			return writeErr
		}
		return file.Close()
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return errors.New("agency public key must be a regular non-symlink file")
	}
	existing, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoded, decodeErr := base64.RawURLEncoding.DecodeString(strings.TrimSpace(string(existing)))
	if decodeErr != nil || !ed25519.PublicKey(decoded).Equal(publicKey) {
		return errors.New("agency public key does not match private key")
	}
	return nil
}

func (signer *Signer) Sign(proposal model.AgencyProposal) (model.SignedAgencyProposal, error) {
	if signer == nil || len(signer.private) != ed25519.PrivateKeySize {
		return model.SignedAgencyProposal{}, errors.New("agency signer is unavailable")
	}
	payload, err := json.Marshal(proposal)
	if err != nil {
		return model.SignedAgencyProposal{}, err
	}
	signature := ed25519.Sign(signer.private, payload)
	return model.SignedAgencyProposal{Payload: base64.RawURLEncoding.EncodeToString(payload), Signature: base64.RawURLEncoding.EncodeToString(signature), KeyID: signer.keyID}, nil
}

func (signer *Signer) Verify(signed model.SignedAgencyProposal) (model.AgencyProposal, error) {
	payload, err := base64.RawURLEncoding.DecodeString(signed.Payload)
	if err != nil {
		return model.AgencyProposal{}, err
	}
	signature, err := base64.RawURLEncoding.DecodeString(signed.Signature)
	if err != nil || signed.KeyID != signer.keyID || !ed25519.Verify(signer.public, payload, signature) {
		return model.AgencyProposal{}, errors.New("agency signature is invalid")
	}
	var proposal model.AgencyProposal
	if err := json.Unmarshal(payload, &proposal); err != nil {
		return model.AgencyProposal{}, err
	}
	return proposal, nil
}

func BuildProposal(draft model.AgencyProposalDraft, now time.Time, policy Policy) (model.AgencyProposal, error) {
	if !policy.Enabled {
		return model.AgencyProposal{}, errors.New("agency is disabled")
	}
	if !policy.Valid() {
		return model.AgencyProposal{}, errors.New("agency policy is invalid")
	}
	draft.ID, draft.TenantID, draft.SessionID = strings.TrimSpace(draft.ID), strings.TrimSpace(draft.TenantID), strings.TrimSpace(draft.SessionID)
	draft.Reason, draft.IdempotencyKey, draft.CausalChainID = strings.TrimSpace(draft.Reason), strings.TrimSpace(draft.IdempotencyKey), strings.TrimSpace(draft.CausalChainID)
	if draft.ID == "" || draft.TenantID == "" || draft.SessionID == "" || draft.IdempotencyKey != draft.ID || draft.CausalChainID == "" {
		return model.AgencyProposal{}, errors.New("proposal id, matching idempotency key, tenant, session, and causal chain are required")
	}
	if !boundedString(draft.ID, maxIdentifierBytes) || !boundedString(draft.TenantID, maxIdentifierBytes) || !boundedString(draft.SessionID, maxSessionIDBytes) ||
		!boundedString(draft.CausalChainID, maxIdentifierBytes) || !boundedOptionalString(strings.TrimSpace(draft.ParentProposalID), maxIdentifierBytes) {
		return model.AgencyProposal{}, errors.New("proposal identifiers exceed the authority contract")
	}
	capability, err := capabilityFor(draft.Action)
	if err != nil {
		return model.AgencyProposal{}, err
	}
	if draft.Reason == "" || len([]byte(draft.Reason)) > policy.MaxReasonBytes {
		return model.AgencyProposal{}, errors.New("proposal reason is empty or exceeds the byte limit")
	}
	if !finite(draft.ExpectedUtility) || draft.ExpectedUtility <= 0 || draft.ExpectedUtility > 1 || !finite(draft.Priority) || draft.Priority < 0 || draft.Priority > 1 {
		return model.AgencyProposal{}, errors.New("expected utility must be in (0,1] and priority in [0,1]")
	}
	if draft.NotBefore.IsZero() || draft.NotBefore.After(now.Add(policy.MaxFuture)) || !draft.ExpiresAt.After(now) || draft.ExpiresAt.After(now.Add(policy.MaxFuture)) || !draft.NotBefore.Before(draft.ExpiresAt) || draft.ExpiresAt.Sub(draft.NotBefore) > policy.MaxTTL {
		return model.AgencyProposal{}, errors.New("proposal availability and expiry exceed policy")
	}
	if draft.Action == model.AgencySchedule {
		if draft.ScheduledFor == nil || draft.ScheduledFor.Before(draft.NotBefore) || !draft.ScheduledFor.Before(draft.ExpiresAt) {
			return model.AgencyProposal{}, errors.New("schedule proposals require scheduled_for within the validity window")
		}
	} else if draft.ScheduledFor != nil {
		return model.AgencyProposal{}, errors.New("scheduled_for is valid only for schedule proposals")
	}
	if draft.CausalChainDepth < 0 || draft.CausalChainDepth > policy.MaxChainDepth || (draft.CausalChainDepth == 0 && draft.ParentProposalID != "") || (draft.CausalChainDepth > 0 && strings.TrimSpace(draft.ParentProposalID) == "") {
		return model.AgencyProposal{}, errors.New("causal chain depth or parent binding violates policy")
	}
	evidence := append([]string(nil), draft.EvidenceIDs...)
	if len(evidence) == 0 || len(evidence) > policy.MaxEvidence {
		return model.AgencyProposal{}, errors.New("proposal requires a bounded non-empty evidence set")
	}
	sort.Strings(evidence)
	for index, id := range evidence {
		if strings.TrimSpace(id) == "" || !boundedString(id, maxIdentifierBytes) || (index > 0 && id == evidence[index-1]) {
			return model.AgencyProposal{}, errors.New("proposal evidence ids must be non-empty and unique")
		}
	}
	return model.AgencyProposal{
		ID: draft.ID, TenantID: draft.TenantID, SessionID: draft.SessionID, Action: draft.Action, Reason: draft.Reason,
		EvidenceIDs: evidence, ExpectedUtility: draft.ExpectedUtility, Priority: draft.Priority, RequiredCapability: capability,
		NotBefore: draft.NotBefore.UTC(), ScheduledFor: utcPointer(draft.ScheduledFor), ExpiresAt: draft.ExpiresAt.UTC(), IdempotencyKey: draft.IdempotencyKey,
		CausalChainID: draft.CausalChainID, ParentProposalID: strings.TrimSpace(draft.ParentProposalID), CausalChainDepth: draft.CausalChainDepth,
		CreatedAt: now.UTC(), ContractVersion: model.ContractVersion,
	}, nil
}

func boundedString(value string, maximum int) bool {
	return value != "" && len([]byte(value)) <= maximum
}

func boundedOptionalString(value string, maximum int) bool {
	return value == "" || len([]byte(value)) <= maximum
}

func capabilityFor(action model.AgencyAction) (string, error) {
	switch action {
	case model.AgencyWake:
		return CapabilityWake, nil
	case model.AgencyNotify:
		return CapabilityNotify, nil
	case model.AgencySchedule:
		return CapabilitySchedule, nil
	default:
		return "", errors.New("agency action must be wake, notify, or schedule")
	}
}

func utcPointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
