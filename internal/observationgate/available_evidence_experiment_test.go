package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

func TestAvailableEvidenceContracts(t *testing.T) {
	base, _, err := arrivalSwitchSetup(0, 3, 0)
	if err != nil {
		t.Fatal(err)
	}
	j := &subsetDelayJournal{state: subsetState{base: base, enabled: true}}
	if err := j.publish(observationpreserved.Models{}, nil); err != nil {
		t.Fatal(err)
	}
	p, err := j.predict(0, observationexperiment.Frames(137, "contract"), 0)
	if err != nil {
		t.Fatal(err)
	}
	before := *j
	x, err := availableEvidenceCapture(j, 0, 0, 0)
	t.Logf("same-view reconstruction absolute difference %.17g", math.Abs(x.Available-p.P))
	if err != nil || x.Available != x.Original || math.Abs(x.Available-p.P) > 1e-14 || x.ConsumedMask != p.Mask {
		t.Fatal("same view", x.Available, p.P, err)
	}
	if _, err := availableEvidenceCapture(j, 0, 511, 137); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, *j) {
		t.Fatal("diagnostic mutated learner")
	}
	bit := p.Mask & (^p.Mask + 1)
	if bit == 0 {
		t.Fatal("empty requested view")
	}
	for _, v := range [][2]uint16{{512, 0}, {0, 1}, {bit, (p.Values & bit) ^ bit}} {
		if _, err := availableEvidenceCapture(j, 0, v[0], v[1]); err == nil {
			t.Fatal("invalid/conflicting view")
		}
	}
	if _, err := availableEvidenceCapture(j, 1, 0, 0); err == nil {
		t.Fatal("unissued origin")
	}
	for _, delay := range []bool{false, true} {
		base, m, err := arrivalSwitchSetup(0, 3, 0)
		if err != nil {
			t.Fatal(err)
		}
		cfg := forestDependenceCase{subsetBreadthCase{Name: arrivalSwitchNames[3], Mode: "stable", Target: "bit"}, 0}
		outcome := func(x uint16, c int, ref bool, r *rand.Rand) bool {
			y := arrivalSwitchTruth(x, m, 3, c, ref)
			if r.Float64() < .05 {
				y = !y
			}
			return y
		}
		want, wa, err := creditLearningRun(base, cfg, 3, 0, 2026092431, delay, outcome, true)
		if err != nil {
			t.Fatal(err)
		}
		var diag []availableEvidenceFrame
		got, ga, err := availableEvidenceRun(base, cfg, 3, 0, 2026092431, delay, outcome, true, &diag)
		if err != nil || len(diag) != 512 || !reflect.DeepEqual(want, got) || !reflect.DeepEqual(wa, ga) {
			t.Fatal("diagnostic parity", delay, err)
		}
	}
}

type availableEvidenceRecord struct {
	Phase, Case        string
	Index              int
	Immediate, Delayed []availableEvidenceFrame
}

func TestAvailableEvidenceExperiment(t *testing.T) {
	in, out := os.Getenv("EVENTFRAME_AVAILABLE_EVIDENCE_INPUT"), os.Getenv("EVENTFRAME_AVAILABLE_EVIDENCE_OUTPUT")
	if in == "" || out == "" {
		t.Skip("opt-in consumed available-view diagnostic")
	}
	raw, err := os.ReadFile(in)
	if err != nil {
		t.Fatal(err)
	}
	parentHash := sha256.Sum256(raw)
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
	paths = append(paths, "../../go.mod", "../../go.sum", "../../docs/experiments/mmm-available-evidence-v1-contract.md", "../../research/available-evidence-summary.mjs")
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
	if err := enc.Encode(map[string]any{"Sources": sources, "Hashes": hashes, "ParentSHA256": hex.EncodeToString(parentHash[:]), "Consumed": true}); err != nil {
		t.Fatal(err)
	}
	n := 0
	for {
		var p creditLearningRecord
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
			t.Fatal("case")
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
		row := availableEvidenceRecord{Phase: p.Phase, Case: p.Case, Index: p.Index}
		for schedule := 0; schedule < 2; schedule++ {
			var diag []availableEvidenceFrame
			got, acq, err := availableEvidenceRun(base, cfg, scenario, p.Index, int64(2026092431+100*phase), schedule == 1, outcome, true, &diag)
			want := p.Immediate
			if schedule == 1 {
				want = p.Delayed
			}
			if err != nil || len(diag) != 512 || !reflect.DeepEqual(got, want) || !reflect.DeepEqual(acq, p.Acquisition[schedule]) {
				t.Fatal("archive replay drift", p.Case, p.Index, schedule, err)
			}
			if schedule == 0 {
				row.Immediate = diag
			} else {
				row.Delayed = diag
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
		t.Fatal("incomplete", n)
	}
}
