package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/transfergenerator"
)

const softV120TransferBase int64 = 2200112200
const softV120BooleanBase int64 = 2204112300

type softV120Data = softV118Data
type softV120Step struct {
	X       uint16
	Y       bool
	Q       float64
	Delay   int
	Missing bool
	P       [15]float64
}
type softV120Fit struct {
	Clock                                 int
	Origins                               [2][]int
	RidgeBeta                             [2][ridgeDimension]float64
	RidgeStats                            [2]ridgeFitStats
	VariationalMean                       [2][ridgeDimension]float64
	VariationalCov                        [2][ridgeDimension][ridgeDimension]float64
	VariationalBound, VariationalResidual [2]float64
	VariationalIterations                 [2]int
	SegmentStarts                         [2][]float64
	SegmentEvidence                       [2]float64
}
type softV120Record struct {
	Phase, Case, Index, Schedule int
	Teacher                      *transfergenerator.Description
	Rules                        [2]uint16
	Seeds                        [5]int64
	Initial                      [16]observation.Sample
	Steps                        []softV120Step
	Fits                         []softV120Fit
	Metrics                      [15][2]stackV93Metric
	LogLoss                      [15][2]float64
	Realized, RealizedLog        [15]float64
	Arrived                      int
}

func softV120Run(data softV120Data, phase, scenario, index, schedule int) (softV120Record, error) {
	r := softV120Record{Phase: phase, Case: scenario, Index: index, Schedule: schedule, Teacher: data.Teacher, Rules: data.Rules, Seeds: data.Seeds, Initial: data.Initial}
	if schedule < 0 || schedule > 1 {
		return r, fmt.Errorf("schedule")
	}
	history := append([]observation.Sample(nil), data.Initial[:]...)
	var arrived [256]bool
	var predictions [15][512]float64
	control, err := newMarkovAdvice(.001)
	if err != nil {
		return r, err
	}
	for clock := 0; clock < 288; clock++ {
		for i, s := range r.Steps {
			if !arrived[i] && !s.Missing && i+s.Delay <= clock {
				arrived[i] = true
				if err := control.deliver(uint64(i), s.Y); err != nil {
					return r, err
				}
			}
		}
		if clock >= 32 {
			if err := control.expireBefore(uint64(min(clock-31, 256))); err != nil {
				return r, err
			}
		}
		if clock >= 256 {
			continue
		}
		if clock%32 == 0 {
			available := make([]int, 0, 272)
			for i := -16; i < clock; i++ {
				if i < 0 || arrived[i] {
					available = append(available, i)
				}
			}
			fit := softV120Fit{Clock: clock}
			packets := make([]segmentPacket, clock+16)
			for i, s := range data.Initial {
				packets[i] = segmentPacket{Bits: s.Bits, Outcome: s.Outcome, Arrives: 0}
			}
			for i, s := range r.Steps {
				at := i + s.Delay
				if s.Missing {
					at = -1
				}
				packets[i+16] = segmentPacket{Bits: s.X, Outcome: s.Y, Arrives: at}
			}
			for window, cap := range []int{64, 32} {
				origins := available
				if len(origins) > cap {
					origins = origins[len(origins)-cap:]
				}
				fit.Origins[window] = append([]int(nil), origins...)
				samples := make([]observation.Sample, len(origins))
				for j, i := range origins {
					samples[j] = history[i+16]
				}
				g, err := fitSubset(samples)
				if err != nil {
					return r, err
				}
				predictions[window*2] = g.predictions
				p, err := fitBooleanSpecialist(samples)
				if err != nil {
					return r, err
				}
				predictions[window*2+1] = p.predictions
				linear, err := fitRidgeLogistic(samples)
				if err != nil {
					return r, fmt.Errorf("ridge at clock%d/window%d: %w", clock, window, err)
				}
				fit.RidgeBeta[window], fit.RidgeStats[window] = linear.beta, linear.stats
				for x := range predictions[4+window] {
					predictions[4+window][x], err = linear.predict(uint16(x))
					if err != nil {
						return r, err
					}
				}
				tree, err := fitContextTree(samples)
				if err != nil {
					return r, err
				}
				predictions[6+window] = tree.predictions
				vb, err := fitVariationalLogistic(samples)
				if err != nil {
					return r, fmt.Errorf("variational at clock%d/window%d: %w", clock, window, err)
				}
				fit.VariationalMean[window], fit.VariationalCov[window] = vb.mean, vb.cov
				fit.VariationalBound[window], fit.VariationalResidual[window], fit.VariationalIterations[window] = vb.bound, vb.residual, vb.iterations
				for x := range predictions[8+window] {
					predictions[8+window][x], err = vb.predict(uint16(x))
					if err != nil {
						return r, err
					}
				}
				segmented, err := fitSegmentPosterior(-16, clock, packets, cap, .01, .95)
				if err != nil {
					return r, fmt.Errorf("segment clock%d/window%d: %w", clock, window, err)
				}
				if !reflect.DeepEqual(segmented.origins, origins) {
					return r, fmt.Errorf("segment label mismatch")
				}
				predictions[10+window] = segmented.predictions
				fit.SegmentStarts[window] = segmented.lastStart
				fit.SegmentEvidence[window] = segmented.logEvidence
				gz, bz := math.Inf(-1), math.Inf(-1)
				for mask := range g.evidence {
					gz = segmentLogAdd(gz, segmentPrior()[mask]+g.evidence[mask])
					bz = segmentLogAdd(bz, segmentPrior()[mask]+p.evidence[mask])
				}
				z := segmentLogAdd(math.Log(.95)+gz, math.Log(.05)+bz)
				wg, wb := math.Exp(math.Log(.95)+gz-z), math.Exp(math.Log(.05)+bz-z)
				for x := range predictions[13+window] {
					predictions[13+window][x] = wg*g.predictions[x] + wb*p.predictions[x]
				}
			}
			r.Fits = append(r.Fits, fit)
		}
		x := data.Frames[clock].X
		s := softV120Step{X: x}
		for arm, table := range predictions {
			s.P[arm] = table[x]
		}
		raw := [4]float64{s.P[0], s.P[1], s.P[2], s.P[3]}
		weights := control.current.weights()
		for j, p := range raw {
			s.P[12] += weights[j+1] * p
		}
		if err := control.issue(uint64(clock), raw); err != nil {
			return r, err
		}
		// Evaluator truth and current label are accessed only after prediction.
		s.Y, s.Q = data.Frames[clock].Y, data.Q[clock]
		if schedule == 1 {
			s.Delay, s.Missing = int(data.Frames[clock].Delay), data.Frames[clock].Missing
		}
		for arm, p := range s.P {
			if math.IsNaN(p) || math.IsInf(p, 0) || p <= 0 || p >= 1 {
				return r, fmt.Errorf("invalid forecast")
			}
			brier := (p-s.Q)*(p-s.Q) + s.Q*(1-s.Q)
			acc := 1 - s.Q
			if p >= .5 {
				acc = s.Q
			}
			logLoss := -s.Q*math.Log(p) - (1-s.Q)*math.Log1p(-p)
			r.Metrics[arm][0].Brier += brier / 256
			r.Metrics[arm][0].Accuracy += acc / 256
			r.LogLoss[arm][0] += logLoss / 256
			if clock >= 192 {
				r.Metrics[arm][1].Brier += brier / 64
				r.Metrics[arm][1].Accuracy += acc / 64
				r.LogLoss[arm][1] += logLoss / 64
			}
			y := 0.
			if s.Y {
				y = 1
			}
			r.Realized[arm] += (p - y) * (p - y) / 256
			r.RealizedLog[arm] += (-y*math.Log(p) - (1-y)*math.Log1p(-p)) / 256
		}
		r.Steps = append(r.Steps, s)
		history = append(history, observation.Sample{Bits: x, Outcome: s.Y})
		if !s.Missing && s.Delay == 0 {
			arrived[clock] = true
			if err := control.deliver(uint64(clock), s.Y); err != nil {
				return r, err
			}
		}
	}
	for _, v := range arrived {
		if v {
			r.Arrived++
		}
	}
	return r, nil
}

