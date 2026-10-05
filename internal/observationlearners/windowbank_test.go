package observationlearners

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestWindowBankControl(t *testing.T) {
	for _, c := range []struct {
		w          [4]float64
		mode, want int
	}{
		{[4]float64{.4, .2, .2, .2}, 0, 1}, {[4]float64{.4, .2, .2, .2}, 1, 0},
		{[4]float64{.1, .2, .35, .35}, 1, 1}, {[4]float64{.2, .5, .2, .1}, 1, 2},
		{[4]float64{.1, .6, .2, .1}, 2, 3},
	} {
		if windowGuide(c.w, c.mode) != c.want {
			t.Fatal("duplicate/tie guide")
		}
	}
	base, e := Base(Scenarios[7], 7, 0)
	if e != nil {
		t.Fatal(e)
	}
	a, e := RunBreadthStream(base, "design", 7, 0, 0, 2026108202, 2, 1, 0)
	if e != nil {
		t.Fatal(e)
	}
	b, e := RunWindowBankStream(base, "design", 7, 0, 0, 2026108202, 2, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	a.FitNS, a.TreeNS, b.FitNS, b.TreeNS = 0, 0, 0, 0
	a.Mode = b.Mode
	if !reflect.DeepEqual(a, b) {
		t.Fatal("subset64 control")
	}
	for mode := 1; mode < 3; mode++ {
		c, e := RunWindowBankStream(base, "design", 7, 0, 0, 2026108202, 2, mode, 0)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(a.Inputs, c.Inputs) {
			t.Fatal("input pairing")
		}
		for i, tick := range c.Ticks {
			if tick.Outcome != a.Ticks[i].Outcome || !reflect.DeepEqual(tick.Delivered, a.Ticks[i].Delivered) {
				t.Fatal("feedback pairing")
			}
			for _, arm := range []int{0, 1, 3} {
				if tick.Predictions[arm] != a.Ticks[i].Predictions[arm] {
					t.Fatal("unchanged arm")
				}
			}
		}
	}
}

func TestWindowBankArtifactReplay(t *testing.T) {
	path := os.Getenv("EVENTFRAME_WINDOWBANK_ARTIFACT")
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
	var header struct{ ExpectedRecords int }
	if e = decoder.Decode(&header); e != nil {
		t.Fatal(e)
	}
	n := 0
	e = RunWindowBank(func(r BreadthRecord) error {
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
	if n != header.ExpectedRecords {
		t.Fatal("count")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		t.Fatal("trailing records")
	}
}
