package observationlearners

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestBrierLifecycle(t *testing.T) {
	for _, rho := range []float64{-1, 1, math.NaN(), math.Inf(1)} {
		if _, e := newBrierAggregator(rho); e == nil {
			t.Fatal("invalid share")
		}
	}
	var zero brierAggregator
	if _, e := zero.predict(0, [2]float64{0, 1}); e == nil {
		t.Fatal("zero state")
	}
	s, _ := newBrierAggregator(.001)
	for _, p := range [][2]float64{{-1, 0}, {0, 2}, {math.NaN(), 0}, {0, math.Inf(1)}} {
		before := *s
		if _, e := s.predict(0, p); e == nil || before != *s {
			t.Fatal("invalid prediction mutation")
		}
	}
	before := *s
	if e := s.observe(0, true); e == nil || before != *s {
		t.Fatal("unissued outcome")
	}
	if _, e := s.predict(1, [2]float64{0, 1}); e == nil || before != *s {
		t.Fatal("sequence mutation")
	}
	f, e := s.predict(0, [2]float64{.2, .8})
	if e != nil {
		t.Fatal(e)
	}
	stored := s.pending
	f.P = 1
	f.Experts[0] = 1
	if s.pending != stored {
		t.Fatal("returned forecast aliases state")
	}
	before = *s
	if _, e := s.predict(1, [2]float64{0, 1}); e == nil || before != *s {
		t.Fatal("double issuance")
	}
	if e := s.observe(1, false); e == nil || before != *s {
		t.Fatal("wrong outcome id")
	}
	if e := s.observe(0, false); e != nil {
		t.Fatal(e)
	}
	before = *s
	if e := s.observe(0, true); e == nil || before != *s {
		t.Fatal("duplicate outcome")
	}
	s.issued = math.MaxUint64
	before = *s
	if _, e := s.predict(math.MaxUint64, [2]float64{}); e == nil || before != *s {
		t.Fatal("counter overflow")
	}
}

func TestBrierRecoveryAndMissingFeedback(t *testing.T) {
	s, _ := newBrierAggregator(0)
	for step := uint64(0); step < 4010; step++ {
		if _, e := s.predict(step, [2]float64{0, 1}); e != nil {
			t.Fatal(e)
		}
		if e := s.observe(step, step >= 2000); e != nil {
			t.Fatal(e)
		}
		if math.IsInf(s.logWeights[0], 0) || math.IsInf(s.logWeights[1], 0) {
			t.Fatal("irreversible underflow")
		}
	}
	if s.weights()[1] <= .5 {
		t.Fatal("lost expert failed to recover")
	}
	// Reusing the initial forecast without feedback is outside the contract.
	// Its loss against the perfect generic expert exceeds the claimed bound.
	loss := 1000 * math.Pow(1-brierGenericPrior, 2)
	if loss <= brierGenericBound(1000, 0) {
		t.Fatal("missing-feedback falsifier vacuous")
	}
	x, _ := newBrierAggregator(0)
	if _, e := x.predict(0, [2]float64{0, 1}); e != nil {
		t.Fatal(e)
	}
	if _, e := x.predict(1, [2]float64{0, 1}); e == nil {
		t.Fatal("missing feedback silently skipped")
	}
	if brierGenericBound(16, .001)/16 >= .01 {
		t.Fatal("share budget")
	}
}

var brierCases = []string{"random", "adaptive_adversary", "alternating_extremes", "regime_flip", "changing_experts", "generic_better", "challenger_better"}

type brierKernelRecord struct {
	Case                                             string
	Index, Steps                                     int
	Share                                            float64
	Loss                                             [3]float64
	MaxRegret, MaxBoundViolation, MaxJensenViolation float64
	FinalWeights                                     [2]float64
	RecoveryLag, PostCorrect                         int
	Tape                                             string
}

func brierPanelRun(scenario, index int) ([2]brierKernelRecord, error) {
	var result [2]brierKernelRecord
	states := [2]*brierAggregator{}
	for arm, rho := range []float64{0, .001} {
		states[arm], _ = newBrierAggregator(rho)
		result[arm] = brierKernelRecord{Case: brierCases[scenario], Index: index, Steps: 4096, Share: rho, RecoveryLag: -1}
	}
	seed := int64(2042119400 + scenario*10000 + index*10)
	xs := rand.New(rand.NewSource(seed))
	ys := rand.New(rand.NewSource(seed + 1))
	tapes := [2]hashWriter{{h: sha256.New()}, {h: sha256.New()}}
	for step := 0; step < 4096; step++ {
		experts := [2]float64{xs.Float64(), xs.Float64()}
		switch scenario {
		case 2, 3:
			experts = [2]float64{0, 1}
		case 4:
			experts = [2]float64{float64(step%17) / 16, 1 - float64(step%17)/16}
		case 5:
			experts = [2]float64{.2, .8}
		case 6:
			experts = [2]float64{.8, .2}
		}
		var forecasts [2]brierForecast
		for arm, s := range states {
			f, e := s.predict(uint64(step), experts)
			if e != nil {
				return result, e
			}
			forecasts[arm] = f
		}
		y := ys.Float64() < .5
		switch scenario {
		case 1:
			harm := func(y float64) float64 {
				v := 0.
				for _, f := range forecasts {
					v += (f.P-y)*(f.P-y) - (experts[0]-y)*(experts[0]-y)
				}
				return v
			}
			y = harm(1) > harm(0)
		case 2:
			y = step%2 == 1
		case 3:
			y = step >= 2048
		case 4:
			y = (step/128)%2 == 1
		case 5, 6:
			y = false
		}
		target := 0.
		if y {
			target = 1
		}
		for arm, s := range states {
			r := &result[arm]
			f := forecasts[arm]
			loss := func(p float64) float64 { return (p - target) * (p - target) }
			r.Loss[0] += loss(experts[0])
			r.Loss[1] += loss(experts[1])
			r.Loss[2] += loss(f.P)
			regret := r.Loss[2] - r.Loss[0]
			r.MaxRegret = math.Max(r.MaxRegret, regret)
			r.MaxBoundViolation = math.Max(r.MaxBoundViolation, regret-brierGenericBound(uint64(step+1), r.Share))
			jensen := f.Weights[0]*math.Exp(-brierRate*loss(experts[0])) + f.Weights[1]*math.Exp(-brierRate*loss(experts[1])) - math.Exp(-brierRate*loss(f.P))
			r.MaxJensenViolation = math.Max(r.MaxJensenViolation, jensen)
			if scenario == 3 && step >= 2048 {
				if (f.P >= .5) == y {
					r.PostCorrect++
				}
				if r.RecoveryLag < 0 && f.Weights[1] > .5 {
					r.RecoveryLag = step - 2048
				}
			}
			tapes[arm].write(f, y)
			if e := s.observe(uint64(step), y); e != nil {
				return result, e
			}
		}
	}
	for arm, s := range states {
		result[arm].FinalWeights = s.weights()
		result[arm].Tape = hex.EncodeToString(tapes[arm].h.Sum(nil))
	}
	return result, nil
}

