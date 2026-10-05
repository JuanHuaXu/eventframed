package observationlearners

import (
	"encoding/json"
	"io"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func TestSparseConvergenceBudget(t *testing.T) {
	s := make([]observation.Sample, 16)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 29), Outcome: i%3 == 0}
	}
	for _, limit := range []int{0, -1, 1025} {
		if _, err := fitSparseVRVMBudget(s, limit); err == nil {
			t.Fatal("invalid budget accepted")
		}
	}
	a, err := fitSparseVRVM(s)
	if err != nil {
		t.Fatal(err)
	}
	b, err := fitSparseVRVMBudget(s, 64)
	if err != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("default changed", err)
	}
	c, err := fitSparseVRVMBudget(s, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.trace) < len(a.trace) || !reflect.DeepEqual(c.trace[:len(a.trace)], a.trace) {
		t.Fatal("extended prefix differs")
	}
	if c.trace[len(c.trace)-1][2] < a.trace[len(a.trace)-1][2] {
		t.Fatal("extended bound degraded")
	}
	t.Logf("iterations=%d stop=%s lastBound=%g", len(c.trace), c.stop, c.trace[len(c.trace)-1][2])

	// Clock zero admits only initial labels: all future outcome/metadata fields
	// must remain evaluator-only even when the optimizer uses a larger budget.
	r := softV120Record{Steps: make([]softV120Step, 256), Fits: make([]softV120Fit, 8)}
	copy(r.Initial[:], s)
	for i := -16; i < 0; i++ {
		r.Fits[0].Origins[0] = append(r.Fits[0].Origins[0], i)
	}
	base, err := runSparsePilotBudget(r, 0, 64, 1024)
	if err != nil {
		t.Fatal(err)
	}
	for i := range r.Steps {
		r.Steps[i].Y = true
		r.Steps[i].Q = math.NaN()
		r.Steps[i].P = [15]float64{math.NaN()}
	}
	r.Case = 999
	poison, err := runSparsePilotBudget(r, 0, 64, 1024)
	if err != nil || base.P != poison.P || base.Mean != poison.Mean || !reflect.DeepEqual(base.Trace, poison.Trace) {
		t.Fatal("extended future/evaluator leak", err)
	}
}

func TestSparseConvergenceCollect(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SPARSE_CONVERGENCE_SOURCE")
	if path == "" {
		t.Skip("explicit research source")
	}
	in, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	out, err := os.OpenFile(os.Getenv("EVENTFRAME_SPARSE_CONVERGENCE_OUTPUT"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	dec, enc := json.NewDecoder(in), json.NewEncoder(out)
	var header softV120Artifact
	if err = dec.Decode(&header); err != nil || header.Version != "soft-learners-v120" {
		t.Fatal("header", err)
	}
	count := 0
	for {
		var r softV120Record
		err = dec.Decode(&r)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if r.Index != 0 {
			continue
		}
		v, err := runSparsePilotBudget(r, 128, 64, 1024)
		if err != nil {
			t.Fatalf("record %d: %v", count, err)
		}
		if err = enc.Encode(v); err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 84 {
		t.Fatal("count", count)
	}
	if err = out.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Logf("fits=%d forecasts=%d", count, count*32)
}

func BenchmarkSparseVRVMConvergence(b *testing.B) {
	s := make([]observation.Sample, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 7), Outcome: i%3 == 0}
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := fitSparseVRVMBudget(s, 1024); err != nil {
			b.Fatal(err)
		}
	}
}
