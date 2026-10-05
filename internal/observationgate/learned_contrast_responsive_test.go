package observationgate

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

var responsiveCases = [...]transferCase{
	{Name: "early160", Change: 160, Noise: .05, Missing: .20},
	{Name: "late360", Change: 360, Noise: .05, Missing: .20},
	{Name: "late400", Change: 400, Noise: .05, Missing: .20},
	{Name: "noisy256", Change: 256, Noise: .20, Missing: .20},
	{Name: "delayed256", Change: 256, Noise: .05, Missing: .25, Delay: 32},
	{Name: "skew256", Change: 256, Noise: .05, Missing: .20, Skew: true},
	{Name: "stable20", Change: 512, Noise: .20, Missing: .20, Stable: true},
	{Name: "majority_ood", Change: 256, Noise: .05, Missing: .20, OOD: true},
}

type responsiveMonitor struct {
	Log     [8][10]float64
	AlertAt int
	Updates int
}

func newResponsiveMonitor() responsiveMonitor { return responsiveMonitor{AlertAt: -1} }

func (m *responsiveMonitor) observe(origin int, x uint16, y bool, baseP float64, clock int) error {
	if origin < 0 || origin > clock || clock >= learnedSteps ||
		baseP <= 0 || baseP >= 1 || math.IsNaN(baseP) {
		return errors.New("invalid responsive monitor input")
	}
	var logRatio [10]float64
	for h := 1; h <= 10; h++ {
		p := learnedRule(h, x, baseP)
		if y {
			logRatio[h-1] = math.Log(p) - math.Log(baseP)
		} else {
			logRatio[h-1] = math.Log1p(-p) - math.Log1p(-baseP)
		}
	}
	maximum := math.Inf(-1)
	for j := range m.Log {
		if origin >= j*64 {
			for h := range m.Log[j] {
				m.Log[j][h] += logRatio[h]
			}
		}
		for _, v := range m.Log[j] {
			maximum = math.Max(maximum, v)
		}
	}
	total := 0.
	for _, row := range m.Log {
		for _, v := range row {
			total += math.Exp(v - maximum)
		}
	}
	if total <= 0 || math.IsNaN(total) || math.IsInf(total, 0) {
		return errors.New("invalid responsive monitor wealth")
	}
	m.Updates++
	if m.AlertAt < 0 && maximum+math.Log(total/80) >= math.Log(100) {
		m.AlertAt = clock
	}
	return nil
}

func responsiveNominate(policy string, x uint16, base, meanUncertainty float64,
	state *learnedState, remaining, clocks int, rate float64, rng *rand.Rand) bool {
	if remaining <= 0 {
		return false
	}
	if remaining >= clocks {
		return true
	}
	switch policy {
	case "uncertainty":
		rate *= 4 * base * (1 - base) / meanUncertainty
	case "learned_disagreement":
		alt := state.bestAlternative()
		if (learnedRule(alt, x, base) >= .5) == (base >= .5) {
			return false
		}
		rate *= 2
	}
	return rng.Float64() < math.Min(1, rate)
}

type responsiveRaw struct {
	Data           learnedArm
	AlertAt        int
	MonitorUpdates int
}

func responsiveRunArm(tape [learnedSteps]learnedTapeTick, base *[512]float64,
	meanUncertainty float64, policy string, rng *rand.Rand, correctAlternative, delay int) (responsiveRaw, error) {
	r := responsiveRaw{Data: learnedArm{Policy: policy}, AlertAt: -1}
	state, monitor := newLearnedState(), newResponsiveMonitor()
	var due [learnedSteps][]int
	for clock, event := range tape {
		pb := base[event.X]
		p := state.forecast(event.X, pb)
		if p <= 0 || p >= 1 || math.IsNaN(p) {
			return r, errors.New("invalid responsive working forecast")
		}
		alt := state.bestAlternative()
		tick := learnedArmTick{P: p, Alternate: alt}
		loss := (p - boolFloat(event.Y)) * (p - boolFloat(event.Y))
		r.Data.FullBrier += loss / learnedSteps
		if clock >= 256 {
			r.Data.PostBrier += loss / 256
			if clock < 320 {
				r.Data.EarlyBrier += loss / 64
			}
		}
		rate := .20
		if monitor.AlertAt >= 0 && clock > monitor.AlertAt && clock <= monitor.AlertAt+64 {
			rate = .60
		}
		if responsiveNominate(policy, event.X, pb, meanUncertainty, &state,
			learnedBudget-r.Data.Nominated, learnedSteps-clock, rate, rng) {
			tick.Selected = true
			r.Data.Nominated++
			if clock >= 256 {
				r.Data.NominatedPost++
				if alt == correctAlternative {
					r.Data.CorrectPost++
				}
			}
			if !event.Missing {
				arrival := clock + delay
				if arrival < learnedSteps {
					due[arrival] = append(due[arrival], clock)
				} else {
					r.Data.Pending++
				}
			}
		}
		for _, origin := range due[clock] {
			if origin > clock || tape[origin].Missing {
				return r, errors.New("future or missing responsive label")
			}
			r.Data.Arrived++
			tick.Delivered = append(tick.Delivered, origin)
			observed := tape[origin]
			if err := state.observe(observed.X, observed.Y, base); err != nil {
				return r, err
			}
			if err := monitor.observe(origin, observed.X, observed.Y, base[observed.X], clock); err != nil {
				return r, err
			}
		}
		r.Data.Ticks = append(r.Data.Ticks, tick)
	}
	if r.Data.Nominated != learnedBudget || r.Data.Arrived+r.Data.Pending > learnedBudget ||
		monitor.Updates != r.Data.Arrived {
		return r, errors.New("invalid responsive accounting")
	}
	r.AlertAt, r.MonitorUpdates = monitor.AlertAt, monitor.Updates
	return r, nil
}

