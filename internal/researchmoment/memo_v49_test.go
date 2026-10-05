package researchmoment

import (
	"math"
	"reflect"
	"testing"
)

func TestMemoV49ExactState(t *testing.T) {
	base := make([]float64, 31)
	for i := range base {
		base[i] = .25 + .675*float64(i)/float64(len(base)-1)
	}
	base[7], base[8] = .5, math.Nextafter(.5, 1)
	for _, family := range []string{"narrow", "rich"} {
		for _, prior := range []string{"density", "moment"} {
			for _, shared := range []bool{false, true} {
				for _, strength := range []float64{.125, 32} {
					for _, hazard := range []float64{0, .125} {
						cfg := Config{family, prior, strength, hazard, shared}
						a, err := New(base, 1, len(base)*4, cfg)
						if err != nil {
							t.Fatal(err)
						}
						b, err := NewMemoV49(base, 1, len(base)*4, cfg)
						if err != nil {
							t.Fatal(err)
						}
						if !reflect.DeepEqual(a, b) {
							t.Fatal("constructor state differs", cfg)
						}
						clock := int64(0)
						for round := 0; round < 3; round++ {
							var ta, tb []Ticket
							for m := range base {
								clock++
								x, e := a.Issue(m, clock)
								if e != nil {
									t.Fatal(e)
								}
								y, e := b.Issue(m, clock)
								if e != nil {
									t.Fatal(e)
								}
								if x.Forecast() != y.Forecast() {
									t.Fatal("issued forecast differs")
								}
								ta, tb = append(ta, x), append(tb, y)
							}
							for m := len(base) - 1; m >= 0; m-- {
								clock++
								if (m+round)%7 == 0 {
									if e := a.Cancel(ta[m], clock); e != nil {
										t.Fatal(e)
									}
									if e := b.Cancel(tb[m], clock); e != nil {
										t.Fatal(e)
									}
								} else {
									x, e := a.Resolve(ta[m], (m+round)%3 != 0, clock)
									if e != nil {
										t.Fatal(e)
									}
									y, e := b.Resolve(tb[m], (m+round)%3 != 0, clock)
									if e != nil {
										t.Fatal(e)
									}
									if x != y {
										t.Fatal("receipt differs")
									}
								}
							}
							if !reflect.DeepEqual(a, b) {
								t.Fatal("delayed/canceled state differs", cfg)
							}
						}
						if e := a.BeginEpoch(2, clock+1); e != nil {
							t.Fatal(e)
						}
						if e := b.BeginEpoch(2, clock+1); e != nil {
							t.Fatal(e)
						}
						if !reflect.DeepEqual(a, b) {
							t.Fatal("epoch state differs")
						}
					}
				}
			}
		}
	}
}

func TestMemoV49InvalidAndOwnership(t *testing.T) {
	cfg := Config{Family: "rich", Prior: "moment", Strength: 2, Hazard: 1. / 16, Shared: true}
	check := func(base []float64, epoch uint64, cap int, c Config) {
		t.Helper()
		a, e := New(base, epoch, cap, c)
		b, f := NewMemoV49(base, epoch, cap, c)
		if (e == nil) != (f == nil) || (e != nil && e.Error() != f.Error()) || !reflect.DeepEqual(a, b) {
			t.Fatal("contract differs", e, f)
		}
	}
	for _, base := range [][]float64{nil, {.5}, make([]float64, 201), {.24, .5}, {.5, 1}, {math.NaN(), .5}, {math.Inf(1), .5}} {
		check(base, 1, 1, cfg)
	}
	for _, cap := range []int{0, 129} {
		check([]float64{.25, .925}, 1, cap, cfg)
	}
	check([]float64{.25, .925}, 0, 1, cfg)
	for j := 0; j < 7; j++ {
		c := cfg
		switch j {
		case 0:
			c.Family = "bad"
		case 1:
			c.Prior = "bad"
		case 2:
			c.Strength = 0
		case 3:
			c.Strength = 33
		case 4:
			c.Strength = math.NaN()
		case 5:
			c.Hazard = 1
		case 6:
			c.Hazard = math.Inf(1)
		}
		check([]float64{.25, .925}, 1, 1, c)
	}
	base := make([]float64, 200)
	for i := range base {
		base[i] = .25 + .675*float64(i)/199
	}
	a, e := New(base, 1, 12800, cfg)
	if e != nil {
		t.Fatal(e)
	}
	b, e := NewMemoV49(base, 1, 12800, cfg)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("full-cap eviction differs")
	}
	base[0] = .9
	if a.base[0] != .25 || b.base[0] != .25 {
		t.Fatal("input alias")
	}
	b.prior[0] = 0
	if a.prior[0] == 0 {
		t.Fatal("cross-constructor alias")
	}
}
