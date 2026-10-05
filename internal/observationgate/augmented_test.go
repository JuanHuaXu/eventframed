package observationgate

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAugmentationContracts(t *testing.T) {
	qs := [][4]float64{{.25, .25, .25, .25}, {.1, .7, .1, .1}, {.4, .4, .1, .1}}
	ms := [][4]float64{{}, {1, -1, 1, -1}, {.8, .8, 0, 0}, {-1, 1, 1, 1}}
	mu := [4]float64{.8, -.1, -.1, 0}
	clipped := false
	for _, q := range qs {
		for _, m := range ms {
			eta, c, e := augmentation(q, m)
			if e != nil || eta < 0 || eta > 1 {
				t.Fatal("invalid coefficient", e)
			}
			clipped = clipped || eta < 1
			got, want := 0., 0.
			for i := range q {
				got += q[i] * (.25*mu[i]/q[i] + eta*c[i])
				want += mu[i] / 4
				for _, d := range []float64{-1, 1} {
					z := .25*d/q[i] + eta*c[i]
					for _, s := range []float64{-1, 1} {
						if 1+.25*(s*z-.15) < .08-1e-12 {
							t.Fatal("factor bound")
						}
					}
				}
			}
			if math.Abs(got-want) > 1e-12 {
				t.Fatal("incorrect model biased evidence")
			}
		}
	}
	if !clipped {
		t.Fatal("coefficient cap not exercised")
	}
	g, old := augmentedGate{}, importanceGate{}
	for i := 0; i < 512; i++ {
		d := float64(i%3 - 1)
		q := [4]float64{.4, .4, .1, .1}
		a, e := g.observe(d, i%4, q, [4]float64{})
		b, _ := old.observe(d, q[i%4])
		if e != nil || a != b {
			t.Fatal("zero-model parity")
		}
	}
	before := g
	for _, q := range [][4]float64{{.1, .1, .1, .1}, {math.NaN(), .25, .25, .25}, {0, .5, .25, .25}} {
		if _, e := g.observe(0, 0, q, [4]float64{}); e == nil || g != before {
			t.Fatal("invalid proposal state mutation")
		}
	}
	if _, e := g.observe(0, 4, qs[0], ms[0]); e == nil || g != before {
		t.Fatal("invalid channel")
	}
	if _, e := g.observe(2, 0, qs[0], ms[0]); e == nil || g != before {
		t.Fatal("invalid outcome")
	}
	if _, e := g.observe(0, 0, qs[0], [4]float64{math.Inf(1)}); e == nil || g != before {
		t.Fatal("invalid model")
	}
	p := residualAllocation{}
	for i := 0; i < 200; i++ {
		before := p
		q, m := p.proposal()
		sum := 0.
		for j := range q {
			sum += q[j]
			if q[j] < .1 || q[j] > .7 || math.IsNaN(q[j]) || math.Abs(m[j]) > 1 {
				t.Fatal("bad moments/proposal")
			}
		}
		if math.Abs(sum-1) > 1e-12 || p != before {
			t.Fatal("proposal changes past")
		}
		if e := p.observe(i%4, float64(i%3-1)); e != nil {
			t.Fatal(e)
		}
	}
	pBefore := p
	if e := p.observe(-1, 0); e == nil || p != pBefore {
		t.Fatal("invalid channel history")
	}
}

type augmentedRecord struct {
	Split, Scenario string
	Seed            int64
	First           [4]int
	Counts          [3][4]int
	Clipped         int
	TapeSHA256      string
}

func augmentedRun(s int, split string, seed int64) augmentedRecord {
	r := augmentedRecord{Split: split, Scenario: allocationNames[s], Seed: seed, First: [4]int{-1, -1, -1, -1}}
	rng := rand.New(rand.NewSource(seed))
	oldPolicy := allocationPolicy{}
	newPolicy := residualAllocation{}
	var gates [3]importanceGate
	aug := augmentedGate{}
	hash := sha256.New()
	choose := func(q [4]float64, u float64) int {
		cdf := 0.
		for i, v := range q {
			cdf += v
			if u < cdf {
				return i
			}
		}
		return 3
	}
	for tick := 0; tick < 512; tick++ {
		qOld := oldPolicy.proposal()
		qNew, m := newPolicy.proposal()
		eta, _, err := augmentation(qNew, m)
		if err != nil {
			panic(err)
		}
		if eta < 1 {
			r.Clipped++
		}
		u := rng.Float64()
		channels := [3]int{int(u * 4), choose(qOld, u), choose(qNew, u)}
		var latent [4]float64
		for i := range latent {
			latent[i] = rng.Float64()
		}
		var ds [3]int
		for arm, c := range channels {
			ds[arm] = channelDifference(s, tick, c, latent[c])
			r.Counts[arm][c]++
			prop := .25
			if arm == 1 {
				prop = qOld[c]
			}
			if arm == 2 {
				prop = qNew[c]
			}
			a, e := gates[arm].observe(float64(ds[arm]), prop)
			if e != nil {
				panic(e)
			}
			if a && r.First[arm] < 0 {
				r.First[arm] = tick
			}
		}
		a, e := aug.observe(float64(ds[2]), channels[2], qNew, m)
		if e != nil {
			panic(e)
		}
		if a && r.First[3] < 0 {
			r.First[3] = tick
		}
		var tape [110]byte
		for j, values := range [3][4]float64{qOld, qNew, m} {
			for i, v := range values {
				binary.BigEndian.PutUint64(tape[32*j+8*i:], math.Float64bits(v))
			}
		}
		binary.BigEndian.PutUint64(tape[96:], math.Float64bits(eta))
		for i := range ds {
			tape[104+i] = byte(channels[i])
			tape[107+i] = byte(ds[i] + 1)
		}
		_, _ = hash.Write(tape[:])
		_ = oldPolicy.observe(channels[1], float64(ds[1]))
		_ = newPolicy.observe(channels[2], float64(ds[2]))
	}
	r.TapeSHA256 = hex.EncodeToString(hash.Sum(nil))
	return r
}

