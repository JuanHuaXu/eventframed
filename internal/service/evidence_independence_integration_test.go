package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

func TestRepeatedCapturedConversationCannotMonopolizePacket(t *testing.T) {
	ctx := context.Background()
	memory := memorystore.New()
	embedder, err := embed.NewHashEmbedder(16)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := service.New(memory, embedder, service.Config{DefaultRecallK: 10, DefaultPackK: 10, DefaultTokenBudget: 10_000, OverfetchMultiplier: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Close()

	now := time.Now().UTC().Add(-time.Minute)
	for index := 0; index < 3; index++ {
		at := now.Add(time.Duration(index) * time.Second)
		id := fmt.Sprintf("repeated-turn-%d", index)
		turn := model.TurnCapture{
			ID: id, TenantID: "tenant-a", SessionID: fmt.Sprintf("session-%d", index), Sequence: uint64(index + 1), RunID: fmt.Sprintf("run-%d", index),
			UserText: "What city is the project office in?", AssistantText: "The project office is in Example City.",
			OccurredAt: at, ObservedAt: at, AvailableAt: at,
		}
		if _, err := runtime.CaptureTurn(ctx, model.CaptureTurnRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Turn: turn}); err != nil {
			t.Fatal(err)
		}
	}
	packet, err := runtime.Recall(ctx, model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "current", Query: "project office city",
		AsOf: now.Add(time.Minute), RecallK: 10, PackK: 10, TokenBudget: 10_000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if packet.Recalled != 3 || packet.Packed != 1 || packet.CorrelatedSuppressed != 2 {
		t.Fatalf("repetition reached context sink: recalled=%d packed=%d suppressed=%d", packet.Recalled, packet.Packed, packet.CorrelatedSuppressed)
	}
	marked := 0
	for _, decision := range packet.BayesianShadow.Decisions {
		if decision.CorrelatedSuppressed {
			marked++
		}
	}
	if marked != 2 {
		t.Fatalf("repetition was not marked as correlated selective evidence: %+v", packet.BayesianShadow)
	}
	encoded, err := json.Marshal(packet)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "evidence_group_key") {
		t.Fatal("derived claim fingerprint escaped the daemon boundary")
	}
}
