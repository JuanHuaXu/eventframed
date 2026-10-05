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

func TestAllocationContracts(t *testing.T) {
	p := allocationPolicy{}
	if p.proposal() != [4]float64{.25, .25, .25, .25} {
		t.Fatal("initial proposal")
	}
	for i := 0; i < 128; i++ {
		if e := p.observe(i%4, float64(i%3-1)); e != nil {
			t.Fatal(e)
		}
	}
	before := p
	q := p.proposal()
	sum := 0.
	for _, v := range q {
		sum += v
		if v < .1 || v > .7 {
			t.Fatal("coverage")
		}
	}
	if math.Abs(sum-1) > 1e-14 || p != before {
		t.Fatal("proposal mutated history")
	}
	for _, input := range []struct {
		c int
		d float64
	}{{-1, 0}, {4, 0}, {0, math.NaN()}, {0, 2}} {
		if e := p.observe(input.c, input.d); e == nil || p != before {
			t.Fatal("invalid observation changed state")
		}
	}
	q = [4]float64{.4, .2, .2, .2}
	mu := [4]float64{1, -2. / 15, -2. / 15, -2. / 15}
	weighted, naive := 0., 0.
	for i := range q {
		weighted += q[i] * (.25 / q[i]) * mu[i]
		naive += q[i] * mu[i]
	}
	if math.Abs(weighted-.15) > 1e-14 || naive <= .15 {
		t.Fatal("selection correction")
	}
	for _, sign := range []float64{-1, 1} {
		if 1+.25*(sign*weighted-.15) > 1+1e-14 {
			t.Fatal("invalid null factor")
		}
	}
	if 1+.25*(-2.5-.15) < .3375-1e-14 {
		t.Fatal("nonpositive factor")
	}
	g := importanceGate{}
	old := StartPoolGate{}
	for i := 0; i < 512; i++ {
		d := float64(i%5%3 - 1)
		a, e := g.observe(d, .25)
		b, _ := old.Observe(d)
		if e != nil || a != b[2] {
			t.Fatal("uniform pooled parity")
		}
	}
	gb := g
	for _, bad := range []float64{.09, .71, math.NaN(), math.Inf(1)} {
		if _, e := g.observe(0, bad); e == nil || g != gb {
			t.Fatal("bad propensity")
		}
	}
	if _, e := g.observe(math.NaN(), .25); e == nil || g != gb {
		t.Fatal("bad evidence")
	}
}

var allocationNames = []string{"symmetric", "sparse_null", "heterogeneous_boundary", "cancelling", "homogeneous128", "sparse128", "sparse256", "negative256", "sparse384", "weak256"}

func allocationChange(s int) int {
	if s < 4 {
		return 512
	}
	if s == 4 || s == 5 {
		return 128
	}
	if s == 8 {
		return 384
	}
	return 256
}

func channelDifference(s, t, c int, u float64) int {
	plus, minus := .05, .05
	switch s {
	case 0:
		plus, minus = .5, .5
	case 2:
		if c == 0 {
			plus, minus = 1, 0
		} else {
			plus, minus = 0, 2./15
		}
	case 3:
		plus, minus = 0, 0
		if c == 0 {
			plus, minus = .9, .1
		}
		if c == 1 {
			plus, minus = .1, .9
		}
	default:
		if t >= allocationChange(s) {
			plus, minus = 0, 0
			switch s {
			case 4:
				plus, minus = .5, .1
			case 5, 6, 8:
				if c < 2 {
					plus, minus = .85, .05
				}
			case 7:
				if c < 2 {
					plus, minus = .05, .85
				}
			case 9:
				if c == 0 {
					plus, minus = .8, 0
				}
			}
		}
	}
	if u < plus {
		return 1
	}
	if u < plus+minus {
		return -1
	}
	return 0
}

type allocationRecord struct {
	Split, Scenario string
	Seed            int64
	First           [3]int
	Counts          [2][4]int  // uniform and shared adaptive selection, respectively
	Means           [3]float64 // uniform, unweighted adaptive, corrected adaptive
	TapeSHA256      string
}

