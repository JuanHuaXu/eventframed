package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/transfergenerator"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const transferDiagnosisV117Parent = "cad9f6f36879280044f0148895910f9fc10e3d0df902da95cc8276980f079d1e"

type transferDiagnosisV117Profile struct {
	Clock int
	Full  [4][512]float64
}
type transferDiagnosisV117Step struct {
	Raw     [3][4]float64
	Generic float64
}
type transferDiagnosisV117Record struct {
	Spec     transfergenerator.Spec
	Schedule int
	Profiles []transferDiagnosisV117Profile
	Steps    []transferDiagnosisV117Step
}
type transferDiagnosisV117Artifact struct {
	Version, ParentSHA256 string
	Hashes                map[string]string
	Records               []transferDiagnosisV117Record
}

func transferDiagnosisV117Run(r transferRecord) (transferDiagnosisV117Record, error) {
	out := transferDiagnosisV117Record{Spec: r.Teacher.Spec, Schedule: r.Schedule}
	if (r.Schedule != 0 && r.Schedule != 1) || len(r.Fits) != 8 || len(r.Steps) != 256 {
		return out, fmt.Errorf("parent shape")
	}
	data, teacher, err := transfergenerator.Generate(r.Teacher.Spec)
	if err != nil || teacher.Describe() != r.Teacher {
		return out, fmt.Errorf("teacher replay: %v", err)
	}
	history := make([]observation.Sample, 0, 272)
	for _, s := range data.Initial {
		history = append(history, observation.Sample{Bits: s.X, Outcome: s.Y})
	}
	for i, s := range r.Steps {
		p := data.Frames[i]
		q, err := teacher.Truth(p.X, i)
		if err != nil || p.X != s.X || p.Y != s.Y || q != s.Q {
			return out, fmt.Errorf("packet replay")
		}
		delay, missing := 0, false
		if r.Schedule == 1 {
			delay, missing = int(p.Delay), p.Missing
		}
		if delay != s.Delay || missing != s.Missing {
			return out, fmt.Errorf("schedule replay")
		}
		history = append(history, observation.Sample{Bits: s.X, Outcome: s.Y})
	}
	var weights [512]float64
	for i := range weights {
		weights[i] = 1
	}
	var models [4]*ConditionalForest
	for t, s := range r.Steps {
		if t%32 == 0 {
			fit := r.Fits[t/32]
			if fit.Clock != t {
				return out, fmt.Errorf("fit clock")
			}
			origins := make([]int, 0, 272)
			for i := -16; i < t; i++ {
				if i < 0 || !r.Steps[i].Missing && i+r.Steps[i].Delay <= t {
					origins = append(origins, i)
				}
			}
			var samples [2][]observation.Sample
			for j, cap := range []int{64, 32} {
				want := origins
				if len(want) > cap {
					want = want[len(want)-cap:]
				}
				if !reflect.DeepEqual(want, fit.Origins[j]) {
					return out, fmt.Errorf("as-of origins")
				}
				for _, origin := range want {
					samples[j] = append(samples[j], history[origin+16])
				}
			}
			for j := 0; j < 2; j++ {
				var err error
				models[2*j], err = NewSubsetConditional(samples[j], weights)
				if err != nil {
					return out, err
				}
				models[2*j+1], err = NewBooleanConditional(samples[j])
				if err != nil {
					return out, err
				}
			}
			profile := transferDiagnosisV117Profile{Clock: t}
			for k, m := range models {
				for x := 0; x < 512; x++ {
					p, err := m.Forecast(511, uint16(x))
					if err != nil {
						return out, err
					}
					profile.Full[k][x] = p
				}
			}
			out.Profiles = append(out.Profiles, profile)
		}
		row := transferDiagnosisV117Step{}
		for j, arm := range []int{0, 1, 2} {
			for k, m := range models {
				p, err := m.Forecast(s.Mask[arm], s.X&s.Mask[arm])
				if err != nil {
					return out, err
				}
				row.Raw[j][k] = p
			}
		}
		p, err := models[0].Forecast(s.Mask[0], s.X&s.Mask[0])
		if err != nil {
			return out, err
		}
		if p != s.P[0] {
			return out, fmt.Errorf("generic parent forecast changed at %d", t)
		}
		row.Generic = p
		out.Steps = append(out.Steps, row)
	}
	return out, nil
}

