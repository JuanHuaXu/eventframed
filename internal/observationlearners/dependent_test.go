package observationlearners

import (
	"compress/gzip"
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"reflect"
	"testing"
)

func TestDependentArtifactReplay(t *testing.T) {
	path := os.Getenv("EVENTFRAME_DEPENDENT_ARTIFACT")
	if path == "" {
		t.Skip("explicit research replay artifact required")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	var artifact struct{ Records []DependentRecord }
	if err := json.NewDecoder(gz).Decode(&artifact); err != nil {
		t.Fatal(err)
	}
	got, err := RunDependent()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(artifact.Records) {
		t.Fatal("record count")
	}
	for i := range got {
		got[i].FitNS, got[i].TreeNS = 0, 0
		artifact.Records[i].FitNS, artifact.Records[i].TreeNS = 0, 0
		if !reflect.DeepEqual(got[i], artifact.Records[i]) {
			t.Fatalf("record %d replay mismatch", i)
		}
	}
}

func TestDependentGenerators(t *testing.T) {
	for g := 0; g < 3; g++ {
		r := rand.New(rand.NewSource(77))
		ones, disagree := 0, 0
		for i := 0; i < 100000; i++ {
			x := dependentInput(r, g)
			if x >= 512 {
				t.Fatal("invalid input")
			}
			if x&1 != 0 {
				ones++
			}
			if (x&1 != 0) != (x&2 != 0) {
				disagree++
			}
		}
		mean, want := float64(ones)/100000, .5
		if g == 1 {
			want = .2
		}
		if math.Abs(mean-want) > .01 {
			t.Fatalf("generator%d mean%g", g, mean)
		}
		want = []float64{.5, .68, .18}[g]
		if math.Abs(float64(disagree)/100000-want) > .01 {
			t.Fatalf("generator%d dependence%d", g, disagree)
		}
	}
}

func TestDependentFairParity(t *testing.T) {
	base, err := Base(Scenarios[2], 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	old, err := RunRetainedStream(base, "test", 2, 0, 0, 2026104202)
	if err != nil {
		t.Fatal(err)
	}
	got, err := RunDependentStream(base, "test", 2, 0, 0, 2026104202, 0)
	if err != nil {
		t.Fatal(err)
	}
	old.FitNS, old.TreeNS, got.FitNS, got.TreeNS = 0, 0, 0, 0
	if !reflect.DeepEqual(old, got.RetainedRecord) {
		t.Fatal("fair control changed v8 policy")
	}
	rng := rand.New(rand.NewSource(Seed(2026104202, 2, 0, 0, 0)))
	for i, tick := range got.Ticks {
		x := dependentInput(rng, 0)
		if x != got.Inputs[i] || truth(x, i, Scenarios[2], rng) != tick.Outcome {
			t.Fatal("input/label replay")
		}
		for a := 0; a < 4; a++ {
			p := tick.Predictions[a].P
			if math.IsNaN(p) || p <= 0 || p >= 1 || got.Views[i][a].Cost > 6 {
				t.Fatal("forecast/view bounds")
			}
		}
		for _, origin := range tick.Delivered {
			if origin > i {
				t.Fatal("future outcome")
			}
		}
	}
	if _, err := RunDependentStream(base, "test", 2, 0, 0, 1, 3); err == nil {
		t.Fatal("invalid generator accepted")
	}
}
