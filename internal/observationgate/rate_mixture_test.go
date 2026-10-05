package observationgate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestRateMixtureContracts(t *testing.T) {
	var g rateMixture
	if math.Abs(g.logWealth()) > 1e-14 {
		t.Fatal("initial wealth")
	}
	var p residualAllocation
	var direct [3][8][2]float64
	for a := range direct {
		for j := range direct[a] {
			direct[a][j] = [2]float64{1, 1}
		}
	}
	for tick := 0; tick < 512; tick++ {
		long, e := preparePredictiveBet(p)
		if e != nil {
			t.Fatal(e)
		}
		short, e := prepareWindowBet(p, 8)
		if e != nil {
			t.Fatal(e)
		}
		fixed := long
		fixed.rate = [2]float64{.25, .25}
		c := tick % 4
		d := float64(tick%3 - 1)
		if tick >= 128 {
			d = 1
			if tick%5 == 0 {
				d = -1
			}
		}
		if _, e := g.observe(d, c, long, short); e != nil {
			t.Fatal(e)
		}
		for a, s := range [3]predictiveBet{fixed, long, short} {
			z := .25*d/s.q[c] + s.eta*s.c[c]
			for j := 0; j < 8; j++ {
				if tick < j*64 {
					continue
				}
				for k, sign := range []float64{1, -1} {
					direct[a][j][k] *= 1 + s.rate[k]*(sign*z-.15)
				}
			}
		}
		wealth := 0.
		for a, w := range []float64{.5, .25, .25} {
			v := 0.
			for _, pair := range direct[a] {
				v += (pair[0] + pair[1]) / 16
			}
			wealth += w * v
		}
		if math.Abs(math.Log(wealth)-g.logWealth()) > 1e-11 {
			t.Fatal("direct wealth mismatch", tick)
		}
		before := g
		bad := short
		bad.rate[0] = math.NaN()
		if _, e := g.observe(d, c, long, bad); e == nil || g != before {
			t.Fatal("partial state after late rejection")
		}
		bad = short
		bad.m[0] += .001
		if _, e := g.observe(d, c, long, bad); e == nil || g != before {
			t.Fatal("dependency mismatch")
		}
		if _, e := g.observe(2, c, long, short); e == nil || g != before {
			t.Fatal("bad evidence")
		}
		_ = p.observe(c, d)
	}
}

type mixtureRecord struct {
	Split, Scenario     string
	Seed                int64
	First               [5]int
	Counts              [4]int
	LongTape, ShortTape string
}

func mixtureRun(scenario int, split string, seed int64) mixtureRecord {
	r := mixtureRecord{Split: split, Scenario: allocationNames[scenario], Seed: seed, First: [5]int{-1, -1, -1, -1, -1}}
	var p residualAllocation
	var g rateMixture
	rng := rand.New(rand.NewSource(seed))
	for tick := 0; tick < 512; tick++ {
		long, e := preparePredictiveBet(p)
		if e != nil {
			panic(e)
		}
		short, e := prepareWindowBet(p, 8)
		if e != nil {
			panic(e)
		}
		u := rng.Float64()
		channel := 3
		cum := 0.
		for c, q := range long.q {
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
		d := float64(channelDifference(scenario, tick, channel, latent[channel]))
		r.Counts[channel]++
		a, e := g.observe(d, channel, long, short)
		if e != nil {
			panic(e)
		}
		if a && r.First[4] < 0 {
			r.First[4] = tick
		}
		for i, c := range g.components {
			if c.base.alert && r.First[i+1] < 0 {
				r.First[i+1] = tick
			}
		}
		if e := p.observe(channel, d); e != nil {
			panic(e)
		}
	}
	// Controls are computed after the candidate and verify component alarms.
	old := windowRun(scenario, split, seed)
	if old.Counts != r.Counts {
		panic("query drift")
	}
	for i := 1; i < 4; i++ {
		if old.First[i] != r.First[i] {
			panic("component drift")
		}
	}
	r.First[0] = old.First[0]
	r.LongTape = old.LongTape
	r.ShortTape = old.Tape
	return r
}

func TestMixtureReplay(t *testing.T) {
	for s := range allocationNames {
		a := mixtureRun(s, "test", int64(7900+s))
		if a != mixtureRun(s, "test", a.Seed) {
			t.Fatal("replay")
		}
	}
}

func TestMixtureV79Experiment(t *testing.T) {
	output, replay := os.Getenv("EVENTFRAME_RATE_MIXTURE_OUT"), os.Getenv("EVENTFRAME_RATE_MIXTURE_REPLAY")
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
	for _, p := range []string{"docs/experiments/mmm-rate-mixture-v79-protocol.md", "research/rate-mixture-v79-summary.mjs", "research/paired-risk-v73.mjs", "go.mod", "go.sum"} {
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
	}{"v79", hashes}); e != nil {
		t.Fatal(e)
	}
	for phase, split := range []string{"design", "confirmation"} {
		for s := range allocationNames {
			for i := 0; i < 512; i++ {
				r := mixtureRun(s, split, int64(2026117901+phase)*1000000+int64(s*1000+i))
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
			t.Fatal("replay/source mismatch")
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

func BenchmarkRateMixture(b *testing.B) {
	for _, mode := range []string{"balanced", "positive"} {
		b.Run(mode, func(b *testing.B) {
			var p residualAllocation
			var g rateMixture
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if i%512 == 0 {
					p = residualAllocation{}
					g = rateMixture{}
				}
				long, e := preparePredictiveBet(p)
				if e != nil {
					b.Fatal(e)
				}
				short, e := prepareWindowBet(p, 8)
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
				predictiveBenchAlert, e = g.observe(d, i%4, long, short)
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