func TestTransferDiagnosisV117Smoke(t *testing.T) {
	for f := transfergenerator.Additive; f <= transfergenerator.LocalTable; f++ {
		for m := transfergenerator.Stationary; m <= transfergenerator.Gradual; m++ {
			data, teacher, err := transfergenerator.Generate(transfergenerator.Spec{SeedBase: transfergenerator.ValidationSeedBase, Family: f, Mode: m})
			if err != nil {
				t.Fatal(err)
			}
			for schedule := 0; schedule < 2; schedule++ {
				r, err := transferRunEvidence(data, teacher, schedule)
				if err != nil {
					t.Fatal(err)
				}
				d, err := transferDiagnosisV117Run(r)
				if err != nil || len(d.Profiles) != 8 || len(d.Steps) != 256 {
					t.Fatal("coverage", err)
				}
				r.Steps[100].Y = !r.Steps[100].Y
				if _, err := transferDiagnosisV117Run(r); err == nil {
					t.Fatal("changed label accepted")
				}
				r.Steps[100].Y = !r.Steps[100].Y
				r.Fits[1].Origins[0][0] = 32
				if _, err := transferDiagnosisV117Run(r); err == nil {
					t.Fatal("future fit origin accepted")
				}
			}
		}
	}
}
func TestTransferDiagnosisV117(t *testing.T) {
	output := os.Getenv("EVENTFRAME_TRANSFER_DIAGNOSIS_V117")
	if output == "" {
		t.Skip("explicit artifact required")
	}
	replay := os.Getenv("EVENTFRAME_TRANSFER_DIAGNOSIS_V117_REPLAY") == "1"
	if !replay {
		if _, err := os.Stat(output); !os.IsNotExist(err) {
			t.Fatal("output exists or cannot inspect", err)
		}
	}
	root := "../.."
	raw, err := os.ReadFile(filepath.Join(root, "docs/experiments/mmm-transfer-v116.json"))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != transferDiagnosisV117Parent {
		t.Fatal("parent hash")
	}
	var parent transferV116Artifact
	if err = json.Unmarshal(raw, &parent); err != nil {
		t.Fatal(err)
	}
	if len(parent.Records) != 1152 || len(parent.Hashes) != 154 {
		t.Fatal("parent shape")
	}
	a := transferDiagnosisV117Artifact{Version: "transfer-diagnosis-v117", ParentSHA256: transferDiagnosisV117Parent, Hashes: map[string]string{}}
	for name, h := range parent.Hashes {
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != h {
			t.Fatal("parent source changed", name, err)
		}
		a.Hashes[name] = h
	}
	for _, name := range []string{"internal/observationlearners/transfer_diagnosis_v117_test.go", "research/transfer-diagnosis-v117-summary.mjs", "docs/experiments/mmm-transfer-diagnosis-v117-protocol.md"} {
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		a.Hashes[name] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	for _, r := range parent.Records {
		d, err := transferDiagnosisV117Run(r)
		if err != nil {
			t.Fatal(r.Teacher.Spec, r.Schedule, err)
		}
		a.Records = append(a.Records, d)
	}
	if replay {
		b, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		var old transferDiagnosisV117Artifact
		if err = json.Unmarshal(b, &old); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, old) {
			t.Fatal("diagnosis replay mismatch")
		}
		t.Log("exact replay", len(a.Records), "records", len(a.Hashes), "hashes")
		return
	}
	f, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.NewEncoder(f).Encode(a); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	t.Log("wrote", len(a.Records), "records", len(a.Hashes), "hashes")
}
