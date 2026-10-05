package observationgate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

type subsetBreadthCase struct {
	Name, Mode, Target string
	Dependent          bool
}

var subsetBreadthCases = []subsetBreadthCase{
	{"stable_uniform", "stable", "bit", false},
	{"bit_uniform", "member_shift", "bit", false},
	{"xor2_uniform", "member_shift", "xor2", false},
	{"majority3_uniform", "member_shift", "majority3", false},
	{"mux_uniform", "member_shift", "mux", false},
	{"parity4_uniform", "member_shift", "parity4", false},
	{"bit_dependent", "member_shift", "bit", true},
	{"xor2_dependent", "member_shift", "xor2", true},
	{"stable_dependent", "stable", "bit", true},
	{"null_uniform", "null", "bit", false},
}

func breadthInput(x uint16, dependent bool) uint16 {
	if dependent {
		x &^= 1 | 16
		if x&4 != 0 {
			x |= 1
		}
		if x&2 != 0 {
			x |= 16
		}
	}
	return x
}
func breadthRule(x uint16, target string) bool {
	b := func(i uint) bool { return x&(1<<i) != 0 }
	switch target {
	case "bit":
		return b(2)
	case "xor2":
		return b(1) != b(2)
	case "majority3":
		n := 0
		for i := uint(0); i < 3; i++ {
			if b(i) {
				n++
			}
		}
		return n >= 2
	case "mux":
		if b(0) {
			return b(1)
		}
		return b(2)
	case "parity4":
		return ((b(1) != b(2)) != b(3)) != b(4)
	default:
		panic("undeclared breadth rule")
	}
}
func breadthLabel(x uint16, local, null bool, rng *rand.Rand, target string) bool {
	if null {
		return rng.Intn(2) == 1
	}
	y := ((x&64 != 0) != (x&128 != 0)) != (x&256 != 0)
	if local {
		y = breadthRule(x, target)
	}
	if rng.Float64() < .05 {
		y = !y
	}
	return y
}
func breadthBase(c subsetBreadthCase, seed int64) (*observation.Model, error) {
	rng := rand.New(rand.NewSource(seed))
	samples := make([]observation.Sample, 4096)
	for i := range samples {
		x := breadthInput(uint16(rng.Intn(512)), c.Dependent)
		samples[i] = observation.Sample{Bits: x, Outcome: breadthLabel(x, false, c.Mode == "null", rng, c.Target)}
	}
	return observation.Fit(samples)
}

func TestSubsetBreadthContracts(t *testing.T) {
	for x := uint16(0); x < 512; x++ {
		v := breadthInput(x, true)
		if (v&1 != 0) != (v&4 != 0) || (v&16 != 0) != (v&2 != 0) || v&^(uint16(17)) != x&^(uint16(17)) {
			t.Fatal("dependent mapping")
		}
		if breadthInput(x, false) != x {
			t.Fatal("uniform mapping")
		}
		a, b, c := x&1 != 0, x&2 != 0, x&4 != 0
		if breadthRule(x, "bit") != c || breadthRule(x, "xor2") != (b != c) || breadthRule(x, "majority3") != ((a && b) || (a && c) || (b && c)) || breadthRule(x, "mux") != ((a && b) || (!a && c)) {
			t.Fatal("rule truth table")
		}
		n := 0
		for i := uint(1); i <= 4; i++ {
			if x&(1<<i) != 0 {
				n++
			}
		}
		if breadthRule(x, "parity4") != (n%2 == 1) {
			t.Fatal("parity rule")
		}
		// All new rules depend only on the first5 bits, available via views
		// (scope0,depth2) and (scope1,depth1) at total cost5.
		for _, target := range []string{"bit", "xor2", "majority3", "mux", "parity4"} {
			if breadthRule(x, target) != breadthRule(x^480, target) {
				t.Fatal("unaffordable hidden dependency")
			}
		}
	}
	seen := map[int64]bool{}
	for j := range subsetBreadthCases {
		train := int64(2026118300)*1000000 + int64(j)
		if seen[train] {
			t.Fatal("train collision")
		}
		seen[train] = true
		for phase := 0; phase < 2; phase++ {
			for i := 0; i < 64; i++ {
				for role := 0; role < 4; role++ {
					s := observationpreserved.Seed(int64(2026118301+10*j+phase), j, i, role)
					if seen[s] {
						t.Fatal("stream collision")
					}
					seen[s] = true
				}
			}
		}
	}
	cfg := subsetBreadthCases[1]
	base, e := breadthBase(cfg, 2026118300000001)
	if e != nil {
		t.Fatal(e)
	}
	a := newSubsetTrial(base, 1)
	r, e := subsetBreadthRun(base, "unit", 1, 0, 2026118301, a, cfg)
	if e != nil {
		t.Fatal(e)
	}
	b := newSubsetTrial(base, 1)
	rr, e := subsetBreadthRun(base, "unit", 1, 0, 2026118301, b, cfg)
	if e != nil || r != rr || a.metrics != b.metrics || !bytes.Equal(a.hash.Sum(nil), b.hash.Sum(nil)) {
		t.Fatal("breadth replay", e)
	}
}

