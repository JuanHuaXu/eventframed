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

func sourcesV36(root string, hashes map[string]string) error {
	if len(hashes) != len(filesV36) {
		return fmt.Errorf("source count")
	}
	for _, path := range filesV36 {
		b, err := os.ReadFile(filepath.Join(root, path))
		if err != nil || hashV34(b) != hashes[path] {
			return fmt.Errorf("frozen source changed %s: %v", path, err)
		}
	}
	return nil
}

func closeV36(a, b float64) bool {
	return !math.IsNaN(a) && !math.IsInf(a, 0) && math.Abs(a-b) <= 3e-10
}

func auditDelayedV36(w delayedWorldV36, seed int64) error {
	g, r := -1, -1
	for j, name := range []string{"tight", "wide"} {
		if w.Geometry == name {
			g = j
		}
	}
	for j, name := range regimesV36 {
		if w.Regime == name {
			r = j
		}
	}
	if g < 0 || r < 0 || w.World < 0 || w.World >= 16 || w.Kind != "world" || len(w.Arms) != len(schedulesV36) {
		return fmt.Errorf("world domain")
	}
	expected := makeV36(seed, g, r, w.World)
	if expected.Seed != w.Seed || !reflect.DeepEqual(expected.Base, w.Base) || !reflect.DeepEqual(expected.Before, w.Before) || !reflect.DeepEqual(expected.After, w.After) || !reflect.DeepEqual(expected.Outcomes, w.Outcomes) {
		return fmt.Errorf("population RNG replay")
	}
	for j, a := range w.Arms {
		if a.Schedule != schedulesV36[j] || len(a.Issued) != 2400 || len(a.Receipts) != 2400 || a.PeakPending < 1 || a.PeakPending > 2400 {
			return fmt.Errorf("arm domain")
		}
		b, err := runDelayedV36(expected, a.Schedule)
		if err != nil {
			return err
		}
		scoreDelayedV36(expected, &b)
		x, y := a, b
		x.Costs, y.Costs = delayedCostsV36{}, delayedCostsV36{}
		if !reflect.DeepEqual(x, y) {
			return fmt.Errorf("issued-law/receipt/snapshot seed replay")
		}
		c := a.Costs
		if c.SetupNS < 0 || c.ScheduleNS < 0 || c.IssueNS < 0 || c.ResolveNS < 0 || c.SnapshotNS < 0 || c.AccountedNS != c.SetupNS+c.IssueNS+c.ResolveNS+c.SnapshotNS || c.ElapsedNS < c.AccountedNS+c.ScheduleNS {
			return fmt.Errorf("cost accounting")
		}
		seen := map[int]bool{}
		lastTick, lastTrial := int64(-1), -1
		for _, receipt := range a.Receipts {
			trial := (receipt.TrialOrdinal-1)*150 + receipt.Member
			if receipt.Member < 0 || receipt.Member >= 150 || receipt.TrialOrdinal < 1 || receipt.TrialOrdinal > 16 || receipt.Epoch != 1 || seen[trial] || receipt.IssuedAt != int64(trial) || receipt.ArrivedAt < receipt.IssuedAt || receipt.ArrivedAt < lastTick || (receipt.ArrivedAt == lastTick && trial <= lastTrial) || receipt.Forecast != a.Issued[trial] || receipt.Useful != w.Outcomes[receipt.TrialOrdinal-1][receipt.Member] {
				return fmt.Errorf("original evidence binding/order")
			}
			seen[trial], lastTick, lastTrial = true, receipt.ArrivedAt, trial
		}
		m, _ := NewShape(w.Base)
		arrived := 0
		for _, s := range a.Snapshots {
			for arrived < len(a.Receipts) && a.Receipts[arrived].ArrivedAt <= int64(s.Tick) {
				r := a.Receipts[arrived]
				m.n[r.Member]++
				if r.Useful {
					m.success[r.Member]++
				}
				arrived++
			}
			issued := min(2400, s.Tick+1)
			if s.Arrived != arrived || s.Pending != issued-arrived || len(s.Forecast) != 150 {
				return fmt.Errorf("snapshot as-of counts")
			}
			_, q := shapeReference(m)
			for i, got := range s.Forecast {
				if got <= 0 || got >= 1 || !closeV36(got, q[i]) {
					return fmt.Errorf("independent batch arrived-set law")
				}
			}
			rates := w.Before
			if s.Tick >= 1200 {
				rates = w.After
			}
			order := make([]int, 150)
			brier, priority, useful, bias := 0., 0., 0., 0.
			for i, p := range rates {
				order[i] = i
				q := s.Forecast[i]
				loss := q*q - 2*q*p + p
				weight := 1.
				if i < 10 {
					weight = 3
				}
				brier += loss / 150
				priority += loss * weight / 170
			}
			sort.SliceStable(order, func(i, j int) bool { return s.Forecast[order[i]] > s.Forecast[order[j]] })
			for _, i := range order[:10] {
				useful += rates[i] / 10
				bias += (s.Forecast[i] - rates[i]) / 10
			}
			if !closeV36(s.Brier, brier) || !closeV36(s.PriorityBrier, priority) || !closeV36(s.PacketUsefulness, useful) || !closeV36(s.PacketBias, bias) {
				return fmt.Errorf("risk/priority/packet reconstruction")
			}
		}
		final := a.Snapshots[len(a.Snapshots)-1]
		if final.Arrived != 2400 || final.Pending != 0 {
			return fmt.Errorf("undrained feedback")
		}
		baseline := w.Arms[0].Snapshots[len(w.Arms[0].Snapshots)-1]
		for i, q := range final.Forecast {
			if !closeV36(q, baseline.Forecast[i]) {
				return fmt.Errorf("full-evidence order invariance")
			}
		}
		issuedRisk := 0.
		for trial, q := range a.Issued {
			if q <= 0 || q >= 1 || math.IsNaN(q) {
				return fmt.Errorf("issued forecast domain")
			}
			rates := w.Before
			if trial >= 1200 {
				rates = w.After
			}
			p := rates[trial%150]
			issuedRisk += ((q-p)*(q-p) + p*(1-p)) / 2400
		}
		if !closeV36(a.IssuedBrier, issuedRisk) {
			return fmt.Errorf("original issued expected Brier")
		}
	}
	return nil
}

