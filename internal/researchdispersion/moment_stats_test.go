package researchdispersion

import (
	"fmt"
	"strconv"
	"strings"
)

func allocationMomentV41(b []byte) (int64, error) {
	maximum, count := int64(0), 0
	for _, line := range strings.Split(string(b), "\n") {
		v := strings.Fields(line)
		if len(v) == 0 || !strings.HasPrefix(v[0], "BenchmarkMomentConstructor150-") {
			continue
		}
		if len(v) != 8 || v[5] != "B/op" {
			return 0, fmt.Errorf("moment allocation row")
		}
		n, e := strconv.ParseInt(v[4], 10, 64)
		if e != nil || n < 0 {
			return 0, fmt.Errorf("moment allocation value")
		}
		maximum = max(maximum, n)
		count++
	}
	if count != 3 {
		return 0, fmt.Errorf("moment allocation coverage %d", count)
	}
	return maximum, nil
}
func reportMomentV41(groups map[string][]orientationWorldV38, worlds int, allocation int64) map[string]any {
	fields := map[string]func(windowArmV37) float64{"IssuedBrier": func(a windowArmV37) float64 { return a.IssuedBrier }, "IssuedPriority": func(a windowArmV37) float64 { return a.IssuedPriority }, "Recovery": func(a windowArmV37) float64 { return a.Recovery }, "FinalBrier": func(a windowArmV37) float64 { return windowAtV37(a, -1).Brier }, "FinalPriority": func(a windowArmV37) float64 { return windowAtV37(a, -1).PriorityBrier }, "FinalUsefulness": func(a windowArmV37) float64 { return windowAtV37(a, -1).PacketUsefulness }}
	pairs := map[string][][2]int{"MomentMinusDensity": {{3, 4}, {5, 6}, {7, 8}, {9, 10}}, "RichMinusNarrow": {{3, 7}, {4, 8}, {5, 9}, {6, 10}}, "Strength2Minus4": {{5, 3}, {6, 4}, {9, 7}, {10, 8}}, "MomentOnlyStrength2Minus4": {{6, 4}, {10, 8}}}
	stats, gates, contrasts := map[string]any{}, map[string]any{}, map[string]any{}
	failures := map[string]int{}
	for _, mode := range modesMomentV41[3:] {
		failures[mode] = 0
	}
	for name, ws := range groups {
		if len(ws) != worlds {
			panic("moment statistics coverage")
		}
		for si, schedule := range schedulesV39 {
			key := name + "/" + schedule
			cell := map[string]any{}
			maximum := int64(0)
			for j, mode := range modesMomentV41 {
				row := map[string]any{}
				cost := int64(0)
				for field, metric := range fields {
					x := make([]float64, worlds)
					for k, w := range ws {
						a := w.Arms[si*11+j]
						x[k] = metric(a.windowArmV37)
						cost = max(cost, a.Costs.ElapsedNS)
					}
					row[field] = estimatePriorV40(x)
				}
				row["MaximumElapsedNS"] = cost
				maximum = max(maximum, cost)
				cell[mode] = row
			}
			stats[key] = cell
			factor := map[string]any{}
			for label, ps := range pairs {
				metrics := map[string]any{}
				for field, metric := range fields {
					x := make([]float64, worlds)
					for k, w := range ws {
						for _, p := range ps {
							gain := metric(w.Arms[si*11+p[0]].windowArmV37) - metric(w.Arms[si*11+p[1]].windowArmV37)
							if field == "FinalUsefulness" {
								gain = -gain
							}
							x[k] += gain / float64(len(ps))
						}
					}
					metrics[field] = estimatePriorV40(x)
				}
				factor[label] = metrics
			}
			contrasts[key] = factor
			if worlds == 1 {
				continue
			}
			stationary := false
			for _, r := range append(append([]string(nil), regimesV39[:4]...), "stationary_noise10") {
				stationary = stationary || strings.HasSuffix(name, "/"+r)
			}
			decisions := map[string]any{}
			for j := 3; j < 11; j++ {
				pass := maximum <= 400000000 && allocation <= 8<<20
				checks := map[string]any{"WorkPass": maximum <= 400000000, "AllocationPass": allocation <= 8<<20}
				for control := 0; control < 2; control++ {
					for _, field := range []string{"IssuedBrier", "IssuedPriority", "FinalBrier", "FinalPriority", "FinalUsefulness"} {
						x := make([]float64, worlds)
						for k, w := range ws {
							a, b := w.Arms[si*11+j], w.Arms[si*11+control]
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
						checks[modesMomentV41[control]+"/"+field] = map[string]any{"Gain": v, "Require001Improvement": improve, "Pass": ok}
					}
				}
				if len(ws[0].Population.Changes) > 0 {
					x, baseline := make([]float64, worlds), make([]float64, worlds)
					for k, w := range ws {
						a, b := w.Arms[si*11+j], w.Arms[si*11]
						x[k], baseline[k] = b.Recovery-a.Recovery, b.Recovery
					}
					v, control := interval34(x), interval34(baseline)
					ok := v.Lower > 0 && v.Mean >= .1*control.Mean
					pass = pass && ok
					checks["Recovery"] = map[string]any{"Gain": v, "Control": control, "Pass": ok}
				}
				if !pass {
					failures[modesMomentV41[j]]++
				}
				decisions[modesMomentV41[j]] = map[string]any{"Checks": checks, "Pass": pass}
			}
			gates[key] = decisions
		}
	}
	result := map[string]any{"Groups": stats, "FactorialContrasts": contrasts, "MaximumConstructorBytes": allocation, "CellsPerCandidate": 84, "WorldsPerCell": worlds, "QualityAdoptionEvaluated": worlds == 16}
	if worlds == 16 {
		result["Gates"], result["FailedCells"] = gates, failures
	} else {
		result["Gates"], result["FailedCells"], result["DiagnosticOnly"] = nil, nil, true
	}
	return result
}