type responsiveTapeTick struct {
	X       uint16
	Y       bool
	Missing bool
	PTrue   float64
	BaseP   float64
}

type responsiveArm struct {
	Data           transferArm
	AlertAt        int
	MonitorUpdates int
}

type responsiveRow struct {
	Kind, Split, Scenario string
	Fit, Stream           int
	Tape                  [learnedSteps]responsiveTapeTick
	Arms                  [6]responsiveArm
}

func TestLearnedContrastResponsive(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_RESPONSIVE_OUT"), os.Getenv("EVENTFRAME_RESPONSIVE_SPLIT")
	if path == "" && split == "" {
		t.Skip("opt-in research study")
	}
	if path == "" || (split != "design" && split != "confirmation") {
		t.Fatal("output path and design|confirmation split required")
	}
	seed, fitOffset := int64(2026102701), 1800
	if split == "confirmation" {
		seed, fitOffset = 2026102702, 1900
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := make(map[string]string)
	for _, name := range []string{
		"docs/experiments/mmm-observation-responsive-v1-protocol.md",
		"internal/observationgate/learned_contrast_responsive_test.go",
		"internal/observationgate/learned_contrast_transfer_test.go",
		"internal/observationgate/learned_contrast_v2_test.go",
		"internal/observationlearners/experiment.go", "internal/observation/forecast.go",
	} {
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
	gz := gzip.NewWriter(f)
	defer gz.Close()
	enc := json.NewEncoder(gz)
	if err := enc.Encode(map[string]any{"kind": "manifest", "split": split,
		"seedBase": seed, "fitOffset": fitOffset, "trialsPerCase": 256,
		"budget": learnedBudget, "monitorStarts": 8, "alternatives": 10,
		"threshold": 100, "baseRate": .20, "burstRate": .60, "burstClocks": 64,
		"stateBytes": unsafe.Sizeof(learnedState{}) + unsafe.Sizeof(responsiveMonitor{}),
		"hashes":     hashes}); err != nil {
		t.Fatal(err)
	}
	for caseID, c := range responsiveCases {
		for fit := 0; fit < 16; fit++ {
			baseScenario := observationlearners.Scenarios[0]
			if c.Noise == .20 {
				baseScenario = observationlearners.Scenarios[1]
			}
			base, e := observationlearners.Base(baseScenario, caseID, fit+fitOffset)
			if e != nil {
				t.Fatal(e)
			}
			var probability [512]float64
			meanUncertainty := 0.
			for x := range probability {
				p, e := base.ForecastObserved(511, uint16(x))
				if e != nil || p <= 0 || p >= 1 {
					t.Fatalf("base forecast %s x=%d: %v", c.Name, x, e)
				}
				probability[x] = p
				meanUncertainty += 4 * p * (1 - p) / 512
			}
			for stream := 0; stream < 16; stream++ {
				tape := transferTape(seed, c, caseID, fit, stream)
				row := responsiveRow{Kind: "trial", Split: split, Scenario: c.Name,
					Fit: fit, Stream: stream}
				var oldTape [learnedSteps]learnedTapeTick
				for clock, event := range tape {
					oldTape[clock] = learnedTapeTick{X: event.X, Y: event.Y, Missing: event.Missing}
					row.Tape[clock] = responsiveTapeTick{X: event.X, Y: event.Y,
						Missing: event.Missing, PTrue: event.PTrue, BaseP: probability[event.X]}
				}
				correct := -1
				if !c.Stable && !c.OOD {
					correct = 3
				}
				for policy, name := range learnedPolicies {
					selectorSeed := learnedSeed(seed, caseID, fit, stream, 10+policy)
					oldRNG := rand.New(rand.NewSource(selectorSeed))
					old, e := learnedRunArm(oldTape, &probability, meanUncertainty,
						name, oldRNG, correct, c.Delay)
					if e != nil {
						t.Fatalf("old %s/%d/%d/%s: %v", c.Name, fit, stream, name, e)
					}
					row.Arms[policy] = responsiveArm{Data: transferMeasure(old, tape, c), AlertAt: -1}
					newRNG := rand.New(rand.NewSource(selectorSeed))
					current, e := responsiveRunArm(oldTape, &probability, meanUncertainty,
						name, newRNG, correct, c.Delay)
					if e != nil {
						t.Fatalf("responsive %s/%d/%d/%s: %v", c.Name, fit, stream, name, e)
					}
					row.Arms[policy+3] = responsiveArm{Data: transferMeasure(current.Data, tape, c),
						AlertAt: current.AlertAt, MonitorUpdates: current.MonitorUpdates}
				}
				if err := enc.Encode(row); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}

func TestResponsiveMonitorOriginAndArrival(t *testing.T) {
	m := newResponsiveMonitor()
	if err := m.observe(1, 4, true, .05, 0); err == nil {
		t.Fatal("future origin accepted")
	}
	if m.Updates != 0 || m.AlertAt != -1 {
		t.Fatal("rejected input changed monitor")
	}
	for clock := 0; clock < 8 && m.AlertAt < 0; clock++ {
		if err := m.observe(clock, 4, true, .05, clock); err != nil {
			t.Fatal(err)
		}
	}
	if m.AlertAt < 0 || m.AlertAt > 7 {
		t.Fatal("contradictory arrived evidence did not alert")
	}
}
