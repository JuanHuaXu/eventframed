package observationgate

import (
	"errors"
	"math/bits"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
)

// The budget is unique observed coordinates, not bytes, backend requests or CPU.
// Audit overlap funds completion; there is no advance against future savings.
func monitorCreditStep(credit, bounded int, audit bool) (full bool, charge, next int, err error) {
	if credit < 0 || bounded < 0 || bounded > 18 {
		return false, 0, credit, errors.New("invalid credit input")
	}
	if audit {
		if credit > int(^uint(0)>>1)-bounded {
			return false, 0, credit, errors.New("credit overflow")
		}
		return true, 18, credit + bounded, nil
	}
	if credit >= 18-bounded {
		return true, 18, credit - (18 - bounded), nil
	}
	return false, bounded, credit, nil
}

type monitorCreditReader struct {
	raw          observation.Reader
	epoch        uint64
	mask, values uint16
	charged      int
}

func newMonitorCreditReader(r observation.Reader) *monitorCreditReader {
	return &monitorCreditReader{raw: r, epoch: r.Epoch()}
}
func (r *monitorCreditReader) Epoch() uint64 { return r.raw.Epoch() }
func (r *monitorCreditReader) Read(v observation.View) (uint16, uint16, error) {
	if v.Scope < 0 || v.Scope > 2 || v.Depth < 0 || v.Depth > 2 || r.raw.Epoch() != r.epoch {
		return 0, 0, errors.New("invalid cached view or epoch")
	}
	m := v.Mask()
	if m&^r.mask == 0 {
		return m, r.values & m, nil
	}
	got, value, err := r.raw.Read(v)
	if err != nil {
		return 0, 0, err
	}
	if got != m || value&^got != 0 || r.raw.Epoch() != r.epoch {
		return 0, 0, errors.New("inconsistent reader result")
	}
	if value&r.mask != r.values&m {
		return 0, 0, errors.New("changed cached coordinate")
	}
	r.charged += bits.OnesCount16(m &^ r.mask)
	r.mask |= m
	r.values = (r.values &^ m) | value
	return m, value, nil
}

func TestMonitorCreditContracts(t *testing.T) {
	for b := 0; b <= 18; b++ {
		for c := 0; c <= 40; c++ {
			for _, audit := range []bool{false, true} {
				full, charge, next, e := monitorCreditStep(c, b, audit)
				if e != nil {
					t.Fatal(e)
				}
				allow := b
				if audit {
					allow += 18
				}
				if next < 0 || next != c+allow-charge || charge < b || charge > 18 || audit && !full || full != (audit || c >= 18-b) {
					t.Fatal("ledger", c, b, audit)
				}
			}
		}
	}
	for _, x := range [][2]int{{-1, 0}, {0, -1}, {0, 19}} {
		if _, _, _, e := monitorCreditStep(x[0], x[1], false); e == nil {
			t.Fatal("invalid accepted")
		}
	}
	if _, _, _, e := monitorCreditStep(int(^uint(0)>>1), 1, true); e == nil {
		t.Fatal("overflow")
	}
	c, old, spent := 0, 0, 0
	for i := 0; i < 10000; i++ {
		b, a := i%19, i%97 == 0
		_, cost, n, e := monitorCreditStep(c, b, a)
		if e != nil {
			t.Fatal(e)
		}
		old += b
		if a {
			old += 18
		}
		spent += cost
		c = n
		if spent > old || c != old-spent {
			t.Fatal("prefix budget")
		}
	}
	for x := uint16(0); x < 512; x++ {
		r := newMonitorCreditReader(observationexperiment.Frames(x, "cache-contract"))
		for depth := 0; depth < 3; depth++ {
			for scope := 0; scope < 3; scope++ {
				v := observation.View{Scope: scope, Depth: depth}
				m, y, e := r.Read(v)
				if e != nil || m != v.Mask() || y != x&m {
					t.Fatal("view", x, e)
				}
			}
		}
		if r.mask != 511 || r.values != x || r.charged != 9 {
			t.Fatal("completion")
		}
		for scope := 0; scope < 3; scope++ {
			if _, _, e := r.Read(observation.View{Scope: scope, Depth: 2}); e != nil {
				t.Fatal(e)
			}
		}
		if r.charged != 9 {
			t.Fatal("duplicate charge")
		}
	}
}

type monitorCreditMock struct {
	epoch uint64
	flip  bool
	calls int
}

func (r *monitorCreditMock) Epoch() uint64 { return r.epoch }
func (r *monitorCreditMock) Read(v observation.View) (uint16, uint16, error) {
	r.calls++
	if r.flip {
		r.epoch++
	}
	return v.Mask(), 0, nil
}
func TestMonitorCreditEpoch(t *testing.T) {
	raw := &monitorCreditMock{epoch: 1}
	r := newMonitorCreditReader(raw)
	if _, _, e := r.Read(observation.View{}); e != nil {
		t.Fatal(e)
	}
	raw.epoch++
	if _, _, e := r.Read(observation.View{}); e == nil {
		t.Fatal("cached stale epoch")
	}
	r = newMonitorCreditReader(raw)
	raw.flip = true
	if _, _, e := r.Read(observation.View{}); e == nil || r.charged != 0 || r.mask != 0 {
		t.Fatal("midread epoch")
	}
}
