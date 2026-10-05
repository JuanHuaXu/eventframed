package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/bits"
	"os"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

type scores struct {
	N            int
	Final        float64
	Base         float64
	Challenger   float64
	Long         float64
	Neutral      float64
	Short        float64
	Tree         float64
	ObservedBits float64
}

func squared(probability float64, outcome bool) float64 {
	y := 0.0
	if outcome {
		y = 1
	}
	difference := probability - y
	return difference * difference
}

func (s *scores) add(final float64, experts [4]float64, inner [4]float64, observed int, outcome bool) error {
	for _, p := range []float64{final, experts[0], experts[1], experts[2], experts[3], inner[0], inner[1]} {
		if math.IsNaN(p) || p < 0 || p > 1 {
			return fmt.Errorf("invalid archived probability %g", p)
		}
	}
	s.N++
	s.Final += squared(final, outcome)
	s.Base += squared(experts[0], outcome)
	s.Challenger += squared(experts[1], outcome)
	s.Long += squared(experts[2], outcome)
	s.Neutral += squared(experts[3], outcome)
	s.Short += squared(inner[0], outcome)
	s.Tree += squared(inner[1], outcome)
	s.ObservedBits += float64(observed)
	return nil
}

func (s *scores) accumulate(other scores) {
	s.N += other.N
	s.Final += other.Final
	s.Base += other.Base
	s.Challenger += other.Challenger
	s.Long += other.Long
	s.Neutral += other.Neutral
	s.Short += other.Short
	s.Tree += other.Tree
	s.ObservedBits += other.ObservedBits
}

func (s scores) normalized() scores {
	if s.N == 0 {
		return s
	}
	d := float64(s.N)
	s.Final /= d
	s.Base /= d
	s.Challenger /= d
	s.Long /= d
	s.Neutral /= d
	s.Short /= d
	s.Tree /= d
	s.ObservedBits /= d
	return s
}

type recordScores struct {
	Scenario string
	Fit      int
	Stream   int
	Post     [4]scores
	Early    [4]scores
}

type cellScores struct {
	Scenario string
	Window   string
	Arms     [4]scores
}

func sourceHash(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func inspect(record observationlearners.RetainedRecord) (recordScores, error) {
	if record.Split != "confirmation" || (record.Scenario != "shift128" && record.Scenario != "shift256") {
		return recordScores{}, fmt.Errorf("record outside frozen scope")
	}
	if len(record.Ticks) != 512 || len(record.Views) != 512 || len(record.Inner) != 512 {
		return recordScores{}, fmt.Errorf("incomplete archived trace")
	}
	change := 128
	if record.Scenario == "shift256" {
		change = 256
	}
	out := recordScores{Scenario: record.Scenario, Fit: record.Fit, Stream: record.Stream}
	for clock := change; clock < 512; clock++ {
		tick := record.Ticks[clock]
		for arm := 0; arm < 4; arm++ {
			forecast := tick.Predictions[arm]
			inner := record.Inner[clock][arm]
			switch arm {
			case 0:
				if math.Abs(forecast.Experts[1]-inner[0]) > 1e-12 {
					return recordScores{}, fmt.Errorf("short expert mapping at clock %d", clock)
				}
			case 1:
				if math.Abs(forecast.Experts[1]-inner[1]) > 1e-12 {
					return recordScores{}, fmt.Errorf("tree expert mapping at clock %d", clock)
				}
			case 3:
				if math.Abs(forecast.Experts[1]-(inner[0]+inner[1])/2) > 1e-12 {
					return recordScores{}, fmt.Errorf("static blend mapping at clock %d", clock)
				}
			}
			mask := uint16(0)
			trace := record.Views[clock][arm].Trace
			if len(trace) > 0 {
				mask = trace[len(trace)-1].Observed
			}
			observed := bits.OnesCount16(mask)
			if observed > 6 {
				return recordScores{}, fmt.Errorf("observation budget exceeded at clock %d", clock)
			}
			if err := out.Post[arm].add(forecast.P, forecast.Experts, inner, observed, tick.Outcome); err != nil {
				return recordScores{}, err
			}
			if clock < change+64 {
				if err := out.Early[arm].add(forecast.P, forecast.Experts, inner, observed, tick.Outcome); err != nil {
					return recordScores{}, err
				}
			}
		}
	}
	for arm := 0; arm < 4; arm++ {
		if math.Abs(out.Post[arm].normalized().Final-record.Post[arm].Brier) > 1e-12 {
			return recordScores{}, fmt.Errorf("archived final Brier mismatch: %s fit%d stream%d arm%d", record.Scenario, record.Fit, record.Stream, arm)
		}
	}
	return out, nil
}

func run(input, mode, output string) error {
	if mode != "uniform" && mode != "latent" {
		return fmt.Errorf("undeclared input mode %s", mode)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		return fmt.Errorf("output exists or inaccessible: %s", output)
	}
	started := time.Now()
	inputSHA256, err := sourceHash(input)
	if err != nil {
		return err
	}
	protocolSHA256, err := sourceHash("docs/experiments/mmm-retained-truth-v10-experts-protocol.md")
	if err != nil {
		return err
	}
	sourceSHA256, err := sourceHash("cmd/research-retained-experts/main.go")
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
	var records []recordScores
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
			var record observationlearners.RetainedRecord
			if err := decoder.Decode(&record); err != nil {
				return err
			}
			if record.Split != "confirmation" || (record.Scenario != "shift128" && record.Scenario != "shift256") {
				continue
			}
			result, err := inspect(record)
			if err != nil {
				return err
			}
			records = append(records, result)
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
	if len(records) != 48 {
		return fmt.Errorf("expected 48 scoped records, got %d", len(records))
	}
	var cells []cellScores
	for _, scenario := range []string{"shift128", "shift256"} {
		for _, window := range []string{"post", "early64"} {
			cell := cellScores{Scenario: scenario, Window: window}
			count := 0
			for _, record := range records {
				if record.Scenario != scenario {
					continue
				}
				count++
				for arm := range cell.Arms {
					if window == "post" {
						cell.Arms[arm].accumulate(record.Post[arm])
					} else {
						cell.Arms[arm].accumulate(record.Early[arm])
					}
				}
			}
			if count != 24 {
				return fmt.Errorf("expected 24 records for %s", scenario)
			}
			for arm := range cell.Arms {
				cell.Arms[arm] = cell.Arms[arm].normalized()
			}
			cells = append(cells, cell)
		}
	}
	result := map[string]any{
		"mode":             mode,
		"inputSHA256":      inputSHA256,
		"protocolSHA256":   protocolSHA256,
		"sourceSHA256":     sourceSHA256,
		"records":          records,
		"cells":            cells,
		"wallMilliseconds": float64(time.Since(started)) / float64(time.Millisecond),
		"limits":           "Consumed-data expert diagnosis only; no expert selection, new fit or service claim.",
	}
	out, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(result)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	fmt.Printf("mode=%s records=%d wall=%s\n", mode, len(records), time.Since(started))
	for _, cell := range cells {
		fmt.Printf("%s/%s fixed=%.5f short=%.5f tree=%.5f replacement=%.5f\n",
			cell.Scenario, cell.Window, cell.Arms[0].Final,
			cell.Arms[0].Short, cell.Arms[1].Tree, cell.Arms[1].Final)
	}
	return nil
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: research-retained-experts FULL.json.gz MODE OUTPUT.json")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2], os.Args[3]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
