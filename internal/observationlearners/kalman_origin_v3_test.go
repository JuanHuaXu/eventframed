package observationlearners

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
	"unsafe"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

const originV3DesignSeed int64 = 2026102201
const originV3ConfirmationSeed int64 = 2026102202

var originV3Scenarios = [...]int{0, 2, 5, 7, 8}
var originV3Noise = [...]float64{.0005, .01}

type originV3Filter struct {
	State      KalmanResidual
	LastOrigin int
	Noise      float64
}

func newOriginV3Filter(noise float64) originV3Filter {
	return originV3Filter{State: NewKalmanResidual(), LastOrigin: -1, Noise: noise}
}

func (f *originV3Filter) stateAt(clock int) (KalmanResidual, error) {
	if clock <= f.LastOrigin {
		return KalmanResidual{}, fmt.Errorf("nonmonotone filter clock %d after %d", clock, f.LastOrigin)
	}
	next := f.State
	for i := range next.Cov {
		next.Cov[i][i] += float64(clock-f.LastOrigin) * f.Noise
	}
	return next, nil
}

func (f *originV3Filter) predictAt(clock int, base float64, features [kalmanDim]float64) (float64, error) {
	state, err := f.stateAt(clock)
	if err != nil {
		return 0, err
	}
	return state.Predict(base, features), nil
}

func (f *originV3Filter) observeAt(origin int, base float64, features [kalmanDim]float64, outcome bool) error {
	state, err := f.stateAt(origin)
	if err != nil {
		return err
	}
	if err := state.Observe(base, features, outcome); err != nil {
		return err
	}
	f.State, f.LastOrigin = state, origin
	return nil
}

type originV3Metrics struct {
	Full, Early, Tail, DelayedBlind, DelayedAfter [7]float64
}

type originV3Row struct {
	Kind, Split, Scenario string
	Fit, Stream           int
	Control               earlyV2Row
	OriginPredictions     [2][]float64
	Metric                originV3Metrics
	ForecastP99NS         [2]int64
	UpdateP99NS           [2]int64
	OriginStateBytes      int
}

