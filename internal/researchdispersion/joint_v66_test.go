// Isolated coherent sequence study; no production or truth certification.
package researchdispersion

import (
	"bufio"
	"encoding/json"
	"fmt"
	joint "github.com/JuanHuaXu/eventframed/internal/researchjointsequence"
	ref "github.com/JuanHuaXu/eventframed/internal/researchjointsequenceref"
	paired "github.com/JuanHuaXu/eventframed/internal/researchpaired"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

func chooseJointV66(m *joint.Model, tickets []joint.Ticket, round, at int, mode string, rng *rand.Rand) (decisionPairedV60, error) {
	d := decisionPairedV60{Round: round, At: at}
	if mode == "random" {
		d.Members = append([]int(nil), rng.Perm(150)[:25]...)
		return d, nil
	}
	if mode != "uncertainty" && mode != "information" && mode != "falsification" && mode != "predictive" {
		return d, fmt.Errorf("joint policy")
	}
	d.Options = make([]paired.Option, 150)
	if mode == "predictive" {
		d.Values = make([]float64, 150)
	}
	order := make([]int, 150)
	scores := make([]float64, 150)
	for i := range order {
		order[i] = i
		o, e := m.Query(tickets[round*150+i], mode)
		if e != nil {
			return d, e
		}
		d.Options[i] = paired.Option{Observed: o.Observed, Uncertainty: o.Uncertainty, Information: o.Information, EdgeCut: o.EdgeCut}
		switch mode {
		case "uncertainty":
			scores[i] = o.Uncertainty
		case "information":
			scores[i] = o.Information
		case "falsification":
			scores[i] = o.EdgeCut
		case "predictive":
			scores[i] = o.Value
			d.Values[i] = o.Value
		}
	}
	sort.SliceStable(order, func(i, j int) bool { return scores[order[i]] > scores[order[j]] })
	d.Members = append([]int(nil), order[:25]...)
	return d, nil
}

func runJointV66(p populationPairedV60, mode, schedule string) (armPairedV60, error) {
	w := p.World
	if mode == "full" || mode == "adaptive" {
		a, e := runMomentV41(w, mode, schedule)
		return armPairedV60{orientationArmV38: a}, e
	}
	started := time.Now()
	a := armPairedV60{orientationArmV38: orientationArmV38{ReadyAt: -1, windowArmV37: windowArmV37{Mode: mode, Schedule: schedule, Issued: make([]float64, 2400), ExpertIssued: make([][4]float64, 2400)}}, ObservedIssued: make([]float64, 2400)}
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
	m, e := joint.New(w.Base, 1, 2800, joint.Config{Hazard: 1. / 16})
	a.Breakdown.SetupNS = time.Since(phase).Nanoseconds()
	if e != nil {
		return a, e
	}
	tickets := make([]joint.Ticket, 2400)
	audits := make([]joint.Ticket, 2400)
	requestedAt := make([]int, 2400)
	rng := rand.New(rand.NewSource(channelSeedPairedV60(w.Seed, "random_policy")))
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
			// Both clean Y and W1 laws are recorded BEFORE the issue expiry step.
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
				d, e := chooseJointV66(m, tickets, nextRound, tick, mode, rng)
				a.Breakdown.ProposalNS += time.Since(phase).Nanoseconds()
				if e != nil {
					return a, e
				}
				a.Decisions = append(a.Decisions, d)
				for _, i := range d.Members {
					k := nextRound*150 + i
					phase = time.Now()
					audits[k], e = m.RequestSecond(tickets[k], int64(tick))
					a.Breakdown.RequestNS += time.Since(phase).Nanoseconds()
					if e != nil {
						return a, e
					}
					requestedAt[k] = tick
					delay, e := pairedDelayV60(p, k, schedule)
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
			a.AuditOutcomes = append(a.AuditOutcomes, auditOutcomePairedV60{Trial: k, RequestedAt: requestedAt[k], ArrivedAt: tick, Forecast: audits[k].Forecast(), Available: p.Available[k], Value: value, Expired: false})
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
			a.Snapshots = append(a.Snapshots, s)
			a.Breakdown.SnapshotNS += time.Since(phase).Nanoseconds()
		}
	}
	c := a.Breakdown
	c.AccountedNS = c.ScheduleNS + c.SetupNS + c.IssueNS + c.FirstResolveNS + c.ProposalNS + c.RequestNS + c.SecondResolveNS + c.SnapshotNS
	c.ElapsedNS = time.Since(started).Nanoseconds()
	a.Breakdown = c
	a.Costs = delayedCostsV36{SetupNS: c.SetupNS, ScheduleNS: c.ScheduleNS, IssueNS: c.IssueNS + c.ProposalNS + c.RequestNS, ResolveNS: c.FirstResolveNS + c.SecondResolveNS, SnapshotNS: c.SnapshotNS, AccountedNS: c.AccountedNS, ElapsedNS: c.ElapsedNS}
	return a, nil
}

