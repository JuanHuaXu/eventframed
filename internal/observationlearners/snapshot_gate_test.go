package observationlearners

import (
	"math"
	"reflect"
	"testing"
)

func TestSnapshotGateExplicitStartMixture(t *testing.T) {
	for _, n := range []int{4, 8} {
		var ids [8]uint64
		for i := 0; i < n; i++ {
			ids[i] = uint64(i + 1)
		}
		var g snapshotGate
		if err := g.reset(n/4-1, ids, snapshotGateBoundary); err != nil {
			t.Fatal(err)
		}
		var history [][8]float64
		var ys []bool
		var frozen [8][9]float64
		var rejected [8]bool
		for count := 1; count <= 32; count++ {
			raw := snapshotRaw(ids, uint64(count))
			raw[0] = .001
			y := count%7 != 0
			history = append(history, raw)
			ys = append(ys, y)
			if err := g.observe(g.epoch, ids, raw, y); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < n; i++ {
				var terms [9]float64
				z := 0.
				for j := 0; j <= n; j++ {
					if j == i+1 {
						continue
					}
					mass := .5
					if j > 0 {
						mass = .5 / float64(n-1)
					}
					wealth := float64(32 - count)
					for start := 0; start < count; start++ {
						product := 1.
						for k := start; k < count; k++ {
							p := history[k][i]
							q := .5
							if j > 0 {
								q = history[k][j-1]
							}
							if !ys[k] {
								p = 1 - p
								q = 1 - q
							}
							product *= q / p
						}
						wealth += product
					}
					terms[j] = mass * wealth / 32
					z += terms[j]
				}
				if !rejected[i] && z >= snapshotGateBoundary {
					rejected[i] = true
					for j := range terms {
						frozen[i][j] = math.Log(terms[j] / z)
					}
				}
				if rejected[i] != g.tests[i].rejected {
					t.Fatal("rejection mismatch", n, count, i)
				}
				if rejected[i] {
					for j := range terms {
						a, b := frozen[i][j], g.credits[i][j]
						if a != b && math.Abs(a-b) > 1e-10 {
							t.Fatal("first crossing credit", n, count, i, j, a, b)
						}
					}
				}
			}
		}
		before := g
		if err := g.observe(g.epoch, ids, history[0], true); err == nil || !reflect.DeepEqual(g, before) {
			t.Fatal("33rd observation must fail atomically")
		}
	}
}

func TestSnapshotGateFourRoleCompatibility(t *testing.T) {
	var g snapshotGate
	ids := [8]uint64{1, 2, 3, 4}
	if err := g.reset(0, ids, 6400); err != nil {
		t.Fatal(err)
	}
	var old evidenceRouting
	w := [4]float64{.4, .3, .2, .1}
	wide := [8]float64{.4, .3, .2, .1}
	for i := uint64(0); i < 32; i++ {
		raw := [4]float64{.001, .9, .7, .4}
		full := [8]float64{.001, .9, .7, .4}
		want, err := old.predict(i, raw, w)
		if err != nil {
			t.Fatal(err)
		}
		got, err := g.route(ids, wide)
		if err != nil {
			t.Fatal(err)
		}
		for j := range want.Weights {
			if math.Abs(want.Weights[j]-got[j]) > 1e-12 {
				t.Fatal("archived four-role routing")
			}
		}
		if err := old.observe(i, true); err != nil {
			t.Fatal(err)
		}
		if err := g.observe(0, ids, full, true); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSnapshotGateRetainedRejectionAndCreditIdentity(t *testing.T) {
	var g snapshotGate
	ids := [8]uint64{1, 2, 3, 4}
	if err := g.reset(0, ids, snapshotGateBoundary); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		if err := g.observe(0, ids, [8]float64{.001, .9, .8, .7}, true); err != nil {
			t.Fatal(err)
		}
	}
	if !g.tests[0].rejected {
		t.Fatal("fixture not rejected")
	}
	ids = [8]uint64{1, 2, 3, 4, 5, 6, 7, 8}
	if err := g.reset(1, ids, snapshotGateBoundary); err != nil {
		t.Fatal(err)
	}
	if !g.tests[0].rejected {
		t.Fatal("retained snapshot resurrected")
	}
	w := [8]float64{.95, .01, .01, .01, .005, .005, .005, .005}
	r, err := g.route(ids, w)
	if err != nil || r[1] != 0 {
		t.Fatal("rejected snapshot received mass", err)
	}
	// Isolate a credit aimed at an identity that will retire in slot0.
	g.tests[4].rejected = true
	for j := range g.credits[4] {
		g.credits[4][j] = math.Inf(-1)
	}
	g.credits[4][0] = math.Log(.1)
	g.credits[4][1] = math.Log(.9)
	ids = [8]uint64{9, 10, 11, 12, 5, 6, 7, 8}
	if err := g.reset(2, ids, snapshotGateBoundary); err != nil {
		t.Fatal(err)
	}
	if g.tests[0].rejected || !g.tests[4].rejected || !math.IsInf(g.credits[4][1], -1) {
		t.Fatal("generation credit/rejection alias")
	}
	w = [8]float64{0, 0, 0, 0, 1, 0, 0, 0}
	r, err = g.route(ids, w)
	if err != nil || r[0] != 1 {
		t.Fatal("retired credit leaked to new recipient", r, err)
	}
	before := g
	bad := [8]uint64{9, 10, 11, 12, 13, 14, 15, 16}
	bad[0], bad[1] = bad[1], bad[0]
	if err := g.reset(3, bad, snapshotGateBoundary); err == nil || !reflect.DeepEqual(g, before) {
		t.Fatal("identity relocation bypass")
	}
	if err := g.reset(2, ids, snapshotGateBoundary); err == nil || !reflect.DeepEqual(g, before) {
		t.Fatal("same-window test restart")
	}
	next := [8]uint64{9, 10, 11, 12, 13, 14, 15, 16}
	if err := g.reset(3, next, snapshotGateBoundary/2); err == nil || !reflect.DeepEqual(g, before) {
		t.Fatal("gate budget changed mid-run")
	}
}

func TestSnapshotGateAllRejectedAndAtomic(t *testing.T) {
	var g snapshotGate
	ids := [8]uint64{1, 2, 3, 4, 5, 6, 7, 8}
	if err := g.reset(1, ids, snapshotGateBoundary); err != nil {
		t.Fatal(err)
	}
	raw := [8]float64{.001, .001, .001, .001, .001, .001, .001, .001}
	for i := 0; i < 8; i++ {
		if err := g.observe(1, ids, raw, true); err != nil {
			t.Fatal(err)
		}
	}
	w := [8]float64{.125, .125, .125, .125, .125, .125, .125, .125}
	r, err := g.route(ids, w)
	if err != nil || r[0] != 1 {
		t.Fatal("neutral fallback", r, err)
	}
	before := g
	bad := raw
	bad[7] = math.NaN()
	if err := g.observe(1, ids, bad, true); err == nil || !reflect.DeepEqual(g, before) {
		t.Fatal("invalid last emission partially committed")
	}
	if err := g.observe(0, ids, raw, true); err == nil || !reflect.DeepEqual(g, before) {
		t.Fatal("wrong window admitted")
	}
}
