package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

func checkJournal(t *testing.T, g forecastJournal) {
	t.Helper()
	n := 0
	for i, e := range g.entries {
		if e.active {
			n++
			if e.origin < 0 || e.origin >= g.issued || e.origin%64 != i {
				t.Fatal("entry identity")
			}
		}
	}
	if n != g.stats.Pending || n > 64 || g.issued != g.stats.Applied+g.stats.Stale+g.stats.Censored+n {
		t.Fatal("journal accounting", g.issued, g.stats)
	}
}

func journalModels(t *testing.T) (*observation.Model, *subsetTrial, observationpreserved.Models) {
	t.Helper()
	base, e := observationpreserved.Base(false)
	if e != nil {
		t.Fatal(e)
	}
	samples := make([]observation.Sample, 64)
	for i := range samples {
		x := uint16(i * 7 % 512)
		samples[i] = observation.Sample{Bits: x, Outcome: x&4 != 0}
	}
	trial := newSubsetTrial(base, 1)
	if e := trial.fit(samples); e != nil {
		t.Fatal(e)
	}
	count, e := observation.Fit(samples)
	if e != nil {
		t.Fatal(e)
	}
	return base, trial, observationpreserved.Models{Short: count, Local: count, Pooled: count, Version: 1}
}

func TestJournalImmediateParity(t *testing.T) {
	base, trial, models := journalModels(t)
	for _, enabled := range []bool{false, true} {
		state := subsetState{base: base, enabled: enabled}
		g := forecastJournal{base: base, enabled: enabled}
		m := models
		if e := g.publish(m, trial.model); e != nil {
			t.Fatal(e)
		}
		for i := 0; i < 1024; i++ {
			if i > 0 && i%64 == 0 {
				m.Version++
				if e := g.publish(m, trial.model); e != nil {
					t.Fatal(e)
				}
			}
			rd := observationexperiment.Frames(uint16(i*13%512), "journal-parity")
			want, e := state.predict(rd, m, trial.model, i, 0)
			if e != nil {
				t.Fatal(e)
			}
			got, e := g.predict(rd, i, 0)
			if e != nil || got != want {
				t.Fatal("immediate forecast parity", enabled, i, e)
			}
			if g.mix != state.mix || g.inner != state.inner {
				t.Fatal("issuance learned from unobserved label")
			}
			y, auth := i%3 == 0, i == 100
			if e := state.observe(i, y, auth); e != nil {
				t.Fatal(e)
			}
			status, e := g.deliver(i, y, auth)
			if e != nil || status != "applied" || g.mix != state.mix || g.inner != state.inner || g.split != state.split {
				t.Fatal("immediate update parity", i, e)
			}
			checkJournal(t, g)
		}
	}
}

type failedJournalReader struct{ reads int }

func (r *failedJournalReader) Epoch() uint64 { return 1 }
func (r *failedJournalReader) Read(observation.View) (uint16, uint16, error) {
	r.reads++
	return 0, 0, errors.New("synthetic unavailable read")
}

