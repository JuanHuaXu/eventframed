package observationlearners

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type oneChangeFit struct {
	Predictions                [512]float64
	Cuts                       []int
	Weights, LeftLog, RightLog []float64
	LogEvidence                float64
}

// A single coherent snapshot model. Cut zero denotes H0; every other cut
// uses independent parameter draws on disjoint sides of the SAME evidence.
// No hidden boundary, arrival schedule, or teacher value enters this fitter.
func fitOneChange(samples []observation.Sample) (*oneChangeFit, error) {
	table, err := buildSegmentLikelihoodsBatch(context.Background(), samples, .95)
	if err != nil {
		return nil, err
	}
	n := len(samples)
	m := &oneChangeFit{Cuts: []int{0}, LeftLog: []float64{0}, RightLog: []float64{table.logM[0][n]}}
	logs := []float64{table.logM[0][n]}
	if n >= 16 {
		logs[0] += math.Log(.9)
		for cut := 8; cut <= n-8; cut++ {
			m.Cuts = append(m.Cuts, cut)
			m.LeftLog = append(m.LeftLog, table.logM[0][cut])
			m.RightLog = append(m.RightLog, table.logM[cut][n])
			logs = append(logs, math.Log(.1/float64(n-15))+table.logM[0][cut]+table.logM[cut][n])
		}
	}
	m.LogEvidence = math.Inf(-1)
	for _, l := range logs {
		m.LogEvidence = segmentLogAdd(m.LogEvidence, l)
	}
	for i, l := range logs {
		w := math.Exp(l - m.LogEvidence)
		m.Weights = append(m.Weights, w)
		for x, p := range table.tail[m.Cuts[i]] {
			m.Predictions[x] += w * p
		}
	}
	for _, p := range m.Predictions {
		if math.IsNaN(p) || p <= 0 || p >= 1 {
			return nil, fmt.Errorf("invalid one-change predictive")
		}
	}
	return m, nil
}

type oneChangeRecord struct {
	Phase, Case, Index, Schedule, Clock int
	Origins                             []int
	Fit                                 *oneChangeFit
	Predictions                         [32]float64
	Error                               string `json:",omitempty"`
}

func oneChangeSnapshot(input softV120Record, clock int) oneChangeRecord {
	r := oneChangeRecord{Phase: input.Phase, Case: input.Case, Index: input.Index, Schedule: input.Schedule, Clock: clock}
	if len(input.Steps) != 256 || clock < 0 || clock > 224 {
		r.Error = "invalid snapshot"
		return r
	}
	for j := -16; j < clock; j++ {
		if j < 0 || (!input.Steps[j].Missing && j+input.Steps[j].Delay <= clock) {
			r.Origins = append(r.Origins, j)
		}
	}
	r.Origins = r.Origins[max(0, len(r.Origins)-64):]
	samples := make([]observation.Sample, len(r.Origins))
	for i, j := range r.Origins {
		if j < 0 {
			samples[i] = input.Initial[j+16]
		} else {
			samples[i] = observation.Sample{Bits: input.Steps[j].X, Outcome: input.Steps[j].Y}
		}
	}
	var err error
	r.Fit, err = fitOneChange(samples)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	for j := range r.Predictions {
		x := input.Steps[clock+j].X
		if x >= 512 {
			r.Error = "invalid query"
			return r
		}
		r.Predictions[j] = r.Fit.Predictions[x]
	}
	return r
}

