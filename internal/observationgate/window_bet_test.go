package observationgate

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestWindowBetContracts(t *testing.T) {
	var p residualAllocation
	var history [4][]float64
	rng := rand.New(rand.NewSource(7800))
	for tick := 0; tick < 400; tick++ {
		c := rng.Intn(4)
		d := float64(rng.Intn(3) - 1)
		_ = p.observe(c, d)
		history[c] = append(history[c], d)
		before := p
		old, e := preparePredictiveBet(p)
		if e != nil {
			t.Fatal(e)
		}
		full, e := prepareWindowBet(p, 32)
		if e != nil || old != full || p != before {
			t.Fatal("full window parity/immutability", tick)
		}
		short, e := prepareWindowBet(p, 8)
		if e != nil || short.q != old.q || short.m != old.m || short.c != old.c || short.eta != old.eta {
			t.Fatal("query contract changed")
		}
		var probs [4][3]float64
		for i, hs := range history {
			probs[i] = [3]float64{.5, 1, .5}
			n := min(8, len(hs))
			for _, v := range hs[len(hs)-n:] {
				probs[i][int(v)+1]++
			}
			for j := range probs[i] {
				probs[i][j] /= float64(n + 2)
			}
		}
		// Independent append-only histories catch ring-window ordering errors.
		oracle := diagnosticOracle(short, probs)
		for k := 0; k < 2; k++ {
			if math.Abs(short.rate[k]-oracle.rate[k]) > 1e-7 {
				t.Fatal("short-window optimum", tick)
			}
		}
		g := predictiveGate{}
		if _, e := g.observe(d, c, short); e != nil {
			t.Fatal(e)
		}
		if tick%40 == 0 {
			flipped := p
			permuted := p
			for i := range p.values {
				for j, v := range p.values[i] {
					flipped.values[i][j] = -v
				}
				target := (i + 1) % 4
				permuted.values[target] = p.values[i]
				permuted.count[target] = p.count[i]
				permuted.next[target] = p.next[i]
			}
			f, e := prepareWindowBet(flipped, 8)
			if e != nil || math.Abs(f.rate[0]-short.rate[1]) > 1e-7 || math.Abs(f.rate[1]-short.rate[0]) > 1e-7 {
				t.Fatal("sign equivariance")
			}
			v, e := prepareWindowBet(permuted, 8)
			if e != nil || math.Abs(v.rate[0]-short.rate[0]) > 1e-7 || math.Abs(v.rate[1]-short.rate[1]) > 1e-7 {
				t.Fatal("coordinate equivariance")
			}
			for k := 0; k < 2; k++ {
				best := diagnosticGrowth(short, probs, k, short.rate[k])
				for j := 0; j <= 1000; j++ {
					rate := .8 * float64(j) / 1000
					s := short
					s.rate[k] = rate
					g := predictiveGate{}
					if _, e := g.observe(0, 0, s); e == nil && diagnosticGrowth(s, probs, k, rate) > best+1e-8 {
						t.Fatal("grid optimum")
					}
				}
			}
		}
	}
	for _, w := range []int{-1, 0, 33} {
		if _, e := prepareWindowBet(p, w); e == nil {
			t.Fatal("invalid window")
		}
	}
	bad := p
	bad.values[0][(bad.next[0]+31)%32] = .5
	if _, e := prepareWindowBet(bad, 8); e == nil {
		t.Fatal("nonternary observation")
	}
}

type windowRecord struct {
	Split, Scenario string
	Seed            int64
	First           [4]int // uniform / fixed augmented / long rate / short rate
	Counts          [4]int
	Tape, LongTape  string
	Zero            [2]int
	Rates           [2]float64
}