func TestJournalLifecycle(t *testing.T) {
	base, trial, m := journalModels(t)
	g := forecastJournal{base: base, enabled: true}
	if e := g.publish(m, trial.model); e != nil {
		t.Fatal(e)
	}
	bad := &failedJournalReader{}
	before := g
	if _, e := g.predict(bad, 0, 0); e == nil || g != before || bad.reads != 1 {
		t.Fatal("failed read mutated issuance")
	}
	for i := 0; i < 64; i++ {
		if _, e := g.predict(observationexperiment.Frames(uint16(i), "journal-ring"), i, 0); e != nil {
			t.Fatal(e)
		}
	}
	before = g
	reads := bad.reads
	if _, e := g.predict(bad, 64, 0); e == nil || g != before || bad.reads != reads {
		t.Fatal("backpressure read or overwrote")
	}
	// Arrival order differs from origin order, but each update uses its own
	// saved expert probabilities, never the latest prediction in the ring.
	want := g.mix
	for _, i := range []int{7, 2, 63, 0} {
		saved := g.entries[i].forecast
		want = want.Observe(saved.p.Experts, i%2 == 0, 1)
		if status, e := g.deliver(i, i%2 == 0, false); e != nil || status != "applied" || g.mix != want {
			t.Fatal("arrival update mismatch")
		}
		before = g
		if _, e := g.deliver(i, true, true); e == nil || g != before {
			t.Fatal("duplicate mutation")
		}
	}
	if _, e := g.predict(observationexperiment.Frames(511, "ring-reuse"), 64, 0); e != nil {
		t.Fatal(e)
	}
	before = g
	if _, e := g.deliver(0, true, true); e == nil || g != before {
		t.Fatal("old origin consumed reused slot")
	}
	m.Version++
	if e := g.publish(m, trial.model); e != nil {
		t.Fatal(e)
	}
	weights, inner := g.mix, g.inner
	if status, e := g.deliver(64, true, true); e != nil || status != "stale" || g.mix != weights || g.inner != inner || g.split {
		t.Fatal("stale publication credit")
	}
	if _, e := g.expireBefore(g.issued); e != nil {
		t.Fatal(e)
	}
	if _, e := g.predict(observationexperiment.Frames(0, "split-generation"), 65, 0); e != nil {
		t.Fatal(e)
	}
	if _, e := g.predict(observationexperiment.Frames(511, "split-generation"), 66, 0); e != nil {
		t.Fatal(e)
	}
	if status, e := g.deliver(66, true, true); e != nil || status != "applied" || !g.split || g.generation != 1 {
		t.Fatal("split generation")
	}
	weights, inner = g.mix, g.inner
	if status, e := g.deliver(65, false, true); e != nil || status != "stale" || g.mix != weights || g.inner != inner {
		t.Fatal("pre-split credit")
	}
	if _, e := g.predict(observationexperiment.Frames(0, "censor"), 67, 0); e != nil {
		t.Fatal(e)
	}
	weights, inner = g.mix, g.inner
	if e := g.censor(67); e != nil || g.mix != weights || g.inner != inner {
		t.Fatal("censor applied label")
	}
	for _, origin := range []int{-1, 0, 67, 68, 1000} {
		before = g
		if _, e := g.deliver(origin, true, false); e == nil || g != before {
			t.Fatal("invalid origin mutation")
		}
	}
	for _, cut := range []int{-1, g.issued + 1} {
		before = g
		if _, e := g.expireBefore(cut); e == nil || g != before {
			t.Fatal("invalid cutoff mutation")
		}
	}
	before = g
	if e := g.publish(m, trial.model); e == nil || g != before {
		t.Fatal("stale publication")
	}
	checkJournal(t, g)
}

type journalStressRecord struct {
	Seed                 int64
	Issued               int
	Stats                forecastJournalStats
	MaxPending, Rejected int
	Tape                 string
}

func journalStress(t *testing.T, seed int64) journalStressRecord {
	t.Helper()
	base, trial, models := journalModels(t)
	g := forecastJournal{base: base, enabled: true}
	if e := g.publish(models, trial.model); e != nil {
		t.Fatal(e)
	}
	type packet struct {
		origin, due int
		y           bool
	}
	var packets []packet
	rng := rand.New(rand.NewSource(seed))
	r := journalStressRecord{Seed: seed}
	hash := sha256.New()
	enc := json.NewEncoder(hash)
	for tick := 0; tick < 2048; tick++ {
		if tick > 0 && tick%64 == 0 {
			models.Version++
			if e := g.publish(models, trial.model); e != nil {
				t.Fatal(e)
			}
		}
		if _, e := g.expireBefore(max(0, tick-48)); e != nil {
			t.Fatal(e)
		}
		x := uint16(rng.Intn(512))
		p, e := g.predict(observationexperiment.Frames(x, "journal-stress"), tick, 0)
		if e != nil {
			t.Fatal(e)
		}
		if e := enc.Encode(p); e != nil {
			t.Fatal(e)
		}
		// Outcomes belong only to the simulator packet until their due time.
		if rng.Float64() >= .2 {
			packets = append(packets, packet{tick, tick + rng.Intn(32), x&4 != 0})
		}
		r.MaxPending = max(r.MaxPending, g.stats.Pending)
		for i := 0; i < len(packets); {
			q := packets[i]
			if q.due > tick {
				i++
				continue
			}
			status, e := g.deliver(q.origin, q.y, q.origin == 300)
			if e != nil {
				r.Rejected++
				status = "rejected"
			}
			if e := enc.Encode(struct {
				Tick, Origin int
				Status       string
				Mix, Inner   bayes.ForecastMix
				Stats        forecastJournalStats
			}{tick, q.origin, status, g.mix, g.inner, g.stats}); e != nil {
				t.Fatal(e)
			}
			before := g
			if _, e := g.deliver(q.origin, !q.y, true); e == nil || g != before {
				t.Fatal("stress duplicate accepted")
			}
			packets = append(packets[:i], packets[i+1:]...)
		}
		checkJournal(t, g)
	}
	// Explicit end-of-stream censoring, not synthetic completion labels.
	if _, e := g.expireBefore(g.issued); e != nil {
		t.Fatal(e)
	}
	checkJournal(t, g)
	r.Issued, r.Stats, r.Tape = g.issued, g.stats, hex.EncodeToString(hash.Sum(nil))
	return r
}

