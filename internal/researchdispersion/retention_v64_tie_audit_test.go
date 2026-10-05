// Independent replay ordering diagnostic; original failed artifact retained.
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
	"reflect"
	"sort"
	"testing"
)

var retentionOrderingStats struct {
	Decisions, AmbiguousReferenceOrders                  int
	MaxAbsoluteScoreDifference, MaxReferenceCutoffRegret float64
}

func independentRetentionV64TieAudit(p populationPairedV60, a armPairedV60) error {
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
						if d.Options[i].Information != 0 || d.Options[i].EdgeCut != 0 {
							return fmt.Errorf("undeclared acquisition fields")
						}
						retentionOrderingStats.MaxAbsoluteScoreDifference = math.Max(retentionOrderingStats.MaxAbsoluteScoreDifference, math.Abs(h-d.Options[i].Uncertainty))
					}
					sort.SliceStable(order, func(i, j int) bool { return scores[order[i]] > scores[order[j]] })
					order = order[:25]
					retentionOrderingStats.Decisions++
					if !reflect.DeepEqual(order, d.Members) {
						retentionOrderingStats.AmbiguousReferenceOrders++
					}
					// Recompute stable sorting EXACTLY on independently checked recorded
					// scores. Do not demand bitwise ranks from different arithmetic at
					// saturated probabilities; explicitly bound reference-cutoff regret.
					chosen := map[int]bool{}
					minimum := math.Inf(1)
					for _, i := range d.Members {
						if i < 0 || i >= 150 || chosen[i] {
							return fmt.Errorf("selection member")
						}
						chosen[i] = true
						minimum = math.Min(minimum, scores[i])
					}
					regret := 0.
					for i, h := range scores {
						if !chosen[i] {
							regret = math.Max(regret, h-minimum)
						}
					}
					if regret > 6e-10 {
						return fmt.Errorf("reference cutoff error")
					}
					retentionOrderingStats.MaxReferenceCutoffRegret = math.Max(retentionOrderingStats.MaxReferenceCutoffRegret, regret)
					order = make([]int, 150)
					for i := range order {
						order[i] = i
					}
					sort.SliceStable(order, func(i, j int) bool { return d.Options[order[i]].Uncertainty > d.Options[order[j]].Uncertainty })
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

func TestRetentionV64TieAudit(t *testing.T) {
	path := os.Getenv("EVENTFRAME_RETENTION_V64_AUDIT")
	if path == "" {
		t.Skip("explicit artifact")
	}
	old := os.Getenv("EVENTFRAME_RETENTION_V64_FREEZE")
	if e := os.Setenv("EVENTFRAME_RETENTION_V64_FREEZE", os.Getenv("EVENTFRAME_RETENTION_V64_DATA_FREEZE")); e != nil {
		t.Fatal(e)
	}
	sources := sourcesRetentionV64(t)
	if e := os.Setenv("EVENTFRAME_RETENTION_V64_FREEZE", old); e != nil {
		t.Fatal(e)
	}
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
				if e = independentRetentionV64TieAudit(w.Population, a); e != nil {
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
	if path := os.Getenv("EVENTFRAME_RETENTION_V64_TIE_STATS"); path != "" {
		b, e := json.MarshalIndent(retentionOrderingStats, "", "  ")
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
}
