package researchdispersion

import (
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

func shapeSourcesV35(root string, hashes map[string]string) error {
	if len(hashes) != len(filesV35) {
		return fmt.Errorf("source count")
	}
	for _, p := range filesV35 {
		b, err := os.ReadFile(filepath.Join(root, p))
		if err != nil || hashV34(b) != hashes[p] {
			return fmt.Errorf("frozen source changed %s: %v", p, err)
		}
	}
	return nil
}

func auditShapeV35(w worldV35, seed int64) error {
	g, r := -1, -1
	for k, name := range []string{"tight", "wide"} {
		if w.Geometry == name {
			g = k
		}
	}
	for k, name := range regimesV35 {
		if w.Regime == name {
			r = k
		}
	}
	if g < 0 || r < 0 || w.Kind != "world" || w.World < 0 || w.World >= 32 || len(w.Trace) != 2400 || len(w.ShapeTrace) != 2400 || len(w.Snapshots) != 25 || len(w.ShapeSnapshots) != 5 {
		return fmt.Errorf("world domain")
	}
	expected := makeV35(seed, g, r, w.World)
	if expected.Seed != w.Seed || !reflect.DeepEqual(expected.Base, w.Base) || !reflect.DeepEqual(expected.Rates, w.Rates) || !reflect.DeepEqual(expected.Outcomes, w.Outcomes) {
		return fmt.Errorf("generated population")
	}
	if err := runV34(&expected.worldV34, 2400); err != nil {
		return err
	}
	if err := runShapeV35(&expected); err != nil {
		return err
	}
	for j := range expected.Snapshots {
		scoreV34(&expected.Snapshots[j], expected.Rates)
	}
	for j := range expected.ShapeSnapshots {
		scoreV34(&expected.ShapeSnapshots[j].snapV34, expected.Rates)
	}
	a, b := w, expected
	a.Snapshots = append([]snapV34(nil), w.Snapshots...)
	a.ShapeSnapshots = append([]shapeSnapshotV35(nil), w.ShapeSnapshots...)
	b.Snapshots = append([]snapV34(nil), expected.Snapshots...)
	b.ShapeSnapshots = append([]shapeSnapshotV35(nil), expected.ShapeSnapshots...)
	clearV35(&a)
	clearV35(&b)
	if !reflect.DeepEqual(a, b) {
		return fmt.Errorf("non-timing seed replay")
	}
	near := func(a, b float64) bool { return !math.IsNaN(a) && math.Abs(a-b) <= 3e-9 }
	refs := make([]*Model, 3)
	for j, mode := range modesV34[:3] {
		refs[j], _ = New(w.Base, mode)
	}
	shape, _ := NewShape(w.Base)
	n, s := make([]uint16, 150), make([]uint16, 150)
	seen := map[[2]int]bool{}
	previous := map[string]costV34{}
	for step, tr := range w.Trace {
		ordinal, position := step/150+1, step%150
		key := [2]int{tr.Index, ordinal}
		if tr.Index < 0 || tr.Index >= 150 || tr.Ordinal != ordinal || seen[key] || !near(tr.Probability, 1/float64(150-position)) || tr.Useful != w.Outcomes[ordinal-1][tr.Index] {
			return fmt.Errorf("trial eligibility/probability/outcome")
		}
		seen[key] = true
		n[tr.Index]++
		if tr.Useful {
			s[tr.Index]++
		}
		cut := step + 1
		for j, budget := range cutsV34 {
			if cut != budget {
				continue
			}
			shape.n, shape.success = n, s
			shapeWeights, shapeLaw := shapeReference(shape)
			var kernelWeights [ShapeKernels]float64
			for z, weight := range shapeWeights {
				kernelWeights[z%ShapeKernels] += weight
			}
			ss := w.ShapeSnapshots[j]
			if ss.Model != "shape" || ss.Budget != cut {
				return fmt.Errorf("candidate snapshot domain")
			}
			for k, weight := range kernelWeights {
				if !near(weight, ss.Shapes[k]) {
					return fmt.Errorf("batch shape marginal")
				}
			}
			for k, mode := range append(append([]string(nil), modesV34...), "shape") {
				var law []float64
				var snapshot snapV34
				if mode == "shape" {
					law, snapshot = shapeLaw, ss.snapV34
				} else {
					snapshot = w.Snapshots[j*5+k]
					if snapshot.Model != mode || snapshot.Budget != cut {
						return fmt.Errorf("control snapshot domain")
					}
					if k < 3 {
						refs[k].n, refs[k].success = n, s
						weights, forecasts := reference(refs[k], mode)
						law = forecasts
						var marginal [Strengths]float64
						for z, weight := range weights {
							marginal[z%Strengths] += weight
						}
						for z := range marginal {
							if !near(marginal[z], snapshot.Dispersion[z]) {
								return fmt.Errorf("batch control marginal")
							}
						}
					} else {
						law = make([]float64, 150)
						for i, p := range w.Base {
							law[i] = p
							if mode == "local" {
								law[i] = (2*p + float64(s[i])) / (2 + float64(n[i]))
							}
						}
					}
				}
				if len(snapshot.Forecast) != 150 || len(snapshot.Packet) != 10 {
					return fmt.Errorf("law dimensions")
				}
				for i, q := range law {
					if !near(q, snapshot.Forecast[i]) {
						return fmt.Errorf("batch integral law")
					}
				}
				if err := scoreAuditV35(snapshot, w.Rates); err != nil {
					return err
				}
				c, old := snapshot.Costs, previous[mode]
				if c.SetupNS < 0 || c.ProbeNS < 0 || c.UpdateNS < 0 || c.SnapshotNS < 0 || c.EarlierSnapshotsNS < 0 || c.AccountedNS != c.SetupNS+c.ProbeNS+c.UpdateNS+c.SnapshotNS+c.EarlierSnapshotsNS || c.ProbeNS < old.ProbeNS || c.UpdateNS < old.UpdateNS || cut != 32 && (c.SetupNS != old.SetupNS || c.EarlierSnapshotsNS != old.EarlierSnapshotsNS+old.SnapshotNS) || cut == 32 && c.EarlierSnapshotsNS != 0 {
					return fmt.Errorf("phase accounting")
				}
				previous[mode] = c
			}
		}
	}
	oldWork := w.NominationNS
	for mode, c := range previous {
		if mode != "shape" {
			oldWork += c.AccountedNS
		}
	}
	if w.NominationNS < 0 || w.ElapsedNS < oldWork || w.ShapeElapsedNS < previous["shape"].AccountedNS {
		return fmt.Errorf("elapsed accounting")
	}
	return nil
}

func scoreAuditV35(s snapV34, rates []float64) error {
	order := make([]int, 150)
	var risk, priority, useful, packetRisk, bias float64
	for i, p := range rates {
		order[i] = i
		q := s.Forecast[i]
		v := q*q - 2*q*p + p
		risk += v
		priority += v
		if i < 10 {
			priority += 2 * v
		}
	}
	sort.SliceStable(order, func(i, j int) bool { return s.Forecast[order[i]] > s.Forecast[order[j]] })
	if !reflect.DeepEqual(order[:10], s.Packet) {
		return fmt.Errorf("packet ordering")
	}
	for _, i := range s.Packet {
		p, q := rates[i], s.Forecast[i]
		useful += p
		packetRisk += q*q - 2*q*p + p
		bias += q - p
	}
	for _, pair := range [][2]float64{{risk / 150, s.Brier}, {priority / 170, s.PriorityBrier}, {useful / 10, s.PacketUsefulness}, {packetRisk / 10, s.PacketBrier}, {bias / 10, s.PacketBias}} {
		if math.IsNaN(pair[1]) || math.Abs(pair[0]-pair[1]) > 3e-10 {
			return fmt.Errorf("future score denominator")
		}
	}
	return nil
}

func TestStudyAuditV35(t *testing.T) {
	input := os.Getenv("EVENTFRAME_SHAPE_V35_AUDIT")
	if input == "" {
		t.Skip("explicit raw input required")
	}
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var manifest manifestV34
	if err := dec.Decode(&manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Kind != "manifest" || manifest.Worlds != 1024 || manifest.Split != "design" && manifest.Split != "confirmation" || manifest.SeedBase != map[string]int64{"design": 2026103503, "confirmation": 2026103504}[manifest.Split] {
		t.Fatal("manifest")
	}
	if err := shapeSourcesV35(rootV34(t), manifest.Sources); err != nil {
		t.Fatal(err)
	}
	groups := map[string][][]snapV34{}
	shapeByCell := map[string][][5][ShapeKernels]float64{}
	ids, seeds := map[string]bool{}, map[int64]bool{}
	for count := 0; count < manifest.Worlds; count++ {
		var w worldV35
		if err := dec.Decode(&w); err != nil {
			t.Fatal(err)
		}
		if err := auditShapeV35(w, manifest.SeedBase); err != nil {
			t.Fatal(count, err)
		}
		name := w.Geometry + "/" + w.Regime
		id := fmt.Sprintf("%s/%d", name, w.World)
		if ids[id] || seeds[w.Seed] {
			t.Fatal("duplicate world/cross-cell seed")
		}
		ids[id], seeds[w.Seed] = true, true
		var masses [5][ShapeKernels]float64
		for j, ss := range w.ShapeSnapshots {
			w.Snapshots = append(w.Snapshots, ss.snapV34)
			masses[j] = ss.Shapes
		}
		groups[name] = append(groups[name], w.Snapshots)
		shapeByCell[name] = append(shapeByCell[name], masses)
	}
	var extra worldV35
	if err := dec.Decode(&extra); err != io.EOF {
		t.Fatal("extra record", err)
	}
	if len(groups) != 32 {
		t.Fatal("missing cells")
	}
	get := func(world []snapV34, model string, cut int) snapV34 {
		for _, s := range world {
			if s.Model == model && s.Budget == cut {
				return s
			}
		}
		panic("missing snapshot")
	}
	fields := map[string]func(snapV34) float64{"Brier": func(s snapV34) float64 { return s.Brier }, "PriorityBrier": func(s snapV34) float64 { return s.PriorityBrier }, "PacketUsefulness": func(s snapV34) float64 { return s.PacketUsefulness }, "PacketBrier": func(s snapV34) float64 { return s.PacketBrier }, "PacketBias": func(s snapV34) float64 { return s.PacketBias }}
	stats, gates := map[string]any{}, map[string]any{}
	benchmark, err := os.ReadFile(filepath.Join(rootV34(t), "docs/experiments/mmm-shape-v35-benchmarks.txt"))
	if err != nil {
		t.Fatal(err)
	}
	allocationMax, allocationRows := int64(0), 0
	for _, line := range strings.Split(string(benchmark), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "BenchmarkShapeConstruct150-10" {
			continue
		}
		if len(fields) != 8 || fields[5] != "B/op" {
			t.Fatal("construction benchmark format")
		}
		value, err := strconv.ParseInt(fields[4], 10, 64)
		if err != nil || value < 0 {
			t.Fatal("construction benchmark value", err)
		}
		allocationMax = max(allocationMax, value)
		allocationRows++
	}
	if allocationRows != 3 {
		t.Fatal("missing construction measurements")
	}
	allocationPass := allocationMax <= 1<<20
	allPass, failed := true, 0
	for name, worlds := range groups {
		if len(worlds) != 32 {
			t.Fatal("cell trajectory count")
		}
		cell := map[int]any{}
		maximum := int64(0)
		for _, cut := range cutsV34 {
			byModel := map[string]any{}
			for _, model := range append(append([]string(nil), modesV34...), "shape") {
				row := map[string]any{}
				maxCost := int64(0)
				for field, value := range fields {
					x := make([]float64, 32)
					for j, world := range worlds {
						x[j] = value(get(world, model, cut))
					}
					row[field] = interval34(x)
				}
				for _, world := range worlds {
					maxCost = max(maxCost, get(world, model, cut).Costs.AccountedNS)
				}
				maximum = max(maximum, maxCost)
				row["MaximumAccountedNS"] = maxCost
				if model == "shape" {
					var mean [ShapeKernels]float64
					for _, weights := range shapeByCell[name] {
						for k, value := range weights[indexCutV34(cut)] {
							mean[k] += value / 32
						}
					}
					row["MeanShapes"] = mean
				}
				byModel[model] = row
			}
			cell[cut] = byModel
		}
		stats[name] = cell
		checks, pass := map[string]any{}, allocationPass
		checks["Allocation"] = map[string]any{"MaximumBytes": allocationMax, "LimitBytes": 1 << 20, "Pass": allocationPass}
		for _, control := range []string{"local", "fixed2", "adaptive"} {
			for _, field := range []string{"Brier", "PriorityBrier", "PacketUsefulness"} {
				x := make([]float64, 32)
				for j, world := range worlds {
					a, b := get(world, "shape", 2400), get(world, control, 2400)
					x[j] = fields[field](b) - fields[field](a)
					if field == "PacketUsefulness" {
						x[j] = -x[j]
					}
				}
				v := interval34(x)
				improve := false
				for _, r := range []string{"aligned", "reversed", "calibrated", "mean_shared"} {
					improve = improve || name == "tight/"+r || name == "wide/"+r
				}
				improve = improve && control == "fixed2" && field != "PacketUsefulness"
				independent := control == "adaptive" && (name == "tight/independent" || name == "wide/independent") && field != "PacketUsefulness"
				ok := v.Lower >= -.01
				if improve {
					ok = v.Mean >= .005 && v.Lower > 0
				} else if independent {
					ok = v.Lower > 0
				}
				checks[control+"/"+field] = map[string]any{"Interval": v, "Improve005": improve, "IndependentGain": independent, "Pass": ok}
				pass = pass && ok
			}
		}
		x := make([]float64, 32)
		for j, world := range worlds {
			x[j] = get(world, "shape", 2400).PacketBias
		}
		bias := interval34(x)
		upper, limit := math.Abs(bias.Mean)+3.5*bias.SE, .10
		if name == "tight/independent" || name == "wide/independent" {
			limit = .05
		}
		checks["Bias"] = map[string]any{"Interval": bias, "MagnitudeUpper": upper, "Limit": limit, "Pass": upper <= limit}
		checks["Work"] = map[string]any{"MaximumNS": maximum, "LimitNS": 50000000, "Pass": maximum <= 50000000}
		pass = pass && upper <= limit && maximum <= 50000000
		gates[name] = map[string]any{"Checks": checks, "Pass": pass}
		allPass = allPass && pass
		if !pass {
			failed++
		}
	}
	raw, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	auditor, err := os.ReadFile(filepath.Join(rootV34(t), "internal/researchdispersion/shape_audit_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	result := map[string]any{"Study": "shape-v35", "Split": manifest.Split, "SHA256": hashV34(raw), "SourceHashes": manifest.Sources, "PostCollectionAuditorSHA256": hashV34(auditor), "PreCollectionBenchmarkSHA256": hashV34(benchmark), "Worlds": 1024, "Snapshots": 30720, "WholeGoalsComplete": false, "Goal7EqualTotalCost": false, "Groups": stats, "Gates": gates, "ComponentPass": allPass, "FailedCells": failed}
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(rootV34(t), "docs/experiments/mmm-shape-v35-"+manifest.Split+"-audit.json")
	if err := os.WriteFile(out, append(b, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("verified1024 worlds/30720 snapshots/2457600 issued steps; pass=%v failed=%d/32", allPass, failed)
}

func TestShapeAuditorNegativeV35(t *testing.T) {
	hashes := map[string]string{}
	for _, p := range filesV35 {
		b, err := os.ReadFile(filepath.Join(rootV34(t), p))
		if err != nil {
			t.Fatal(err)
		}
		hashes[p] = hashV34(b)
	}
	if err := shapeSourcesV35(rootV34(t), hashes); err != nil {
		t.Fatal("valid source control", err)
	}
	hashes[filesV35[3]] = "corrupted"
	if err := shapeSourcesV35(rootV34(t), hashes); err == nil {
		t.Fatal("corrupted source accepted")
	}
	w := makeV35(2026103599, 1, 13, 0)
	if err := runV34(&w.worldV34, 2400); err != nil {
		t.Fatal(err)
	}
	if err := runShapeV35(&w); err != nil {
		t.Fatal(err)
	}
	for j := range w.Snapshots {
		scoreV34(&w.Snapshots[j], w.Rates)
	}
	for j := range w.ShapeSnapshots {
		scoreV34(&w.ShapeSnapshots[j].snapV34, w.Rates)
	}
	if err := auditShapeV35(w, 2026103599); err != nil {
		t.Fatal("valid tape control", err)
	}
	for name, mutate := range map[string]func(*worldV35){
		"issued":     func(w *worldV35) { w.ShapeTrace[0] += .1 },
		"law":        func(w *worldV35) { w.ShapeSnapshots[0].Forecast[0] += .1 },
		"weights":    func(w *worldV35) { w.ShapeSnapshots[0].Shapes[0] += .1 },
		"outcome":    func(w *worldV35) { w.Trace[0].Useful = !w.Trace[0].Useful },
		"cost":       func(w *worldV35) { w.ShapeSnapshots[0].Costs.AccountedNS++ },
		"nomination": func(w *worldV35) { w.Trace[0].Probability += .1 },
	} {
		t.Run(name, func(t *testing.T) {
			b, err := json.Marshal(w)
			if err != nil {
				t.Fatal(err)
			}
			var copy worldV35
			if err := json.Unmarshal(b, &copy); err != nil {
				t.Fatal(err)
			}
			mutate(&copy)
			if err := auditShapeV35(copy, 2026103599); err == nil {
				t.Fatal("corrupted tape accepted")
			}
		})
	}
}
