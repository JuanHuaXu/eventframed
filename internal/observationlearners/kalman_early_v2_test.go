package observationlearners

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
	"unsafe"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

var earlyV2Scenarios = [...]int{0, 2, 5, 7, 8}
var earlyV2Noise = [...]float64{.0005, .002, .01}

type earlyV2Tick struct {
	X                 uint16
	Y, Audit, Missing bool
	P                 [5]float64
	Delivered         []int `json:",omitempty"`
}

type earlyV2Metric struct {
	Full, Early, Tail [5]float64
}

type earlyV2Row struct {
	Kind, Split, Scenario string
	Fit, Stream           int
	Ticks                 []earlyV2Tick
	Metric                earlyV2Metric
	Audits, Available     int
	Updates, Pending      int
	UpdateP99NS           [3]int64
	StateBytes            int
}

type earlyV2Packet struct {
	Origin, Due int
	X           uint16
	Y, Audit    bool
	Base        float64
	H           [kalmanDim]float64
}

// The tape and all observations are shared by the five arms. A due packet is
// applied only after the current tick's forecasts have been recorded.
func runEarlyV2(base *observation.Model, split string, scenario, fit, stream int, seed int64) (earlyV2Row, error) {
	s := Scenarios[scenario]
	r := earlyV2Row{Kind: "trial", Split: split, Scenario: s.Name, Fit: fit, Stream: stream,
		StateBytes: int(unsafe.Sizeof(KalmanResidual{}))}
	var k [3]KalmanResidual
	for i := range k {
		k[i] = NewKalmanResidual()
	}
	rng := rand.New(rand.NewSource(Seed(seed, scenario, fit, stream, 0)))
	audit := rand.New(rand.NewSource(Seed(seed, scenario, fit, stream, 1)))
	missing := rand.New(rand.NewSource(Seed(seed, scenario, fit, stream, 2)))
	var short *observation.Model
	var samples []observation.Sample
	var packets []earlyV2Packet
	var durations [3][]int64
	for clock := 0; clock < 512; clock++ {
		x := uint16(rng.Intn(512))
		pb := forecast(base, x)
		h := kalmanFeatures(x, pb)
		tick := earlyV2Tick{X: x, Audit: audit.Float64() < .25, Missing: missing.Float64() < s.Missing}
		tick.P[0], tick.P[1] = pb, forecast(short, x)
		for i := range k {
			k[i].Advance()
			for j := range k[i].Cov {
				k[i].Cov[j][j] += earlyV2Noise[i] - earlyV2Noise[0]
			}
			tick.P[i+2] = k[i].Predict(pb, h)
		}
		tick.Y = truth(x, clock, s, rng)
		if !tick.Missing {
			packets = append(packets, earlyV2Packet{Origin: clock, Due: clock + s.Delay,
				X: x, Y: tick.Y, Audit: tick.Audit, Base: pb, H: h})
		}
		for len(packets) > 0 && packets[0].Due <= clock {
			p := packets[0]
			packets = packets[1:]
			if p.Origin > clock || p.Due != clock {
				return r, errors.New("future or stale feedback delivery")
			}
			tick.Delivered = append(tick.Delivered, p.Origin)
			r.Available++
			if !p.Audit {
				continue
			}
			r.Audits++
			samples = append(samples, observation.Sample{Bits: p.X, Outcome: p.Y})
			for i := range k {
				start := time.Now()
				if err := k[i].Observe(p.Base, p.H, p.Y); err != nil {
					return r, err
				}
				durations[i] = append(durations[i], time.Since(start).Nanoseconds())
			}
			if r.Audits >= 32 && r.Audits%16 == 0 {
				var err error
				short, err = observation.Fit(samples[max(0, len(samples)-64):])
				if err != nil {
					return r, err
				}
			}
		}
		r.Ticks = append(r.Ticks, tick)
	}
	r.Pending, r.Updates = len(packets), r.Audits
	if r.Available+r.Pending > 512 || r.Audits > r.Available || r.StateBytes > 4096 {
		return r, errors.New("invalid feedback or state accounting")
	}
	for i := range durations {
		if len(durations[i]) != r.Updates {
			return r, errors.New("missing filter updates")
		}
		sorted := append([]int64(nil), durations[i]...)
		sortDurations(sorted)
		if len(sorted) > 0 {
			r.UpdateP99NS[i] = sorted[int(math.Ceil(.99*float64(len(sorted))))-1]
		}
	}
	start := s.Change
	if start >= 512 {
		start = 0
	}
	for clock, tick := range r.Ticks {
		for i, p := range tick.P {
			if p < 0 || p > 1 || math.IsNaN(p) || math.IsInf(p, 0) {
				return r, errors.New("invalid forecast")
			}
			loss := contrastFreeBrier(p, tick.Y)
			r.Metric.Full[i] += loss / 512
			if clock >= start && clock < start+64 {
				r.Metric.Early[i] += loss / 64
			}
			if clock >= 384 {
				r.Metric.Tail[i] += loss / 128
			}
		}
	}
	return r, nil
}

