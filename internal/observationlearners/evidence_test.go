package observationlearners

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestEvidenceRecomputesAndReplays(t *testing.T) {
	root := filepath.Join("..", "..")
	f, e := os.Open(filepath.Join(root, "docs/experiments/mmm-learners-v5.json.gz"))
	if os.IsNotExist(e) {
		t.Skip("not run yet")
	}
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	gz, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer gz.Close()
	var out Output
	if e = json.NewDecoder(gz).Decode(&out); e != nil {
		t.Fatal(e)
	}
	if len(out.Records) != 480 {
		t.Fatal("incomplete experiment")
	}
	for path, want := range out.Hashes {
		b, e := os.ReadFile(filepath.Join(root, path))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			t.Fatalf("source changed: %s", path)
		}
	}
	for _, r := range out.Records {
		if len(r.Ticks) != 512 {
			t.Fatal("missing predictions")
		}
		var scenario Scenario
		for _, s := range Scenarios {
			if s.Name == r.Scenario {
				scenario = s
			}
		}
		seen := map[int]bool{}
		audits := 0
		for i, tick := range r.Ticks {
			for _, p := range tick.Predictions {
				if p.P <= 0 || p.P >= 1 || math.IsNaN(p.P) {
					t.Fatal("invalid forecast")
				}
			}
			for _, origin := range tick.Delivered {
				if seen[origin] || origin+scenario.Delay != i || r.Ticks[origin].Missing {
					t.Fatal("availability violation")
				}
				seen[origin] = true
				if r.Ticks[origin].Audit {
					audits++
				}
			}
			if tick.Audits != audits || tick.ForestNodes > 155 {
				t.Fatal("state accounting")
			}
		}
		if r.Available != len(seen) || r.Audits != audits {
			t.Fatal("label accounting")
		}
		post := scenario.Change
		if post >= 512 {
			post = 256
		}
		for arm := 0; arm < 7; arm++ {
			if !reflect.DeepEqual(score(r.Ticks, arm, 0), r.Full[arm]) || !reflect.DeepEqual(score(r.Ticks, arm, post), r.Post[arm]) {
				t.Fatal("score mismatch")
			}
		}
	}
	copy := Output{Records: out.Records}
	Summarize(&copy)
	if !reflect.DeepEqual(copy.Comparisons, out.Comparisons) || !reflect.DeepEqual(copy.Verdicts, out.Verdicts) {
		t.Fatal("verdict mismatch")
	}
	replay, e := Run(nil)
	if e != nil {
		t.Fatal(e)
	}
	for i := range out.Records {
		out.Records[i].FitNS = 0
		replay.Records[i].FitNS = 0
	}
	if !reflect.DeepEqual(out.Records, replay.Records) {
		t.Fatal("replay mismatch")
	}
}
