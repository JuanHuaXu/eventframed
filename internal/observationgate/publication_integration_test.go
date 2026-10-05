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

type publicationRecord struct {
	Phase, Case        string
	Index              int
	Seed, TrainSeed    int64
	Immediate, Delayed publicationResult
}

func TestPublicationExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_PUBLICATION")
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
	paths = append(paths, "../../docs/experiments/mmm-publication-v1-contract.md", "../../research/publication-summary.mjs", "../../go.mod", "../../go.sum")
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
	seeds := [2]int64{2027010351, 2027010451}
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
				i, err := publicationRun(base, cfg, ci, index, seed, false)
				if err != nil {
					t.Fatal(err)
				}
				d, err := publicationRun(base, cfg, ci, index, seed, true)
				if err != nil {
					t.Fatal(err)
				}
				if err := enc.Encode(publicationRecord{label, cfg.Name, index, seed, train, i, d}); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestPublicationIntegrationContracts(t *testing.T) {
	for ci, cfg := range forestDelayCases()[:2] {
		base, err := forestDependenceBase(cfg, 2026092191+int64(ci))
		if err != nil {
			t.Fatal(err)
		}
		for _, delayed := range []bool{false, true} {
			ref, err := forestDelayRun(base, cfg, ci, 0, 2026092192, delayed)
			if err != nil {
				t.Fatal(err)
			}
			got, err := publicationRun(base, cfg, ci, 0, 2026092192, delayed)
			if err != nil {
				t.Fatal(err)
			}
			if got.Arms[0] != ref.Arms[0] || got.Applied[0] != ref.Applied[0] || got.Stale[0] != ref.Stale[0] || got.Censored[0] != ref.Censored[0] || len(got.Fits) != len(ref.Fits) {
				t.Fatal("control drift")
			}
			for n, f := range got.Frames {
				if f.Predictions[0] != ref.Frames[n].Predictions[0] {
					t.Fatal("forecast drift")
				}
			}
			for n, f := range got.Fits {
				if f.Clock != ref.Fits[n].Clock || len(f.Train)+len(f.Validation) != len(f.Origins) || len(f.Validation) != 16 {
					t.Fatal("partition/cadence")
				}
				if f.Train[len(f.Train)-1] >= f.Validation[0] {
					t.Fatal("partition overlap")
				}
				if f.Clock+1 < 512 {
					next := got.Frames[f.Clock+1]
					for k, w := range next.Predictions[2].Weights {
						if w != f.Weights[1][k] {
							t.Fatal("warm publication mismatch")
						}
					}
				}
			}
		}
	}
}
