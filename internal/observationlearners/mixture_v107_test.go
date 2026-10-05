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

type mixtureV107Publication struct {
	Clock        int
	Origins      [2][]int
	Rows         [512][5]float64
	Truth        [512]float64
	Profile      mixtureProfile
	Raw, Neutral mixtureOptimum
}
type mixtureV107Record struct {
	Phase, Case  string
	Index        int
	Publications []mixtureV107Publication
}
type mixtureV107Artifact struct {
	Version, ParentSHA256 string
	Hashes                map[string]string
	Records               []mixtureV107Record
}

func mixtureV107Reconstruct(r arrivalV106Record, base int64) (mixtureV107Record, error) {
	out := mixtureV107Record{Phase: r.Phase, Case: r.Case, Index: r.Index}
	phase, scenario := 0, 10
	if r.Phase == "confirmation" {
		phase = 1
	} else if r.Phase != "design" {
		return out, fmt.Errorf("unknown phase")
	}
	if r.Case == "parity_to_majority" {
		scenario = 11
	} else if r.Case != "majority_to_parity" {
		return out, fmt.Errorf("unknown case")
	}
	if r.Schedule != 1 || r.Index < 0 || r.Index >= 32 || len(r.Steps) != 256 || len(r.Fits) != 8 {
		return out, fmt.Errorf("record shape")
	}
	seed := base + int64(phase*1000000+scenario*10000+r.Index*10)
	xs, ys := rand.New(rand.NewSource(seed+1)), rand.New(rand.NewSource(seed+2))
	truth := func(x uint16, clock int) float64 {
		name, rule := "majority3", r.Rules[0]
		if scenario == 11 {
			name = "parity4"
		}
		if clock >= 128 {
			rule = r.Rules[1]
			if scenario == 10 {
				name = "parity4"
			} else {
				name = "majority3"
			}
		}
		return stackV93Truth(x, rule, name)
	}
	history := make([]observation.Sample, 272)
	for clock := -16; clock < 256; clock++ {
		x := stackV93Input(uint16(xs.Intn(512)), r.Case)
		q := truth(x, clock)
		y := ys.Float64() < q
		history[clock+16] = observation.Sample{Bits: x, Outcome: y}
		if clock >= 0 {
			s := r.Steps[clock]
			if s.X != x || s.Q != q || s.Y != y {
				return out, fmt.Errorf("stream mismatch at %d", clock)
			}
		}
	}
	var weights [512]float64
	for i := range weights {
		weights[i] = 1
	}
	for fi, f := range r.Fits {
		if f.Clock != 32*fi {
			return out, fmt.Errorf("publication clock")
		}
		var eligible []int
		for i := -16; i < 0; i++ {
			eligible = append(eligible, i)
		}
		for i := 0; i < f.Clock; i++ {
			s := r.Steps[i]
			if !s.Missing && i+s.Delay <= f.Clock {
				eligible = append(eligible, i)
			}
		}
		if len(eligible) > 64 {
			eligible = eligible[len(eligible)-64:]
		}
		short := eligible
		if len(short) > 32 {
			short = short[len(short)-32:]
		}
		if !reflect.DeepEqual(f.Origins, [2][]int{eligible, short}) {
			return out, fmt.Errorf("as-of origins mismatch")
		}
		if f.Clock < 128 {
			continue
		}
		var samples [2][]observation.Sample
		for j, origins := range f.Origins {
			for _, i := range origins {
				samples[j] = append(samples[j], history[i+16])
			}
		}
		var models [4]*ConditionalForest
		var err error
		models[0], err = NewSubsetConditional(samples[0], weights)
		if err != nil {
			return out, err
		}
		models[1], err = NewBooleanConditional(samples[0])
		if err != nil {
			return out, err
		}
		models[2], err = NewSubsetConditional(samples[1], weights)
		if err != nil {
			return out, err
		}
		models[3], err = NewBooleanConditional(samples[1])
		if err != nil {
			return out, err
		}
		for clock := f.Clock; clock < f.Clock+32; clock++ {
			s := r.Steps[clock]
			p, err := models[0].Forecast(s.Mask[0], s.X&s.Mask[0])
			if err != nil || p != s.P[0] {
				return out, fmt.Errorf("issued generic mismatch %d: %v", clock, err)
			}
		}
		p := mixtureV107Publication{Clock: f.Clock, Origins: f.Origins}
		for x := uint16(0); x < 512; x++ {
			p.Truth[x] = truth(x, f.Clock)
			p.Rows[x][4] = .5
			for j, m := range models {
				p.Rows[x][j], err = m.Forecast(511, x)
				if err != nil {
					return out, err
				}
			}
		}
		p.Profile, err = makeMixtureProfile(p.Rows[:], p.Truth[:])
		if err != nil {
			return out, err
		}
		p.Raw, err = p.Profile.minimum(4)
		if err != nil {
			return out, err
		}
		p.Neutral, err = p.Profile.minimum(5)
		if err != nil {
			return out, err
		}
		out.Publications = append(out.Publications, p)
	}
	return out, nil
}

func TestMixtureV107(t *testing.T) {
	output := os.Getenv("EVENTFRAME_MIXTURE_V107")
	if output == "" {
		t.Skip("explicit diagnostic artifact required")
	}
	raw, err := os.ReadFile("../../docs/experiments/mmm-arrival-v106.json")
	if err != nil {
		t.Fatal(err)
	}
	parentHash := fmt.Sprintf("%x", sha256.Sum256(raw))
	if parentHash != "bc2d8289e2ed1910c7fdf0a425265453f2ea5e1b282b8687622624fb1fdf9d45" {
		t.Fatal("parent changed")
	}
	var parent arrivalV106Artifact
	if err = json.Unmarshal(raw, &parent); err != nil {
		t.Fatal(err)
	}
	a := mixtureV107Artifact{Version: "v107", ParentSHA256: parentHash, Hashes: map[string]string{}}
	for name, hash := range parent.Hashes {
		a.Hashes[name] = hash
	}
	for _, name := range []string{"internal/observationlearners/mixture_oracle.go", "internal/observationlearners/mixture_oracle_test.go", "internal/observationlearners/mixture_v107_test.go", "research/mixture-v107-protocol.md", "research/mixture-v107-summary.mjs"} {
		b, err := os.ReadFile(filepath.Join("../..", name))
		if err != nil {
			t.Fatal(err)
		}
		a.Hashes[name] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	for name, hash := range a.Hashes {
		b, err := os.ReadFile(filepath.Join("../..", name))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
			t.Fatal("source changed", name, err)
		}
	}
	seen := map[string]bool{}
	for _, r := range parent.Records {
		if r.Schedule != 1 || (r.Case != "majority_to_parity" && r.Case != "parity_to_majority") {
			continue
		}
		key := fmt.Sprintf("%s/%s/%d", r.Phase, r.Case, r.Index)
		if seen[key] {
			t.Fatal("duplicate record")
		}
		seen[key] = true
		got, err := mixtureV107Reconstruct(r, parent.SeedBase)
		if err != nil {
			t.Fatal(key, err)
		}
		a.Records = append(a.Records, got)
	}
	if len(a.Records) != 128 {
		t.Fatal("record coverage")
	}
	if os.Getenv("EVENTFRAME_MIXTURE_V107_REPLAY") == "1" {
		b, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		var old mixtureV107Artifact
		if err = json.Unmarshal(b, &old); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, old) {
			t.Fatal("reconstruction changed")
		}
		return
	}
	raw, err = json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write(raw); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
}
