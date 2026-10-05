package observationlearners

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type regimeMatchedRecord struct {
	Phase, Case, Index, Schedule, Clock              int
	Origins                                          [10][]int
	Predictions                                      [10][32]float64
	GenericMass, GenericLog, BooleanLog, LogEvidence [10]float64
	Error                                            string `json:",omitempty"`
}

func regimeSubset(full []int, n, phase, c, index, clock, rep int) []int {
	type item struct {
		origin int
		hash   [32]byte
	}
	if n < 0 || n > len(full) {
		panic("subset count")
	}
	ranked := make([]item, len(full))
	for i, j := range full {
		ranked[i] = item{j, sha256.Sum256([]byte(fmt.Sprintf("family-regime-match-v1:%d:%d:%d:%d:%d:%d", phase, c, index, clock, rep, j)))}
	}
	sort.Slice(ranked, func(i, j int) bool {
		cmp := bytes.Compare(ranked[i].hash[:], ranked[j].hash[:])
		return cmp < 0 || (cmp == 0 && ranked[i].origin < ranked[j].origin)
	})
	out := make([]int, n)
	for i := range out {
		out[i] = ranked[i].origin
	}
	sort.Ints(out)
	return out
}

func regimeMatchedFit(input softV120Record, clock int) regimeMatchedRecord {
	r := regimeMatchedRecord{Phase: input.Phase, Case: input.Case, Index: input.Index, Schedule: input.Schedule, Clock: clock}
	if len(input.Steps) != 256 || clock < 128 || clock > 224 {
		r.Error = "invalid diagnostic clock"
		return r
	}
	for j := -16; j < clock; j++ {
		if j < 0 || (!input.Steps[j].Missing && j+input.Steps[j].Delay <= clock) {
			r.Origins[0] = append(r.Origins[0], j)
		}
	}
	r.Origins[0] = r.Origins[0][max(0, len(r.Origins[0])-64):]
	for _, j := range r.Origins[0] {
		if j >= 128 {
			r.Origins[1] = append(r.Origins[1], j)
		}
	}
	for rep := 0; rep < 8; rep++ {
		r.Origins[rep+2] = regimeSubset(r.Origins[0], len(r.Origins[1]), r.Phase, r.Case, r.Index, clock, rep)
	}
	for arm, origins := range r.Origins {
		samples := make([]observation.Sample, len(origins))
		for i, j := range origins {
			if j < 0 {
				samples[i] = input.Initial[j+16]
			} else {
				samples[i] = observation.Sample{Bits: input.Steps[j].X, Outcome: input.Steps[j].Y}
			}
		}
		m := &familyEvidenceFit{genericMass: .95}
		for x := range m.predictions {
			m.predictions[x] = .5
		}
		if len(samples) > 0 {
			var err error
			m, err = fitFamilyEvidence(samples, .95)
			if err != nil {
				r.Error = err.Error()
				return r
			}
		}
		r.GenericMass[arm], r.GenericLog[arm], r.BooleanLog[arm], r.LogEvidence[arm] = m.genericMass, m.genericLog, m.booleanLog, m.logEvidence
		for j := 0; j < 32; j++ {
			x := input.Steps[clock+j].X
			if x >= 512 {
				r.Error = "invalid query"
				return r
			}
			r.Predictions[arm][j] = m.predictions[x]
		}
	}
	return r
}

func TestRegimeMatchedContracts(t *testing.T) {
	var synthetic softV120Record
	synthetic.Steps = make([]softV120Step, 256)
	for i := range synthetic.Steps {
		synthetic.Steps[i].X = uint16(i % 512)
		synthetic.Steps[i].Y = i%3 == 0
	}
	empty := regimeMatchedFit(synthetic, 128)
	if empty.Error != "" {
		t.Fatal(empty.Error)
	}
	for a := 1; a < 10; a++ {
		if len(empty.Origins[a]) != 0 || empty.GenericMass[a] != .95 || empty.LogEvidence[a] != 0 {
			t.Fatal("empty prior")
		}
		for _, p := range empty.Predictions[a] {
			if p != .5 {
				t.Fatal("fake evidence")
			}
		}
	}
	all := regimeMatchedFit(synthetic, 224)
	if all.Error != "" {
		t.Fatal(all.Error)
	}
	for a := 1; a < 10; a++ {
		if !reflect.DeepEqual(all.Origins[a], all.Origins[0]) || all.Predictions[a] != all.Predictions[0] {
			t.Fatal("all-current control")
		}
	}
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("real consumed fixture path required")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var header softV120Artifact
	if err := d.Decode(&header); err != nil {
		t.Fatal(err)
	}
	checked := 0
	for {
		var input softV120Record
		err := d.Decode(&input)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if input.Phase != 0 || input.Index != 0 || input.Schedule != 1 || (input.Case != 0 && input.Case != 20) {
			continue
		}
		before, _ := json.Marshal(input)
		for _, clock := range []int{160, 192, 224} {
			r := regimeMatchedFit(input, clock)
			if r.Error != "" {
				t.Fatal(r.Error)
			}
			for a, origins := range r.Origins {
				if a > 0 && len(origins) != len(r.Origins[1]) {
					t.Fatal("unequal n")
				}
				for i, j := range origins {
					if i > 0 && origins[i-1] >= j {
						t.Fatal("duplicate/unordered")
					}
					if j >= clock || input.Steps[j].Missing || j+input.Steps[j].Delay > clock {
						t.Fatal("unavailable label")
					}
				}
			}
			changed := input
			changed.Steps = append([]softV120Step(nil), input.Steps...)
			for j := range changed.Steps {
				s := &changed.Steps[j]
				s.Q = 1 - s.Q
				if j >= clock || s.Missing || j+s.Delay > clock {
					s.Y = !s.Y
				}
			}
			if !reflect.DeepEqual(regimeMatchedFit(changed, clock), r) {
				t.Fatal("future label/Q leakage")
			}
			checked++
		}
		after, _ := json.Marshal(input)
		if string(before) != string(after) {
			t.Fatal("input mutation")
		}
	}
	if checked != 6 {
		t.Fatalf("checked%d", checked)
	}
	t.Log("empty prior, all-current equality, six count-matched/as-of snapshot checks PASS")
}
