package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
)

type competitionView struct {
	P, Short                                             float64
	InnerAdvice, OuterAdvice, InnerWeights, OuterWeights [4]float64
}
type competitionFrame struct{ Views [2]competitionView }

func windowCompetitionReplay(p innerArrivalResult, d []availableEvidenceFrame, w eventWindowResult) ([]competitionFrame, error) {
	if len(p.Frames) != 512 || len(d) != 512 || len(w.Frames) != 512 {
		return nil, fmt.Errorf("incomplete tape")
	}
	batches := make([][]int, 544)
	for i, f := range p.Frames {
		if f.Arrival < i || f.Arrival > i+31 {
			return nil, fmt.Errorf("invalid arrival")
		}
		if !f.Missing {
			batches[f.Arrival] = append(batches[f.Arrival], i)
		}
	}
	var control bayes.ForecastMix
	var inner, outer [2]bayes.ForecastMix
	var issued []competitionFrame
	splitClock := -1
	for clock, batch := range batches {
		if clock < 512 {
			x := d[clock]
			if x.Origin != clock {
				return nil, fmt.Errorf("origin mismatch")
			}
			cw := control.Weights
			if cw == [4]float64{} {
				cw = [4]float64{.7, .1, .1, .1}
			}
			for i, v := range cw {
				if math.Abs(v-x.OuterWeights[i]) > 1e-12 {
					return nil, fmt.Errorf("control weights drift at %d", clock)
				}
			}
			var frame competitionFrame
			for view := 0; view < 2; view++ {
				want := x.Original
				if view == 1 {
					want = x.Available
				}
				if math.Abs(control.Forecast(x.Experts[view])-want) > 1e-12 {
					return nil, fmt.Errorf("control law drift")
				}
				raw := w.Frames[clock].Raw[1][view]
				a := [4]float64{x.Experts[view][1], raw[0], raw[1], raw[1]}
				if !bayes.ValidForecastExperts(a) {
					return nil, fmt.Errorf("invalid inner advice")
				}
				short := inner[view].Forecast(a)
				e := x.Experts[view]
				e[1] = short
				frame.Views[view] = competitionView{outer[view].Forecast(e), short, a, e, inner[view].Weights, outer[view].Weights}
			}
			issued = append(issued, frame)
		}
		for _, origin := range batch {
			if splitClock >= 0 && origin <= splitClock {
				continue
			}
			f := p.Frames[origin]
			control = control.Observe(d[origin].Experts[0], f.Y, 1)
			for view := 0; view < 2; view++ {
				advice := issued[origin].Views[view]
				inner[view] = inner[view].Observe(advice.InnerAdvice, f.Y, 1)
				outer[view] = outer[view].Observe(advice.OuterAdvice, f.Y, 1)
			}
		}
		if clock == p.Arms[2].SplitAt {
			splitClock = clock
			control = coupledArrivalRevoke(control)
			for view := 0; view < 2; view++ {
				outer[view] = coupledArrivalRevoke(outer[view])
			}
		}
	}
	return issued, nil
}

func TestWindowCompetitionJournal(t *testing.T) {
	// Synthetic control advice remains neutral, enabling future/missing-label
	// checks without granting the candidate access to a future control trajectory.
	p := innerArrivalResult{}
	p.Arms[2].SplitAt = -1
	d := make([]availableEvidenceFrame, 512)
	w := eventWindowResult{Frames: make([]eventWindowFrame, 512)}
	var c bayes.ForecastMix
	for i := 0; i < 512; i++ {
		p.Frames = append(p.Frames, innerArrivalFrame{Arrival: i, Missing: true})
		e := [4]float64{.5, .5, .5, .5}
		d[i] = availableEvidenceFrame{Origin: i, Original: c.Forecast(e), Available: c.Forecast(e), Experts: [2][4]float64{e, e}, OuterWeights: [4]float64{.7, .1, .1, .1}}
		w.Frames[i].Raw[1] = [2][2]float64{{.9, .1}, {.8, .2}}
	}
	// One delivered label at clock2; update the neutral control weights in tape.
	p.Frames[0].Missing = false
	p.Frames[0].Arrival = 2
	p.Frames[0].Y = true
	c = c.Observe(d[0].Experts[0], true, 1)
	for i := 3; i < 512; i++ {
		d[i].OuterWeights = c.Weights
	}
	a, e := windowCompetitionReplay(p, d, w)
	if e != nil {
		t.Fatal(e)
	}
	p.Frames[0].Y = false
	b, e := windowCompetitionReplay(p, d, w)
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i <= 2; i++ {
		if a[i] != b[i] {
			t.Fatal("future label leakage")
		}
	}
	if a[3] == b[3] {
		t.Fatal("feedback ignored")
	}
	p.Frames[0].Y = true
	p.Frames[1].Y = true
	b, e = windowCompetitionReplay(p, d, w)
	if e != nil {
		t.Fatal(e)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("missing label leakage")
		}
	}
	p.Frames[0].Arrival = -1
	if _, e := windowCompetitionReplay(p, d, w); e == nil {
		t.Fatal("invalid arrival accepted")
	}
}

