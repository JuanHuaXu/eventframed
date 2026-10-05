package researchswitch

import (
	"bufio"
	"encoding/json"
	"os"
	"runtime"
	"testing"
)

func TestScoredV51Allocation(t *testing.T) {
	input, output := os.Getenv("EVENTFRAME_SCORED_V51_FIXTURE"), os.Getenv("EVENTFRAME_SCORED_V51_ALLOCATION")
	if input == "" || output == "" {
		t.Skip("explicit allocation input/output required")
	}
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := json.NewDecoder(bufio.NewReader(f))
	var manifest map[string]any
	var fixture studyFixture
	if err := d.Decode(&manifest); err != nil {
		t.Fatal(err)
	}
	if err := d.Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	bytes := map[string]uint64{}
	for _, style := range append([]string{"legacy"}, scoreModesV51...) {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		cfg := Config{Prior: []float64{.8, .1, .1}, Hazard: 1. / 16, Trials: 2400, Pending: 2400}
		if style == "legacy" {
			p, err := NewMemoHybridV49(fixture.World.Population.Base, cfg, 16, 1./2400, 1./150, 1)
			if err != nil {
				t.Fatal(err)
			}
			runtime.ReadMemStats(&after)
			runtime.KeepAlive(p)
		} else {
			p, err := NewScoredHybridV51(fixture.World.Population.Base, cfg, 16, 1./2400, 1./150, 1, style)
			if err != nil {
				t.Fatal(err)
			}
			runtime.ReadMemStats(&after)
			runtime.KeepAlive(p)
		}
		bytes[style] = after.TotalAlloc - before.TotalAlloc
	}
	result, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer result.Close()
	if err := json.NewEncoder(result).Encode(map[string]any{"constructorBytes": bytes, "rssBound": false, "servingLatencyClaim": false}); err != nil {
		t.Fatal(err)
	}
	if err := result.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Log("constructor allocated bytes, not RSS", bytes)
}
