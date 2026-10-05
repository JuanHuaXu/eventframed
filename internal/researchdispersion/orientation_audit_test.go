package researchdispersion

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func sourcesV38(root string, hashes map[string]string) error {
	if len(hashes) != len(filesV38) {
		return fmt.Errorf("source count")
	}
	for _, p := range filesV38 {
		b, err := os.ReadFile(filepath.Join(root, p))
		if err != nil || hashV34(b) != hashes[p] {
			return fmt.Errorf("frozen source %s: %v", p, err)
		}
	}
	return nil
}

func independentOrientationV38(w windowWorldV37, arm orientationArmV38) error {
	a := arm.windowArmV37
	due, _ := scheduleV36(delayedWorldV36{Seed: w.Seed, Outcomes: w.Outcomes}, a.Schedule)
	arrived, labels := make([]uint64, 150), make([]uint64, 150)
	ref, _ := NewShape(w.Base)
	weights := [4]float64{1}
	if a.Mode == "adaptive" {
		weights = windowPrior
	}
	control := a.Mode == "full" || a.Mode == "adaptive"
	anchorCount, position, snapshot, peak := 0, 0, 0, 0
	ready := -1
	var template []float64
	nodes := []orientationNode{}
	indices := make([]int, 2400)
	for i := range indices {
		indices[i] = -1
	}
	hazard := 0.
	if a.Mode == "orientation" {
		hazard = 1. / 600
	}
	last := 4800
	for len(due[last]) == 0 {
		last--
	}
	for tick := 0; tick <= last; tick++ {
		if tick < 2400 {
			i := tick % 150
			peak = max(peak, tick+1-position)
			q := 0.
			if control {
				for j, p := range a.ExpertIssued[tick] {
					q += weights[j] * p
				}
			} else {
				if a.ExpertIssued[tick] != ([4]float64{}) {
					return fmt.Errorf("orientation falsely serializes expert rows")
				}
				if ready < 0 {
					q, _ = ref.Predict(i)
				} else {
					f := orientationReference(template, hazard, nodes)
					next := (1-hazard)*f + hazard*(1-f)
					q = template[i]*(1-next) + (1-template[i])*next
					indices[tick] = len(nodes)
					nodes = append(nodes, orientationNode{member: i})
				}
			}
			if !closeV36(q, a.Issued[tick]) {
				return fmt.Errorf("independent original issue law %d", tick)
			}
		}
		for _, trial := range due[tick] {
			if position >= len(a.Receipts) {
				return fmt.Errorf("missing receipt")
			}
			i, ordinal := trial%150, trial/150
			r := a.Receipts[position]
			if r.Member != i || r.TrialOrdinal != ordinal+1 || r.Epoch != 1 || r.IssuedAt != int64(trial) || r.ArrivedAt != int64(tick) || r.Useful != w.Outcomes[ordinal][i] || r.Forecast != a.Issued[trial] || arrived[i]&(1<<ordinal) != 0 {
				return fmt.Errorf("receipt identity/original law")
			}
			arrived[i] |= 1 << ordinal
			if r.Useful {
				labels[i] |= 1 << ordinal
			}
			if control && a.Mode == "adaptive" {
				sum := 0.
				for j, p := range a.ExpertIssued[trial] {
					if !r.Useful {
						p = 1 - p
					}
					weights[j] *= p
					sum += weights[j]
				}
				for j := range weights {
					weights[j] = weights[j]/sum*(599./600) + windowPrior[j]/600
				}
			} else if !control {
				if ordinal < 4 {
					if err := ref.Observe(i, int(ref.n[i])+1, r.Useful); err != nil {
						return err
					}
					anchorCount++
					if anchorCount == 600 {
						ready = tick
						_, template = shapeReference(ref)
						for issued := 600; issued <= min(tick, 2399); issued++ {
							member, round := issued%150, issued/150
							indices[issued] = len(nodes)
							bit := uint64(1) << round
							nodes = append(nodes, orientationNode{member: member, observed: arrived[member]&bit != 0, useful: labels[member]&bit != 0})
						}
					}
				} else if ready >= 0 {
					index := indices[trial]
					if index < 0 || index >= len(nodes) || nodes[index].observed {
						return fmt.Errorf("independent nomination identity")
					}
					nodes[index].observed, nodes[index].useful = true, r.Useful
				}
			}
			position++
		}
		if (tick < 2400 && tick%150 == 149) || tick == last {
			if snapshot >= len(a.Snapshots) {
				return fmt.Errorf("snapshot missing")
			}
			s := a.Snapshots[snapshot]
			if s.Tick != tick || s.Arrived != position || s.Pending != min(tick+1, 2400)-position || len(s.Forecast) != 150 {
				return fmt.Errorf("snapshot as-of domain")
			}
			if control {
				if err := independentWindowSnapshotV37(w, a, s, arrived, labels, weights); err != nil {
					return err
				}
			} else {
				_, q := shapeReference(ref)
				f := 0.
				if ready >= 0 {
					p := orientationReference(template, hazard, nodes)
					f = (1-hazard)*p + hazard*(1-p)
					for i := range q {
						q[i] = (1-f)*template[i] + f*(1-template[i])
					}
				}
				if !closeV36(f, s.Weights[0]) || s.Weights[1] != 0 || s.Weights[2] != 0 || s.Weights[3] != 0 {
					return fmt.Errorf("independent hidden-state mixture")
				}
				for i, value := range q {
					if !closeV36(value, s.Forecast[i]) {
						return fmt.Errorf("independent hidden-state forecast")
					}
				}
			}
			if err := independentMetricsV38(w, s); err != nil {
				return err
			}
			snapshot++
		}
	}
	if peak != a.PeakPending || position != 2400 || snapshot != len(a.Snapshots) || arm.ReadyAt != ready || len(arm.Template) != len(template) {
		return fmt.Errorf("drain/template/nomination coverage")
	}
	for i, p := range template {
		if !closeV36(p, arm.Template[i]) {
			return fmt.Errorf("batch-integral template")
		}
	}
	return nil
}

