package researchcalibration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

type traceV27 struct {
	Index    int
	Useful   bool
	Q, Score float64
}
type armV27 struct {
	Policy, Setting                                                 string
	Trace                                                           []traceV27
	Forecast                                                        []float64
	Weights                                                         [Hypotheses]float64
	Packet                                                          []int
	Brier, PriorityBrier, PacketUsefulness, PacketBrier, PacketBias float64
	Budget, Used, Unused, LabelCost                                 int64
	SetupNS, SelectionNS, UpdateNS, FinalNS, TotalNS                int64
}
type worldV27 struct {
	Kind, Geometry, Regime string
	World                  int
	Seed                   int64
	Theta                  int
	Base, Rates            []float64
	Labels                 []bool
	Arms                   []armV27
}

func baseV27(step float64) []float64 {
	b := []float64{.925, .925}
	for i := 0; i < 198; i++ {
		b = append(b, .65*(math.Cos(step*float64(i+1))+1)/2+.275)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(b)))
	return b[:150]
}

// Fixture-only Marsaglia-Tsang gamma sampler (2000), DOI 10.1145/358407.358414.
// For shape<1 use Gamma(a+1)*U^(1/a); independent gammas give a Beta rate.
func gammaV27(rng *rand.Rand, a float64) float64 {
	if a < 1 {
		return gammaV27(rng, a+1) * math.Pow(1-rng.Float64(), 1/a)
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
func betaV27(rng *rand.Rand, a, b float64) float64 {
	x, y := gammaV27(rng, a), gammaV27(rng, b)
	return x / (x + y)
}

func TestBetaGeneratorV27(t *testing.T) {
	for _, ab := range [][2]float64{{.04, 1.96}, {.4, 1.6}, {2, 2}, {1.96, .04}} {
		rng := rand.New(rand.NewSource(729111))
		n := 100000.0
		s, s2 := 0.0, 0.0
		for i := 0; i < int(n); i++ {
			v := betaV27(rng, ab[0], ab[1])
			if math.IsNaN(v) || v < 0 || v > 1 {
				t.Fatal("invalid beta rate")
			}
			s += v
			s2 += v * v
		}
		mean := ab[0] / (ab[0] + ab[1])
		variance := mean * (1 - mean) / (ab[0] + ab[1] + 1)
		if math.Abs(s/n-mean) > .005 || math.Abs(s2/n-s*s/n/n-variance) > .005 {
			t.Fatalf("bad sampler moments %v %g %g", ab, s/n, s2/n-s*s/n/n)
		}
	}
}

func scoreV27(a *armV27, rates []float64) {
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

func runV27(base []float64, labels []bool, policy, setting string, labelCost, seed int64) armV27 {
	start := time.Now()
	a := armV27{Policy: policy, Setting: setting, LabelCost: labelCost}
	m, err := New(base)
	if err != nil {
		panic(err)
	}
	a.SetupNS = time.Since(start).Nanoseconds()
	n := len(base)
	setup := int64(20 * n * Hypotheses)
	final := int64(3*n*Hypotheses + n*int(math.Ceil(math.Log2(float64(n)))))
	update := int64(5 * Hypotheses)
	a.Used = setup + final
	if setting != "fixed32" {
		a.Budget = setup + final + 32*(labelCost+update+SelectionCost(n, "information"))
	}
	rng := rand.New(rand.NewSource(seed + 303))
	for k := 0; k < n; k++ {
		if setting == "fixed32" && k == 32 {
			break
		}
		cost := SelectionCost(n, policy) + update + labelCost
		if a.Budget > 0 && a.Used+cost > a.Budget {
			break
		}
		s := time.Now()
		i, sc, err := m.Select(policy, rng)
		a.SelectionNS += time.Since(s).Nanoseconds()
		if err != nil {
			panic(err)
		}
		q, _ := m.Predict(i)
		a.Trace = append(a.Trace, traceV27{i, labels[i], q, sc})
		s = time.Now()
		err = m.Observe(i, labels[i])
		a.UpdateNS += time.Since(s).Nanoseconds()
		if err != nil {
			panic(err)
		}
		a.Used += cost
	}
	s := time.Now()
	a.Forecast = make([]float64, n)
	for i := range a.Forecast {
		a.Forecast[i], _ = m.Predict(i)
	}
	a.Weights = m.weights
	a.FinalNS = time.Since(s).Nanoseconds()
	a.TotalNS = time.Since(start).Nanoseconds()
	if a.Budget > 0 {
		a.Unused = a.Budget - a.Used
	}
	return a
}

func localV27(base []float64, labels []bool) armV27 {
	a := armV27{Policy: "local", Setting: "fixed32", Forecast: append([]float64(nil), base...)}
	for i := 0; i < 32; i++ {
		y := 0.0
		if labels[i] {
			y = 1
		}
		a.Trace = append(a.Trace, traceV27{Index: i, Useful: labels[i], Q: base[i]})
		a.Forecast[i] = (2*base[i] + y) / 3
	}
	return a
}

func TestBudgetAndFutureBoundaryV27(t *testing.T) {
	b := baseV27(.02)
	x, y := make([]bool, 150), make([]bool, 150)
	for i := 32; i < 150; i++ {
		y[i] = true
	}
	a, c := runV27(b, x, "head", "fixed32", 0, 13), runV27(b, y, "head", "fixed32", 0, 13)
	for i := range a.Forecast {
		if a.Forecast[i] != c.Forecast[i] {
			t.Fatal("unobserved future labels changed forecast")
		}
	}
	for _, policy := range []string{"head", "random", "stratified", "uncertainty", "information"} {
		for _, cost := range []int64{10000, 100000, 1000000} {
			v := runV27(b, x, policy, fmt.Sprint(cost), cost, 13)
			if v.Used > v.Budget || v.Unused != v.Budget-v.Used || len(v.Trace) < 32 {
				t.Fatal("invalid budget admission")
			}
			if policy == "information" && len(v.Trace) != 32 {
				t.Fatal("information anchor changed")
			}
		}
	}
}

func TestExperimentV27(t *testing.T) {
	out := os.Getenv("EVENTFRAME_CALIBRATION_V27_OUT")
	if out == "" {
		t.Skip("opt-in research experiment")
	}
	split := os.Getenv("EVENTFRAME_CALIBRATION_V27_SPLIT")
	seedBase := int64(2026102703)
	if split == "confirmation" {
		seedBase = 2026102704
	} else if split != "design" {
		t.Fatal("unknown split")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := map[string]string{}
	for _, p := range []string{"internal/researchcalibration/model.go", "internal/researchcalibration/model_test.go", "internal/researchcalibration/experiment_test.go", "research/calibration-observation-v27-verify.mjs", "docs/experiments/mmm-calibration-observation-v27-protocol.md"} {
		b, e := os.ReadFile(filepath.Join(root, p))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	f, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err = enc.Encode(map[string]any{"Kind": "manifest", "Split": split, "SeedBase": seedBase, "Sources": hashes, "Worlds": 384}); err != nil {
		t.Fatal(err)
	}
	for g, geo := range []string{"tight", "wide"} {
		step := .005
		if g == 1 {
			step = .020
		}
		base := baseV27(step)
		for r, regime := range []string{"independent", "aligned", "reversed", "calibrated", "smooth", "matched"} {
			for w := 0; w < 32; w++ {
				seed := seedBase + int64(g*10000000+r*1000000+w*1000)
				truth := rand.New(rand.NewSource(seed + 101))
				draw := rand.New(rand.NewSource(seed + 202))
				record := worldV27{Kind: "world", Geometry: geo, Regime: regime, World: w, Seed: seed, Theta: -1, Base: base, Rates: make([]float64, 150), Labels: make([]bool, 150)}
				matched, _ := New(base)
				if regime == "matched" {
					u := truth.Float64()
					sum := 0.0
					for h, weight := range matched.weights {
						sum += weight
						if u < sum {
							record.Theta = h
							break
						}
					}
					if record.Theta < 0 {
						t.Fatal("prior draw failed")
					}
				}
				perm := truth.Perm(150)
				for i := range record.Rates {
					rank := float64(i) / 149
					p := 0.0
					switch regime {
					case "independent":
						p = .2
						if perm[i] < 75 {
							p = .8
						}
					case "aligned":
						p = .9 - .8*rank
					case "reversed":
						p = .1 + .8*rank
					case "calibrated":
						p = base[i]
					case "smooth":
						p = .1 + .8*math.Pow(math.Sin(math.Pi*rank), 2)
					case "matched":
						mean := matched.p[i*Hypotheses+record.Theta]
						p = betaV27(truth, 2*mean, 2*(1-mean))
					}
					record.Rates[i] = p
					record.Labels[i] = draw.Float64() < p
				}
				record.Arms = append(record.Arms, localV27(base, record.Labels))
				for _, setting := range []struct {
					Name string
					Cost int64
				}{{"fixed32", 0}, {"cost10000", 10000}, {"cost100000", 100000}, {"cost1000000", 1000000}} {
					for _, policy := range []string{"head", "random", "uncertainty", "information", "stratified"} {
						record.Arms = append(record.Arms, runV27(base, record.Labels, policy, setting.Name, setting.Cost, seed))
					}
				}
				for i := range record.Arms {
					scoreV27(&record.Arms[i], record.Rates)
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
