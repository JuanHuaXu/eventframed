package observationlearners

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
)

// Build only from as-of evidence. Outcomes outside the selected support are
// never copied into the model, even though the replay tape contains them.
func regimeQueryTapeView(input softV120Record, clock int) ([]segmentPacket, []int, []uint16, error) {
	if len(input.Steps) != 256 || clock < 8 || clock > 255 {
		return nil, nil, nil, fmt.Errorf("invalid tape clock")
	}
	available := make([]int, 0, clock+16)
	for j := -16; j < clock; j++ {
		if j < 0 || (!input.Steps[j].Missing && j+input.Steps[j].Delay <= clock) {
			available = append(available, j)
		}
	}
	available = available[max(0, len(available)-63):]
	history := make([]segmentPacket, clock+16)
	for i := range history {
		j := i - 16
		history[i].Arrives = -1
		if j < 0 {
			history[i].Bits = input.Initial[j+16].Bits
		} else {
			history[i].Bits = input.Steps[j].X
		}
	}
	for _, j := range available {
		history[j+16].Arrives = clock
		if j < 0 {
			history[j+16].Outcome = input.Initial[j+16].Outcome
		} else {
			history[j+16].Outcome = input.Steps[j].Y
		}
	}
	var pool []int
	for j := clock - 8; j < clock; j++ {
		s := input.Steps[j]
		if s.Missing || j+s.Delay > clock {
			pool = append(pool, j)
		}
	}
	probes := make([]uint16, 8)
	for i := range probes {
		probes[i] = input.Steps[clock-7+i].X
		if probes[i] >= 512 {
			return nil, nil, nil, fmt.Errorf("invalid visible probe")
		}
	}
	return history, pool, probes, nil
}

func regimeQueryFromTape(ctx context.Context, input softV120Record, clock int) (*regimeQueryState, []int, []uint16, error) {
	history, pool, probes, err := regimeQueryTapeView(input, clock)
	if err != nil {
		return nil, nil, nil, err
	}
	s, err := newRegimeQueryState(ctx, -16, clock, history)
	return s, pool, probes, err
}

func TestRegimeQueryTape(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("explicit consumed source required")
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
	if header.Version != "soft-learners-v120" {
		t.Fatal("source version")
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
		if input.Index != 0 || (input.Case != 0 && input.Case != 19 && input.Case != 20) {
			continue
		}
		const clock = 160
		before, _ := json.Marshal(input)
		s, pool, probes, err := regimeQueryFromTape(context.Background(), input, clock)
		if err != nil {
			t.Fatal(err)
		}
		if len(s.base.origins) != 63 {
			t.Fatal("support count")
		}
		if input.Schedule == 0 && len(pool) != 0 {
			t.Fatal("nomination of known evidence")
		}
		if input.Schedule == 1 && len(pool) == 0 {
			t.Fatal("missing delayed fixture")
		}
		mutant := input
		mutant.Steps = append([]softV120Step(nil), input.Steps...)
		retained := map[int]bool{}
		for _, j := range s.base.origins {
			retained[j] = true
		}
		for j := range mutant.Steps {
			v := &mutant.Steps[j]
			v.Q = 1 - v.Q
			if !retained[j] {
				v.Y = !v.Y
			}
			if j > clock {
				v.X ^= 511
			}
		}
		other, otherPool, otherProbes, err := regimeQueryFromTape(context.Background(), mutant, clock)
		if err != nil || !reflect.DeepEqual(s, other) || !reflect.DeepEqual(pool, otherPool) || !reflect.DeepEqual(probes, otherProbes) {
			t.Fatal("tape outcome/Q/future input leaked")
		}
		if len(pool) > 0 {
			j := pool[len(pool)-1]
			r, err := s.value(context.Background(), j, probes)
			if err != nil {
				t.Fatal(err)
			}
			r2, err := other.value(context.Background(), j, otherProbes)
			if err != nil || !reflect.DeepEqual(r, r2) {
				t.Fatal("query value leakage")
			}
		}
		after, _ := json.Marshal(input)
		if string(before) != string(after) {
			t.Fatal("tape mutation")
		}
		checked++
	}
	if checked != 12 {
		t.Fatal("fixture count", checked)
	}
	t.Log("12 tape views and six delayed query values: erased outcomes, future X, Q, complete-delivery pool and ownership PASS")
}
