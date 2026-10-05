package observationlearners

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"time"
)

type researchCovariateManifest struct {
	Mode            string
	OriginalPath    string
	OriginalSHA256  string
	GeneratedPath   string
	GeneratedSHA256 string
	OverlayPath     string
	ProtocolSHA256  string
	BuilderSHA256   string
	ReplacedSites   int
}

func researchHash(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func researchScore(ticks []Tick, arm, start int) Metrics {
	m := Metrics{N: len(ticks) - start}
	for _, tick := range ticks[start:] {
		p := tick.Predictions[arm].P
		y := 0.0
		if tick.Outcome {
			y = 1
			m.LogLoss -= math.Log(p)
		} else {
			m.LogLoss -= math.Log1p(-p)
		}
		m.Brier += (p - y) * (p - y)
		if (p >= 0.5) == tick.Outcome {
			m.Accuracy++
		} else if p <= 0.1 || p >= 0.9 {
			m.ConfidentErrors++
		}
	}
	m.Brier /= float64(m.N)
	m.LogLoss /= float64(m.N)
	m.Accuracy /= float64(m.N)
	return m
}

func researchCheckMetrics(t *testing.T, got, want Metrics) {
	t.Helper()
	if got.N != want.N || got.ConfidentErrors != want.ConfidentErrors ||
		math.Abs(got.Brier-want.Brier) > 1e-12 ||
		math.Abs(got.LogLoss-want.LogLoss) > 1e-12 ||
		math.Abs(got.Accuracy-want.Accuracy) > 1e-12 {
		t.Fatalf("independent score mismatch: got %+v want %+v", got, want)
	}
}

func researchWriteJSON(path string, value any, compressed bool) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	var writeErr error
	if compressed {
		gz := gzip.NewWriter(f)
		writeErr = json.NewEncoder(gz).Encode(value)
		if err := gz.Close(); writeErr == nil {
			writeErr = err
		}
	} else {
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		writeErr = enc.Encode(value)
	}
	if err := f.Close(); writeErr == nil {
		writeErr = err
	}
	return writeErr
}

func TestResearchCovariateRetained(t *testing.T) {
	manifestPath := os.Getenv("EVENTFRAME_COVARIATE_MANIFEST")
	fullPath := os.Getenv("EVENTFRAME_COVARIATE_FULL")
	summaryPath := os.Getenv("EVENTFRAME_COVARIATE_SUMMARY")
	if manifestPath == "" && fullPath == "" && summaryPath == "" {
		t.Skip("opt-in covariate-generator experiment")
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
	var manifest researchCovariateManifest
	if err := json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Mode != "biased" && manifest.Mode != "latent" {
		t.Fatal("undeclared covariate mode")
	}
	if manifest.ReplacedSites != 2 || manifest.OriginalSHA256 != "ecbc832eb2ea3d71957fe0ecd0da226325c60c5b4e247cc7ab245ddcdda86370" {
		t.Fatal("source contract mismatch")
	}
	for _, pair := range [][2]string{{manifest.OriginalPath, manifest.OriginalSHA256}, {manifest.GeneratedPath, manifest.GeneratedSHA256}, {"../../docs/experiments/mmm-retained-covariates-v9-protocol.md", manifest.ProtocolSHA256}} {
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
	out.Hashes = map[string]string{
		"original retained.go":  manifest.OriginalSHA256,
		"generated retained.go": manifest.GeneratedSHA256,
		"covariate protocol":    manifest.ProtocolSHA256,
		"overlay builder":       manifest.BuilderSHA256,
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
	for _, verdict := range out.Verdicts {
		if verdict.Arm == "" {
			t.Fatal(fmt.Errorf("empty arm verdict"))
		}
	}
}
