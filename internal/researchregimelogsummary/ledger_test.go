package researchregimelogsummary

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
)

type hidden struct{ state, start int }

// Dense latent trajectory plus last global reset; no candidate filters or masks
// computed from its own posterior. Only the externally supplied masks are used.
func restrictedDense(base []float64, cfg Config, rows []Row, support [][]Key) ([]float64, float64) {
	prior, width := oraclePrior(base), 1<<len(base)
	p := map[hidden]float64{}
	for s, w := range prior {
		p[hidden{s, -1}] = w
	}
	logZ := 0.
	for t, row := range rows {
		next := map[hidden]float64{}
		for from, w := range p {
			for to, pi := range prior {
				if cfg.Reset > 0 {
					next[hidden{to, t}] += w * cfg.Reset * pi
				}
				fh, fm, th, tm := from.state/width, from.state%width, to/width, to%width
				if cfg.Reset < 1 && fh == th && fm&^(1<<row.Member) == tm&^(1<<row.Member) {
					x := oracleHigh(base[row.Member], th)
					if tm&(1<<row.Member) == 0 {
						x = 1 - x
					}
					q := cfg.Hazard * x
					if fm == tm {
						q += 1 - cfg.Hazard
					}
					next[hidden{to, from.start}] += w * (1 - cfg.Reset) * q
				}
			}
		}
		allowed := map[Key]bool{}
		for _, k := range support[t] {
			allowed[k] = true
		}
		z := 0.
		for s, w := range next {
			if !allowed[Key{s.state / width, s.start}] {
				delete(next, s)
				continue
			}
			w *= oracleFactor(s.state, len(base), row)
			next[s] = w
			z += w
		}
		if z == 0 {
			return nil, math.Inf(-1)
		}
		for s := range next {
			next[s] /= z
		}
		logZ += math.Log(z)
		p = next
	}
	out := make([]float64, len(prior))
	for s, w := range p {
		out[s.state] += w
	}
	return out, logZ
}

type audit struct {
	SummaryParityChecks                                        int
	MaxSummaryParity                                           float64
	OldEnvelopeComparisons                                     int
	MaxOldEnvelopeDifference                                   float64
	EnvelopeOracleChecks, NonvacuousEnvelopes                  int
	MaxEnvelopeActualTV                                        float64
	Configurations, Checks, Publications, Queries, Abstentions int
	MaxOracle, MaxTower, MaxReceipt                            float64
	Full64                                                     bool
	RepruningChanged                                           bool
	WholeGoalValidation                                        bool
	IncrementalChecks, CacheChecks                             int
	MaxReference                                               float64
	FullWorkload, OldReplay                                    bool
	ProtectedChecks                                            int
	NoResetUnderflow                                           bool
	ExactLongPairLogProbability                                float64
	LongLogRescue, MemberOddsRescue                            bool
	MaxLogOracle                                               float64
}

var report audit

func TestMain(m *testing.M) {
	code := m.Run()
	if path := os.Getenv("EVENTFRAME_REGIME_V83_REPORT"); path != "" {
		b, e := json.MarshalIndent(report, "", "  ")
		if e == nil {
			e = os.WriteFile(path, append(b, '\n'), 0600)
		}
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			code = 1
		}
	}
	os.Exit(code)
}

func near(t *testing.T, a, b float64) {
	t.Helper()
	if !finite(a) || !finite(b) || math.Abs(a-b) > 2e-11 {
		t.Fatalf("oracle comparison %.17g %.17g", a, b)
	}
	report.Checks++
	report.MaxOracle = math.Max(report.MaxOracle, math.Abs(a-b))
}

