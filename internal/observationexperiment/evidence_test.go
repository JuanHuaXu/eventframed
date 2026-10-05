package observationexperiment

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

func TestPublishedPilotEvidenceRecomputes(t *testing.T) {
	root := filepath.Join("..", "..")
	f, err := os.Open(filepath.Join(root, "docs/experiments/mmm-attention-v1.json.gz"))
	if os.IsNotExist(err) {
		t.Skip("pilot not run yet")
	}
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	var out Output
	if err := json.NewDecoder(gz).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.Rows) != 2*7*Trials*8 {
		t.Fatal("incomplete pilot")
	}
	for path, expected := range out.Hashes {
		content, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(content)
		if hex.EncodeToString(h[:]) != expected {
			t.Fatalf("source/protocol changed: %s", path)
		}
	}
	type rowKey struct {
		Split, Family, Policy string
		Trial                 int
	}
	keys := map[rowKey]bool{}
	for _, r := range out.Rows {
		key := rowKey{r.Split, r.Family, r.Policy, r.Trial}
		if keys[key] {
			t.Fatal("duplicate prediction")
		}
		keys[key] = true
		truth := 0.
		if r.Outcome {
			truth = 1
		}
		p := r.Result.Probability
		if p <= 0 || p >= 1 || math.IsNaN(p) {
			t.Fatal("invalid probability")
		}
		brier := (p - truth) * (p - truth)
		loss := -math.Log1p(-p)
		if r.Outcome {
			loss = -math.Log(p)
		}
		if math.Abs(brier-r.Brier) > 1e-14 || math.Abs(loss-r.LogLoss) > 1e-14 || r.Correct != ((p >= .5) == r.Outcome) {
			t.Fatal("incorrect score")
		}
		cap := 6
		if r.Policy == "exhaustive" {
			cap = 9
		}
		if r.Result.Cost > cap || r.Result.Observed > r.Result.Cost {
			t.Fatal("budget breached")
		}
	}
	copy := Output{Rows: out.Rows}
	Summarize(&copy)
	if !reflect.DeepEqual(copy.Summaries, out.Summaries) || !reflect.DeepEqual(copy.Comparisons, out.Comparisons) || copy.PrimaryPassed != out.PrimaryPassed || copy.AnyConfirmationHarm != out.AnyConfirmationHarm {
		t.Fatal("summary mismatch")
	}
	// Independently regenerate every trajectory and replay every policy. No
	// timing fields enter the result, so full deterministic equality is possible.
	replayed, err := Run()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(replayed.Rows, out.Rows) {
		t.Fatal("full replay mismatch")
	}
}
