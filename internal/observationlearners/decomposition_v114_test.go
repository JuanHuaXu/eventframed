package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

const decompositionV114Parent = "54b6220b6936227e4e1e45954a67522a5f7eb152adb32a3bc8ef5f5283550860"

type decompositionV114Profile struct {
	Clock int
	Full  [4][512]float64
}
type decompositionV114Step struct {
	Raw     [3][4]float64
	Generic float64
}
type decompositionV114Record struct {
	Phase, Case string
	Index       int
	Profiles    []decompositionV114Profile
	Steps       []decompositionV114Step
}
type decompositionV114Artifact struct {
	Version, ParentSHA256 string
	Hashes                map[string]string
	Records               []decompositionV114Record
}

func decompositionV114Run(r hazardV113Record) (decompositionV114Record, error) {
	out := decompositionV114Record{Phase: r.Phase, Case: r.Case, Index: r.Index}
	phase := 0
	if r.Phase == "confirmation" {
		phase = 1
	} else if r.Phase != "design" {
		return out, fmt.Errorf("phase")
	}
	scenario := -1
	for i, c := range replicationV102Cases {
		if c == r.Case {
			scenario = i
		}
	}
	if scenario < 0 || r.Schedule != 1 || len(r.Fits) != 8 || len(r.Steps) != 256 {
		return out, fmt.Errorf("parent shape")
	}
	seed := int64(2150111300 + phase*1000000 + scenario*10000 + r.Index*10)
	xs, ys := rand.New(rand.NewSource(seed+1)), rand.New(rand.NewSource(seed+2))
	history := make([]observation.Sample, 0, 272)
	name := r.Case
	if scenario == 10 {
		name = "majority3"
	} else if scenario == 11 {
		name = "parity4"
	}
	for i := 0; i < 16; i++ {
		x := stackV93Input(uint16(xs.Intn(512)), r.Case)
		history = append(history, observation.Sample{Bits: x, Outcome: ys.Float64() < stackV93Truth(x, r.Rules[0], name)})
	}
	for _, s := range r.Steps {
		if x := stackV93Input(uint16(xs.Intn(512)), r.Case); x != s.X {
			return out, fmt.Errorf("input replay")
		}
		if y := ys.Float64() < s.Q; y != s.Y {
			return out, fmt.Errorf("outcome replay")
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
			profile := decompositionV114Profile{Clock: t}
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
		row := decompositionV114Step{}
		for j, arm := range []int{5, 7, 8} {
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

func TestDecompositionV114Smoke(t *testing.T) {
	for _, c := range []int{3, 10, 11} {
		r, err := hazardV113Run(0, c, 0, 1, 2150111300)
		if err != nil {
			t.Fatal(err)
		}
		d, err := decompositionV114Run(r)
		if err != nil {
			t.Fatal(err)
		}
		if len(d.Profiles) != 8 || len(d.Steps) != 256 {
			t.Fatal("coverage")
		}
		r.Fits[1].Origins[0][0] = 32
		if _, err := decompositionV114Run(r); err == nil {
			t.Fatal("future-origin corruption accepted")
		}
	}
}

func TestDecompositionV114(t *testing.T) {
	outPath := os.Getenv("EVENTFRAME_DECOMPOSITION_V114")
	if outPath == "" {
		t.Skip("explicit artifact required")
	}
	root := filepath.Join("..", "..")
	raw, err := os.ReadFile(filepath.Join(root, "docs/experiments/mmm-hazard-v113.json"))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != decompositionV114Parent {
		t.Fatal("parent digest")
	}
	var parent hazardV113Artifact
	if err := json.Unmarshal(raw, &parent); err != nil {
		t.Fatal(err)
	}
	a := decompositionV114Artifact{Version: "v114", ParentSHA256: decompositionV114Parent, Hashes: map[string]string{}}
	for file, hash := range parent.Hashes {
		data, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != hash {
			t.Fatal("parent source changed", file)
		}
		a.Hashes[file] = hash
	}
	for _, file := range []string{"internal/observationlearners/decomposition_v114_test.go", "research/decomposition-v114-summary.mjs", "docs/experiments/mmm-decomposition-v114-protocol.md"} {
		data, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatal(err)
		}
		a.Hashes[file] = fmt.Sprintf("%x", sha256.Sum256(data))
	}
	for _, r := range parent.Records {
		if r.Schedule != 1 || r.Case != "parity4" && r.Case != "majority_to_parity" && r.Case != "parity_to_majority" {
			continue
		}
		d, err := decompositionV114Run(r)
		if err != nil {
			t.Fatal(r.Phase, r.Case, r.Index, err)
		}
		a.Records = append(a.Records, d)
	}
	if len(a.Records) != 192 {
		t.Fatal("coverage")
	}
	if os.Getenv("EVENTFRAME_DECOMPOSITION_V114_REPLAY") == "1" {
		data, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatal(err)
		}
		var old decompositionV114Artifact
		if err := json.Unmarshal(data, &old); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, old) {
			t.Fatal("replay mismatch")
		}
		return
	}
	f, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(f).Encode(a); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
