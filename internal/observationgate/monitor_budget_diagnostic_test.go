package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/bits"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
)

func monitorBudgetSelected(origin int, audit bool) bool { return origin%2 == 0 || audit }

func monitorBudgetRead(x uint16) (uint16, uint16, int, error) {
	r := observationexperiment.Frames(x, "budget-counterfactual")
	var mask, value uint16
	cost := 0
	for scope := 0; scope < 3; scope++ {
		m, v, err := r.Read(observation.View{Scope: scope, Depth: 2})
		if err != nil {
			return 0, 0, 0, err
		}
		cost += bits.OnesCount16(m &^ mask)
		mask |= m
		value |= v
	}
	return mask, value, cost, nil
}

type monitorBudgetRecord struct {
	Phase, Case, Schedule                                                          string
	Index, OriginalCost, CandidateCost, OriginalSplit, SelectedPairs, ArrivedPairs int
	Candidate                                                                      fullAuditGateArm
	Masks                                                                          [][2]uint16
}

func monitorBudgetReplay(t testing.TB, r arrivalSwitchRecord) []monitorBudgetRecord {
	t.Helper()
	phase := 0
	if r.Phase == "cohort2" {
		phase = 1
	}
	scenario := -1
	for i, n := range arrivalSwitchNames {
		if n == r.Case {
			scenario = i
		}
	}
	if scenario < 0 {
		t.Fatal("unknown case")
	}
	base, masks, err := arrivalSwitchSetup(phase, scenario, r.Index)
	if err != nil || masks != r.Masks {
		t.Fatal("setup", err)
	}
	var result []monitorBudgetRecord
	for schedule, s := range []innerArrivalResult{r.Immediate, r.Delayed} {
		z := monitorBudgetRecord{Phase: r.Phase, Case: r.Case, Schedule: []string{"Immediate", "Delayed"}[schedule], Index: r.Index, OriginalCost: s.MonitorCost + s.AuditCost, OriginalSplit: s.Arms[2].SplitAt, Candidate: fullAuditGateArm{FirstAllow: -1, FirstNomination: -1, SplitAt: -1}, Masks: make([][2]uint16, 512)}
		var predictions [512][2]float64
		// Frozen forecasts require no replay-future labels. Charge acquisition on
		// nomination even for missing feedback; gate updates happen only on release.
		for i, f := range s.Frames {
			if !monitorBudgetSelected(i, f.Audit) {
				continue
			}
			z.SelectedPairs++
			for side, x := range []uint16{f.X, f.RX} {
				m, v, c, err := monitorBudgetRead(x)
				if err != nil || m != 511 || v != x || c != 9 {
					t.Fatal("read contract", err)
				}
				z.Masks[i][side] = m
				z.CandidateCost += c
				predictions[i][side], err = base.ForecastObserved(m, v)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
		var g pairedInvestigator
		var mon observationrescue.Monitor
		head := 0
		var released []int
		for clock := 0; clock < 544; clock++ {
			for head < 512 && head <= clock {
				f := s.Frames[head]
				if f.Missing {
					if head >= max(0, clock-31) {
						break
					}
					head++
					continue
				}
				if f.Arrival > clock {
					break
				}
				released = append(released, head)
				if monitorBudgetSelected(head, f.Audit) {
					z.ArrivedPairs++
					lc, rc := (predictions[head][0] >= .5) == f.Y, (predictions[head][1] >= .5) == f.RY
					rev, cp := mon.Observe(rc, lc, true)
					nom := bayes.RevisionSplits(rev.Action)
					allow, err := g.observe(g.next, rc, lc)
					if err != nil {
						t.Fatal(err)
					}
					a := &z.Candidate
					if allow && a.FirstAllow < 0 {
						a.FirstAllow = clock
					}
					if nom && a.FirstNomination < 0 {
						a.FirstNomination = clock
					}
					if allow && nom && a.SplitAt < 0 {
						a.SplitAt = clock
					}
					a.Trace = append(a.Trace, switchGateTrace{clock, head, rc, lc, allow, nom, cp, g.evidence.logWealth()})
				}
				head++
			}
		}
		if len(released) != len(s.Released) {
			t.Fatal("release length")
		}
		for i, id := range released {
			if id != s.Released[i] {
				t.Fatal("release drift")
			}
		}
		if head != 512 || z.CandidateCost != 18*z.SelectedPairs || len(z.Candidate.Trace) != z.ArrivedPairs {
			t.Fatal("accounting")
		}
		result = append(result, z)
	}
	return result
}

func TestMonitorBudgetContracts(t *testing.T) {
	for x := uint16(0); x < 512; x++ {
		m, v, c, e := monitorBudgetRead(x)
		if e != nil || m != 511 || v != x || c != 9 {
			t.Fatal("reader", x, e)
		}
	}
	for i := 0; i < 512; i++ {
		if !monitorBudgetSelected(i, true) || monitorBudgetSelected(i, false) != (i%2 == 0) {
			t.Fatal("nomination")
		}
	}
}

func TestMonitorBudgetDiagnostic(t *testing.T) {
	in, out := os.Getenv("EVENTFRAME_MONITOR_BUDGET_INPUT"), os.Getenv("EVENTFRAME_MONITOR_BUDGET_OUTPUT")
	if in == "" || out == "" {
		t.Skip("opt-in budget diagnostic")
	}
	raw, err := os.ReadFile(in)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(raw)
	f, err := os.Open(in)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var header json.RawMessage
	if err := d.Decode(&header); err != nil {
		t.Fatal(err)
	}
	var records []monitorBudgetRecord
	for {
		var r arrivalSwitchRecord
		if err := d.Decode(&r); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		records = append(records, monitorBudgetReplay(t, r)...)
	}
	if len(records) != 256 {
		t.Fatal("incomplete")
	}
	paths, err := filepath.Glob("../../internal/*/*.go")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, "../../go.mod", "../../go.sum", "../../docs/experiments/mmm-monitor-budget-v1-contract.md", "../../research/monitor-budget-summary.mjs")
	sources, hashes := map[string]string{}, map[string]string{}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		v := sha256.Sum256(b)
		sources[p[6:]] = string(b)
		hashes[p[6:]] = hex.EncodeToString(v[:])
	}
	o, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()
	if err := json.NewEncoder(o).Encode(map[string]any{"ParentSHA256": hex.EncodeToString(h[:]), "Sources": sources, "Hashes": hashes, "Records": records, "Consumed": true}); err != nil {
		t.Fatal(err)
	}
}
