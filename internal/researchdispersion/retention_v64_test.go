// Isolated V64 quality ablation: actual working window mixture, consumed worlds.
package researchdispersion

import (
	"bufio"
	"encoding/json"
	"fmt"
	selector "github.com/JuanHuaXu/eventframed/internal/researchretention"
	law "github.com/JuanHuaXu/eventframed/internal/researchretentionlaw"
	bank "github.com/JuanHuaXu/eventframed/internal/researchwindowjournal"
	ref "github.com/JuanHuaXu/eventframed/internal/researchwindowjournalref"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

// The observer receives neither the generator nor future receipt values,
// availability or arrival times. Both stages act on the same issued mixture.
func chooseRetentionV64(m *bank.Bank, tickets []bank.BankTicket, round, at int, mode string, rng *rand.Rand) (decisionPairedV60, error) {
	d := decisionPairedV60{Round: round, At: at}
	if mode == "random" {
		d.Members = append([]int(nil), rng.Perm(150)[:25]...)
		return d, nil
	}
	if mode != "uncertainty" {
		return d, fmt.Errorf("retention policy")
	}
	d.Options = make([]bank.Option, 150)
	order := make([]int, 150)
	for i := range order {
		order[i] = i
		q, e := m.QuerySecond(tickets[round*150+i])
		if e != nil {
			return d, e
		}
		d.Options[i].Observed = q
		if q > 0 && q < 1 {
			d.Options[i].Uncertainty = -q*math.Log(q) - (1-q)*math.Log1p(-q)
		}
	}
	sort.SliceStable(order, func(i, j int) bool { return d.Options[order[i]].Uncertainty > d.Options[order[j]].Uncertainty })
	d.Members = append([]int(nil), order[:25]...)
	return d, nil
}

func runRetentionV64(p populationPairedV60, mode, schedule string) (armPairedV60, error) {
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
	m, e := bank.NewBank(w.Base, 1, 2800, 7, [3]int{600, 1200, 2400})
	a.Breakdown.SetupNS = time.Since(phase).Nanoseconds()
	if e != nil {
		return a, e
	}
	tickets := make([]bank.BankTicket, 2400)
	audits := make([]bank.BankTicket, 2400)
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
				d, e := chooseRetentionV64(m, tickets, nextRound, tick, mode, rng)
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
			a.AuditOutcomes = append(a.AuditOutcomes, auditOutcomePairedV60{Trial: k, RequestedAt: requestedAt[k], ArrivedAt: tick, Forecast: audits[k].Forecast(), Available: p.Available[k], Value: value, Expired: k < min(tick+1, 2400)-600})
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

var modesRetentionV64 = []string{"full", "adaptive", "no_pair", "random", "uncertainty"}

func sourcesRetentionV64(t *testing.T) map[string]string {
	t.Helper()
	b, e := os.ReadFile(os.Getenv("EVENTFRAME_RETENTION_V64_FREEZE"))
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

func TestRetentionV64Experiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_RETENTION_V64_OUT")
	if path == "" {
		t.Skip("explicit isolated output")
	}
	sources := sourcesRetentionV64(t)
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	b := bufio.NewWriter(f)
	enc := json.NewEncoder(b)
	if e = enc.Encode(manifestV34{Kind: "manifest", Split: "diagnostic", SeedBase: 2026105407, Worlds: 40, Sources: sources}); e != nil {
		t.Fatal(e)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesPairedV60 {
			w := worldPairedV60{Population: makePairedV60(2026105407, g, r, 0)}
			for _, schedule := range schedulesV39 {
				for _, mode := range modesRetentionV64 {
					a, e := runRetentionV64(w.Population, mode, schedule)
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
	if e = b.Flush(); e != nil {
		t.Fatal(e)
	}
	if e = f.Sync(); e != nil {
		t.Fatal(e)
	}
}

// Independent tree reconstruction plus the previously frozen finite-HMM
// selector. Replay nominal evidence and choices, never the generator's truth.
func independentRetentionV64(p populationPairedV60, a armPairedV60) error {
	if len(a.Issued) != 2400 || len(a.ObservedIssued) != 2400 {
		return fmt.Errorf("shape")
	}
	var refs [3]*ref.Reference
	windows := [3]int{600, 1200, 2400}
	for k, w := range windows {
		var e error
		refs[k], e = ref.New(p.World.Base, bank.Config{Depth: 7, Window: w})
		if e != nil {
			return e
		}
	}
	s, e := selector.New(150, 1)
	if e != nil {
		return e
	}
	tickets := make([]selector.Ticket, 2400)
	audits := make([]selector.Ticket, 2400)
	active := make([][3]bool, 2400)
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
	predict := func(i int) (float64, float64, selector.Forecasts, error) {
		var joints [3]law.Joint
		for k, r := range refs {
			j, e := r.LatentJoint(i)
			if e != nil {
				return 0, 0, selector.Forecasts{}, e
			}
			joints[k] = j
		}
		f, e := law.Forecasts(joints)
		if e != nil {
			return 0, 0, f, e
		}
		ws, e := s.Weights(i)
		if e != nil {
			return 0, 0, f, e
		}
		q, o := 0., 0.
		for k, w := range ws {
			q += w * f[k].Clean
			o += w * (f[k].Joint[2] + f[k].Joint[3])
		}
		return q, o, f, nil
	}
	for tick := 0; tick <= last; tick++ {
		if tick < 2400 {
			i := tick % 150
			q, o, f, e := predict(i)
			if e != nil {
				return e
			}
			if !closeV36(q, a.Issued[tick]) || !closeV36(o, a.ObservedIssued[tick]) {
				return fmt.Errorf("issued law %d", tick)
			}
			tickets[tick], e = s.Issue(i, int64(tick), f)
			if e != nil {
				return e
			}
			if !closeV36(tickets[tick].Forecast(), a.Issued[tick]) {
				return fmt.Errorf("selector issue")
			}
			for _, r := range refs {
				if e = r.Issue(i); e != nil {
					return e
				}
			}
			issued++
		}
		for _, k := range due[tick] {
			value := p.World.Outcomes[k/150][k%150]
			r, e := s.Resolve(tickets[k], value, int64(tick))
			if e != nil {
				return e
			}
			if receipts >= len(a.Receipts) {
				return fmt.Errorf("receipt shape")
			}
			x := a.Receipts[receipts]
			if x.Member != r.Member || x.TrialOrdinal != r.Ordinal || x.Epoch != r.Epoch || x.IssuedAt != r.IssuedAt || x.ArrivedAt != r.ArrivedAt || x.Useful != r.Value || !closeV36(x.Forecast, r.Forecast) {
				return fmt.Errorf("receipt metadata")
			}
			receipts++
			for _, r := range refs {
				if e = r.Observe(k%150, k/150+1, 1, value); e != nil {
					return e
				}
			}
			firstCount[k/150]++
		}
		for nextRound < 16 && firstCount[nextRound] == 150 {
			if a.Mode != "no_pair" {
				if decisions >= len(a.Decisions) {
					return fmt.Errorf("decision shape")
				}
				d := a.Decisions[decisions]
				if d.Round != nextRound || d.At != tick || len(d.Members) != 25 {
					return fmt.Errorf("decision metadata")
				}
				var order []int
				if a.Mode == "random" {
					order = rng.Perm(150)[:25]
					if len(d.Options) != 0 {
						return fmt.Errorf("random scored candidates")
					}
				} else if a.Mode == "uncertainty" {
					if len(d.Options) != 150 {
						return fmt.Errorf("option shape")
					}
					order = make([]int, 150)
					scores := make([]float64, 150)
					for i := range order {
						order[i] = i
						q, e := s.QuerySecond(tickets[nextRound*150+i])
						if e != nil {
							return e
						}
						h := 0.
						if q > 0 && q < 1 {
							h = -q*math.Log(q) - (1-q)*math.Log1p(-q)
						}
						scores[i] = h
						if !closeV36(q, d.Options[i].Observed) || !closeV36(h, d.Options[i].Uncertainty) {
							return fmt.Errorf("acquisition law")
						}
					}
					sort.SliceStable(order, func(i, j int) bool { return scores[order[i]] > scores[order[j]] })
					order = order[:25]
				} else {
					return fmt.Errorf("mode")
				}
				if !reflect.DeepEqual(order, d.Members) {
					return fmt.Errorf("selection")
				}
				for _, i := range order {
					k := nextRound*150 + i
					audits[k], e = s.RequestSecond(tickets[k], int64(tick))
					if e != nil {
						return e
					}
					requests++
					for h, w := range windows {
						if issued-k <= w {
							if e = refs[h].Audit(i, nextRound+1); e != nil {
								return e
							}
							active[k][h] = true
						}
					}
					delay, e := pairedDelayV60(p, k, a.Schedule)
					if e != nil {
						return e
					}
					seconds[tick+delay] = append(seconds[tick+delay], k)
					last = max(last, tick+delay)
				}
				decisions++
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
			delay, _ := pairedDelayV60(p, k, a.Schedule)
			if x.Trial != k || x.RequestedAt != tick-delay || x.ArrivedAt != tick || x.Available != p.Available[k] || x.Value != value || x.Expired != (k < issued-600) || !closeV36(x.Forecast, audits[k].Forecast()) {
				return fmt.Errorf("audit metadata")
			}
			if p.Available[k] {
				if _, e = s.Resolve(audits[k], p.Second[k], int64(tick)); e != nil {
					return e
				}
				for h, r := range refs {
					if active[k][h] {
						if e = r.Observe(k%150, k/150+1, 2, p.Second[k]); e != nil {
							return e
						}
					}
				}
			} else {
				if e = s.Cancel(audits[k], int64(tick)); e != nil {
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
				r, _, _, e := predict(i)
				if e != nil {
					return e
				}
				if !closeV36(q, r) {
					return fmt.Errorf("snapshot law")
				}
			}
		}
	}
	if receipts != 2400 || nextRound != 16 || snapshots != len(a.Snapshots) || decisions != len(a.Decisions) || outcomes != len(a.AuditOutcomes) {
		return fmt.Errorf("terminal shape")
	}
	if (a.Mode == "no_pair" && outcomes != 0) || (a.Mode != "no_pair" && outcomes != 400) {
		return fmt.Errorf("equal requests")
	}
	copy := a.windowArmV37
	copy.IssuedBrier, copy.IssuedPriority, copy.Recovery = 0, 0, 0
	copy.Snapshots = append([]windowSnapshotV37(nil), a.Snapshots...)
	scoreWindowV37(p.World, &copy)
	if !closeV36(copy.IssuedBrier, a.IssuedBrier) || !closeV36(copy.IssuedPriority, a.IssuedPriority) || !closeV36(copy.Recovery, a.Recovery) {
		return fmt.Errorf("metrics")
	}
	for i, x := range copy.Snapshots {
		z := a.Snapshots[i]
		if !closeV36(x.Brier, z.Brier) || !closeV36(x.PriorityBrier, z.PriorityBrier) || !closeV36(x.PacketUsefulness, z.PacketUsefulness) || !closeV36(x.PacketBias, z.PacketBias) {
			return fmt.Errorf("snapshot metrics")
		}
	}
	c := a.Breakdown
	if c.AccountedNS != c.ScheduleNS+c.SetupNS+c.IssueNS+c.FirstResolveNS+c.ProposalNS+c.RequestNS+c.SecondResolveNS+c.SnapshotNS || c.AccountedNS > c.ElapsedNS || c.ElapsedNS != a.Costs.ElapsedNS {
		return fmt.Errorf("cost accounting")
	}
	return nil
}

func TestRetentionV64Audit(t *testing.T) {
	path := os.Getenv("EVENTFRAME_RETENTION_V64_AUDIT")
	if path == "" {
		t.Skip("explicit artifact")
	}
	sources := sourcesRetentionV64(t)
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
			if !reflect.DeepEqual(w.Population, makePairedV60(2026105407, g, r, 0)) || len(w.Arms) != 15 {
				t.Fatal("population/shape")
			}
			for k, a := range w.Arms {
				if a.Mode != modesRetentionV64[k%5] || a.Schedule != schedulesV39[k/5] {
					t.Fatal("arm order")
				}
				if a.Mode == "full" || a.Mode == "adaptive" {
					continue
				}
				if e = independentRetentionV64(w.Population, a); e != nil {
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
	t.Log("independently reconstructed clean and observed issue forecasts", checks)
}

func TestRetentionV64FutureAndCorruptions(t *testing.T) {
	p := makePairedV60(2026105407, 0, 14, 0)
	a, e := runRetentionV64(p, "random", "fixed150")
	if e != nil {
		t.Fatal(e)
	}
	scoreWindowV37(p.World, &a.windowArmV37)
	if e = independentRetentionV64(p, a); e != nil {
		t.Fatal(e)
	}
	clone := func() armPairedV60 {
		b, _ := json.Marshal(a)
		var out armPairedV60
		if e = json.Unmarshal(b, &out); e != nil {
			t.Fatal(e)
		}
		return out
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
		if e = independentRetentionV64(p, x); e == nil {
			t.Fatal("corruption accepted", field)
		}
	}
	q := p
	q.Second = append([]bool(nil), p.Second...)
	for k := 1800; k < 2400; k++ {
		q.Second[k] = !q.Second[k]
	}
	z, e := runRetentionV64(q, "random", "fixed150")
	if e != nil {
		t.Fatal(e)
	}
	changed := false
	for k := range a.Issued {
		if k <= 1800 && a.Issued[k] != z.Issued[k] {
			t.Fatal("future leaked", k)
		}
		if k > 1800 && a.Issued[k] != z.Issued[k] {
			changed = true
		}
	}
	if !changed {
		t.Fatal("vacuous future fork")
	}
	t.Log("9 corrupted fields rejected; future fork differs only after visible evidence")
}
