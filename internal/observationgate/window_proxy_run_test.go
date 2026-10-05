package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"io"
	"math"
	"math/bits"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Copied isolated driver; only guide dispatch changes. Disabled replay must be exact.
func windowProxyRun(base *observation.Model, p innerArrivalResult, d []availableEvidenceFrame, reference []competitionFrame, enabled bool) (windowCouplingResult, error) {
	var result windowCouplingResult
	if base == nil || len(p.Frames) != 512 || len(d) != 512 || reference != nil && len(reference) != 512 {
		return result, fmt.Errorf("incomplete coupling inputs")
	}
	batches := make([][]int, 544)
	for i, f := range p.Frames {
		if f.Arrival < i || f.Arrival > i+31 {
			return result, fmt.Errorf("arrival")
		}
		if !f.Missing {
			batches[f.Arrival] = append(batches[f.Arrival], i)
		}
	}
	var state [2][3]bayes.ForecastMix
	m := windowGuideModels{Base: base}
	var local, pooled *observation.Model
	fitIndex, fitClock, splitClock := 0, -1, -1
	for clock, batch := range batches {
		if clock < 512 {
			f, x := p.Frames[clock], d[clock]
			if x.Origin != clock || x.RequestedValues != f.X&x.RequestedMask {
				return result, fmt.Errorf("mask binding")
			}
			if splitClock >= 0 {
				m.Long = local
			} else {
				m.Long = pooled
			}
			row := windowCouplingFrame{OldAvailable: m.LabelCount != nil && m.LabelSubset != nil, FitClock: fitClock}
			for arm := 0; arm < 2; arm++ {
				for layer := 0; layer < 3; layer++ {
					row.Weights[arm][layer] = state[arm][layer].Weights
				}
			}
			fixed, e := windowFixedAdvice(m, state[0], x.RequestedMask, x.RequestedValues, f.Predictions[2].Cost)
			if e != nil {
				return result, e
			}
			row.Arms[0] = fixed
			reader, e := windowSeededReader(f.X, x.MonitorMask, x.MonitorValues)
			if e != nil {
				return result, e
			}
			row.Arms[1], e = windowProxyPredict(m, state[1][0], state[1][1], state[1][2], reader, f.Seed, enabled)
			if e != nil {
				return result, e
			}
			row.Incremental = [2]int{bits.OnesCount16(x.RequestedMask &^ x.MonitorMask), reader.charged - bits.OnesCount16(x.MonitorMask)}
			for arm, a := range row.Arms {
				if a.Cost > 6 || a.Mask >= 512 || a.Values != f.X&a.Mask || row.Incremental[arm] != bits.OnesCount16(a.Mask&^x.MonitorMask) {
					return result, fmt.Errorf("cost or mask")
				}
			}
			if reference != nil {
				want := reference[clock].Views[0]
				if math.Abs(fixed.P-want.P) > 1e-12 || math.Abs(fixed.Outer[1]-want.Short) > 1e-12 {
					return result, fmt.Errorf("fixed law drift at %d", clock)
				}
				for i := 0; i < 4; i++ {
					if math.Abs(fixed.NewInner[i]-want.InnerAdvice[i]) > 1e-12 || math.Abs(fixed.Outer[i]-want.OuterAdvice[i]) > 1e-12 || math.Abs(state[0][1].Weights[i]-want.InnerWeights[i]) > 1e-12 || math.Abs(state[0][2].Weights[i]-want.OuterWeights[i]) > 1e-12 {
						return result, fmt.Errorf("fixed advice drift at %d", clock)
					}
				}
			}
			result.Frames = append(result.Frames, row)
		}
		for _, origin := range batch {
			if splitClock < 0 || origin > splitClock {
				issued := result.Frames[origin]
				y := p.Frames[origin].Y
				for arm, a := range issued.Arms {
					if issued.OldAvailable {
						state[arm][0] = state[arm][0].Observe(a.OldInner, y, 1)
					}
					state[arm][1] = state[arm][1].Observe(a.NewInner, y, 1)
					state[arm][2] = state[arm][2].Observe(a.Outer, y, 1)
				}
			}
		}
		if clock == p.Arms[2].SplitAt {
			splitClock = clock
			for arm := 0; arm < 2; arm++ {
				state[arm][2] = coupledArrivalRevoke(state[arm][2])
			}
		}
		if fitIndex < len(p.Fits) && p.Fits[fitIndex].Clock == clock {
			f := p.Fits[fitIndex]
			ids, recent, e := eventWindowIDs(p.Frames, f)
			if e != nil {
				return result, e
			}
			label, e := fitEventWindow(p.Frames, ids)
			if e != nil {
				return result, e
			}
			event, e := fitEventWindow(p.Frames, recent)
			if e != nil {
				return result, e
			}
			m.LabelCount, m.LabelSubset, m.EventCount, m.EventSubset = label.count, label.subset, event.count, event.subset
			ls, rs := make([]observation.Sample, len(f.Origins)), make([]observation.Sample, len(f.Origins))
			for j, i := range f.Origins {
				z := p.Frames[i]
				ls[j] = observation.Sample{Bits: z.X, Outcome: z.Y}
				rs[j] = observation.Sample{Bits: z.RX, Outcome: z.RY}
			}
			local, e = observation.Fit(ls)
			if e != nil {
				return result, e
			}
			n := len(ls)
			pool := append(append([]observation.Sample{}, ls[max(0, n-128):]...), rs[max(0, n-128):]...)
			pooled, e = observation.Fit(pool)
			if e != nil {
				return result, e
			}
			result.Fits = append(result.Fits, eventWindowFit{clock, ids, recent})
			fitClock = clock
			fitIndex++
		}
	}
	if fitIndex != len(p.Fits) {
		return result, fmt.Errorf("unprocessed fits")
	}
	return result, nil
}

