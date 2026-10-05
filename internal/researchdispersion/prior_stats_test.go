package researchdispersion

import (
	"fmt"
	"strconv"
	"strings"
)

func allocationPriorV40(b []byte) (int64, error) {
	maximum, count := int64(0), 0
	for _, line := range strings.Split(string(b), "\n") {
		v := strings.Fields(line)
		if len(v) == 0 || !strings.HasPrefix(v[0], "BenchmarkPriorConstructor150-") {
			continue
		}
		if len(v) != 8 || v[5] != "B/op" {
			return 0, fmt.Errorf("prior allocation row")
		}
		n, e := strconv.ParseInt(v[4], 10, 64)
		if e != nil || n < 0 {
			return 0, fmt.Errorf("prior allocation value")
		}
		maximum = max(maximum, n)
		count++
	}
	if count != 3 {
		return 0, fmt.Errorf("prior allocation coverage %d", count)
	}
	return maximum, nil
}
func estimatePriorV40(x []float64) any {
	if len(x) == 1 {
		return map[string]any{"Mean": x[0], "N": 1, "Interval": nil, "DiagnosticOnly": true}
	}
	return interval34(x)
}
func reportPriorV40(groups map[string][]orientationWorldV38, worlds int, allocation int64) map[string]any {
	fields := map[string]func(windowArmV37) float64{"IssuedBrier": func(a windowArmV37) float64 { return a.IssuedBrier }, "IssuedPriority": func(a windowArmV37) float64 { return a.IssuedPriority }, "Recovery": func(a windowArmV37) float64 { return a.Recovery }, "FinalBrier": func(a windowArmV37) float64 { return windowAtV37(a, -1).Brier }, "FinalPriority": func(a windowArmV37) float64 { return windowAtV37(a, -1).PriorityBrier }, "FinalUsefulness": func(a windowArmV37) float64 { return windowAtV37(a, -1).PacketUsefulness }}
	stats, gates, contrasts := map[string]any{}, map[string]any{}, map[string]any{}
	failures := map[string]int{}
	for _, mode := range modesPriorV40[3:] {
		failures[mode] = 0
	}
	for name, ws := range groups {
		if len(ws) != worlds {
			panic("prior statistics coverage")
		}
		for si, schedule := range schedulesV39 {
			key := name + "/" + schedule
			cell := map[string]any{}
			maximum := int64(0)
			for j, mode := range modesPriorV40 {
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
			for _, f := range []string{"SharedMinusPrivate", "InverseMinusRaw", "Strength2Minus4"} {
				metrics := map[string]any{}
				for field, metric := range fields {
					x := make([]float64, worlds)
					for k, w := range ws {
						for a := 3; a < 11; a++ {
							b := -1
							switch f {
							case "SharedMinusPrivate":
								if a%2 == 1 {
									b = a + 1
								}
							case "InverseMinusRaw":
								if a < 7 {
									b = a + 4
								}
							case "Strength2Minus4":
								if a == 5 || a == 6 || a == 9 || a == 10 {
									b = a - 2
								}
							}
							if b >= 0 {
								gain := metric(w.Arms[si*11+a].windowArmV37) - metric(w.Arms[si*11+b].windowArmV37)
								if field == "FinalUsefulness" {
									gain = -gain
								}
								x[k] += gain / 4
							}
						}
					}
					metrics[field] = estimatePriorV40(x)
				}
				factor[f] = metrics
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
						checks[modesPriorV40[control]+"/"+field] = map[string]any{"Gain": v, "Require001Improvement": improve, "Pass": ok}
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
					failures[modesPriorV40[j]]++
				}
				decisions[modesPriorV40[j]] = map[string]any{"Checks": checks, "Pass": pass}
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
