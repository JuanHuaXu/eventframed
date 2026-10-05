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
	"testing"
)

func TestPredictiveBetContracts(t *testing.T) {
	p := residualAllocation{}
	s, e := preparePredictiveBet(p)
	if e != nil || s.rate != [2]float64{} {
		t.Fatal("prior should not bet", e, s.rate)
	}
	for i := 0; i < 128; i++ {
		_ = p.observe(i%4, 1)
	}
	s, e = preparePredictiveBet(p)
	if e != nil || s.rate[0] <= .25 || s.rate[1] != 0 {
		t.Fatal("positive evidence did not change bet", e, s.rate)
	}
	old := s
	_ = p.observe(0, -1)
	if s != old {
		t.Fatal("snapshot changed after feedback")
	}
	g := predictiveGate{}
	if _, e := g.observe(1, 0, s); e != nil {
		t.Fatal(e)
	}
	before := g
	for _, d := range []float64{.5, math.NaN(), math.Inf(1)} {
		if _, e := g.observe(d, 0, s); e == nil || g != before {
			t.Fatal("invalid outcome state")
		}
	}
	for _, bad := range []predictiveBet{{}, func() predictiveBet { b := s; b.rate[0] = math.NaN(); return b }(), func() predictiveBet { b := s; b.q[0] = math.NaN(); return b }(), func() predictiveBet { b := s; b.c[0]++; return b }()} {
		if _, e := g.observe(0, 0, bad); e == nil || g != before {
			t.Fatal("invalid snapshot state")
		}
	}
	// Check rates against their own predicted growth without using simulator laws.
	for seed := int64(0); seed < 30; seed++ {
		p = residualAllocation{}
		rng := rand.New(rand.NewSource(seed))
		for i := 0; i < 128; i++ {
			_ = p.observe(rng.Intn(4), float64(rng.Intn(3)-1))
		}
		s, e = preparePredictiveBet(p)
		if e != nil {
			t.Fatal(e)
		}
		for k, sign := range []float64{1, -1} {
			lower := math.Inf(1)
			var xs, probs [12]float64
			for c := 0; c < 4; c++ {
				counts := [3]float64{.5, 1, .5}
				for j := 0; j < p.count[c]; j++ {
					counts[int(p.values[c][j])+1]++
				}
				for d := -1; d <= 1; d++ {
					i := c*3 + d + 1
					xs[i] = sign*(.25*float64(d)/s.q[c]+s.eta*s.c[c]) - .15
					probs[i] = s.q[c] * counts[d+1] / float64(p.count[c]+2)
					lower = math.Min(lower, xs[i])
				}
			}
			cap := .8
			if lower < 0 {
				cap = math.Min(cap, .92/(-lower))
			}
			growth := func(rate float64) float64 {
				v := 0.
				for i, x := range xs {
					v += probs[i] * math.Log1p(rate*x)
				}
				return v
			}
			for j := 0; j <= 1000; j++ {
				if growth(s.rate[k])+1e-12 < growth(cap*float64(j)/1000) {
					t.Fatal("optimizer below grid")
				}
			}
			for _, x := range xs {
				if 1+s.rate[k]*x < .08-1e-12 {
					t.Fatal("factor bound")
				}
			}
		}
	}
}

type predictiveRecord struct {
	augmentedRecord
	RateTapeSHA256 string
	ZeroRates      [2]int
	MeanRates      [2]float64
}