func TestSoftV120AsOf(t *testing.T) {
	for _, c := range []int{0, 5, 8, 12, 19, 20} {
		data, err := softV118Generate(0, c, 0, transfergenerator.ValidationSeedBase, 2160111500)
		if err != nil {
			t.Fatal(err)
		}
		r, err := softV120Run(data, 0, c, 0, 1)
		if err != nil {
			t.Fatal(err)
		}
		for _, fit := range r.Fits {
			for window, cap := range []int{64, 32} {
				var want []int
				for i := -16; i < fit.Clock; i++ {
					if i < 0 || !data.Frames[i].Missing && i+int(data.Frames[i].Delay) <= fit.Clock {
						want = append(want, i)
					}
				}
				if len(want) > cap {
					want = want[len(want)-cap:]
				}
				if !reflect.DeepEqual(want, fit.Origins[window]) {
					t.Fatal("as-of mismatch")
				}
			}
		}
		changed := data
		changed.Frames[160].Y = !changed.Frames[160].Y
		for i := range changed.Q {
			changed.Q[i] = .5
		}
		other, err := softV120Run(changed, 0, c, 0, 1)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i <= 160; i++ {
			if r.Steps[i].P != other.Steps[i].P {
				t.Fatal("current/future label or oracle leaked")
			}
		}
	}
}

