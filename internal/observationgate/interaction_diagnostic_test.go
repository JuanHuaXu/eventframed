package observationgate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

type interactionWindow struct {
	N, RequiredFrames, Covered, SubsetGuide, AgeGuide int
	Guide                                             [3]int
	Available                                         [5]int
	Brier                                             [5][2]float64
	Weights                                           [4]float64
	InnerWeights                                      [3]float64
	InnerAvailable                                    int
	MixtureBrier, FullMixtureBrier                    float64
}
type interactionPending struct {
	p             agePending
	required      uint16
	available     [5]bool
	probabilities [5][2]float64
	fullMixture   float64
	ageGuide      bool
}
type interactionDiagnostic struct {
	Windows [8]interactionWindow
	next    int
	pending *interactionPending
}

func interactionMask(step int, cfg subsetBreadthCase) (uint16, error) {
	if cfg.Mode == "null" {
		return 0, nil
	}
	if cfg.Mode == "stable" || step < 256 {
		return 448, nil
	}
	switch cfg.Target {
	case "bit":
		return 4, nil
	case "xor2":
		return 6, nil
	case "majority3", "mux":
		return 7, nil
	case "parity4":
		return 30, nil
	}
	return 0, fmt.Errorf("unknown interaction mask")
}
func (d *interactionDiagnostic) prepare(step int, x uint16, cfg subsetBreadthCase, p agePending, models observationpreserved.Models, base *observation.Model, retained, age *observationlearners.ConditionalForest, mix, inner bayes.ForecastMix) error {
	if d.pending != nil || step != d.next || step < 0 || step >= 512 || x >= 512 || p.p.Values != x&p.p.Mask {
		return fmt.Errorf("invalid interaction preparation")
	}
	required, e := interactionMask(step, cfg)
	if e != nil {
		return e
	}
	pending := interactionPending{p: p, required: required}
	long := models.Pooled
	if p.p.Split {
		long = models.Local
	}
	for i, m := range []*observation.Model{base, models.Short, nil, nil, long} {
		if m == nil {
			continue
		}
		pending.available[i] = true
		for j, mask := range []uint16{p.p.Mask, 511} {
			v, e := m.ForecastObserved(mask, x&mask)
			if e != nil {
				return e
			}
			pending.probabilities[i][j] = v
		}
	}
	for i, m := range []*observationlearners.ConditionalForest{retained, age} {
		if m == nil {
			continue
		}
		pending.available[i+2] = true
		for j, mask := range []uint16{p.p.Mask, 511} {
			v, e := m.Forecast(mask, x&mask)
			if e != nil {
				return e
			}
			pending.probabilities[i+2][j] = v
		}
	}
	for view := 0; view < 2; view++ {
		get := func(i int) float64 {
			if !pending.available[i] {
				return .5
			}
			return pending.probabilities[i][view]
		}
		experts := [4]float64{get(0), get(1), get(4), .5}
		if p.available {
			a := get(2)
			if pending.available[3] {
				a = get(3)
			}
			inside := [4]float64{get(1), get(2), a, a}
			if view == 0 {
				for i := range inside {
					if math.Abs(inside[i]-p.inner[i]) > 1e-14 {
						return fmt.Errorf("inner reconstruction")
					}
				}
			}
			experts[1] = inner.Forecast(inside)
		}
		if view == 0 {
			for i := range experts {
				if math.Abs(experts[i]-p.p.Experts[i]) > 1e-14 {
					return fmt.Errorf("expert reconstruction")
				}
			}
			if math.Abs(mix.Forecast(experts)-p.p.P) > 1e-14 {
				return fmt.Errorf("mixture reconstruction")
			}
		} else {
			pending.fullMixture = mix.Forecast(experts)
		}
	}
	pending.ageGuide = p.subsetGuide && age != nil && p.innerWeights[2]+p.innerWeights[3] > p.innerWeights[1]
	d.pending = &pending
	return nil
}
func (d *interactionDiagnostic) observe(step int, y bool) error {
	if d.pending == nil || step != d.next {
		return fmt.Errorf("missing interaction preparation")
	}
	p := d.pending
	w := &d.Windows[step/64]
	w.N++
	if p.p.p.Guide < 0 || p.p.p.Guide >= 3 {
		return fmt.Errorf("invalid diagnostic guide")
	}
	w.Guide[p.p.p.Guide]++
	if p.required != 0 {
		w.RequiredFrames++
		if p.p.p.Mask&p.required == p.required {
			w.Covered++
		}
	}
	if p.p.subsetGuide {
		w.SubsetGuide++
	}
	if p.ageGuide {
		w.AgeGuide++
	}
	target := 0.
	if y {
		target = 1
	}
	sq := func(p float64) float64 { return (p - target) * (p - target) }
	w.MixtureBrier += sq(p.p.p.P)
	w.FullMixtureBrier += sq(p.fullMixture)
	for i := range w.Weights {
		w.Weights[i] += p.p.p.Weights[i]
	}
	if p.p.available {
		w.InnerAvailable++
		w.InnerWeights[0] += p.p.innerWeights[0]
		w.InnerWeights[1] += p.p.innerWeights[1]
		w.InnerWeights[2] += p.p.innerWeights[2] + p.p.innerWeights[3]
	}
	for i, available := range p.available {
		if available {
			w.Available[i]++
			for j, v := range p.probabilities[i] {
				w.Brier[i][j] += sq(v)
			}
		}
	}
	d.next++
	d.pending = nil
	return nil
}

type interactionRecord struct {
	Original ageBreadthRecord
	Windows  [8]interactionWindow
}

