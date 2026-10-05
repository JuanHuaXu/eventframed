package observationlearners

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"
)

type researchTruthManifest struct {
	Mode              string
	OriginalPath      string
	OriginalSHA256    string
	GeneratedPath     string
	GeneratedSHA256   string
	OverlayPath       string
	ProtocolSHA256    string
	BuilderSHA256     string
	ReplacementCounts []struct {
		OldText string
		Count   int
	}
}

func TestResearchTruthRetained(t *testing.T) {
	manifestPath := os.Getenv("EVENTFRAME_TRUTH_MANIFEST")
	fullPath := os.Getenv("EVENTFRAME_TRUTH_FULL")
	summaryPath := os.Getenv("EVENTFRAME_TRUTH_SUMMARY")
	if manifestPath == "" && fullPath == "" && summaryPath == "" {
		t.Skip("opt-in outcome-family experiment")
	}
	if manifestPath == "" || fullPath == "" || summaryPath == "" {
		t.Fatal("manifest, full journal, and summary paths must all be set")
	}
	for _, path := range []string{fullPath, summaryPath} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("output already exists or cannot be checked: %s", path)
		}
	}
	b, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest researchTruthManifest
	if err := json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Mode != "uniform" && manifest.Mode != "latent" {
		t.Fatal("undeclared mode")
	}
	if manifest.OriginalSHA256 != "ecbc832eb2ea3d71957fe0ecd0da226325c60c5b4e247cc7ab245ddcdda86370" || len(manifest.ReplacementCounts) != 5 {
		t.Fatal("source contract mismatch")
	}
	for i, count := range []int{2, 1, 1, 1, 1} {
		if manifest.ReplacementCounts[i].Count != count {
			t.Fatal("unexpected replacement count")
		}
	}
	for _, pair := range [][2]string{{manifest.OriginalPath, manifest.OriginalSHA256}, {manifest.GeneratedPath, manifest.GeneratedSHA256}, {"../../docs/experiments/mmm-retained-truth-v10-protocol.md", manifest.ProtocolSHA256}} {
		got, err := researchHash(pair[0])
		if err != nil || got != pair[1] {
			t.Fatalf("source hash mismatch %s: %s, %v", pair[0], got, err)
		}
	}
	started := time.Now()
	out, err := RunRetained()
	if err != nil {
		t.Fatal(err)
	}
	runWall := time.Since(started)
	if len(out.Records) != 480 || len(out.Verdicts) != 3 {
		t.Fatalf("unexpected run size: %d records, %d verdicts", len(out.Records), len(out.Verdicts))
	}
	var checkedPredictions, checkedDeliveries int
	for _, record := range out.Records {
		if len(record.Ticks) != 512 || len(record.Views) != 512 || len(record.Inner) != 512 {
			t.Fatal("incomplete forecast journal")
		}
		post := 256
		found := false
		for _, scenario := range Scenarios {
			if scenario.Name == record.Scenario {
				found = true
				if scenario.Change < 512 {
					post = scenario.Change
				}
			}
		}
		if !found {
			t.Fatal("unknown scenario")
		}
		var auditDelivered, missing int
		for clock, tick := range record.Ticks {
			if tick.Missing {
				missing++
			}
			for arm := 0; arm < 4; arm++ {
				p := tick.Predictions[arm].P
				if !(p > 0 && p < 1) || math.IsNaN(p) {
					t.Fatalf("invalid pre-outcome probability at %d: %g", clock, p)
				}
				checkedPredictions++
			}
			for _, origin := range tick.Delivered {
				if origin < 0 || origin > clock || record.Ticks[origin].Missing {
					t.Fatalf("future or missing feedback at clock %d, origin %d", clock, origin)
				}
				if record.Ticks[origin].Audit {
					auditDelivered++
				}
				checkedDeliveries++
			}
		}
		if record.Available+record.Pending+missing != 512 || auditDelivered != record.Audits {
			t.Fatalf("availability or audit accounting: %+v", record.Record)
		}
		for arm := 0; arm < 4; arm++ {
			researchCheckMetrics(t, record.Full[arm], researchScore(record.Ticks, arm, 0))
			researchCheckMetrics(t, record.Post[arm], researchScore(record.Ticks, arm, post))
		}
	}
	checksSHA256, err := researchHash("research_covariates_run_test.go")
	if err != nil {
		t.Fatal(err)
	}
	out.Hashes = map[string]string{
		"original retained.go":  manifest.OriginalSHA256,
		"generated retained.go": manifest.GeneratedSHA256,
		"outcome protocol":      manifest.ProtocolSHA256,
		"overlay builder":       manifest.BuilderSHA256,
		"score-check helpers":   checksSHA256,
	}
	if err := researchWriteJSON(fullPath, out, true); err != nil {
		t.Fatal(err)
	}
	for i := range out.Records {
		out.Records[i].Ticks = nil
		out.Records[i].Views = nil
		out.Records[i].Inner = nil
	}
	if err := researchWriteJSON(summaryPath, out, false); err != nil {
		t.Fatal(err)
	}
	t.Logf("mode=%s records=%d predictions=%d deliveries=%d wall=%s verdicts=%+v", manifest.Mode, len(out.Records), checkedPredictions, checkedDeliveries, runWall, out.Verdicts)
}
