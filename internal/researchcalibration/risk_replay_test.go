package researchcalibration

import (
	"bufio"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// Post-collection check: regenerate every policy's exact Go RNG trace from
// archived seeds/tapes without changing the frozen generator or source hashes.
func TestReplayRiskV28(t *testing.T) {
	path := os.Getenv("EVENTFRAME_RISK_V28_REPLAY")
	if path == "" {
		t.Skip("opt-in exact research replay")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	scan.Buffer(make([]byte, 65536), 4<<20)
	if !scan.Scan() {
		t.Fatal("missing manifest")
	}
	worlds, arms := 0, 0
	for scan.Scan() {
		var world worldV27
		if err = json.Unmarshal(scan.Bytes(), &world); err != nil {
			t.Fatal(err)
		}
		for _, want := range world.Arms {
			var got armV27
			if want.Policy == "local" {
				got = localV27(world.Base, world.Labels)
			} else {
				got = runV28(world.Base, world.Labels, want.Policy, want.Setting, want.LabelCost, world.Seed)
			}
			scoreV27(&got, world.Rates)
			got.SetupNS, got.SelectionNS, got.UpdateNS, got.FinalNS, got.TotalNS = 0, 0, 0, 0, 0
			want.SetupNS, want.SelectionNS, want.UpdateNS, want.FinalNS, want.TotalNS = 0, 0, 0, 0, 0
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("replay mismatch %s/%s/%d %s/%s", world.Geometry, world.Regime, world.World, want.Setting, want.Policy)
			}
			arms++
		}
		worlds++
	}
	if err = scan.Err(); err != nil {
		t.Fatal(err)
	}
	if worlds != 384 || arms != 6528 {
		t.Fatalf("incomplete replay: %d worlds %d arms", worlds, arms)
	}
}
