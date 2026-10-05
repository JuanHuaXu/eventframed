package researchregimelogsummary

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	ref "github.com/JuanHuaXu/eventframed/internal/researchregimeledger"
)

func reference(t testing.TB, m *Ledger, rows []Row, support [][]Key) ref.Result {
	t.Helper()
	rr := make([]ref.Row, len(rows))
	for i, r := range rows {
		rr[i] = ref.Row{Member: r.Member, First: r.First, Second: r.Second}
	}
	ss := make([][]ref.Key, len(support))
	for i, ks := range support {
		ss[i] = make([]ref.Key, len(ks))
		for j, k := range ks {
			ss[i][j] = ref.Key{Class: k.Class, Start: k.Start}
		}
	}
	r, e := ref.Conditional(m.base, ref.Config{Reset: m.cfg.Reset, Hazard: m.cfg.Hazard, Cap: m.cfg.Cap}, rr, ss)
	if e != nil {
		t.Fatal("reference condition rejected", e)
	}
	return r
}

func referenceNear(t *testing.T, a, b float64) {
	t.Helper()
	near(t, a, b)
	report.IncrementalChecks++
	report.MaxReference = math.Max(report.MaxReference, math.Abs(a-b))
}

func checkCold(t *testing.T, m *Ledger) {
	t.Helper()
	r := reference(t, m, m.rows, m.state.Support)
	referenceNear(t, m.state.LogEvidence, r.LogEvidence)
	referenceNear(t, m.state.Discard, r.Discard)
	report.OldEnvelopeComparisons++
	report.MaxOldEnvelopeDifference = math.Max(report.MaxOldEnvelopeDifference, math.Abs(m.state.Envelope-r.Envelope))
	// V81 invents tiny removed mass even when all paths are retained, then
	// amplifies it across unavailable factors. Its scalar bound is not an oracle.
	cold, e := Conditional(m.base, m.cfg, m.rows, m.state.Support)
	if e != nil {
		t.Fatal(e)
	}
	referenceNear(t, m.state.Envelope, cold.Envelope)
	if m.state.Components != r.Components || m.state.MaxComponents != r.MaxComponents {
		t.Fatal("component-count mismatch")
	}
	for i, p := range r.Forecast {
		referenceNear(t, m.state.Forecast[i], p)
	}
	if len(m.cache) > CacheCount {
		t.Fatal("prefix cache grew")
	}
	previous := 0
	for _, p := range m.cache {
		if p.count <= previous || p.count%CacheStride != 0 || p.count > len(m.rows) || p.state.Support != nil || p.state.Forecast != nil {
			t.Fatal("prefix shape or ordering")
		}
		previous = p.count
		r := reference(t, m, m.rows[:p.count], m.state.Support[:p.count])
		referenceNear(t, p.state.LogEvidence, r.LogEvidence)
		referenceNear(t, p.state.Discard, r.Discard)
		report.OldEnvelopeComparisons++
		report.MaxOldEnvelopeDifference = math.Max(report.MaxOldEnvelopeDifference, math.Abs(p.state.Envelope-r.Envelope))
		cold, e := Conditional(m.base, m.cfg, m.rows[:p.count], m.state.Support[:p.count])
		if e != nil {
			t.Fatal(e)
		}
		referenceNear(t, p.state.Envelope, cold.Envelope)
		if len(m.base) <= 4 {
			u, e := p.state.Joint(len(m.base))
			if e != nil {
				t.Fatal(e)
			}
			v, e := r.Joint(len(m.base))
			if e != nil {
				t.Fatal(e)
			}
			for i := range u {
				referenceNear(t, u[i], v[i])
			}
		}
		report.CacheChecks++
	}
}

func cloneCache(m *Ledger) []prefix {
	out := append([]prefix(nil), m.cache...)
	for i := range out {
		out[i].state.parts = append([]component(nil), out[i].state.parts...)
	}
	return out
}