func TestWindowProxyExperiment(t *testing.T) {
	paths := []string{os.Getenv("EVENTFRAME_PROXY_COMP"), os.Getenv("EVENTFRAME_PROXY_ADVICE"), os.Getenv("EVENTFRAME_PROXY_PARENT"), os.Getenv("EVENTFRAME_PROXY_ORIGINAL")}
	out := os.Getenv("EVENTFRAME_PROXY_OUTPUT")
	if out == "" || paths[0] == "" || paths[1] == "" || paths[2] == "" || paths[3] == "" {
		t.Skip("opt-in synthetic coupling experiment")
	}
	hash := func(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
	want := []string{"ecb3cf7f99d7f1460499c4d4ba9e6473530c47b8d824a7ee61ac386305ad640b", "3a5fb1fd119c7c14aaeb1864a95e3895a18781bc3e21e75ef5f7dcef485cb76e", "4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655", "ab3756fc535db2a1cc47d3c97629d5e2e291452284dd1238702e7cebf840dda5"}
	var ds []*json.Decoder
	for i, path := range paths {
		raw, e := os.ReadFile(path)
		if e != nil || hash(raw) != want[i] {
			t.Fatal("input", e)
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
		ds = append(ds, d)
	}
	o, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer o.Close()
	enc := json.NewEncoder(o)
	files, e := filepath.Glob("../../internal/*/*.go")
	if e != nil {
		t.Fatal(e)
	}
	files = append(files, "../../docs/experiments/mmm-window-proxy-v1-contract.md", "../../research/window-proxy-summary.mjs", "../../go.mod", "../../go.sum")
	sources, hashes := map[string]string{}, map[string]string{}
	for _, path := range files {
		raw, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		sources[path[6:]] = string(raw)
		hashes[path[6:]] = hash(raw)
	}
	if e = enc.Encode(map[string]any{"InputSHA256": want, "Sources": sources, "Hashes": hashes, "Consumed": true}); e != nil {
		t.Fatal(e)
	}
	n := 0
	for {
		var c struct {
			Phase, Case        string
			Index              int
			Immediate, Delayed []competitionFrame
		}
		if e = ds[0].Decode(&c); e == io.EOF {
			break
		} else if e != nil {
			t.Fatal(e)
		}
		var d availableEvidenceRecord
		var p creditLearningRecord
		if e = ds[1].Decode(&d); e != nil {
			t.Fatal(e)
		}
		if e = ds[2].Decode(&p); e != nil {
			t.Fatal(e)
		}
		if c.Phase != p.Phase || c.Case != p.Case || c.Index != p.Index || d.Phase != p.Phase || d.Case != p.Case || d.Index != p.Index {
			t.Fatal("key")
		}
		var previous struct {
			Phase, Case        string
			Index              int
			Immediate, Delayed windowCouplingResult
		}
		if e = ds[3].Decode(&previous); e != nil {
			t.Fatal(e)
		}
		if previous.Phase != p.Phase || previous.Case != p.Case || previous.Index != p.Index {
			t.Fatal("previous key")
		}
		phase := 0
		if p.Phase == "cohort2" {
			phase = 1
		}
		scenario := -1
		for i, name := range arrivalSwitchNames {
			if name == p.Case {
				scenario = i
			}
		}
		if scenario < 0 {
			t.Fatal("scenario")
		}
		base, m, e := arrivalSwitchSetup(phase, scenario, p.Index)
		if e != nil || m != p.Masks {
			t.Fatal("base setup", e)
		}
		row := struct {
			Phase, Case        string
			Index              int
			Immediate, Delayed windowCouplingResult
		}{Phase: p.Phase, Case: p.Case, Index: p.Index}
		disabled, e := windowProxyRun(base, p.Immediate, d.Immediate, c.Immediate, false)
		if e != nil || !reflect.DeepEqual(disabled, previous.Immediate) {
			t.Fatal("disabled immediate drift", e)
		}
		disabled, e = windowProxyRun(base, p.Delayed, d.Delayed, c.Delayed, false)
		if e != nil || !reflect.DeepEqual(disabled, previous.Delayed) {
			t.Fatal("disabled delayed drift", e)
		}
		row.Immediate, e = windowProxyRun(base, p.Immediate, d.Immediate, c.Immediate, true)
		if e != nil {
			t.Fatal(p.Phase, p.Case, p.Index, "immediate", e)
		}
		row.Delayed, e = windowProxyRun(base, p.Delayed, d.Delayed, c.Delayed, true)
		if e != nil {
			t.Fatal(p.Phase, p.Case, p.Index, "delayed", e)
		}
		for j, pair := range [][2]windowCouplingResult{{row.Immediate, previous.Immediate}, {row.Delayed, previous.Delayed}} {
			if !reflect.DeepEqual(pair[0].Fits, pair[1].Fits) {
				t.Fatal("fit drift", j)
			}
			for i, f := range pair[0].Frames {
				old := pair[1].Frames[i]
				if f.Arms[0] != old.Arms[0] || f.Weights[0] != old.Weights[0] || f.Incremental[0] != old.Incremental[0] || f.OldAvailable != old.OldAvailable || f.FitClock != old.FitClock {
					t.Fatal("fixed drift", j, i)
				}
			}
		}
		if e = enc.Encode(row); e != nil {
			t.Fatal(e)
		}
		n++
		if p.Index == 15 {
			t.Log(p.Phase, p.Case, "complete")
		}
	}
	if n != 128 {
		t.Fatal("incomplete", n)
	}
	for _, d := range ds[1:] {
		var h json.RawMessage
		if e = d.Decode(&h); e != io.EOF {
			t.Fatal("trailing")
		}
	}
	if e = o.Sync(); e != nil {
		t.Fatal(e)
	}
}