func inspect(t *testing.T, m *Ledger) []float64 {
	t.Helper()
	s := m.Snapshot()
	p, z := restrictedDense(m.base, m.cfg, s.Rows, s.Support)
	if math.IsInf(z, -1) {
		t.Fatal("published impossible evidence")
	}
	u, e := m.state.Joint(len(m.base))
	if e != nil {
		t.Fatal(e)
	}
	for i := range p {
		near(t, p[i], u[i])
	}
	near(t, z, s.LogEvidence)
	for i, q := range s.NextClean {
		near(t, q, denseForecast(m.base, m.cfg, p, i))
	}
	if s.Token != m.identity(s.Rows, s.Support, s.AsOf, s.Version, s.SupportEpoch) {
		t.Fatal("publication identity drift")
	}
	report.Publications++
	return p
}

func queried(t *testing.T, m *Ledger, ordinal, which int) (Query, bool) {
	t.Helper()
	before := m.Snapshot()
	q, e := m.Pending(before.Token, ordinal, which)
	if !reflect.DeepEqual(before, m.Snapshot()) {
		t.Fatal("query published or mutated")
	}
	var probs [2]float64
	var expected [2][]float64
	zeros := 0
	for value := 0; value < 2; value++ {
		rows := append([]Row(nil), before.Rows...)
		if which == 1 {
			rows[ordinal].First = value
		} else {
			rows[ordinal].Second = value
		}
		p, z := restrictedDense(m.base, m.cfg, rows, before.Support)
		if math.IsInf(z, -1) {
			zeros++
			continue
		}
		probs[value] = math.Exp(z - before.LogEvidence)
		expected[value] = make([]float64, len(m.base))
		for i := range m.base {
			expected[value][i] = denseForecast(m.base, m.cfg, p, i)
		}
	}
	if e != nil {
		if zeros == 0 {
			t.Fatal("query abstained on two supported branches", e)
		}
		report.Abstentions++
		return Query{}, false
	}
	if zeros > 0 {
		t.Fatal("unsupported query silently assigned mass")
	}
	for value := 0; value < 2; value++ {
		near(t, q.Branches[value].Probability, probs[value])
		for i, p := range expected[value] {
			near(t, q.Branches[value].NextClean[i], p)
		}
	}
	defect := math.Abs(q.Branches[0].Probability + q.Branches[1].Probability - 1)
	for i, p := range before.NextClean {
		weighted := 0.
		for _, branch := range q.Branches {
			weighted += branch.Probability * branch.NextClean[i]
		}
		defect = math.Max(defect, math.Abs(p-weighted))
	}
	near(t, defect, 0)
	report.MaxTower = math.Max(report.MaxTower, defect)
	report.Queries++
	return q, true
}

func receiptOracle(t *testing.T, m *Ledger, r Receipt, p []float64) {
	t.Helper()
	member := m.rows[r.Ordinal].Member
	width := 1 << len(m.base)
	clean, first := 0., 0.
	for state, w := range p {
		rate := .02
		if state%width&(1<<member) != 0 {
			rate = .98
		}
		eta := []float64{0, .1, .2}[(state/width)%3]
		clean += w * rate
		first += w * (eta + (1-2*eta)*rate)
	}
	near(t, clean, r.Clean)
	near(t, first, r.First)
	report.MaxReceipt = math.Max(report.MaxReceipt, math.Max(math.Abs(clean-r.Clean), math.Abs(first-r.First)))
	q, ok := queried(t, m, r.Ordinal, 1)
	if !ok {
		t.Fatal("first observation always has positive support")
	}
	near(t, r.First, q.Branches[1].Probability)
}