func originV3P99(durations []int64) int64 {
	if len(durations) == 0 {
		return 0
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	return durations[int(math.Ceil(.99*float64(len(durations))))-1]
}

func originV3Brier(ticks []earlyV2Tick, predictions []float64, start, end int) float64 {
	sum := 0.0
	for i := start; i < end; i++ {
		sum += contrastFreeBrier(predictions[i], ticks[i].Y)
	}
	return sum / float64(end-start)
}

func runOriginV3(base *observation.Model, split string, scenario, fit, stream int, seed int64) (originV3Row, error) {
	control, err := runEarlyV2(base, split, scenario, fit, stream, seed)
	if err != nil {
		return originV3Row{}, err
	}
	row := originV3Row{Kind: "trial", Split: split, Scenario: control.Scenario,
		Fit: fit, Stream: stream, Control: control, OriginStateBytes: int(unsafe.Sizeof(originV3Filter{}))}
	var filters [2]originV3Filter
	var predictionDurations, updateDurations [2][]int64
	for arm := range filters {
		filters[arm] = newOriginV3Filter(originV3Noise[arm])
		row.OriginPredictions[arm] = make([]float64, len(control.Ticks))
	}
	for clock, tick := range control.Ticks {
		features := kalmanFeatures(tick.X, tick.P[0])
		for arm := range filters {
			start := time.Now()
			prediction, err := filters[arm].predictAt(clock, tick.P[0], features)
			predictionDurations[arm] = append(predictionDurations[arm], time.Since(start).Nanoseconds())
			if err != nil || !isFinite(prediction) || prediction < 0 || prediction > 1 {
				return row, fmt.Errorf("invalid origin prediction clock=%d arm=%d: %v", clock, arm, err)
			}
			row.OriginPredictions[arm][clock] = prediction
		}
		for _, origin := range tick.Delivered {
			if origin < 0 || origin > clock || origin >= len(control.Ticks) ||
				origin+Scenarios[scenario].Delay != clock {
				return row, fmt.Errorf("invalid delivery origin=%d clock=%d", origin, clock)
			}
			packet := control.Ticks[origin]
			if packet.Missing {
				return row, errors.New("missing packet was delivered")
			}
			if !packet.Audit {
				continue
			}
			packetFeatures := kalmanFeatures(packet.X, packet.P[0])
			for arm := range filters {
				start := time.Now()
				err := filters[arm].observeAt(origin, packet.P[0], packetFeatures, packet.Y)
				updateDurations[arm] = append(updateDurations[arm], time.Since(start).Nanoseconds())
				if err != nil {
					return row, fmt.Errorf("invalid origin update clock=%d origin=%d arm=%d: %w", clock, origin, arm, err)
				}
			}
		}
	}
	for arm := range filters {
		if len(updateDurations[arm]) != control.Updates {
			return row, fmt.Errorf("origin update count=%d control=%d", len(updateDurations[arm]), control.Updates)
		}
		row.ForecastP99NS[arm] = originV3P99(predictionDurations[arm])
		row.UpdateP99NS[arm] = originV3P99(updateDurations[arm])
	}
	metrics := &row.Metric
	start := Scenarios[scenario].Change
	if start >= 512 {
		start = 0
	}
	for arm := 0; arm < 7; arm++ {
		predictions := make([]float64, len(control.Ticks))
		for clock, tick := range control.Ticks {
			if arm < 5 {
				predictions[clock] = tick.P[arm]
			} else {
				predictions[clock] = row.OriginPredictions[arm-5][clock]
			}
		}
		metrics.Full[arm] = originV3Brier(control.Ticks, predictions, 0, 512)
		metrics.Early[arm] = originV3Brier(control.Ticks, predictions, start, start+64)
		metrics.Tail[arm] = originV3Brier(control.Ticks, predictions, 384, 512)
		if scenario == 7 {
			metrics.DelayedBlind[arm] = originV3Brier(control.Ticks, predictions, 256, 272)
			metrics.DelayedAfter[arm] = originV3Brier(control.Ticks, predictions, 272, 336)
		}
	}
	return row, nil
}

func TestKalmanOriginV3ZeroDelayParity(t *testing.T) {
	base, err := Base(Scenarios[2], 2, 400)
	if err != nil {
		t.Fatal(err)
	}
	row, err := runOriginV3(base, "design", 2, 0, 0, originV3DesignSeed)
	if err != nil {
		t.Fatal(err)
	}
	for clock, tick := range row.Control.Ticks {
		for arm, control := range []int{2, 4} {
			if diff := math.Abs(row.OriginPredictions[arm][clock] - tick.P[control]); diff > 1e-10 {
				t.Fatalf("zero-delay clock=%d arm=%d diff=%g", clock, arm, diff)
			}
		}
	}
}

func TestKalmanOriginV3NoFutureFeedback(t *testing.T) {
	base, err := Base(Scenarios[7], 7, 400)
	if err != nil {
		t.Fatal(err)
	}
	row, err := runOriginV3(base, "design", 7, 0, 0, originV3DesignSeed)
	if err != nil {
		t.Fatal(err)
	}
	for clock := 0; clock <= 16; clock++ {
		for arm, control := range []int{2, 4} {
			if math.Abs(row.OriginPredictions[arm][clock]-row.Control.Ticks[clock].P[control]) > 1e-10 {
				t.Fatalf("pre-feedback forecast changed at clock=%d arm=%d", clock, arm)
			}
		}
	}
}

func TestKalmanOriginV3(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_KALMAN_ORIGIN_OUT"), os.Getenv("EVENTFRAME_KALMAN_ORIGIN_SPLIT")
	if path == "" && split == "" {
		t.Skip("opt-in research study")
	}
	if path == "" || (split != "design" && split != "confirmation") {
		t.Fatal("output path and design|confirmation split required")
	}
	seed, fitOffset := originV3DesignSeed, 400
	if split == "confirmation" {
		seed, fitOffset = originV3ConfirmationSeed, 500
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := make(map[string]string)
	for _, name := range []string{
		"docs/experiments/mmm-kalman-origin-v3-protocol.md",
		"internal/observationlearners/kalman_origin_v3_test.go",
		"internal/observationlearners/kalman_early_v2_test.go",
		"internal/observationlearners/kalman_window.go",
		"internal/observationlearners/experiment.go",
		"internal/observation/forecast.go",
	} {
		content, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(content)
		hashes[name] = hex.EncodeToString(hash[:])
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(map[string]any{"kind": "manifest", "split": split,
		"seedBase": seed, "fitOffset": fitOffset, "hashes": hashes,
		"processNoise": originV3Noise, "scenarios": originV3Scenarios}); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range originV3Scenarios {
		for fit := 0; fit < 4; fit++ {
			base, err := Base(Scenarios[scenario], scenario, fit+fitOffset)
			if err != nil {
				t.Fatal(err)
			}
			for stream := 0; stream < 8; stream++ {
				row, err := runOriginV3(base, split, scenario, fit, stream, seed)
				if err != nil {
					t.Fatalf("scenario=%s fit=%d stream=%d: %v", Scenarios[scenario].Name, fit, stream, err)
				}
				if err := encoder.Encode(row); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if err := file.Sync(); err != nil {
		t.Fatal(err)
	}
}
