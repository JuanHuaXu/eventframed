package researchswitch

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"testing"
)

// Paired diagnostic only. Both static policies use alpha=0, the SAME shared
// children, advice, prior, original nominations and arrival sets. The only
// difference is whether expert-state evidence pools globally or per task.
func TestLocalV47GlobalStaticPair(t *testing.T) {
	input, out := os.Getenv("EVENTFRAME_SPECIALIST_V47_FIXTURE"), os.Getenv("EVENTFRAME_SPECIALIST_V47_PAIR")
	if out == "" {
		t.Skip("explicit paired diagnostic required")
	}
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	d := json.NewDecoder(bufio.NewReader(f))
	b := bufio.NewWriter(w)
	e := json.NewEncoder(b)
	var manifest map[string]any
	if err = d.Decode(&manifest); err != nil || manifest["Split"] != "diagnostic" || manifest["Worlds"] != float64(28) || manifest["SeedBase"] != float64(2026104707) {
		t.Fatal("diagnostic manifest", err)
	}
	manifest["Kind"] = "global_static_pair_manifest"
	if err = e.Encode(manifest); err != nil {
		t.Fatal(err)
	}
	count := 0
	for {
		var fixture studyFixture
		if err = d.Decode(&fixture); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		r := studyRecord{Seed: fixture.World.Population.Seed}
		for s := 0; s < 3; s++ {
			a, err := collectStudy(fixture, "static", s)
			if err != nil {
				t.Fatal(err)
			}
			if err = auditStudy(fixture, a, s); err != nil {
				t.Fatal("unchanged independent global reference", err)
			}
			r.Arms = append(r.Arms, a)
		}
		if err = e.Encode(r); err != nil {
			t.Fatal(err)
		}
		count++
		t.Log(fixture.World.Population.Geometry + "/" + fixture.World.Population.Regime)
	}
	if count != 28 {
		t.Fatal("paired coverage")
	}
	if err = b.Flush(); err != nil {
		t.Fatal(err)
	}
	if err = w.Sync(); err != nil {
		t.Fatal(err)
	}
}
