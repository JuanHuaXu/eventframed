package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

func forestDelayCases() []forestDependenceCase {
	return []forestDependenceCase{
		{subsetBreadthCase: subsetBreadthCase{Name: "copy_bit", Mode: "member_shift", Target: "bit"}, Pattern: 6},
		{subsetBreadthCase: subsetBreadthCase{Name: "copy_xor2", Mode: "member_shift", Target: "xor2"}, Pattern: 6},
		{subsetBreadthCase: subsetBreadthCase{Name: "noise10_bit", Mode: "member_shift", Target: "bit"}, Pattern: 1},
		{subsetBreadthCase: subsetBreadthCase{Name: "reverses_xor2", Mode: "member_shift", Target: "xor2"}, Pattern: 5},
		{subsetBreadthCase: subsetBreadthCase{Name: "stable_noise10", Mode: "stable", Target: "bit"}, Pattern: 1},
		{subsetBreadthCase: subsetBreadthCase{Name: "null_uniform", Mode: "null", Target: "bit"}, Pattern: 0},
	}
}

func TestForestDelayContracts(t *testing.T) {
	for ci, cfg := range forestDelayCases()[:2] {
		base, err := forestDependenceBase(cfg, 2026092191+int64(ci))
		if err != nil {
			t.Fatal(err)
		}
		c, f, j := newSubsetTrial(base, 1), newSubsetTrial(base, 1), newSubsetTrial(base, 1)
		if _, err := forestDependenceRun(base, "unit", ci, 0, 2026092192, c, cfg, f, j); err != nil {
			t.Fatal(err)
		}
		got, err := forestDelayRun(base, cfg, ci, 0, 2026092192, false)
		if err != nil {
			t.Fatal(err)
		}
		for a, ref := range []*subsetTrial{c, j} {
			if got.Arms[a].Full != ref.metrics.Full || got.Arms[a].Post != ref.metrics.Post || got.Arms[a].SplitAt != ref.metrics.SplitAt {
				t.Fatalf("immediate parity %s arm%d: %+v / %+v", cfg.Name, a, got.Arms[a], ref.metrics)
			}
		}
		delayed, err := forestDelayRun(base, cfg, ci, 0, 2026092192, true)
		if err != nil {
			t.Fatal(err)
		}
		replay, err := forestDelayRun(base, cfg, ci, 0, 2026092192, true)
		if err != nil || !reflect.DeepEqual(delayed, replay) {
			t.Fatal("delayed replay", err)
		}
		for n, f := range got.Frames {
			d := delayed.Frames[n]
			if f.X != d.X || f.RX != d.RX || f.Y != d.Y || f.RY != d.RY || f.Audit != d.Audit || f.Seed != d.Seed {
				t.Fatal("schedule changed latent tape")
			}
		}
		for _, fit := range delayed.Fits {
			for _, id := range fit.Origins {
				f := delayed.Frames[id]
				if f.Missing || !f.Audit || f.Arrival > fit.Clock {
					t.Fatal("fit leakage")
				}
			}
		}
	}
}

type forestDelayRecord struct {
	Phase, Case        string
	Index              int
	Seed, TrainSeed    int64
	Immediate, Delayed forestDelayResult
}

func TestForestDelayExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_FOREST_DELAY")
	if path == "" {
		t.Skip("opt-in delayed quality")
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
	paths = append(paths, "../../docs/experiments/mmm-forest-delay-v1-contract.md", "../../research/forest-delay-summary.mjs", "../../go.mod", "../../go.sum")
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
	seeds := [2]int64{2027010151, 2027010251}
	seen := map[int64]bool{}
	for _, seed := range seeds {
		for ci := range forestDelayCases() {
			for index := 0; index < 16; index++ {
				for role := 0; role < 8; role++ {
					s := observationpreserved.Seed(seed, ci, index, role)
					if seen[s] {
						t.Fatal("seed collision")
					}
					seen[s] = true
				}
			}
		}
	}
	if err := enc.Encode(map[string]any{"Sources": sources, "Hashes": hashes, "Seeds": seeds, "PerCell": 16}); err != nil {
		t.Fatal(err)
	}
	for phase, label := range []string{"design", "confirmation"} {
		for ci, cfg := range forestDelayCases() {
			for index := 0; index < 16; index++ {
				seed := seeds[phase]
				train := observationpreserved.Seed(seed, ci, index, 4)
				base, err := forestDependenceBase(cfg, train)
				if err != nil {
					t.Fatal(err)
				}
				i, err := forestDelayRun(base, cfg, ci, index, seed, false)
				if err != nil {
					t.Fatal(err)
				}
				d, err := forestDelayRun(base, cfg, ci, index, seed, true)
				if err != nil {
					t.Fatal(err)
				}
				if err := enc.Encode(forestDelayRecord{label, cfg.Name, index, seed, train, i, d}); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
