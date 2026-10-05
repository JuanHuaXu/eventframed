// Generated mechanically from the V54 fixture. Identical consumed populations.
package researchdispersion

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	paired "github.com/JuanHuaXu/eventframed/internal/researchpairedfast"
	ref "github.com/JuanHuaXu/eventframed/internal/researchpairedref"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"
	"time"
)

var regimesPairedV55 = append(append([]string(nil), regimesNoiseV53...), "partial_noise10_missing20", "stationary_noise20_missing20")
var modesPairedV55 = []string{"full", "adaptive", "no_pair", "random", "uncertainty", "information", "falsification"}

// Domain-separated streams avoid additive-salt collisions across nearby
// reserved cohort seeds. These are simulation streams, not source signatures.
func channelSeedPairedV55(seed int64, channel string) int64 {
	b, e := json.Marshal(struct {
		Namespace string
		World     int64
		Channel   string
	}{"paired-v54", seed, channel})
	if e != nil {
		panic(e)
	}
	h := sha256.Sum256(b)
	return int64(binary.LittleEndian.Uint64(h[:8]))
}

type populationPairedV55 struct {
	World                    windowWorldV37
	Truth, Second, Available []bool
	UniformLag               []int
}
type decisionPairedV55 struct {
	Round, At int
	Members   []int
	Options   []paired.Option
}
type auditOutcomePairedV55 struct {
	Trial, RequestedAt, ArrivedAt int
	Forecast                      float64
	Available, Value              bool
}
type costsPairedV55 struct{ SetupNS, ScheduleNS, IssueNS, FirstResolveNS, ProposalNS, RequestNS, SecondResolveNS, SnapshotNS, AccountedNS, ElapsedNS int64 }
type armPairedV55 struct {
	orientationArmV38
	ObservedIssued []float64
	NoiseIssued    [][3]float64
	NoiseSnapshots [][3]float64
	Decisions      []decisionPairedV55
	AuditOutcomes  []auditOutcomePairedV55
	Breakdown      costsPairedV55
}
type worldPairedV55 struct {
	Population populationPairedV55
	Arms       []armPairedV55
}

func makePairedV55(seed int64, g, r, id int) populationPairedV55 {
	var w windowWorldV37
	if r < 18 {
		w = makeNoiseV53(seed, g, r, id)
	} else {
		old := 13
		if r == 19 {
			old = 14
		}
		w = makeNoiseV53(seed+int64(r)*1000000000000, g, old, id)
		w.Regime = regimesPairedV55[r]
	}
	p := populationPairedV55{World: w, Truth: make([]bool, 2400), Second: make([]bool, 2400), Available: make([]bool, 2400), UniformLag: make([]int, 2400)}
	truth := rand.New(rand.NewSource(channelSeedPairedV55(w.Seed, "truth")))
	first := rand.New(rand.NewSource(channelSeedPairedV55(w.Seed, "first")))
	second := rand.New(rand.NewSource(channelSeedPairedV55(w.Seed, "second")))
	available := rand.New(rand.NewSource(channelSeedPairedV55(w.Seed, "availability")))
	lag := rand.New(rand.NewSource(channelSeedPairedV55(w.Seed, "secondary_lag")))
	eta := 0.
	switch r {
	case 12, 13, 16, 17, 18:
		eta = .1
	case 14, 15, 19:
		eta = .2
	}
	for round := 0; round < 16; round++ {
		shared := false
		if r == 16 {
			shared = first.Float64() < eta
		}
		for i := 0; i < 150; i++ {
			k := round*150 + i
			y := truth.Float64() < w.Rates[round][i]
			a, b := y, y
			if r == 16 {
				if shared {
					a, b = !y, !y
				}
			} else {
				if first.Float64() < eta {
					a = !y
				}
				if second.Float64() < eta {
					b = !y
				}
			}
			p.Truth[k], p.Second[k] = y, b
			p.World.Outcomes[round][i] = a
			p.Available[k] = true
			if r >= 18 {
				p.Available[k] = available.Float64() >= .2
			}
			p.UniformLag[k] = lag.Intn(300)
		}
	}
	return p
}
func pairedDelayV55(p populationPairedV55, k int, schedule string) (int, error) {
	switch schedule {
	case "immediate":
		return 0, nil
	case "fixed150":
		return 150, nil
	case "uniform299":
		return p.UniformLag[k], nil
	}
	return 0, fmt.Errorf("paired schedule")
}
func optionScorePairedV55(o paired.Option, mode string) float64 {
	switch mode {
	case "uncertainty":
		return o.Uncertainty
	case "information":
		return o.Information
	case "falsification":
		return o.EdgeCut
	}
	return 0
}

