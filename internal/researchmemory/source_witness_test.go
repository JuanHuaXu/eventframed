package researchmemory

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func sourceWitnessFixture() (ServiceBinding, string, model.Event, []byte) {
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	event := model.Event{
		ID: "event-b", TenantID: "tenant", SessionID: "session", Sequence: 7, Kind: "observation",
		Content: "private source content", OccurredAt: now, ObservedAt: now, AvailableAt: now,
		Who: model.Field{Value: "private subject"}, What: model.Field{Value: "private action"},
		Where: model.Field{Value: "private location"}, When: model.Field{Value: "private date"},
		Why: model.Field{Value: "private cause"}, How: model.Field{Value: "private method"},
		Provenance: model.Provenance{Producer: "fixture"},
	}
	binding := ServiceBinding{Tenant: "tenant", JournalID: "journal", EventID: event.ID, Snapshot: model.Snapshot{RuntimeVersion: 1, ContractVersion: 1}}
	digest := sha256.Sum256([]byte("private query"))
	return binding, hex.EncodeToString(digest[:]), event, []byte("0123456789abcdef0123456789abcdef")
}

func TestSourceWitnessBindsCompressedSourceWithoutExportingIt(t *testing.T) {
	binding, digest, event, key := sourceWitnessFixture()
	w, err := NewSourceWitness("lab-v1", key, binding, digest, event, 7, .6)
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := w.Verify("lab-v1", key, binding, digest, event, 7, .6); err != nil || !ok {
		t.Fatalf("original witness rejected: match=%v err=%v", ok, err)
	}
	contentChanged := event
	contentChanged.Content = "different non-feature text"
	if ok, err := w.Verify("lab-v1", key, binding, digest, contentChanged, 7, .6); err != nil || !ok {
		t.Fatalf("non-feature content invalidated witness: match=%v err=%v", ok, err)
	}
	for _, tc := range []struct {
		name     string
		keyID    string
		key      []byte
		binding  ServiceBinding
		digest   string
		event    model.Event
		features uint16
		baseline float64
		witness  SourceWitness
	}{
		{"wrong-key", "lab-v1", []byte("fedcba9876543210fedcba9876543210"), binding, digest, event, 7, .6, w},
		{"wrong-query", "lab-v1", key, binding, strings.Repeat("a", 64), event, 7, .6, w},
		{"wrong-features", "lab-v1", key, binding, digest, event, 6, .6, w},
		{"wrong-baseline", "lab-v1", key, binding, digest, event, 7, .7, w},
		{"wrong-journal", "lab-v1", key, ServiceBinding{Tenant: "tenant", JournalID: "other", EventID: event.ID, Snapshot: binding.Snapshot}, digest, event, 7, .6, w},
		{"wrong-fields", "lab-v1", key, binding, digest, func() model.Event { v := event; v.What.Value = "different"; return v }(), 7, .6, w},
		{"wrong-identity", "lab-v1", key, binding, digest, func() model.Event { v := event; v.ObservedAt = v.ObservedAt.Add(time.Second); return v }(), 7, .6, w},
		{"wrong-key-id", "lab-v2", key, binding, digest, event, 7, .6, w},
		{"wrong-contract", "lab-v1", key, binding, digest, event, 7, .6, func() SourceWitness { v := w; v.FeatureContract = "other"; return v }()},
		{"tampered-mac", "lab-v1", key, binding, digest, event, 7, .6, func() SourceWitness { v := w; v.MAC = strings.Repeat("0", 64); return v }()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := tc.witness.Verify(tc.keyID, tc.key, tc.binding, tc.digest, tc.event, tc.features, tc.baseline)
			if err == nil && ok {
				t.Fatal("changed input retained source")
			}
		})
	}
	if _, err := NewSourceWitness("lab-v1", []byte("short"), binding, digest, event, 7, .6); err == nil {
		t.Fatal("short source key accepted")
	}
	if _, err := NewSourceWitness("private subject", key, binding, digest, event, 7, .6); err == nil {
		t.Fatal("free-text key ID accepted")
	}
	missingTime := event
	missingTime.ObservedAt = time.Time{}
	if _, err := NewSourceWitness("lab-v1", key, binding, digest, missingTime, 7, .6); err == nil {
		t.Fatal("incomplete event identity accepted")
	}
	encoded, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"private", "event-b", "journal", "query", string(key)} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("witness leaked %q", forbidden)
		}
	}
}

func TestSourceWitnessOfflineCost(t *testing.T) {
	if os.Getenv("EVENTFRAME_WITNESS_COST") == "" {
		t.Skip("run explicitly for the frozen source-witness cost screen")
	}
	binding, digest, event, key := sourceWitnessFixture()
	durations := make([]time.Duration, 100)
	for i := range durations {
		start := time.Now()
		for j := 0; j < 200; j++ {
			if _, err := NewSourceWitness("lab-v1", key, binding, digest, event, 7, .6); err != nil {
				t.Fatal(err)
			}
		}
		durations[i] = time.Since(start)
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	p95 := durations[94]
	t.Logf("200 source witnesses p95=%s", p95)
	if p95 >= 2*time.Millisecond {
		t.Fatalf("source witness component exceeds frozen 2ms screen: %s", p95)
	}
}
