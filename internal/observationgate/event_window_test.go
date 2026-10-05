package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

type eventWindowModels struct {
	count  *observation.Model
	subset *observationlearners.ConditionalForest
}
type eventWindowFit struct {
	Clock              int
	LabelIDs, EventIDs []int
}
type eventWindowFrame struct {
	Raw        [2][2][2]float64 // window(label,event), view(requested,available), model(count,subset)
	Short, Law [2][2]float64
	FitClock   int
}
type eventWindowResult struct {
	Frames []eventWindowFrame
	Fits   []eventWindowFit
}

// Select by evidence availability first, then by origin age. Never accept a
// future/missing label just because its origin falls inside the time window.
func eventWindowIDs(frames []innerArrivalFrame, fit forestDelayFit) ([]int, []int, error) {
	if fit.Clock < 0 {
		return nil, nil, fmt.Errorf("negative fit clock")
	}
	last := -1
	var recent []int
	for _, i := range fit.Origins {
		if i <= last || i < 0 || i >= len(frames) || i > fit.Clock {
			return nil, nil, fmt.Errorf("invalid origin order")
		}
		f := frames[i]
		if !f.Audit || f.Missing || f.Arrival < i || f.Arrival > fit.Clock {
			return nil, nil, fmt.Errorf("unavailable training label")
		}
		if i > fit.Clock-64 {
			recent = append(recent, i)
		}
		last = i
	}
	ids := append([]int(nil), fit.Origins[max(0, len(fit.Origins)-64):]...)
	return ids, recent, nil
}
func fitEventWindow(frames []innerArrivalFrame, ids []int) (eventWindowModels, error) {
	var m eventWindowModels
	if len(ids) == 0 {
		return m, nil
	}
	samples := make([]observation.Sample, len(ids))
	for j, i := range ids {
		samples[j] = observation.Sample{Bits: frames[i].X, Outcome: frames[i].Y}
	}
	var err error
	m.count, err = observation.Fit(samples)
	if err != nil {
		return m, err
	}
	var weights [512]float64
	for i := range weights {
		weights[i] = 1
	}
	m.subset, err = observationlearners.NewSubsetConditional(samples, weights)
	return m, err
}
func (m eventWindowModels) forecast(mask, values uint16) ([2]float64, error) {
	p := [2]float64{.5, .5}
	if mask >= 512 || values&^mask != 0 {
		return p, fmt.Errorf("invalid observed view")
	}
	if m.count == nil && m.subset == nil {
		return p, nil
	}
	if m.count == nil || m.subset == nil {
		return p, fmt.Errorf("incomplete fit")
	}
	var err error
	p[0], err = m.count.ForecastObserved(mask, values)
	if err != nil {
		return p, err
	}
	p[1], err = m.subset.Forecast(mask, values)
	return p, err
}
func eventWindowReplay(p innerArrivalResult, advice []availableEvidenceFrame) (eventWindowResult, error) {
	var result eventWindowResult
	if len(p.Frames) != 512 || len(advice) != 512 {
		return result, fmt.Errorf("incomplete tape")
	}
	var models [2]eventWindowModels
	fitIndex, fitClock := 0, -1
	for clock, d := range advice {
		if d.Origin != clock {
			return result, fmt.Errorf("advice origin")
		}
		frame := eventWindowFrame{FitClock: fitClock}
		for w, m := range models {
			for v, view := range [][2]uint16{{d.RequestedMask, d.RequestedValues}, {d.ConsumedMask, d.ConsumedValues}} {
				raw, err := m.forecast(view[0], view[1])
				if err != nil {
					return result, err
				}
				frame.Raw[w][v] = raw
				q := .5
				if fitClock >= 0 {
					q = (bayes.ForecastMix{Weights: d.InnerWeights}).Forecast([4]float64{raw[0], raw[1], raw[1], raw[1]})
				}
				frame.Short[w][v] = q
				experts := d.Experts[v]
				experts[1] = q
				frame.Law[w][v] = (bayes.ForecastMix{Weights: d.OuterWeights}).Forecast(experts)
				if w == 0 && (math.Abs(q-d.Experts[v][1]) > 1e-12 || math.Abs(frame.Law[w][v]-[]float64{d.Original, d.Available}[v]) > 1e-12) {
					return result, fmt.Errorf("reference reconstruction drift clock=%d view=%d short=%g wanted=%g", clock, v, q, d.Experts[v][1])
				}
			}
		}
		result.Frames = append(result.Frames, frame)
		// Fitting occurs after the clock's issued forecast, exactly as in the parent.
		if fitIndex < len(p.Fits) && p.Fits[fitIndex].Clock == clock {
			f := p.Fits[fitIndex]
			ids, recent, err := eventWindowIDs(p.Frames, f)
			if err != nil {
				return result, err
			}
			models[0], err = fitEventWindow(p.Frames, ids)
			if err != nil {
				return result, err
			}
			models[1], err = fitEventWindow(p.Frames, recent)
			if err != nil {
				return result, err
			}
			result.Fits = append(result.Fits, eventWindowFit{clock, ids, recent})
			fitClock = clock
			fitIndex++
		}
	}
	if fitIndex < len(p.Fits) && p.Fits[fitIndex].Clock < 512 {
		return result, fmt.Errorf("missed publication")
	}
	return result, nil
}

