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
	"math/rand"
	"os"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

type measure struct {
	N                int
	ModelBrier       float64
	ViewOracleBrier  float64
	FullOracleBrier  float64
	MeanObservedBits float64
}

func (m *measure) add(model, view, full float64, observed int, outcome bool) {
	y := 0.0
	if outcome {
		y = 1
	}
	m.N++
	m.ModelBrier += (model - y) * (model - y)
	m.ViewOracleBrier += (view - y) * (view - y)
	m.FullOracleBrier += (full - y) * (full - y)
	m.MeanObservedBits += float64(observed)
}

func (m measure) normalized() measure {
	if m.N == 0 {
		return m
	}
	d := float64(m.N)
	m.ModelBrier /= d
	m.ViewOracleBrier /= d
	m.FullOracleBrier /= d
	m.MeanObservedBits /= d
	return m
}

type rowResult struct {
	Scenario string
	Fit      int
	Stream   int
	Post     [4]measure
	Early    [4]measure
}

type cellResult struct {
	Scenario string
	Window   string
	Arms     [4]measure
}

type oracleKey struct {
	Mask  uint16
	Value uint16
	Local bool
}

type oracle struct {
	mass  [512]float64
	cache map[oracleKey]float64
}

func fileHash(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func bit(x uint16, position uint) bool { return x&(1<<position) != 0 }

func majority(a, b, c bool) bool {
	return (a && b) || (a && c) || (b && c)
}

func truthProbability(x uint16, local bool) float64 {
	y := majority(bit(x, 0), bit(x, 3), bit(x, 6))
	if local {
		y = majority(bit(x, 2), bit(x, 5), bit(x, 8))
	}
	if y {
		return 0.95
	}
	return 0.05
}

func drawInput(rng *rand.Rand, mode string) uint16 {
	if mode == "uniform" {
		return uint16(rng.Intn(512))
	}
	latent := rng.Intn(2) == 1
	var result uint16
	for position := 0; position < 9; position++ {
		value := false
		switch position {
		case 0, 1, 6, 7:
			value = latent
			if rng.Float64() < 0.15 {
				value = !value
			}
		default:
			value = rng.Intn(2) == 1
		}
		if value {
			result |= 1 << position
		}
	}
	return result
}

func inputMass(x uint16, mode string) float64 {
	if mode == "uniform" {
		return 1.0 / 512
	}
	sum := 0.0
	for _, latent := range []bool{false, true} {
		product := 1.0
		for _, position := range []uint{0, 1, 6, 7} {
			if bit(x, position) == latent {
				product *= 0.85
			} else {
				product *= 0.15
			}
		}
		sum += 0.5 * product
	}
	return sum / 32
}

func newOracle(mode string) (*oracle, error) {
	if mode != "uniform" && mode != "latent" {
		return nil, fmt.Errorf("undeclared input mode: %s", mode)
	}
	o := &oracle{cache: map[oracleKey]float64{}}
	sum := 0.0
	for x := 0; x < 512; x++ {
		o.mass[x] = inputMass(uint16(x), mode)
		if o.mass[x] <= 0 {
			return nil, fmt.Errorf("zero input mass at %d", x)
		}
		sum += o.mass[x]
	}
	if math.Abs(sum-1) > 1e-12 {
		return nil, fmt.Errorf("input mass sums to %.15g", sum)
	}
	return o, nil
}

func (o *oracle) selectedView(mask, value uint16, local bool) (float64, error) {
	if mask > 511 || value&^mask != 0 {
		return 0, fmt.Errorf("invalid observed mask/value %d/%d", mask, value)
	}
	key := oracleKey{mask, value, local}
	if cached, ok := o.cache[key]; ok {
		return cached, nil
	}
	numerator, denominator := 0.0, 0.0
	for x := 0; x < 512; x++ {
		if uint16(x)&mask != value {
			continue
		}
		weight := o.mass[x]
		denominator += weight
		numerator += weight * truthProbability(uint16(x), local)
	}
	if denominator <= 0 {
		return 0, fmt.Errorf("empty view support %d/%d", mask, value)
	}
	p := numerator / denominator
	o.cache[key] = p
	return p, nil
}

func scenarioIndex(name string) (int, observationlearners.Scenario, error) {
	for i, scenario := range observationlearners.Scenarios {
		if scenario.Name == name {
			return i, scenario, nil
		}
	}
	return 0, observationlearners.Scenario{}, fmt.Errorf("unknown scenario %s", name)
}

func inspect(record observationlearners.RetainedRecord, mode string, o *oracle) (rowResult, error) {
	index, scenario, err := scenarioIndex(record.Scenario)
	if err != nil {
		return rowResult{}, err
	}
	if record.Split != "confirmation" || (record.Scenario != "shift128" && record.Scenario != "shift256") {
		return rowResult{}, fmt.Errorf("out-of-scope record %s/%s", record.Split, record.Scenario)
	}
	if len(record.Ticks) != 512 || len(record.Views) != 512 {
		return rowResult{}, fmt.Errorf("incomplete record")
	}
	seed := observationlearners.Seed(2026100125, index, record.Fit, record.Stream, 0)
	rng := rand.New(rand.NewSource(seed))
	result := rowResult{Scenario: record.Scenario, Fit: record.Fit, Stream: record.Stream}
	for clock, tick := range record.Ticks {
		x := drawInput(rng, mode)
		local := clock >= scenario.Change
		expected := truthProbability(x, local) == 0.95
		if rng.Float64() < scenario.Noise {
			expected = !expected
		}
		if expected != tick.Outcome {
			return rowResult{}, fmt.Errorf("outcome mismatch %s fit%d stream%d clock%d", record.Scenario, record.Fit, record.Stream, clock)
		}
		var masks, values [4]uint16
		for arm := 0; arm < 4; arm++ {
			view := record.Views[clock][arm]
			if len(view.Trace) > 0 {
				last := view.Trace[len(view.Trace)-1]
				masks[arm], values[arm] = last.Observed, last.Values
			}
			if masks[arm] > 511 || values[arm]&^masks[arm] != 0 || x&masks[arm] != values[arm] {
				return rowResult{}, fmt.Errorf("observed value mismatch %s fit%d stream%d clock%d arm%d", record.Scenario, record.Fit, record.Stream, clock, arm)
			}
		}
		if !local {
			continue
		}
		full := truthProbability(x, true)
		for arm := 0; arm < 4; arm++ {
			conditional, err := o.selectedView(masks[arm], values[arm], true)
			if err != nil {
				return rowResult{}, err
			}
			observed := bits.OnesCount16(masks[arm])
			model := tick.Predictions[arm].P
			result.Post[arm].add(model, conditional, full, observed, tick.Outcome)
			if clock < scenario.Change+64 {
				result.Early[arm].add(model, conditional, full, observed, tick.Outcome)
			}
		}
	}
	return result, nil
}

func addMeasure(dst *measure, src measure) {
	dst.N += src.N
	dst.ModelBrier += src.ModelBrier
	dst.ViewOracleBrier += src.ViewOracleBrier
	dst.FullOracleBrier += src.FullOracleBrier
	dst.MeanObservedBits += src.MeanObservedBits
}

func run(input, mode, output string) error {
	started := time.Now()
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		return fmt.Errorf("output exists or inaccessible: %s", output)
	}
	inputSHA256, err := fileHash(input)
	if err != nil {
		return err
	}
	protocolSHA256, err := fileHash("docs/experiments/mmm-retained-truth-v10-oracle-protocol.md")
	if err != nil {
		return err
	}
	sourceSHA256, err := fileHash("cmd/research-retained-truth-oracle/main.go")
	if err != nil {
		return err
	}
	o, err := newOracle(mode)
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
	var rows []rowResult
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
			row, err := inspect(record, mode, o)
			if err != nil {
				return err
			}
			rows = append(rows, row)
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
	if len(rows) != 48 {
		return fmt.Errorf("expected 48 confirmation shift records, got %d", len(rows))
	}
	cells := make([]cellResult, 0, 4)
	for _, scenario := range []string{"shift128", "shift256"} {
		for _, window := range []string{"post", "early64"} {
			cell := cellResult{Scenario: scenario, Window: window}
			count := 0
			for _, row := range rows {
				if row.Scenario != scenario {
					continue
				}
				count++
				for arm := 0; arm < 4; arm++ {
					if window == "post" {
						addMeasure(&cell.Arms[arm], row.Post[arm])
					} else {
						addMeasure(&cell.Arms[arm], row.Early[arm])
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
		"records":          rows,
		"cells":            cells,
		"cachedViews":      len(o.cache),
		"wallMilliseconds": float64(time.Since(started)) / float64(time.Millisecond),
		"limits":           "Post-hoc consumed-data oracle diagnostic only; no predictor, gate, or service changes.",
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
	fmt.Printf("mode=%s records=%d cachedViews=%d wall=%s\n", mode, len(rows), len(o.cache), time.Since(started))
	for _, cell := range cells {
		fmt.Printf("%s/%s fixed=%.5f view=%.5f full=%.5f bits=%.2f\n", cell.Scenario, cell.Window,
			cell.Arms[0].ModelBrier, cell.Arms[0].ViewOracleBrier, cell.Arms[0].FullOracleBrier, cell.Arms[0].MeanObservedBits)
	}
	return nil
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: research-retained-truth-oracle FULL.json.gz MODE OUTPUT.json")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2], os.Args[3]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
