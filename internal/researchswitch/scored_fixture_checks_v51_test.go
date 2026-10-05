package researchswitch

import (
	"bufio"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestScoredV51StreamIdentity(t *testing.T) {
	for _, style := range append([]string{"legacy"}, scoreModesV51...) {
		for _, bad := range []string{"", "wrong", "legacy", "log_mean", "brier_strong"} {
			if bad == style {
				continue
			}
			raw := `{"Kind":"hybrid_study_manifest","Worlds":1,"Style":"` + bad + `"}`
			if _, err := auditScoredStreamV51(strings.NewReader(`{}`), strings.NewReader(raw), style); err == nil {
				t.Fatal("incorrect style admitted", style, bad)
			}
		}
	}
}

func TestScoredV51FixtureChecks(t *testing.T) {
	input, output := os.Getenv("EVENTFRAME_SCORED_V51_FIXTURE"), os.Getenv("EVENTFRAME_SCORED_V51_CHECKS")
	if input == "" || output == "" {
		t.Skip("explicit fixture and check output required")
	}
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := json.NewDecoder(bufio.NewReader(f))
	var manifest map[string]any
	var fixture studyFixture
	if err = d.Decode(&manifest); err != nil {
		t.Fatal(err)
	}
	if err = d.Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if manifest["SeedBase"] != float64(2026105107) || manifest["Worlds"] != float64(28) {
		t.Fatal("fresh diagnostic identity")
	}
	futureChecks, corruptionChecks := 0, 0
	const cut = 1200
	for _, style := range scoreModesV51 {
		for _, mode := range hybridModesV48 {
			for s := 0; s < 3; s++ {
				original, err := collectScoredV51(fixture, mode, s, style)
				if err != nil {
					t.Fatal(err)
				}
				fork := fixture
				fork.World.Population.Outcomes = make([][]bool, len(fixture.World.Population.Outcomes))
				for j, row := range fixture.World.Population.Outcomes {
					fork.World.Population.Outcomes[j] = append([]bool(nil), row...)
				}
				flipped := 0
				for tick := cut; tick < len(fixture.Due[s]); tick++ {
					for _, trial := range fixture.Due[s][tick] {
						fork.World.Population.Outcomes[trial/150][trial%150] = !fork.World.Population.Outcomes[trial/150][trial%150]
						flipped++
					}
				}
				if flipped == 0 {
					t.Fatal("vacuous future fork")
				}
				changed, err := collectScoredV51(fork, mode, s, style)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(original.Issued[:cut], changed.Issued[:cut]) || !reflect.DeepEqual(original.Advice[:cut], changed.Advice[:cut]) || !reflect.DeepEqual(original.Heads[:cut], changed.Heads[:cut]) {
					t.Fatal("future outcomes alter earlier issued laws", style, mode, s)
				}
				for j, snapshot := range original.Snapshots {
					if snapshot.Tick < cut && !reflect.DeepEqual(snapshot, changed.Snapshots[j]) {
						t.Fatal("future outcomes alter earlier snapshot")
					}
				}
				for j, receipt := range original.Receipts {
					if receipt.ArrivedAt < cut && receipt != changed.Receipts[j] {
						t.Fatal("future outcomes alter earlier receipt")
					}
				}
				if reflect.DeepEqual(original.Snapshots[len(original.Snapshots)-1].Forecast, changed.Snapshots[len(changed.Snapshots)-1].Forecast) {
					t.Fatal("revealed contrary labels did not change final law")
				}
				futureChecks++
				if mode != "round" || s != 2 {
					continue
				}
				if err := auditScoredV51(fixture, original, s, style); err != nil {
					t.Fatal("positive audit", err)
				}
				corrupt := []func(*hybridArmV48){
					func(a *hybridArmV48) { a.Issued[0] += .01 },
					func(a *hybridArmV48) { a.Advice[0][0] += .01 },
					func(a *hybridArmV48) { a.Heads[0][0] += .01 },
					func(a *hybridArmV48) { a.Receipts[0].Forecast += .01 },
					func(a *hybridArmV48) { a.Receipts[0].TrialOrdinal++ },
					func(a *hybridArmV48) { a.Receipts[0].Epoch++ },
					func(a *hybridArmV48) { a.Receipts[0].Useful = !a.Receipts[0].Useful },
					func(a *hybridArmV48) { a.Receipts[0].ArrivedAt++ },
					func(a *hybridArmV48) { a.Receipts = a.Receipts[:len(a.Receipts)-1] },
					func(a *hybridArmV48) { a.Snapshots[0].ScopeWeights[0] += .01 },
					func(a *hybridArmV48) { a.Snapshots[0].GlobalWeights[0] += .01 },
					func(a *hybridArmV48) { a.Snapshots[0].Weights[0][0] += .01 },
					func(a *hybridArmV48) { a.Snapshots[0].Forecast[0] += .01 },
					func(a *hybridArmV48) { a.Snapshots = a.Snapshots[:len(a.Snapshots)-1] },
					func(a *hybridArmV48) { a.Costs.AccountedNS++ },
					func(a *hybridArmV48) { a.PeakPending++ },
					func(a *hybridArmV48) { a.Schedule = "immediate" },
				}
				b, err := json.Marshal(original)
				if err != nil {
					t.Fatal(err)
				}
				for id, mutate := range corrupt {
					var bad hybridArmV48
					if err := json.Unmarshal(b, &bad); err != nil {
						t.Fatal(err)
					}
					mutate(&bad)
					if err := auditScoredV51(fixture, bad, s, style); err == nil {
						t.Fatal("semantic corruption accepted", style, id)
					}
					corruptionChecks++
				}
			}
		}
	}
	out, err := json.MarshalIndent(map[string]any{"futurePrefixCases": futureChecks, "semanticCorruptionsRejected": corruptionChecks, "fourStyles": scoreModesV51, "allThreePoliciesAndSchedules": true, "revealedForkIsNonvacuous": true, "qualityClaim": false}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	result, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer result.Close()
	if _, err := result.Write(append(out, '\n')); err != nil {
		t.Fatal(err)
	}
	if err := result.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Log("future prefixes", futureChecks, "semantic corruptions", corruptionChecks)
}
