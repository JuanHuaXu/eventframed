package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/bits"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
)

type windowCouplingFrame struct {
	Arms         [2]windowGuideAdvice
	Weights      [2][3][4]float64
	Incremental  [2]int
	OldAvailable bool
	FitClock     int
}
type windowCouplingResult struct {
	Frames []windowCouplingFrame
	Fits   []eventWindowFit
}

func windowSeededReader(x, mask, values uint16) (*monitorCreditReader, error) {
	if x >= 512 || mask >= 512 || values&^mask != 0 || values != x&mask {
		return nil, fmt.Errorf("invalid paid monitor seed")
	}
	r := newMonitorCreditReader(observationexperiment.Frames(x, "window-private"))
	// Seed only independently captured, already-paid monitoring evidence. This
	// does not change Reader responses or authorize additional consumed bits.
	r.mask, r.values, r.charged = mask, values, bits.OnesCount16(mask)
	return r, nil
}

func windowFixedAdvice(m windowGuideModels, weights [3]bayes.ForecastMix, mask, values uint16, cost int) (windowGuideAdvice, error) {
	a := windowGuideAdvice{Guide: "fixed", Mask: mask, Values: values, Cost: cost, Outer: [4]float64{.5, .5, .5, .5}, OldInner: [4]float64{.5, .5, .5, .5}}
	for i, model := range []*observation.Model{m.Base, m.LabelCount, m.Long} {
		if model != nil {
			p, e := model.ForecastObserved(mask, values)
			if e != nil {
				return a, e
			}
			a.Outer[i] = p
		}
	}
	a.OldInner[0] = a.Outer[1]
	old := a.Outer[1]
	if m.LabelCount != nil && m.LabelSubset != nil {
		p, e := m.LabelSubset.Forecast(mask, values)
		if e != nil {
			return a, e
		}
		a.OldInner[1], a.OldInner[2], a.OldInner[3] = p, p, p
		old = weights[0].Forecast(a.OldInner)
	}
	a.NewInner = [4]float64{old, .5, .5, .5}
	if m.EventCount != nil {
		p, e := m.EventCount.ForecastObserved(mask, values)
		if e != nil {
			return a, e
		}
		a.NewInner[1] = p
	}
	if m.EventSubset != nil {
		p, e := m.EventSubset.Forecast(mask, values)
		if e != nil {
			return a, e
		}
		a.NewInner[2], a.NewInner[3] = p, p
	}
	a.Outer[1] = weights[1].Forecast(a.NewInner)
	a.P = weights[2].Forecast(a.Outer)
	return a, nil
}

func windowCouplingRun(base *observation.Model, p innerArrivalResult, d []availableEvidenceFrame, reference []competitionFrame) (windowCouplingResult, error) {
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
			row.Arms[1], e = windowGuidePredict(m, state[1][0], state[1][1], state[1][2], reader, f.Seed)
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

func TestWindowCouplingContracts(t *testing.T) {
	a, e := windowSeededReader(7, 1, 1)
	if e != nil {
		t.Fatal(e)
	}
	b, _ := windowSeededReader(7, 1, 1)
	mask, v, e := a.Read(observation.View{Scope: 0, Depth: 2})
	if e != nil || mask != 7 || v != 7 || a.charged != 3 || b.mask != 1 || b.charged != 1 {
		t.Fatal("private cache", e)
	}
	if _, e := windowSeededReader(0, 1, 1); e == nil {
		t.Fatal("bad seed")
	}
	base, e := observation.Fit([]observation.Sample{{Bits: 0, Outcome: false}, {Bits: 1, Outcome: true}})
	if e != nil {
		t.Fatal(e)
	}
	p := innerArrivalResult{}
	p.Arms[2].SplitAt = -1
	d := make([]availableEvidenceFrame, 512)
	for i := 0; i < 512; i++ {
		x := uint16(i % 512)
		p.Frames = append(p.Frames, innerArrivalFrame{X: x, RX: x, Y: i%2 != 0, RY: i%2 != 0, Arrival: i, Audit: i < 32})
		p.Frames[i].Predictions[2].Cost = 6
		d[i] = availableEvidenceFrame{Origin: i, RequestedMask: 63, RequestedValues: x & 63}
	}
	ids := make([]int, 32)
	for i := range ids {
		ids[i] = i
	}
	p.Fits = []forestDelayFit{{Clock: 31, Origins: ids}}
	p.Frames[42].Missing = true
	first, e := windowCouplingRun(base, p, d, nil)
	if e != nil {
		t.Fatal(e)
	}
	p.Frames[42].Y = !p.Frames[42].Y
	second, e := windowCouplingRun(base, p, d, nil)
	if e != nil || !reflect.DeepEqual(first, second) {
		t.Fatal("missing leakage", e)
	}
	p.Frames[100].Y = !p.Frames[100].Y
	third, e := windowCouplingRun(base, p, d, nil)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(second.Frames[:101], third.Frames[:101]) {
		t.Fatal("future feedback leakage")
	}
	if first.Frames[31].OldAvailable || !first.Frames[32].OldAvailable {
		t.Fatal("same-clock fit leakage")
	}
}

func TestWindowCouplingExperiment(t *testing.T) {
	paths := []string{os.Getenv("EVENTFRAME_COUPLE_COMP"), os.Getenv("EVENTFRAME_COUPLE_ADVICE"), os.Getenv("EVENTFRAME_COUPLE_PARENT")}
	out := os.Getenv("EVENTFRAME_COUPLE_OUTPUT")
	if out == "" || paths[0] == "" || paths[1] == "" || paths[2] == "" {
		t.Skip("opt-in synthetic coupling experiment")
	}
	hash := func(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
	want := []string{"ecb3cf7f99d7f1460499c4d4ba9e6473530c47b8d824a7ee61ac386305ad640b", "3a5fb1fd119c7c14aaeb1864a95e3895a18781bc3e21e75ef5f7dcef485cb76e", "4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655"}
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
	files = append(files, "../../docs/experiments/mmm-window-coupling-v1-contract.md", "../../research/window-coupling-summary.mjs", "../../go.mod", "../../go.sum")
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
		row.Immediate, e = windowCouplingRun(base, p.Immediate, d.Immediate, c.Immediate)
		if e != nil {
			t.Fatal(p.Phase, p.Case, p.Index, "immediate", e)
		}
		row.Delayed, e = windowCouplingRun(base, p.Delayed, d.Delayed, c.Delayed)
		if e != nil {
			t.Fatal(p.Phase, p.Case, p.Index, "delayed", e)
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
