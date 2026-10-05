package observationlearners

import (
	"compress/gzip"
	"encoding/json"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"math"
	"os"
	"reflect"
	"testing"
)

func TestConditionalObserverParity(t *testing.T) {
	f := NewForest(33)
	for x := 0; x < 512; x++ {
		f.Update(uint16(x), x%3 == 0)
	}
	var w [512]float64
	for x := range w {
		w[x] = 1
	}
	m, e := NewConditionalForest(f, w)
	if e != nil {
		t.Fatal(e)
	}
	for x := uint16(0); x < 512; x++ {
		rd := observationexperiment.Frames(x, "parity")
		a, e := RunForestObserver(f, rd, rd.Epoch())
		if e != nil {
			t.Fatal(e)
		}
		b, e := RunConditionalObserver(m, rd, rd.Epoch())
		if e != nil {
			t.Fatal(e)
		}
		if a.Cost != b.Cost || a.Stop != b.Stop || len(a.Trace) != len(b.Trace) || math.Abs(a.Probability-b.Probability) > 1e-12 {
			t.Fatal("observer parity")
		}
		for i := range a.Trace {
			if a.Trace[i].View != b.Trace[i].View || a.Trace[i].Values != b.Trace[i].Values {
				t.Fatal("acquisition changed")
			}
		}
		if _, e := RunConditionalObserver(m, rd, rd.Epoch()+1); e == nil {
			t.Fatal("stale epoch accepted")
		}
	}
}

func TestAcquisitionControl(t *testing.T) {
	base, e := Base(Scenarios[2], 2, 0)
	if e != nil {
		t.Fatal(e)
	}
	a, e := RunDependentStream(base, "design", 2, 0, 0, 2026105202, 2)
	if e != nil {
		t.Fatal(e)
	}
	b, e := RunAcquisitionStream(base, "design", 2, 0, 0, 2026105202, 2, 0)
	if e != nil {
		t.Fatal(e)
	}
	a.FitNS, a.TreeNS, b.FitNS, b.TreeNS = 0, 0, 0, 0
	if !reflect.DeepEqual(a, b.DependentRecord) {
		t.Fatal("uniform control changed")
	}
	for mode := 1; mode < 3; mode++ {
		c, e := RunAcquisitionStream(base, "design", 2, 0, 0, 2026105202, 2, mode)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(a.Inputs, c.Inputs) {
			t.Fatal("different inputs")
		}
		for step, tick := range c.Ticks {
			if tick.Outcome != a.Ticks[step].Outcome || tick.Audit != a.Ticks[step].Audit {
				t.Fatal("unpaired data")
			}
			if tick.Predictions[0] != a.Ticks[step].Predictions[0] {
				t.Fatal("fixed control changed")
			}
			for arm := 0; arm < 4; arm++ {
				if c.Views[step][arm].Cost > 6 {
					t.Fatal("budget")
				}
			}
		}
	}
}

func TestAcquisitionArtifactReplay(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_ARTIFACT")
	if path == "" {
		t.Skip("explicit artifact required")
	}
	f, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	var a struct{ Records []AcquisitionRecord }
	if e := json.NewDecoder(z).Decode(&a); e != nil {
		t.Fatal(e)
	}
	b, e := RunAcquisition()
	if e != nil {
		t.Fatal(e)
	}
	if len(a.Records) != len(b) {
		t.Fatal("record count")
	}
	for i := range b {
		a.Records[i].FitNS, a.Records[i].TreeNS, b[i].FitNS, b[i].TreeNS = 0, 0, 0, 0
		if !reflect.DeepEqual(a.Records[i], b[i]) {
			t.Fatalf("replay%d", i)
		}
	}
}