func TestEventWindowContracts(t *testing.T) {
	f := make([]innerArrivalFrame, 65)
	ids := make([]int, 65)
	for i := range f {
		f[i] = innerArrivalFrame{X: uint16(i), Y: i%2 == 0, Arrival: i, Audit: true}
		ids[i] = i
	}
	a, b, e := eventWindowIDs(f, forestDelayFit{Clock: 64, Origins: ids})
	if e != nil || !reflect.DeepEqual(a, ids[1:]) || !reflect.DeepEqual(b, ids[1:]) {
		t.Fatal("window endpoints", e)
	}
	_, b, e = eventWindowIDs(f, forestDelayFit{Clock: 100, Origins: []int{0}})
	if e != nil || len(b) != 0 {
		t.Fatal("empty window", e)
	}
	m, e := fitEventWindow(f, nil)
	if e != nil {
		t.Fatal(e)
	}
	p, e := m.forecast(511, 64)
	if e != nil || p != [2]float64{.5, .5} {
		t.Fatal("empty forecast")
	}
	for _, bad := range []forestDelayFit{{Clock: 64, Origins: []int{2, 1}}, {Clock: 0, Origins: []int{1}}, {Clock: 64, Origins: []int{65}}} {
		if _, _, e := eventWindowIDs(f, bad); e == nil {
			t.Fatal("invalid support")
		}
	}
	for k := 0; k < 3; k++ {
		g := append([]innerArrivalFrame(nil), f...)
		if k == 0 {
			g[64].Missing = true
		}
		if k == 1 {
			g[64].Audit = false
		}
		if k == 2 {
			g[64].Arrival = 65
		}
		if _, _, e := eventWindowIDs(g, forestDelayFit{Clock: 64, Origins: ids}); e == nil {
			t.Fatal("unavailable evidence")
		}
	}
	m, e = fitEventWindow(f, ids[1:])
	if e != nil {
		t.Fatal(e)
	}
	for mask := uint16(0); mask < 512; mask++ {
		p, e := m.forecast(mask, 64&mask)
		if e != nil || p[0] < 0 || p[0] > 1 || p[1] < 0 || p[1] > 1 {
			t.Fatal("law", e)
		}
	}
}

func TestEventWindowExperiment(t *testing.T) {
	in, parent, out := os.Getenv("EVENTFRAME_WINDOW_INPUT"), os.Getenv("EVENTFRAME_WINDOW_PARENT"), os.Getenv("EVENTFRAME_WINDOW_OUTPUT")
	if in == "" || parent == "" || out == "" {
		t.Skip("opt-in consumed event-window diagnostic")
	}
	read := func(p string) []byte {
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	hash := func(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
	ir, pr := read(in), read(parent)
	if hash(ir) != "3a5fb1fd119c7c14aaeb1864a95e3895a18781bc3e21e75ef5f7dcef485cb76e" || hash(pr) != "4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655" {
		t.Fatal("wrong input")
	}
	f, e := os.Open(in)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	g, e := os.Open(parent)
	if e != nil {
		t.Fatal(e)
	}
	defer g.Close()
	d, pd := json.NewDecoder(f), json.NewDecoder(g)
	var header json.RawMessage
	if e = d.Decode(&header); e != nil {
		t.Fatal(e)
	}
	if e = pd.Decode(&header); e != nil {
		t.Fatal(e)
	}
	o, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer o.Close()
	enc := json.NewEncoder(o)
	paths := []string{"../../internal/observationgate/event_window_test.go", "../../docs/experiments/mmm-event-window-v1-contract.md", "../../research/event-window-summary.mjs", "../../go.mod", "../../go.sum"}
	for _, dir := range []string{"observation", "observationlearners", "bayes"} {
		more, e := filepath.Glob("../../internal/" + dir + "/*.go")
		if e != nil {
			t.Fatal(e)
		}
		paths = append(paths, more...)
	}
	sources, hashes := map[string]string{}, map[string]string{}
	for _, p := range paths {
		b := read(p)
		sources[p[6:]] = string(b)
		hashes[p[6:]] = hash(b)
	}
	if e = enc.Encode(map[string]any{"InputSHA256": hash(ir), "ParentSHA256": hash(pr), "Sources": sources, "Hashes": hashes, "Consumed": true}); e != nil {
		t.Fatal(e)
	}
	n := 0
	for {
		var r availableEvidenceRecord
		if e = d.Decode(&r); e == io.EOF {
			break
		} else if e != nil {
			t.Fatal(e)
		}
		var p creditLearningRecord
		if e = pd.Decode(&p); e != nil {
			t.Fatal(e)
		}
		if r.Phase != p.Phase || r.Case != p.Case || r.Index != p.Index {
			t.Fatal("key mismatch")
		}
		row := struct {
			Phase, Case        string
			Index              int
			Immediate, Delayed eventWindowResult
		}{Phase: r.Phase, Case: r.Case, Index: r.Index}
		row.Immediate, e = eventWindowReplay(p.Immediate, r.Immediate)
		if e != nil {
			t.Fatal(r.Phase, r.Case, r.Index, "Immediate", e)
		}
		row.Delayed, e = eventWindowReplay(p.Delayed, r.Delayed)
		if e != nil {
			t.Fatal(r.Phase, r.Case, r.Index, "Delayed", e)
		}
		if e = enc.Encode(row); e != nil {
			t.Fatal(e)
		}
		n++
		if r.Index == 15 {
			t.Log(r.Phase, r.Case, "complete")
		}
	}
	if n != 128 {
		t.Fatal("incomplete", n)
	}
	if e = pd.Decode(&header); e != io.EOF {
		t.Fatal("trailing parent")
	}
	if e = o.Sync(); e != nil {
		t.Fatal(e)
	}
}
