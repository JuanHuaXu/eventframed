package researchdispersion

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// Recompute from the complete as-of member history, not the learner's cached
// suffix. Log-space prior construction and unnormalized rank-one transitions
// provide an independent check on its normalization and delayed chronology.
func batchContinuingV39(base, hazard float64, issued int, arrived, labels uint64) float64 {
	var prior, mass [21]float64
	sum := 0.
	for z := range prior {
		q := (float64(z) + .5) / 21
		prior[z] = math.Exp((4*base-1)*math.Log(q) + (4*(1-base)-1)*math.Log1p(-q))
		sum += prior[z]
	}
	for z := range prior {
		prior[z] /= sum
		mass[z] = prior[z]
	}
	for j := 0; j < issued; j++ {
		if j > 0 {
			sum = 0
			for _, p := range mass {
				sum += p
			}
			for z := range mass {
				mass[z] = (1-hazard)*mass[z] + hazard*prior[z]*sum
			}
		}
		if arrived&(1<<j) != 0 {
			for z := range mass {
				q := (float64(z) + .5) / 21
				if labels&(1<<j) == 0 {
					q = 1 - q
				}
				mass[z] *= q
			}
		}
	}
	sum, mean, initial := 0., 0., 0.
	for z, p := range mass {
		sum += p
		mean += p * (float64(z) + .5) / 21
		initial += prior[z] * (float64(z) + .5) / 21
	}
	if issued == 0 {
		return initial
	}
	return (1-hazard)*mean/sum + hazard*initial
}

func independentContinuingV39(w windowWorldV37, arm orientationArmV38) error {
	a := arm.windowArmV37
	h, err := hazardV39(a.Mode)
	if err != nil {
		return err
	}
	if arm.ReadyAt != -1 || len(arm.Template) != 0 {
		return fmt.Errorf("local mode claims frozen template")
	}
	due, _ := scheduleV36(delayedWorldV36{Seed: w.Seed, Outcomes: w.Outcomes}, a.Schedule)
	var issued [150]int
	var arrived, labels [150]uint64
	position, snapshot, peak := 0, 0, 0
	last := len(due) - 1
	for len(due[last]) == 0 {
		last--
	}
	for tick := 0; tick <= last; tick++ {
		if tick < 2400 {
			i := tick % 150
			q := batchContinuingV39(w.Base[i], h, issued[i], arrived[i], labels[i])
			if !closeV36(q, a.Issued[tick]) || a.ExpertIssued[tick] != ([4]float64{}) {
				return fmt.Errorf("independent local original law %d", tick)
			}
			issued[i]++
			peak = max(peak, tick+1-position)
		}
		for _, trial := range due[tick] {
			if position >= len(a.Receipts) {
				return fmt.Errorf("missing local receipt")
			}
			i, ordinal := trial%150, trial/150
			r := a.Receipts[position]
			if r.Member != i || r.TrialOrdinal != ordinal+1 || r.Epoch != 1 || r.IssuedAt != int64(trial) || r.ArrivedAt != int64(tick) || r.Useful != w.Outcomes[ordinal][i] || r.Forecast != a.Issued[trial] || arrived[i]&(1<<ordinal) != 0 {
				return fmt.Errorf("local receipt identity/original law")
			}
			arrived[i] |= 1 << ordinal
			if r.Useful {
				labels[i] |= 1 << ordinal
			}
			position++
		}
		if (tick < 2400 && tick%150 == 149) || tick == last {
			if snapshot >= len(a.Snapshots) {
				return fmt.Errorf("local snapshot missing")
			}
			s := a.Snapshots[snapshot]
			if s.Tick != tick || s.Arrived != position || s.Pending != min(tick+1, 2400)-position || len(s.Forecast) != 150 || s.Weights != ([4]float64{}) {
				return fmt.Errorf("local snapshot as-of domain")
			}
			for i, q := range s.Forecast {
				if !closeV36(q, batchContinuingV39(w.Base[i], h, issued[i], arrived[i], labels[i])) {
					return fmt.Errorf("independent local batch law")
				}
			}
			if err := independentMetricsV38(w, s); err != nil {
				return err
			}
			snapshot++
		}
	}
	if position != 2400 || snapshot != len(a.Snapshots) || peak != a.PeakPending {
		return fmt.Errorf("local full drain/nomination coverage")
	}
	return nil
}

