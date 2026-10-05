package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
	"os"
	"path/filepath"
	"testing"
)

func TestArrivalBreadthContracts(t *testing.T) {
	for ci, cfg := range forestDependenceCases() {
		base, err := forestDependenceBase(cfg, observationpreserved.Seed(2026092231, ci, 0, 4))
		if err != nil {
			t.Fatal(err)
		}
		for _, delay := range []bool{false, true} {
			want, err := forestDelayRun(base, cfg, ci, 0, 2026092231, delay)
			if err != nil {
				t.Fatal(err)
			}
			got, err := innerArrivalRun(base, cfg, ci, 0, 2026092231, delay)
			if err != nil {
				t.Fatal(err)
			}
			if got.Arms[0].Full != want.Arms[0].Full || got.Arms[0].Post != want.Arms[0].Post || got.Arms[0].SplitAt != want.Arms[0].SplitAt {
				t.Fatal("broader control drift", cfg.Name)
			}
			for i, f := range got.Frames {
				if f.Predictions[0] != want.Frames[i].Predictions[0] {
					t.Fatal("forecast drift", cfg.Name, i)
				}
			}
		}
	}
}

func TestArrivalBreadthExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ARRIVAL_BREADTH")
	if path == "" {
		t.Skip("opt-in consumed breadth")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	paths, err := filepath.Glob("../../internal/*/*.go")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, "../../go.mod", "../../go.sum", "../../docs/experiments/mmm-arrival-breadth-v1-contract.md", "../../research/arrival-breadth-summary.mjs")
	sources, hashes := map[string]string{}, map[string]string{}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		sources[p[6:]] = string(b)
		hashes[p[6:]] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(map[string]any{"Sources": sources, "Hashes": hashes, "Consumed": true, "Seeds": [2]int64{2026092231, 2026092331}, "PerCell": 16}); err != nil {
		t.Fatal(err)
	}
	for phase, label := range []string{"design", "confirmation"} {
		for ci, cfg := range forestDependenceCases() {
			for index := 0; index < 16; index++ {
				seed := int64(2026092231 + 100*phase)
				train := observationpreserved.Seed(seed, ci, index, 4)
				base, err := forestDependenceBase(cfg, train)
				if err != nil {
					t.Fatal(err)
				}
				i, err := innerArrivalRun(base, cfg, ci, index, seed, false)
				if err != nil {
					t.Fatal(err)
				}
				d, err := innerArrivalRun(base, cfg, ci, index, seed, true)
				if err != nil {
					t.Fatal(err)
				}
				if err := enc.Encode(innerArrivalRecord{label, cfg.Name, index, seed, train, i, d}); err != nil {
					t.Fatal(err)
				}
			}
			if err := f.Sync(); err != nil {
				t.Fatal(err)
			}
			t.Log(label, cfg.Name, "complete")
		}
	}
}
