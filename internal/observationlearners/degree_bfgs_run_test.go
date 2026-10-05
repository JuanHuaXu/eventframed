package observationlearners

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type degreeBFGSRecord struct {
	Phase, Case, Index, Schedule, Clock, Window int
	LearnNoise                                  bool
	Origins                                     []int
	Stats                                       degreeBFGSStats
	OldStats                                    learnedDegreeStats
	Predictions, OldPredictions, Q              [32]float64
	Y                                           [32]bool
}

func TestDegreeBFGSRun(t *testing.T) {
	src, prior, dst := os.Getenv("EVENTFRAME_BFGS_SOURCE"), os.Getenv("EVENTFRAME_BFGS_PRIOR"), os.Getenv("EVENTFRAME_BFGS_OUTPUT")
	if src == "" {
		t.Skip("explicit research input required")
	}
	if prior == "" || dst == "" {
		t.Fatal("prior/output required")
	}
	f, e := os.Open(src)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	p, e := os.Open(prior)
	if e != nil {
		t.Fatal(e)
	}
	defer p.Close()
	out, e := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer out.Close()
	dec, olddec, enc := json.NewDecoder(f), json.NewDecoder(p), json.NewEncoder(out)
	var header softV120Artifact
	if e = dec.Decode(&header); e != nil || header.Version != "soft-learners-v120" {
		t.Fatal("header", e)
	}
	count, read := 0, 0
	for {
		var s softV120Record
		var old learnedDegreeRecord
		e = dec.Decode(&s)
		if e == io.EOF {
			if e = olddec.Decode(&old); e != io.EOF {
				t.Fatal("extra prior records")
			}
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		if e = olddec.Decode(&old); e != nil {
			t.Fatal(e)
		}
		if s.Phase != old.Phase || s.Case != old.Case || s.Index != old.Index || s.Schedule != old.Schedule {
			t.Fatal("source alignment")
		}
		read++
		if s.Index != 0 {
			continue
		}
		for _, clock := range []int{0, 128, 224} {
			for w, cap := range []int{64, 32} {
				var origins []int
				for i := -16; i < clock; i++ {
					if i < 0 || (!s.Steps[i].Missing && i+s.Steps[i].Delay <= clock) {
						origins = append(origins, i)
					}
				}
				if len(origins) > cap {
					origins = origins[len(origins)-cap:]
				}
				if s.Fits[clock/32].Clock != clock || !reflect.DeepEqual(origins, s.Fits[clock/32].Origins[w]) {
					t.Fatal("as-of origins")
				}
				samples := make([]observation.Sample, len(origins))
				for j, i := range origins {
					if i < 0 {
						samples[j] = s.Initial[i+16]
					} else {
						samples[j] = observation.Sample{Bits: s.Steps[i].X, Outcome: s.Steps[i].Y}
					}
				}
				for mode := 0; mode < 2; mode++ {
					m, st, e := fitDegreeBFGS(samples, mode == 1)
					if e != nil {
						t.Fatalf("record%d: %v", count, e)
					}
					r := degreeBFGSRecord{Phase: s.Phase, Case: s.Case, Index: s.Index, Schedule: s.Schedule, Clock: clock, Window: w, LearnNoise: mode == 1, Origins: origins, Stats: st, OldStats: old.Optimization[clock/32][2*mode+w]}
					if math.Abs(st.Initial-r.OldStats.Initial) > 1e-10 {
						t.Fatal("initial objective mismatch")
					}
					for j := 0; j < 32; j++ {
						r.Predictions[j], e = m.predict(s.Steps[clock+j].X)
						if e != nil {
							t.Fatal(e)
						}
						r.OldPredictions[j] = old.P[clock+j][2+2*mode+w]
						r.Q[j] = s.Steps[clock+j].Q
						r.Y[j] = s.Steps[clock+j].Y
					}
					if e = enc.Encode(r); e != nil {
						t.Fatal(e)
					}
					count++
				}
			}
		}
	}
	if read != 2688 || count != 1008 {
		t.Fatal(read, count)
	}
	if e = out.Sync(); e != nil {
		t.Fatal(e)
	}
	t.Logf("read=%d fits=%d", read, count)
}

func TestDegreeBFGSFreeSystems(t *testing.T) {
	var b degreeMatrix
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			b[i][j] = float64((i + 1) * (j + 1))
			if i == j {
				b[i][j] += 1
			}
		}
	}
	rhs := [5]float64{1, -2, 3, -4, 5}
	for mask := 0; mask < 32; mask++ {
		var free []int
		for i := 0; i < 5; i++ {
			if mask&(1<<i) != 0 {
				free = append(free, i)
			}
		}
		x, e := degreeSmallSolve(b, rhs, free)
		if e != nil {
			t.Fatal(e)
		}
		for _, i := range free {
			v := -rhs[i]
			for j := 0; j < 5; j++ {
				v += b[i][j] * x[j]
			}
			if math.Abs(v) > 1e-11 {
				t.Fatal("free-system residual", v)
			}
		}
	}
}

func BenchmarkDegreeBFGSFit(b *testing.B) {
	spectralInit()
	s := make([]observation.Sample, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 19 % 512), Outcome: i%3 == 0}
	}
	for _, noise := range []bool{false, true} {
		b.Run(fmt.Sprint(noise), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, _, e := fitDegreeBFGS(s, noise); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}

func TestDegreeBFGSNonfinite(t *testing.T) {
	for _, bad := range []learnedDegreeEvaluation{{Value: math.NaN()}, {Value: math.Inf(1)}, {Gradient: [5]float64{math.NaN()}}} {
		_, _, e := degreeBoxMinimize(func([5]float64, bool) (learnedDegreeEvaluation, error) { return bad, nil }, [5]float64{}, true, 64)
		if e == nil {
			t.Fatal("nonfinite objective/gradient accepted")
		}
	}
}