func contrastFreeBrier(p float64, y bool) float64 {
	if y {
		return (1 - p) * (1 - p)
	}
	return p * p
}

func sortDurations(v []int64) {
	for i := 1; i < len(v); i++ {
		for j := i; j > 0 && v[j] < v[j-1]; j-- {
			v[j], v[j-1] = v[j-1], v[j]
		}
	}
}

func TestKalmanEarlyV2(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_KALMAN_EARLY_OUT"), os.Getenv("EVENTFRAME_KALMAN_EARLY_SPLIT")
	if path == "" && split == "" {
		t.Skip("opt-in research study")
	}
	if path == "" || (split != "design" && split != "confirmation") {
		t.Fatal("output path and design|confirmation split required")
	}
	seed := int64(2026102101)
	fitOffset := 200
	if split == "confirmation" {
		seed = 2026102102
		fitOffset = 300
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := make(map[string]string)
	for _, name := range []string{"docs/experiments/mmm-kalman-early-v2-protocol.md",
		"internal/observationlearners/kalman_early_v2_test.go", "internal/observationlearners/kalman_window.go",
		"internal/observationlearners/experiment.go", "internal/observation/forecast.go"} {
		b, e := os.ReadFile(filepath.Join(root, name))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(map[string]any{"kind": "manifest", "split": split, "seedBase": seed,
		"fitOffset": fitOffset, "hashes": hashes, "processNoise": earlyV2Noise,
		"scenarios": earlyV2Scenarios}); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range earlyV2Scenarios {
		for fit := 0; fit < 4; fit++ {
			base, e := Base(Scenarios[scenario], scenario, fit+fitOffset)
			if e != nil {
				t.Fatal(e)
			}
			for stream := 0; stream < 8; stream++ {
				r, e := runEarlyV2(base, split, scenario, fit, stream, seed)
				if e != nil {
					t.Fatalf("scenario=%s fit=%d stream=%d: %v", Scenarios[scenario].Name, fit, stream, e)
				}
				if e := enc.Encode(r); e != nil {
					t.Fatal(e)
				}
			}
		}
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}

func TestKalmanEarlyV2NoFutureFeedback(t *testing.T) {
	base, err := Base(Scenarios[7], 7, 200)
	if err != nil {
		t.Fatal(err)
	}
	r, err := runEarlyV2(base, "design", 7, 0, 0, 2026102101)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= 16; i++ {
		for k := 2; k < 5; k++ {
			if r.Ticks[i].P[k] != r.Ticks[i].P[0] {
				t.Fatalf("filter saw undelivered label at clock=%d arm=%d", i, k)
			}
		}
		for _, origin := range r.Ticks[i].Delivered {
			if origin+16 != i {
				t.Fatalf("invalid delivery origin=%d clock=%d", origin, i)
			}
		}
	}
}
