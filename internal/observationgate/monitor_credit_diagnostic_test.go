package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
)

type monitorCreditAcquisition struct {
	Bounded, Allowance, Charge, Credit int
	Audit, Full                        bool
}

type monitorCreditRecord struct {
	Acquisition                                                                    []monitorCreditAcquisition
	Values                                                                         [][2]uint16
	Phase, Case, Schedule                                                          string
	Index, OriginalCost, CandidateCost, OriginalSplit, SelectedPairs, ArrivedPairs int
	Candidate                                                                      fullAuditGateArm
	Masks                                                                          [][2]uint16
}

func monitorCreditReplay(t testing.TB, r arrivalSwitchRecord) []monitorCreditRecord {
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
	var result []monitorCreditRecord
	for schedule, s := range []innerArrivalResult{r.Immediate, r.Delayed} {
		z := monitorCreditRecord{Phase: r.Phase, Case: r.Case, Schedule: []string{"Immediate", "Delayed"}[schedule], Index: r.Index, OriginalCost: s.MonitorCost + s.AuditCost, OriginalSplit: s.Arms[2].SplitAt, Candidate: fullAuditGateArm{FirstAllow: -1, FirstNomination: -1, SplitAt: -1}, Masks: make([][2]uint16, 512)}
		var predictions [512][2]float64
		z.Values = make([][2]uint16, 512)
		credit, allowanceTotal := 0, 0
		for i, f := range s.Frames {
			readers := [2]*monitorCreditReader{
				newMonitorCreditReader(observationexperiment.Frames(f.X, "credit-live")),
				newMonitorCreditReader(observationexperiment.Frames(f.RX, "credit-reference")),
			}
			bounded := 0
			for side, reader := range readers {
				p, err := observation.Run(base, reader, 1, "mmm", 0)
				if err != nil {
					t.Fatal(err)
				}
				original := f.Baseline
				if side == 1 {
					original = f.Reference
				}
				if math.Abs(p.Probability-original) > 1e-12 || p.Cost != reader.charged {
					t.Fatal("bounded drift")
				}
				predictions[i][side] = p.Probability
				bounded += reader.charged
			}
			full, charge, next, err := monitorCreditStep(credit, bounded, f.Audit)
			if err != nil {
				t.Fatal(err)
			}
			if full {
				z.SelectedPairs++
				for side, reader := range readers {
					for scope := 0; scope < 3; scope++ {
						if _, _, err := reader.Read(observation.View{Scope: scope, Depth: 2}); err != nil {
							t.Fatal(err)
						}
					}
					if reader.mask != 511 {
						t.Fatal("incomplete full view")
					}
					predictions[i][side], err = base.ForecastObserved(reader.mask, reader.values)
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			actual := 0
			for side, reader := range readers {
				actual += reader.charged
				z.Masks[i][side] = reader.mask
				z.Values[i][side] = reader.values
			}
			if actual != charge {
				t.Fatal("charged versus observed mismatch")
			}
			allowance := bounded
			if f.Audit {
				allowance += 18
			}
			allowanceTotal += allowance
			z.CandidateCost += actual
			credit = next
			if z.CandidateCost > allowanceTotal || credit != allowanceTotal-z.CandidateCost {
				t.Fatal("prefix budget")
			}
			z.Acquisition = append(z.Acquisition, monitorCreditAcquisition{bounded, allowance, charge, credit, f.Audit, full})
		}
		if allowanceTotal != z.OriginalCost {
			t.Fatal("original total mismatch")
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
				{
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
		if head != 512 || len(z.Acquisition) != 512 || len(z.Candidate.Trace) != z.ArrivedPairs {
			t.Fatal("accounting")
		}
		result = append(result, z)
	}
	return result
}

func TestMonitorCreditDiagnostic(t *testing.T) {
	in, out := os.Getenv("EVENTFRAME_MONITOR_CREDIT_INPUT"), os.Getenv("EVENTFRAME_MONITOR_CREDIT_OUTPUT")
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
	var records []monitorCreditRecord
	for {
		var r arrivalSwitchRecord
		if err := d.Decode(&r); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		records = append(records, monitorCreditReplay(t, r)...)
	}
	if len(records) != 256 {
		t.Fatal("incomplete")
	}
	paths, err := filepath.Glob("../../internal/*/*.go")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, "../../go.mod", "../../go.sum", "../../docs/experiments/mmm-monitor-credit-v1-contract.md", "../../research/monitor-credit-summary.mjs")
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
