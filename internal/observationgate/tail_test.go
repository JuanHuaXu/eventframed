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
	"reflect"
	"testing"
)

func TestTailMixtureAlgebra(t *testing.T) {
	g := TailGate{}
	var expected [2]bool
	for i := 0; i < 512; i++ {
		d := float64((i*7)%3 - 1)
		if i >= 128 {
			d = 1
		}
		a, e := g.Observe(d)
		if e != nil {
			t.Fatal(e)
		}
		for arm, weight := range []float64{.9, .5} {
			crossed := false
			for j := range g.base.Starts {
				if i < j*64 {
					continue
				}
				wealth := 0.
				for sign := 0; sign < 2; sign++ {
					s := g.base.Starts[j][sign]
					grid := 0.
					for _, x := range s.Log {
						grid += math.Exp(x) / 5
					}
					wealth += (weight*math.Exp(s.Log[2]) + (1-weight)*grid) / 2
				}
				if wealth >= 800 {
					crossed = true
				}
			}
			expected[arm] = expected[arm] || crossed
			if expected[arm] != a[arm+2] {
				t.Fatal("log mixture differs from direct wealth crossing")
			}
		}
	}
	before := g
	if _, e := g.Observe(math.NaN()); e == nil || !reflect.DeepEqual(g, before) {
		t.Fatal("invalid observation changed gate")
	}
}

func TestTailEvidenceReplay(t *testing.T) {
	root := filepath.Join("..", "..")
	f, e := os.Open(filepath.Join(root, "docs/experiments/mmm-gate-tail-v2.json.gz"))
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	var out TailOutput
	if e = json.NewDecoder(z).Decode(&out); e != nil {
		t.Fatal(e)
	}
	for p, want := range out.Hashes {
		b, e := os.ReadFile(filepath.Join(root, p))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			t.Fatal("source changed", p)
		}
	}
	for _, r := range out.Records {
		g := TailGate{}
		rng := rand.New(rand.NewSource(r.Seed))
		j := -1
		for i, name := range Names {
			if name == r.Scenario {
				j = i
			}
		}
		if j < 0 {
			t.Fatal("unknown scenario")
		}
		previous := 0
		first := [4]int{-1, -1, -1, -1}
		for i, d := range r.Differences {
			if want := difference(j, i, previous, rng); want != d {
				t.Fatal("outcome tape differs", r.Seed)
			}
			previous = d
			a, e := g.Observe(float64(d))
			if e != nil {
				t.Fatal(e)
			}
			for arm, v := range a {
				if v && first[arm] < 0 {
					first[arm] = i
				}
			}
		}
		if first != r.First {
			t.Fatal("first alert mismatch", r.Seed)
		}
	}
	summary, pass, bounds := out.Summary, out.Pass, out.MissBounds
	SummarizeTail(&out)
	if !reflect.DeepEqual(summary, out.Summary) || pass != out.Pass || !reflect.DeepEqual(bounds, out.MissBounds) {
		t.Fatal("summary mismatch")
	}
}