func TestSoftV120Seeds(t *testing.T) {
	used := map[int64]bool{}
	for phase := int64(0); phase < 2; phase++ {
		for c := int64(0); c < 21; c++ {
			for index := int64(0); index < 32; index++ {
				for role := int64(0); role < 5; role++ {
					seed := softV120BooleanBase + phase*1000000 + (c-9)*10000 + index*10 + role
					if c < 9 {
						seed = softV120TransferBase + phase*1000000 + (c/3)*100000 + (c%3)*10000 + index*10 + role
					}
					seed %= 2147483647
					if used[seed] {
						t.Fatal("within-quality seed collision")
					}
					used[seed] = true
				}
			}
		}
	}
	for _, base := range []int64{2026119000, 2030119100, 2034119200, 2038119300, 2042119400, 2046119500, 2050119600, 2054119700, 2062119900, 2066110000, 2078110100, 2082110200, 2090110300, 2100110400, 2110110600, 2120110800, 2130110900, 2140111100, 2144111200, 2150111300, 2160111500, 2176111600, 2180111700, 2184111800, 2188111900, 2192112000, 2196112100} {
		for phase := int64(0); phase < 2; phase++ {
			for c := int64(0); c < 30; c++ {
				for index := int64(0); index < 128; index++ {
					for role := int64(0); role < 5; role++ {
						if used[(base+phase*1000000+c*10000+index*10+role)%2147483647] {
							t.Fatal("archived seed collision", base)
						}
					}
				}
			}
		}
	}
	for _, base := range []int64{3070119900, 3090110000, 3110110100, 3130110200} {
		for mode := int64(0); mode < 2; mode++ {
			for family := int64(0); family < 4096; family++ {
				for test := int64(0); test < 64; test++ {
					if used[(base+mode*10000000+family*1000+test)%2147483647] {
						t.Fatal("archived null seed")
					}
				}
			}
		}
	}
}

type softV120Artifact struct {
	Version                   string
	TransferBase, BooleanBase int64
	Hashes                    map[string]string
	Records                   []softV120Record `json:"-"`
	Workers                   int
	Hazard, GenericMass       float64
}

