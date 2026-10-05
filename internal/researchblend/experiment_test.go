package researchblend

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchcalibration"
	"github.com/JuanHuaXu/eventframed/internal/researchpartition"
)

type traceV30 struct {
	Index                 int
	Useful                bool
	Q, Probability, Score float64
}
type armV30 struct {
	Model, Policy                                                   string
	Trace                                                           []traceV30
	Forecast                                                        []float64
	Weights                                                         [3]float64
	Packet                                                          []int
	Brier, PriorityBrier, PacketUsefulness, PacketBrier, PacketBias float64
	SetupNS, SelectionNS, UpdateNS, FinalNS, TotalNS                int64
}
type worldV30 struct {
	Kind, Geometry, Regime              string
	World, Partition                    int
	Seed                                int64
	Base, Coordinates, Rates, LeafMeans []float64
	Labels                              []bool
	Arms                                []armV30
}
type predictorV30 interface {
	Predict(int) (float64, error)
	Observe(int, bool) error
}
type localV30 struct {
	base         []float64
	seen, useful []bool
}

func (m *localV30) Predict(i int) (float64, error) {
	q := m.base[i]
	if m.seen[i] {
		y := 0.
		if m.useful[i] {
			y = 1
		}
		q = (2*q + y) / 3
	}
	return q, nil
}
func (m *localV30) Observe(i int, y bool) error { m.seen[i], m.useful[i] = true, y; return nil }

func baseV30(step float64) []float64 {
	b := []float64{.925, .925}
	for j := 1; j <= 198; j++ {
		b = append(b, .65*(math.Cos(step*float64(j))+1)/2+.275)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(b)))
	return b[:150]
}
func gammaV30(rng *rand.Rand, a float64) float64 {
	if a < 1 {
		return gammaV30(rng, a+1) * math.Pow(1-rng.Float64(), 1/a)
	}
	d := a - 1./3
	c := 1 / math.Sqrt(9*d)
	for {
		x := rng.NormFloat64()
		v := 1 + c*x
		if v <= 0 {
			continue
		}
		v = v * v * v
		u := 1 - rng.Float64()
		if u < 1-.0331*x*x*x*x || math.Log(u) < .5*x*x+d*(1-v+math.Log(v)) {
			return d * v
		}
	}
}
func betaV30(rng *rand.Rand, a, b float64) float64 {
	x, y := gammaV30(rng, a), gammaV30(rng, b)
	return x / (x + y)
}

// Fixed32 control nomination: same first-pass stratum schedule as the blend,
// without constructing or updating a hidden blend model for the controls.
func stratumV30(n, k int, rng *rand.Rand) (int, float64) {
	bucket, x := 0, k
	for bit := 0; bit < 5; bit++ {
		bucket = (bucket << 1) | (x & 1)
		x >>= 1
	}
	lo, hi := bucket*n/32, (bucket+1)*n/32
	return lo + rng.Intn(hi-lo), 1 / float64(hi-lo)
}

