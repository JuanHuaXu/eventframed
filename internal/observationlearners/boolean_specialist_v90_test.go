package observationlearners

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/bits"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

var booleanV90Cases = []string{"parity1", "parity2", "parity3", "parity4", "complement4", "majority3", "mux3", "constant", "null", "dependent4"}

type booleanV90Metric struct{ Brier, Accuracy float64 }
type booleanV90Record struct {
	Phase, Case                   string
	Index, N                      int
	Rule                          uint16
	TrainSHA256, PredictionSHA256 string
	// Arm order: unchanged subset, Boolean specialist. View order: full, mask63.
	Metrics [2][2]booleanV90Metric
	Oracle  [2]float64
}
type booleanV90Artifact struct {
	Protocol, Runtime string
	Hashes            map[string]string
	Records           []booleanV90Record
}

func booleanV90Cost(mask uint16) int {
	mask |= 1
	cost := 0
	for shift := 0; shift < 9; shift += 3 {
		cost += bits.Len16((mask >> shift) & 7)
	}
	return cost
}
func booleanV90Masks(k, phase int) []uint16 {
	var pool []uint16
	index := 0
	for mask := uint16(1); mask < 512; mask++ {
		if bits.OnesCount16(mask) == k && booleanV90Cost(mask) <= 6 {
			if index%2 == phase {
				pool = append(pool, mask)
			}
			index++
		}
	}
	return pool
}
func booleanV90Input(x uint16, name string) uint16 {
	if name == "dependent4" {
		x = (x &^ 256) | (((x >> 1) & 1) << 8)
	}
	return x
}
func booleanV90Truth(x, rule uint16, name string) float64 {
	if name == "null" {
		return .5
	}
	y := bits.OnesCount16(x&rule)%2 == 1
	switch name {
	case "complement4":
		y = !y
	case "constant":
		y = false
	case "majority3":
		y = bits.OnesCount16(x&rule) >= 2
	case "mux3":
		var selected [3]bool
		i := 0
		for bit := uint16(1); bit < 512; bit <<= 1 {
			if rule&bit != 0 {
				selected[i] = x&bit != 0
				i++
			}
		}
		y = selected[2]
		if selected[0] {
			y = selected[1]
		}
	}
	if y {
		return .95
	}
	return .05
}
func booleanV90Data(phase, scenario, index int) (uint16, []observation.Sample) {
	seed := int64(2026119000 + phase*1000000 + scenario*10000 + index*10)
	rules := rand.New(rand.NewSource(seed))
	xs := rand.New(rand.NewSource(seed + 1))
	ys := rand.New(rand.NewSource(seed + 2))
	arities := [10]int{1, 2, 3, 4, 4, 3, 3, 0, 0, 4}
	rule := uint16(0)
	if k := arities[scenario]; k != 0 {
		pool := booleanV90Masks(k, phase)
		rule = pool[rules.Intn(len(pool))]
	}
	name := booleanV90Cases[scenario]
	s := make([]observation.Sample, 64)
	for i := range s {
		x := booleanV90Input(uint16(xs.Intn(512)), name)
		s[i] = observation.Sample{Bits: x, Outcome: ys.Float64() < booleanV90Truth(x, rule, name)}
	}
	return rule, s
}
func booleanV90SHA(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
func booleanV90Score(phase, scenario, index, n int) (booleanV90Record, error) {
	rule, samples := booleanV90Data(phase, scenario, index)
	samples = samples[:n]
	r := booleanV90Record{Phase: []string{"design", "confirmation"}[phase], Case: booleanV90Cases[scenario], Index: index, N: n, Rule: rule}
	raw, err := json.Marshal(samples)
	if err != nil {
		return r, err
	}
	r.TrainSHA256 = booleanV90SHA(raw)
	var w [512]float64
	for i := range w {
		w[i] = 1
	}
	control, err := NewSubsetConditional(samples, w)
	if err != nil {
		return r, err
	}
	candidate, err := NewBooleanConditional(samples)
	if err != nil {
		return r, err
	}
	var predictions [2][2][512]float64
	for view, mask := range []uint16{511, 63} {
		var mass, positive [512]float64
		for rawX := uint16(0); rawX < 512; rawX++ {
			x := booleanV90Input(rawX, r.Case)
			truth := booleanV90Truth(x, rule, r.Case)
			mass[x&mask]++
			positive[x&mask] += truth
			for arm, m := range []*ConditionalForest{control, candidate} {
				p, err := m.Forecast(mask, x&mask)
				if err != nil || math.IsNaN(p) || p <= 0 || p >= 1 {
					return r, fmt.Errorf("invalid forecast %v %v", p, err)
				}
				predictions[arm][view][rawX] = p
				metric := &r.Metrics[arm][view]
				metric.Brier += ((p-truth)*(p-truth) + truth*(1-truth)) / 512
				accuracy := 1 - truth
				if p >= .5 {
					accuracy = truth
				}
				metric.Accuracy += accuracy / 512
			}
		}
		for i, count := range mass {
			if count > 0 {
				p := positive[i] / count
				r.Oracle[view] += count / 512 * p * (1 - p)
			}
		}
	}
	raw, err = json.Marshal(predictions)
	if err != nil {
		return r, err
	}
	r.PredictionSHA256 = booleanV90SHA(raw)
	return r, nil
}

func TestBooleanV90Generator(t *testing.T) {
	for k := 1; k <= 4; k++ {
		seen := map[uint16]bool{}
		for phase := 0; phase < 2; phase++ {
			pool := booleanV90Masks(k, phase)
			if len(pool) == 0 {
				t.Fatal("empty rule pool")
			}
			for _, mask := range pool {
				if seen[mask] || bits.OnesCount16(mask) != k || booleanV90Cost(mask) > 6 {
					t.Fatal("invalid/disjoint rule pool")
				}
				seen[mask] = true
			}
		}
	}
	for scenario := range booleanV90Cases {
		r, err := booleanV90Score(0, scenario, 0, 16)
		if err != nil {
			t.Fatal(err)
		}
		for _, arm := range r.Metrics {
			for view, m := range arm {
				if m.Brier+1e-12 < r.Oracle[view] || m.Brier > 1 || m.Accuracy < 0 || m.Accuracy > 1 {
					t.Fatal("score bounds")
				}
			}
		}
		if r.Case == "null" {
			for _, arm := range r.Metrics {
				for _, m := range arm {
					if math.Abs(m.Accuracy-.5) > 1e-12 || m.Brier < .25-1e-12 {
						t.Fatal("null evaluator")
					}
				}
			}
		}
	}
}

func TestBooleanV90(t *testing.T) {
	out, replay := os.Getenv("EVENTFRAME_BOOLEAN_V90_OUT"), os.Getenv("EVENTFRAME_BOOLEAN_V90_REPLAY")
	if out == "" && replay == "" {
		t.Skip("explicit experiment destination or replay required")
	}
	if out != "" && replay != "" {
		t.Fatal("choose generation or replay")
	}
	root := filepath.Join("..", "..")
	a := booleanV90Artifact{Protocol: "mmm-boolean-specialist-v90", Runtime: runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH, Hashes: map[string]string{}}
	paths := []string{"go.mod", "go.sum", "docs/experiments/mmm-boolean-specialist-v90-protocol.md", "research/boolean-specialist-proposal.md", "internal/observationlearners/boolean_specialist.go", "internal/observationlearners/boolean_specialist_test.go", "internal/observationlearners/boolean_specialist_v90_test.go", "internal/observationlearners/subset.go", "internal/observationlearners/subset_test.go", "internal/observationlearners/conditional.go", "internal/observation/controller.go"}
	for _, path := range paths {
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		a.Hashes[path] = booleanV90SHA(data)
	}
	var previous booleanV90Artifact
	if replay != "" {
		raw, err := os.ReadFile(replay)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, &previous); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(previous.Hashes, a.Hashes) || previous.Protocol != a.Protocol || previous.Runtime != a.Runtime || len(previous.Records) != 3840 {
			t.Fatal("artifact contract mismatch")
		}
	}
	for phase := 0; phase < 2; phase++ {
		for scenario, name := range booleanV90Cases {
			for index := 0; index < 64; index++ {
				for _, n := range []int{16, 32, 64} {
					r, err := booleanV90Score(phase, scenario, index, n)
					if err != nil {
						t.Fatal(err)
					}
					if replay != "" && r != previous.Records[len(a.Records)] {
						t.Fatal("record mismatch", phase, scenario, index, n)
					}
					a.Records = append(a.Records, r)
				}
			}
			t.Log(phase, name, "complete")
		}
	}
	if out != "" {
		f, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		err = json.NewEncoder(f).Encode(a)
		closeErr := f.Close()
		if err != nil || closeErr != nil {
			t.Fatal(err, closeErr)
		}
	}
}