var modesJointV66 = []string{"full", "adaptive", "no_pair", "random", "uncertainty", "information", "falsification", "predictive"}

func sourcesJointV66(t *testing.T) map[string]string {
	t.Helper()
	b, e := os.ReadFile(os.Getenv("EVENTFRAME_JOINT_V66_FREEZE"))
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Files map[string]string `json:"files"`
	}
	if e = json.Unmarshal(b, &f); e != nil || len(f.Files) == 0 {
		t.Fatal("freeze", e)
	}
	for p, h := range f.Files {
		b, e := os.ReadFile(filepath.Join(rootV34(t), p))
		if e != nil || hashV34(b) != h {
			t.Fatal("frozen source", p, e)
		}
	}
	return f.Files
}
func TestJointV66Experiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_JOINT_V66_OUT")
	if path == "" {
		t.Skip("explicit isolated output")
	}
	sources := sourcesJointV66(t)
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
		for r := range regimesPairedV60 {
			w := worldPairedV60{Population: makePairedV60(2026105407, g, r, 0)}
			for _, schedule := range schedulesV39 {
				for _, mode := range modesJointV66 {
					a, e := runJointV66(w.Population, mode, schedule)
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

func independentJointV66(p populationPairedV60, a armPairedV60) error {
	if len(a.Issued) != 2400 || len(a.ObservedIssued) != 2400 {
		return fmt.Errorf("shape")
	}
	r, e := ref.New(p.World.Base, 1./16)
	if e != nil {
		return e
	}
	due, e := scheduleV36(delayedWorldV36{Seed: p.World.Seed, Outcomes: p.World.Outcomes}, a.Schedule)
	if e != nil {
		return e
	}
	seconds := make([][]int, len(due))
	firstCount := [16]int{}
	last := len(due) - 1
	for len(due[last]) == 0 {
		last--
	}
	issued, nextRound, decisions, receipts, snapshots, outcomes, requests := 0, 0, 0, 0, 0, 0, 0
	rng := rand.New(rand.NewSource(channelSeedPairedV60(p.World.Seed, "random_policy")))
	requestedAt := make([]int, 2400)
	requestForecast := make([]float64, 2400)
	for tick := 0; tick <= last; tick++ {
		if tick < 2400 {
			q, o, e := r.Predict(tick % 150)
			if e != nil {
				return e
			}
			if !closeV36(q, a.Issued[tick]) || !closeV36(o, a.ObservedIssued[tick]) {
				return fmt.Errorf("issue law %d", tick)
			}
			if e = r.Issue(tick % 150); e != nil {
				return e
			}
			issued++
		}
		for _, k := range due[tick] {
			value := p.World.Outcomes[k/150][k%150]
			if e = r.Observe(k%150, k/150+1, 1, value); e != nil {
				return e
			}
			if receipts >= len(a.Receipts) {
				return fmt.Errorf("receipt shape")
			}
			x := a.Receipts[receipts]
			receipts++
			if x.Member != k%150 || x.TrialOrdinal != k/150+1 || x.Epoch != 1 || x.IssuedAt != int64(k) || x.ArrivedAt != int64(tick) || x.Useful != value || !closeV36(x.Forecast, a.ObservedIssued[k]) {
				return fmt.Errorf("receipt metadata")
			}
			firstCount[k/150]++
		}
		for nextRound < 16 && firstCount[nextRound] == 150 {
			if a.Mode != "no_pair" {
				if decisions >= len(a.Decisions) {
					return fmt.Errorf("decision shape")
				}
				d := a.Decisions[decisions]
				decisions++
				if d.Round != nextRound || d.At != tick || len(d.Members) != 25 {
					return fmt.Errorf("decision metadata")
				}
				var order []int
				if a.Mode == "random" {
					order = rng.Perm(150)[:25]
					if len(d.Options) != 0 || len(d.Values) != 0 {
						return fmt.Errorf("random candidates")
					}
				} else {
					if len(d.Options) != 150 || ((a.Mode == "predictive") != (len(d.Values) == 150)) {
						return fmt.Errorf("option shape")
					}
					order = make([]int, 150)
					scores := make([]float64, 150)
					for i := range order {
						order[i] = i
						o, e := r.Query(i, nextRound+1)
						if e != nil {
							return e
						}
						x := d.Options[i]
						if !closeV36(o.Observed, x.Observed) || !closeV36(o.Uncertainty, x.Uncertainty) {
							return fmt.Errorf("query probability")
						}
						switch a.Mode {
						case "uncertainty":
							if x.Information != 0 || x.EdgeCut != 0 {
								return fmt.Errorf("undeclared uncertainty fields")
							}
							scores[i] = x.Uncertainty
						case "information", "falsification":
							if !closeV36(o.Information, x.Information) || !closeV36(o.EdgeCut, x.EdgeCut) {
								return fmt.Errorf("information law")
							}
							scores[i] = x.Information
							if a.Mode == "falsification" {
								scores[i] = x.EdgeCut
							}
						case "predictive":
							if !closeV36(o.Value, d.Values[i]) || x.Information != 0 || x.EdgeCut != 0 {
								return fmt.Errorf("prediction value")
							}
							scores[i] = d.Values[i]
						default:
							return fmt.Errorf("mode")
						}
					}
					// Exact sorting on independently validated recorded scores avoids
					// demanding identical ULP ordering from different arithmetic.
					sort.SliceStable(order, func(i, j int) bool { return scores[order[i]] > scores[order[j]] })
					order = order[:25]
				}
				if !reflect.DeepEqual(order, d.Members) {
					return fmt.Errorf("selection")
				}
				for _, i := range order {
					k := nextRound*150 + i
					o, e := r.Query(i, nextRound+1)
					if e != nil {
						return e
					}
					requestForecast[k] = o.Observed
					requestedAt[k] = tick
					if e = r.Audit(i, nextRound+1); e != nil {
						return e
					}
					requests++
					delay, e := pairedDelayV60(p, k, a.Schedule)
					if e != nil {
						return e
					}
					seconds[tick+delay] = append(seconds[tick+delay], k)
					last = max(last, tick+delay)
				}
			}
			nextRound++
		}
		for _, k := range seconds[tick] {
			if outcomes >= len(a.AuditOutcomes) {
				return fmt.Errorf("audit shape")
			}
			x := a.AuditOutcomes[outcomes]
			outcomes++
			value := p.Available[k] && p.Second[k]
			if x.Trial != k || x.RequestedAt != requestedAt[k] || x.ArrivedAt != tick || x.Available != p.Available[k] || x.Value != value || x.Expired || !closeV36(x.Forecast, requestForecast[k]) {
				return fmt.Errorf("audit metadata")
			}
			if p.Available[k] {
				if e = r.Observe(k%150, k/150+1, 2, p.Second[k]); e != nil {
					return e
				}
			}
		}
		if (tick < 2400 && tick%150 == 149) || tick == last {
			if snapshots >= len(a.Snapshots) {
				return fmt.Errorf("snapshot shape")
			}
			x := a.Snapshots[snapshots]
			snapshots++
			if x.Tick != tick || x.Arrived != receipts || x.Pending != issued-receipts+requests-outcomes || len(x.Forecast) != 150 {
				return fmt.Errorf("snapshot metadata")
			}
			for i, q := range x.Forecast {
				p, _, e := r.Predict(i)
				if e != nil {
					return e
				}
				if !closeV36(q, p) {
					return fmt.Errorf("snapshot law")
				}
			}
		}
	}
	if receipts != 2400 || nextRound != 16 || snapshots != len(a.Snapshots) || decisions != len(a.Decisions) || outcomes != len(a.AuditOutcomes) {
		return fmt.Errorf("terminal shape")
	}
	if (a.Mode == "no_pair" && outcomes != 0) || (a.Mode != "no_pair" && outcomes != 400) {
		return fmt.Errorf("request budget")
	}
	x := a.windowArmV37
	x.IssuedBrier, x.IssuedPriority, x.Recovery = 0, 0, 0
	x.Snapshots = append([]windowSnapshotV37(nil), a.Snapshots...)
	scoreWindowV37(p.World, &x)
	if !closeV36(x.IssuedBrier, a.IssuedBrier) || !closeV36(x.IssuedPriority, a.IssuedPriority) || !closeV36(x.Recovery, a.Recovery) {
		return fmt.Errorf("metrics")
	}
	for i, s := range x.Snapshots {
		z := a.Snapshots[i]
		if !closeV36(s.Brier, z.Brier) || !closeV36(s.PriorityBrier, z.PriorityBrier) || !closeV36(s.PacketUsefulness, z.PacketUsefulness) || !closeV36(s.PacketBias, z.PacketBias) {
			return fmt.Errorf("snapshot metrics")
		}
	}
	c := a.Breakdown
	if c.AccountedNS != c.ScheduleNS+c.SetupNS+c.IssueNS+c.FirstResolveNS+c.ProposalNS+c.RequestNS+c.SecondResolveNS+c.SnapshotNS || c.AccountedNS > c.ElapsedNS || c.ElapsedNS != a.Costs.ElapsedNS {
		return fmt.Errorf("cost accounting")
	}
	return nil
}

func TestJointV66Audit(t *testing.T) {
	path := os.Getenv("EVENTFRAME_JOINT_V66_AUDIT")
	if path == "" {
		t.Skip("explicit artifact")
	}
	sources := sourcesJointV66(t)
	f, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	d := json.NewDecoder(bufio.NewReader(f))
	var m manifestV34
	if e = d.Decode(&m); e != nil || m.Kind != "manifest" || m.Split != "diagnostic" || m.SeedBase != 2026105407 || m.Worlds != 40 || !reflect.DeepEqual(m.Sources, sources) {
		t.Fatal("manifest", e)
	}
	checks := 0
	for g := 0; g < 2; g++ {
		for r := range regimesPairedV60 {
			var w worldPairedV60
			if e = d.Decode(&w); e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(w.Population, makePairedV60(2026105407, g, r, 0)) || len(w.Arms) != 24 {
				t.Fatal("population/shape")
			}
			for k, a := range w.Arms {
				if a.Mode != modesJointV66[k%8] || a.Schedule != schedulesV39[k/8] {
					t.Fatal("arm order")
				}
				if k%8 < 2 {
					continue
				}
				if e = independentJointV66(w.Population, a); e != nil {
					t.Fatal(g, r, a.Schedule, a.Mode, e)
				}
				checks += 2400
			}
			t.Log("replayed", g, w.Population.World.Regime)
		}
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		t.Fatal("trailing data", e)
	}
	t.Log("independent issued packets (clean and W1)", checks)
}
func TestJointV66FutureAndCorruptions(t *testing.T) {
	p := makePairedV60(2026105407, 0, 14, 0)
	a, e := runJointV66(p, "random", "fixed150")
	if e != nil {
		t.Fatal(e)
	}
	scoreWindowV37(p.World, &a.windowArmV37)
	if e = independentJointV66(p, a); e != nil {
		t.Fatal(e)
	}
	clone := func() armPairedV60 {
		b, _ := json.Marshal(a)
		var x armPairedV60
		if e = json.Unmarshal(b, &x); e != nil {
			t.Fatal(e)
		}
		return x
	}
	for _, field := range []string{"issue", "observed", "receipt", "request", "choice", "snapshot", "metric", "pending", "cost"} {
		x := clone()
		switch field {
		case "issue":
			x.Issued[0] += .1
		case "observed":
			x.ObservedIssued[0] += .1
		case "receipt":
			x.Receipts[0].ArrivedAt++
		case "request":
			x.AuditOutcomes[0].Forecast += .1
		case "choice":
			x.Decisions[0].Members[0] = x.Decisions[0].Members[1]
		case "snapshot":
			x.Snapshots[0].Forecast[0] += .1
		case "metric":
			x.Recovery++
		case "pending":
			x.Snapshots[0].Pending++
		case "cost":
			x.Breakdown.AccountedNS++
		}
		if e = independentJointV66(p, x); e == nil {
			t.Fatal("corruption", field)
		}
	}
	q := p
	q.Second = append([]bool(nil), p.Second...)
	for k := 1800; k < 2400; k++ {
		q.Second[k] = !q.Second[k]
	}
	b, e := runJointV66(q, "random", "fixed150")
	if e != nil {
		t.Fatal(e)
	}
	changed := false
	for k := range a.Issued {
		if k <= 1800 && a.Issued[k] != b.Issued[k] {
			t.Fatal("future leak", k)
		}
		if k > 1800 && a.Issued[k] != b.Issued[k] {
			changed = true
		}
	}
	if !changed {
		t.Fatal("vacuous fork")
	}
	t.Log("9 corruptions rejected and nonvacuous future fork")
}
