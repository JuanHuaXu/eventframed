package researchpartition

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
)

type traceV29 struct {
	Index  int
	Useful bool
	Q      float64
}
type armV29 struct {
	Model, Policy, Setting                                                            string
	Trace                                                                             []traceV29
	Forecast                                                                          []float64
	Weights                                                                           [Trees + 1]float64
	Packet                                                                            []int
	Brier, PriorityBrier, PacketUsefulness, PacketBrier, PacketBias                   float64
	Budget, Used, Unused, LabelCost, SetupNS, SelectionNS, UpdateNS, FinalNS, TotalNS int64
}
type worldV29 struct {
	Kind, Geometry, Regime              string
	World, Partition                    int
	Seed                                int64
	Base, Coordinates, Rates, LeafMeans []float64
	Labels                              []bool
	Arms                                []armV29
}
type predictorV29 interface {
	Predict(int) (float64, error)
	Observe(int, bool) error
}

func baseV29(step float64) []float64 {
	b := []float64{.925, .925}
	for i := 0; i < 198; i++ {
		b = append(b, .65*(math.Cos(step*float64(i+1))+1)/2+.275)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(b)))
	return b[:150]
}
func gammaV29(rng *rand.Rand, a float64) float64 {
	if a < 1 {
		return gammaV29(rng, a+1) * math.Pow(1-rng.Float64(), 1/a)
	}
	d := a - 1.0/3
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
func betaV29(rng *rand.Rand, a, b float64) float64 {
	x, y := gammaV29(rng, a), gammaV29(rng, b)
	return x / (x + y)
}

func TestBetaGeneratorV29(t *testing.T) {
	for _, ab := range [][2]float64{{.04, 1.96}, {.4, 1.6}, {1.6, .4}, {2, 2}} {
		rng := rand.New(rand.NewSource(881903))
		sum, square := 0.0, 0.0
		for i := 0; i < 100000; i++ {
			v := betaV29(rng, ab[0], ab[1])
			if math.IsNaN(v) || v < 0 || v > 1 {
				t.Fatal("invalid sampled probability")
			}
			sum += v
			square += v * v
		}
		mean := ab[0] / (ab[0] + ab[1])
		variance := mean * (1 - mean) / (ab[0] + ab[1] + 1)
		if math.Abs(sum/100000-mean) > .005 || math.Abs(square/100000-sum*sum/1e10-variance) > .005 {
			t.Fatal("Beta sampler moments failed")
		}
	}
}
func selectionCostV29(n int, policy string) int64 {
	switch policy {
	case "head", "random":
		return int64(n + 1)
	case "stratified":
		size, bits := 1, 0
		for size < n {
			size *= 2
			bits++
		}
		return int64(size * bits)
	case "uncertainty":
		return int64(n*(3*27+12) + n)
	case "information":
		return int64(n*(6*27+12) + n)
	}
	panic("unsupported selection")
}

func runV29(base, coordinate []float64, labels []bool, model, policy, setting string, labelCost, seed int64) armV29 {
	start := time.Now()
	a := armV29{Model: model, Policy: policy, Setting: setting, LabelCost: labelCost}
	var p predictorV29
	var part *Model
	var old *researchcalibration.Model
	var err error
	if model == "partition" {
		part, err = New(base, coordinate)
		p = part
	} else if model == "old" {
		old, err = researchcalibration.New(base)
		p = old
	} else {
		panic("unsupported model")
	}
	if err != nil {
		panic(err)
	}
	a.SetupNS = time.Since(start).Nanoseconds()
	n := len(base)
	setup := int64(20 * n * 27)
	final := int64(3*n*27 + n*int(math.Ceil(math.Log2(float64(n)))))
	update := int64(5*27 + 16)
	a.Used = setup + final
	if setting != "fixed32" {
		a.Budget = setup + final + 32*(labelCost+update+selectionCostV29(n, "information"))
	}
	rng := rand.New(rand.NewSource(seed + 303))
	for k := 0; k < n; k++ {
		if setting == "fixed32" && k == 32 {
			break
		}
		cost := selectionCostV29(n, policy) + update + labelCost
		if a.Budget > 0 && a.Used+cost > a.Budget {
			break
		}
		s := time.Now()
		index := -1
		if part != nil {
			index, err = part.Select(policy, rng)
		} else {
			index, _, err = old.Select(policy, rng)
		}
		a.SelectionNS += time.Since(s).Nanoseconds()
		if err != nil {
			panic(err)
		}
		q, _ := p.Predict(index)
		a.Trace = append(a.Trace, traceV29{index, labels[index], q})
		s = time.Now()
		err = p.Observe(index, labels[index])
		a.UpdateNS += time.Since(s).Nanoseconds()
		if err != nil {
			panic(err)
		}
		a.Used += cost
	}
	s := time.Now()
	a.Forecast = make([]float64, n)
	for i := range a.Forecast {
		a.Forecast[i], _ = p.Predict(i)
	}
	if part != nil {
		a.Weights = part.weights
	}
	a.FinalNS = time.Since(s).Nanoseconds()
	a.TotalNS = time.Since(start).Nanoseconds()
	if a.Budget > 0 {
		a.Unused = a.Budget - a.Used
	}
	return a
}

func controlV29(base []float64, labels []bool, model string) armV29 {
	a := armV29{Model: model, Policy: "head", Setting: "fixed32", Forecast: append([]float64(nil), base...)}
	if model == "local" {
		for i := 0; i < 32; i++ {
			y := 0.0
			if labels[i] {
				y = 1
			}
			a.Trace = append(a.Trace, traceV29{i, labels[i], base[i]})
			a.Forecast[i] = (2*base[i] + y) / 3
		}
	}
	return a
}
func scoreV29(a *armV29, rates []float64) {
	n := len(rates)
	a.Packet = make([]int, n)
	for i, q := range a.Forecast {
		p := rates[i]
		risk := (q-p)*(q-p) + p*(1-p)
		a.Brier += risk / float64(n)
		w := 1.0
		if i < 10 {
			w = 3
		}
		a.PriorityBrier += w * risk / float64(n+20)
		a.Packet[i] = i
	}
	sort.SliceStable(a.Packet, func(i, j int) bool { return a.Forecast[a.Packet[i]] > a.Forecast[a.Packet[j]] })
	a.Packet = a.Packet[:10]
	for _, i := range a.Packet {
		q, p := a.Forecast[i], rates[i]
		a.PacketUsefulness += p / 10
		a.PacketBias += (q - p) / 10
		a.PacketBrier += ((q-p)*(q-p) + p*(1-p)) / 10
	}
}

func TestFutureAndEvidenceParityV29(t *testing.T) {
	b, r := inputs(150)
	labels := make([]bool, 150)
	for i := range labels {
		labels[i] = i%3 == 0
	}
	for _, policy := range []string{"stratified", "uncertainty", "random"} {
		a := runV29(b, r, labels, "partition", policy, "fixed32", 0, 913)
		changed := append([]bool(nil), labels...)
		seen := map[int]bool{}
		for _, v := range a.Trace {
			seen[v.Index] = true
		}
		for i := range changed {
			if !seen[i] {
				changed[i] = !changed[i]
			}
		}
		c := runV29(b, r, changed, "partition", policy, "fixed32", 0, 913)
		if !reflect.DeepEqual(a.Trace, c.Trace) || !reflect.DeepEqual(a.Forecast, c.Forecast) || a.Weights != c.Weights {
			t.Fatal("unselected labels entered inference")
		}
	}
	a, c := runV29(b, r, labels, "partition", "stratified", "fixed32", 0, 913), runV29(b, r, labels, "old", "stratified", "fixed32", 0, 913)
	for i, v := range a.Trace {
		if v.Index != c.Trace[i].Index || v.Useful != c.Trace[i].Useful {
			t.Fatal("matched evidence differs")
		}
	}
	for _, model := range []string{"partition", "old"} {
		for _, cost := range []int64{100000, 1000000} {
			v := runV29(b, r, labels, model, "stratified", "cost", cost, 913)
			if v.Used > v.Budget || v.Unused != v.Budget-v.Used || len(v.Trace) < 32 || len(v.Trace) > 150 {
				t.Fatal("invalid budget")
			}
		}
	}
}

func TestExperimentV29(t *testing.T) {
	out := os.Getenv("EVENTFRAME_PARTITION_V29_OUT")
	if out == "" {
		t.Skip("opt-in research experiment")
	}
	split := os.Getenv("EVENTFRAME_PARTITION_V29_SPLIT")
	seedBase := int64(2026102903)
	if split == "confirmation" {
		seedBase = 2026102904
	} else if split != "design" {
		t.Fatal("unknown split")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := map[string]string{}
	for _, path := range []string{"internal/researchpartition/model.go", "internal/researchpartition/model_test.go", "internal/researchpartition/experiment_test.go", "internal/researchcalibration/model.go", "research/partition-v29-verify.mjs", "docs/experiments/mmm-partition-v29-protocol.md"} {
		b, e := os.ReadFile(filepath.Join(root, path))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		hashes[path] = hex.EncodeToString(h[:])
	}
	f, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err = enc.Encode(map[string]any{"Kind": "manifest", "Split": split, "SeedBase": seedBase, "Worlds": 576, "Sources": hashes}); err != nil {
		t.Fatal(err)
	}
	for g, geo := range []string{"tight", "wide"} {
		step := .005
		if g == 1 {
			step = .020
		}
		base := baseV29(step)
		coordinate := make([]float64, 150)
		for i := range coordinate {
			coordinate[i] = float64(i) / 149
		}
		for regimeIndex, regime := range []string{"independent", "aligned", "reversed", "calibrated", "curved", "shifted_peak", "alternating", "baseline_matched", "tree_matched"} {
			for w := 0; w < 32; w++ {
				seed := seedBase + int64(g*10000000+regimeIndex*1000000+w*1000)
				truth := rand.New(rand.NewSource(seed + 101))
				draw := rand.New(rand.NewSource(seed + 202))
				record := worldV29{Kind: "world", Geometry: geo, Regime: regime, World: w, Partition: -1, Seed: seed, Base: base, Coordinates: coordinate, Rates: make([]float64, 150), Labels: make([]bool, 150)}
				part, _ := New(base, coordinate)
				if regime == "tree_matched" {
					u, mass := truth.Float64(), 0.0
					for k, p := range part.partitions {
						mass += p.prior
						if u < mass {
							record.Partition = k
							break
						}
					}
					if record.Partition < 0 {
						t.Fatal("tree prior draw failed")
					}
					record.LeafMeans = make([]float64, 15)
					for _, node := range part.partitions[record.Partition].node {
						if record.LeafMeans[node] == 0 {
							v := truth.Float64()
							for v == 0 {
								v = truth.Float64()
							}
							record.LeafMeans[node] = v
						}
					}
				}
				perm := truth.Perm(150)
				for i, r := range coordinate {
					p := 0.0
					switch regime {
					case "independent":
						p = .2
						if perm[i] < 75 {
							p = .8
						}
					case "aligned":
						p = .9 - .8*r
					case "reversed":
						p = .1 + .8*r
					case "calibrated":
						p = base[i]
					case "curved":
						p = .1 + .8*math.Pow(math.Sin(math.Pi*r), 2)
					case "shifted_peak":
						center := .25 + .5*float64(w%8)/7
						p = .1 + .8*math.Exp(-math.Pow((r-center)/.14, 2))
					case "alternating":
						p = .2
						if i%2 == 0 {
							p = .8
						}
					case "baseline_matched":
						p = betaV29(truth, 2*base[i], 2*(1-base[i]))
					case "tree_matched":
						node := part.partitions[record.Partition].node[part.bin[i]]
						mean := record.LeafMeans[node]
						p = betaV29(truth, 2*mean, 2*(1-mean))
					}
					if math.IsNaN(p) || p < 0 || p > 1 {
						t.Fatal("invalid true rate")
					}
					record.Rates[i] = p
					record.Labels[i] = draw.Float64() < p
				}
				record.Arms = append(record.Arms, controlV29(base, record.Labels, "baseline"), controlV29(base, record.Labels, "local"))
				for _, setting := range []struct {
					Name string
					Cost int64
				}{{"fixed32", 0}, {"cost100000", 100000}, {"cost1000000", 1000000}} {
					for _, modelPolicy := range [][2]string{{"old", "stratified"}, {"old", "information"}, {"partition", "head"}, {"partition", "random"}, {"partition", "stratified"}, {"partition", "uncertainty"}} {
						record.Arms = append(record.Arms, runV29(base, coordinate, record.Labels, modelPolicy[0], modelPolicy[1], setting.Name, setting.Cost, seed))
					}
				}
				for i := range record.Arms {
					scoreV29(&record.Arms[i], record.Rates)
				}
				if err = enc.Encode(record); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if err = f.Sync(); err != nil {
		t.Fatal(err)
	}
}
