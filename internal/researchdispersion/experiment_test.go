package researchdispersion

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

var regimesV34 = []string{"aligned", "reversed", "calibrated", "curved", "independent", "permuted_curved", "narrow_peak", "baseline_matched", "mean_half", "mean8", "mean32", "mean_shared"}
var cutsV34 = []int{32, 150, 300, 600, 2400}
var modesV34 = []string{"adaptive", "fixed2", "shared", "local", "baseline"}
var filesV34 = []string{"internal/researchdispersion/model.go", "internal/researchdispersion/model_test.go", "internal/researchdispersion/experiment_test.go", "docs/experiments/mmm-dispersion-v34-preflight.md", "docs/experiments/mmm-dispersion-v34-protocol.md"}

type trialV34 struct {
	Index, Ordinal int
	Useful         bool
	Probability    float64
	Q              [4]float64
}

type costV34 struct{ SetupNS, ProbeNS, UpdateNS, SnapshotNS, EarlierSnapshotsNS, AccountedNS int64 }
type snapV34 struct {
	Model                                                           string
	Budget                                                          int
	Forecast                                                        []float64
	Dispersion                                                      [Strengths]float64
	Packet                                                          []int
	Brier, PriorityBrier, PacketUsefulness, PacketBrier, PacketBias float64
	Costs                                                           costV34
}
type worldV34 struct {
	Kind, Geometry, Regime  string
	World                   int
	Seed                    int64
	Base, Rates             []float64
	Outcomes                [][]bool
	Trace                   []trialV34
	Snapshots               []snapV34
	NominationNS, ElapsedNS int64
}
type manifestV34 struct {
	Kind, Split string
	SeedBase    int64
	Worlds      int
	Sources     map[string]string
}

func gammaV34(r *rand.Rand, a float64) float64 {
	if a < 1 {
		u := r.Float64()
		for u == 0 {
			u = r.Float64()
		}
		return gammaV34(r, a+1) * math.Pow(u, 1/a)
	}
	d, c := a-1./3, 1/math.Sqrt(9*(a-1./3))
	for {
		x := r.NormFloat64()
		v := 1 + c*x
		if v <= 0 {
			continue
		}
		v = v * v * v
		u := r.Float64()
		if u < 1-.0331*x*x*x*x || math.Log(u) < .5*x*x+d*(1-v+math.Log(v)) {
			return d * v
		}
	}
}

func makeV34(baseSeed int64, geometry, regime, id int) worldV34 {
	w := worldV34{Kind: "world", Geometry: []string{"tight", "wide"}[geometry], Regime: regimesV34[regime], World: id, Seed: baseSeed + int64(geometry*100000000+regime*1000000+id*1000)}
	step := []float64{.005, .02}[geometry]
	w.Base = []float64{.925, .925}
	for i := 1; i <= 198; i++ {
		w.Base = append(w.Base, .65*(math.Cos(step*float64(i))+1)/2+.275)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(w.Base)))
	w.Base = w.Base[:150]
	w.Rates, w.Outcomes = make([]float64, 150), make([][]bool, 16)
	truth, draw := rand.New(rand.NewSource(w.Seed+101)), rand.New(rand.NewSource(w.Seed+202))
	perm := truth.Perm(150)
	for i, b := range w.Base {
		x, p, c := float64(i)/149, 0., 0.
		switch w.Regime {
		case "aligned":
			p = .9 - .8*x
		case "reversed":
			p = .1 + .8*x
		case "calibrated":
			p = b
		case "curved", "permuted_curved":
			if w.Regime == "permuted_curved" {
				x = float64(perm[i]) / 149
			}
			p = .1 + .8*math.Pow(math.Sin(math.Pi*x), 2)
		case "narrow_peak":
			center := .25 + .5*float64(id%8)/7
			p = .1 + .8*math.Exp(-.5*math.Pow((x-center)/.04, 2))
		case "independent":
			p = .2
			if perm[i] < 75 {
				p = .8
			}
		case "baseline_matched":
			p, c = b, 2
		default:
			p = (.9*b - .05) / .8
			c = map[string]float64{"mean_half": .5, "mean8": 8, "mean32": 32}[w.Regime]
		}
		if c > 0 {
			a, z := gammaV34(truth, c*p), gammaV34(truth, c*(1-p))
			p = a / (a + z)
		}
		w.Rates[i] = p
	}
	for k := range w.Outcomes {
		w.Outcomes[k] = make([]bool, 150)
		for i, p := range w.Rates {
			w.Outcomes[k][i] = draw.Float64() < p
		}
	}
	return w
}

