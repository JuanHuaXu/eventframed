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
	"sort"
	"strconv"
	"strings"
	"testing"
)

func sourcesV37(root string, hashes map[string]string) error {
	if len(hashes) != len(filesV37) {
		return fmt.Errorf("source count")
	}
	for _, path := range filesV37 {
		b, err := os.ReadFile(filepath.Join(root, path))
		if err != nil || hashV34(b) != hashes[path] {
			return fmt.Errorf("frozen source changed %s: %v", path, err)
		}
	}
	return nil
}

func widthsV37(mode string) []int {
	switch mode {
	case "full":
		return []int{64}
	case "fixed4":
		return []int{4}
	case "fixed8":
		return []int{8}
	case "fixed16":
		return []int{16}
	case "adaptive":
		return []int{64, 4, 8, 16}
	}
	return nil
}

func independentWindowSnapshotV37(w windowWorldV37, a windowArmV37, s windowSnapshotV37, arrived, labels []uint64, weights [4]float64) error {
	widths := widthsV37(a.Mode)
	q := make([]float64, 150)
	ref, _ := NewShape(w.Base)
	for j, width := range widths {
		for i, mask := range arrived {
			n, count := 0, 0
			for ordinal := 63; ordinal >= 0 && n < width; ordinal-- {
				bit := uint64(1) << ordinal
				if mask&bit == 0 {
					continue
				}
				n++
				if labels[i]&bit != 0 {
					count++
				}
			}
			ref.n[i], ref.success[i] = uint16(n), uint16(count)
		}
		_, child := shapeReference(ref)
		for i, v := range child {
			q[i] += weights[j] * v
		}
	}
	for j, value := range s.Weights {
		if !closeV36(value, weights[j]) {
			return fmt.Errorf("independent expert weight replay")
		}
	}
	for i, value := range s.Forecast {
		if value <= 0 || value >= 1 || !closeV36(value, q[i]) {
			return fmt.Errorf("independent batch window mixture law")
		}
	}
	rates := w.Rates[min(15, s.Tick/150)]
	order := make([]int, 150)
	risk, priority, useful, bias := 0., 0., 0., 0.
	for i, p := range rates {
		order[i] = i
		v := s.Forecast[i]
		loss := v*v - 2*v*p + p
		risk += loss / 150
		priority += loss / 170
		if i < 10 {
			priority += 2 * loss / 170
		}
	}
	sort.SliceStable(order, func(i, j int) bool { return s.Forecast[order[i]] > s.Forecast[order[j]] })
	for _, i := range order[:10] {
		useful += rates[i] / 10
		bias += (s.Forecast[i] - rates[i]) / 10
	}
	if !closeV36(risk, s.Brier) || !closeV36(priority, s.PriorityBrier) || !closeV36(useful, s.PacketUsefulness) || !closeV36(bias, s.PacketBias) {
		return fmt.Errorf("future packet/risk reconstruction")
	}
	return nil
}