func TestWindowCompetitionExperiment(t *testing.T) {
	paths := []string{os.Getenv("EVENTFRAME_COMP_WINDOW"), os.Getenv("EVENTFRAME_COMP_ADVICE"), os.Getenv("EVENTFRAME_COMP_PARENT")}
	out := os.Getenv("EVENTFRAME_COMP_OUTPUT")
	if paths[0] == "" || paths[1] == "" || paths[2] == "" || out == "" {
		t.Skip("opt-in consumed competition")
	}
	hash := func(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
	want := []string{"5d5c0d8b4fc0447f738311019f4475387cf51f81969733578942335ab189b332", "3a5fb1fd119c7c14aaeb1864a95e3895a18781bc3e21e75ef5f7dcef485cb76e", "4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655"}
	var decoders []*json.Decoder
	for i, path := range paths {
		raw, e := os.ReadFile(path)
		if e != nil || hash(raw) != want[i] {
			t.Fatal("input hash", e)
		}
		f, e := os.Open(path)
		if e != nil {
			t.Fatal(e)
		}
		defer f.Close()
		d := json.NewDecoder(f)
		var h json.RawMessage
		if e = d.Decode(&h); e != nil {
			t.Fatal(e)
		}
		decoders = append(decoders, d)
	}
	o, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer o.Close()
	enc := json.NewEncoder(o)
	sources, hashes := map[string]string{}, map[string]string{}
	for _, path := range []string{"internal/observationgate/window_competition_test.go", "internal/observationgate/coupled_arrival_run_test.go", "internal/bayes/forecast_mix.go", "docs/experiments/mmm-window-competition-v1-contract.md", "research/window-competition-summary.mjs", "go.mod", "go.sum"} {
		raw, e := os.ReadFile("../../" + path)
		if e != nil {
			t.Fatal(e)
		}
		sources[path] = string(raw)
		hashes[path] = hash(raw)
	}
	if e = enc.Encode(map[string]any{"InputSHA256": want, "Sources": sources, "Hashes": hashes, "Consumed": true}); e != nil {
		t.Fatal(e)
	}
	n := 0
	for {
		var w struct {
			Phase, Case        string
			Index              int
			Immediate, Delayed eventWindowResult
		}
		if e = decoders[0].Decode(&w); e == io.EOF {
			break
		} else if e != nil {
			t.Fatal(e)
		}
		var d availableEvidenceRecord
		var p creditLearningRecord
		if e = decoders[1].Decode(&d); e != nil {
			t.Fatal(e)
		}
		if e = decoders[2].Decode(&p); e != nil {
			t.Fatal(e)
		}
		if w.Phase != d.Phase || w.Case != d.Case || w.Index != d.Index || w.Phase != p.Phase || w.Case != p.Case || w.Index != p.Index {
			t.Fatal("key mismatch")
		}
		row := struct {
			Phase, Case        string
			Index              int
			Immediate, Delayed []competitionFrame
		}{Phase: w.Phase, Case: w.Case, Index: w.Index}
		row.Immediate, e = windowCompetitionReplay(p.Immediate, d.Immediate, w.Immediate)
		if e != nil {
			t.Fatal(w.Phase, w.Case, w.Index, e)
		}
		row.Delayed, e = windowCompetitionReplay(p.Delayed, d.Delayed, w.Delayed)
		if e != nil {
			t.Fatal(w.Phase, w.Case, w.Index, e)
		}
		if e = enc.Encode(row); e != nil {
			t.Fatal(e)
		}
		n++
	}
	if n != 128 {
		t.Fatal("incomplete rows", n)
	}
	for _, d := range decoders[1:] {
		var x json.RawMessage
		if e = d.Decode(&x); e != io.EOF {
			t.Fatal("trailing rows")
		}
	}
	if e = o.Sync(); e != nil {
		t.Fatal(e)
	}
}