func TestOneChangeContracts(t *testing.T) {
	for _, n := range []int{0, 1, 15, 16, 18} {
		samples := ridgeTestSamples(n)
		before := append([]observation.Sample{}, samples...)
		m, err := fitOneChange(samples)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, append([]observation.Sample{}, samples...)) {
			t.Fatal("input mutation")
		}
		if len(m.Cuts) != 1+max(0, n-15) {
			t.Fatal("cut support")
		}
		den := segmentTestMarginal(samples, .95)
		if n >= 16 {
			den *= .9
		}
		for cut := 8; cut <= n-8; cut++ {
			den += .1 / float64(n-15) * segmentTestMarginal(samples[:cut], .95) * segmentTestMarginal(samples[cut:], .95)
		}
		if math.Abs(m.LogEvidence-math.Log(den)) > 1e-11 {
			t.Fatal("joint evidence")
		}
		total := 0.
		for i, cut := range m.Cuts {
			w := segmentTestMarginal(samples[cut:], .95)
			if n >= 16 {
				if cut == 0 {
					w *= .9
				} else {
					w *= .1 / float64(n-15) * segmentTestMarginal(samples[:cut], .95)
				}
			}
			if math.Abs(m.Weights[i]-w/den) > 1e-11 {
				t.Fatal("joint posterior weight")
			}
			total += m.Weights[i]
		}
		if math.Abs(total-1) > 1e-12 {
			t.Fatal("normalization")
		}
		for _, x := range []uint16{0, 17, 511} {
			p := 0.
			for i, cut := range m.Cuts {
				tail := append([]observation.Sample{}, samples[cut:]...)
				z := segmentTestMarginal(tail, .95)
				tail = append(tail, observation.Sample{Bits: x, Outcome: true})
				p += m.Weights[i] * segmentTestMarginal(tail, .95) / z
			}
			if math.Abs(p-m.Predictions[x]) > 1e-11 {
				t.Fatal("predictive numerator")
			}
		}
	}
	for _, bad := range [][]observation.Sample{make([]observation.Sample, 65), {{Bits: 512}}} {
		if _, err := fitOneChange(bad); err == nil {
			t.Fatal("bad sample accepted")
		}
	}
	samples := ridgeTestSamples(18)
	want, err := fitOneChange(samples)
	if err != nil {
		t.Fatal(err)
	}
	var got [4]*oneChangeFit
	var errs [4]error
	var wg sync.WaitGroup
	for i := range got {
		wg.Add(1)
		go func(i int) { defer wg.Done(); got[i], errs[i] = fitOneChange(samples) }(i)
	}
	wg.Wait()
	for i := range got {
		if errs[i] != nil || !reflect.DeepEqual(got[i], want) {
			t.Fatal("concurrent drift")
		}
	}
	t.Log("five support/evidence fixtures, 15 direct predictive ratios, and four concurrent fits PASS")
}

func TestOneChangeAsOf(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("explicit consumed source required")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var h softV120Artifact
	if err := d.Decode(&h); err != nil {
		t.Fatal(err)
	}
	checked := 0
	for {
		var input softV120Record
		err := d.Decode(&input)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if input.Phase != 0 || input.Index != 0 || input.Schedule != 1 || (input.Case != 0 && input.Case != 20) {
			continue
		}
		before, _ := json.Marshal(input)
		const clock = 160
		r := oneChangeSnapshot(input, clock)
		if r.Error != "" {
			t.Fatal(r.Error)
		}
		mutant := input
		mutant.Steps = append([]softV120Step{}, input.Steps...)
		for j := range mutant.Steps {
			s := &mutant.Steps[j]
			s.Q = 1 - s.Q
			if j >= clock || s.Missing || j+s.Delay > clock {
				s.Y = !s.Y
			}
		}
		if !reflect.DeepEqual(r, oneChangeSnapshot(mutant, clock)) {
			t.Fatal("unavailable outcome leaked")
		}
		for i, j := range r.Origins {
			if i > 0 && j <= r.Origins[i-1] {
				t.Fatal("duplicate/unordered label")
			}
			if j >= 0 && (j >= clock || input.Steps[j].Missing || j+input.Steps[j].Delay > clock) {
				t.Fatal("ineligible label")
			}
		}
		after, _ := json.Marshal(input)
		if string(before) != string(after) {
			t.Fatal("tape mutated")
		}
		checked++
	}
	if checked != 2 {
		t.Fatal("fixture count", checked)
	}
	t.Log("two full64 snapshots: future/current/unavailable Y and Q poisoning, as-of origins, ownership PASS")
}

var oneChangeSink *oneChangeFit

func BenchmarkOneChange64(b *testing.B) {
	samples := ridgeTestSamples(64)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var err error
		oneChangeSink, err = fitOneChange(samples)
		if err != nil {
			b.Fatal(err)
		}
	}
}