func auditWindowV37(w windowWorldV37, seed int64) error {
	g, r := -1, -1
	for j, name := range []string{"tight", "wide"} {
		if w.Geometry == name {
			g = j
		}
	}
	for j, name := range regimesV37 {
		if w.Regime == name {
			r = j
		}
	}
	if g < 0 || r < 0 || w.World < 0 || w.World >= 16 || w.Kind != "world" || len(w.Arms) != 10 {
		return fmt.Errorf("world domain")
	}
	expected := makeV37(seed, g, r, w.World)
	if expected.Seed != w.Seed || !reflect.DeepEqual(expected.Base, w.Base) || !reflect.DeepEqual(expected.Rates, w.Rates) || !reflect.DeepEqual(expected.Outcomes, w.Outcomes) || !reflect.DeepEqual(expected.Changes, w.Changes) {
		return fmt.Errorf("population RNG replay")
	}
	for index, a := range w.Arms {
		if a.Mode != modesV37[index%5] || a.Schedule != schedulesV37[index/5] || len(a.Issued) != 2400 || len(a.ExpertIssued) != 2400 || len(a.Receipts) != 2400 || a.PeakPending < 1 || a.PeakPending > 2400 {
			return fmt.Errorf("arm domain")
		}
		b, err := runWindowV37(expected, a.Mode, a.Schedule)
		if err != nil {
			return err
		}
		scoreWindowV37(expected, &b)
		x, y := a, b
		x.Costs, y.Costs = delayedCostsV36{}, delayedCostsV36{}
		if !reflect.DeepEqual(x, y) {
			return fmt.Errorf("original law/expert/receipt/snapshot RNG replay")
		}
		c := a.Costs
		if c.SetupNS < 0 || c.ScheduleNS < 0 || c.IssueNS < 0 || c.ResolveNS < 0 || c.SnapshotNS < 0 || c.AccountedNS != c.SetupNS+c.IssueNS+c.ResolveNS+c.SnapshotNS || c.ElapsedNS < c.AccountedNS+c.ScheduleNS {
			return fmt.Errorf("phase accounting")
		}
		weights := [4]float64{1}
		if a.Mode == "adaptive" {
			weights = [4]float64{.85, .05, .05, .05}
		}
		arrived, labels := make([]uint64, 150), make([]uint64, 150)
		position, snapshot := 0, 0
		last := int(a.Receipts[len(a.Receipts)-1].ArrivedAt)
		for tick := 0; tick <= last; tick++ {
			if tick < 2400 {
				row, q := a.ExpertIssued[tick], 0.
				for j, value := range row {
					if j >= len(widthsV37(a.Mode)) {
						if value != 0 {
							return fmt.Errorf("unused expert row")
						}
						continue
					}
					if value <= 0 || value >= 1 || math.IsNaN(value) {
						return fmt.Errorf("expert forecast domain")
					}
					q += weights[j] * value
				}
				if !closeV36(a.Issued[tick], q) {
					return fmt.Errorf("issue-time mixture replay")
				}
			}
			for position < len(a.Receipts) && a.Receipts[position].ArrivedAt == int64(tick) {
				r := a.Receipts[position]
				if r.Member < 0 || r.Member >= 150 || r.TrialOrdinal < 1 || r.TrialOrdinal > 16 || r.Epoch != 1 {
					return fmt.Errorf("receipt domain")
				}
				trial := (r.TrialOrdinal-1)*150 + r.Member
				bit := uint64(1) << (r.TrialOrdinal - 1)
				if arrived[r.Member]&bit != 0 || r.IssuedAt != int64(trial) || r.ArrivedAt < r.IssuedAt || r.Forecast != a.Issued[trial] || r.Useful != w.Outcomes[r.TrialOrdinal-1][r.Member] {
					return fmt.Errorf("original receipt/retention identity")
				}
				arrived[r.Member] |= bit
				if r.Useful {
					labels[r.Member] |= bit
				}
				if a.Mode == "adaptive" {
					sum := 0.
					for j, value := range a.ExpertIssued[trial] {
						if !r.Useful {
							value = 1 - value
						}
						weights[j] *= value
						sum += weights[j]
					}
					for j := range weights {
						weights[j] = (599.0/600)*weights[j]/sum + [4]float64{.85, .05, .05, .05}[j]/600
					}
				}
				position++
			}
			if snapshot < len(a.Snapshots) && a.Snapshots[snapshot].Tick == tick {
				s := a.Snapshots[snapshot]
				if len(s.Forecast) != 150 || s.Arrived != position || s.Pending != min(2400, tick+1)-position {
					return fmt.Errorf("snapshot as-of counts")
				}
				if err := independentWindowSnapshotV37(w, a, s, arrived, labels, weights); err != nil {
					return err
				}
				snapshot++
			}
		}
		if position != 2400 || snapshot != len(a.Snapshots) || a.Snapshots[len(a.Snapshots)-1].Pending != 0 {
			return fmt.Errorf("drain coverage")
		}
		risk, priority := 0., 0.
		for trial, q := range a.Issued {
			p := w.Rates[trial/150][trial%150]
			loss := q*q - 2*q*p + p
			risk += loss / 2400
			priority += loss / (16 * 170)
			if trial%150 < 10 {
				priority += 2 * loss / (16 * 170)
			}
		}
		if !closeV36(risk, a.IssuedBrier) || !closeV36(priority, a.IssuedPriority) {
			return fmt.Errorf("original expected score")
		}
		// Independent scan of each declared phase; no future/after-drain
		// snapshot can shorten a phase's restricted recovery delay.
		recovery := 0.
		for j, start := range w.Changes {
			end := 16
			if j+1 < len(w.Changes) {
				end = w.Changes[j+1]
			}
			delay := end - start + 1
			for round := start + 1; round < end; round++ {
				a, b := windowAtV37(a, (round-1)*150+149), windowAtV37(a, round*150+149)
				if a.Brier <= .20 && b.Brier <= .20 && a.PacketUsefulness >= .75 && b.PacketUsefulness >= .75 {
					delay = round - start + 1
					break
				}
			}
			recovery += float64(delay) / float64(len(w.Changes))
		}
		if !closeV36(recovery, a.Recovery) {
			return fmt.Errorf("phase recovery reconstruction")
		}
	}
	return nil
}

func windowAtV37(a windowArmV37, tick int) windowSnapshotV37 {
	if tick < 0 {
		return a.Snapshots[len(a.Snapshots)-1]
	}
	for _, s := range a.Snapshots {
		if s.Tick == tick {
			return s
		}
	}
	panic("missing snapshot")
}

