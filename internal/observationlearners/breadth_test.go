package observationlearners

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"math/rand"
	"os"
	"reflect"
	"testing"
)

func TestBreadthArtifactReplay(t *testing.T) {
	path := os.Getenv("EVENTFRAME_BREADTH_ARTIFACT")
	if path == "" {
		t.Skip("explicit artifact required")
	}
	f, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	decoder := json.NewDecoder(z)
	var header struct{ ExpectedRecords int }
	if e = decoder.Decode(&header); e != nil {
		t.Fatal(e)
	}
	n := 0
	e = RunBreadth(func(r BreadthRecord) error {
		var old BreadthRecord
		if e := decoder.Decode(&old); e != nil {
			return e
		}
		r.FitNS, r.TreeNS, old.FitNS, old.TreeNS = 0, 0, 0, 0
		if !reflect.DeepEqual(r, old) {
			t.Fatalf("replay mismatch %d", n)
		}
		n++
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	if n != header.ExpectedRecords {
		t.Fatal("record count")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		t.Fatal("trailing records")
	}
}

func TestBreadthTruthAndSeeds(t *testing.T) {
	s := Scenario{Name: "shift", Change: 128}
	for family := 0; family < 2; family++ {
		for x := uint16(0); x < 512; x++ {
			for _, tick := range []int{-1, 127, 128, 511} {
				offset := uint(6)
				if tick >= 128 {
					offset = 0
				}
				n := 0
				for b := offset; b < offset+3; b++ {
					if x&(1<<b) != 0 {
						n++
					}
				}
				want := n >= 2
				if family == 1 {
					choose := offset
					if x&(1<<(offset+2)) != 0 {
						choose++
					}
					want = x&(1<<choose) != 0
				}
				if breadthTruth(x, tick, s, family, rand.New(rand.NewSource(1))) != want {
					t.Fatal("truth table")
				}
			}
		}
	}
	seen := map[int64]bool{}
	for _, base := range []int64{2026107201, 2026107202, 2026107203} {
		for family := 0; family < 2; family++ {
			for g := 0; g < 3; g++ {
				for _, j := range []int{0, 2, 6, 7} {
					for fit := 0; fit < 6; fit++ {
						for stream := 0; stream < 2; stream++ {
							for role := 0; role < 5; role++ {
								seed := Seed(base, 30*family+10*g+j, fit, stream, role)
								if seen[seed] {
									t.Fatal("seed collision")
								}
								seen[seed] = true
							}
						}
					}
				}
			}
		}
	}
}
