package observationlearners

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type spikeContinuousRecord struct {
	Phase, Case, Index, Schedule int
	P                            [256]float64
	FitCounts, Capped            [8]int
}

func TestSpikeContinuousCollect(t *testing.T) {
	src, dir, prefix := os.Getenv("EVENTFRAME_SPIKE_CONTINUOUS_SOURCE"), os.Getenv("EVENTFRAME_SPIKE_CONTINUOUS_CACHE"), os.Getenv("EVENTFRAME_SPIKE_CONTINUOUS_PREFIX")
	if src == "" {
		t.Skip("explicit research source")
	}
	if dir == "" || prefix == "" {
		t.Fatal("cache directory and output prefix required")
	}
	open := func(path string) *os.File {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close() })
		return f
	}
	var outputs []*os.File
	create := func(path string) *json.Encoder {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		outputs = append(outputs, f)
		t.Cleanup(func() { f.Close() })
		return json.NewEncoder(f)
	}
	cache := map[int]*json.Decoder{}
	for c, name := range map[int]string{0: "early", 128: "breadth", 224: "late"} {
		cache[c] = json.NewDecoder(open(filepath.Join(dir, "mmm-spike-cadence-"+name+".jsonl")))
	}
	writers := map[int]*json.Encoder{}
	for _, c := range []int{32, 64, 96, 160, 192} {
		writers[c] = create(fmt.Sprintf("%s-clock%d.jsonl", prefix, c))
	}
	compact := create(prefix + ".jsonl")
	dec := json.NewDecoder(open(src))
	var header softV120Artifact
	if err := dec.Decode(&header); err != nil || header.Version != "soft-learners-v120" {
		t.Fatal("header", err)
	}
	count, totalFits, newFits, caps := 0, 0, 0, 0
	for {
		var r softV120Record
		err := dec.Decode(&r)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if r.Index < 0 || r.Index >= 8 {
			continue
		}
		out := spikeContinuousRecord{Phase: r.Phase, Case: r.Case, Index: r.Index, Schedule: r.Schedule}
		var last *spikeCadenceFit
		for c := 0; c < 256; c += 32 {
			var v spikeCadenceRecord
			if cached := cache[c]; cached != nil {
				if err = cached.Decode(&v); err != nil {
					t.Fatal(err)
				}
			} else {
				v, err = runSpikeCadenceAt(r, c)
				if err != nil {
					t.Fatal(err)
				}
				if err = writers[c].Encode(v); err != nil {
					t.Fatal(err)
				}
				newFits += len(v.Fits)
			}
			clock := 128
			if v.Clock != nil {
				clock = *v.Clock
			}
			if clock != c || v.Phase != r.Phase || v.Case != r.Case || v.Index != r.Index || v.Schedule != r.Schedule {
				t.Fatal("cached identity/order mismatch")
			}
			if len(v.Fits) == 0 {
				t.Fatal("empty fit list")
			}
			if last != nil && reflect.DeepEqual(last.Origins, v.Fits[0].Origins) {
				a, b := *last, v.Fits[0]
				a.Clock, b.Clock = 0, 0
				if !reflect.DeepEqual(a, b) {
					t.Fatal("same-window boundary refit changed law")
				}
			}
			last = &v.Fits[len(v.Fits)-1]
			for i := 0; i < 32; i++ {
				s := r.Steps[c+i]
				if v.X[i] != s.X || v.Y[i] != s.Y || v.Q[i] != s.Q || v.Control[i] != s.P {
					t.Fatal("cached source mismatch")
				}
				out.P[c+i] = v.P[i]
			}
			out.FitCounts[c/32] = len(v.Fits)
			for _, f := range v.Fits {
				if f.Stop == "iteration-cap" {
					out.Capped[c/32]++
					caps++
				}
			}
			totalFits += len(v.Fits)
		}
		if err = compact.Encode(out); err != nil {
			t.Fatal(err)
		}
		count++
		if count%84 == 0 {
			t.Logf("records=%d totalFits=%d newFits=%d caps=%d", count, totalFits, newFits, caps)
		}
	}
	if count != 672 {
		t.Fatal("record count", count)
	}
	for _, d := range cache {
		var extra spikeCadenceRecord
		if err := d.Decode(&extra); err != io.EOF {
			t.Fatal("extra cache row", err)
		}
	}
	for _, f := range outputs {
		if err := f.Sync(); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("complete records=%d forecasts=%d totalFits=%d newFits=%d caps=%d", count, count*256, totalFits, newFits, caps)
}
