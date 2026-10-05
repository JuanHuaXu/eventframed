package researchmemory

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

func boundTransferFixture(count int, aUseful, bUseful bool) ([]BoundLabel, model.Snapshot, time.Time) {
	old := model.Snapshot{RuntimeVersion: 1, ContractVersion: 1}
	target := model.Snapshot{RuntimeVersion: 2, ContractVersion: 1}
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	labels := make([]BoundLabel, count)
	for i := range labels {
		event, useful, features := "source-a", aUseful, uint16(0)
		if i%2 == 1 {
			event, useful, features = "source-b", bUseful, 1
		}
		at := start.Add(time.Duration(i) * time.Second)
		labels[i] = BoundLabel{
			Prediction: RecordedPrediction{
				Contract:   RecordContract,
				Seed:       42,
				Prediction: Prediction{ID: uint64(i + 1), Probability: .5, Features: features, Epoch: 1},
				At:         at,
				Outer:      [4]float64{.5, .5, .5, .5},
				Binding:    &ServiceBinding{Tenant: "tenant", JournalID: fmt.Sprintf("journal-%d", i), EventID: event, Snapshot: old},
			},
			Feedback: RecordedFeedback{ID: uint64(i + 1), Useful: useful, Available: at.Add(time.Second)},
		}
	}
	return labels, target, start.Add(time.Duration(count+1) * time.Second)
}

func retainSourceB(_ context.Context, target model.Snapshot, label BoundLabel) (bool, error) {
	return target.RuntimeVersion == 2 && label.Prediction.Binding.EventID == "source-b", nil
}

func TestBoundTransferRetainsOnlyValidatedSourceAndResetsMixture(t *testing.T) {
	labels, target, cutoff := boundTransferFixture(64, true, false)
	a, count, err := RebuildFromBoundLabels(context.Background(), target, "tenant", 2, 42, cutoff, labels, retainSourceB)
	if err != nil || count != 32 {
		t.Fatalf("rebuild count=%d err=%v", count, err)
	}
	if a.outer.Weights != [4]float64{} || a.inner.Weights != [4]float64{} || len(a.pending) != 0 {
		t.Fatal("old mixture or pending predictions crossed epoch")
	}
	expected := make([]observation.Sample, 32)
	for i := range expected {
		expected[i] = observation.Sample{Bits: 1, Outcome: false}
	}
	short, err := observation.Fit(expected)
	if err != nil {
		t.Fatal(err)
	}
	long, err := observation.Fit(expected)
	if err != nil {
		t.Fatal(err)
	}
	forest := observationlearners.NewForest(42)
	for _, sample := range expected {
		forest.Update(sample.Bits, sample.Outcome)
	}
	for feature := uint16(0); feature < 512; feature++ {
		gotShort, _ := a.short.ForecastObserved(511, feature)
		wantShort, _ := short.ForecastObserved(511, feature)
		gotLong, _ := a.long.ForecastObserved(511, feature)
		wantLong, _ := long.ForecastObserved(511, feature)
		if gotShort != wantShort || gotLong != wantLong || a.forest.Predict(feature) != forest.Predict(feature) {
			t.Fatalf("retained component mismatch at feature %d", feature)
		}
	}
	changedA, _, _ := boundTransferFixture(64, false, false)
	aOnlyChanged, _, err := RebuildFromBoundLabels(context.Background(), target, "tenant", 2, 42, cutoff, changedA, retainSourceB)
	if err != nil {
		t.Fatal(err)
	}
	changedB, _, _ := boundTransferFixture(64, true, true)
	bChanged, _, err := RebuildFromBoundLabels(context.Background(), target, "tenant", 2, 42, cutoff, changedB, retainSourceB)
	if err != nil {
		t.Fatal(err)
	}
	for feature := uint16(0); feature < 512; feature++ {
		base, _ := a.Freeze().Score(feature, .5, 2, cutoff)
		noStaleEffect, _ := aOnlyChanged.Freeze().Score(feature, .5, 2, cutoff)
		if base != noStaleEffect {
			t.Fatalf("rejected source influenced feature %d", feature)
		}
	}
	base, _ := a.Freeze().Score(1, .5, 2, cutoff)
	opposite, _ := bChanged.Freeze().Score(1, .5, 2, cutoff)
	if math.Abs(base-opposite) < .001 {
		t.Fatal("retained outcome reversal had no predictive effect")
	}
	if _, err := a.Freeze().Score(1, .5, 1, cutoff); err == nil {
		t.Fatal("old epoch accepted rebuilt model")
	}
}