// This function cannot access the generator's truth, rates, second labels or
// source availability. Random does not run the candidate scoring loop.
func choosePairedV55(m *paired.Model, tickets []paired.Ticket, round, at int, mode string, rng *rand.Rand) (decisionPairedV55, error) {
	d := decisionPairedV55{Round: round, At: at}
	if mode == "random" {
		d.Members = append([]int(nil), rng.Perm(150)[:25]...)
		return d, nil
	}
	if mode != "uncertainty" && mode != "information" && mode != "falsification" {
		return d, fmt.Errorf("paired policy")
	}
	d.Options = make([]paired.Option, 150)
	order := make([]int, 150)
	for i := range order {
		order[i] = i
		var e error
		d.Options[i], e = m.Query(tickets[round*150+i], mode)
		if e != nil {
			return d, e
		}
	}
	sort.SliceStable(order, func(i, j int) bool {
		return optionScorePairedV55(d.Options[order[i]], mode) > optionScorePairedV55(d.Options[order[j]], mode)
	})
	d.Members = append([]int(nil), order[:25]...)
	return d, nil
}
func runPairedV55(p populationPairedV55, mode, schedule string) (armPairedV55, error) {
	w := p.World
	if mode == "full" || mode == "adaptive" {
		a, e := runMomentV41(w, mode, schedule)
		return armPairedV55{orientationArmV38: a}, e
	}
	started := time.Now()
	a := armPairedV55{orientationArmV38: orientationArmV38{ReadyAt: -1, windowArmV37: windowArmV37{Mode: mode, Schedule: schedule, Issued: make([]float64, 2400), ExpertIssued: make([][4]float64, 2400)}}, ObservedIssued: make([]float64, 2400), NoiseIssued: make([][3]float64, 2400)}
	phase := time.Now()
	due, e := scheduleV36(delayedWorldV36{Seed: w.Seed, Outcomes: w.Outcomes}, schedule)
	a.Breakdown.ScheduleNS = time.Since(phase).Nanoseconds()
	if e != nil {
		return a, e
	}
	secondary := make([][]int, len(due))
	firstCount := [16]int{}
	last := len(due) - 1
	for len(due[last]) == 0 {
		last--
	}
	phase = time.Now()
	m, e := paired.New(w.Base, 1, 2800, paired.Config{Strength: 2, Hazard: 1. / 16})
	a.Breakdown.SetupNS = time.Since(phase).Nanoseconds()
	if e != nil {
		return a, e
	}
	tickets := make([]paired.Ticket, 2400)
	audits := make([]paired.Ticket, 2400)
	requestedAt := make([]int, 2400)
	rng := rand.New(rand.NewSource(channelSeedPairedV55(w.Seed, "random_policy")))
	nextRound := 0
	for tick := 0; tick <= last; tick++ {
		if tick < 2400 {
			phase = time.Now()
			_, observed, e := m.Predict(tick % 150)
			if e != nil {
				return a, e
			}
			t, e := m.Issue(tick%150, int64(tick))
			a.Breakdown.IssueNS += time.Since(phase).Nanoseconds()
			if e != nil {
				return a, e
			}
			tickets[tick] = t
			a.Issued[tick] = t.Forecast()
			a.ObservedIssued[tick] = observed
			a.NoiseIssued[tick] = m.NoiseWeights()
			a.PeakPending = max(a.PeakPending, m.Pending())
		}
		for _, k := range due[tick] {
			phase = time.Now()
			r, e := m.Resolve(tickets[k], w.Outcomes[k/150][k%150], int64(tick))
			a.Breakdown.FirstResolveNS += time.Since(phase).Nanoseconds()
			if e != nil {
				return a, e
			}
			a.Receipts = append(a.Receipts, DelayedReceipt{Member: r.Member, TrialOrdinal: r.Ordinal, Epoch: r.Epoch, IssuedAt: r.IssuedAt, ArrivedAt: r.ArrivedAt, Forecast: r.Forecast, Useful: r.Value})
			firstCount[k/150]++
		}
		for nextRound < 16 && firstCount[nextRound] == 150 {
			if mode != "no_pair" {
				phase = time.Now()
				d, e := choosePairedV55(m, tickets, nextRound, tick, mode, rng)
				a.Breakdown.ProposalNS += time.Since(phase).Nanoseconds()
				if e != nil {
					return a, e
				}
				a.Decisions = append(a.Decisions, d)
				for _, i := range d.Members {
					k := nextRound*150 + i
					phase = time.Now()
					audits[k], e = m.RequestAudit(tickets[k], int64(tick))
					a.Breakdown.RequestNS += time.Since(phase).Nanoseconds()
					if e != nil {
						return a, e
					}
					requestedAt[k] = tick
					delay, e := pairedDelayV55(p, k, schedule)
					if e != nil {
						return a, e
					}
					if tick+delay >= len(secondary) {
						return a, fmt.Errorf("secondary time cap")
					}
					secondary[tick+delay] = append(secondary[tick+delay], k)
					last = max(last, tick+delay)
					a.PeakPending = max(a.PeakPending, m.Pending())
				}
			}
			nextRound++
		}
		for _, k := range secondary[tick] {
			phase = time.Now()
			if p.Available[k] {
				_, e = m.Resolve(audits[k], p.Second[k], int64(tick))
			} else {
				e = m.Cancel(audits[k], int64(tick))
			}
			a.Breakdown.SecondResolveNS += time.Since(phase).Nanoseconds()
			if e != nil {
				return a, e
			}
			value := false
			if p.Available[k] {
				value = p.Second[k]
			}
			a.AuditOutcomes = append(a.AuditOutcomes, auditOutcomePairedV55{Trial: k, RequestedAt: requestedAt[k], ArrivedAt: tick, Forecast: audits[k].Forecast(), Available: p.Available[k], Value: value})
		}
		if (tick < 2400 && tick%150 == 149) || tick == last {
			phase = time.Now()
			s := windowSnapshotV37{delayedSnapshotV36: delayedSnapshotV36{Tick: tick, Arrived: len(a.Receipts), Pending: m.Pending(), Forecast: make([]float64, 150)}}
			for i := range s.Forecast {
				s.Forecast[i], _, e = m.Predict(i)
				if e != nil {
					return a, e
				}
			}
			a.NoiseSnapshots = append(a.NoiseSnapshots, m.NoiseWeights())
			a.Snapshots = append(a.Snapshots, s)
			a.Breakdown.SnapshotNS += time.Since(phase).Nanoseconds()
		}
	}
	c := a.Breakdown
	c.AccountedNS = c.SetupNS + c.IssueNS + c.FirstResolveNS + c.ProposalNS + c.RequestNS + c.SecondResolveNS + c.SnapshotNS
	c.ElapsedNS = time.Since(started).Nanoseconds()
	a.Breakdown = c
	a.Costs = delayedCostsV36{SetupNS: c.SetupNS, ScheduleNS: c.ScheduleNS, IssueNS: c.IssueNS + c.ProposalNS + c.RequestNS, ResolveNS: c.FirstResolveNS + c.SecondResolveNS, SnapshotNS: c.SnapshotNS, AccountedNS: c.AccountedNS, ElapsedNS: c.ElapsedNS}
	return a, nil
}
func sourcesPairedV55(t *testing.T) map[string]string {
	t.Helper()
	b, e := os.ReadFile(os.Getenv("EVENTFRAME_PAIRED_V55_FREEZE"))
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Files map[string]string `json:"files"`
	}
	if e = json.Unmarshal(b, &f); e != nil || len(f.Files) == 0 {
		t.Fatal("paired freeze", e)
	}
	for p, h := range f.Files {
		b, e := os.ReadFile(filepath.Join(rootV34(t), p))
		if e != nil || hashV34(b) != h {
			t.Fatal("paired source", p, e)
		}
	}
	return f.Files
}
func TestPairedV55Experiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_PAIRED_V55_OUT")
	if path == "" {
		t.Skip("explicit isolated output")
	}
	sources := sourcesPairedV55(t)
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	buf := bufio.NewWriter(f)
	enc := json.NewEncoder(buf)
	if e = enc.Encode(manifestV34{Kind: "manifest", Split: "diagnostic", SeedBase: 2026105407, Worlds: 40, Sources: sources}); e != nil {
		t.Fatal(e)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesPairedV55 {
			w := worldPairedV55{Population: makePairedV55(2026105407, g, r, 0)}
			for _, schedule := range schedulesV39 {
				for _, mode := range modesPairedV55 {
					a, e := runPairedV55(w.Population, mode, schedule)
					if e != nil {
						t.Fatal(e)
					}
					scoreWindowV37(w.Population.World, &a.windowArmV37)
					w.Arms = append(w.Arms, a)
				}
			}
			if e = enc.Encode(w); e != nil {
				t.Fatal(e)
			}
			t.Log(w.Population.World.Geometry, w.Population.World.Regime)
		}
	}
	if e = buf.Flush(); e != nil {
		t.Fatal(e)
	}
	if e = f.Sync(); e != nil {
		t.Fatal(e)
	}
}

