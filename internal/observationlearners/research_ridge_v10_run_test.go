package observationlearners

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/bits"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type ridgeV10Tick struct {
	Outcome   bool       `json:"y"`
	Ready     bool       `json:"ready"`
	Incumbent [4]float64 `json:"incumbent"`
	Candidate [4]float64 `json:"candidate"`
}

type ridgeV10Publication struct {
	Clock, Audits, LastOrigin int
	Failure                   string `json:",omitempty"`
}

type ridgeV10Record struct {
	Split, Scenario                 string
	Fit, Stream                     int
	Audits, Available, Publications int
	FitNS, CompileNS, ForecastNS    int64
	PublicationsLog                 []ridgeV10Publication
	Ticks                           []ridgeV10Tick
}

type ridgeV10Output struct {
	Mode, JournalSHA256, ProtocolSHA256, SourceSHA256 string
	Records                                           []ridgeV10Record
	WallNS                                            int64
}

func ridgeV10Hash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func ridgeV10Input(mode string, rng *rand.Rand) uint16 {
	if mode == "uniform" {
		return uint16(rng.Intn(512))
	}
	latent := rng.Intn(2) == 1
	var x uint16
	for bit := 0; bit < 9; bit++ {
		value := false
		switch bit {
		case 0, 1, 6, 7:
			value = latent
			if rng.Float64() < .15 {
				value = !value
			}
		default:
			value = rng.Intn(2) == 1
		}
		if value {
			x |= 1 << bit
		}
	}
	return x
}

func ridgeV10Label(x uint16, t int, s Scenario, rng *rand.Rand) bool {
	if s.Name == "null" {
		return rng.Intn(2) == 1
	}
	local := t >= s.Change
	if s.Name == "gradual" {
		probability := math.Max(0, math.Min(1, float64(t-128)/256))
		local = rng.Float64() < probability
	}
	if s.Name == "recurring" {
		local = t >= 0 && (t/128)%2 == 1
	}
	bit := func(i uint) bool { return x&(1<<i) != 0 }
	majority := func(a, b, c bool) bool { return (a && b) || (a && c) || (b && c) }
	y := majority(bit(0), bit(3), bit(6))
	if local {
		y = majority(bit(2), bit(5), bit(8))
		if s.Name == "interaction" {
			y = (bit(0) && bit(1)) || (bit(2) && bit(3))
		}
	}
	if rng.Float64() < s.Noise {
		y = !y
	}
	return y
}

func ridgeV10Replay(record RetainedRecord, mode string) (ridgeV10Record, error) {
	var scenario Scenario
	j := -1
	for i, s := range Scenarios {
		if s.Name == record.Scenario {
			j, scenario = i, s
			break
		}
	}
	if j < 0 || (record.Split != "design" && record.Split != "confirmation") || record.Fit < 0 || record.Fit >= 3 || record.Stream < 0 || record.Stream >= 8 {
		return ridgeV10Record{}, fmt.Errorf("invalid record identity: %s/%s/%d/%d", record.Split, record.Scenario, record.Fit, record.Stream)
	}
	if len(record.Ticks) != 512 || len(record.Views) != 512 {
		return ridgeV10Record{}, fmt.Errorf("incomplete archived record")
	}
	phase := int64(0)
	if record.Split == "confirmation" {
		phase = 1
	}
	rng := rand.New(rand.NewSource(Seed(2026100124+phase, j, record.Fit, record.Stream, 0)))
	out := ridgeV10Record{Split: record.Split, Scenario: record.Scenario, Fit: record.Fit, Stream: record.Stream}
	var inputs [512]uint16
	var delivered [512]bool
	var audits []observation.Sample
	var law *ConditionalForest
	var weights [512]float64
	for i := range weights {
		weights[i] = 1
	}
	for t, tick := range record.Ticks {
		x := ridgeV10Input(mode, rng)
		inputs[t] = x
		if ridgeV10Label(x, t, scenario, rng) != tick.Outcome {
			return ridgeV10Record{}, fmt.Errorf("outcome reconstruction mismatch at %d", t)
		}
		forecast := ridgeV10Tick{Outcome: tick.Outcome, Ready: law != nil}
		for arm := 0; arm < 4; arm++ {
			trace := record.Views[t][arm].Trace
			var mask, value uint16
			if len(trace) > 0 {
				mask, value = trace[len(trace)-1].Observed, trace[len(trace)-1].Values
			}
			if mask >= 512 || value&^mask != 0 || x&mask != value || bits.OnesCount16(mask) > 6 {
				return ridgeV10Record{}, fmt.Errorf("invalid selected view at %d arm %d", t, arm)
			}
			p := tick.Predictions[arm].P
			if p <= 0 || p >= 1 || math.IsNaN(p) {
				return ridgeV10Record{}, fmt.Errorf("invalid incumbent at %d arm %d", t, arm)
			}
			forecast.Incumbent[arm], forecast.Candidate[arm] = p, p
			if law != nil {
				start := time.Now()
				q, err := law.Forecast(mask, value)
				out.ForecastNS += int64(time.Since(start))
				if err != nil || q <= 0 || q >= 1 || math.IsNaN(q) {
					return ridgeV10Record{}, fmt.Errorf("ridge forecast at %d arm %d: %g, %v", t, arm, q, err)
				}
				forecast.Candidate[arm] = q
			}
		}
		out.Ticks = append(out.Ticks, forecast)
		for _, origin := range tick.Delivered {
			if origin < 0 || origin > t || delivered[origin] || record.Ticks[origin].Missing || origin+scenario.Delay != t {
				return ridgeV10Record{}, fmt.Errorf("invalid delivery at %d from %d", t, origin)
			}
			delivered[origin] = true
			out.Available++
			if !record.Ticks[origin].Audit {
				continue
			}
			out.Audits++
			audits = append(audits, observation.Sample{Bits: inputs[origin], Outcome: record.Ticks[origin].Outcome})
			if len(audits) > 64 {
				audits = audits[1:]
			}
			if out.Audits < 32 || out.Audits%16 != 0 {
				continue
			}
			publication := ridgeV10Publication{Clock: t, Audits: out.Audits, LastOrigin: origin}
			start := time.Now()
			model, err := fitRidgeLogistic(audits)
			out.FitNS += int64(time.Since(start))
			if err == nil {
				start = time.Now()
				var compiled *ConditionalForest
				compiled, err = model.conditional(weights)
				out.CompileNS += int64(time.Since(start))
				if err == nil {
					law = compiled
					out.Publications++
				}
			}
			if err != nil {
				publication.Failure = err.Error()
			}
			out.PublicationsLog = append(out.PublicationsLog, publication)
		}
		if out.Audits != tick.Audits {
			return ridgeV10Record{}, fmt.Errorf("audit counter mismatch at %d: %d != %d", t, out.Audits, tick.Audits)
		}
	}
	if out.Available != record.Available || out.Audits != record.Audits {
		return ridgeV10Record{}, fmt.Errorf("terminal delivery mismatch")
	}
	return out, nil
}

