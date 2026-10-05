package libravdbstore

import (
	"context"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

var priorContextSinkV26 context.Context

// Include the scope provider's frame extraction, SHA256, JSON and allocations.
// No storage, inference, Recall or packing is included in this operation.
func BenchmarkResearchContextPriorV26Scope(b *testing.B) {
	request := model.RecallRequest{TenantID: "tenant-a", Query: "public vector search fixture", Embedding: denseQueryV6(),
		EmbeddingModel: "public-model", AsOf: time.Date(2026, 9, 30, 0, 0, 5, 0, time.UTC)}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, err := priorContextV26(context.Background(), request)
		if err != nil {
			b.Fatal(err)
		}
		priorContextSinkV26 = ctx
	}
}