func auditContinuingV39(w orientationWorldV38, seed int64) error {
	g, r := -1, -1
	for i, name := range []string{"tight", "wide"} {
		if w.Population.Geometry == name {
			g = i
		}
	}
	for i, name := range regimesV39 {
		if w.Population.Regime == name {
			r = i
		}
	}
	if g < 0 || r < 0 || w.Population.World < 0 || w.Population.World >= 16 || len(w.Arms) != 21 || !reflect.DeepEqual(w.Population, makeV39(seed, g, r, w.Population.World)) {
		return fmt.Errorf("population/RNG/domain")
	}
	for j, arm := range w.Arms {
		a := arm.windowArmV37
		if len(a.Issued) != 2400 || len(a.ExpertIssued) != 2400 || a.Mode != modesV39[j%7] || a.Schedule != schedulesV39[j/7] {
			return fmt.Errorf("arm shape/order")
		}
		c := a.Costs
		if c.SetupNS < 0 || c.ScheduleNS < 0 || c.IssueNS < 0 || c.ResolveNS < 0 || c.SnapshotNS < 0 || c.AccountedNS != c.SetupNS+c.IssueNS+c.ResolveNS+c.SnapshotNS || c.ElapsedNS < c.AccountedNS+c.ScheduleNS {
			return fmt.Errorf("time accounting")
		}
		// Exact rerun catches serialized auxiliary rows; it is not the independent
		// model check below and its timings never replace the collected timings.
		replay, err := runContinuingV39(w.Population, a.Mode, a.Schedule)
		if err != nil {
			return err
		}
		x, y := arm, replay
		x.Costs, y.Costs = delayedCostsV36{}, delayedCostsV36{}
		scoreWindowV37(w.Population, &y.windowArmV37)
		if !reflect.DeepEqual(x, y) {
			return fmt.Errorf("complete arithmetic replay")
		}
		if j%7 < 4 {
			if err := independentOrientationV38(w.Population, arm); err != nil {
				return err
			}
		} else if err := independentContinuingV39(w.Population, arm); err != nil {
			return err
		}
		risk, priority := 0., 0.
		for trial, q := range a.Issued {
			p := w.Population.Rates[trial/150][trial%150]
			weight := 1.
			if trial%150 < 10 {
				weight = 3
			}
			loss := q*q - 2*q*p + p
			risk += loss / 2400
			priority += weight * loss / (16 * 170)
		}
		if !closeV36(risk, a.IssuedBrier) || !closeV36(priority, a.IssuedPriority) {
			return fmt.Errorf("independent issued risk")
		}
		recovery := 0.
		for phase, start := range w.Population.Changes {
			end := 16
			if phase+1 < len(w.Population.Changes) {
				end = w.Population.Changes[phase+1]
			}
			delay := end - start + 1
			for round := start + 1; round < end; round++ {
				one, two := windowAtV37(a, (round-1)*150+149), windowAtV37(a, round*150+149)
				if one.Brier <= .20 && two.Brier <= .20 && one.PacketUsefulness >= .75 && two.PacketUsefulness >= .75 {
					delay = round - start + 1
					break
				}
			}
			recovery += float64(delay) / float64(len(w.Population.Changes))
		}
		if !closeV36(recovery, a.Recovery) {
			return fmt.Errorf("independent phase recovery")
		}
	}
	return nil
}

func sourcesContinuingV39(root string, hashes map[string]string) error {
	if len(hashes) != len(filesV39) {
		return fmt.Errorf("source count")
	}
	for _, p := range filesV39 {
		b, err := os.ReadFile(filepath.Join(root, p))
		if err != nil || hashV34(b) != hashes[p] {
			return fmt.Errorf("frozen source %s: %v", p, err)
		}
	}
	return nil
}

func allocationContinuingV39(b []byte) (int64, error) {
	maxBytes, count := int64(0), 0
	for _, line := range strings.Split(string(b), "\n") {
		v := strings.Fields(line)
		if len(v) == 0 || !strings.HasPrefix(v[0], "BenchmarkContinuingNew150-") {
			continue
		}
		if len(v) != 8 || v[5] != "B/op" {
			return 0, fmt.Errorf("allocation row")
		}
		n, err := strconv.ParseInt(v[4], 10, 64)
		if err != nil || n < 0 {
			return 0, fmt.Errorf("allocation value")
		}
		maxBytes, count = max(maxBytes, n), count+1
	}
	if count != 3 {
		return 0, fmt.Errorf("allocation coverage %d", count)
	}
	return maxBytes, nil
}

