package observationlearners

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestPartialEvidence(t *testing.T) {
	f, e := os.Open("../../docs/experiments/mmm-partial-v7.json.gz")
	if os.IsNotExist(e) {
		t.Skip("not run")
	}
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	var o PartialOutput
	if e = json.NewDecoder(z).Decode(&o); e != nil {
		t.Fatal(e)
	}
	if len(o.Records) != 480 {
		t.Fatal("incomplete experiment")
	}
	for p, want := range o.Hashes {
		b, e := os.ReadFile("../../" + p)
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			t.Fatalf("source drift %s", p)
		}
	}
	for _, r := range o.Records {
		if len(r.Ticks) != 512 {
			t.Fatal("truncated trace")
		}
		var s Scenario
		for _, v := range Scenarios {
			if v.Name == r.Scenario {
				s = v
			}
		}
		seen := map[int]bool{}
		audits := 0
		for t0, tick := range r.Ticks {
			for _, origin := range tick.Delivered {
				if origin < 0 || origin > t0 || origin+s.Delay != t0 || seen[origin] || r.Ticks[origin].Missing {
					t.Fatal("invalid delivery")
				}
				seen[origin] = true
				if r.Ticks[origin].Audit {
					audits++
				}
			}
			if tick.Audits != audits {
				t.Fatal("audit leakage")
			}
		}
		if r.Available != len(seen) || r.Audits != audits {
			t.Fatal("feedback accounting")
		}
	}
	replay, e := RunPartial()
	if e != nil {
		t.Fatal(e)
	}
	for i := range o.Records {
		o.Records[i].FitNS = 0
		o.Records[i].TreeNS = 0
		replay.Records[i].FitNS = 0
		replay.Records[i].TreeNS = 0
	}
	if !reflect.DeepEqual(o.Records, replay.Records) || !reflect.DeepEqual(o.Comparisons, replay.Comparisons) || !reflect.DeepEqual(o.Verdicts, replay.Verdicts) {
		t.Fatal("replay mismatch")
	}
}
