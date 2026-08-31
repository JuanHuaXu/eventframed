package translation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/fuzzing"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func BenchmarkEvaluate32Stages256DimensionsCold(b *testing.B) {
	request, chains, embedder := benchmarkFixture(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		predictor, err := fuzzing.NewEmbeddingNominationPredictor(context.Background(), embedder, request.Query)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := Evaluate(context.Background(), request, chains, predictor); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEvaluate32Stages256DimensionsCached(b *testing.B) {
	request, chains, embedder := benchmarkFixture(b)
	for _, chain := range [][]model.Event{chains.DomainABaseline, chains.DomainARevealed, chains.DomainBBaseline, chains.DomainBRevealed} {
		for i := range chain {
			vector, err := embed.Document(embedder, chain[i].FrameText())
			if err != nil {
				b.Fatal(err)
			}
			chain[i].Embedding, chain[i].EmbeddingModel = vector, embedder.ModelKey()
		}
	}
	cache := fuzzing.NewNominationCache(8, 512)
	predictor, err := fuzzing.NewEmbeddingNominationPredictorWithCache(context.Background(), embedder, request.Query, cache)
	if err != nil {
		b.Fatal(err)
	}
	if _, err := Evaluate(context.Background(), request, chains, predictor); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Evaluate(context.Background(), request, chains, predictor); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRejectStructuralDivergence32Stages(b *testing.B) {
	request, chains, _ := benchmarkFixture(b)
	chains.DomainBRevealed[1].Where.Value = "undeclared movement"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := EvaluateWithFactory(context.Background(), request, chains, func(context.Context) (Predictor, error) {
			b.Fatal("structural divergence constructed predictor")
			return nil, nil
		}); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkFixture(b *testing.B) (model.ChainTranslationRequest, Chains, *embed.HashEmbedder) {
	b.Helper()
	now := time.Now().UTC()
	chains := Chains{}
	request := model.ChainTranslationRequest{TenantID: "tenant-a", Query: "aligned chain", AsOf: now, InvariantThreshold: .02, TranslationThreshold: .05}
	for stage := 0; stage < 32; stage++ {
		beforeA, afterA := fmt.Sprintf("a-before-%d", stage), fmt.Sprintf("a-after-%d", stage)
		beforeB, afterB := fmt.Sprintf("b-before-%d", stage), fmt.Sprintf("b-after-%d", stage)
		availableAt := now.Add(time.Duration(stage-64) * time.Minute)
		chains.DomainABaseline = append(chains.DomainABaseline, testutil.Event(fmt.Sprintf("a0-%d", stage), beforeA, availableAt))
		chains.DomainARevealed = append(chains.DomainARevealed, testutil.Event(fmt.Sprintf("a1-%d", stage), afterA, availableAt))
		chains.DomainBBaseline = append(chains.DomainBBaseline, testutil.Event(fmt.Sprintf("b0-%d", stage), beforeB, availableAt))
		chains.DomainBRevealed = append(chains.DomainBRevealed, testutil.Event(fmt.Sprintf("b1-%d", stage), afterB, availableAt))
		request.StageMaps = append(request.StageMaps, model.ChainStageMap{Stage: stage, DomainAField: model.FuzzWhat, DomainBField: model.FuzzWhat,
			DomainABefore: beforeA, DomainAAfter: afterA, DomainBBefore: beforeB, DomainBAfter: afterB,
			CorrespondenceID: fmt.Sprintf("map-%d", stage), ValidityEvidence: "benchmark"})
	}
	request.DomainA = chainIDsForBenchmark(chains.DomainABaseline, chains.DomainARevealed)
	request.DomainB = chainIDsForBenchmark(chains.DomainBBaseline, chains.DomainBRevealed)
	embedder, _ := embed.NewHashEmbedder(256)
	return request, chains, embedder
}

func chainIDsForBenchmark(baseline, revealed []model.Event) model.ChainTrajectory {
	result := model.ChainTrajectory{BaselineEventIDs: make([]string, len(baseline)), RevealedEventIDs: make([]string, len(revealed))}
	for i := range baseline {
		result.BaselineEventIDs[i], result.RevealedEventIDs[i] = baseline[i].ID, revealed[i].ID
	}
	return result
}
