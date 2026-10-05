package researchcalibration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func riskCostV28(n int) int64 { return int64(6*n*Hypotheses*Hypotheses + 12*n*Hypotheses + 40*n) }

func runV28(base []float64, labels []bool, policy, setting string, labelCost, seed int64) armV27 {
	start := time.Now()
	a := armV27{Policy: policy, Setting: setting, LabelCost: labelCost}
	m, err := New(base)
	if err != nil {
		panic(err)
	}
	a.SetupNS = time.Since(start).Nanoseconds()
	n := len(base)
	priority := make([]float64, n)
	for i := range priority {
		priority[i] = 1
		if i < 10 {
			priority[i] = 3
		}
	}
	setup := int64(20 * n * Hypotheses)
	final := int64(3*n*Hypotheses + n*int(math.Ceil(math.Log2(float64(n)))))
	update := int64(5 * Hypotheses)
	selection := SelectionCost(n, policy)
	if policy == "risk" {
		selection = riskCostV28(n)
	}
	a.Used = setup + final
	if setting != "fixed32" {
		a.Budget = setup + final + 32*(labelCost+update+riskCostV28(n))
	}
	rng := rand.New(rand.NewSource(seed + 303))
	for k := 0; k < n; k++ {
		if setting == "fixed32" && k == 32 {
			break
		}
		cost := selection + update + labelCost
		if a.Budget > 0 && a.Used+cost > a.Budget {
			break
		}
		s := time.Now()
		i, score := -1, 0.0
		if policy == "risk" {
			i, score, err = m.SelectRisk(priority)
		} else {
			i, score, err = m.Select(policy, rng)
		}
		a.SelectionNS += time.Since(s).Nanoseconds()
		if err != nil {
			panic(err)
		}
		q, _ := m.Predict(i)
		a.Trace = append(a.Trace, traceV27{i, labels[i], q, score})
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

func TestBudgetAndFutureBoundaryV28(t *testing.T) {
	b := baseV27(.02)
	labels := make([]bool, 150)
	for i := range labels {
		labels[i] = i%3 == 0
	}
	a := runV28(b, labels, "risk", "fixed32", 0, 113)
	changed := append([]bool(nil), labels...)
	observed := map[int]bool{}
	for _, v := range a.Trace {
		observed[v.Index] = true
	}
	for i := range changed {
		if !observed[i] {
			changed[i] = !changed[i]
		}
	}
	c := runV28(b, changed, "risk", "fixed32", 0, 113)
	if !reflect.DeepEqual(a.Trace, c.Trace) || !reflect.DeepEqual(a.Forecast, c.Forecast) || a.Weights != c.Weights {
		t.Fatal("unseen future tape entered risk acquisition")
	}
	for _, p := range []string{"risk", "random", "uncertainty", "information"} {
		for _, cost := range []int64{10000, 100000, 1000000} {
			v := runV28(b, labels, p, "cost", cost, 113)
			if v.Used > v.Budget || v.Unused != v.Budget-v.Used || len(v.Trace) < 32 || len(v.Trace) > 150 {
				t.Fatal("budget admission violated")
			}
			if p == "risk" && len(v.Trace) != 32 {
				t.Fatal("risk anchor changed")
			}
		}
	}
}

func TestExperimentV28(t *testing.T) {
	out := os.Getenv("EVENTFRAME_RISK_V28_OUT")
	if out == "" {
		t.Skip("opt-in research experiment")
	}
	split := os.Getenv("EVENTFRAME_RISK_V28_SPLIT")
	seedBase := int64(2026102803)
	if split == "confirmation" {
		seedBase = 2026102804
	} else if split != "design" {
		t.Fatal("unknown split")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := map[string]string{}
	for _, p := range []string{"internal/researchcalibration/model.go", "internal/researchcalibration/risk.go", "internal/researchcalibration/risk_test.go", "internal/researchcalibration/risk_experiment_test.go", "internal/researchcalibration/experiment_test.go", "research/risk-acquisition-v28-verify.mjs", "docs/experiments/mmm-risk-acquisition-v28-protocol.md"} {
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
	if err = enc.Encode(map[string]any{"Kind": "manifest", "Split": split, "SeedBase": seedBase, "Worlds": 384, "Sources": hashes}); err != nil {
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
					u, sum := truth.Float64(), 0.0
					for h, v := range matched.weights {
						sum += v
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
					for _, policy := range []string{"random", "uncertainty", "information", "risk"} {
						record.Arms = append(record.Arms, runV28(base, record.Labels, policy, setting.Name, setting.Cost, seed))
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