func allocationRun(s int, split string, seed int64) allocationRecord {
	r := allocationRecord{Split: split, Scenario: allocationNames[s], Seed: seed, First: [3]int{-1, -1, -1}}
	rng := rand.New(rand.NewSource(seed))
	policy := allocationPolicy{}
	var gates [3]importanceGate
	hash := sha256.New()
	for tick := 0; tick < 512; tick++ {
		q := policy.proposal()
		u := rng.Float64()
		uniform := int(u * 4)
		selected := 3
		cdf := 0.
		for c, v := range q {
			cdf += v
			if u < cdf {
				selected = c
				break
			}
		}
		// Only the simulator owns the counterfactual outcome tape; policy state
		// receives the chosen channel after q and both actions are committed.
		var latent [4]float64
		for c := range latent {
			latent[c] = rng.Float64()
		}
		d0, d1 := channelDifference(s, tick, uniform, latent[uniform]), channelDifference(s, tick, selected, latent[selected])
		r.Counts[0][uniform]++
		r.Counts[1][selected]++
		for arm := range gates {
			d, propensity := float64(d1), .25
			if arm == 0 {
				d = float64(d0)
			}
			if arm == 2 {
				propensity = q[selected]
			}
			a, err := gates[arm].observe(d, propensity)
			if err != nil {
				panic(err)
			}
			if a && r.First[arm] < 0 {
				r.First[arm] = tick
			}
			r.Means[arm] += .25 * d / propensity / 512
		}
		var tape [36]byte
		for c, v := range q {
			binary.BigEndian.PutUint64(tape[c*8:], math.Float64bits(v))
		}
		tape[32], tape[33], tape[34], tape[35] = byte(uniform), byte(selected), byte(d0+1), byte(d1+1)
		_, _ = hash.Write(tape[:])
		_ = policy.observe(selected, float64(d1))
	}
	r.TapeSHA256 = hex.EncodeToString(hash.Sum(nil))
	return r
}

func TestAllocationReplayAndCounts(t *testing.T) {
	for s := range allocationNames {
		a := allocationRun(s, "test", int64(2026117200+s))
		b := allocationRun(s, "test", int64(2026117200+s))
		if !reflect.DeepEqual(a, b) {
			t.Fatal("replay changed")
		}
		for _, counts := range a.Counts {
			n := 0
			for _, v := range counts {
				n += v
			}
			if n != 512 {
				t.Fatal("query count")
			}
		}
	}
}

func TestAllocationV72Experiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ALLOCATION_ARTIFACT")
	if path == "" {
		t.Skip("opt-in experiment")
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
	paths = append(paths, filepath.Join(root, "docs/experiments/mmm-evidence-allocation-v72-protocol.md"), filepath.Join(root, "go.mod"), filepath.Join(root, "go.sum"))
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
	}{"v72", hashes}); e != nil {
		t.Fatal(e)
	}
	n := 0
	for phase, split := range []string{"design", "confirmation"} {
		for s := range allocationNames {
			for stream := 0; stream < 512; stream++ {
				r := allocationRun(s, split, int64(2026117201+phase)*1000000+int64(s*1000+stream))
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

func TestAllocationV72ArtifactReplay(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ALLOCATION_REPLAY")
	if path == "" {
		t.Skip("opt-in full replay")
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
	if header.Version != "v72" {
		t.Fatal("wrong version")
	}
	for p, want := range header.Hashes {
		data, e := os.ReadFile(filepath.Join("../..", p))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(data)
		if hex.EncodeToString(h[:]) != want {
			t.Fatal("source hash", p)
		}
	}
	for phase, split := range []string{"design", "confirmation"} {
		for s := range allocationNames {
			for stream := 0; stream < 512; stream++ {
				var got allocationRecord
				if e := dec.Decode(&got); e != nil {
					t.Fatal(e)
				}
				want := allocationRun(s, split, int64(2026117201+phase)*1000000+int64(s*1000+stream))
				if !reflect.DeepEqual(got, want) {
					t.Fatal("record replay mismatch", s, stream)
				}
			}
		}
	}
	var extra any
	if e := dec.Decode(&extra); e != io.EOF {
		t.Fatal("unexpected trailing artifact content", e)
	}
}

var allocationBenchQ [4]float64

func BenchmarkAllocationObserve(b *testing.B) {
	p, g := allocationPolicy{}, importanceGate{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if i%512 == 0 {
			p = allocationPolicy{}
			g = importanceGate{}
		}
		q := p.proposal()
		c := i % 4
		d := float64(i%3 - 1)
		_, _ = g.observe(d, q[c])
		_ = p.observe(c, d)
		allocationBenchQ = q
	}
}
