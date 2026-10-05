package researchdispersion

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// This NEW test-only bridge reuses the immutable original generator/control
// functions. It never imports the new switching candidate (avoids a cycle).
func hybridCohortV48(split string) (int64, int, error) {
	switch split {
	case "diagnostic":
		return 2026104807, 1, nil
	case "design":
		return 2026104809, 16, nil
	case "confirmation":
		return 2026104811, 16, nil
	}
	return 0, 0, fmt.Errorf("invalid switch cohort")
}

func TestSpecialistActualSeedSeparationV48(t *testing.T) {
	seen := map[int64]bool{}
	for _, base := range []int64{2026103903, 2026103904, 2026104001, 2026104003, 2026104004, 2026104101, 2026104103, 2026104104, 2026104307, 2026104309, 2026104311, 2026104707, 2026104709, 2026104711, 2026104807, 2026104809, 2026104811} {
		for g := 0; g < 2; g++ {
			for r := range regimesV39 {
				for id := 0; id < 16; id++ {
					seed := makeV39(base, g, r, id).Seed
					if seen[seed] {
						t.Fatal("actual world-seed collision", seed)
					}
					seen[seed] = true
				}
			}
		}
	}
}

func TestSpecialistFixtureV48(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_HYBRID_V48_FIXTURE"), os.Getenv("EVENTFRAME_HYBRID_V48_SPLIT")
	if path == "" {
		t.Skip("explicit fixture output required")
	}
	base, n, err := hybridCohortV48(split)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b := bufio.NewWriter(f)
	enc := json.NewEncoder(b)
	if err := enc.Encode(map[string]any{"Kind": "fixture_manifest", "Split": split, "SeedBase": base, "Worlds": 28 * n, "IndependentlyAuditedControls": true}); err != nil {
		t.Fatal(err)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesV39 {
			for id := 0; id < n; id++ {
				w := orientationWorldV38{Population: makeV39(base, g, r, id)}
				due := make([][][]int, 3)
				for s, schedule := range schedulesV39 {
					due[s], err = scheduleV36(delayedWorldV36{Seed: w.Population.Seed, Outcomes: w.Population.Outcomes}, schedule)
					if err != nil {
						t.Fatal(err)
					}
					for _, mode := range []string{"full", "adaptive", "rich_moment2"} {
						a, err := runMomentV41(w.Population, mode, schedule)
						if err != nil {
							t.Fatal(err)
						}
						scoreWindowV37(w.Population, &a.windowArmV37)
						if mode == "rich_moment2" {
							err = independentMomentV41(w.Population, a)
							if err == nil {
								err = independentPriorMetricsV43(w.Population, a)
							}
						} else {
							err = independentOrientationV38(w.Population, a)
						}
						if err != nil {
							t.Fatal("independent original control audit", err)
						}
						w.Arms = append(w.Arms, a)
					}
				}
				if err := enc.Encode(map[string]any{"World": w, "Due": due}); err != nil {
					t.Fatal(err)
				}
			}
			t.Log([]string{"tight", "wide"}[g] + "/" + regimesV39[r])
		}
	}
	if err := b.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}
