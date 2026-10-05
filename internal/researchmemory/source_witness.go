package researchmemory

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"time"
	"unicode/utf8"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const SourceWitnessVersion = "research-source-hmac-v1"

// SourceWitness is a keyed commitment to the original compressed input. The
// key is never stored here; possession of a MAC does not authorize feedback.
type SourceWitness struct {
	Version         string `json:"version"`
	KeyID           string `json:"key_id"`
	FeatureContract string `json:"feature_contract"`
	MAC             string `json:"mac"`
}

func validWitnessKeyID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	for _, r := range id {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.') {
			return false
		}
	}
	return true
}

func (w SourceWitness) Validate() error {
	if w.Version != SourceWitnessVersion || !validWitnessKeyID(w.KeyID) || w.FeatureContract == "" || len(w.FeatureContract) > 128 || len(w.MAC) != 64 {
		return errors.New("invalid source witness")
	}
	decoded, err := hex.DecodeString(w.MAC)
	if err != nil || hex.EncodeToString(decoded) != w.MAC {
		return errors.New("invalid source witness MAC")
	}
	return nil
}

func NewSourceWitness(keyID string, key []byte, binding ServiceBinding, queryDigest string, event model.Event, features uint16, baseline float64) (SourceWitness, error) {
	w := SourceWitness{Version: SourceWitnessVersion, KeyID: keyID, FeatureContract: FeatureContract}
	mac, err := sourceWitnessMAC(w, key, binding, queryDigest, event, features, baseline)
	if err != nil {
		return SourceWitness{}, err
	}
	w.MAC = hex.EncodeToString(mac)
	return w, nil
}

func (w SourceWitness) Verify(keyID string, key []byte, binding ServiceBinding, queryDigest string, event model.Event, features uint16, baseline float64) (bool, error) {
	if err := w.Validate(); err != nil {
		return false, err
	}
	if w.KeyID != keyID || w.FeatureContract != FeatureContract {
		return false, nil
	}
	want, err := sourceWitnessMAC(w, key, binding, queryDigest, event, features, baseline)
	if err != nil {
		return false, err
	}
	got, _ := hex.DecodeString(w.MAC)
	return hmac.Equal(got, want), nil
}

func sourceWitnessMAC(w SourceWitness, key []byte, binding ServiceBinding, queryDigest string, event model.Event, features uint16, baseline float64) ([]byte, error) {
	if len(key) < 32 || !validWitnessKeyID(w.KeyID) || binding.validate() != nil || event.ID != binding.EventID || event.TenantID != binding.Tenant || event.SessionID == "" || event.Kind == "" || event.Provenance.Producer == "" || event.OccurredAt.IsZero() || event.ObservedAt.IsZero() || event.AvailableAt.IsZero() || features >= 512 || math.IsNaN(baseline) || math.IsInf(baseline, 0) || baseline < 0 || baseline > 1 || len(queryDigest) != 64 {
		return nil, errors.New("invalid source witness inputs")
	}
	queryBytes, err := hex.DecodeString(queryDigest)
	if err != nil || hex.EncodeToString(queryBytes) != queryDigest {
		return nil, errors.New("invalid source query digest")
	}
	fields := [6]string{event.Who.Value, event.What.Value, event.Where.Value, event.When.Value, event.Why.Value, event.How.Value}
	for _, value := range fields {
		if len(value) > 2048 || !utf8.ValidString(value) {
			return nil, errors.New("invalid compressed source field")
		}
	}
	payload, err := json.Marshal(struct {
		Version         string         `json:"version"`
		KeyID           string         `json:"key_id"`
		FeatureContract string         `json:"feature_contract"`
		Binding         ServiceBinding `json:"binding"`
		QueryDigest     string         `json:"query_digest"`
		SessionID       string         `json:"session_id"`
		Sequence        uint64         `json:"sequence"`
		Kind            string         `json:"kind"`
		OccurredAt      string         `json:"occurred_at"`
		ObservedAt      string         `json:"observed_at"`
		AvailableAt     string         `json:"available_at"`
		Producer        string         `json:"producer"`
		Fields          [6]string      `json:"fields"`
		Features        uint16         `json:"features"`
		Baseline        float64        `json:"baseline"`
	}{w.Version, w.KeyID, w.FeatureContract, binding, queryDigest,
		event.SessionID, event.Sequence, event.Kind,
		event.OccurredAt.UTC().Format(time.RFC3339Nano),
		event.ObservedAt.UTC().Format(time.RFC3339Nano),
		event.AvailableAt.UTC().Format(time.RFC3339Nano),
		event.Provenance.Producer, fields, features, baseline})
	if err != nil {
		return nil, err
	}
	h := hmac.New(sha256.New, key)
	_, _ = h.Write(payload)
	return h.Sum(nil), nil
}