func TestStudyAuditV39(t *testing.T) {
	input := os.Getenv("EVENTFRAME_CONTINUING_V39_AUDIT")
	if input == "" {
		t.Skip("explicit audit input required")
	}
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	hash := sha256.New()
	d := json.NewDecoder(bufio.NewReader(io.TeeReader(f, hash)))
	d.DisallowUnknownFields()
	var manifest manifestV34
	if err := d.Decode(&manifest); err != nil {
		t.Fatal(err)
	}
	seed := int64(2026103903)
	if manifest.Split == "confirmation" {
		seed++
	} else if manifest.Split != "design" {
		t.Fatal("split")
	}
	if manifest.Kind != "manifest" || manifest.Worlds != 448 || manifest.SeedBase != seed {
		t.Fatal("manifest")
	}
	if err := sourcesContinuingV39(rootV34(t), manifest.Sources); err != nil {
		t.Fatal(err)
	}
	groups := map[string][]orientationWorldV38{}
	seen := map[int64]bool{}
	count, snapshots := 0, 0
	for {
		var w orientationWorldV38
		err := d.Decode(&w)
		if err == io.EOF {
			break
		}
		if err != nil || seen[w.Population.Seed] {
			t.Fatal("decode/duplicate", err)
		}
		if err := auditContinuingV39(w, seed); err != nil {
			t.Fatal(w.Population.Geometry, w.Population.Regime, w.Population.World, err)
		}
		seen[w.Population.Seed] = true
		count++
		for _, a := range w.Arms {
			snapshots += len(a.Snapshots)
		}
		key := w.Population.Geometry + "/" + w.Population.Regime
		groups[key] = append(groups[key], w)
	}
	if count != 448 || len(groups) != 28 || snapshots != 448*350 {
		t.Fatal("coverage", count, len(groups), snapshots)
	}
	benchmarkPath := os.Getenv("EVENTFRAME_CONTINUING_V39_BENCHMARK")
	if benchmarkPath == "" {
		benchmarkPath = filepath.Join(rootV34(t), "research/continuing-v39/benchmarks.log")
	}
	benchmark, err := os.ReadFile(benchmarkPath)
	if err != nil {
		t.Fatal(err)
	}
	allocation, err := allocationContinuingV39(benchmark)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]func(windowArmV37) float64{
		"IssuedBrier": func(a windowArmV37) float64 { return a.IssuedBrier }, "IssuedPriority": func(a windowArmV37) float64 { return a.IssuedPriority }, "Recovery": func(a windowArmV37) float64 { return a.Recovery },
		"FinalBrier": func(a windowArmV37) float64 { return windowAtV37(a, -1).Brier }, "FinalPriority": func(a windowArmV37) float64 { return windowAtV37(a, -1).PriorityBrier }, "FinalUsefulness": func(a windowArmV37) float64 { return windowAtV37(a, -1).PacketUsefulness },
	}
	stats, gates := map[string]any{}, map[string]any{}
	failures := map[string]int{"local0": 0, "local32": 0, "local16": 0}
	for name, worlds := range groups {
		if len(worlds) != 16 {
			t.Fatal("cell coverage")
		}
		for scheduleIndex, schedule := range schedulesV39 {
			key := name + "/" + schedule
			cell, decisions := map[string]any{}, map[string]any{}
			maximum := int64(0)
			for j, mode := range modesV39 {
				row := map[string]any{}
				cost := int64(0)
				for field, metric := range fields {
					x := make([]float64, 16)
					for k, w := range worlds {
						a := w.Arms[scheduleIndex*7+j]
						x[k] = metric(a.windowArmV37)
						cost = max(cost, a.Costs.ElapsedNS)
					}
					row[field] = interval34(x)
				}
				row["MaximumElapsedNS"] = cost
				maximum = max(maximum, cost)
				cell[mode] = row
			}
			stationary := false
			for _, r := range append(append([]string(nil), regimesV39[:4]...), "stationary_noise10") {
				stationary = stationary || strings.HasSuffix(name, "/"+r)
			}
			for j := 4; j < 7; j++ {
				pass := maximum <= 400000000 && allocation <= 8<<20
				checks := map[string]any{"WorkPass": maximum <= 400000000, "AllocationPass": allocation <= 8<<20}
				for control := 0; control <= 1; control++ {
					for _, field := range []string{"IssuedBrier", "IssuedPriority", "FinalBrier", "FinalPriority", "FinalUsefulness"} {
						x := make([]float64, 16)
						for k, w := range worlds {
							a, b := w.Arms[scheduleIndex*7+j], w.Arms[scheduleIndex*7+control]
							x[k] = fields[field](b.windowArmV37) - fields[field](a.windowArmV37)
							if field == "FinalUsefulness" {
								x[k] = -x[k]
							}
						}
						v := interval34(x)
						improve := control == 0 && !stationary && (field == "IssuedBrier" || field == "IssuedPriority")
						ok := v.Lower >= -.01
						if improve {
							ok = v.Mean >= .01 && v.Lower > 0
						}
						pass = pass && ok
						checks[modesV39[control]+"/"+field] = map[string]any{"Gain": v, "Require001Improvement": improve, "Pass": ok}
					}
				}
				if len(worlds[0].Population.Changes) > 0 {
					x, baseline := make([]float64, 16), make([]float64, 16)
					for k, w := range worlds {
						a, b := w.Arms[scheduleIndex*7+j], w.Arms[scheduleIndex*7]
						x[k], baseline[k] = b.Recovery-a.Recovery, b.Recovery
					}
					v, control := interval34(x), interval34(baseline)
					ok := v.Lower > 0 && v.Mean >= .1*control.Mean
					pass = pass && ok
					checks["Recovery"] = map[string]any{"Gain": v, "Control": control, "Pass": ok}
				}
				if !pass {
					failures[modesV39[j]]++
				}
				decisions[modesV39[j]] = map[string]any{"Checks": checks, "Pass": pass}
			}
			stats[key], gates[key] = cell, decisions
		}
	}
	result := map[string]any{"Study": "continuing-v39", "Split": manifest.Split, "SHA256": hex.EncodeToString(hash.Sum(nil)), "Sources": manifest.Sources, "BenchmarkSHA256": hashV34(benchmark), "Worlds": count, "Arms": count * 21, "Snapshots": snapshots, "DistinctTrials": count * 2400, "FailedCells": failures, "CellsPerCandidate": 84, "MaximumConstructorBytes": allocation, "WholeGoalsComplete": false, "Goal7EqualTotalCost": false, "Groups": stats, "Gates": gates}
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSuffix(input, ".jsonl") + "-audit.json"
	out, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	if _, err = out.Write(append(b, '\n')); err != nil {
		t.Fatal(err)
	}
	if err = out.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Logf("verified %d worlds/%d snapshots; failed cells %v /84", count, snapshots, failures)
}

