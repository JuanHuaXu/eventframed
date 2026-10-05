package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/transfergenerator"
)

const softV118TransferBase int64 = 2184111800
const softV118BooleanBase int64 = 2188111900

type softV118Data struct {
	Initial [16]observation.Sample
	Frames  [256]transfergenerator.Observation
	Q       [256]float64
	Teacher *transfergenerator.Description
	Rules   [2]uint16
	Seeds   [5]int64
}
type softV118Step struct {
	X       uint16
	Y       bool
	Q       float64
	Delay   int
	Missing bool
	P       [8]float64
}
type softV118Fit struct {
	Clock      int
	Origins    [2][]int
	RidgeBeta  [2][ridgeDimension]float64
	RidgeStats [2]ridgeFitStats
}
type softV118Record struct {
	Phase, Case, Index, Schedule int
	Teacher                      *transfergenerator.Description
	Rules                        [2]uint16
	Seeds                        [5]int64
	Initial                      [16]observation.Sample
	Steps                        []softV118Step
	Fits                         []softV118Fit
	Metrics                      [8][2]stackV93Metric
	LogLoss                      [8][2]float64
	Realized, RealizedLog        [8]float64
	Arrived                      int
}

func softV118Generate(phase, scenario, index int, transferBase, booleanBase int64) (softV118Data, error) {
	var out softV118Data
	if phase < 0 || phase > 1 || scenario < 0 || scenario >= 21 || index < 0 || index >= 32 {
		return out, fmt.Errorf("generator identity")
	}
	if scenario < 9 {
		data, teacher, err := transfergenerator.Generate(transfergenerator.Spec{SeedBase: transferBase, Phase: phase, Family: transfergenerator.Family(scenario / 3), Mode: transfergenerator.Mode(scenario % 3), Index: index})
		if err != nil {
			return out, err
		}
		d := teacher.Describe()
		out.Teacher = &d
		out.Seeds = d.Seeds
		out.Frames = data.Frames
		for i, s := range data.Initial {
			out.Initial[i] = observation.Sample{Bits: s.X, Outcome: s.Y}
		}
		for i, s := range data.Frames {
			out.Q[i], err = teacher.Truth(s.X, i)
			if err != nil {
				return out, err
			}
		}
		return out, nil
	}
	c := scenario - 9
	seed := booleanBase + int64(phase*1000000+c*10000+index*10)
	for i := range out.Seeds {
		out.Seeds[i] = seed + int64(i)
	}
	rules, xs, ys, delays, misses := rand.New(rand.NewSource(seed)), rand.New(rand.NewSource(seed+1)), rand.New(rand.NewSource(seed+2)), rand.New(rand.NewSource(seed+3)), rand.New(rand.NewSource(seed+4))
	arities := [12]int{1, 2, 3, 4, 4, 3, 3, 0, 0, 4, 3, 4}
	for k := range out.Rules {
		arity := arities[c]
		if c >= 10 && k == 1 {
			arity = 7 - arity
		}
		if arity > 0 {
			pool := stackV93Masks(arity, phase)
			out.Rules[k] = pool[rules.Intn(len(pool))]
		}
	}
	truth := func(x uint16, t int) float64 {
		name, rule := replicationV102Cases[c], out.Rules[0]
		if c >= 10 {
			name = "majority3"
			if c == 11 {
				name = "parity4"
			}
			if t >= 128 {
				rule = out.Rules[1]
				if c == 10 {
					name = "parity4"
				} else {
					name = "majority3"
				}
			}
		}
		return stackV93Truth(x, rule, name)
	}
	for t := -16; t < 256; t++ {
		x := stackV93Input(uint16(xs.Intn(512)), replicationV102Cases[c])
		q := truth(x, t)
		y := ys.Float64() < q
		if t < 0 {
			out.Initial[t+16] = observation.Sample{Bits: x, Outcome: y}
		} else {
			out.Frames[t] = transfergenerator.Observation{X: x, Y: y, Delay: uint8(delays.Intn(32)), Missing: misses.Float64() < .2}
			out.Q[t] = q
		}
	}
	return out, nil
}