type subsetBreadthRecord struct {
	Config                subsetBreadthCase
	TrainSeed, StreamBase int64
	Result                subsetRecord
}

func TestSubsetBreadthV83(t *testing.T) {
	output, replay := os.Getenv("EVENTFRAME_SUBSET_BREADTH_OUT"), os.Getenv("EVENTFRAME_SUBSET_BREADTH_REPLAY")
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
	var header struct {
		Version string
		Hashes  map[string]string
	}
	var dec *json.Decoder
	if replay != "" {
		f, e := os.Open(replay)
		if e != nil {
			t.Fatal(e)
		}
		defer f.Close()
		dec = json.NewDecoder(f)
		dec.DisallowUnknownFields()
		if e := dec.Decode(&header); e != nil {
			t.Fatal(e)
		}
		if header.Version != "v83" {
			t.Fatal("version")
		}
		for p, want := range header.Hashes {
			if !filepath.IsLocal(p) {
				t.Fatal("nonlocal source")
			}
			b, e := os.ReadFile(filepath.Join(root, p))
			if e != nil {
				t.Fatal(e)
			}
			h := sha256.Sum256(b)
			if hex.EncodeToString(h[:]) != want {
				t.Fatal("source changed", p)
			}
		}
	} else {
		header.Version = "v83"
		header.Hashes = map[string]string{}
		var paths []string
		for _, dir := range []string{"observationgate", "observation", "observationlearners", "observationpreserved", "observationrescue", "observationexperiment", "bayes", "model"} {
			ps, e := filepath.Glob(filepath.Join(root, "internal", dir, "*.go"))
			if e != nil {
				t.Fatal(e)
			}
			paths = append(paths, ps...)
		}
		for _, p := range []string{"docs/experiments/mmm-subset-breadth-v83-protocol.md", "research/subset-breadth-v83-summary.mjs", "go.mod", "go.sum"} {
			paths = append(paths, filepath.Join(root, p))
		}
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
			header.Hashes[rel] = hex.EncodeToString(h[:])
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if e := enc.Encode(header); e != nil {
		t.Fatal(e)
	}
	for phase, split := range []string{"design", "confirmation"} {
		for j, cfg := range subsetBreadthCases {
			train := int64(2026118300)*1000000 + int64(j)
			seed := int64(2026118301 + 10*j + phase)
			base, e := breadthBase(cfg, train)
			if e != nil {
				t.Fatal(e)
			}
			for i := 0; i < 64; i++ {
				trial := newSubsetTrial(base, 1)
				r, e := subsetBreadthRun(base, split, j, i, seed, trial, cfg)
				if e != nil {
					t.Fatal(e)
				}
				if trial.state.next != 512 || trial.state.pending != nil || trial.fits*3 != r.Fits || trial.metrics.SplitAt != r.Arms[3].SplitAt {
					t.Fatal("candidate accounting")
				}
				got := subsetBreadthRecord{cfg, train, seed, subsetRecord{r, trial.metrics, trial.fits, trial.subsetGuides, hex.EncodeToString(trial.hash.Sum(nil))}}
				if dec != nil {
					var want subsetBreadthRecord
					if e := dec.Decode(&want); e != nil {
						t.Fatal(e)
					}
					if got != want {
						t.Fatal("replay mismatch", split, cfg.Name, i)
					}
				} else {
					if e := enc.Encode(got); e != nil {
						t.Fatal(e)
					}
				}
			}
			t.Log(split, cfg.Name, "complete")
		}
	}
	if dec != nil {
		var extra any
		if e := dec.Decode(&extra); e != io.EOF {
			t.Fatal("extra records", e)
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
}