func TestStudyAuditV37(t *testing.T) {
	input := os.Getenv("EVENTFRAME_WINDOW_V37_AUDIT")
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
	seed := int64(2026103703)
	if manifest.Split == "confirmation" {
		seed++
	} else if manifest.Split != "design" {
		t.Fatal("split")
	}
	if manifest.Kind != "manifest" || manifest.SeedBase != seed || manifest.Worlds != 256 {
		t.Fatal("manifest")
	}
	if err := sourcesV37(rootV34(t), manifest.Sources); err != nil {
		t.Fatal(err)
	}
	groups := map[string][]windowWorldV37{}
	seen := map[int64]bool{}
	count, snapshots := 0, 0
	for {
		var w windowWorldV37
		err := d.Decode(&w)
		if err == io.EOF {
			break
		}
		if err != nil || seen[w.Seed] {
			t.Fatal("world decoding/duplicate", err)
		}
		if err := auditWindowV37(w, seed); err != nil {
			t.Fatal(wNameV37(w), err)
		}
		seen[w.Seed] = true
		count++
		for _, a := range w.Arms {
			snapshots += len(a.Snapshots)
		}
		name := w.Geometry + "/" + w.Regime
		groups[name] = append(groups[name], w)
	}
	if count != 256 || len(groups) != 16 || snapshots != 256*(5*16+5*17) {
		t.Fatal("coverage", count, len(groups), snapshots)
	}
	benchmark, err := os.ReadFile(filepath.Join(rootV34(t), "docs/experiments/mmm-window-v37-benchmarks.txt"))
	if err != nil {
		t.Fatal(err)
	}
	alloc, rows := int64(0), 0
	for _, line := range strings.Split(string(benchmark), "\n") {
		v := strings.Fields(line)
		if len(v) == 0 || v[0] != "BenchmarkWindowObserverNew-10" {
			continue
		}
		if len(v) != 8 || v[5] != "B/op" {
			t.Fatal("benchmark format")
		}
		n, err := strconv.ParseInt(v[4], 10, 64)
		if err != nil || n < 0 {
			t.Fatal("benchmark allocation", err)
		}
		alloc = max(alloc, n)
		rows++
	}
	if rows != 3 {
		t.Fatal("benchmark rows")
	}
	stats, gates := map[string]any{}, map[string]any{}
	all, failed := true, 0
	fields := map[string]func(windowArmV37) float64{
		"IssuedBrier": func(a windowArmV37) float64 { return a.IssuedBrier }, "IssuedPriority": func(a windowArmV37) float64 { return a.IssuedPriority }, "Recovery": func(a windowArmV37) float64 { return a.Recovery },
		"FinalBrier": func(a windowArmV37) float64 { return windowAtV37(a, -1).Brier }, "FinalPriority": func(a windowArmV37) float64 { return windowAtV37(a, -1).PriorityBrier }, "FinalUsefulness": func(a windowArmV37) float64 { return windowAtV37(a, -1).PacketUsefulness }, "FinalBias": func(a windowArmV37) float64 { return windowAtV37(a, -1).PacketBias },
	}
	for name, worlds := range groups {
		if len(worlds) != 16 {
			t.Fatal("cell coverage")
		}
		for scheduleIndex, schedule := range schedulesV37 {
			key := name + "/" + schedule
			cell, checks := map[string]any{}, map[string]any{}
			maximum := int64(0)
			for j, mode := range modesV37 {
				row := map[string]any{}
				maxCost := int64(0)
				for field, metric := range fields {
					x := make([]float64, 16)
					for k, w := range worlds {
						x[k] = metric(w.Arms[scheduleIndex*5+j])
						maxCost = max(maxCost, w.Arms[scheduleIndex*5+j].Costs.AccountedNS)
					}
					row[field] = interval34(x)
				}
				row["MaximumAccountedNS"] = maxCost
				maximum = max(maximum, maxCost)
				cell[mode] = row
			}
			pass := alloc <= 8<<20 && maximum <= 400000000
			checks["Work"] = map[string]any{"MaximumNS": maximum, "LimitNS": 400000000, "Pass": maximum <= 400000000}
			checks["Allocation"] = map[string]any{"MaximumBytes": alloc, "LimitBytes": 8 << 20, "Pass": alloc <= 8<<20}
			shift := strings.HasSuffix(name, "/abrupt") || strings.HasSuffix(name, "/late") || strings.HasSuffix(name, "/recurring") || strings.HasSuffix(name, "/gradual")
			for _, field := range []string{"IssuedBrier", "IssuedPriority", "FinalBrier", "FinalPriority", "FinalUsefulness"} {
				x := make([]float64, 16)
				for k, w := range worlds {
					a, b := w.Arms[scheduleIndex*5+4], w.Arms[scheduleIndex*5]
					x[k] = fields[field](b) - fields[field](a)
					if field == "FinalUsefulness" {
						x[k] = -x[k]
					}
				}
				v := interval34(x)
				improve := shift && (field == "IssuedBrier" || field == "IssuedPriority")
				ok := v.Lower >= -.01
				if improve {
					ok = v.Mean >= .01 && v.Lower > 0
				}
				pass = pass && ok
				checks[field] = map[string]any{"Gain": v, "Require001Improvement": improve, "Pass": ok}
			}
			if len(worlds[0].Changes) > 0 {
				x, base := make([]float64, 16), make([]float64, 16)
				for k, w := range worlds {
					a, b := w.Arms[scheduleIndex*5+4], w.Arms[scheduleIndex*5]
					x[k] = b.Recovery - a.Recovery
					base[k] = b.Recovery
				}
				v, control := interval34(x), interval34(base)
				ok := v.Lower > 0 && v.Mean >= .1*control.Mean
				pass = pass && ok
				checks["Recovery"] = map[string]any{"Gain": v, "Control": control, "Pass": ok}
			}
			gates[key] = map[string]any{"Checks": checks, "Pass": pass}
			stats[key] = cell
			all = all && pass
			if !pass {
				failed++
			}
		}
	}
	auditor, err := os.ReadFile(filepath.Join(rootV34(t), "internal/researchdispersion/window_audit_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	result := map[string]any{"Study": "window-v37", "Split": manifest.Split, "SHA256": hex.EncodeToString(hash.Sum(nil)), "SourceHashes": manifest.Sources, "AuditorSHA256": hashV34(auditor), "BenchmarkSHA256": hashV34(benchmark), "Worlds": count, "Arms": count * 10, "Snapshots": snapshots, "DistinctTrials": count * 2400, "ComponentPass": all, "FailedCells": failed, "WholeGoalsComplete": false, "Goal7EqualTotalCost": false, "Groups": stats, "Gates": gates}
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(rootV34(t), "docs/experiments/mmm-window-v37-"+manifest.Split+"-audit.json")
	if err := os.WriteFile(out, append(b, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("verified256 worlds/%d snapshots; pass=%v failed=%d/32", snapshots, all, failed)
}

func TestWindowAuditorNegativeV37(t *testing.T) {
	w := makeV37(2026103799, 1, 4, 0)
	for _, schedule := range schedulesV37 {
		for _, mode := range modesV37 {
			a, err := runWindowV37(w, mode, schedule)
			if err != nil {
				t.Fatal(err)
			}
			scoreWindowV37(w, &a)
			w.Arms = append(w.Arms, a)
		}
	}
	if err := auditWindowV37(w, 2026103799); err != nil {
		t.Fatal("valid control", err)
	}
	for name, mutate := range map[string]func(*windowWorldV37){
		"issued": func(w *windowWorldV37) { w.Arms[0].Issued[0] += .1 }, "expert": func(w *windowWorldV37) { w.Arms[4].ExpertIssued[0][0] += .1 }, "receipt": func(w *windowWorldV37) { w.Arms[0].Receipts[0].Forecast += .1 },
		"weights": func(w *windowWorldV37) { w.Arms[4].Snapshots[0].Weights[0] += .1 }, "retention": func(w *windowWorldV37) { w.Arms[0].Receipts[1].TrialOrdinal++ }, "law": func(w *windowWorldV37) { w.Arms[0].Snapshots[0].Forecast[0] += .1 },
		"cost": func(w *windowWorldV37) { w.Arms[0].Costs.AccountedNS++ }, "risk": func(w *windowWorldV37) { w.Arms[0].IssuedPriority += .1 }, "recovery": func(w *windowWorldV37) { w.Arms[4].Recovery++ }, "outcome": func(w *windowWorldV37) { w.Outcomes[0][0] = !w.Outcomes[0][0] },
	} {
		t.Run(name, func(t *testing.T) {
			b, _ := json.Marshal(w)
			var copy windowWorldV37
			if err := json.Unmarshal(b, &copy); err != nil {
				t.Fatal(err)
			}
			mutate(&copy)
			if err := auditWindowV37(copy, 2026103799); err == nil {
				t.Fatal("corruption accepted")
			}
		})
	}
	hashes := map[string]string{}
	for _, path := range filesV37 {
		b, err := os.ReadFile(filepath.Join(rootV34(t), path))
		if err != nil {
			t.Fatal(err)
		}
		hashes[path] = hashV34(b)
	}
	if err := sourcesV37(rootV34(t), hashes); err != nil {
		t.Fatal(err)
	}
	hashes[filesV37[0]] = "corrupted"
	if sourcesV37(rootV34(t), hashes) == nil {
		t.Fatal("changed source accepted")
	}
}