func runV30(base, r []float64, labels []bool, model, policy string, seed int64) armV30 {
	start := time.Now()
	a := armV30{Model: model, Policy: policy, Trace: make([]traceV30, 0, 32)}
	var p predictorV30
	var blend *Model
	var old *researchcalibration.Model
	var part *researchpartition.Model
	var err error
	switch model {
	case "blend":
		blend, err = New(base, r)
		p = blend
	case "old":
		old, err = researchcalibration.New(base)
		p = old
	case "partition":
		part, err = researchpartition.New(base, r)
		p = part
	case "local", "baseline":
		p = &localV30{base: append([]float64(nil), base...), seen: make([]bool, len(base)), useful: make([]bool, len(base))}
	default:
		panic("unknown model")
	}
	if err != nil {
		panic(err)
	}
	a.SetupNS = time.Since(start).Nanoseconds()
	rng := rand.New(rand.NewSource(seed + 303))
	if model != "baseline" {
		for k := 0; k < 32; k++ {
			s := time.Now()
			index := -1
			probability, score := 0., 0.
			if blend != nil {
				v, e := blend.Select(policy, rng)
				err = e
				index, probability, score = v.Index, v.Probability, v.Score
			} else if policy == "stratified_random" {
				index, probability = stratumV30(len(base), k, rng)
			} else if old != nil {
				index, score, err = old.Select(policy, rng)
				probability = 1
				if policy == "random" {
					probability = 1 / float64(len(base)-k)
				}
			} else {
				panic("unsupported model policy")
			}
			a.SelectionNS += time.Since(s).Nanoseconds()
			if err != nil {
				panic(err)
			}
			q, e := p.Predict(index)
			if e != nil {
				panic(e)
			}
			a.Trace = append(a.Trace, traceV30{index, labels[index], q, probability, score})
			s = time.Now()
			if e = p.Observe(index, labels[index]); e != nil {
				panic(e)
			}
			a.UpdateNS += time.Since(s).Nanoseconds()
		}
	}
	s := time.Now()
	a.Forecast = make([]float64, len(base))
	for i := range base {
		a.Forecast[i], err = p.Predict(i)
		if err != nil {
			panic(err)
		}
	}
	if blend != nil {
		a.Weights = blend.Weights()
	}
	a.FinalNS = time.Since(s).Nanoseconds()
	a.TotalNS = time.Since(start).Nanoseconds()
	return a
}
func scoreV30(a *armV30, rates []float64) {
	indices := make([]int, len(rates))
	for i, q := range a.Forecast {
		indices[i] = i
		v := (q-rates[i])*(q-rates[i]) + rates[i]*(1-rates[i])
		a.Brier += v / 150
		w := 1.
		if i < 10 {
			w = 3
		}
		a.PriorityBrier += w * v / 170
	}
	sort.SliceStable(indices, func(i, j int) bool { return a.Forecast[indices[i]] > a.Forecast[indices[j]] })
	a.Packet = append([]int(nil), indices[:10]...)
	for _, i := range a.Packet {
		q, p := a.Forecast[i], rates[i]
		a.PacketUsefulness += p / 10
		a.PacketBrier += ((q-p)*(q-p) + p*(1-p)) / 10
		a.PacketBias += (q - p) / 10
	}
}

var regimesV30 = []string{"independent", "aligned", "reversed", "calibrated", "curved", "shifted_peak", "alternating", "phase_alternating", "permuted_curved", "baseline_matched", "tree_matched", "narrow_peak"}
var armsV30 = [][2]string{{"baseline", "head"}, {"local", "stratified_random"}, {"old", "stratified_random"}, {"old", "random"}, {"old", "uncertainty"}, {"old", "information"}, {"partition", "stratified_random"}, {"blend", "stratified_random"}, {"blend", "random"}, {"blend", "uncertainty"}, {"blend", "family_information"}}

func makeWorldV30(seedBase int64, g, regimeIndex, w int) worldV30 {
	step := .005
	geo := "tight"
	if g == 1 {
		step = .02
		geo = "wide"
	}
	seed := seedBase + int64(g*10000000+regimeIndex*1000000+w*1000)
	r := worldV30{Kind: "world", Geometry: geo, Regime: regimesV30[regimeIndex], World: w, Partition: -1, Seed: seed, Base: baseV30(step), Coordinates: make([]float64, 150), Rates: make([]float64, 150), Labels: make([]bool, 150)}
	truth, draw := rand.New(rand.NewSource(seed+101)), rand.New(rand.NewSource(seed+202))
	trees := treeReferences(0, 8, 0, 0)
	if r.Regime == "tree_matched" {
		u, mass := truth.Float64(), 0.
		for k, p := range trees {
			mass += p.prior
			if u < mass {
				r.Partition = k
				break
			}
		}
		if r.Partition < 0 {
			panic("invalid tree draw")
		}
		r.LeafMeans = make([]float64, 15)
		for _, node := range trees[r.Partition].nodes {
			if r.LeafMeans[node] == 0 {
				v := truth.Float64()
				for v == 0 {
					v = truth.Float64()
				}
				r.LeafMeans[node] = v
			}
		}
	}
	perm := truth.Perm(150)
	for i := range r.Base {
		x := float64(i) / 149
		r.Coordinates[i] = x
		p := 0.
		switch r.Regime {
		case "independent":
			p = .2
			if perm[i] < 75 {
				p = .8
			}
		case "aligned":
			p = .9 - .8*x
		case "reversed":
			p = .1 + .8*x
		case "calibrated":
			p = r.Base[i]
		case "curved":
			p = .1 + .8*math.Pow(math.Sin(math.Pi*x), 2)
		case "shifted_peak", "narrow_peak":
			center := .25 + .5*float64(w%8)/7
			width := .14
			if r.Regime == "narrow_peak" {
				width = .04
			}
			p = .1 + .8*math.Exp(-math.Pow((x-center)/width, 2))
		case "alternating", "phase_alternating":
			p = .2
			if (i%2 == 0) == (r.Regime == "alternating") {
				p = .8
			}
		case "permuted_curved":
			p = .1 + .8*math.Pow(math.Sin(math.Pi*float64(perm[i])/149), 2)
		case "baseline_matched":
			p = betaV30(truth, 2*r.Base[i], 2*(1-r.Base[i]))
		case "tree_matched":
			mean := r.LeafMeans[trees[r.Partition].nodes[min(7, int(8*x))]]
			p = betaV30(truth, 2*mean, 2*(1-mean))
		}
		if math.IsNaN(p) || p < 0 || p > 1 {
			panic("invalid true probability")
		}
		r.Rates[i] = p
		r.Labels[i] = draw.Float64() < p
	}
	for _, mp := range armsV30 {
		a := runV30(r.Base, r.Coordinates, r.Labels, mp[0], mp[1], seed)
		scoreV30(&a, r.Rates)
		r.Arms = append(r.Arms, a)
	}
	return r
}