func TestInteractionDiagnosticContracts(t *testing.T) {
	cfg := subsetBreadthCases[5]
	base, e := breadthBase(cfg, 2026118999)
	if e != nil {
		t.Fatal(e)
	}
	for _, schedule := range []feedbackSchedule{learningSchedules[0], learningSchedules[3]} {
		a, e := ageBreadthRun(base, "unit", 5, 0, 2026118999, schedule, cfg)
		if e != nil {
			t.Fatal(e)
		}
		var d interactionDiagnostic
		b, e := interactionDiagnosticRun(base, "unit", 5, 0, 2026118999, schedule, cfg, &d)
		if e != nil || a != b {
			t.Fatal("diagnostic changed original", e)
		}
		total := 0.
		for _, w := range d.Windows {
			if w.N != 64 || w.Covered > w.RequiredFrames || w.AgeGuide > w.SubsetGuide || w.FullMixtureBrier < 0 || w.FullMixtureBrier > 64 {
				t.Fatal("window bounds")
			}
			total += w.MixtureBrier
		}
		if math.Abs(total-a.Arms[1].Full.Brier) > 1e-10 || d.pending != nil || d.next != 512 {
			t.Fatal("score parity")
		}
		if e := d.observe(512, true); e == nil {
			t.Fatal("unprepared outcome")
		}
	}
	for _, tc := range []struct {
		step, index int
		want        uint16
	}{{0, 5, 448}, {300, 5, 30}, {300, 2, 6}, {300, 9, 0}, {300, 8, 448}} {
		m, e := interactionMask(tc.step, subsetBreadthCases[tc.index])
		if e != nil || m != tc.want {
			t.Fatal("mask contract")
		}
	}
}

func TestInteractionDiagnosticV89(t *testing.T) {
	out, replay := os.Getenv("EVENTFRAME_INTERACTION_DIAG_OUT"), os.Getenv("EVENTFRAME_INTERACTION_DIAG_REPLAY")
	if out == "" && replay == "" {
		t.Skip("opt-in")
	}
	if out != "" && replay != "" {
		t.Fatal("choose one")
	}
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(filepath.Join(root, "docs/experiments/mmm-age-breadth-v88.jsonl"))
	if e != nil {
		t.Fatal(e)
	}
	parent := json.NewDecoder(bytes.NewReader(raw))
	parent.DisallowUnknownFields()
	var ph struct {
		Version string
		Hashes  map[string]string
	}
	if e := parent.Decode(&ph); e != nil || ph.Version != "v88" {
		t.Fatal("parent header", e)
	}
	h := sha256.Sum256(raw)
	header := struct {
		Version, ParentSHA256 string
		Hashes                map[string]string
	}{"v89", hex.EncodeToString(h[:]), ph.Hashes}
	var dec *json.Decoder
	if replay != "" {
		f, e := os.Open(replay)
		if e != nil {
			t.Fatal(e)
		}
		defer f.Close()
		dec = json.NewDecoder(f)
		dec.DisallowUnknownFields()
		var want struct {
			Version, ParentSHA256 string
			Hashes                map[string]string
		}
		if e := dec.Decode(&want); e != nil || want.Version != "v89" || want.ParentSHA256 != header.ParentSHA256 {
			t.Fatal("header", e)
		}
		header = want
	} else {
		for _, p := range []string{"internal/observationgate/interaction_diagnostic_test.go", "internal/observationgate/interaction_diagnostic_run_test.go", "research/generate-interaction-diagnostic-v89.mjs", "docs/experiments/mmm-interaction-diagnostic-v89-protocol.md"} {
			b, e := os.ReadFile(filepath.Join(root, p))
			if e != nil {
				t.Fatal(e)
			}
			h := sha256.Sum256(b)
			header.Hashes[p] = hex.EncodeToString(h[:])
		}
	}
	for p, want := range header.Hashes {
		if !filepath.IsLocal(p) {
			t.Fatal("source path")
		}
		b, e := os.ReadFile(filepath.Join(root, p))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			t.Fatal("source changed", p)
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if e := enc.Encode(header); e != nil {
		t.Fatal(e)
	}
	for phase, split := range []string{"design", "confirmation"} {
		for j, cfg := range subsetBreadthCases {
			for _, schedule := range []feedbackSchedule{learningSchedules[0], learningSchedules[3]} {
				for i := 0; i < 64; i++ {
					var original ageBreadthRecord
					if e := parent.Decode(&original); e != nil {
						t.Fatal(e)
					}
					seed := int64(2026118800)*1000000 + int64(j*10000+phase*1000+i)
					if original.TrainSeed != seed || original.Config != cfg {
						t.Fatal("parent configuration")
					}
					base, e := breadthBase(cfg, seed)
					if e != nil {
						t.Fatal(e)
					}
					var d interactionDiagnostic
					r, e := interactionDiagnosticRun(base, split, j, i, int64(2026118801+10*j+phase), schedule, cfg, &d)
					if e != nil || r != original.ageRecord {
						t.Fatal("original parity", split, cfg.Name, i, e)
					}
					record := interactionRecord{original, d.Windows}
					if dec != nil {
						var want interactionRecord
						if e := dec.Decode(&want); e != nil || want != record {
							t.Fatal("diagnostic replay", e)
						}
					} else {
						if e := enc.Encode(record); e != nil {
							t.Fatal(e)
						}
					}
				}
				t.Log(split, cfg.Name, schedule.Name, "parity complete")
			}
		}
	}
	var extra any
	if e := parent.Decode(&extra); e != io.EOF {
		t.Fatal("parent trailing data", e)
	}
	if dec != nil {
		if e := dec.Decode(&extra); e != io.EOF {
			t.Fatal("trailing data", e)
		}
		return
	}
	f, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if _, e := f.Write(buf.Bytes()); e != nil {
		t.Fatal(e)
	}
	if e := f.Sync(); e != nil {
		t.Fatal(e)
	}
}
