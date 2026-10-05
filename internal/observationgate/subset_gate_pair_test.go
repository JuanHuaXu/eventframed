package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

func TestSubsetGatePairParity(t *testing.T) {
	for scenario := 0; scenario < 5; scenario++ {
		base, err := observationpreserved.Base(scenario == 4)
		if err != nil {
			t.Fatal(err)
		}
		control := newSubsetTrial(base, scenario)
		a, err := subsetIntegrationRun(base, "unit", scenario, 0, 2026092199, control)
		if err != nil {
			t.Fatal(err)
		}
		current, old := newSubsetTrial(base, scenario), newSubsetTrial(base, scenario)
		b, err := subsetGatePairRun(base, "unit", scenario, 0, 2026092199, current, old)
		if err != nil || a != b || current.metrics != control.metrics || hex.EncodeToString(current.hash.Sum(nil)) != hex.EncodeToString(control.hash.Sum(nil)) {
			t.Fatal("control drift", scenario, err)
		}
		if old.metrics.SplitAt != b.Arms[2].SplitAt || current.metrics.SplitAt != b.Arms[3].SplitAt || old.fits != current.fits || old.model != current.model {
			t.Fatal("gate/fit mismatch", scenario)
		}
		if old.state.next != 512 || current.state.next != 512 || base.Support() != 4096 {
			t.Fatal("state accounting")
		}
	}
}

func TestSubsetGatePairExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SUBSET_GATE_PAIR_ARTIFACT")
	if path == "" {
		t.Skip("opt-in paired gate exploratory study")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	paths, err := filepath.Glob("../../internal/*/*.go")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, "../../go.mod", "../../go.sum", "../../docs/experiments/mmm-subset-gate-pair-v1-contract.md")
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		name := filepath.ToSlash(path[6:])
		h := sha256.Sum256(data)
		sources[name] = string(data)
		hashes[name] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(map[string]any{"Kind": "header", "Sources": sources, "Hashes": hashes, "PerCell": 16, "Seeds": [2]int64{2026092111, 2026092112}}); err != nil {
		t.Fatal(err)
	}
	for phase := 0; phase < 2; phase++ {
		for scenario, name := range observationpreserved.Scenarios {
			for index := 0; index < 16; index++ {
				seed := int64(2026092111 + phase)
				fitSeed := observationpreserved.Seed(seed, scenario, index, 4)
				base, fitHash, err := memberFreshBase(name == "null", fitSeed)
				if err != nil {
					t.Fatal(err)
				}
				current, old := newSubsetTrial(base, scenario), newSubsetTrial(base, scenario)
				r, err := subsetGatePairRun(base, []string{"design", "confirmation"}[phase], scenario, index, seed, current, old)
				if err != nil {
					t.Fatal(err)
				}
				if old.metrics.SplitAt != r.Arms[2].SplitAt || current.metrics.SplitAt != r.Arms[3].SplitAt || old.fits != current.fits {
					t.Fatal("pair mismatch")
				}
				if err := enc.Encode(struct {
					Original                      memberRecord
					OldSubset, MixtureSubset      integrationArm
					Fits                          int
					FitSeed                       int64
					FitHash, OldTape, MixtureTape string
				}{r, old.metrics, current.metrics, current.fits, fitSeed, fitHash, hex.EncodeToString(old.hash.Sum(nil)), hex.EncodeToString(current.hash.Sum(nil))}); err != nil {
					t.Fatal(err)
				}
			}
			if err := f.Sync(); err != nil {
				t.Fatal(err)
			}
			t.Log(phase, name, "16 complete")
		}
	}
}