func windowRun(scenario int, split string, seed int64) windowRecord {
	r := windowRecord{Split: split, Scenario: allocationNames[scenario], Seed: seed, First: [4]int{-1, -1, -1, -1}}
	var p residualAllocation
	var g predictiveGate
	rng := rand.New(rand.NewSource(seed))
	hash := sha256.New()
	for tick := 0; tick < 512; tick++ {
		s, e := prepareWindowBet(p, 8)
		if e != nil {
			panic(e)
		}
		u := rng.Float64()
		channel := 3
		cum := 0.
		for c, q := range s.q {
			cum += q
			if u < cum {
				channel = c
				break
			}
		}
		var latent [4]float64
		for c := range latent {
			latent[c] = rng.Float64()
		}
		d := channelDifference(scenario, tick, channel, latent[channel])
		r.Counts[channel]++
		a, e := g.observe(float64(d), channel, s)
		if e != nil {
			panic(e)
		}
		if a && r.First[3] < 0 {
			r.First[3] = tick
		}
		var tape [90]byte
		for j, vs := range [2][4]float64{s.q, s.m} {
			for c, v := range vs {
				binary.BigEndian.PutUint64(tape[j*32+c*8:], math.Float64bits(v))
			}
		}
		binary.BigEndian.PutUint64(tape[64:], math.Float64bits(s.eta))
		for k, v := range s.rate {
			binary.BigEndian.PutUint64(tape[72+k*8:], math.Float64bits(v))
			r.Rates[k] += v / 512
			if v == 0 {
				r.Zero[k]++
			}
		}
		tape[88], tape[89] = byte(channel), byte(d+1)
		_, _ = hash.Write(tape[:])
		if e := p.observe(channel, float64(d)); e != nil {
			panic(e)
		}
	}
	r.Tape = hex.EncodeToString(hash.Sum(nil))
	old := predictiveRun(scenario, split, seed)
	if old.Counts[2] != r.Counts {
		panic("query count drift")
	}
	r.First[0], r.First[1], r.First[2] = old.First[0], old.First[2], old.First[3]
	r.LongTape = old.RateTapeSHA256
	return r
}

func TestWindowRunReplay(t *testing.T) {
	for s := range allocationNames {
		a := windowRun(s, "test", int64(7800+s))
		if a != windowRun(s, "test", a.Seed) {
			t.Fatal("replay drift")
		}
	}
}

func TestWindowV78Experiment(t *testing.T) {
	output, replay := os.Getenv("EVENTFRAME_SHORT_RATE_OUT"), os.Getenv("EVENTFRAME_SHORT_RATE_REPLAY")
	if output == "" && replay == "" {
		t.Skip("opt-in")
	}
	if output != "" && replay != "" {
		t.Fatal("choose write or replay")
	}
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	paths, e := filepath.Glob(filepath.Join(root, "internal/observationgate/*.go"))
	if e != nil {
		t.Fatal(e)
	}
	for _, p := range []string{"docs/experiments/mmm-short-rate-v78-protocol.md", "research/short-rate-v78-summary.mjs", "research/paired-risk-v73.mjs", "go.mod", "go.sum"} {
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
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if e := enc.Encode(struct {
		Version string
		Hashes  map[string]string
	}{"v78", hashes}); e != nil {
		t.Fatal(e)
	}
	for phase, split := range []string{"design", "confirmation"} {
		for s := range allocationNames {
			for i := 0; i < 512; i++ {
				r := windowRun(s, split, int64(2026117801+phase)*1000000+int64(s*1000+i))
				if e := enc.Encode(r); e != nil {
					t.Fatal(e)
				}
			}
		}
	}
	if replay != "" {
		want, e := os.ReadFile(replay)
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(want, buf.Bytes()) {
			t.Fatal("artifact replay/source mismatch")
		}
		return
	}
	f, e := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if _, e := f.Write(buf.Bytes()); e != nil {
		t.Fatal(e)
	}
	if e := f.Sync(); e != nil {
		t.Fatal(e)
	}
	t.Log("wrote10240 fresh paired streams")
}

func BenchmarkWindowObserve(b *testing.B) {
	for _, mode := range []string{"balanced", "positive"} {
		b.Run(mode, func(b *testing.B) {
			var p residualAllocation
			var g predictiveGate
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if i%512 == 0 {
					p = residualAllocation{}
					g = predictiveGate{}
				}
				s, e := prepareWindowBet(p, 8)
				if e != nil {
					b.Fatal(e)
				}
				d := float64(i%3 - 1)
				if mode == "positive" {
					d = 1
					if i%5 == 0 {
						d = -1
					}
				}
				predictiveBenchAlert, e = g.observe(d, i%4, s)
				if e != nil {
					b.Fatal(e)
				}
				if e := p.observe(i%4, d); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}