func TestCanonicalIssueRevealAndRefresh(t *testing.T) {
	all := []float64{.25, .47, .925}
	for members := 1; members <= 3; members++ {
		for _, kappa := range []float64{0, 1. / 16, .5, 1} {
			for _, lambda := range []float64{0, .25, 1} {
				for _, cap := range []int{9, 18, 36} {
					cfg := Config{kappa, lambda, cap}
					m, e := New(all[:members], cfg)
					if e != nil {
						t.Fatal(e)
					}
					report.Configurations++
					for i := 0; i < 12; i++ {
						before := m.Snapshot()
						r, e := m.Issue(i%members, int64(i))
						if e != nil {
							t.Fatal(e)
						}
						after := m.Snapshot()
						if !reflect.DeepEqual(before.Support, after.Support[:i]) || after.Version != before.Version+1 || after.SupportEpoch != before.SupportEpoch+1 {
							t.Fatal("issue refit historical masks or hid epoch")
						}
						receiptOracle(t, m, r, inspect(t, m))
					}
					for _, i := range []int{8, 0, 5, 2, 11, 4} {
						before := m.Snapshot()
						q, ok := queried(t, m, i, 1)
						if !ok {
							t.Fatal("first query support")
						}
						if e = m.Reveal(before.Token, i, 1, i%2, 12); e != nil {
							t.Fatal(e)
						}
						after := m.Snapshot()
						if !reflect.DeepEqual(before.Support, after.Support) || after.SupportEpoch != before.SupportEpoch || after.Version != before.Version+1 {
							t.Fatal("reveal changed support")
						}
						for j, p := range q.Branches[i%2].NextClean {
							near(t, p, after.NextClean[j])
						}
						inspect(t, m)
					}
					for _, i := range []int{0, 8, 11} {
						before := m.Snapshot()
						queried(t, m, i, 2)
						// Same-valued second observation is supported even in eta0.
						if e = m.Reveal(before.Token, i, 2, i%2, 12); e != nil {
							t.Fatal(e)
						}
						if !reflect.DeepEqual(before.Support, m.Snapshot().Support) {
							t.Fatal("pair refit masks")
						}
						inspect(t, m)
					}
					before := m.Snapshot()
					if e = m.Refresh(before.Token, 13); e != nil {
						t.Fatal(e)
					}
					if m.token.SupportEpoch != before.SupportEpoch+1 || m.token.Version != before.Version+1 || !reflect.DeepEqual(before.Rows, m.rows) {
						t.Fatal("refresh boundary")
					}
					inspect(t, m)
					old, e := Run(m.base, m.cfg, m.rows)
					if e != nil {
						t.Fatal(e)
					}
					if !reflect.DeepEqual(before.Support, old.Support) {
						report.RepruningChanged = true
					}
				}
			}
		}
	}
	if !report.RepruningChanged {
		t.Fatal("re-pruning comparison was vacuous")
	}
	t.Logf("%d configurations, %d publications, %d checks, max oracle %g, tower %g, receipt %g, explicit query abstentions %d", report.Configurations, report.Publications, report.Checks, report.MaxOracle, report.MaxTower, report.MaxReceipt, report.Abstentions)
}

