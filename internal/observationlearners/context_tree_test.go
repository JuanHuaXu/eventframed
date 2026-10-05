package observationlearners

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"math/bits"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// Independent reference: enumerate prior choices recursively and multiply
// ordinary Beta sequence probabilities, without context indexing or log weights.
func directTreeEvidence(samples []observation.Sample, used uint16) float64 {
	leaf, yes := 1., 0
	for i, s := range samples {
		p := (float64(yes) + .5) / (float64(i) + 1)
		if s.Outcome {
			yes++
		} else {
			p = 1 - p
		}
		leaf *= p
	}
	depth := bits.OnesCount16(used)
	if depth == 3 {
		return leaf
	}
	total := .5 * leaf
	for b := 0; b < 9; b++ {
		bit := uint16(1) << b
		if used&bit != 0 {
			continue
		}
		var children [2][]observation.Sample
		for _, s := range samples {
			i := 0
			if s.Bits&bit != 0 {
				i = 1
			}
			children[i] = append(children[i], s)
		}
		total += .5 / float64(9-depth) * directTreeEvidence(children[0], used|bit) * directTreeEvidence(children[1], used|bit)
	}
	return total
}

func TestContextTreeEvidenceAndSymmetry(t *testing.T) {
	samples := []observation.Sample{{Bits: 0}, {Bits: 7, Outcome: true}, {Bits: 19}, {Bits: 511, Outcome: true}}
	original := append([]observation.Sample(nil), samples...)
	m, err := fitContextTree(samples)
	if err != nil {
		t.Fatal(err)
	}
	z := directTreeEvidence(samples, 0)
	for _, x := range []uint16{0, 7, 123, 511} {
		s := append(append([]observation.Sample(nil), samples...), observation.Sample{Bits: x, Outcome: true})
		want := directTreeEvidence(s, 0) / z
		if math.Abs(m.predictions[x]-want) > 1e-12 {
			t.Fatalf("x=%d got %g want %g", x, m.predictions[x], want)
		}
	}
	if !reflect.DeepEqual(samples, original) {
		t.Fatal("fit mutated labels")
	}
	permuted := append([]observation.Sample(nil), samples...)
	complement := append([]observation.Sample(nil), samples...)
	rotate := func(x uint16) uint16 { return ((x << 1) & 511) | (x >> 8) }
	for i := range samples {
		permuted[i].Bits = rotate(samples[i].Bits)
		complement[i].Outcome = !samples[i].Outcome
	}
	mp, _ := fitContextTree(permuted)
	mc, _ := fitContextTree(complement)
	mr, _ := fitContextTree(samples)
	for x, p := range m.predictions {
		if p <= 0 || p >= 1 || math.IsNaN(p) {
			t.Fatal("invalid probability")
		}
		if math.Abs(p-mp.predictions[rotate(uint16(x))]) > 1e-12 || math.Abs(p+mc.predictions[x]-1) > 1e-12 {
			t.Fatal("symmetry violated")
		}
	}
	if *m != *mr {
		t.Fatal("nondeterministic replay")
	}
	for _, bad := range [][]observation.Sample{nil, make([]observation.Sample, 257), {{Bits: 512}}} {
		if _, e := fitContextTree(bad); e == nil {
			t.Fatal("accepted invalid samples")
		}
	}
}

func contextTruth(x uint16, offset, family int) bool {
	a, b, c := x&(1<<offset) != 0, x&(1<<(offset+1)) != 0, x&(1<<(offset+2)) != 0
	switch family {
	case 0:
		return (a && b) || (a && c) || (b && c)
	case 1:
		if c {
			return b
		}
		return a
	case 2:
		return a != (b != c)
	default:
		return (a != (b != c)) != (x&(1<<(offset+3)) != 0)
	}
}

type contextScreenRow struct {
	Phase                  string
	Seed                   int64
	Family, NewLabels      int
	Noise                  float64
	SubsetBrier, TreeBrier float64
	SubsetFitNS, TreeFitNS int64
}