func closeOptionPairedV55(a, b paired.Option) bool {
	return closeV36(a.Observed, b.Observed) && closeV36(a.Uncertainty, b.Uncertainty) && closeV36(a.Information, b.Information) && closeV36(a.EdgeCut, b.EdgeCut)
}
func independentPairedV55(p populationPairedV55, a armPairedV55) error {
	w := p.World
	r, e := ref.New(w.Base, paired.Config{Strength: 2, Hazard: 1. / 16})
	if e != nil {
		return e
	}
	due, e := scheduleV36(delayedWorldV36{Seed: w.Seed, Outcomes: w.Outcomes}, a.Schedule)
	if e != nil {
		return e
	}
	last := len(due) - 1
	for len(due[last]) == 0 {
		last--
	}
	secondary := make([][]int, len(due))
	firstCount := [16]int{}
	firstPos, secondPos, snapshot, decision, nextRound, peak := 0, 0, 0, 0, 0, 0
	requestedAt := make([]int, 2400)
	originalAuditForecast := make([]float64, 2400)
	rng := rand.New(rand.NewSource(channelSeedPairedV55(w.Seed, "random_policy")))
	for tick := 0; tick <= last; tick++ {
		if tick < 2400 {
			q, o, e := r.Predict(tick % 150)
			if e != nil || !closeV36(q, a.Issued[tick]) || !closeV36(o, a.ObservedIssued[tick]) || a.ExpertIssued[tick] != ([4]float64{}) {
				return fmt.Errorf("paired independent first forecast %d %v", tick, e)
			}
			weights, e := r.NoiseWeights()
			if e != nil {
				return e
			}
			for j, v := range weights {
				if !closeV36(v, a.NoiseIssued[tick][j]) {
					return fmt.Errorf("paired first weights")
				}
			}
			if e = r.Issue(tick % 150); e != nil {
				return e
			}
			peak = max(peak, tick+1+decision*25-firstPos-secondPos)
		}
		for _, k := range due[tick] {
			if firstPos >= len(a.Receipts) {
				return fmt.Errorf("missing paired first receipt")
			}
			receipt := a.Receipts[firstPos]
			if receipt.Member != k%150 || receipt.TrialOrdinal != k/150+1 || receipt.Epoch != 1 || receipt.IssuedAt != int64(k) || receipt.ArrivedAt != int64(tick) || receipt.Forecast != a.Issued[k] || receipt.Useful != w.Outcomes[k/150][k%150] {
				return fmt.Errorf("paired first identity")
			}
			if e = r.Observe(k%150, k/150+1, 1, receipt.Useful); e != nil {
				return e
			}
			firstPos++
			firstCount[k/150]++
		}
		for nextRound < 16 && firstCount[nextRound] == 150 {
			if a.Mode != "no_pair" {
				if decision >= len(a.Decisions) {
					return fmt.Errorf("missing decision")
				}
				d := a.Decisions[decision]
				if d.Round != nextRound || d.At != tick || len(d.Members) != 25 {
					return fmt.Errorf("decision availability/order")
				}
				if a.Mode == "random" {
					want := rng.Perm(150)[:25]
					if len(d.Options) != 0 || !reflect.DeepEqual(want, d.Members) {
						return fmt.Errorf("random selection")
					}
				} else {
					if len(d.Options) != 150 {
						return fmt.Errorf("score frontier truncated")
					}
					for i, o := range d.Options {
						gold, e := r.Options(i, nextRound+1)
						if e != nil {
							return e
						}
						switch a.Mode {
						case "uncertainty":
							gold.Information, gold.EdgeCut = 0, 0
						case "information":
							gold.EdgeCut = 0
						case "falsification":
							gold.Uncertainty, gold.Information = 0, 0
						default:
							return fmt.Errorf("paired policy domain")
						}
						if !closeOptionPairedV55(o, gold) {
							return fmt.Errorf("independent origin option %d", i)
						}
					}
					order := make([]int, 150)
					for i := range order {
						order[i] = i
					}
					sort.SliceStable(order, func(i, j int) bool {
						return optionScorePairedV55(d.Options[order[i]], a.Mode) > optionScorePairedV55(d.Options[order[j]], a.Mode)
					})
					if !reflect.DeepEqual(order[:25], d.Members) {
						return fmt.Errorf("selection before full scoring")
					}
				}
				seen := map[int]bool{}
				for n, i := range d.Members {
					if i < 0 || i >= 150 || seen[i] {
						return fmt.Errorf("selected duplicate/domain")
					}
					seen[i] = true
					k := nextRound*150 + i
					option, e := r.Options(i, nextRound+1)
					if e != nil {
						return e
					}
					originalAuditForecast[k] = option.Observed
					requestedAt[k] = tick
					if e = r.Audit(i, nextRound+1); e != nil {
						return e
					}
					delay, e := pairedDelayV55(p, k, a.Schedule)
					if e != nil {
						return e
					}
					secondary[tick+delay] = append(secondary[tick+delay], k)
					last = max(last, tick+delay)
					peak = max(peak, min(tick+1, 2400)+decision*25+n+1-firstPos-secondPos)
				}
				decision++
			}
			nextRound++
		}
		for _, k := range secondary[tick] {
			if secondPos >= len(a.AuditOutcomes) {
				return fmt.Errorf("missing audit outcome")
			}
			s := a.AuditOutcomes[secondPos]
			value := false
			if p.Available[k] {
				value = p.Second[k]
			}
			if s.Trial != k || s.RequestedAt != requestedAt[k] || s.ArrivedAt != tick || s.Available != p.Available[k] || s.Value != value || !closeV36(s.Forecast, originalAuditForecast[k]) {
				return fmt.Errorf("origin second receipt")
			}
			if s.Available {
				if e = r.Observe(k%150, k/150+1, 2, s.Value); e != nil {
					return e
				}
			}
			secondPos++
		}
		if (tick < 2400 && tick%150 == 149) || tick == last {
			if snapshot >= len(a.Snapshots) || snapshot >= len(a.NoiseSnapshots) {
				return fmt.Errorf("paired snapshot missing")
			}
			s := a.Snapshots[snapshot]
			if s.Tick != tick || s.Arrived != firstPos || s.Pending != min(tick+1, 2400)+decision*25-firstPos-secondPos || len(s.Forecast) != 150 || s.Weights != ([4]float64{}) {
				return fmt.Errorf("paired snapshot counts")
			}
			for i, q := range s.Forecast {
				v, _, e := r.Predict(i)
				if e != nil || !closeV36(q, v) {
					return fmt.Errorf("paired snapshot law")
				}
			}
			weights, e := r.NoiseWeights()
			if e != nil {
				return e
			}
			for j, v := range weights {
				if !closeV36(v, a.NoiseSnapshots[snapshot][j]) {
					return fmt.Errorf("paired snapshot weights")
				}
			}
			if e = independentMetricsV38(w, s); e != nil {
				return e
			}
			snapshot++
		}
	}
	expected := 16
	if a.Mode == "no_pair" {
		expected = 0
	}
	if firstPos != 2400 || secondPos != expected*25 || decision != expected || len(a.Decisions) != expected || len(a.AuditOutcomes) != expected*25 || snapshot != len(a.Snapshots) || snapshot != len(a.NoiseSnapshots) || peak != a.PeakPending || a.ReadyAt != -1 || len(a.Template) != 0 {
		return fmt.Errorf("paired drain/nomination cap")
	}
	return nil
}
func auditPairedArmV55(p populationPairedV55, a armPairedV55) error {
	if len(a.Issued) != 2400 || len(a.ExpertIssued) != 2400 || len(a.Receipts) != 2400 {
		return fmt.Errorf("paired arm shape")
	}
	c := a.Costs
	if c.SetupNS < 0 || c.ScheduleNS < 0 || c.IssueNS < 0 || c.ResolveNS < 0 || c.SnapshotNS < 0 || c.AccountedNS != c.SetupNS+c.IssueNS+c.ResolveNS+c.SnapshotNS || c.ElapsedNS < c.AccountedNS+c.ScheduleNS {
		return fmt.Errorf("paired cost")
	}
	if a.Mode == "full" || a.Mode == "adaptive" {
		if len(a.ObservedIssued) != 0 || len(a.Decisions) != 0 || len(a.AuditOutcomes) != 0 || len(a.NoiseIssued) != 0 || len(a.NoiseSnapshots) != 0 || a.Breakdown != (costsPairedV55{}) {
			return fmt.Errorf("control extras")
		}
		if e := independentOrientationV38(p.World, a.orientationArmV38); e != nil {
			return e
		}
		return auditPairedMetricsV55(p, a)
	}
	if len(a.ObservedIssued) != 2400 || len(a.NoiseIssued) != 2400 {
		return fmt.Errorf("paired observed shape")
	}
	b := a.Breakdown
	if b.SetupNS < 0 || b.ScheduleNS < 0 || b.IssueNS < 0 || b.FirstResolveNS < 0 || b.ProposalNS < 0 || b.RequestNS < 0 || b.SecondResolveNS < 0 || b.SnapshotNS < 0 || b.AccountedNS != b.SetupNS+b.IssueNS+b.FirstResolveNS+b.ProposalNS+b.RequestNS+b.SecondResolveNS+b.SnapshotNS || b.ElapsedNS < b.AccountedNS+b.ScheduleNS || c.SetupNS != b.SetupNS || c.ScheduleNS != b.ScheduleNS || c.IssueNS != b.IssueNS+b.ProposalNS+b.RequestNS || c.ResolveNS != b.FirstResolveNS+b.SecondResolveNS || c.SnapshotNS != b.SnapshotNS || c.AccountedNS != b.AccountedNS || c.ElapsedNS != b.ElapsedNS {
		return fmt.Errorf("paired full operation accounting")
	}
	if e := independentPairedV55(p, a); e != nil {
		return e
	}
	return auditPairedMetricsV55(p, a)
}
func auditPairedMetricsV55(p populationPairedV55, a armPairedV55) error {
	risk, priority, recovery := 0., 0., 0.
	for k, q := range a.Issued {
		rate := p.World.Rates[k/150][k%150]
		weight := 1.
		if k%150 < 10 {
			weight = 3
		}
		loss := q*q - 2*q*rate + rate
		risk += loss / 2400
		priority += weight * loss / (16 * 170)
	}
	for j, start := range p.World.Changes {
		end := 16
		if j+1 < len(p.World.Changes) {
			end = p.World.Changes[j+1]
		}
		delay := end - start + 1
		for round := start + 1; round < end; round++ {
			one, two := windowAtV37(a.windowArmV37, (round-1)*150+149), windowAtV37(a.windowArmV37, round*150+149)
			if one.Brier <= .20 && two.Brier <= .20 && one.PacketUsefulness >= .75 && two.PacketUsefulness >= .75 {
				delay = round - start + 1
				break
			}
		}
		recovery += float64(delay) / float64(len(p.World.Changes))
	}
	if !closeV36(risk, a.IssuedBrier) || !closeV36(priority, a.IssuedPriority) || !closeV36(recovery, a.Recovery) {
		return fmt.Errorf("paired derived metrics")
	}
	return nil
}
func TestPairedV55Audit(t *testing.T) {
	path := os.Getenv("EVENTFRAME_PAIRED_V55_AUDIT")
	if path == "" {
		t.Skip("explicit paired artifact")
	}
	sources := sourcesPairedV55(t)
	f, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	d := json.NewDecoder(bufio.NewReader(f))
	d.DisallowUnknownFields()
	var m manifestV34
	if e = d.Decode(&m); e != nil || m.Kind != "manifest" || m.Split != "diagnostic" || m.SeedBase != 2026105407 || m.Worlds != 40 || !reflect.DeepEqual(m.Sources, sources) {
		t.Fatal("paired manifest", e)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesPairedV55 {
			var w worldPairedV55
			if e = d.Decode(&w); e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(w.Population, makePairedV55(2026105407, g, r, 0)) || len(w.Arms) != 21 {
				t.Fatal("paired population")
			}
			for j, a := range w.Arms {
				if a.Mode != modesPairedV55[j%7] || a.Schedule != schedulesV39[j/7] {
					t.Fatal("paired arm order")
				}
				if e = auditPairedArmV55(w.Population, a); e != nil {
					t.Fatal(g, r, j, e)
				}
			}
			t.Log(w.Population.World.Geometry, w.Population.World.Regime)
		}
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		t.Fatal("paired trailing artifact", e)
	}
}
func TestPairedV55FutureAndCorruptions(t *testing.T) {
	for _, mode := range modesPairedV55[2:] {
		for _, schedule := range schedulesV39 {
			p, x := makePairedV55(740541, 1, 18, 0), makePairedV55(740541, 1, 18, 0)
			for k := 1200; k < 2400; k++ {
				x.World.Outcomes[k/150][k%150] = !x.World.Outcomes[k/150][k%150]
				x.Second[k] = !x.Second[k]
				x.Truth[k] = !x.Truth[k]
			}
			a, e := runPairedV55(p, mode, schedule)
			if e != nil {
				t.Fatal(e)
			}
			b, e := runPairedV55(x, mode, schedule)
			if e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(a.Issued[:1201], b.Issued[:1201]) || !reflect.DeepEqual(a.ObservedIssued[:1201], b.ObservedIssued[:1201]) || !reflect.DeepEqual(a.NoiseIssued[:1201], b.NoiseIssued[:1201]) {
				t.Fatal("paired future leak")
			}
			if reflect.DeepEqual(a.Issued[1500:], b.Issued[1500:]) {
				t.Fatal("paired revealed fork vacuous")
			}
			for j, d := range a.Decisions {
				if d.At < 1200 && !reflect.DeepEqual(d, b.Decisions[j]) {
					t.Fatal("decision future leak")
				}
			}
			for j, s := range a.Snapshots {
				if s.Tick < 1200 && (!reflect.DeepEqual(s, b.Snapshots[j]) || a.NoiseSnapshots[j] != b.NoiseSnapshots[j]) {
					t.Fatal("snapshot future leak")
				}
			}
			if schedule != "uniform299" || mode == "no_pair" {
				continue
			}
			scoreWindowV37(p.World, &a.windowArmV37)
			if e = auditPairedArmV55(p, a); e != nil {
				t.Fatal(e)
			}
			edits := []func(*armPairedV55){func(z *armPairedV55) { z.Issued[1] += .01 }, func(z *armPairedV55) { z.NoiseIssued[1][0] += .01 }, func(z *armPairedV55) { z.Decisions[0].Members[0] = z.Decisions[0].Members[1] }, func(z *armPairedV55) { z.AuditOutcomes[0].Forecast += .01 }, func(z *armPairedV55) { z.AuditOutcomes[0].RequestedAt++ }, func(z *armPairedV55) { z.AuditOutcomes[0].Available = !z.AuditOutcomes[0].Available }, func(z *armPairedV55) { z.Snapshots = z.Snapshots[:len(z.Snapshots)-1] }, func(z *armPairedV55) { z.Breakdown.ProposalNS++ }}
			for _, edit := range edits {
				data, e := json.Marshal(a)
				if e != nil {
					t.Fatal(e)
				}
				var bad armPairedV55
				if e = json.Unmarshal(data, &bad); e != nil {
					t.Fatal(e)
				}
				edit(&bad)
				if e = auditPairedArmV55(p, bad); e == nil {
					t.Fatal("paired corruption accepted")
				}
			}
		}
	}
}
func TestPairedV55SeedSeparation(t *testing.T) {
	seen := map[int64]bool{}
	channels := map[int64]bool{}
	for _, seed := range []int64{2026105307, 2026105309, 2026105311, 2026105407, 2026105409, 2026105411} {
		for g := 0; g < 2; g++ {
			for r := range regimesPairedV55 {
				for id := 0; id < 16; id++ {
					s := makePairedV55(seed, g, r, id).World.Seed
					if seen[s] {
						t.Fatal("paired seed collision", s)
					}
					seen[s] = true
					for _, name := range []string{"truth", "first", "second", "availability", "secondary_lag", "random_policy"} {
						q := channelSeedPairedV55(s, name)
						if channels[q] {
							t.Fatal("RNG channel collision", q)
						}
						channels[q] = true
					}
				}
			}
		}
	}
	t.Log("distinct actual seeds", len(seen))
	t.Log("distinct domain-separated channels", len(channels))
}
func TestPairedV55Allocation(t *testing.T) {
	path := os.Getenv("EVENTFRAME_PAIRED_V55_ALLOCATION")
	if path == "" {
		t.Skip("explicit allocation")
	}
	base := makePairedV55(2026105407, 0, 0, 0).World.Base
	var maximum uint64
	for i := 0; i < 5; i++ {
		runtime.GC()
		var a, b runtime.MemStats
		runtime.ReadMemStats(&a)
		m, e := paired.New(base, 1, 2800, paired.Config{Strength: 2, Hazard: 1. / 16})
		if e != nil {
			t.Fatal(e)
		}
		runtime.ReadMemStats(&b)
		runtime.KeepAlive(m)
		maximum = max(maximum, b.TotalAlloc-a.TotalAlloc)
	}
	b, e := json.MarshalIndent(map[string]uint64{"modelAllocatedBytes": maximum}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if _, e = f.Write(append(b, '\n')); e != nil {
		t.Fatal(e)
	}
	if e = f.Sync(); e != nil {
		t.Fatal(e)
	}
}

func TestPairedV55ControlMetricGuard(t *testing.T) {
	p := makePairedV55(740549, 1, 8, 0)
	for _, mode := range []string{"full", "adaptive"} {
		a, e := runPairedV55(p, mode, "immediate")
		if e != nil {
			t.Fatal(e)
		}
		scoreWindowV37(p.World, &a.windowArmV37)
		if e = auditPairedArmV55(p, a); e != nil {
			t.Fatal(e)
		}
		for _, which := range []int{0, 1, 2} {
			bad := a
			switch which {
			case 0:
				bad.IssuedBrier += .01
			case 1:
				bad.IssuedPriority += .01
			case 2:
				bad.Recovery++
			}
			if e = auditPairedArmV55(p, bad); e == nil {
				t.Fatal("control aggregate corruption accepted", mode, which)
			}
		}
	}
}