func atV36(a delayedArmV36, tick int) delayedSnapshotV36 {
	if tick < 0 {
		return a.Snapshots[len(a.Snapshots)-1]
	}
	for _, s := range a.Snapshots {
		if s.Tick == tick {
			return s
		}
	}
	panic("missing as-of snapshot")
}

func TestStudyAuditV36(t *testing.T) {
	input := os.Getenv("EVENTFRAME_DELAYED_V36_AUDIT")
	if input == "" {
		t.Skip("explicit input required")
	}
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	hash := sha256.New()
	decoder := json.NewDecoder(bufio.NewReader(io.TeeReader(f, hash)))
	decoder.DisallowUnknownFields()
	var manifest manifestV34
	if err := decoder.Decode(&manifest); err != nil {
		t.Fatal(err)
	}
	seed := int64(2026103603)
	if manifest.Split == "confirmation" {
		seed++
	} else if manifest.Split != "design" {
		t.Fatal("split")
	}
	if manifest.Kind != "manifest" || manifest.SeedBase != seed || manifest.Worlds != 192 {
		t.Fatal("manifest")
	}
	if err := sourcesV36(rootV34(t), manifest.Sources); err != nil {
		t.Fatal(err)
	}
	groups := map[string][]delayedWorldV36{}
	seen := map[int64]bool{}
	count, snapshots, expectedSnapshots, maximum := 0, 0, 0, int64(0)
	for {
		var w delayedWorldV36
		err := decoder.Decode(&w)
		if err == io.EOF {
			break
		}
		if err != nil || seen[w.Seed] {
			t.Fatal("invalid or duplicate world", err)
		}
		if err := auditDelayedV36(w, seed); err != nil {
			t.Fatal(w.Geometry, w.Regime, w.World, err)
		}
		seen[w.Seed] = true
		for _, a := range w.Arms {
			snapshots += len(a.Snapshots)
			// Final drain can coincide with tick2399 (immediate AND burst).
			// Derive the checkpoint union from the already audited arrivals.
			expectedSnapshots += 3
			if a.Receipts[len(a.Receipts)-1].ArrivedAt != 2399 {
				expectedSnapshots++
			}
			maximum = max(maximum, a.Costs.AccountedNS)
		}
		name := w.Geometry + "/" + w.Regime
		groups[name] = append(groups[name], w)
		count++
	}
	if count != 192 || len(groups) != 12 || snapshots != expectedSnapshots {
		t.Fatal("coverage", count, len(groups), snapshots)
	}
	benchmark, err := os.ReadFile(filepath.Join(rootV34(t), "docs/experiments/mmm-delayed-v36-benchmarks.txt"))
	if err != nil {
		t.Fatal(err)
	}
	alloc, rows := int64(0), 0
	for _, line := range strings.Split(string(benchmark), "\n") {
		v := strings.Fields(line)
		if len(v) == 0 || v[0] != "BenchmarkDelayedShapeNew-10" {
			continue
		}
		if len(v) != 8 || v[5] != "B/op" {
			t.Fatal("benchmark format")
		}
		value, err := strconv.ParseInt(v[4], 10, 64)
		if err != nil || value < 0 {
			t.Fatal("allocation", err)
		}
		rows++
		alloc = max(alloc, value)
	}
	if rows != 3 {
		t.Fatal("benchmark coverage")
	}
	stats, paired := map[string]any{}, map[string]any{}
	for name, worlds := range groups {
		if len(worlds) != 16 {
			t.Fatal("cell coverage")
		}
		cell, deltas := map[string]any{}, map[string]any{}
		for k, schedule := range schedulesV36 {
			issued, issueDifference := make([]float64, 16), make([]float64, 16)
			maxWork, peak := int64(0), 0
			for j, w := range worlds {
				a := w.Arms[k]
				issued[j] = a.IssuedBrier
				issueDifference[j] = a.IssuedBrier - w.Arms[0].IssuedBrier
				maxWork, peak = max(maxWork, a.Costs.AccountedNS), max(peak, a.PeakPending)
			}
			asof, harm := map[string]any{}, map[string]any{}
			for _, tick := range []int{599, 1199, 2399, -1} {
				label := strconv.Itoa(tick)
				if tick < 0 {
					label = "final"
				}
				metrics := map[string]func(delayedSnapshotV36) float64{
					"Brier":            func(s delayedSnapshotV36) float64 { return s.Brier },
					"PriorityBrier":    func(s delayedSnapshotV36) float64 { return s.PriorityBrier },
					"PacketUsefulness": func(s delayedSnapshotV36) float64 { return s.PacketUsefulness },
					"PacketBias":       func(s delayedSnapshotV36) float64 { return s.PacketBias },
					"Arrived":          func(s delayedSnapshotV36) float64 { return float64(s.Arrived) },
				}
				values, differences := map[string]any{}, map[string]any{}
				for field, metric := range metrics {
					x, d := make([]float64, 16), make([]float64, 16)
					for j, w := range worlds {
						x[j] = metric(atV36(w.Arms[k], tick))
						d[j] = x[j] - metric(atV36(w.Arms[0], tick))
					}
					values[field], differences[field] = interval34(x), interval34(d)
				}
				asof[label], harm[label] = values, differences
			}
			cell[schedule] = map[string]any{"IssuedBrier": interval34(issued), "MaximumAccountedNS": maxWork, "PeakPending": peak, "AsOf": asof}
			deltas[schedule] = map[string]any{"IssuedBrierHarm": interval34(issueDifference), "AsOfDifference": harm}
		}
		stats[name], paired[name] = cell, deltas
	}
	auditor, err := os.ReadFile(filepath.Join(rootV34(t), "internal/researchdispersion/delayed_audit_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	result := map[string]any{"Study": "delayed-v36", "Split": manifest.Split, "SHA256": hex.EncodeToString(hash.Sum(nil)), "SourceHashes": manifest.Sources, "AuditorSHA256": hashV34(auditor), "BenchmarkSHA256": hashV34(benchmark), "Worlds": count, "Arms": count * 6, "Snapshots": snapshots, "DistinctTrials": count * 2400, "LearnerCopiesAreNotIndependentTrials": true, "ComponentPass": maximum <= 50000000 && alloc <= 2<<20, "MaximumAccountedNS": maximum, "MaximumConstructorBytes": alloc, "WholeGoalsComplete": false, "ModelAdequacyValidated": false, "Goal7EqualTotalCost": false, "Groups": stats, "PairedDelayHarm": paired}
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(rootV34(t), "docs/experiments/mmm-delayed-v36-"+manifest.Split+"-audit.json")
	if err := os.WriteFile(out, append(b, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("verified %d worlds/%d snapshots; max learner work %.3fms; pass=%v", count, snapshots, float64(maximum)/1e6, result["ComponentPass"])
}

func TestDelayedAuditorNegativeV36(t *testing.T) {
	w := makeV36(2026103699, 1, 1, 0)
	for _, name := range schedulesV36 {
		a, err := runDelayedV36(w, name)
		if err != nil {
			t.Fatal(err)
		}
		scoreDelayedV36(w, &a)
		w.Arms = append(w.Arms, a)
	}
	if err := auditDelayedV36(w, 2026103699); err != nil {
		t.Fatal("valid control", err)
	}
	for name, mutate := range map[string]func(*delayedWorldV36){
		"issued":     func(w *delayedWorldV36) { w.Arms[0].Issued[0] += .1 },
		"receipt":    func(w *delayedWorldV36) { w.Arms[0].Receipts[0].Forecast += .1 },
		"outcome":    func(w *delayedWorldV36) { w.Outcomes[0][0] = !w.Outcomes[0][0] },
		"future-law": func(w *delayedWorldV36) { w.Arms[0].Snapshots[0].Forecast[0] += .1 },
		"risk":       func(w *delayedWorldV36) { w.Arms[0].Snapshots[0].Brier += .1 },
		"time":       func(w *delayedWorldV36) { w.Arms[1].Receipts[0].ArrivedAt++ },
		"cost":       func(w *delayedWorldV36) { w.Arms[0].Costs.AccountedNS++ },
	} {
		t.Run(name, func(t *testing.T) {
			b, _ := json.Marshal(w)
			var copy delayedWorldV36
			if err := json.Unmarshal(b, &copy); err != nil {
				t.Fatal(err)
			}
			mutate(&copy)
			if err := auditDelayedV36(copy, 2026103699); err == nil {
				t.Fatal("corruption accepted")
			}
		})
	}
	hashes := map[string]string{}
	for _, p := range filesV36 {
		b, err := os.ReadFile(filepath.Join(rootV34(t), p))
		if err != nil {
			t.Fatal(err)
		}
		hashes[p] = hashV34(b)
	}
	if err := sourcesV36(rootV34(t), hashes); err != nil {
		t.Fatal(err)
	}
	hashes[filesV36[0]] = "corrupted"
	if err := sourcesV36(rootV34(t), hashes); err == nil {
		t.Fatal("changed source accepted")
	}
}

func TestDelayedCoupledFuturePrefixV36(t *testing.T) {
	w, v := makeV36(2026103699, 1, 1, 0), makeV36(2026103699, 1, 1, 0)
	for round := 8; round < 16; round++ {
		for i := range v.Outcomes[round] {
			v.Outcomes[round][i] = !v.Outcomes[round][i]
		}
	}
	a, err := runDelayedV36(w, "outcome_coupled")
	if err != nil {
		t.Fatal(err)
	}
	b, err := runDelayedV36(v, "outcome_coupled")
	if err != nil || !reflect.DeepEqual(a.Issued[:1201], b.Issued[:1201]) {
		t.Fatal("future coupled schedule affected earlier law", err)
	}
}
