package observationgate

import (
	"bytes"
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

// Simulator-only knowledge must never enter preparePredictiveBet or serving.
func diagnosticLaw(s, tick int) (p [4][3]float64) {
	for c := range p {
		for j := 0; j < 60; j++ {
			p[c][channelDifference(s, tick, c, (float64(j)+.5)/60)+1]++
		}
		for j := range p[c] {
			p[c][j] /= 60
		}
	}
	return
}

func diagnosticGrowth(s predictiveBet, p [4][3]float64, k int, rate float64) float64 {
	sign := 1.
	if k == 1 {
		sign = -1
	}
	v := 0.
	for c := range p {
		for j, probability := range p[c] {
			x := sign*(.25*float64(j-1)/s.q[c]+s.eta*s.c[c]) - .15
			v += s.q[c] * probability * math.Log1p(rate*x)
		}
	}
	return v
}

func diagnosticOracle(s predictiveBet, p [4][3]float64) predictiveBet {
	for k, sign := range []float64{1, -1} {
		cap := .8
		for c := range p {
			for _, d := range []float64{-1, 1} {
				x := sign*(.25*d/s.q[c]+s.eta*s.c[c]) - .15
				if x < 0 {
					cap = math.Min(cap, .92/-x)
				}
			}
		}
		derivative := func(rate float64) float64 {
			v := 0.
			for c := range p {
				for j, probability := range p[c] {
					x := sign*(.25*float64(j-1)/s.q[c]+s.eta*s.c[c]) - .15
					v += s.q[c] * probability * x / (1 + rate*x)
				}
			}
			return v
		}
		if derivative(0) <= 0 {
			s.rate[k] = 0
			continue
		}
		if derivative(cap) >= 0 {
			s.rate[k] = cap
			continue
		}
		lo, hi := 0., cap
		for j := 0; j < 32; j++ {
			mid := (lo + hi) / 2
			if derivative(mid) > 0 {
				lo = mid
			} else {
				hi = mid
			}
		}
		s.rate[k] = (lo + hi) / 2
	}
	return s
}

type diagnosticWindow struct {
	N            int
	Rate, Growth [3]float64 // sums, fixed / actual / oracle
	Zero         [3]int
}
type diagnosticRecord struct {
	Split, Scenario string
	Seed            int64
	First           [3]int
	Tape            string
	Windows         [6]diagnosticWindow
}

func diagnosticRun(scenario int, split string, seed int64) diagnosticRecord {
	r := diagnosticRecord{Split: split, Scenario: allocationNames[scenario], Seed: seed, First: [3]int{-1, -1, -1}}
	var policy residualAllocation
	var gates [3]predictiveGate
	rng := rand.New(rand.NewSource(seed))
	hash := sha256.New()
	at := allocationChange(scenario)
	signIndex := 0
	if scenario == 7 {
		signIndex = 1
	}
	for tick := 0; tick < 512; tick++ {
		s, e := preparePredictiveBet(policy)
		if e != nil {
			panic(e)
		}
		law := diagnosticLaw(scenario, tick)
		fixed := s
		fixed.rate = [2]float64{.25, .25}
		variants := [3]predictiveBet{fixed, s, diagnosticOracle(s, law)}
		window := -1
		offset := tick - at
		switch {
		case offset >= -64 && offset < 0:
			window = 0
		case offset >= 0 && offset < 32:
			window = 1
		case offset >= 32 && offset < 64:
			window = 2
		case offset >= 64 && offset < 128:
			window = 3
		case offset >= 128 && offset < 256:
			window = 4
		case offset >= 256:
			window = 5
		}
		if window >= 0 {
			w := &r.Windows[window]
			w.N++
			for a, v := range variants {
				w.Rate[a] += v.rate[signIndex]
				w.Growth[a] += diagnosticGrowth(v, law, signIndex, v.rate[signIndex])
				if v.rate[signIndex] == 0 {
					w.Zero[a]++
				}
			}
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
		for a, v := range variants {
			alert, e := gates[a].observe(float64(d), channel, v)
			if e != nil {
				panic(e)
			}
			if alert && r.First[a] < 0 {
				r.First[a] = tick
			}
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
		}
		tape[88], tape[89] = byte(channel), byte(d+1)
		_, _ = hash.Write(tape[:])
		if e := policy.observe(channel, float64(d)); e != nil {
			panic(e)
		}
	}
	r.Tape = hex.EncodeToString(hash.Sum(nil))
	return r
}

func TestRateDiagnosticContracts(t *testing.T) {
	for scenario := range allocationNames {
		for _, tick := range []int{0, 511} {
			p := diagnosticLaw(scenario, tick)
			for c, ps := range p {
				plus, minus := .05, .05
				switch {
				case scenario == 0:
					plus, minus = .5, .5
				case scenario == 2:
					plus, minus = 0, 2./15
					if c == 0 {
						plus, minus = 1, 0
					}
				case scenario == 3:
					plus, minus = 0, 0
					if c == 0 {
						plus, minus = .9, .1
					}
					if c == 1 {
						plus, minus = .1, .9
					}
				case tick >= allocationChange(scenario):
					plus, minus = 0, 0
					switch scenario {
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
				want := [3]float64{minus, 1 - plus - minus, plus}
				for j := range ps {
					if math.Abs(ps[j]-want[j]) > 1e-14 {
						t.Fatal("generator enumeration", scenario, tick, c, ps, want)
					}
				}
			}
			var history residualAllocation
			for i := 0; i < 128; i++ {
				_ = history.observe(i%4, float64(channelDifference(scenario, tick, i%4, (float64(i%60)+.5)/60)))
			}
			s, e := preparePredictiveBet(history)
			if e != nil {
				t.Fatal(e)
			}
			oracle := diagnosticOracle(s, p)
			for k := 0; k < 2; k++ {
				best := diagnosticGrowth(s, p, k, oracle.rate[k])
				for j := 0; j <= 1000; j++ {
					rate := .8 * float64(j) / 1000
					v := s
					v.rate[k] = rate
					g := predictiveGate{}
					if _, e := g.observe(0, 0, v); e == nil && diagnosticGrowth(s, p, k, rate) > best+1e-8 {
						t.Fatal("oracle optimization")
					}
				}
			}
			g := predictiveGate{}
			if _, e := g.observe(0, 0, oracle); e != nil {
				t.Fatal(e)
			}
		}
	}
	a := diagnosticRun(4, "test", 2026117601004000)
	if a != diagnosticRun(4, "test", a.Seed) {
		t.Fatal("replay")
	}
	v := predictiveRun(4, "test", a.Seed)
	if a.Tape != v.RateTapeSHA256 || a.First[0] != v.First[2] || a.First[1] != v.First[3] {
		t.Fatal("control drift")
	}
}

func TestRateDiagnosticV77(t *testing.T) {
	output, replay := os.Getenv("EVENTFRAME_RATE_DIAGNOSTIC_OUT"), os.Getenv("EVENTFRAME_RATE_DIAGNOSTIC_REPLAY")
	if output == "" && replay == "" {
		t.Skip("opt-in post-hoc diagnostic")
	}
	if output != "" && replay != "" {
		t.Fatal("choose write or replay")
	}
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(filepath.Join(root, "docs/experiments/mmm-predictive-bet-v76.jsonl"))
	if e != nil {
		t.Fatal(e)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	var oldHeader struct {
		Version string
		Hashes  map[string]string
	}
	if e := dec.Decode(&oldHeader); e != nil {
		t.Fatal(e)
	}
	if oldHeader.Version != "v76" {
		t.Fatal("source version")
	}
	hashes := oldHeader.Hashes
	for path, want := range hashes {
		b, e := os.ReadFile(filepath.Join(root, path))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			t.Fatal("source changed", path)
		}
	}
	for _, path := range []string{"internal/observationgate/rate_diagnostic_test.go", "docs/experiments/mmm-rate-diagnostic-v77-protocol.md", "docs/experiments/mmm-predictive-bet-v76.jsonl"} {
		b, e := os.ReadFile(filepath.Join(root, path))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		hashes[path] = hex.EncodeToString(h[:])
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if e := enc.Encode(struct {
		Version string
		Hashes  map[string]string
	}{"v77-posthoc", hashes}); e != nil {
		t.Fatal(e)
	}
	for phase, split := range []string{"design", "confirmation"} {
		for scenario := range allocationNames {
			for i := 0; i < 512; i++ {
				var old predictiveRecord
				if e := dec.Decode(&old); e != nil {
					t.Fatal(e)
				}
				seed := int64(2026117601+phase)*1000000 + int64(scenario*1000+i)
				if old.Seed != seed || old.Split != split || old.Scenario != allocationNames[scenario] {
					t.Fatal("source order")
				}
				r := diagnosticRun(scenario, split, seed)
				if r.Tape != old.RateTapeSHA256 || r.First[0] != old.First[2] || r.First[1] != old.First[3] {
					t.Fatal("original path mismatch", seed)
				}
				if e := enc.Encode(r); e != nil {
					t.Fatal(e)
				}
			}
		}
	}
	var extra any
	if e := dec.Decode(&extra); e != io.EOF {
		t.Fatal("trailing source data", e)
	}
	if replay != "" {
		want, e := os.ReadFile(replay)
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(want, buf.Bytes()) {
			t.Fatal("diagnostic replay changed")
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
	t.Log("verified original tapes and wrote 10240 diagnostic streams")
}