func TestBoundTransferRejectsUnverifiedAndMalformedHistory(t *testing.T) {
	labels, target, cutoff := boundTransferFixture(64, true, false)
	for _, tc := range []struct {
		name   string
		change func([]BoundLabel) []BoundLabel
		check  BoundLabelValidator
	}{
		{"wrong-tenant", func(v []BoundLabel) []BoundLabel { v[0].Prediction.Binding.Tenant = "other"; return v }, retainSourceB},
		{"duplicate", func(v []BoundLabel) []BoundLabel {
			v[1].Prediction.Binding.JournalID = v[0].Prediction.Binding.JournalID
			v[1].Prediction.Binding.EventID = v[0].Prediction.Binding.EventID
			return v
		}, retainSourceB},
		{"duplicate-id", func(v []BoundLabel) []BoundLabel {
			v[1].Prediction.Prediction.ID = v[0].Prediction.Prediction.ID
			v[1].Feedback.ID = v[0].Feedback.ID
			return v
		}, retainSourceB},
		{"mismatched-id", func(v []BoundLabel) []BoundLabel { v[1].Feedback.ID++; return v }, retainSourceB},
		{"unbound", func(v []BoundLabel) []BoundLabel { v[0].Prediction.Binding = nil; return v }, retainSourceB},
		{"out-of-order", func(v []BoundLabel) []BoundLabel {
			v[0].Feedback.Available = v[1].Feedback.Available.Add(time.Second)
			return v
		}, retainSourceB},
		{"early", func(v []BoundLabel) []BoundLabel {
			v[0].Feedback.Available = v[0].Prediction.At.Add(-time.Second)
			return v
		}, retainSourceB},
		{"future", func(v []BoundLabel) []BoundLabel { v[63].Feedback.Available = cutoff.Add(time.Second); return v }, retainSourceB},
		{"feature-range", func(v []BoundLabel) []BoundLabel { v[0].Prediction.Prediction.Features = 512; return v }, retainSourceB},
		{"bad-contract", func(v []BoundLabel) []BoundLabel { v[0].Prediction.Contract = "other"; return v }, retainSourceB},
		{"nil-validator", func(v []BoundLabel) []BoundLabel { return v }, nil},
		{"validator-error", func(v []BoundLabel) []BoundLabel { return v }, func(context.Context, model.Snapshot, BoundLabel) (bool, error) {
			return false, errors.New("source unavailable")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			copyLabels := append([]BoundLabel(nil), labels...)
			for i := range copyLabels {
				binding := *copyLabels[i].Prediction.Binding
				copyLabels[i].Prediction.Binding = &binding
			}
			copyLabels = tc.change(copyLabels)
			got, count, err := RebuildFromBoundLabels(context.Background(), target, "tenant", 2, 42, cutoff, copyLabels, tc.check)
			if err == nil || got != nil || count != 0 {
				t.Fatalf("invalid history exposed state: count=%d err=%v", count, err)
			}
		})
	}
	cold, count, err := RebuildFromBoundLabels(context.Background(), target, "tenant", 2, 42, cutoff, labels, func(context.Context, model.Snapshot, BoundLabel) (bool, error) { return false, nil })
	if err != nil || count != 0 {
		t.Fatalf("all-rejected history count=%d err=%v", count, err)
	}
	if got, err := cold.Freeze().Score(1, .6, 2, cutoff); err != nil || got != .6 {
		t.Fatalf("cold model changed baseline: p=%v err=%v", got, err)
	}
	oversized, _, oversizedCutoff := boundTransferFixture(257, true, false)
	if rebuilt, count, err := RebuildFromBoundLabels(context.Background(), target, "tenant", 2, 42, oversizedCutoff, oversized, retainSourceB); err == nil || rebuilt != nil || count != 0 {
		t.Fatal("oversized history exposed state")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if rebuilt, count, err := RebuildFromBoundLabels(ctx, target, "tenant", 2, 42, cutoff, labels, retainSourceB); !errors.Is(err, context.Canceled) || rebuilt != nil || count != 0 {
		t.Fatalf("canceled rebuild exposed state: count=%d err=%v", count, err)
	}
}

func TestBoundTransferOfflineCost(t *testing.T) {
	if os.Getenv("EVENTFRAME_TRANSFER_COST") == "" {
		t.Skip("run explicitly for the frozen offline cost screen")
	}
	for _, count := range []int{64, 256} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			labels, target, cutoff := boundTransferFixture(count, true, false)
			durations := make([]time.Duration, 100)
			for i := range durations {
				start := time.Now()
				if _, _, err := RebuildFromBoundLabels(context.Background(), target, "tenant", 2, 42, cutoff, labels, retainSourceB); err != nil {
					t.Fatal(err)
				}
				durations[i] = time.Since(start)
			}
			sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
			p95 := durations[94]
			allocs := testing.AllocsPerRun(5, func() {
				_, _, err := RebuildFromBoundLabels(context.Background(), target, "tenant", 2, 42, cutoff, labels, retainSourceB)
				if err != nil {
					panic(err)
				}
			})
			t.Logf("offline reconstruction n=%d p95=%s allocations=%.0f", count, p95, allocs)
			if p95 >= 100*time.Millisecond {
				t.Fatalf("offline reconstruction exceeds frozen 100ms screen: %s", p95)
			}
		})
	}
}
