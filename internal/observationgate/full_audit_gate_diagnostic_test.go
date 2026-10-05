package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
)

type fullAuditGateArm struct {
	FirstAllow, FirstNomination, SplitAt int
	Trace                                []switchGateTrace
}
type fullAuditGateRecord struct {
	Phase, Case, Schedule         string
	Index                         int
	AccuracyBefore, AccuracyAfter [2]float64
	Arms                          [3]fullAuditGateArm
	AuditPairs, ArrivedAuditPairs int
}

func fullAuditGateReplay(t testing.TB, r arrivalSwitchRecord) []fullAuditGateRecord {
	t.Helper()
	phase := 0
	if r.Phase == "cohort2" {
		phase = 1
	}
	scenario := -1
	for i, name := range arrivalSwitchNames {
		if name == r.Case {
			scenario = i
		}
	}
	if scenario < 0 {
		t.Fatal("unknown case")
	}
	base, masks, err := arrivalSwitchSetup(phase, scenario, r.Index)
	if err != nil || masks != r.Masks {
		t.Fatal("base mismatch", err)
	}
	// Enumeration is evaluation only; runtime full-view calls below are audit-gated.
	var bounded [512]float64
	var before, after [2]float64
	for x := uint16(0); x < 512; x++ {
		v, err := observation.Run(base, observationexperiment.Frames(x, "full-audit-diagnostic"), 1, "mmm", 0)
		if err != nil {
			t.Fatal(err)
		}
		bounded[x] = v.Probability
		full, err := base.ForecastObserved(511, x)
		if err != nil {
			t.Fatal(err)
		}
		for a, p := range []float64{bounded[x], full} {
			if (p >= .5) == arrivalSwitchTruth(x, masks, scenario, 255, false) {
				before[a] += .95 / 512
			} else {
				before[a] += .05 / 512
			}
			if (p >= .5) == arrivalSwitchTruth(x, masks, scenario, 256, false) {
				after[a] += .95 / 512
			} else {
				after[a] += .05 / 512
			}
		}
	}
	var out []fullAuditGateRecord
	for schedule, s := range []innerArrivalResult{r.Immediate, r.Delayed} {
		z := fullAuditGateRecord{Phase: r.Phase, Case: r.Case, Schedule: []string{"Immediate", "Delayed"}[schedule], Index: r.Index, AccuracyBefore: before, AccuracyAfter: after}
		for a := range z.Arms {
			z.Arms[a] = fullAuditGateArm{FirstAllow: -1, FirstNomination: -1, SplitAt: -1}
		}
		var gs [3]pairedInvestigator
		var ms [3]observationrescue.Monitor
		var released []int
		for _, f := range s.Frames {
			if f.Audit {
				z.AuditPairs++
			}
		}
		head := 0
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
				if math.Abs(bounded[f.X]-f.Baseline) > 1e-12 || math.Abs(bounded[f.RX]-f.Reference) > 1e-12 {
					t.Fatal("original forecast drift")
				}
				if f.Audit {
					z.ArrivedAuditPairs++
				}
				for a := range gs {
					if a > 0 && !f.Audit {
						continue
					}
					lp, rp := f.Baseline, f.Reference
					if a == 2 {
						lp, err = base.ForecastObserved(511, f.X)
						if err != nil {
							t.Fatal(err)
						}
						rp, err = base.ForecastObserved(511, f.RX)
						if err != nil {
							t.Fatal(err)
						}
					}
					rc, lc := (rp >= .5) == f.RY, (lp >= .5) == f.Y
					rev, cp := ms[a].Observe(rc, lc, true)
					allow, err := gs[a].observe(gs[a].next, rc, lc)
					if err != nil {
						t.Fatal(err)
					}
					nom := bayes.RevisionSplits(rev.Action)
					arm := &z.Arms[a]
					if allow && arm.FirstAllow < 0 {
						arm.FirstAllow = clock
					}
					if nom && arm.FirstNomination < 0 {
						arm.FirstNomination = clock
					}
					if allow && nom && arm.SplitAt < 0 {
						arm.SplitAt = clock
					}
					arm.Trace = append(arm.Trace, switchGateTrace{clock, head, rc, lc, allow, nom, cp, gs[a].evidence.logWealth()})
				}
				released = append(released, head)
				head++
			}
		}
		if head != 512 || !reflect.DeepEqual(released, s.Released) || z.Arms[0].SplitAt != s.Arms[2].SplitAt {
			t.Fatal("original gate replay mismatch")
		}
		for a := 1; a < 3; a++ {
			if len(z.Arms[a].Trace) != z.ArrivedAuditPairs {
				t.Fatal("audit support mismatch")
			}
		}
		out = append(out, z)
	}
	return out
}

func TestFullAuditGateDiagnostic(t *testing.T) {
	in, out := os.Getenv("EVENTFRAME_FULL_AUDIT_INPUT"), os.Getenv("EVENTFRAME_FULL_AUDIT_OUTPUT")
	if in == "" || out == "" {
		t.Skip("opt-in consumed audit replay")
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
	var result []fullAuditGateRecord
	for {
		var r arrivalSwitchRecord
		if err := d.Decode(&r); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		result = append(result, fullAuditGateReplay(t, r)...)
	}
	if len(result) != 256 {
		t.Fatal("incomplete replay")
	}
	paths, err := filepath.Glob("../../internal/*/*.go")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, "../../go.mod", "../../go.sum", "../../docs/experiments/mmm-full-audit-gate-v1-contract.md", "../../research/full-audit-gate-summary.mjs")
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
	if err := json.NewEncoder(o).Encode(map[string]any{"ParentSHA256": hex.EncodeToString(h[:]), "Records": result, "Sources": sources, "Hashes": hashes, "Consumed": true}); err != nil {
		t.Fatal(err)
	}
}