func TestJournalStressV84(t *testing.T) {
	output, replay := os.Getenv("EVENTFRAME_JOURNAL_OUT"), os.Getenv("EVENTFRAME_JOURNAL_REPLAY")
	if output == "" && replay == "" {
		t.Skip("opt-in")
	}
	if output != "" && replay != "" {
		t.Fatal("choose write or replay")
	}
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	type artifact struct {
		Version string
		Hashes  map[string]string
		Records []journalStressRecord
	}
	a := artifact{Version: "v84", Hashes: map[string]string{}}
	if replay != "" {
		b, e := os.ReadFile(replay)
		if e != nil {
			t.Fatal(e)
		}
		if e := json.Unmarshal(b, &a); e != nil {
			t.Fatal(e)
		}
		if a.Version != "v84" || len(a.Records) != 20 {
			t.Fatal("archive shape")
		}
	}
	if replay == "" {
		var paths []string
		for _, dir := range []string{"observationgate", "observation", "observationlearners", "observationpreserved", "observationrescue", "observationexperiment", "bayes", "model"} {
			ps, e := filepath.Glob(filepath.Join(root, "internal", dir, "*.go"))
			if e != nil {
				t.Fatal(e)
			}
			paths = append(paths, ps...)
		}
		for _, p := range []string{"docs/experiments/mmm-feedback-journal-v84-protocol.md", "go.mod", "go.sum"} {
			paths = append(paths, filepath.Join(root, p))
		}
		for _, p := range paths {
			b, e := os.ReadFile(p)
			if e != nil {
				t.Fatal(e)
			}
			h := sha256.Sum256(b)
			rel, e := filepath.Rel(root, p)
			if e != nil {
				t.Fatal(e)
			}
			a.Hashes[rel] = hex.EncodeToString(h[:])
		}
	} else {
		for p, want := range a.Hashes {
			if !filepath.IsLocal(p) {
				t.Fatal("source path")
			}
			b, e := os.ReadFile(filepath.Join(root, p))
			if e != nil {
				t.Fatal(e)
			}
			h := sha256.Sum256(b)
			if hex.EncodeToString(h[:]) != want {
				t.Fatal("source changed", p)
			}
		}
	}
	for i := 0; i < 20; i++ {
		r := journalStress(t, 2026118401+int64(i))
		if replay != "" {
			if r != a.Records[i] {
				t.Fatal("stress replay mismatch")
			}
		} else {
			a.Records = append(a.Records, r)
		}
	}
	if replay != "" {
		return
	}
	f, e := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if e := json.NewEncoder(f).Encode(a); e != nil {
		t.Fatal(e)
	}
	if e := f.Sync(); e != nil {
		t.Fatal(e)
	}
}

func BenchmarkForecastJournal(b *testing.B) {
	base, e := observationpreserved.Base(false)
	if e != nil {
		b.Fatal(e)
	}
	for _, journal := range []bool{false, true} {
		name := "direct"
		if journal {
			name = "journal"
		}
		b.Run(name, func(b *testing.B) {
			g := forecastJournal{base: base, enabled: true}
			s := subsetState{base: base, enabled: true}
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				rd := observationexperiment.Frames(uint16(i%512), "journal-benchmark")
				if journal {
					if _, e := g.predict(rd, i, 0); e != nil {
						b.Fatal(e)
					}
					if _, e := g.deliver(i, i%2 == 0, false); e != nil {
						b.Fatal(e)
					}
				} else {
					if _, e := s.predict(rd, observationpreserved.Models{}, nil, i, 0); e != nil {
						b.Fatal(e)
					}
					if e := s.observe(i, i%2 == 0, false); e != nil {
						b.Fatal(e)
					}
				}
			}
		})
	}
}