func TestFull64DelayedAndFutureForks(t *testing.T) {
	base := []float64{.25, .925}
	for _, kappa := range []float64{0, 1. / 16, .5, 1} {
		for _, cap := range []int{9, 18, 36} {
			cfg := Config{kappa, .25, cap}
			a, e := New(base, cfg)
			if e != nil {
				t.Fatal(e)
			}
			b, e := New(base, cfg)
			if e != nil {
				t.Fatal(e)
			}
			for i := 0; i < 64; i++ {
				x, e := a.Issue(i%2, int64(i))
				if e != nil {
					t.Fatal(e)
				}
				y, e := b.Issue(i%2, int64(i))
				if e != nil {
					t.Fatal(e)
				}
				if x != y {
					t.Fatal("as-of issue fork")
				}
			}
			for _, i := range []int{63, 0, 32, 17, 48, 1} {
				queried(t, a, i, 1)
				if e = a.Reveal(a.token, i, 1, i%2, 64); e != nil {
					t.Fatal(e)
				}
				if e = b.Reveal(b.token, i, 1, i%2, 64); e != nil {
					t.Fatal(e)
				}
				if !reflect.DeepEqual(a.Snapshot(), b.Snapshot()) {
					t.Fatal("equal arrival fork")
				}
				inspect(t, a)
			}
			// Fork a newly issued event. Under kappa1 an OLD event also has
			// no influence on the current state, not only on the next forecast.
			ra, e := a.Issue(0, 65)
			if e != nil {
				t.Fatal(e)
			}
			rb, e := b.Issue(0, 65)
			if e != nil || ra != rb {
				t.Fatal("unrevealed latest issue fork", e)
			}
			before := a.Snapshot()
			if e = a.Reveal(a.token, ra.Ordinal, 1, 0, 66); e != nil {
				t.Fatal(e)
			}
			if e = b.Reveal(b.token, rb.Ordinal, 1, 1, 66); e != nil {
				t.Fatal(e)
			}
			x, e := a.state.Joint(2)
			if e != nil {
				t.Fatal(e)
			}
			y, e := b.state.Joint(2)
			if e != nil {
				t.Fatal(e)
			}
			if reflect.DeepEqual(x, y) {
				t.Fatal("revealed current posterior fork vacuous")
			}
			if kappa == 1 {
				// A guaranteed reset severs influence on the next event, not on
				// the current posterior: require the independent limiting law.
				for i, p := range a.Snapshot().NextClean {
					near(t, p, b.Snapshot().NextClean[i])
				}
			} else if reflect.DeepEqual(a.Snapshot().NextClean, b.Snapshot().NextClean) {
				t.Fatal("persistent future predictive fork vacuous")
			}
			if reflect.DeepEqual(before.Token, a.token) {
				t.Fatal("future reveal identity unchanged")
			}
			inspect(t, a)
			inspect(t, b)
		}
	}
	report.Full64 = true
}

func TestOwnershipBindingsAndFailedOperations(t *testing.T) {
	base := []float64{.3, .8}
	m, e := New(base, Config{1. / 16, .25, 18})
	if e != nil {
		t.Fatal(e)
	}
	base[0] = .8
	if m.base[0] != .3 {
		t.Fatal("constructor retained mutable input")
	}
	for i := 0; i < 4; i++ {
		if _, e = m.Issue(i%2, int64(i)); e != nil {
			t.Fatal(e)
		}
	}
	stale := m.token
	if e = m.Reveal(m.token, 0, 1, 1, 4); e != nil {
		t.Fatal(e)
	}
	before := m.Snapshot()
	for _, f := range []func() error{
		func() error { return m.Reveal(stale, 1, 1, 1, 4) },
		func() error { return m.Reveal(m.token, 0, 1, 1, 4) },
		func() error { return m.Reveal(m.token, 1, 2, 1, 4) },
		func() error { return m.Reveal(m.token, 4, 1, 1, 4) },
		func() error { return m.Reveal(m.token, 1, 1, 2, 4) },
		func() error { return m.Reveal(m.token, 1, 1, 1, 3) },
		func() error { return m.Refresh(stale, 4) },
		func() error { return m.Refresh(m.token, 3) },
		func() error { _, e := m.Issue(-1, 4); return e },
		func() error { _, e := m.Issue(0, 3); return e },
		func() error { _, e := m.Pending(stale, 1, 1); return e },
	} {
		if e = f(); e == nil || !reflect.DeepEqual(before, m.Snapshot()) {
			t.Fatal("invalid operation accepted/published", e)
		}
	}
	s := m.Snapshot()
	s.Rows[0].First = 0
	s.Support[0][0].Class = 8
	s.NextClean[0] = 0
	if !reflect.DeepEqual(before, m.Snapshot()) {
		t.Fatal("snapshot aliases ledger")
	}
	q, ok := queried(t, m, 1, 1)
	if !ok {
		t.Fatal("query support")
	}
	q.Branches[0].NextClean[0] = 0
	if !reflect.DeepEqual(before, m.Snapshot()) {
		t.Fatal("query result aliases ledger")
	}
	if e = m.Refresh(m.token, 4); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Pending(before.Token, 1, 1); e == nil {
		t.Fatal("refresh did not invalidate query")
	}
	for _, bad := range []Config{{math.NaN(), .25, 9}, {.1, math.Inf(1), 9}, {.1, .25, 1}, {.1, .25, 257}} {
		if _, e = New([]float64{.3}, bad); e == nil {
			t.Fatal("invalid configuration")
		}
	}
	// A zero-noise-only path forbids opposite same-Y observations. This
	// controlled internal fixture is valid, not a production support mutation.
	z, e := New([]float64{.3}, Config{0, .25, 9})
	if e != nil {
		t.Fatal(e)
	}
	rows := []Row{{0, 1, -1}}
	r, e := Conditional(z.base, z.cfg, rows, [][]Key{{{0, -1}}})
	if e != nil {
		t.Fatal(e)
	}
	z.publish(rows, r, 1, 1)
	before = z.Snapshot()
	if e = z.Reveal(z.token, 0, 2, 0, 2); e == nil || !reflect.DeepEqual(before, z.Snapshot()) {
		t.Fatal("impossible reveal fail-open")
	}
	if _, e = z.Pending(z.token, 0, 2); e == nil || !reflect.DeepEqual(before, z.Snapshot()) {
		t.Fatal("impossible query fail-open")
	}
	if e = z.Reveal(z.token, 0, 2, 1, 2); e != nil {
		t.Fatal("supported adjacent update rejected", e)
	}
}