func TestCacheBoundariesLateEvidenceAndRefresh(t *testing.T) {
	for _, cfg := range []Config{{1. / 16, .25, 9}, {.5, 0, 18}, {1, .25, 36}, {0, 1, 9}} {
		m, e := New([]float64{.3, .8}, cfg)
		if e != nil {
			t.Fatal(e)
		}
		for i := 0; i < 320; i++ {
			if _, e = m.Issue(i%2, int64(i)); e != nil {
				t.Fatal(e)
			}
			if i == 62 || i == 63 || i == 64 || i == 126 || i == 127 || i == 128 || i == 255 || i == 319 {
				checkCold(t, m)
			}
		}
		for _, target := range []int{255, 64, 63, 128, 319, 65, 0, 256} {
			before, cache := m.Snapshot(), cloneCache(m)
			q, e := m.Pending(m.token, target, 1)
			if e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(before, m.Snapshot()) || !reflect.DeepEqual(cache, m.cache) {
				t.Fatal("query mutated cached evidence")
			}
			for value := 0; value < 2; value++ {
				r := reference(t, m, m.evidenceRows(target, 1, value), m.state.Support)
				referenceNear(t, q.Branches[value].Probability, math.Exp(r.LogEvidence-m.state.LogEvidence))
				for i, p := range r.Forecast {
					referenceNear(t, q.Branches[value].NextClean[i], p)
				}
			}
			if target == 0 {
				report.OldReplay = true
			}
			if e = m.Reveal(m.token, target, 1, target%2, 320); e != nil {
				t.Fatal(e)
			}
			checkCold(t, m)
			frozen := cloneCache(m)
			if e = m.Reveal(m.token, target, 1, 0, 320); e == nil || !reflect.DeepEqual(frozen, m.cache) {
				t.Fatal("failed update changed cache")
			}
		}
		if e = m.Refresh(m.token, 321); e != nil {
			t.Fatal(e)
		}
		checkCold(t, m)
	}
}

func TestUncappedCachedPrefixAndAtomicHistoryLimit(t *testing.T) {
	m, e := New([]float64{.3, .8}, Config{1. / 16, .25, 0})
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 64; i++ {
		if _, e = m.Issue(i%2, int64(i)); e != nil {
			t.Fatal(e)
		}
	}
	checkCold(t, m)
	before, cache := m.Snapshot(), cloneCache(m)
	if _, e = m.Issue(0, 64); e == nil || !reflect.DeepEqual(before, m.Snapshot()) || !reflect.DeepEqual(cache, m.cache) {
		t.Fatal("uncapped configured limit changed or published")
	}
	if e = m.Reveal(m.token, 0, 1, 1, 64); e != nil {
		t.Fatal(e)
	}
	checkCold(t, m)
}

func TestFullWorkloadReferenceAndResourceShape(t *testing.T) {
	for _, members := range []int{150, 200} {
		m := coldFixture(t, members, members*16)
		checkCold(t, m)
		for _, target := range []int{len(m.rows) - 1, 0} {
			cache := cloneCache(m)
			q, e := m.Pending(m.token, target, 1)
			if e != nil {
				t.Fatal(e)
			}
			for value := 0; value < 2; value++ {
				r := reference(t, m, m.evidenceRows(target, 1, value), m.state.Support)
				referenceNear(t, q.Branches[value].Probability, math.Exp(r.LogEvidence-m.state.LogEvidence))
				for i, p := range r.Forecast {
					referenceNear(t, q.Branches[value].NextClean[i], p)
				}
			}
			if !reflect.DeepEqual(cache, m.cache) {
				t.Fatal("full query changed cache")
			}
		}
		payload := 0
		for _, keys := range m.state.Support {
			payload += len(keys) * int(unsafe.Sizeof(Key{}))
		}
		cachePayload := 0
		for _, p := range m.cache {
			cachePayload += len(p.state.parts) * int(unsafe.Sizeof(component{}))
		}
		t.Logf("M%d T%d support payload %d bytes, cached component payload %d bytes; excludes heap/scratch/headers/publication/RSS", members, len(m.rows), payload, cachePayload)
	}
	report.FullWorkload = true
}

func BenchmarkFullFrontierOldFirst(b *testing.B) {
	for _, members := range []int{150, 200} {
		b.Run(stringName(members), func(b *testing.B) {
			m := coldFixture(b, members, members*16)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				q, e := m.Pending(m.token, 0, 1)
				if e != nil {
					b.Fatal(e)
				}
				benchQuery = q
			}
		})
	}
}

func stringName(members int) string {
	if members == 150 {
		return "M150/T2400"
	}
	return "M200/T3200"
}
