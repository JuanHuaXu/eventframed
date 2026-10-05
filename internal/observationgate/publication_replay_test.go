package observationgate

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestPublicationReplay(t *testing.T) {
	path := os.Getenv("EVENTFRAME_PUBLICATION_REPLAY")
	if path == "" {
		t.Skip("opt-in complete forecast replay")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var header json.RawMessage
	if err := dec.Decode(&header); err != nil {
		t.Fatal(err)
	}
	n := 0
	for {
		var r publicationRecord
		if err := dec.Decode(&r); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		ci := -1
		for i, cfg := range forestDelayCases() {
			if cfg.Name == r.Case {
				ci = i
			}
		}
		if ci < 0 {
			t.Fatal("unknown replay case")
		}
		cfg := forestDelayCases()[ci]
		base, err := forestDependenceBase(cfg, r.TrainSeed)
		if err != nil {
			t.Fatal(err)
		}
		for _, delayed := range []bool{false, true} {
			got, err := publicationRun(base, cfg, ci, r.Index, r.Seed, delayed)
			want := r.Immediate
			if delayed {
				want = r.Delayed
			}
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("replay %s %s %d delayed=%v: %v", r.Phase, r.Case, r.Index, delayed, err)
			}
		}
		n++
	}
	if n != 192 {
		t.Fatal("incomplete replay", n)
	}
}

func BenchmarkPublicationFixture(b *testing.B) {
	cfg := forestDelayCases()[1]
	base, err := forestDependenceBase(cfg, 2026092191)
	if err != nil {
		b.Fatal(err)
	}
	for _, delayed := range []bool{false, true} {
		name := "immediate"
		if delayed {
			name = "delayed"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				if _, err := publicationRun(base, cfg, 1, 0, 2026092192, delayed); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