func independentMetricsV38(w windowWorldV37, s windowSnapshotV37) error {
	rates := w.Rates[min(15, s.Tick/150)]
	order := make([]int, 150)
	for i := range order {
		order[i] = i
	}
	// Stable insertion sort is deliberately separate from collector sorting.
	for j := 1; j < len(order); j++ {
		for k := j; k > 0 && s.Forecast[order[k]] > s.Forecast[order[k-1]]; k-- {
			order[k], order[k-1] = order[k-1], order[k]
		}
	}
	risk, priority, useful, bias := 0., 0., 0., 0.
	for i, p := range rates {
		q := s.Forecast[i]
		loss := q*q - 2*q*p + p
		weight := 1.
		if i < 10 {
			weight = 3
		}
		risk += loss / 150
		priority += weight * loss / 170
	}
	for _, i := range order[:10] {
		useful += rates[i] / 10
		bias += (s.Forecast[i] - rates[i]) / 10
	}
	if !closeV36(risk, s.Brier) || !closeV36(priority, s.PriorityBrier) || !closeV36(useful, s.PacketUsefulness) || !closeV36(bias, s.PacketBias) {
		return fmt.Errorf("independent packet/risk metrics")
	}
	return nil
}

func auditOrientationV38(w orientationWorldV38, seed int64) error {
	g, r := -1, -1
	for i, name := range []string{"tight", "wide"} {
		if w.Population.Geometry == name {
			g = i
		}
	}
	for i, name := range regimesV38 {
		if w.Population.Regime == name {
			r = i
		}
	}
	if g < 0 || r < 0 || w.Population.World < 0 || w.Population.World >= 16 || len(w.Arms) != 8 || !reflect.DeepEqual(w.Population, makeV38(seed, g, r, w.Population.World)) {
		return fmt.Errorf("population/RNG/domain")
	}
	for j, arm := range w.Arms {
		a := arm.windowArmV37
		if a.Mode != modesV38[j%4] || a.Schedule != schedulesV37[j/4] {
			return fmt.Errorf("arm order")
		}
		c := a.Costs
		if c.SetupNS < 0 || c.ScheduleNS < 0 || c.IssueNS < 0 || c.ResolveNS < 0 || c.SnapshotNS < 0 || c.AccountedNS != c.SetupNS+c.IssueNS+c.ResolveNS+c.SnapshotNS || c.ElapsedNS < c.AccountedNS+c.ScheduleNS {
			return fmt.Errorf("time accounting")
		}
		replay, err := runOrientationV38(w.Population, a.Mode, a.Schedule)
		if err != nil {
			return err
		}
		scoreWindowV37(w.Population, &replay.windowArmV37)
		a.Costs, arm.Costs, replay.Costs = delayedCostsV36{}, delayedCostsV36{}, delayedCostsV36{}
		if !reflect.DeepEqual(arm, replay) {
			return fmt.Errorf("exact law/receipt/metric replay %s/%s", a.Mode, a.Schedule)
		}
		if err := independentOrientationV38(w.Population, arm); err != nil {
			return fmt.Errorf("%s/%s: %w", a.Mode, a.Schedule, err)
		}
		risk, priority := 0., 0.
		for trial, q := range a.Issued {
			p := w.Population.Rates[trial/150][trial%150]
			loss := q*q - 2*q*p + p
			weight := 1.
			if trial%150 < 10 {
				weight = 3
			}
			risk += loss / 2400
			priority += weight * loss / (16 * 170)
		}
		if !closeV36(risk, a.IssuedBrier) || !closeV36(priority, a.IssuedPriority) {
			return fmt.Errorf("independent issued risks")
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

func allocationOrientationV38(b []byte) (int64, error) {
	maxima, counts := [2]int64{}, [2]int{}
	for _, line := range strings.Split(string(b), "\n") {
		v := strings.Fields(line)
		if len(v) == 0 {
			continue
		}
		index := -1
		if strings.HasPrefix(v[0], "BenchmarkOrientationObserverNew-") {
			index = 0
		} else if strings.HasPrefix(v[0], "BenchmarkOrientationObserverFreezePlan-") {
			index = 1
		}
		if index < 0 {
			continue
		}
		if len(v) != 8 || v[5] != "B/op" {
			return 0, fmt.Errorf("benchmark shape")
		}
		n, err := strconv.ParseInt(v[4], 10, 64)
		if err != nil || n < 0 {
			return 0, fmt.Errorf("benchmark allocation")
		}
		maxima[index] = max(maxima[index], n)
		counts[index]++
	}
	if counts != ([2]int{3, 3}) {
		return 0, fmt.Errorf("benchmark coverage %v", counts)
	}
	return maxima[0] + maxima[1], nil
}

func TestStudyAuditV38(t *testing.T) {
	input := os.Getenv("EVENTFRAME_ORIENTATION_V38_AUDIT")
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
	seed := int64(2026103803)
	if manifest.Split == "confirmation" {
		seed++
	} else if manifest.Split != "design" {
		t.Fatal("split")
	}
	if manifest.Kind != "manifest" || manifest.Worlds != 384 || manifest.SeedBase != seed {
		t.Fatal("manifest")
	}
	if err := sourcesV38(rootV34(t), manifest.Sources); err != nil {
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
			t.Fatal("world decode/duplicate", err)
		}
		if err := auditOrientationV38(w, seed); err != nil {
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
	if count != 384 || len(groups) != 24 || snapshots != 384*132 {
		t.Fatal("coverage", count, len(groups), snapshots)
	}
	benchmark, err := os.ReadFile(filepath.Join(rootV34(t), "docs/experiments/mmm-orientation-v38-benchmarks.txt"))
	if err != nil {
		t.Fatal(err)
	}
	allocation, err := allocationOrientationV38(benchmark)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]func(windowArmV37) float64{
		"IssuedBrier": func(a windowArmV37) float64 { return a.IssuedBrier }, "IssuedPriority": func(a windowArmV37) float64 { return a.IssuedPriority }, "Recovery": func(a windowArmV37) float64 { return a.Recovery },
		"FinalBrier": func(a windowArmV37) float64 { return windowAtV37(a, -1).Brier }, "FinalPriority": func(a windowArmV37) float64 { return windowAtV37(a, -1).PriorityBrier }, "FinalUsefulness": func(a windowArmV37) float64 { return windowAtV37(a, -1).PacketUsefulness }, "FinalBias": func(a windowArmV37) float64 { return windowAtV37(a, -1).PacketBias },
	}
	stats, gates := map[string]any{}, map[string]any{}
	all, failures := true, 0
	for name, worlds := range groups {
		if len(worlds) != 16 {
			t.Fatal("cell coverage")
		}
		for scheduleIndex, schedule := range schedulesV37 {
			key := name + "/" + schedule
			cell, checks := map[string]any{}, map[string]any{}
			maximum := int64(0)
			for j, mode := range modesV38 {
				row := map[string]any{}
				maxCost := int64(0)
				for field, metric := range fields {
					x := make([]float64, 16)
					for k, w := range worlds {
						a := w.Arms[scheduleIndex*4+j]
						x[k] = metric(a.windowArmV37)
						maxCost = max(maxCost, a.Costs.AccountedNS)
					}
					row[field] = interval34(x)
				}
				row["MaximumAccountedNS"] = maxCost
				maximum = max(maximum, maxCost)
				cell[mode] = row
			}
			pass := maximum <= 400000000 && allocation <= 8<<20
			checks["Work"] = map[string]any{"MaximumNS": maximum, "Pass": maximum <= 400000000}
			checks["Allocation"] = map[string]any{"MaximumBytes": allocation, "Pass": allocation <= 8<<20}
			shift := true
			for _, stationary := range regimesV38[:4] {
				if strings.HasSuffix(name, "/"+stationary) {
					shift = false
				}
			}
			for control := 0; control <= 1; control++ {
				for _, field := range []string{"IssuedBrier", "IssuedPriority", "FinalBrier", "FinalPriority", "FinalUsefulness"} {
					x := make([]float64, 16)
					for k, w := range worlds {
						a, b := w.Arms[scheduleIndex*4+3], w.Arms[scheduleIndex*4+control]
						x[k] = fields[field](b.windowArmV37) - fields[field](a.windowArmV37)
						if field == "FinalUsefulness" {
							x[k] = -x[k]
						}
					}
					v := interval34(x)
					improve := control == 0 && shift && (field == "IssuedBrier" || field == "IssuedPriority")
					ok := v.Lower >= -.01
					if improve {
						ok = v.Mean >= .01 && v.Lower > 0
					}
					pass = pass && ok
					checks[modesV38[control]+"/"+field] = map[string]any{"Gain": v, "Require001Improvement": improve, "Pass": ok}
				}
			}
			if len(worlds[0].Population.Changes) > 0 {
				x, base := make([]float64, 16), make([]float64, 16)
				for k, w := range worlds {
					a, b := w.Arms[scheduleIndex*4+3], w.Arms[scheduleIndex*4]
					x[k], base[k] = b.Recovery-a.Recovery, b.Recovery
				}
				v, control := interval34(x), interval34(base)
				ok := v.Lower > 0 && v.Mean >= .1*control.Mean
				pass = pass && ok
				checks["Recovery"] = map[string]any{"Gain": v, "Control": control, "Pass": ok}
			}
			all = all && pass
			if !pass {
				failures++
			}
			stats[key] = cell
			gates[key] = map[string]any{"Checks": checks, "Pass": pass}
		}
	}
	auditor, err := os.ReadFile(filepath.Join(rootV34(t), "internal/researchdispersion/orientation_audit_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	result := map[string]any{"Study": "orientation-v38", "Split": manifest.Split, "SHA256": hex.EncodeToString(hash.Sum(nil)), "Sources": manifest.Sources, "AuditorSHA256": hashV34(auditor), "BenchmarkSHA256": hashV34(benchmark), "Worlds": count, "Arms": count * 8, "Snapshots": snapshots, "DistinctTrials": count * 2400, "ComponentPass": all, "FailedCells": failures, "WholeGoalsComplete": false, "Goal7EqualTotalCost": false, "Groups": stats, "Gates": gates}
	auditSources := map[string]string{}
	for _, path := range []string{"internal/researchdispersion/orientation_audit_test.go", "internal/researchdispersion/window_audit_test.go", "internal/researchdispersion/delayed_audit_test.go"} {
		v, err := os.ReadFile(filepath.Join(rootV34(t), path))
		if err != nil {
			t.Fatal(err)
		}
		auditSources[path] = hashV34(v)
	}
	result["AuditSourceHashes"] = auditSources
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := strings.TrimSuffix(input, ".jsonl") + "-audit.json"
	if err := os.WriteFile(path, append(b, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("verified %d worlds/%d snapshots; adoption pass=%v failures=%d/48", count, snapshots, all, failures)
}

func TestOrientationAuditorNegativeV38(t *testing.T) {
	seed := int64(2026103803)
	w := orientationWorldV38{Population: makeV38(seed, 1, 6, 0)}
	for _, schedule := range schedulesV37 {
		for _, mode := range modesV38 {
			a, err := runOrientationV38(w.Population, mode, schedule)
			if err != nil {
				t.Fatal(err)
			}
			scoreWindowV37(w.Population, &a.windowArmV37)
			w.Arms = append(w.Arms, a)
		}
	}
	if err := auditOrientationV38(w, seed); err != nil {
		t.Fatal("valid control", err)
	}
	b, _ := json.Marshal(w)
	mutations := []func(*orientationWorldV38){
		func(x *orientationWorldV38) { x.Population.Rates[0][0] += .01 },
		func(x *orientationWorldV38) { x.Arms[3].Template[0] += .01 },
		func(x *orientationWorldV38) { x.Arms[3].ReadyAt++ },
		func(x *orientationWorldV38) { x.Arms[3].Issued[2000] += .01 },
		func(x *orientationWorldV38) { x.Arms[3].ExpertIssued[0][0] = .5 },
		func(x *orientationWorldV38) { x.Arms[3].Receipts[1000].TrialOrdinal++ },
		func(x *orientationWorldV38) { x.Arms[3].Snapshots[8].Weights[0] += .01 },
		func(x *orientationWorldV38) { x.Arms[3].Snapshots[8].Forecast[0] += .01 },
		func(x *orientationWorldV38) { x.Arms[3].IssuedPriority += .01 },
		func(x *orientationWorldV38) { x.Arms[3].Recovery++ },
		func(x *orientationWorldV38) { x.Arms[3].Costs.AccountedNS++ },
	}
	for j, change := range mutations {
		var x orientationWorldV38
		json.Unmarshal(b, &x)
		change(&x)
		if err := auditOrientationV38(x, seed); err == nil {
			t.Fatal("corruption accepted", j)
		}
	}
	hashes := map[string]string{}
	for _, p := range filesV38 {
		v, err := os.ReadFile(filepath.Join(rootV34(t), p))
		if err != nil {
			t.Fatal(err)
		}
		hashes[p] = hashV34(v)
	}
	hashes[filesV38[0]] = strings.Repeat("0", 64)
	if err := sourcesV38(rootV34(t), hashes); err == nil {
		t.Fatal("changed source accepted")
	}
	if _, err := allocationOrientationV38([]byte("incomplete")); err == nil {
		t.Fatal("missing benchmark accepted")
	}
}