// Narrow interface keeps the tape writer independent of a concrete hash state.
type hashWriter struct {
	h interface {
		Write([]byte) (int, error)
		Sum([]byte) []byte
	}
}

func (w hashWriter) write(f brierForecast, y bool) {
	var b [49]byte
	binary.LittleEndian.PutUint64(b[:8], f.Sequence)
	for i, x := range []float64{f.Experts[0], f.Experts[1], f.P, f.Weights[0], f.Weights[1]} {
		binary.LittleEndian.PutUint64(b[8+i*8:16+i*8], math.Float64bits(x))
	}
	if y {
		b[48] = 1
	}
	w.h.Write(b[:])
}

func TestBrierPanelContracts(t *testing.T) {
	for scenario := range brierCases {
		rs, e := brierPanelRun(scenario, 0)
		if e != nil {
			t.Fatal(e)
		}
		for _, r := range rs {
			if r.MaxBoundViolation > 1e-8 || r.MaxJensenViolation > 1e-12 {
				t.Fatal("bound", r)
			}
		}
	}
}

func TestBrierKernelV94(t *testing.T) {
	out, replay := os.Getenv("EVENTFRAME_BRIER_V94_OUT"), os.Getenv("EVENTFRAME_BRIER_V94_REPLAY")
	if out == "" && replay == "" {
		t.Skip("explicit artifact path required")
	}
	if out != "" && replay != "" {
		t.Fatal("choose generation or replay")
	}
	type artifact struct {
		Protocol, Runtime string
		Hashes            map[string]string
		Records           []brierKernelRecord
	}
	a := artifact{Protocol: "mmm-brier-kernel-v94", Runtime: runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH, Hashes: map[string]string{}}
	for _, path := range []string{"go.mod", "go.sum", "internal/observationlearners/brier_online.go", "internal/observationlearners/brier_online_test.go", "docs/experiments/mmm-brier-kernel-v94-protocol.md", "research/online-brier-aggregation-proposal.md"} {
		b, e := os.ReadFile(filepath.Join("..", "..", path))
		if e != nil {
			t.Fatal(e)
		}
		sum := sha256.Sum256(b)
		a.Hashes[path] = hex.EncodeToString(sum[:])
	}
	var previous artifact
	if replay != "" {
		b, e := os.ReadFile(replay)
		if e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(b, &previous); e != nil {
			t.Fatal(e)
		}
		if previous.Protocol != a.Protocol || previous.Runtime != a.Runtime || !reflect.DeepEqual(previous.Hashes, a.Hashes) || len(previous.Records) != 448 {
			t.Fatal("artifact mismatch")
		}
	}
	for scenario := range brierCases {
		for index := 0; index < 32; index++ {
			rs, e := brierPanelRun(scenario, index)
			if e != nil {
				t.Fatal(e)
			}
			for _, r := range rs {
				if r.MaxBoundViolation > 1e-8 || r.MaxJensenViolation > 1e-12 {
					t.Fatal("bound failure", r)
				}
				if replay != "" && r != previous.Records[len(a.Records)] {
					t.Fatal("record mismatch")
				}
				a.Records = append(a.Records, r)
			}
		}
		t.Log(brierCases[scenario], "complete")
	}
	if out != "" {
		f, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal(e)
		}
		e = json.NewEncoder(f).Encode(a)
		ce := f.Close()
		if e != nil || ce != nil {
			t.Fatal(e, ce)
		}
	}
}

var brierForecastSink brierForecast

func BenchmarkBrierKernel(b *testing.B) {
	for _, rho := range []float64{0, .001} {
		name := "no_share"
		if rho > 0 {
			name = "share"
		}
		b.Run(name, func(b *testing.B) {
			s, _ := newBrierAggregator(rho)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f, e := s.predict(uint64(i), [2]float64{.2, .8})
				if e != nil {
					b.Fatal(e)
				}
				if e = s.observe(uint64(i), i%3 == 0); e != nil {
					b.Fatal(e)
				}
				brierForecastSink = f
			}
		})
	}
}
