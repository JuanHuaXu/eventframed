package observationlearners

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"io"
	"os"
	"reflect"
	"testing"
)

type gateEpochReader struct {
	observation.Reader
	changed bool
}

func (r *gateEpochReader) Epoch() uint64 {
	e := r.Reader.Epoch()
	if r.changed {
		e++
	}
	return e
}

func TestResearchStopGate(t *testing.T) {
	samples := make([]observation.Sample, 4096)
	f := NewForest(1)
	var w [512]float64
	for i := range samples {
		samples[i] = observation.Sample{Bits: uint16(i % 512), Outcome: true}
		f.Update(uint16(i%512), true)
	}
	for i := range w {
		w[i] = 1
	}
	m, e := observation.Fit(samples)
	if e != nil {
		t.Fatal(e)
	}
	c, e := NewConditionalForest(f, w)
	if e != nil {
		t.Fatal(e)
	}
	reader := observationexperiment.Frames(0, "gate-test")
	for _, accept := range []bool{false, true} {
		gate := func(mask, values uint16) (bool, error) {
			if values&^mask != 0 {
				t.Fatal("hidden values")
			}
			return accept, nil
		}
		a, e := observation.RunStopGate(m, reader, reader.Epoch(), gate)
		if e != nil {
			t.Fatal(e)
		}
		b, e := RunConditionalStopGate(c, reader, reader.Epoch(), gate)
		if e != nil {
			t.Fatal(e)
		}
		var wantA, wantB observation.Result
		if accept {
			wantA, e = observation.Run(m, reader, reader.Epoch(), "mmm", 0)
			if e != nil {
				t.Fatal(e)
			}
			wantB, e = RunConditionalObserver(c, reader, reader.Epoch())
		} else {
			wantA, e = observation.RunFullBudget(m, reader, reader.Epoch())
			if e != nil {
				t.Fatal(e)
			}
			wantB, e = RunConditionalFullBudget(c, reader, reader.Epoch())
		}
		if e != nil || !reflect.DeepEqual(a, wantA) || !reflect.DeepEqual(b, wantB) {
			t.Fatal("gate equivalence", e)
		}
		if accept {
			if a.Stop != "confidence" || b.Stop != "confidence" {
				t.Fatal("accept ignored")
			}
		} else if a.Cost != 6 || b.Cost != 6 {
			t.Fatal("deny ignored")
		}
	}
	sentinel := errors.New("gate error")
	if _, e := observation.RunStopGate(m, reader, reader.Epoch(), nil); e == nil {
		t.Fatal("nil count gate")
	}
	if _, e := RunConditionalStopGate(c, reader, reader.Epoch(), nil); e == nil {
		t.Fatal("nil conditional gate")
	}
	if _, e := observation.RunStopGate(m, reader, reader.Epoch(), func(uint16, uint16) (bool, error) { return false, sentinel }); !errors.Is(e, sentinel) {
		t.Fatal("gate error lost")
	}
	if _, e := RunConditionalStopGate(c, reader, reader.Epoch(), func(uint16, uint16) (bool, error) { return false, sentinel }); !errors.Is(e, sentinel) {
		t.Fatal("conditional gate error lost")
	}
	for kind := 0; kind < 2; kind++ {
		r := &gateEpochReader{Reader: reader}
		gate := func(uint16, uint16) (bool, error) { r.changed = true; return true, nil }
		if kind == 0 {
			_, e = observation.RunStopGate(m, r, r.Epoch(), gate)
		} else {
			_, e = RunConditionalStopGate(c, r, r.Epoch(), gate)
		}
		if e == nil {
			t.Fatal("epoch motion accepted")
		}
	}
}

func TestMixtureStopControl(t *testing.T) {
	base, e := Base(Scenarios[7], 7, 0)
	if e != nil {
		t.Fatal(e)
	}
	for mode := 0; mode < 2; mode++ {
		a, e := RunFullBudgetStream(base, "design", 7, 0, 0, 2026111202, 2, mode, 0)
		if e != nil {
			t.Fatal(e)
		}
		b, e := RunMixtureStopStream(base, "design", 7, 0, 0, 2026111202, 2, mode, 0)
		if e != nil {
			t.Fatal(e)
		}
		a.FitNS, a.TreeNS, b.FitNS, b.TreeNS = 0, 0, 0, 0
		if !reflect.DeepEqual(a, b) {
			t.Fatal("control changed")
		}
	}
	c, e := RunMixtureStopStream(base, "design", 7, 0, 0, 2026111202, 2, 2, 0)
	if e != nil {
		t.Fatal(e)
	}
	for i, tick := range c.Ticks {
		if c.Views[i][2].Stop == "confidence" {
			p := tick.Predictions[2].P
			if p > .1 && p < .9 {
				t.Fatal("stop disagrees with actual emitted mixture")
			}
		}
	}
}

func TestMixtureStopArtifactReplay(t *testing.T) {
	path := os.Getenv("EVENTFRAME_MIXTURESTOP_ARTIFACT")
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
	decoder := json.NewDecoder(z)
	var h struct{ ExpectedRecords int }
	if e = decoder.Decode(&h); e != nil {
		t.Fatal(e)
	}
	n := 0
	e = RunMixtureStop(func(r BreadthRecord) error {
		var old BreadthRecord
		if e := decoder.Decode(&old); e != nil {
			return e
		}
		r.FitNS, r.TreeNS, old.FitNS, old.TreeNS = 0, 0, 0, 0
		if !reflect.DeepEqual(r, old) {
			t.Fatalf("replay%d", n)
		}
		n++
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	if n != h.ExpectedRecords {
		t.Fatal("count")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		t.Fatal("trailing")
	}
}
