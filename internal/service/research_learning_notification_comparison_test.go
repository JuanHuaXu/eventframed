package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"sort"
	"testing"
)

// Both modes use the current worker and general adapter; only waiting differs.
func TestResearchLearningNotificationComparison(t *testing.T) {
	path := os.Getenv("EVENTFRAME_NOTIFICATION_COMPARISON_ARTIFACT")
	if path == "" {
		t.Skip("opt-in adapter comparison")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range []string{"internal/service/research_learning_notification_comparison_test.go", "internal/service/research_learning_notified_test.go", "internal/service/research_learning_adapter_test.go", "internal/researchmemory/completion.go", "internal/service/research_feedback.go", "internal/service/research_frontier.go", "internal/researchmemory/background.go", "internal/researchmemory/adapter.go", "internal/researchpublication/publication.go", "internal/researchpublicationstore/store.go", "internal/researchpublicationstore/mutations.go", "internal/researchpublicationstore/capabilities.go", "docs/experiments/mmm-notification-comparison-v8-protocol.md"} {
		b, e := os.ReadFile("../../" + name)
		if e != nil {
			t.Fatal(e)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if err = enc.Encode(map[string]any{"kind": "header", "Sources": sources, "Hashes": hashes, "expected_arms": 36}); err != nil {
		t.Fatal(err)
	}
	for _, capacity := range []int{16, 64} {
		for trial := 0; trial < 6; trial++ {
			results := map[string]learningAdapterResult{}
			for order := 0; order < 3; order++ {
				kind := []string{"off", "notified", "polling"}[(trial+order)%3]
				var r learningAdapterResult
				switch kind {
				case "off":
					r = learningAdapterArm(t, trial, 0, 192)
				case "polling":
					r = learningAdapterArm(t, trial, capacity, 192)
				case "notified":
					r = learningAdapterResult(learningNotifiedArm(t, trial, capacity, 192))
				}
				if err = enc.Encode(map[string]any{"Mode": kind, "Queue": capacity, "Result": r}); err != nil {
					t.Fatal(err)
				}
				if err = f.Sync(); err != nil {
					t.Fatal(err)
				}
				results[kind] = r
				t.Logf("trial%d %s admitted%d p99_ms%.3f errors%d", trial, kind, r.Admitted, float64(r.P99)/1e6, len(r.Errors))
			}
			off := results["off"]
			for _, kind := range []string{"notified", "polling"} {
				r := results[kind]
				ages := append([]int64(nil), r.AgeNS...)
				sort.Slice(ages, func(i, j int) bool { return ages[i] < ages[j] })
				age := int64(0)
				if len(ages) > 0 {
					age = ages[int(math.Ceil(.95*float64(len(ages))))-1]
				}
				if len(off.Errors)+len(r.Errors) > 0 || off.Overlap == 0 || r.Overlap == 0 || float64(r.P99) > 1.1*float64(off.P99) || r.Admitted < 154 || r.Completed != 50*r.Admitted || r.Failed != 0 || age > 250000000 {
					t.Errorf("trial%d %s failed frozen screen age_ms%.3f", trial, kind, float64(age)/1e6)
				}
			}
		}
	}
}
