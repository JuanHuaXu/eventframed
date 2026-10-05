package observationlearners

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type acquisitionTrainQuery struct{ Clock, Origin, Reveal int }
type acquisitionTrainFit struct {
	Clock   int
	Origins [2][]int
}
type acquisitionTrainResult struct {
	Predictions      [][5]float64
	Queries          []acquisitionTrainQuery
	Fits             []acquisitionTrainFit
	Arrivals         []acquisitionArrival `json:",omitempty"`
	EmptyQueryClocks []int                `json:",omitempty"`
}

type acquisitionArrival struct {
	Clock, Origin  int
	Paid, Surprise bool
	Mixer          bool
}

// Research only: acquisition changes expert fitting as well as the original
// issued-forecast mixer. Truth is read only when a delivery becomes available.
func acquisitionTrainRun(input softV120Record, policy, stop int) (acquisitionTrainResult, error) {
	return acquisitionTrainRunTiming(input, policy, stop, false)
}

func acquisitionTrainRunTiming(input softV120Record, policy, stop int, aligned bool) (acquisitionTrainResult, error) {
	mode := 0
	if aligned {
		mode = 1
	}
	return acquisitionTrainRunMode(input, policy, stop, mode)
}

// Modes0/1 preserve the earlier eight-origin experiments; modes2/3 share
// the frozen32-origin pool and differ only in periodic versus burst timing.
func acquisitionTrainRunMode(input softV120Record, policy, stop, mode int) (acquisitionTrainResult, error) {
	return acquisitionTrainRunCadence(input, policy, stop, mode, 32)
}

// The faster cadence is an explicit extra-compute diagnostic, not a default.
func acquisitionTrainRunCadence(input softV120Record, policy, stop, mode, stride int) (acquisitionTrainResult, error) {
	var out acquisitionTrainResult
	if len(input.Steps) != 256 || policy < 0 || policy > 3 || stop < 0 || stop > 256 || mode < 0 || mode > 3 {
		return out, fmt.Errorf("invalid acquisition run")
	}
	if stride != 32 && (stride != 8 || policy != 0 || mode != 0) {
		return out, fmt.Errorf("invalid diagnostic cadence")
	}
	control, err := newMarkovAdvice(.001)
	if err != nil {
		return out, err
	}
	var known [256]bool
	var paid [256]int
	var tables [4][512]float64
	var scheduler queryBurstSchedule
	for t := 0; t < stop; t++ {
		surprise := false
		observe := func(j int) {
			if mode < 2 || policy == 0 {
				return
			}
			p := out.Predictions[j][4]
			if !input.Steps[j].Y {
				p = 1 - p
			}
			unusual := p < .2
			surprise = surprise || unusual
			out.Arrivals = append(out.Arrivals, acquisitionArrival{t, j, paid[j] > 0 && paid[j] <= t, unusual, j >= t-32})
		}
		for j := 0; j < t; j++ {
			s := input.Steps[j]
			if !known[j] && ((!s.Missing && j+s.Delay <= t) || (paid[j] > 0 && paid[j] <= t)) {
				known[j] = true
				observe(j)
				// Late acquired labels remain valid training evidence, but cannot
				// reopen a journal entry censored on an earlier clock.
				if mode >= 2 && j < t-32 {
					continue
				}
				if err := control.deliver(uint64(j), s.Y); err != nil {
					return out, err
				}
			}
		}
		if t >= 32 {
			if err := control.expireBefore(uint64(t - 31)); err != nil {
				return out, err
			}
		}
		if t%stride == 0 {
			available := make([]int, 0, t+16)
			for j := -16; j < t; j++ {
				if j < 0 || known[j] {
					available = append(available, j)
				}
			}
			fit := acquisitionTrainFit{Clock: t}
			for window, cap := range []int{64, 32} {
				origins := available
				if len(origins) > cap {
					origins = origins[len(origins)-cap:]
				}
				fit.Origins[window] = append([]int(nil), origins...)
				samples := make([]observation.Sample, len(origins))
				for k, j := range origins {
					if j < 0 {
						samples[k] = input.Initial[j+16]
					} else {
						samples[k] = observation.Sample{Bits: input.Steps[j].X, Outcome: input.Steps[j].Y}
					}
				}
				g, err := fitSubset(samples)
				if err != nil {
					return out, err
				}
				b, err := fitBooleanSpecialist(samples)
				if err != nil {
					return out, err
				}
				tables[window*2], tables[window*2+1] = g.predictions, b.predictions
			}
			out.Fits = append(out.Fits, fit)
		}
		x := input.Steps[t].X
		if x >= 512 {
			return out, fmt.Errorf("invalid input")
		}
		var prediction [5]float64
		var raw [4]float64
		weights := control.current.weights()
		for k := range raw {
			raw[k] = tables[k][x]
			prediction[k] = raw[k]
			prediction[4] += weights[k+1] * raw[k]
		}
		if err := control.issue(uint64(t), raw); err != nil {
			return out, err
		}
		out.Predictions = append(out.Predictions, prediction)
		s := input.Steps[t]
		if !s.Missing && s.Delay == 0 {
			known[t] = true
			observe(t)
			if err := control.deliver(uint64(t), s.Y); err != nil {
				return out, err
			}
		}
		queryEnd := t
		queryDue := t > 0 && t%8 == 0
		if mode == 1 {
			if t > 0 && t <= 224 && t%32 == 0 {
				queryDue = false
			}
			// Forecast first, acquire second, reveal before tomorrow's fit.
			if t < 224 && t%32 == 31 {
				queryDue = true
				queryEnd = t + 1
			}
		}
		if policy == 0 || (mode != 3 && !queryDue) {
			continue
		}
		var pool []int
		width := 8
		if mode >= 2 {
			width = 32
		}
		for j := max(0, queryEnd-width); j < queryEnd; j++ {
			if !known[j] && paid[j] == 0 {
				pool = append(pool, j)
			}
		}
		if mode == 3 {
			queryDue = scheduler.step(t, surprise, len(pool) > 0)
		}
		if len(pool) == 0 {
			if mode >= 2 {
				out.EmptyQueryClocks = append(out.EmptyQueryClocks, t)
			}
			continue
		}
		if !queryDue {
			continue
		}
		if mode >= 2 {
			weights = control.current.weights()
		}
		selected := -1
		if policy == 1 {
			sum := sha256.Sum256([]byte(fmt.Sprintf("%d:%d:%d:%d:%d", input.Phase, input.Case, input.Index, input.Schedule, queryEnd)))
			selected = pool[int(uint64(binary.BigEndian.Uint32(sum[:4]))*uint64(len(pool))>>32)]
		} else {
			best := math.Inf(-1)
			for _, j := range pool {
				m, conditional := 0., 0.
				for k := 0; k < 4; k++ {
					p := out.Predictions[j][k]
					m += weights[k+1] * p
					conditional += weights[k+1] * acquisitionEntropy(p)
				}
				score := acquisitionEntropy(m)
				if policy == 3 {
					score -= conditional
				}
				if score > best {
					best = score
					selected = j
				}
			}
		}
		if paid[selected] != 0 {
			return out, fmt.Errorf("duplicate paid query")
		}
		paid[selected] = t + 1
		out.Queries = append(out.Queries, acquisitionTrainQuery{t, selected, t + 1})
	}
	return out, nil
}
func acquisitionEntropy(p float64) float64 {
	p = math.Max(1e-12, math.Min(1-1e-12, p))
	return -p*math.Log(p) - (1-p)*math.Log1p(-p)
}