func TestSoftV120(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SOFT_V120")
	if path == "" {
		t.Skip("explicit artifact required")
	}
	replay := os.Getenv("EVENTFRAME_SOFT_V120_REPLAY") == "1"
	if !replay {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("output exists or inaccessible", err)
		}
	}
	hashes := transferV116Hashes(t)
	for _, name := range []string{"docs/experiments/mmm-soft-learners-v120-protocol.md", "research/soft-learners-v120-summary.mjs", "research/segment-v120-reference.mjs", "research/ridge-challenger-component.md", "research/variational-logistic-component-contract.md", "research/segment-posterior-component.md"} {
		raw, err := os.ReadFile(filepath.Join("../..", name))
		if err != nil {
			t.Fatal(err)
		}
		hashes[name] = fmt.Sprintf("%x", sha256.Sum256(raw))
	}
	a := softV120Artifact{Version: "soft-learners-v120", TransferBase: softV120TransferBase, BooleanBase: softV120BooleanBase, Hashes: hashes, Workers: 4, Hazard: .01, GenericMass: .95}
	var old softV120Artifact
	if replay {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		decoder := json.NewDecoder(f)
		if err = decoder.Decode(&old); err != nil {
			f.Close()
			t.Fatal(err)
		}
		for {
			var r softV120Record
			err = decoder.Decode(&r)
			if err == io.EOF {
				break
			}
			if err != nil {
				f.Close()
				t.Fatal(err)
			}
			old.Records = append(old.Records, r)
			if len(old.Records) > 2688 {
				f.Close()
				t.Fatal("too many records")
			}
		}
		if err = f.Close(); err != nil {
			t.Fatal(err)
		}
		if old.Workers != 4 || old.Hazard != .01 || old.GenericMass != .95 || old.Version != a.Version || old.TransferBase != a.TransferBase || old.BooleanBase != a.BooleanBase || len(old.Records) != 2688 || !reflect.DeepEqual(old.Hashes, hashes) {
			t.Fatal("replay contract")
		}
	}

	for phase := 0; phase < 2; phase++ {
		for c := 0; c < 21; c++ {
			rows, err := softV120Case(phase, c, 32, 4, a.TransferBase, a.BooleanBase)
			if err != nil {
				t.Fatal(phase, c, err)
			}
			for _, r := range rows {
				if replay && !reflect.DeepEqual(r, old.Records[len(a.Records)]) {
					t.Fatal("record replay", r.Phase, r.Case, r.Index, r.Schedule)
				}
				a.Records = append(a.Records, r)
			}
			t.Log("completed phase/case", phase, c)
		}
	}

	if replay {
		t.Log("exact replay", len(a.Records), "records", len(hashes), "hashes")
		return
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	encoder := json.NewEncoder(f)
	if err = encoder.Encode(a); err != nil {
		f.Close()
		t.Fatal(err)
	}
	for _, r := range a.Records {
		if err = encoder.Encode(r); err != nil {
			f.Close()
			t.Fatal(err)
		}
	}
	if err = f.Sync(); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	t.Log("wrote", len(a.Records), "records", len(hashes), "hashes")
}

func softV120Case(phase, c, count, workers int, transferBase, booleanBase int64) ([]softV120Record, error) {
	if count < 1 || count > 32 || workers < 1 || workers > 4 {
		return nil, fmt.Errorf("worker bounds")
	}
	rows := make([]softV120Record, count*2)
	failures := make([]error, count)
	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for index := worker; index < count; index += workers {
				data, err := softV118Generate(phase, c, index, transferBase, booleanBase)
				if err != nil {
					failures[index] = err
					continue
				}
				for schedule := 0; schedule < 2; schedule++ {
					r, err := softV120Run(data, phase, c, index, schedule)
					if err != nil {
						failures[index] = err
						break
					}
					rows[index*2+schedule] = r
				}
			}
		}(worker)
	}
	wg.Wait()
	for i, err := range failures {
		if err != nil {
			return nil, fmt.Errorf("index%d: %w", i, err)
		}
	}
	return rows, nil
}

func TestSoftV120Workers(t *testing.T) {
	serial, err := softV120Case(0, 20, 4, 1, transfergenerator.ValidationSeedBase, 2160111500)
	if err != nil {
		t.Fatal(err)
	}
	parallel, err := softV120Case(0, 20, 4, 4, transfergenerator.ValidationSeedBase, 2160111500)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(serial, parallel) {
		t.Fatal("worker scheduling changed output")
	}
}

func TestSoftV120ReferenceQA(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SOFT_V120_QA")
	if path == "" {
		t.Skip("explicit consumed QA output required")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("QA output exists or inaccessible", err)
	}
	var rows []softV120Record
	for _, c := range []int{0, 20} {
		data, err := softV118Generate(0, c, 0, transfergenerator.ValidationSeedBase, 2160111500)
		if err != nil {
			t.Fatal(err)
		}
		for schedule := 0; schedule < 2; schedule++ {
			r, err := softV120Run(data, 0, c, 0, schedule)
			if err != nil {
				t.Fatal(err)
			}
			rows = append(rows, r)
		}
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, r := range rows {
		if err := enc.Encode(r); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}