func predictiveRun(scenario int, split string, seed int64) predictiveRecord {
	p := residualAllocation{}
	g := predictiveGate{}
	rng := rand.New(rand.NewSource(seed))
	first := -1
	var counts [4]int
	var zero [2]int
	var means [2]float64
	hash := sha256.New()
	for tick := 0; tick < 512; tick++ {
		s, e := preparePredictiveBet(p)
		if e != nil {
			panic(e)
		}
		u := rng.Float64()
		channel := 3
		cum := 0.
		for i, q := range s.q {
			cum += q
			if u < cum {
				channel = i
				break
			}
		}
		var latent [4]float64
		for i := range latent {
			latent[i] = rng.Float64()
		}
		d := channelDifference(scenario, tick, channel, latent[channel])
		counts[channel]++
		a, e := g.observe(float64(d), channel, s)
		if e != nil {
			panic(e)
		}
		if a && first < 0 {
			first = tick
		}
		var tape [90]byte
		for j, vs := range [2][4]float64{s.q, s.m} {
			for i, v := range vs {
				binary.BigEndian.PutUint64(tape[j*32+i*8:], math.Float64bits(v))
			}
		}
		binary.BigEndian.PutUint64(tape[64:], math.Float64bits(s.eta))
		for k, rate := range s.rate {
			binary.BigEndian.PutUint64(tape[72+k*8:], math.Float64bits(rate))
			if rate == 0 {
				zero[k]++
			}
			means[k] += rate / 512
		}
		tape[88], tape[89] = byte(channel), byte(d+1)
		_, _ = hash.Write(tape[:])
		_ = p.observe(channel, float64(d))
	}
	// Controls are run afterward and never enter the candidate's policy state.
	ref := augmentedRun(scenario, split, seed)
	if counts != ref.Counts[2] {
		panic("query policy drift")
	}
	ref.First = [4]int{ref.First[0], ref.First[1], ref.First[3], first}
	return predictiveRecord{ref, hex.EncodeToString(hash.Sum(nil)), zero, means}
}

func TestPredictiveRunReplay(t *testing.T) {
	for s := range allocationNames {
		a := predictiveRun(s, "test", int64(2026117600+s))
		if a != predictiveRun(s, "test", int64(2026117600+s)) {
			t.Fatal("replay mismatch")
		}
	}
}

func TestPredictiveV76Experiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_PREDICTIVE_ARTIFACT")
	if path == "" {
		t.Skip("opt-in")
	}
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	paths, e := filepath.Glob(filepath.Join(root, "internal/observationgate/*.go"))
	if e != nil {
		t.Fatal(e)
	}
	for _, p := range []string{"docs/experiments/mmm-predictive-bet-v76-protocol.md", "research/predictive-v76-summary.mjs", "research/paired-risk-v73.mjs", "go.mod", "go.sum"} {
		paths = append(paths, filepath.Join(root, p))
	}
	hashes := map[string]string{}
	for _, p := range paths {
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
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
	}{"v76", hashes}); e != nil {
		t.Fatal(e)
	}
	for phase, split := range []string{"design", "confirmation"} {
		for s := range allocationNames {
			for i := 0; i < 512; i++ {
				r := predictiveRun(s, split, int64(2026117601+phase)*1000000+int64(s*1000+i))
				if e := enc.Encode(r); e != nil {
					t.Fatal(e)
				}
			}
		}
	}
	if e := f.Sync(); e != nil {
		t.Fatal(e)
	}
	t.Log("wrote10240 paired streams")
}

func TestPredictiveV76Replay(t *testing.T) {
	path := os.Getenv("EVENTFRAME_PREDICTIVE_REPLAY")
	if path == "" {
		t.Skip("opt-in")
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
	if header.Version != "v76" {
		t.Fatal("version")
	}
	for p, want := range header.Hashes {
		b, e := os.ReadFile(filepath.Join("../..", p))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			t.Fatal("hash", p)
		}
	}
	for phase, split := range []string{"design", "confirmation"} {
		for s := range allocationNames {
			for i := 0; i < 512; i++ {
				var got predictiveRecord
				if e := dec.Decode(&got); e != nil {
					t.Fatal(e)
				}
				want := predictiveRun(s, split, int64(2026117601+phase)*1000000+int64(s*1000+i))
				if got != want {
					t.Fatal("replay mismatch")
				}
			}
		}
	}
	var extra any
	if e := dec.Decode(&extra); e != io.EOF {
		t.Fatal("trailing data", e)
	}
}

var predictiveBenchAlert bool

func BenchmarkPredictiveObserve(b *testing.B) {
	p, g := residualAllocation{}, predictiveGate{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if i%512 == 0 {
			p = residualAllocation{}
			g = predictiveGate{}
		}
		s, e := preparePredictiveBet(p)
		if e != nil {
			b.Fatal(e)
		}
		d := float64(i%3 - 1)
		predictiveBenchAlert, _ = g.observe(d, i%4, s)
		_ = p.observe(i%4, d)
	}
}