func TestAugmentedControls(t *testing.T) {
	for s := range allocationNames {
		seed := int64(2026117400 + s)
		a := augmentedRun(s, "test", seed)
		old := allocationRun(s, "test", seed)
		if a.First[0] != old.First[0] || a.First[1] != old.First[2] || a.Counts[0] != old.Counts[0] || a.Counts[1] != old.Counts[1] {
			t.Fatal("old control drift")
		}
		if a != augmentedRun(s, "test", seed) {
			t.Fatal("nondeterministic run")
		}
		for _, cs := range a.Counts {
			n := 0
			for _, c := range cs {
				n += c
			}
			if n != 512 {
				t.Fatal("query budget")
			}
		}
	}
}

func TestAugmentedV74Experiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_AUGMENTED_ARTIFACT")
	if path == "" {
		t.Skip("opt-in research")
	}
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	hashes := map[string]string{}
	paths, e := filepath.Glob(filepath.Join(root, "internal/observationgate/*.go"))
	if e != nil {
		t.Fatal(e)
	}
	for _, p := range []string{"docs/experiments/mmm-augmented-evidence-v74-protocol.md", "research/augmented-v74-summary.mjs", "research/paired-risk-v73.mjs", "go.mod", "go.sum"} {
		paths = append(paths, filepath.Join(root, p))
	}
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
		hashes[rel] = hex.EncodeToString(h[:])
	}
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if e := enc.Encode(struct {
		Version string
		Hashes  map[string]string
	}{"v74", hashes}); e != nil {
		t.Fatal(e)
	}
	n := 0
	for phase, split := range []string{"design", "confirmation"} {
		for s := range allocationNames {
			for stream := 0; stream < 512; stream++ {
				seed := int64(2026117401+phase)*1000000 + int64(s*1000+stream)
				r := augmentedRun(s, split, seed)
				old := allocationRun(s, split, seed)
				if r.First[0] != old.First[0] || r.First[1] != old.First[2] || r.Counts[0] != old.Counts[0] || r.Counts[1] != old.Counts[1] {
					t.Fatal("control mismatch")
				}
				if e := enc.Encode(r); e != nil {
					t.Fatal(e)
				}
				n++
			}
		}
	}
	if e := f.Sync(); e != nil {
		t.Fatal(e)
	}
	t.Logf("wrote %d paired streams", n)
}

func TestAugmentedV74Replay(t *testing.T) {
	path := os.Getenv("EVENTFRAME_AUGMENTED_REPLAY")
	if path == "" {
		t.Skip("opt-in replay")
	}
	f, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var header struct {
		Version string
		Hashes  map[string]string
	}
	if e := dec.Decode(&header); e != nil {
		t.Fatal(e)
	}
	if header.Version != "v74" {
		t.Fatal("version")
	}
	for p, want := range header.Hashes {
		data, e := os.ReadFile(filepath.Join("../..", p))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(data)
		if hex.EncodeToString(h[:]) != want {
			t.Fatal("source changed", p)
		}
	}
	for phase, split := range []string{"design", "confirmation"} {
		for s := range allocationNames {
			for stream := 0; stream < 512; stream++ {
				var r augmentedRecord
				if e := dec.Decode(&r); e != nil {
					t.Fatal(e)
				}
				want := augmentedRun(s, split, int64(2026117401+phase)*1000000+int64(s*1000+stream))
				if !reflect.DeepEqual(r, want) {
					t.Fatal("replay mismatch")
				}
			}
		}
	}
	var extra any
	if e := dec.Decode(&extra); e != io.EOF {
		t.Fatal("trailing content", e)
	}
}

var augmentedBenchAlert bool

func BenchmarkAugmentedObserve(b *testing.B) {
	p, g := residualAllocation{}, augmentedGate{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if i%512 == 0 {
			p = residualAllocation{}
			g = augmentedGate{}
		}
		q, m := p.proposal()
		c := i % 4
		d := float64(i%3 - 1)
		augmentedBenchAlert, _ = g.observe(d, c, q, m)
		_ = p.observe(c, d)
	}
}