func TestContextTreeV70Experiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_CONTEXT_TREE_ARTIFACT")
	if path == "" {
		t.Skip("opt-in research experiment")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	sources := make(map[string]string)
	paths, err := filepath.Glob(filepath.Join(root, "internal/observationlearners/*.go"))
	if err != nil {
		t.Fatal(err)
	}
	opaths, err := filepath.Glob(filepath.Join(root, "internal/observation/*.go"))
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, opaths...)
	paths = append(paths, filepath.Join(root, "docs/experiments/mmm-context-tree-v70-protocol.md"), filepath.Join(root, "go.mod"), filepath.Join(root, "go.sum"))
	for _, p := range paths {
		data, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(data)
		rel, e := filepath.Rel(root, p)
		if e != nil {
			t.Fatal(e)
		}
		sources[rel] = hex.EncodeToString(h[:])
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(struct {
		Version, Go string
		Sources     map[string]string
	}{"v70", runtime.Version(), sources}); err != nil {
		t.Fatal(err)
	}
	rows := 0
	for phase, base := range []int64{2026117001, 2026117002} {
		for family := 0; family < 4; family++ {
			for ni, noise := range []float64{0, .1} {
				for wi, newLabels := range []int{0, 16, 32, 64} {
					for fit := 0; fit < 8; fit++ {
						seed := base + int64(100000*family+10000*ni+1000*wi+10*fit)
						rng := rand.New(rand.NewSource(seed))
						samples := make([]observation.Sample, 64)
						for i := range samples {
							x := uint16(rng.Intn(512))
							offset := 5
							if i >= 64-newLabels {
								offset = 0
							}
							y := contextTruth(x, offset, family)
							if rng.Float64() < noise {
								y = !y
							}
							samples[i] = observation.Sample{Bits: x, Outcome: y}
						}
						row := contextScreenRow{Phase: []string{"design", "confirmation"}[phase], Seed: seed, Family: family, NewLabels: newLabels, Noise: noise}
						var subset *subsetModel
						var tree *contextTree
						fitSubsetArm := func() {
							start := time.Now()
							subset, err = fitSubset(samples)
							row.SubsetFitNS = time.Since(start).Nanoseconds()
							if err != nil {
								t.Fatal(err)
							}
						}
						fitTreeArm := func() {
							start := time.Now()
							tree, err = fitContextTree(samples)
							row.TreeFitNS = time.Since(start).Nanoseconds()
							if err != nil {
								t.Fatal(err)
							}
						}
						if fit%2 == 0 {
							fitSubsetArm()
							fitTreeArm()
						} else {
							fitTreeArm()
							fitSubsetArm()
						}
						offset := 0
						if newLabels == 0 {
							offset = 5
						}
						for x := 0; x < 512; x++ {
							q := noise
							if contextTruth(uint16(x), offset, family) {
								q = 1 - noise
							}
							loss := func(p float64) float64 { return (p-q)*(p-q) + q*(1-q) }
							row.SubsetBrier += loss(subset.predictions[x]) / 512
							row.TreeBrier += loss(tree.predictions[x]) / 512
						}
						if err := enc.Encode(row); err != nil {
							t.Fatal(err)
						}
						rows++
					}
				}
			}
		}
	}
	if rows != 512 {
		t.Fatal(rows)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %d paired fits", rows)
}

func BenchmarkContextTreeFit64(b *testing.B) {
	samples := make([]observation.Sample, 64)
	for i := range samples {
		samples[i] = observation.Sample{Bits: uint16(i * 7), Outcome: i%3 == 0}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := fitContextTree(samples); err != nil {
			b.Fatal(err)
		}
	}
}

var contextPredictionSink float64

func BenchmarkContextTreeLookup(b *testing.B) {
	m, err := fitContextTree([]observation.Sample{{Bits: 0}})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		contextPredictionSink = m.predictions[i&511]
	}
}
