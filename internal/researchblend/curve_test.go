package researchblend

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func clearCurveTimingsV33(c *curveV33) {
	c.NominationNS, c.ElapsedNS = 0, 0
	for i := range c.Snapshots {
		c.Snapshots[i].NominationNS = 0
		c.Snapshots[i].Costs = costsV33{}
	}
}

func TestNominationMultiplePassesV33(t *testing.T) {
	for _, n := range []int{2, 11, 150, 200} {
		base, coordinate := fixture(n)
		for _, seed := range []int64{37, 911, 2026103399} {
			m, err := newNominationV33(n, seed)
			if err != nil {
				t.Fatal(err)
			}
			ref, _ := New(base, coordinate)
			rng := rand.New(rand.NewSource(seed + 303))
			seen := make([]bool, n)
			for k := 0; k < n; k++ {
				got, e := m.next()
				want, f := ref.Select("stratified_random", rng)
				if e != nil || f != nil || got != want || got.Index < 0 || seen[got.Index] {
					t.Fatalf("nomination mismatch n=%d step=%d: %+v versus %+v", n, k, got, want)
				}
				seen[got.Index] = true
				if e = ref.Observe(got.Index, k%3 == 0); e != nil {
					t.Fatal(e)
				}
			}
			if _, e := m.next(); e == nil || m.count != n || !reflect.DeepEqual(m.seen, seen) {
				t.Fatal("exhaustion not fail-closed")
			}
			if _, e := m.next(); e == nil {
				t.Fatal("repeat exhaustion accepted")
			}
		}
	}
	for _, n := range []int{0, 1, 201} {
		if _, err := newNominationV33(n, 1); err == nil {
			t.Fatal("unsupported frontier accepted")
		}
	}
}

func TestCurveNoFutureAtEveryPrefixV33(t *testing.T) {
	base, coordinate := fixture(150)
	labels := make([]bool, 150)
	for i := range labels {
		labels[i] = i%3 == 0
	}
	full := runCurveV33(base, coordinate, labels, 637, 150)
	clearCurveTimingsV33(&full)
	uninstrumented := runCurveWithCheckpointsV33(base, coordinate, labels, 637, 150, nil)
	if !reflect.DeepEqual(full.Trace, uninstrumented.Trace) || len(uninstrumented.Snapshots) != 0 {
		t.Fatal("snapshot queries changed subsequent learning or forecasts")
	}
	for _, budget := range []int{16, 32, 64, 128} {
		future := append([]bool(nil), labels...)
		seen := make([]bool, 150)
		for _, row := range full.Trace[:budget] {
			seen[row.Index] = true
		}
		for i := range future {
			if !seen[i] {
				future[i] = !future[i]
			}
		}
		changed := runCurveV33(base, coordinate, future, 637, budget)
		clearCurveTimingsV33(&changed)
		if !reflect.DeepEqual(full.Trace[:budget], changed.Trace) {
			t.Fatal("unarrived label affected issued forecasts or nomination")
		}
		prefix := []snapshotV33{}
		for _, snap := range full.Snapshots {
			if snap.Budget <= budget {
				prefix = append(prefix, snap)
			}
		}
		if !reflect.DeepEqual(prefix, changed.Snapshots) {
			t.Fatal("future label affected earlier checkpoint state/law/LOO rows")
		}
	}
	for _, model := range []string{"local", "old", "partition", "blend"} {
		control := runV30(base, coordinate, labels, model, "stratified_random", 637)
		for _, snap := range full.Snapshots {
			if snap.Model == model && snap.Budget == 32 {
				if !reflect.DeepEqual(snap.Forecast, control.Forecast) || snap.Weights != control.Weights {
					t.Fatal("first32 law differs from frozen model runner")
				}
			}
		}
	}
}

func TestCurveFutureScoreDenominatorsV33(t *testing.T) {
	base, coordinate := fixture(150)
	labels := make([]bool, 150)
	rates := make([]float64, 150)
	for i := range labels {
		labels[i] = i%3 == 0
		rates[i] = .2 + .6*float64(i%2)
	}
	w := worldV33{Base: base, Coordinates: coordinate, Rates: rates, Labels: labels, curveV33: runCurveV33(base, coordinate, labels, 311, 150)}
	scoreCurveV33(&w)
	if len(w.Snapshots) != 35 || len(w.Trace) != 150 {
		t.Fatal("missing curve checkpoints")
	}
	for _, s := range w.Snapshots {
		if s.ObservedBrier == nil {
			t.Fatal("observed risk missing")
		}
		if s.Budget == 150 {
			if s.UnobservedBrier != nil || math.Abs(*s.ObservedBrier-s.Brier) > 1e-14 {
				t.Fatal("empty unobserved population treated as data")
			}
		} else if s.UnobservedBrier == nil || math.Abs(float64(s.Budget)/150**s.ObservedBrier+float64(150-s.Budget)/150**s.UnobservedBrier-s.Brier) > 1e-14 {
			t.Fatal("split risks do not reconstruct total future risk")
		}
		if s.Costs.AccountedNS != s.Costs.SetupNS+s.Costs.ProbeNS+s.Costs.UpdateNS+s.Costs.SnapshotNS+s.Costs.EarlierSnapshotsNS {
			t.Fatal("cost accounting differs")
		}
		if s.Model == "baseline" && !reflect.DeepEqual(s.Forecast, base) {
			t.Fatal("static baseline learned from labels")
		}
	}
}