func TestEmptySnapshotAndRefreshIdentity(t *testing.T) {
	m, e := New([]float64{.3, .8}, Config{1. / 16, .25, 18})
	if e != nil {
		t.Fatal(e)
	}
	inspect(t, m)
	s := m.Snapshot()
	if s.LawID != m.identity([]Row{}, [][]Key{}, s.AsOf, s.Version, s.SupportEpoch).LawID {
		t.Fatal("empty readback changes law identity")
	}
	if e = m.Refresh(s.Token, 0); e != nil {
		t.Fatal(e)
	}
	inspect(t, m)
	if m.token.SupportEpoch != s.SupportEpoch+1 || m.token.Version != s.Version+1 {
		t.Fatal("empty refresh hides explicit publication")
	}
}

func TestJointDimensionBindingAndRescuedLatestPair(t *testing.T) {
	r, e := Run([]float64{.3}, Config{1. / 16, .25, 9}, nil)
	if e != nil {
		t.Fatal(e)
	}
	for _, members := range []int{0, 2, 3, 4, 5} {
		if _, e = r.Joint(members); e == nil {
			t.Fatal("joint invented another model dimension", members)
		}
	}
	if _, e = r.Joint(1); e != nil {
		t.Fatal(e)
	}
	if _, e = (Result{}).Joint(1); e == nil {
		t.Fatal("uninitialized result became a law")
	}
	m, e := New([]float64{.25, .925}, Config{1. / 16, .25, 36})
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 64; i++ {
		if _, e = m.Issue(i%2, int64(i)); e != nil {
			t.Fatal(e)
		}
	}
	for _, i := range []int{0, 32, 63} {
		if e = m.Reveal(m.token, i, 1, i%2, 64); e != nil {
			t.Fatal(e)
		}
	}
	if latestPairImpossible(m) {
		t.Fatal("protected support lost noisy alternatives")
	}
	if _, ok := queried(t, m, 63, 2); !ok {
		t.Fatal("protected opposite pair unsupported")
	}
	if e = m.Reveal(m.token, 63, 2, 0, 65); e != nil {
		t.Fatal("supported contradictory evidence rejected", e)
	}
	inspect(t, m)
}

// For the latest row, all positive-weight paths with eta0 and known W1 imply
// W2=W1. This is a structural witness, not classification by an error message.
func latestPairImpossible(m *Ledger) bool {
	if len(m.rows) == 0 {
		return false
	}
	r := m.rows[len(m.rows)-1]
	if r.First < 0 || r.Second != -1 {
		return false
	}
	positive := false
	for _, c := range m.state.parts {
		if !math.IsInf(c.logWeight, -1) {
			positive = true
			if c.h%3 != 0 {
				return false
			}
		}
	}
	return positive
}
