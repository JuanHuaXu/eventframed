package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"math/bits"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type originV105Update struct {
	Clock, Origin, Arrival, Wait, Profile int
	IssueVersion, UpdateVersion           uint64
	Raw, Before, After                    [4]float64
	Outcome, BankOnly, RiskMeasured       bool
	RiskBefore, RiskAfter                 float64
}
type originV105Profile struct {
	A [4][4]float64
	B [4]float64
	C float64
}

func (p originV105Profile) risk(w [4]float64) float64 {
	v := p.C
	for i := 0; i < 4; i++ {
		v -= 2 * w[i] * p.B[i]
		for j := 0; j < 4; j++ {
			v += w[i] * w[j] * p.A[i][j]
		}
	}
	return v
}
func makeV105Profile(e *observationExperts, truth func(uint16, int) float64, clock int) originV105Profile {
	var profile originV105Profile
	for x := uint16(0); x < 512; x++ {
		var p [4]float64
		for j := range p {
			p[j], _ = e.models[j].Forecast(511, x)
		}
		q := truth(x, clock)
		profile.C += q / 512
		for i := range p {
			profile.B[i] += q * p[i] / 512
			for j := range p {
				profile.A[i][j] += p[i] * p[j] / 512
			}
		}
	}
	return profile
}
func traceV105Operation(g *routedFeedbackJournal, clock, origin int, y bool, cutoff int, steps []delayedV104Step, trace *[]originV105Update) error {
	before := *g
	var err error
	if cutoff >= 0 {
		err = g.expireBefore(uint64(cutoff))
	} else {
		err = g.deliver(uint64(origin), y)
	}
	if err != nil {
		return err
	}
	if !g.carrySelector {
		return nil
	}
	entries := before.entries
	if cutoff >= 0 {
		for i, e := range entries {
			if e.active && !e.ready && int(e.origin) < cutoff {
				entries[i].censored = true
			}
		}
	} else {
		entries[origin%64].ready = true
		entries[origin%64].outcome = y
	}
	bank := before.core.bank
	for i := before.drainAt; i < g.drainAt; i++ {
		e := entries[i%64]
		if e.censored {
			continue
		}
		if !e.active || !e.ready || e.origin != i {
			return fmt.Errorf("invalid trace entry")
		}
		u := originV105Update{Clock: clock, Origin: int(i), Arrival: int(i) + steps[i].Delay, IssueVersion: e.version, UpdateVersion: g.version, Raw: e.bank.Experts, Outcome: e.outcome, BankOnly: e.version != g.version, Before: bank.weights()}
		u.Wait = clock - u.Arrival
		bank.pending = e.bank
		bank.active = true
		if err := bank.observe(i, e.outcome); err != nil {
			return err
		}
		u.After = bank.weights()
		*trace = append(*trace, u)
	}
	if bank != g.core.bank {
		return fmt.Errorf("traced bank reconstruction changed state")
	}
	return nil
}
func tracedV105Run(phase, scenario, index, schedule int, base int64, trace *[]originV105Update, profiles *[]originV105Profile) (delayedV104Record, error) {
	r := delayedV104Record{Schedule: schedule, Phase: []string{"design", "confirmation"}[phase], Case: replicationV102Cases[scenario], Index: index}
	seed := base + int64(phase*1000000+scenario*10000+index*10)
	rules, xs, ys, delays := rand.New(rand.NewSource(seed)), rand.New(rand.NewSource(seed+1)), rand.New(rand.NewSource(seed+2)), rand.New(rand.NewSource(seed+3))
	arities := [12]int{1, 2, 3, 4, 4, 3, 3, 0, 0, 4, 3, 4}
	for k := range r.Rules {
		a := arities[scenario]
		if scenario >= 10 && k == 1 {
			a = 7 - a
		}
		if a > 0 {
			pool := stackV93Masks(a, phase)
			r.Rules[k] = pool[rules.Intn(len(pool))]
		}
	}
	truth := func(x uint16, step int) float64 {
		name, rule := r.Case, r.Rules[0]
		if scenario >= 10 {
			name = "majority3"
			if scenario == 11 {
				name = "parity4"
			}
			if step >= 128 {
				rule = r.Rules[1]
				if scenario == 10 {
					name = "parity4"
				} else {
					name = "majority3"
				}
			}
		}
		return stackV93Truth(x, rule, name)
	}
	history := make([]observation.Sample, 0, 272)
	for step := -16; step < 0; step++ {
		x := stackV93Input(uint16(xs.Intn(512)), r.Case)
		history = append(history, observation.Sample{Bits: x, Outcome: ys.Float64() < truth(x, step)})
	}

	misses := rand.New(rand.NewSource(seed + 4))
	journals := [2]*routedFeedbackJournal{newRoutedFeedbackJournal(), newRoleRoutedFeedbackJournal()}
	var weights [512]float64
	for j := range weights {
		weights[j] = 1
	}
	var publication *observationExperts
	var arrived [256]bool
	measured := 0
	release := func(clock int) error {
		for origin, s := range r.Steps {
			if !arrived[origin] && !s.Missing && origin+s.Delay <= clock {
				for _, g := range journals {
					if err := traceV105Operation(g, clock, origin, s.Y, -1, r.Steps, trace); err != nil {
						return err
					}
				}
				arrived[origin] = true
			}
		}
		if clock >= 32 {
			for _, g := range journals {
				if err := traceV105Operation(g, clock, -1, false, clock-31, r.Steps, trace); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for clock := 0; clock < 288; clock++ {
		if err := release(clock); err != nil {
			return r, err
		}
		if clock >= 256 {
			continue
		}
		if clock%32 == 0 {
			var eligible []observation.Sample
			var origins []int
			eligible = append(eligible, history[:16]...)
			for i := -16; i < 0; i++ {
				origins = append(origins, i)
			}
			for i := 0; i < clock; i++ {
				if arrived[i] {
					eligible = append(eligible, history[i+16])
					origins = append(origins, i)
				}
			}
			if len(eligible) > 64 {
				eligible = eligible[len(eligible)-64:]
				origins = origins[len(origins)-64:]
			}
			short, shortOrigins := eligible, origins
			if len(short) > 32 {
				short = short[len(short)-32:]
				shortOrigins = shortOrigins[len(shortOrigins)-32:]
			}
			r.Fits = append(r.Fits, delayedV104Fit{Clock: clock, Origins: [2][]int{append([]int(nil), origins...), append([]int(nil), shortOrigins...)}})
			var models [4]*ConditionalForest
			var err error
			models[0], err = NewSubsetConditional(eligible, weights)
			if err != nil {
				return r, err
			}
			models[1], err = NewBooleanConditional(eligible)
			if err != nil {
				return r, err
			}
			models[2], err = NewSubsetConditional(short, weights)
			if err != nil {
				return r, err
			}
			models[3], err = NewBooleanConditional(short)
			if err != nil {
				return r, err
			}
			publication, err = newObservationExperts(models)
			if err != nil {
				return r, err
			}
			*profiles = append(*profiles, makeV105Profile(publication, truth, clock))
			for _, g := range journals {
				if err := g.publish(publication); err != nil {
					return r, err
				}
			}
		}
		for measured < len(*trace) {
			u := &(*trace)[measured]
			if u.Clock != clock {
				return r, fmt.Errorf("trace risk clock mismatch")
			}
			profile := (*profiles)[clock/32]
			u.RiskBefore = profile.risk(u.Before)
			u.RiskAfter = profile.risk(u.After)
			u.RiskMeasured = true
			u.Profile = clock / 32
			measured++
		}
		x := stackV93Input(uint16(xs.Intn(512)), r.Case)
		s := delayedV104Step{X: x}
		generic, err := RunConditionalObserver(&publication.models[0], &jointReader{x: x, epoch: 1}, 1)
		if err != nil {
			return r, err
		}
		s.P[0], s.Mask[0], s.Cost[0] = generic.Probability, generic.Trace[len(generic.Trace)-1].Observed, generic.Cost
		for j, g := range journals {
			got, err := g.predict(uint64(clock), &jointReader{x: x, epoch: 1}, 1)
			if err != nil {
				return r, err
			}
			s.P[j+1], s.Mask[j+1], s.Cost[j+1] = got.Probability, got.Trace[len(got.Trace)-1].Observed, got.Cost
		}
		// The current outcome is unavailable until after all three forecasts.
		s.Q = truth(x, clock)
		s.Y = ys.Float64() < s.Q
		d, missing := delays.Intn(32), misses.Float64() < .2
		if schedule == 1 {
			s.Delay, s.Missing = d, missing
		}
		target := 0.
		if s.Y {
			target = 1
		}
		for arm, p := range s.P {
			if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 || s.Cost[arm] < 1 || s.Cost[arm] > 6 || s.Cost[arm] != bits.OnesCount16(s.Mask[arm]) {
				return r, fmt.Errorf("invalid issued bundle")
			}
			loss := (p-s.Q)*(p-s.Q) + s.Q*(1-s.Q)
			acc := 1 - s.Q
			if p >= .5 {
				acc = s.Q
			}
			r.Metrics[arm][0].Brier += loss / 256
			r.Metrics[arm][0].Accuracy += acc / 256
			if clock >= 128 {
				r.Metrics[arm][1].Brier += loss / 128
				r.Metrics[arm][1].Accuracy += acc / 128
			}
			r.Realized[arm] += (p - target) * (p - target) / 256
		}
		r.Steps = append(r.Steps, s)
		history = append(history, observation.Sample{Bits: x, Outcome: s.Y})
		if !s.Missing && s.Delay == 0 {
			for _, g := range journals {
				if err := traceV105Operation(g, clock, clock, s.Y, -1, r.Steps, trace); err != nil {
					return r, err
				}
			}
			arrived[clock] = true
		}
		for measured < len(*trace) {
			u := &(*trace)[measured]
			if u.Clock != clock {
				return r, fmt.Errorf("immediate trace risk clock mismatch")
			}
			profile := (*profiles)[clock/32]
			u.RiskBefore, u.RiskAfter = profile.risk(u.Before), profile.risk(u.After)
			u.RiskMeasured, u.Profile = true, clock/32
			measured++
		}
		for j, g := range journals {
			z := g.stats
			if z.Issued != z.Applied+z.BankOnly+z.Stale+z.Censored+z.Pending || z.Pending > 64 {
				return r, fmt.Errorf("journal accounting")
			}
			r.Steps[clock].Stats[j] = z
		}
	}
	for j, g := range journals {
		r.Final[j] = g.stats
		if g.stats.Pending != 0 {
			return r, fmt.Errorf("unsettled final journal")
		}
	}
	for _, v := range arrived {
		if v {
			r.Arrived++
		}
	}
	return r, nil
}

type originV105Record struct {
	Record   delayedV104Record
	Updates  []originV105Update
	Profiles []originV105Profile
}

func TestOriginV105Profile(t *testing.T) {
	e, err := newObservationExperts(jointFixture(false, false))
	if err != nil {
		t.Fatal(err)
	}
	truth := func(x uint16, _ int) float64 { return .05 + .9*float64(bits.OnesCount16(x&17)%2) }
	profile := makeV105Profile(e, truth, 0)
	for _, w := range [][4]float64{{1, 0, 0, 0}, {0, 1, 0, 0}, {.25, .25, .25, .25}, {.95, .05 / 3, .05 / 3, .05 / 3}} {
		want := 0.
		for x := uint16(0); x < 512; x++ {
			p := 0.
			for j := range w {
				v, _ := e.models[j].Forecast(511, x)
				p += w[j] * v
			}
			q := truth(x, 0)
			want += ((p-q)*(p-q) + q*(1-q)) / 512
		}
		if math.Abs(want-profile.risk(w)) > 1e-12 {
			t.Fatal("literal Brier profile")
		}
	}
}

func TestOriginV105Smoke(t *testing.T) {
	old, err := delayedV104Run(0, 10, 0, 1, 2100110400)
	if err != nil {
		t.Fatal(err)
	}
	var updates []originV105Update
	var profiles []originV105Profile
	got, err := tracedV105Run(0, 10, 0, 1, 2100110400, &updates, &profiles)
	if err != nil || !reflect.DeepEqual(got, old) {
		t.Fatal("trace side effect", err)
	}
	for _, u := range updates {
		if u.RiskMeasured != (u.Clock < 256) || u.Wait < 0 {
			t.Fatal("trace timing")
		}
	}
}

type originV105Artifact struct {
	Version, ParentSHA256 string
	Hashes                map[string]string
	Records               []originV105Record
}

func TestOriginV105(t *testing.T) {
	output := os.Getenv("EVENTFRAME_ORIGIN_V105")
	if output == "" {
		t.Skip("explicit artifact required")
	}
	raw, err := os.ReadFile("../../docs/experiments/mmm-delayed-v104.json")
	if err != nil {
		t.Fatal(err)
	}
	var parent delayedV104Artifact
	if err := json.Unmarshal(raw, &parent); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "73003cee91b7c7c8b16e6b859391fbeb7da3deb39e92d1bec118175f09f8f760" {
		t.Fatal("parent changed")
	}
	a := originV105Artifact{Version: "v105", ParentSHA256: fmt.Sprintf("%x", sha256.Sum256(raw)), Hashes: map[string]string{}}
	for _, name := range []string{"internal/observationlearners/origin_v105_test.go", "research/origin-v105-summary.mjs", "research/origin-v105-protocol.md"} {
		data, err := os.ReadFile(filepath.Join("../..", name))
		if err != nil {
			t.Fatal(err)
		}
		a.Hashes[name] = fmt.Sprintf("%x", sha256.Sum256(data))
	}
	for _, old := range parent.Records {
		if old.Schedule != 1 || (old.Case != "majority_to_parity" && old.Case != "parity_to_majority") {
			continue
		}
		phase, scenario := 0, 10
		if old.Phase == "confirmation" {
			phase = 1
		}
		if old.Case == "parity_to_majority" {
			scenario = 11
		}
		var d originV105Record
		d.Record, err = tracedV105Run(phase, scenario, old.Index, old.Schedule, parent.SeedBase, &d.Updates, &d.Profiles)
		if err != nil || !reflect.DeepEqual(d.Record, old) {
			t.Fatal("instrumentation changed replay", old.Phase, old.Case, old.Index, err)
		}
		if len(d.Updates) != old.Arrived || len(d.Profiles) != 8 {
			t.Fatal("trace coverage")
		}
		a.Records = append(a.Records, d)
	}
	if len(a.Records) != 128 {
		t.Fatal("diagnostic count")
	}
	if os.Getenv("EVENTFRAME_ORIGIN_V105_REPLAY") == "1" {
		raw, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		var old originV105Artifact
		if err := json.Unmarshal(raw, &old); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(old, a) {
			t.Fatal("trace replay mismatch")
		}
		return
	}
	raw, err = json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(raw); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
