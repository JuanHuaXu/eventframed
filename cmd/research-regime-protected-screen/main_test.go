package main

import (
	"reflect"
	"testing"
)

func TestScheduleAndCommonJournal(t *testing.T) {
	for _, delayed := range []bool{false, true} {
		f := generate(2026105501, "abrupt", .2, delayed, 2, 8)
		if !reflect.DeepEqual(f, generate(2026105501, "abrupt", .2, delayed, 2, 8)) {
			t.Fatal("non-deterministic fixture")
		}
		first, second := 0, 0
		for at, ps := range f.Packets {
			for _, p := range ps {
				e := f.Events[p.Ordinal]
				if at < p.Ordinal || p.At != at {
					t.Fatal("future data")
				}
				if p.Which == 1 {
					first++
					if p.Value != e.First || at != e.FirstAt {
						t.Fatal("first")
					}
				} else {
					second++
					if !e.HasSecond || p.Value != e.Second || at != e.SecondAt || e.FirstAt >= e.SecondAt {
						t.Fatal("second")
					}
				}
			}
		}
		if first != 16 || second != 8 {
			t.Fatal("schedule counts")
		}
		for _, arm := range []string{"v80_top36", "v81_protected36", "v81_reset9"} {
			x, e := run(f, arm)
			if e != nil {
				t.Fatal(e)
			}
			if len(x.Issued) != 16 || len(x.Operations) != 24 || x.AcceptedFirst+x.RejectedFirst != 16 || x.AcceptedSecond+x.RejectedSecond != 8 {
				t.Fatal("missing attempts")
			}
			for i, o := range x.Operations {
				var flat []packet
				for _, ps := range f.Packets {
					flat = append(flat, ps...)
				}
				if o.packet != flat[i] {
					t.Fatal("different journal")
				}
			}
		}
	}
}

func TestFutureSuffixFork(t *testing.T) {
	f := generate(2026105503, "stationary", .1, true, 2, 8)
	g := generate(2026105503, "stationary", .1, true, 2, 8)
	cut := 8
	for i := cut; i < len(g.Events); i++ {
		g.Events[i].First = 1 - g.Events[i].First
		g.Events[i].Second = 1 - g.Events[i].Second
		g.Events[i].Truth = 1 - g.Events[i].Truth
	}
	for at := range g.Packets {
		for i := range g.Packets[at] {
			if g.Packets[at][i].Ordinal >= cut {
				g.Packets[at][i].Value = 1 - g.Packets[at][i].Value
			}
		}
	}
	for _, arm := range []string{"v80_top36", "v81_protected36", "v81_reset9"} {
		x, e := run(f, arm)
		if e != nil {
			t.Fatal(e)
		}
		y, e := run(g, arm)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(x.Issued[:cut], y.Issued[:cut]) {
			t.Fatal("future suffix changed prefix", arm)
		}
	}
}
