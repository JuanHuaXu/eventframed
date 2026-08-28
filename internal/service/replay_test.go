package service_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type replayFixture struct {
	AsOf              time.Time `json:"as_of"`
	RecallK           int       `json:"recall_k"`
	PackK             int       `json:"pack_k"`
	TokenBudget       int       `json:"token_budget"`
	ExpectedPackedIDs []string  `json:"expected_packed_ids"`
	Events            []struct {
		ID          string    `json:"id"`
		Content     string    `json:"content"`
		AvailableAt time.Time `json:"available_at"`
		Priority    float64   `json:"priority"`
	} `json:"events"`
}

func TestDeterministicAvailabilityAndPackingReplay(t *testing.T) {
	payload, err := os.ReadFile("../../testdata/replay/availability_and_packing.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture replayFixture
	if err := json.Unmarshal(payload, &fixture); err != nil {
		t.Fatal(err)
	}
	runtime := newMemoryService(t)
	for _, input := range fixture.Events {
		event := testutil.Event(input.ID, input.Content, input.AvailableAt)
		event.Priority = input.Priority
		event.Embedding = []float32{1, 0, 0, 0, 0, 0, 0, 0}
		event.EmbeddingModel = "feature-hash-v1:d8"
		observe(t, runtime, event)
	}
	packet, err := runtime.Recall(context.Background(), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a",
		Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: "feature-hash-v1:d8", AsOf: fixture.AsOf,
		RecallK: fixture.RecallK, PackK: fixture.PackK, TokenBudget: fixture.TokenBudget,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(packet.Candidates) != len(fixture.ExpectedPackedIDs) {
		t.Fatalf("packed count = %d, want %d", len(packet.Candidates), len(fixture.ExpectedPackedIDs))
	}
	if packet.BayesianShadow.Mode != "shadow" || packet.BayesianShadow.SelectionSupportCertified {
		t.Fatalf("Bayesian shadow escaped certification gate: %+v", packet.BayesianShadow)
	}
	for index, expected := range fixture.ExpectedPackedIDs {
		if actual := packet.Candidates[index].Event.ID; actual != expected {
			t.Fatalf("candidate %d = %q, want %q", index, actual, expected)
		}
	}
}
