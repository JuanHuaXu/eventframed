package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

var benchmarkOpenClawPacket openClawContextPacket

func BenchmarkProjectOpenClawPacket50(b *testing.B) {
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	packet := model.ContextPacket{ProtocolVersion: model.ProtocolVersion, Candidates: make([]model.Candidate, 50)}
	for index := range packet.Candidates {
		event := testutil.Event(fmt.Sprintf("event-%03d", index), "benchmark recalled content", now)
		event.Embedding = make([]float32, 384)
		event.EmbeddingModel = "benchmark:d384"
		packet.Candidates[index] = model.Candidate{Event: event, Similarity: .8, RetrievalScore: .8, Score: .8, EstimatedTokens: 8}
	}
	b.ReportAllocs()
	b.ReportMetric(50, "candidates")
	for b.Loop() {
		benchmarkOpenClawPacket = projectOpenClawPacket(packet)
	}
}