func softV118Run(data softV118Data, phase, scenario, index, schedule int) (softV118Record, error) {
	r := softV118Record{Phase: phase, Case: scenario, Index: index, Schedule: schedule, Teacher: data.Teacher, Rules: data.Rules, Seeds: data.Seeds, Initial: data.Initial}
	if schedule < 0 || schedule > 1 {
		return r, fmt.Errorf("schedule")
	}
	history := append([]observation.Sample(nil), data.Initial[:]...)
	var arrived [256]bool
	var predictions [8][512]float64
	for clock := 0; clock < 288; clock++ {
		for i, s := range r.Steps {
			if !arrived[i] && !s.Missing && i+s.Delay <= clock {
				arrived[i] = true
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
			fit := softV118Fit{Clock: clock}
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
			}
			r.Fits = append(r.Fits, fit)
		}
		x := data.Frames[clock].X
		s := softV118Step{X: x}
		for arm, table := range predictions {
			s.P[arm] = table[x]
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
		}
	}
	for _, v := range arrived {
		if v {
			r.Arrived++
		}
	}
	return r, nil
}

func TestSoftV118AsOf(t *testing.T) {
	for _, c := range []int{0, 5, 8, 12, 19, 20} {
		data, err := softV118Generate(0, c, 0, transfergenerator.ValidationSeedBase, 2160111500)
		if err != nil {
			t.Fatal(err)
		}
		r, err := softV118Run(data, 0, c, 0, 1)
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
		other, err := softV118Run(changed, 0, c, 0, 1)
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

func TestSoftV118Seeds(t *testing.T) {
	used := map[int64]bool{}
	for phase := int64(0); phase < 2; phase++ {
		for c := int64(0); c < 21; c++ {
			for index := int64(0); index < 32; index++ {
				for role := int64(0); role < 5; role++ {
					seed := softV118BooleanBase + phase*1000000 + (c-9)*10000 + index*10 + role
					if c < 9 {
						seed = softV118TransferBase + phase*1000000 + (c/3)*100000 + (c%3)*10000 + index*10 + role
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
	for _, base := range []int64{2026119000, 2030119100, 2034119200, 2038119300, 2042119400, 2046119500, 2050119600, 2054119700, 2062119900, 2066110000, 2078110100, 2082110200, 2090110300, 2100110400, 2110110600, 2120110800, 2130110900, 2140111100, 2144111200, 2150111300, 2160111500, 2176111600, 2180111700} {
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

type softV118Artifact struct {
	Version                   string
	TransferBase, BooleanBase int64
	Hashes                    map[string]string
	Records                   []softV118Record
}

func TestSoftV118(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SOFT_V118")
	if path == "" {
		t.Skip("explicit artifact required")
	}
	replay := os.Getenv("EVENTFRAME_SOFT_V118_REPLAY") == "1"
	if !replay {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("output exists or inaccessible", err)
		}
	}
	hashes := transferV116Hashes(t)
	for _, name := range []string{"docs/experiments/mmm-soft-learners-v118-protocol.md", "research/soft-learners-v118-summary.mjs", "research/ridge-challenger-component.md"} {
		raw, err := os.ReadFile(filepath.Join("../..", name))
		if err != nil {
			t.Fatal(err)
		}
		hashes[name] = fmt.Sprintf("%x", sha256.Sum256(raw))
	}
	a := softV118Artifact{Version: "soft-learners-v118", TransferBase: softV118TransferBase, BooleanBase: softV118BooleanBase, Hashes: hashes}
	var old softV118Artifact
	if replay {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(raw, &old); err != nil {
			t.Fatal(err)
		}
		if old.Version != a.Version || old.TransferBase != a.TransferBase || old.BooleanBase != a.BooleanBase || len(old.Records) != 2688 || !reflect.DeepEqual(old.Hashes, hashes) {
			t.Fatal("replay contract")
		}
	}
	for phase := 0; phase < 2; phase++ {
		for c := 0; c < 21; c++ {
			for index := 0; index < 32; index++ {
				data, err := softV118Generate(phase, c, index, a.TransferBase, a.BooleanBase)
				if err != nil {
					t.Fatal(err)
				}
				for schedule := 0; schedule < 2; schedule++ {
					r, err := softV118Run(data, phase, c, index, schedule)
					if err != nil {
						t.Fatal(phase, c, index, schedule, err)
					}
					if replay && !reflect.DeepEqual(r, old.Records[len(a.Records)]) {
						t.Fatal("record replay", phase, c, index, schedule)
					}
					a.Records = append(a.Records, r)
				}
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
	if err = json.NewEncoder(f).Encode(a); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	t.Log("wrote", len(a.Records), "records", len(hashes), "hashes")
}
