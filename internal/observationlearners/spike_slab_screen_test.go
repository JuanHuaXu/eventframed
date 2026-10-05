package observationlearners

import (
	"encoding/json"
	"io"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type spikeScreenRecord struct {
	Clock int
	spikePilotRecord
}

func TestSpikeScreenAsOf(t *testing.T) {
	r := softV120Record{Steps: make([]softV120Step, 256), Fits: make([]softV120Fit, 8)}
	for i := range r.Initial {
		r.Initial[i] = observation.Sample{Bits: uint16(i * 29), Outcome: i%3 == 0}
	}
	for i := range r.Steps {
		r.Steps[i] = softV120Step{X: uint16(i * 37 % 512), Y: i%3 == 0, Delay: i % 32, Missing: i%5 == 0}
	}
	for c := 0; c <= 224; c += 32 {
		var origins []int
		for i := -16; i < c; i++ {
			if i < 0 || (!r.Steps[i].Missing && i+r.Steps[i].Delay <= c) {
				origins = append(origins, i)
			}
		}
		if len(origins) > 64 {
			origins = origins[len(origins)-64:]
		}
		r.Fits[c/32].Origins[0] = origins
	}
	for _, reverse := range []bool{false, true} {
		for _, c := range []int{0, 128, 224} {
			base, err := runSpikePilotOrdered(r, c, reverse)
			if err != nil {
				t.Fatal(err)
			}
			poison := r
			poison.Steps = append([]softV120Step(nil), r.Steps...)
			for i := range poison.Steps {
				p := &poison.Steps[i]
				p.Q = math.NaN()
				p.P = [15]float64{math.NaN()}
				if i >= c || p.Missing || i+p.Delay > c {
					p.Y = !p.Y
				}
			}
			got, err := runSpikePilotOrdered(poison, c, reverse)
			if err != nil {
				t.Fatal(err)
			}
			if got.P != base.P || !reflect.DeepEqual(got.Trace, base.Trace) {
				t.Fatal("as-of leak", c)
			}
			if c == 128 && !reverse {
				old, err := runSpikePilot(r)
				if err != nil || !reflect.DeepEqual(old, base) {
					t.Fatal("pilot wrapper changed", err)
				}
			}
			for _, i := range base.Origins {
				if i < 0 {
					poison.Initial[i+16].Outcome = !r.Initial[i+16].Outcome
				} else {
					poison.Steps[i].Y = !r.Steps[i].Y
				}
			}
			changed, err := runSpikePilotOrdered(poison, c, reverse)
			if err != nil {
				t.Fatal(err)
			}
			if changed.P == base.P {
				t.Fatal("admitted labels ignored", c)
			}
			for i := range base.P {
				if math.Abs(base.P[i][0]+changed.P[i][0]-1) > 1e-8 {
					t.Fatal("complement", c)
				}
			}
		}
	}
	for _, c := range []int{-1, 1, 225, 256} {
		if _, err := runSpikePilotAt(r, c); err == nil {
			t.Fatal("invalid clock")
		}
	}
}

func TestSpikeScreenCollect(t *testing.T) {
	src, dst := os.Getenv("EVENTFRAME_SPIKE_SCREEN_SOURCE"), os.Getenv("EVENTFRAME_SPIKE_SCREEN_OUTPUT")
	if src == "" {
		t.Skip("explicit research source")
	}
	if dst == "" {
		t.Fatal("output required")
	}
	in, err := os.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	dec, enc := json.NewDecoder(in), json.NewEncoder(out)
	var h softV120Artifact
	if err = dec.Decode(&h); err != nil || h.Version != "soft-learners-v120" {
		t.Fatal("header", err)
	}
	count := 0
	for {
		var r softV120Record
		err = dec.Decode(&r)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if r.Index < 0 || r.Index >= 8 {
			continue
		}
		for _, c := range []int{0, 128, 224} {
			v, err := runSpikePilotAt(r, c)
			if err != nil {
				t.Fatalf("fit%d phase%d case%d index%d schedule%d clock%d: %v", count, r.Phase, r.Case, r.Index, r.Schedule, c, err)
			}
			if err = enc.Encode(spikeScreenRecord{c, v}); err != nil {
				t.Fatal(err)
			}
			count++
		}
	}
	if count != 2016 {
		t.Fatal("count", count)
	}
	if err = out.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Logf("fits=%d forecasts=%d", count, count*32)
}
