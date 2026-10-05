package observationgate

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestStartPoolDirectAndDominance(t *testing.T) {
	for scenario := range Names {
		g, old, inverse := StartPoolGate{}, Gate{}, StartPoolGate{}
		var wealth [8][2][5]float64
		for j := range wealth {
			for k := range wealth[j] {
				for r := range wealth[j][k] {
					wealth[j][k][r] = 1
				}
			}
		}
		var expected [2]bool
		rng := rand.New(rand.NewSource(2026117100 + int64(scenario)))
		previous := 0
		for tick := 0; tick < 512; tick++ {
			d := difference(scenario, tick, previous, rng)
			previous = d
			a, err := g.Observe(float64(d))
			if err != nil {
				t.Fatal(err)
			}
			o, _ := old.Observe(float64(d))
			inv, _ := inverse.Observe(float64(-d))
			if a[0] != o[0] || a[1] != o[1] || a != inv {
				t.Fatal("control/sign parity")
			}
			var total [2]float64
			for j := range wealth {
				for k, sign := range []float64{1, -1} {
					for r, rate := range []float64{.05, .15, .25, .5, .8} {
						if tick >= j*64 {
							wealth[j][k][r] *= 1 + rate*(sign*float64(d)-.15)
						}
						if r == 2 {
							total[0] += wealth[j][k][r] / 16
						}
						total[1] += wealth[j][k][r] / 80
					}
				}
			}
			for arm := 0; arm < 2; arm++ {
				expected[arm] = expected[arm] || total[arm] >= 100
				if a[arm+2] != expected[arm] || (a[arm] && !a[arm+2]) {
					t.Fatal("pool arithmetic/dominance")
				}
			}
		}
		before := g
		for _, bad := range []float64{math.NaN(), math.Inf(1), -1.1, 1.1} {
			if _, err := g.Observe(bad); err == nil || g != before {
				t.Fatal("invalid input mutated state")
			}
		}
	}
}

// Invert the binomial lower tail. This bounds harmful paired discordance,
// not the difference of two independent miss proportions.
func discordanceUpper(k, n int, alpha float64) float64 {
	if k == n {
		return 1
	}
	if k == 0 {
		return -math.Expm1(math.Log(alpha) / float64(n))
	}
	lo, hi := 0., 1.
	lgN, _ := math.Lgamma(float64(n + 1))
	for iteration := 0; iteration < 80; iteration++ {
		p := (lo + hi) / 2
		cdf := 0.
		for i := 0; i <= k; i++ {
			lgI, _ := math.Lgamma(float64(i + 1))
			lgR, _ := math.Lgamma(float64(n - i + 1))
			cdf += math.Exp(lgN - lgI - lgR + float64(i)*math.Log(p) + float64(n-i)*math.Log1p(-p))
		}
		if cdf > alpha {
			lo = p
		} else {
			hi = p
		}
	}
	return hi
}

func TestDiscordanceBound(t *testing.T) {
	if math.Abs(discordanceUpper(0, 10, .05)-(1-math.Pow(.05, .1))) > 1e-14 {
		t.Fatal("zero-count upper")
	}
	// For n=2,k=1, CDF=1-p^2, so inversion is sqrt(1-alpha).
	if math.Abs(discordanceUpper(1, 2, .05)-math.Sqrt(.95)) > 1e-12 {
		t.Fatal("analytic upper")
	}
	last := 0.
	for k := 0; k <= 32; k++ {
		u := discordanceUpper(k, 32, .01)
		if u < last || u > 1 {
			t.Fatal("invalid bound")
		}
		last = u
	}
}

type poolMissBound struct {
	Split, Scenario, Arm string
	N, Harmful           int
	Excess, Upper        float64
}
type poolOutput struct {
	Output
	MissBounds []poolMissBound
}

func TestStartPoolV71Experiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_START_POOL_ARTIFACT")
	if path == "" {
		t.Skip("opt-in research experiment")
	}
	o := poolOutput{}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	o.Hashes = make(map[string]string)
	paths, err := filepath.Glob(filepath.Join(root, "internal/observationgate/*.go"))
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, filepath.Join(root, "docs/experiments/mmm-start-pool-v71-protocol.md"), filepath.Join(root, "go.mod"), filepath.Join(root, "go.sum"))
	for _, path := range paths {
		data, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(data)
		rel, e := filepath.Rel(root, path)
		if e != nil {
			t.Fatal(e)
		}
		o.Hashes[rel] = hex.EncodeToString(h[:])
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	for phase, split := range []string{"design", "confirmation"} {
		for scenario, name := range Names {
			for stream := 0; stream < 512; stream++ {
				seed := int64(2026117101+phase)*1000000 + int64(scenario*1000+stream)
				rng := rand.New(rand.NewSource(seed))
				g := StartPoolGate{}
				old := Gate{}
				r := Record{Split: split, Scenario: name, Seed: seed, First: [4]int{-1, -1, -1, -1}}
				previous := 0
				for tick := 0; tick < 512; tick++ {
					d := difference(scenario, tick, previous, rng)
					previous = d
					r.Differences = append(r.Differences, d)
					a, e := g.Observe(float64(d))
					if e != nil {
						t.Fatal(e)
					}
					baseline, _ := old.Observe(float64(d))
					if a[0] != baseline[0] || a[1] != baseline[1] || (a[0] && !a[2]) || (a[1] && !a[3]) {
						t.Fatal("control/dominance mismatch")
					}
					for arm, v := range a {
						if v && r.First[arm] < 0 {
							r.First[arm] = tick
						}
					}
				}
				o.Records = append(o.Records, r)
			}
		}
	}
	Summarize(&o.Output)
	names := []string{"fixed", "grid", "pooled_fixed", "pooled_grid"}
	for i := range o.Summary {
		o.Summary[i].Arm = names[i%4]
	}
	for _, split := range []string{"design", "confirmation"} {
		for scenario := 5; scenario < len(Names); scenario++ {
			for arm := 1; arm < 4; arm++ {
				b := poolMissBound{Split: split, Scenario: Names[scenario], Arm: names[arm]}
				for _, r := range o.Records {
					if r.Split != split || r.Scenario != Names[scenario] {
						continue
					}
					b.N++
					cm, bm := r.First[arm] < change(scenario), r.First[0] < change(scenario)
					if cm {
						b.Excess++
					}
					if bm {
						b.Excess--
					}
					if cm && !bm {
						b.Harmful++
					}
				}
				b.Excess /= float64(b.N)
				b.Upper = discordanceUpper(b.Harmful, b.N, .05/30)
				o.MissBounds = append(o.MissBounds, b)
				if split == "confirmation" {
					o.Pass[arm-1] = o.Pass[arm-1] && b.Excess <= .01 && b.Upper <= .02
				}
			}
		}
	}
	gz := gzip.NewWriter(f)
	if err := json.NewEncoder(gz).Encode(o); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Logf("streams=%d candidate pass=%v", len(o.Records), o.Pass)
}

func BenchmarkStartPool(b *testing.B) {
	for _, pool := range []bool{false, true} {
		name := "fixed_grid_control"
		if pool {
			name = "start_pool"
		}
		b.Run(name, func(b *testing.B) {
			g, base := StartPoolGate{}, Gate{}
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if i%512 == 0 {
					g = StartPoolGate{}
					base = Gate{}
				}
				d := float64(i%3 - 1)
				if pool {
					_, _ = g.Observe(d)
				} else {
					_, _ = base.Observe(d)
				}
			}
		})
	}
}
