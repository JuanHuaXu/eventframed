package researchdispersion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func auditV34(w worldV34, seed int64) error {
	g, r := -1, -1
	for j, v := range []string{"tight", "wide"} {
		if w.Geometry == v {
			g = j
		}
	}
	for j, v := range regimesV34 {
		if w.Regime == v {
			r = j
		}
	}
	if g < 0 || r < 0 || w.Kind != "world" || w.World < 0 || w.World >= 32 || len(w.Trace) != 2400 || len(w.Snapshots) != 25 {
		return fmt.Errorf("world domain")
	}
	expected := makeV34(seed, g, r, w.World)
	if expected.Seed != w.Seed || !reflect.DeepEqual(expected.Base, w.Base) || !reflect.DeepEqual(expected.Rates, w.Rates) || !reflect.DeepEqual(expected.Outcomes, w.Outcomes) {
		return fmt.Errorf("generated population")
	}
	// Exact RNG replay covers every issued step, not just final predictions.
	if err := runV34(&expected, 2400); err != nil {
		return err
	}
	for j := range expected.Snapshots {
		scoreV34(&expected.Snapshots[j], expected.Rates)
	}
	a, b := w, expected
	a.Snapshots = append([]snapV34(nil), w.Snapshots...)
	b.Snapshots = append([]snapV34(nil), expected.Snapshots...)
	clearV34(&a)
	clearV34(&b)
	if !reflect.DeepEqual(a, b) {
		return fmt.Errorf("non-timing seed replay")
	}
	near := func(got, want float64) bool { return !math.IsNaN(got) && math.Abs(got-want) <= 2e-9 }
	n, s := make([]uint16, 150), make([]uint16, 150)
	refs := make([]*Model, 3)
	for j, mode := range modesV34[:3] {
		refs[j], _ = New(w.Base, mode)
	}
	previous := map[string]costV34{}
	seen := map[[2]int]bool{}
	for step, tr := range w.Trace {
		position, ordinal := step%150, step/150+1
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
		for _, budget := range cutsV34 {
			if cut != budget {
				continue
			}
			for j, mode := range modesV34 {
				ss := w.Snapshots[indexCutV34(cut)*5+j]
				if ss.Model != mode || ss.Budget != cut || len(ss.Forecast) != 150 || len(ss.Packet) != 10 {
					return fmt.Errorf("snapshot domain")
				}
				q := make([]float64, 150)
				var dispersion [Strengths]float64
				if j < 3 {
					refs[j].n, refs[j].success = n, s
					weights, forecasts := reference(refs[j], mode)
					q = forecasts
					for z, mass := range weights {
						dispersion[z%Strengths] += mass
					}
				} else {
					for i, p := range w.Base {
						q[i] = p
						if mode == "local" {
							q[i] = (2*p + float64(s[i])) / (2 + float64(n[i]))
						}
					}
				}
				for i := range q {
					if !near(q[i], ss.Forecast[i]) {
						return fmt.Errorf("batch beta-integral law")
					}
				}
				for k := range dispersion {
					if !near(dispersion[k], ss.Dispersion[k]) {
						return fmt.Errorf("batch dispersion marginal")
					}
				}
				// Independently normalize future risk and sort the packet.
				order := make([]int, 150)
				for i := range order {
					order[i] = i
				}
				// Stable sorting uses the recorded law to avoid tolerance-sized
				// reconstruction noise reordering theoretically tied members.
				sort.SliceStable(order, func(i, j int) bool { return ss.Forecast[order[i]] > ss.Forecast[order[j]] })
				if !reflect.DeepEqual(order[:10], ss.Packet) {
					return fmt.Errorf("stable packet")
				}
				var risk, priority, useful, packetRisk, bias float64
				for i, p := range w.Rates {
					forecast := ss.Forecast[i]
					v := forecast*forecast - 2*forecast*p + p
					risk += v
					priority += v
					if i < 10 {
						priority += 2 * v
					}
				}
				for _, i := range ss.Packet {
					p, forecast := w.Rates[i], ss.Forecast[i]
					useful += p
					packetRisk += forecast*forecast - 2*forecast*p + p
					bias += forecast - p
				}
				for _, pair := range [][2]float64{{risk / 150, ss.Brier}, {priority / 170, ss.PriorityBrier}, {useful / 10, ss.PacketUsefulness}, {packetRisk / 10, ss.PacketBrier}, {bias / 10, ss.PacketBias}} {
					if !near(pair[0], pair[1]) {
						return fmt.Errorf("future-risk denominator")
					}
				}
				c, old := ss.Costs, previous[mode]
				if c.SetupNS < 0 || c.ProbeNS < 0 || c.UpdateNS < 0 || c.SnapshotNS < 0 || c.EarlierSnapshotsNS < 0 || c.AccountedNS != c.SetupNS+c.ProbeNS+c.UpdateNS+c.SnapshotNS+c.EarlierSnapshotsNS || c.ProbeNS < old.ProbeNS || c.UpdateNS < old.UpdateNS || cut != 32 && (c.SetupNS != old.SetupNS || c.EarlierSnapshotsNS != old.EarlierSnapshotsNS+old.SnapshotNS) || cut == 32 && c.EarlierSnapshotsNS != 0 {
					return fmt.Errorf("phase accounting")
				}
				previous[mode] = c
			}
		}
	}
	work := w.NominationNS
	for _, c := range previous {
		work += c.AccountedNS
	}
	if w.NominationNS < 0 || w.ElapsedNS < work {
		return fmt.Errorf("world elapsed accounting")
	}
	return nil
}

