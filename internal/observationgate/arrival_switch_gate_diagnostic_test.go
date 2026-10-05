package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
)

type switchGateTrace struct {
	Clock, Origin                                               int
	ReferenceCorrect, LiveCorrect, Allow, Nominate, ChangePoint bool
	LogWealth                                                   float64
}
type switchGateRecord struct {
	Phase, Case, Schedule                        string
	Index                                        int
	AccuracyBefore, AccuracyAfter, ConditionalTV float64
	FirstAllow, FirstNomination, SplitAt         int
	Trace                                        []switchGateTrace
}

// Exhaustive latent-law calculations below are evaluator-only: neither truth
// nor the diagnostic sees its way back into a forecast or gate update.
func TestArrivalSwitchGateDiagnostic(t *testing.T) {
	in, out := os.Getenv("EVENTFRAME_SWITCH_DIAG_INPUT"), os.Getenv("EVENTFRAME_SWITCH_DIAG_OUTPUT")
	if in == "" || out == "" {
		t.Skip("opt-in consumed trace diagnostic")
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
	var result []switchGateRecord
	for {
		var r arrivalSwitchRecord
		if err := d.Decode(&r); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
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
			t.Fatal("setup drift", err)
		}
		var p [512]float64
		var before, after, tv float64
		for x := uint16(0); x < 512; x++ {
			v, err := observation.Run(base, observationexperiment.Frames(x, "diagnostic"), 1, "mmm", 0)
			if err != nil {
				t.Fatal(err)
			}
			p[x] = v.Probability
			a, b := arrivalSwitchTruth(x, masks, scenario, 255, false), arrivalSwitchTruth(x, masks, scenario, 256, false)
			if (p[x] >= .5) == a {
				before += .95 / 512
			} else {
				before += .05 / 512
			}
			if (p[x] >= .5) == b {
				after += .95 / 512
			} else {
				after += .05 / 512
			}
			if a != b {
				tv += .9 / 512
			}
		}
		for schedule, s := range []innerArrivalResult{r.Immediate, r.Delayed} {
			z := switchGateRecord{Phase: r.Phase, Case: r.Case, Schedule: []string{"Immediate", "Delayed"}[schedule], Index: r.Index, AccuracyBefore: before, AccuracyAfter: after, ConditionalTV: tv, FirstAllow: -1, FirstNomination: -1, SplitAt: -1}
			var gate pairedInvestigator
			var monitor observationrescue.Monitor
			var released []int
			head := 0
			for clock := 0; clock < 544; clock++ {
				for head < 512 && head <= clock {
					f := s.Frames[head]
					if math.Abs(p[f.X]-f.Baseline) > 1e-12 || math.Abs(p[f.RX]-f.Reference) > 1e-12 {
						t.Fatal("issued monitor forecast mismatch")
					}
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
					rc, lc := (f.Reference >= .5) == f.RY, (f.Baseline >= .5) == f.Y
					rev, cp := monitor.Observe(rc, lc, true)
					allow, err := gate.observe(gate.next, rc, lc)
					if err != nil {
						t.Fatal(err)
					}
					nom := bayes.RevisionSplits(rev.Action)
					if allow && z.FirstAllow < 0 {
						z.FirstAllow = clock
					}
					if nom && z.FirstNomination < 0 {
						z.FirstNomination = clock
					}
					if allow && nom && z.SplitAt < 0 {
						z.SplitAt = clock
					}
					z.Trace = append(z.Trace, switchGateTrace{clock, head, rc, lc, allow, nom, cp, gate.evidence.logWealth()})
					released = append(released, head)
					head++
				}
			}
			if head != 512 || !reflect.DeepEqual(released, s.Released) || z.SplitAt != s.Arms[2].SplitAt {
				t.Fatal("gate replay mismatch", r.Case, r.Index, schedule, z.SplitAt, s.Arms[2].SplitAt)
			}
			result = append(result, z)
		}
	}
	if len(result) != 256 {
		t.Fatal("incomplete diagnostic")
	}
	output, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	if err := json.NewEncoder(output).Encode(map[string]any{"RawSHA256": hex.EncodeToString(h[:]), "Records": result, "Limits": "Consumed evaluator-only exhaustive truth and exact issued-monitor replay; not a replacement certificate or intervention."}); err != nil {
		t.Fatal(err)
	}
}
