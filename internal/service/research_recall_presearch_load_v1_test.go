package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type presearchLoadStatsV1 struct {
	recallNS, writeNS, adjustedNS []int64
	searches, mismatches          int
}

func runPresearchLoadArmV1(t *testing.T, block int, arm string) presearchLoadStatsV1 {
	t.Helper()
	ctx := context.Background()
	asOf := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	em, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	db, err := libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/presearch-load.libravdb", Dimension: 8, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
	if err != nil {
		t.Fatal(err)
	}
	wrapped := &searchSnapshotInterleaverV1{EventStore: db, pinBeforeSearch: arm != "current_write"}
	s, err := New(wrapped, em, Config{DefaultRecallK: 50, DefaultPackK: 50, DefaultTokenBudget: 10000})
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	defer s.Close()
	put := func(tenant, id string) {
		event := testutil.Event(id, "public query", asOf.Add(-time.Minute))
		event.TenantID = tenant
		if _, err := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: event}); err != nil {
			t.Fatal(err)
		}
	}
	var out presearchLoadStatsV1
	for i := 0; i < 32; i++ {
		tenant := fmt.Sprintf("tenant-%d-%s-%02d", block, arm, i)
		seedID := fmt.Sprintf("seed-%d-%02d", block, i)
		newID := fmt.Sprintf("new-%d-%02d", block, i)
		put(tenant, seedID)
		var writeNS int64
		if arm != "pin_quiet" {
			wrapped.afterSearch = func() {
				start := time.Now()
				put(tenant, newID)
				writeNS = time.Since(start).Nanoseconds()
			}
		}
		beforeSearches := wrapped.attempts
		start := time.Now()
		packet, err := s.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: tenant, SessionID: tenant, Query: "public query", AsOf: asOf, RecallK: 50, PackK: 50, TokenBudget: 10000})
		rawNS := time.Since(start).Nanoseconds()
		if err != nil {
			t.Fatal(arm, i, err)
		}
		journal, err := db.GetBayesianJournal(ctx, tenant, packet.BayesianShadow.JournalID)
		if err != nil {
			t.Fatal(err)
		}
		seen := make(map[string]bool, len(packet.Candidates))
		for _, candidate := range packet.Candidates {
			if candidate.Event.AvailableAt.After(asOf) {
				t.Fatal("future event reached packet", arm, i, candidate.Event.ID)
			}
			seen[candidate.Event.ID] = true
		}
		if packet.Snapshot != journal.Snapshot || packet.Snapshot != db.Snapshot(ctx) || !seen[seedID] {
			t.Fatal("packet/journal/store snapshot or seed mismatch", arm, i)
		}
		attempts := wrapped.attempts - beforeSearches
		expectedAttempts, expectedCandidates := 1, 1
		if arm == "pin_write" {
			expectedAttempts, expectedCandidates = 2, 2
		}
		if attempts != expectedAttempts || len(seen) != expectedCandidates {
			t.Fatal("unexpected search/candidate count", arm, i, attempts, seen)
		}
		if arm == "pin_write" && !seen[newID] {
			t.Fatal("pinned Recall omitted as-of-visible write", i)
		}
		if arm == "current_write" {
			if seen[newID] {
				t.Fatal("negative control unexpectedly saw after-search write", i)
			}
			out.mismatches++
		}
		out.searches += attempts
		out.recallNS = append(out.recallNS, rawNS)
		out.writeNS = append(out.writeNS, writeNS)
		out.adjustedNS = append(out.adjustedNS, rawNS-writeNS)
	}
	return out
}

func TestResearchRecallPresearchLoadV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_PRESEARCH_LOAD_V1") != "1" {
		t.Skip("opt-in repeated-write pre-search snapshot screen")
	}
	arms := []string{"current_write", "pin_write", "pin_quiet"}
	stats := map[string]presearchLoadStatsV1{}
	for block := 0; block < 3; block++ {
		for offset := 0; offset < len(arms); offset++ {
			arm := arms[(block+offset)%len(arms)]
			result := runPresearchLoadArmV1(t, block, arm)
			pooled := stats[arm]
			pooled.recallNS = append(pooled.recallNS, result.recallNS...)
			pooled.writeNS = append(pooled.writeNS, result.writeNS...)
			pooled.adjustedNS = append(pooled.adjustedNS, result.adjustedNS...)
			pooled.searches += result.searches
			pooled.mismatches += result.mismatches
			stats[arm] = pooled
		}
	}
	for _, arm := range arms {
		r := stats[arm]
		if len(r.recallNS) != 96 || (arm == "current_write" && r.mismatches != 96) {
			t.Fatal("missing arm results or negative control", arm, len(r.recallNS), r.mismatches)
		}
		t.Logf("arm=%s n=%d searches=%d mismatches=%d recall_p50=%s recall_p95=%s recall_p99=%s write_p50=%s write_p99=%s adjusted_p50=%s adjusted_p99=%s",
			arm, len(r.recallNS), r.searches, r.mismatches,
			researchDurableLoadPercentile(r.recallNS, .5), researchDurableLoadPercentile(r.recallNS, .95), researchDurableLoadPercentile(r.recallNS, .99),
			researchDurableLoadPercentile(r.writeNS, .5), researchDurableLoadPercentile(r.writeNS, .99),
			researchDurableLoadPercentile(r.adjustedNS, .5), researchDurableLoadPercentile(r.adjustedNS, .99))
	}
}