func indexCutV34(cut int) int {
	for j, v := range cutsV34 {
		if cut == v {
			return j
		}
	}
	panic("unsupported checkpoint")
}

func sourcesV34(root string, hashes map[string]string) error {
	if len(hashes) != len(filesV34) {
		return fmt.Errorf("source count")
	}
	for _, p := range filesV34 {
		b, err := os.ReadFile(filepath.Join(root, p))
		if err != nil || hashV34(b) != hashes[p] {
			return fmt.Errorf("frozen source changed %s: %v", p, err)
		}
	}
	return nil
}

type intervalV34 struct {
	Mean, SE, Lower, Upper float64
}

func interval34(x []float64) intervalV34 {
	mean := 0.
	for _, v := range x {
		mean += v / float64(len(x))
	}
	variance := 0.
	for _, v := range x {
		variance += (v - mean) * (v - mean)
	}
	se := math.Sqrt(variance / float64(len(x)-1) / float64(len(x)))
	return intervalV34{mean, se, mean - 3.5*se, mean + 3.5*se}
}

func TestStudyAuditV34(t *testing.T) {
	input := os.Getenv("EVENTFRAME_DISPERSION_V34_AUDIT")
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
	if manifest.Kind != "manifest" || manifest.Worlds != 768 || len(manifest.Sources) != len(filesV34) || manifest.Split != "design" && manifest.Split != "confirmation" || manifest.SeedBase != map[string]int64{"design": 2026103403, "confirmation": 2026103404}[manifest.Split] {
		t.Fatal("manifest")
	}
	if err := sourcesV34(rootV34(t), manifest.Sources); err != nil {
		t.Fatal(err)
	}
	groups := map[string][][]snapV34{}
	ids, seeds := map[string]bool{}, map[int64]bool{}
	for k := 0; k < manifest.Worlds; k++ {
		var w worldV34
		if err := dec.Decode(&w); err != nil {
			t.Fatal(err)
		}
		if err := auditV34(w, manifest.SeedBase); err != nil {
			t.Fatal(k, err)
		}
		id := fmt.Sprintf("%s/%s/%d", w.Geometry, w.Regime, w.World)
		if ids[id] || seeds[w.Seed] {
			t.Fatal("duplicate world or cross-cell seed")
		}
		ids[id], seeds[w.Seed] = true, true
		name := w.Geometry + "/" + w.Regime
		groups[name] = append(groups[name], w.Snapshots)
	}
	var extra worldV34
	if err := dec.Decode(&extra); err != io.EOF {
		t.Fatal("extra record", err)
	}
	if len(groups) != 24 {
		t.Fatal("missing cells")
	}
	get := func(v []snapV34, model string, cut int) snapV34 {
		for _, s := range v {
			if s.Model == model && s.Budget == cut {
				return s
			}
		}
		panic("missing snapshot")
	}
	fields := map[string]func(snapV34) float64{"Brier": func(s snapV34) float64 { return s.Brier }, "PriorityBrier": func(s snapV34) float64 { return s.PriorityBrier }, "PacketUsefulness": func(s snapV34) float64 { return s.PacketUsefulness }, "PacketBrier": func(s snapV34) float64 { return s.PacketBrier }, "PacketBias": func(s snapV34) float64 { return s.PacketBias }}
	result := map[string]any{"Study": "dispersion-v34", "Split": manifest.Split, "SourceHashes": manifest.Sources, "Worlds": 768, "WholeGoalsComplete": false, "Goal7EqualTotalCost": false}
	stats, gates := map[string]any{}, map[string]any{}
	allPass, failed := true, 0
	for name, worlds := range groups {
		if len(worlds) != 32 {
			t.Fatal("cell trajectories")
		}
		cell := map[int]any{}
		maximum := int64(0)
		for _, cut := range cutsV34 {
			byModel := map[string]any{}
			for _, model := range modesV34 {
				row := map[string]any{}
				var weights [Strengths]float64
				maxCost := int64(0)
				for field, value := range fields {
					x := make([]float64, 32)
					for j, world := range worlds {
						x[j] = value(get(world, model, cut))
					}
					row[field] = interval34(x)
				}
				for _, world := range worlds {
					s := get(world, model, cut)
					for k := range weights {
						weights[k] += s.Dispersion[k] / 32
					}
					maxCost = max(maxCost, s.Costs.AccountedNS)
				}
				maximum = max(maximum, maxCost)
				row["MeanDispersion"], row["MaximumAccountedNS"] = weights, maxCost
				byModel[model] = row
			}
			cell[cut] = byModel
		}
		stats[name] = cell
		checks := map[string]any{}
		pass := true
		for _, control := range []string{"local", "fixed2"} {
			for _, field := range []string{"Brier", "PriorityBrier", "PacketUsefulness"} {
				x := make([]float64, 32)
				for j, world := range worlds {
					a, b := get(world, "adaptive", 2400), get(world, control, 2400)
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
				ok := v.Lower >= -.01
				if improve {
					ok = v.Mean >= .005 && v.Lower > 0
				}
				checks[control+"/"+field] = map[string]any{"Interval": v, "Improve": improve, "Pass": ok}
				pass = pass && ok
			}
		}
		x := make([]float64, 32)
		for j, world := range worlds {
			x[j] = get(world, "adaptive", 2400).PacketBias
		}
		bias := interval34(x)
		upper := math.Abs(bias.Mean) + 3.5*bias.SE
		checks["Bias"] = map[string]any{"Interval": bias, "MagnitudeUpper": upper, "Pass": upper <= .10}
		checks["Work"] = map[string]any{"MaximumNS": maximum, "Pass": maximum <= 25000000}
		pass = pass && upper <= .10 && maximum <= 25000000
		gates[name] = map[string]any{"Checks": checks, "Pass": pass}
		allPass = allPass && pass
		if !pass {
			failed++
		}
	}
	result["Groups"], result["Gates"], result["ComponentPass"], result["FailedCells"] = stats, gates, allPass, failed
	raw, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	result["SHA256"] = hashV34(raw)
	auditSource, err := os.ReadFile(filepath.Join(rootV34(t), "internal/researchdispersion/audit_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	result["PostCollectionAuditorSHA256"] = hashV34(auditSource)
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(rootV34(t), "docs/experiments/mmm-dispersion-v34-"+manifest.Split+"-audit.json")
	if err := os.WriteFile(output, append(b, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("verified %d worlds / %d snapshots / %d issued steps; component pass=%v failed=%d/24", manifest.Worlds, manifest.Worlds*25, manifest.Worlds*2400, allPass, failed)
}

func TestAuditCorruptionControlsV34(t *testing.T) {
	hashes := map[string]string{}
	for _, p := range filesV34 {
		b, err := os.ReadFile(filepath.Join(rootV34(t), p))
		if err != nil {
			t.Fatal(err)
		}
		hashes[p] = hashV34(b)
	}
	if err := sourcesV34(rootV34(t), hashes); err != nil {
		t.Fatal("valid source control", err)
	}
	hashes[filesV34[0]] = "invalid"
	if err := sourcesV34(rootV34(t), hashes); err == nil {
		t.Fatal("corrupted source hash accepted")
	}
	w := makeV34(2026103499, 0, 3, 0)
	if err := runV34(&w, 2400); err != nil {
		t.Fatal(err)
	}
	for j := range w.Snapshots {
		scoreV34(&w.Snapshots[j], w.Rates)
	}
	if err := auditV34(w, 2026103499); err != nil {
		t.Fatal("valid control rejected", err)
	}
	for name, mutate := range map[string]func(*worldV34){
		"issued":     func(w *worldV34) { w.Trace[0].Q[0] += .1 },
		"outcome":    func(w *worldV34) { w.Trace[0].Useful = !w.Trace[0].Useful },
		"ordinal":    func(w *worldV34) { w.Trace[0].Ordinal++ },
		"law":        func(w *worldV34) { w.Snapshots[0].Forecast[0] += .1 },
		"dispersion": func(w *worldV34) { w.Snapshots[0].Dispersion[0] += .1 },
		"cost":       func(w *worldV34) { w.Snapshots[0].Costs.AccountedNS++ },
	} {
		t.Run(name, func(t *testing.T) {
			b, _ := json.Marshal(w)
			var copy worldV34
			if err := json.NewDecoder(bytes.NewReader(b)).Decode(&copy); err != nil {
				t.Fatal(err)
			}
			mutate(&copy)
			if err := auditV34(copy, 2026103499); err == nil {
				t.Fatal("corrupted tape accepted")
			}
		})
	}
}