func ridgeV10Run(input, mode, output string) error {
	if mode != "uniform" && mode != "latent" {
		return fmt.Errorf("undeclared mode %q", mode)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		return fmt.Errorf("output exists or cannot be checked: %s", output)
	}
	started := time.Now()
	journalSHA, err := ridgeV10Hash(input)
	if err != nil {
		return err
	}
	want := map[string]string{"uniform": "c9a404fbe5c3b0243563e48f229615ae3d3423cb3b17962f8ff1e9df15e5751b", "latent": "3b8ea4d05c6158613d0e169e6e93f1afe4add6efa767676c43746749d86b0066"}
	if journalSHA != want[mode] {
		return fmt.Errorf("journal source hash mismatch for %s", mode)
	}
	protocolSHA, err := ridgeV10Hash("../../docs/experiments/mmm-ridge-v10-component-protocol.md")
	if err != nil {
		return err
	}
	sourceSHA, err := ridgeV10Hash("research_ridge_v10_run_test.go")
	if err != nil {
		return err
	}
	f, err := os.Open(input)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	decoder := json.NewDecoder(gz)
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return fmt.Errorf("invalid journal root: %v", err)
	}
	result := ridgeV10Output{Mode: mode, JournalSHA256: journalSHA, ProtocolSHA256: protocolSHA, SourceSHA256: sourceSHA}
	var seen [2][10][3][8]bool
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return err
		}
		if key != "Records" {
			var discard json.RawMessage
			if err := decoder.Decode(&discard); err != nil {
				return err
			}
			continue
		}
		if token, err := decoder.Token(); err != nil || token != json.Delim('[') {
			return fmt.Errorf("invalid records array: %v", err)
		}
		for decoder.More() {
			var record RetainedRecord
			if err := decoder.Decode(&record); err != nil {
				return err
			}
			replayed, err := ridgeV10Replay(record, mode)
			if err != nil {
				return fmt.Errorf("%s/%s/%d/%d: %w", record.Split, record.Scenario, record.Fit, record.Stream, err)
			}
			phase, scenario := 0, -1
			if record.Split == "confirmation" {
				phase = 1
			}
			for j, s := range Scenarios {
				if s.Name == record.Scenario {
					scenario = j
					break
				}
			}
			if seen[phase][scenario][record.Fit][record.Stream] {
				return fmt.Errorf("duplicate stream identity")
			}
			seen[phase][scenario][record.Fit][record.Stream] = true
			result.Records = append(result.Records, replayed)
		}
		if _, err := decoder.Token(); err != nil {
			return err
		}
	}
	if _, err := decoder.Token(); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("trailing journal content: %v", err)
	}
	if len(result.Records) != 480 {
		return fmt.Errorf("incomplete corpus: %d records", len(result.Records))
	}
	result.WallNS = int64(time.Since(started))
	out, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	compressed := gzip.NewWriter(out)
	writeErr := json.NewEncoder(compressed).Encode(result)
	if err := compressed.Close(); writeErr == nil {
		writeErr = err
	}
	if err := out.Close(); writeErr == nil {
		writeErr = err
	}
	return writeErr
}

func TestResearchRidgeV10Replay(t *testing.T) {
	input, mode, output := os.Getenv("EVENTFRAME_RIDGE_INPUT"), os.Getenv("EVENTFRAME_RIDGE_MODE"), os.Getenv("EVENTFRAME_RIDGE_OUTPUT")
	if input == "" && mode == "" && output == "" {
		t.Skip("opt-in consumed-v10 ridge diagnostic")
	}
	if input == "" || mode == "" || output == "" {
		t.Fatal("input, mode and output must all be set")
	}
	if err := ridgeV10Run(input, mode, output); err != nil {
		t.Fatal(err)
	}
}