func TestNoFutureAndMatchedControlsV30(t *testing.T) {
	b := baseV30(.02)
	r := make([]float64, 150)
	labels := make([]bool, 150)
	for i := range r {
		r[i] = float64(i) / 149
		labels[i] = i%3 == 0
	}
	for _, policy := range []string{"random", "stratified_random", "uncertainty", "family_information"} {
		a := runV30(b, r, labels, "blend", policy, 913)
		changed := append([]bool(nil), labels...)
		seen := make([]bool, 150)
		for _, v := range a.Trace {
			seen[v.Index] = true
		}
		for i := range changed {
			if !seen[i] {
				changed[i] = !changed[i]
			}
		}
		c := runV30(b, r, changed, "blend", policy, 913)
		if !reflect.DeepEqual(a.Trace, c.Trace) || !reflect.DeepEqual(a.Forecast, c.Forecast) || a.Weights != c.Weights {
			t.Fatal("unused labels changed decisions or predictions")
		}
	}
	a := runV30(b, r, labels, "blend", "stratified_random", 913)
	for _, model := range []string{"local", "old", "partition"} {
		c := runV30(b, r, labels, model, "stratified_random", 913)
		for i, v := range a.Trace {
			if v.Index != c.Trace[i].Index || v.Useful != c.Trace[i].Useful || v.Probability != c.Trace[i].Probability {
				t.Fatal("matched evidence differs")
			}
		}
	}
}

func TestExperimentV30(t *testing.T) {
	out := os.Getenv("EVENTFRAME_BLEND_V30_OUT")
	if out == "" {
		t.Skip("opt-in outcome experiment")
	}
	split := os.Getenv("EVENTFRAME_BLEND_V30_SPLIT")
	seedBase := int64(2026103003)
	if split == "confirmation" {
		seedBase = 2026103004
	} else if split != "design" {
		t.Fatal("unknown split")
	}
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	hashes := map[string]string{}
	for _, path := range []string{"internal/researchblend/model.go", "internal/researchblend/model_test.go", "internal/researchblend/experiment_test.go", "internal/researchcalibration/model.go", "internal/researchpartition/model.go", "research/blend-v30-verify.mjs", "docs/experiments/mmm-blend-v30-protocol.md", "docs/experiments/mmm-blend-v30-preflight.md"} {
		b, e := os.ReadFile(filepath.Join(root, path))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		hashes[path] = hex.EncodeToString(h[:])
	}
	f, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if e = enc.Encode(map[string]any{"Kind": "manifest", "Split": split, "SeedBase": seedBase, "Worlds": 768, "Sources": hashes}); e != nil {
		t.Fatal(e)
	}
	for g := 0; g < 2; g++ {
		for regimeIndex := range regimesV30 {
			for w := 0; w < 32; w++ {
				r := makeWorldV30(seedBase, g, regimeIndex, w)
				if e = enc.Encode(r); e != nil {
					t.Fatal(e)
				}
			}
		}
	}
	if e = f.Sync(); e != nil {
		t.Fatal(e)
	}
}
