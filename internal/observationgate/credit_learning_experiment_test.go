package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCreditLearningDisabledParity(t *testing.T) {
	for s, name := range arrivalSwitchNames {
		base, m, err := arrivalSwitchSetup(0, s, 0)
		if err != nil {
			t.Fatal(err)
		}
		cfg := forestDependenceCase{subsetBreadthCase{Name: name, Mode: "stable", Target: "bit"}, 0}
		outcome := func(x uint16, c int, ref bool, r *rand.Rand) bool {
			y := arrivalSwitchTruth(x, m, s, c, ref)
			if r.Float64() < .05 {
				y = !y
			}
			return y
		}
		for _, delay := range []bool{false, true} {
			want, err := innerArrivalRunOutcome(base, cfg, s, 0, 2026092431, delay, outcome)
			if err != nil {
				t.Fatal(err)
			}
			got, acq, err := creditLearningRun(base, cfg, s, 0, 2026092431, delay, outcome, false)
			if err != nil || len(acq) != 0 || !reflect.DeepEqual(want, got) {
				t.Fatal("disabled parity", name, delay, err)
			}
		}
	}
}

type creditLearningRecord struct {
	Phase, Case        string
	Index              int
	Masks              [2]uint16
	Immediate, Delayed innerArrivalResult
	Acquisition        [2][]monitorCreditAcquisition
}

func TestCreditLearningExperiment(t *testing.T) {
	in, shadow, out := os.Getenv("EVENTFRAME_CREDIT_LEARNING_INPUT"), os.Getenv("EVENTFRAME_CREDIT_LEARNING_SHADOW"), os.Getenv("EVENTFRAME_CREDIT_LEARNING_OUTPUT")
	if in == "" || shadow == "" || out == "" {
		t.Skip("opt-in consumed closed loop")
	}
	raw, err := os.ReadFile(in)
	if err != nil {
		t.Fatal(err)
	}
	parentHash := sha256.Sum256(raw)
	sraw, err := os.ReadFile(shadow)
	if err != nil {
		t.Fatal(err)
	}
	shadowHash := sha256.Sum256(sraw)
	var sh struct{ Records []monitorCreditRecord }
	if err := json.Unmarshal(sraw, &sh); err != nil {
		t.Fatal(err)
	}
	if len(sh.Records) != 256 {
		t.Fatal("shadow length")
	}
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
	o, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()
	paths, err := filepath.Glob("../../internal/*/*.go")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, "../../go.mod", "../../go.sum", "../../docs/experiments/mmm-credit-learning-v1-contract.md", "../../research/credit-learning-summary.mjs")
	sources, hashes := map[string]string{}, map[string]string{}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		sources[p[6:]] = string(b)
		hashes[p[6:]] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(o)
	if err := enc.Encode(map[string]any{"Sources": sources, "Hashes": hashes, "ParentSHA256": hex.EncodeToString(parentHash[:]), "ShadowSHA256": hex.EncodeToString(shadowHash[:]), "Consumed": true}); err != nil {
		t.Fatal(err)
	}
	n := 0
	for {
		var p arrivalSwitchRecord
		if err := d.Decode(&p); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		phase := 0
		if p.Phase == "cohort2" {
			phase = 1
		}
		scenario := -1
		for s, name := range arrivalSwitchNames {
			if name == p.Case {
				scenario = s
			}
		}
		if scenario < 0 {
			t.Fatal("unknown case")
		}
		base, m, err := arrivalSwitchSetup(phase, scenario, p.Index)
		if err != nil || m != p.Masks {
			t.Fatal("setup", err)
		}
		cfg := forestDependenceCase{subsetBreadthCase{Name: p.Case, Mode: "stable", Target: "bit"}, 0}
		outcome := func(x uint16, c int, ref bool, r *rand.Rand) bool {
			y := arrivalSwitchTruth(x, m, scenario, c, ref)
			if r.Float64() < .05 {
				y = !y
			}
			return y
		}
		row := creditLearningRecord{Phase: p.Phase, Case: p.Case, Index: p.Index, Masks: m}
		for schedule := 0; schedule < 2; schedule++ {
			got, acq, err := creditLearningRun(base, cfg, scenario, p.Index, int64(2026092431+100*phase), schedule == 1, outcome, true)
			if err != nil {
				t.Fatal(err)
			}
			want := p.Immediate
			if schedule == 1 {
				want = p.Delayed
			}
			ss := sh.Records[2*n+schedule]
			if ss.Phase != p.Phase || ss.Case != p.Case || ss.Index != p.Index || ss.Schedule != []string{"Immediate", "Delayed"}[schedule] {
				t.Fatal("shadow key mismatch")
			}
			if !reflect.DeepEqual(acq, ss.Acquisition) || got.Arms[2].SplitAt != ss.Candidate.SplitAt || got.MonitorCost+got.AuditCost != ss.CandidateCost {
				t.Fatal("shadow drift", p.Case, p.Index, schedule)
			}
			if !reflect.DeepEqual(got.Fits, want.Fits) || !reflect.DeepEqual(got.Released, want.Released) || len(got.Frames) != 512 {
				t.Fatal("training or release drift")
			}
			for i, a := range got.Frames {
				b := want.Frames[i]
				if a.X != b.X || a.RX != b.RX || a.Y != b.Y || a.RY != b.RY || a.Audit != b.Audit || a.Missing != b.Missing || a.Arrival != b.Arrival || a.Seed != b.Seed {
					t.Fatal("latent drift")
				}
			}
			row.Acquisition[schedule] = acq
			if schedule == 0 {
				row.Immediate = got
			} else {
				row.Delayed = got
			}
		}
		if err := enc.Encode(row); err != nil {
			t.Fatal(err)
		}
		n++
		if p.Index == 15 {
			if err := o.Sync(); err != nil {
				t.Fatal(err)
			}
			t.Log(p.Phase, p.Case, "complete")
		}
	}
	if n != 128 {
		t.Fatal("incomplete collection", n)
	}
}