func TestContinuingAuditorNegativeV39(t *testing.T) {
	seed := int64(1233999) // Technical fixture, never a normal cohort seed.
	w := orientationWorldV38{Population: makeV39(seed, 1, 13, 0)}
	for _, schedule := range schedulesV39 {
		for _, mode := range modesV39 {
			a, err := runContinuingV39(w.Population, mode, schedule)
			if err != nil {
				t.Fatal(err)
			}
			scoreWindowV37(w.Population, &a.windowArmV37)
			w.Arms = append(w.Arms, a)
		}
	}
	if err := auditContinuingV39(w, seed); err != nil {
		t.Fatal(err)
	}
	mutations := []func(*orientationWorldV38){
		func(x *orientationWorldV38) { x.Population.Rates[0][0] += .01 },
		func(x *orientationWorldV38) { x.Arms[4].Issued[1000] += .01 },
		func(x *orientationWorldV38) { x.Arms[5].Receipts[1000].TrialOrdinal++ },
		func(x *orientationWorldV38) { x.Arms[6].Snapshots[8].Forecast[0] += .01 },
		func(x *orientationWorldV38) { x.Arms[4].ExpertIssued[0][0] = .5 },
		func(x *orientationWorldV38) { x.Arms[5].IssuedPriority += .01 },
		func(x *orientationWorldV38) { x.Arms[6].Recovery++ },
		func(x *orientationWorldV38) { x.Arms[4].Costs.AccountedNS++ },
		func(x *orientationWorldV38) { x.Arms[5].PeakPending++ },
		func(x *orientationWorldV38) { x.Arms[6].ReadyAt = 100 },
		func(x *orientationWorldV38) { x.Arms[5].Schedule = "fixed150" },
		func(x *orientationWorldV38) { x.Arms[6].Receipts[1] = x.Arms[6].Receipts[0] },
	}
	raw, _ := json.Marshal(w)
	for i, mutate := range mutations {
		var x orientationWorldV38
		if err := json.Unmarshal(raw, &x); err != nil {
			t.Fatal(err)
		}
		mutate(&x)
		if reflect.DeepEqual(w, x) {
			t.Fatalf("corruption %d is a no-op", i)
		}
		if err := auditContinuingV39(x, seed); err == nil {
			t.Fatalf("corruption %d accepted", i)
		}
	}
	if _, err := allocationContinuingV39([]byte("incomplete")); err == nil {
		t.Fatal("missing allocation accepted")
	}
	t.Log("12 semantic corruptions and missing-allocation control rejected")
}