func runV34(w *worldV34, stop int) error {
	if stop < 1 || stop > 2400 {
		return fmt.Errorf("invalid stop")
	}
	start := time.Now()
	var models [3]*Model
	var costs [5]costV34
	for j, mode := range modesV34[:3] {
		s := time.Now()
		var err error
		models[j], err = New(w.Base, mode)
		if err != nil {
			return err
		}
		costs[j].SetupNS = time.Since(s).Nanoseconds()
	}
	var localN, localS [150]int
	s := time.Now()
	rng := rand.New(rand.NewSource(w.Seed + 303))
	w.NominationNS = time.Since(s).Nanoseconds()
	for round := 0; round < 16; round++ {
		s = time.Now()
		order := rng.Perm(150)
		w.NominationNS += time.Since(s).Nanoseconds()
		for position, i := range order {
			tr := trialV34{Index: i, Ordinal: round + 1, Probability: 1 / float64(150-position)}
			for j, m := range models {
				s = time.Now()
				var err error
				tr.Q[j], err = m.Predict(i)
				costs[j].ProbeNS += time.Since(s).Nanoseconds()
				if err != nil {
					return err
				}
			}
			s = time.Now()
			tr.Q[3] = (2*w.Base[i] + float64(localS[i])) / (2 + float64(localN[i]))
			costs[3].ProbeNS += time.Since(s).Nanoseconds()
			tr.Useful = w.Outcomes[round][i]
			for j, m := range models {
				s = time.Now()
				if err := m.Observe(i, round+1, tr.Useful); err != nil {
					return err
				}
				costs[j].UpdateNS += time.Since(s).Nanoseconds()
			}
			s = time.Now()
			localN[i]++
			if tr.Useful {
				localS[i]++
			}
			costs[3].UpdateNS += time.Since(s).Nanoseconds()
			w.Trace = append(w.Trace, tr)
			count := len(w.Trace)
			for _, cut := range cutsV34 {
				if count != cut {
					continue
				}
				for j, mode := range modesV34 {
					s = time.Now()
					ss := snapV34{Model: mode, Budget: cut, Forecast: make([]float64, 150)}
					for k, b := range w.Base {
						if j < 3 {
							ss.Forecast[k], _ = models[j].Predict(k)
						} else if mode == "local" {
							ss.Forecast[k] = (2*b + float64(localS[k])) / (2 + float64(localN[k]))
						} else {
							ss.Forecast[k] = b
						}
					}
					if j < 3 {
						ss.Dispersion = models[j].Dispersion()
					}
					ss.Costs = costs[j]
					ss.Costs.SnapshotNS = time.Since(s).Nanoseconds()
					ss.Costs.AccountedNS = ss.Costs.SetupNS + ss.Costs.ProbeNS + ss.Costs.UpdateNS + ss.Costs.SnapshotNS + ss.Costs.EarlierSnapshotsNS
					costs[j].EarlierSnapshotsNS += ss.Costs.SnapshotNS
					w.Snapshots = append(w.Snapshots, ss)
				}
			}
			if count == stop {
				w.ElapsedNS = time.Since(start).Nanoseconds()
				return nil
			}
		}
	}
	return nil
}

func scoreV34(s *snapV34, rates []float64) {
	indices := make([]int, 150)
	for i, p := range rates {
		indices[i] = i
		q := s.Forecast[i]
		v := (q-p)*(q-p) + p*(1-p)
		weight := 1.
		if i < 10 {
			weight = 3
		}
		s.Brier += v / 150
		s.PriorityBrier += weight * v / 170
	}
	sort.SliceStable(indices, func(i, j int) bool { return s.Forecast[indices[i]] > s.Forecast[indices[j]] })
	s.Packet = append([]int(nil), indices[:10]...)
	for _, i := range s.Packet {
		p, q := rates[i], s.Forecast[i]
		s.PacketUsefulness += p / 10
		s.PacketBrier += ((q-p)*(q-p) + p*(1-p)) / 10
		s.PacketBias += (q - p) / 10
	}
}

func rootV34(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}
func hashV34(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func TestExperimentV34(t *testing.T) {
	out, split := os.Getenv("EVENTFRAME_DISPERSION_V34_OUT"), os.Getenv("EVENTFRAME_DISPERSION_V34_SPLIT")
	if out == "" {
		t.Skip("explicit output required")
	}
	if !filepath.IsAbs(out) || split != "design" && split != "confirmation" {
		t.Fatal("invalid output/split")
	}
	seed := int64(2026103403)
	if split == "confirmation" {
		seed++
	}
	manifest := manifestV34{"manifest", split, seed, 768, map[string]string{}}
	for _, p := range filesV34 {
		b, err := os.ReadFile(filepath.Join(rootV34(t), p))
		if err != nil {
			t.Fatal(err)
		}
		manifest.Sources[p] = hashV34(b)
	}
	f, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	enc := json.NewEncoder(writer)
	if err := enc.Encode(manifest); err != nil {
		t.Fatal(err)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesV34 {
			for k := 0; k < 32; k++ {
				w := makeV34(seed, g, r, k)
				if err := runV34(&w, 2400); err != nil {
					t.Fatal(err)
				}
				for j := range w.Snapshots {
					scoreV34(&w.Snapshots[j], w.Rates)
				}
				if err := enc.Encode(w); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if err := writer.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}

func clearV34(w *worldV34) {
	w.ElapsedNS, w.NominationNS = 0, 0
	for j := range w.Snapshots {
		w.Snapshots[j].Costs = costV34{}
	}
}

func TestNoFutureTrialsV34(t *testing.T) {
	w := makeV34(2026103499, 1, 3, 0)
	if err := runV34(&w, 2400); err != nil {
		t.Fatal(err)
	}
	clearV34(&w)
	for _, cut := range cutsV34[:4] {
		v := makeV34(2026103499, 1, 3, 0)
		seen := map[[2]int]bool{}
		for _, tr := range w.Trace[:cut] {
			seen[[2]int{tr.Ordinal - 1, tr.Index}] = true
		}
		for r, row := range v.Outcomes {
			for i := range row {
				if !seen[[2]int{r, i}] {
					v.Outcomes[r][i] = !row[i]
				}
			}
		}
		if err := runV34(&v, cut); err != nil {
			t.Fatal(err)
		}
		clearV34(&v)
		if !reflect.DeepEqual(v.Trace, w.Trace[:cut]) || !reflect.DeepEqual(v.Snapshots, w.Snapshots[:len(v.Snapshots)]) {
			t.Fatal("unarrived trial changed past", cut)
		}
	}
}