func TestAcquisitionTrainResearchContracts(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("explicit consumed artifact required")
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var header softV120Artifact
	if err := decoder.Decode(&header); err != nil {
		t.Fatal(err)
	}
	if header.Version != "soft-learners-v120" {
		t.Fatal("wrong artifact")
	}
	checked := 0
	for {
		var input softV120Record
		if err := decoder.Decode(&input); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if input.Phase != 0 || input.Index != 0 {
			continue
		}
		if input.Case != 0 && input.Case != 5 && input.Case != 8 && input.Case != 12 && input.Case != 19 && input.Case != 20 {
			continue
		}
		natural, err := acquisitionTrainRun(input, 0, 256)
		if err != nil {
			t.Fatal(err)
		}
		for j, p := range natural.Predictions {
			for k, original := range []int{0, 1, 2, 3, 12} {
				if math.Abs(p[k]-input.Steps[j].P[original]) > 1e-12 {
					t.Fatalf("control mismatch case%d/schedule%d/frame%d/arm%d", input.Case, input.Schedule, j, k)
				}
			}
		}
		for j, fit := range natural.Fits {
			if !reflect.DeepEqual(fit.Origins, input.Fits[j].Origins) {
				t.Fatal("control fit origins mismatch")
			}
		}
		if input.Case == 20 {
			var count = -1
			for policy := 1; policy <= 3; policy++ {
				paid, err := acquisitionTrainRun(input, policy, 256)
				if err != nil {
					t.Fatal(err)
				}
				if count >= 0 && len(paid.Queries) != count {
					t.Fatal("unequal budget")
				}
				count = len(paid.Queries)
				if input.Schedule == 0 && !reflect.DeepEqual(paid.Predictions, natural.Predictions) {
					t.Fatal("complete delivery changed")
				}
				if input.Schedule == 0 && len(paid.Queries) != 0 {
					t.Fatal("paid for already available evidence")
				}
				if input.Schedule == 1 && (reflect.DeepEqual(paid.Fits, natural.Fits) || reflect.DeepEqual(paid.Predictions, natural.Predictions)) {
					t.Fatal("paid acquisition did not affect learning")
				}
				for _, clock := range []int{32, 160} {
					changed := input
					changed.Steps = append([]softV120Step(nil), input.Steps...)
					paidKnown := map[int]bool{}
					for _, q := range paid.Queries {
						if q.Reveal <= clock {
							paidKnown[q.Origin] = true
						}
					}
					for j := range changed.Steps {
						s := &changed.Steps[j]
						s.Q = 1 - s.Q
						if j >= clock || ((s.Missing || j+s.Delay > clock) && !paidKnown[j]) {
							s.Y = !s.Y
						}
					}
					other, err := acquisitionTrainRun(changed, policy, clock+1)
					if err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(other.Predictions, paid.Predictions[:clock+1]) {
						t.Fatal("as-of leak")
					}
					var prefix []acquisitionTrainQuery
					for _, q := range paid.Queries {
						if q.Clock <= clock {
							prefix = append(prefix, q)
						}
					}
					if !reflect.DeepEqual(other.Queries, prefix) {
						t.Fatal("acquisition choice leak")
					}
				}
				for _, fit := range paid.Fits {
					for _, origins := range fit.Origins {
						for _, j := range origins {
							if j < 0 {
								continue
							}
							s := input.Steps[j]
							ok := !s.Missing && j+s.Delay <= fit.Clock
							for _, q := range paid.Queries {
								ok = ok || (q.Origin == j && q.Reveal <= fit.Clock)
							}
							if j >= fit.Clock || !ok {
								t.Fatal("unrevealed training label")
							}
						}
					}
				}
			}
		}
		checked++
	}
	if checked != 12 {
		t.Fatalf("checked%d want12", checked)
	}
	t.Log("12 natural replay fixtures and paid acquisition/as-of contracts PASS")
}
