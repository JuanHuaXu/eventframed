package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

const contrastSteps = 512
const contrastTrials = 1000
const contrastBudget = 128

var contrastCases = [5]string{"stable", "member_shift", "common_shift", "recurring", "null"}
var contrastPolicies = [3]string{"random", "uncertainty", "contrast"}
var contrastStarts = [8]int{0, 64, 128, 192, 256, 320, 384, 448}

func contrastClass(x uint16) uint16 {
	parity := ((x >> 6) ^ (x >> 7) ^ (x >> 8)) & 1
	return (parity << 1) | ((x >> 2) & 1)
}

func contrastEligible(x uint16) bool {
	c := contrastClass(x)
	return ((c >> 1) & 1) != (c & 1)
}

type contrastTick struct {
	x, rx                   uint16
	classMatched, eligible  bool
	y, ry                   bool
	missingLive, missingRef bool
	delayLive, delayRef     int
}

func contrastTape(seed int64, scenarioIndex, trial int) [contrastSteps]contrastTick {
	var rng [7]*rand.Rand
	for role := range rng {
		rng[role] = rand.New(rand.NewSource(observationpreserved.Seed(seed, scenarioIndex, trial, role)))
	}
	var tape [contrastSteps]contrastTick
	scenario := contrastCases[scenarioIndex]
	for clock := range tape {
		t := &tape[clock]
		t.x = uint16(rng[0].Intn(512))
		class := contrastClass(t.x)
		t.eligible = contrastEligible(t.x)
		var menu [4]uint16
		for j := range menu {
			menu[j] = uint16(rng[1].Intn(512))
		}
		t.rx = menu[0]
		for _, candidate := range menu {
			if contrastClass(candidate) == class {
				t.rx = candidate
				t.classMatched = true
				break
			}
		}
		local := (scenario == "member_shift" || scenario == "common_shift") && clock >= 256
		if scenario == "recurring" {
			local = (clock/128)%2 == 1
		}
		t.y = integrationLabel(t.x, local, scenario == "null", rng[2])
		t.ry = integrationLabel(t.rx, scenario == "common_shift" && clock >= 256, scenario == "null", rng[3])
		t.missingLive = rng[4].Float64() < .20
		t.missingRef = rng[5].Float64() < .20
		t.delayLive = rng[6].Intn(32)
		t.delayRef = rng[6].Intn(32)
	}
	return tape
}

type contrastGate struct {
	wealth        [8][2]float64
	flagAt        int
	factorUpdates int
}

func newContrastGate() contrastGate {
	g := contrastGate{flagAt: -1}
	for j := range g.wealth {
		g.wealth[j] = [2]float64{1, 1}
	}
	return g
}

func (g *contrastGate) observe(origin int, z int) error {
	if origin < 0 || origin >= contrastSteps || z < -1 || z > 1 {
		return errors.New("invalid paired contrast")
	}
	for j, start := range contrastStarts {
		if origin < start {
			continue
		}
		for k, sign := range [2]int{-1, 1} {
			factor := 1 + .8*(float64(sign*z)-.1)
			if factor < .12-1e-12 || factor > 1.72+1e-12 {
				return errors.New("unsafe bet")
			}
			g.wealth[j][k] *= factor
			g.factorUpdates++
		}
	}
	return nil
}

func (g *contrastGate) check(clock int) {
	if g.flagAt >= 0 {
		return
	}
	sum := 0.
	for _, pair := range g.wealth {
		sum += pair[0] + pair[1]
	}
	if sum/16 >= 50 {
		g.flagAt = clock
	}
}

type contrastLabel struct {
	origin int
	y      bool
}

type contrastPair struct {
	origin int
	z      int
}

type contrastArm struct {
	Policy                           string
	FullBrier, PostBrier, EarlyBrier float64
	FlagAt                           int
	Nominated, ClassMatched          int
	DeliveredPairs, ArrivedLabels    int
	LiveReads, RefReads              int
	RequestedLabels, FactorUpdates   int
}

type contrastRow struct {
	Kind, Split, Scenario string
	Trial                 int
	Arms                  [3]contrastArm
}

func contrastNominate(policy string, x uint16, pBase, meanU float64, remaining, clocks int, rng *rand.Rand) bool {
	if remaining == 0 {
		return false
	}
	if remaining == clocks {
		return true
	}
	probability := float64(remaining) / float64(clocks)
	switch policy {
	case "uncertainty":
		probability *= 4 * pBase * (1 - pBase) / meanU
	case "contrast":
		if !contrastEligible(x) {
			return false
		}
		probability *= 2
	}
	return rng.Float64() < math.Min(1, probability)
}

func contrastBrier(p float64, y bool) float64 {
	target := 0.
	if y {
		target = 1
	}
	return (p - target) * (p - target)
}

func contrastRunArm(tape [contrastSteps]contrastTick, pBase [contrastSteps]float64, meanU float64,
	policy string, selector *rand.Rand) (contrastArm, error) {
	r := contrastArm{Policy: policy, FlagAt: -1, LiveReads: contrastSteps}
	var labels [contrastSteps][]contrastLabel
	var pairDue [contrastSteps][]contrastPair
	gate := newContrastGate()
	logOdds := 0.
	for clock, tick := range tape {
		localP := .05
		if tick.x&(1<<2) != 0 {
			localP = .95
		}
		weight := 1 / (1 + math.Exp(-logOdds))
		p := (1-weight)*pBase[clock] + weight*localP
		loss := contrastBrier(p, tick.y)
		r.FullBrier += loss
		if clock >= 256 {
			r.PostBrier += loss
			if clock < 320 {
				r.EarlyBrier += loss
			}
		}
		if contrastNominate(policy, tick.x, pBase[clock], meanU,
			contrastBudget-r.Nominated, contrastSteps-clock, selector) {
			r.Nominated++
			r.RefReads += 4
			r.RequestedLabels += 2
			if tick.classMatched {
				r.ClassMatched++
			}
			liveDue, refDue := -1, -1
			if !tick.missingLive {
				liveDue = clock + tick.delayLive
				if liveDue < contrastSteps {
					labels[liveDue] = append(labels[liveDue], contrastLabel{origin: clock, y: tick.y})
				}
			}
			if !tick.missingRef {
				refDue = clock + tick.delayRef
				if refDue < contrastSteps {
					r.ArrivedLabels++
				}
			}
			if tick.classMatched && tick.eligible && liveDue >= 0 && refDue >= 0 &&
				liveDue < contrastSteps && refDue < contrastSteps {
				z := 0
				if tick.y {
					z++
				}
				if tick.ry {
					z--
				}
				if tick.x&(1<<2) == 0 {
					z = -z
				}
				due := max(liveDue, refDue)
				pairDue[due] = append(pairDue[due], contrastPair{origin: clock, z: z})
			}
		}
		// The current label, including a zero-delay label, arrives after this forecast.
		for _, label := range labels[clock] {
			r.ArrivedLabels++
			x := tape[label.origin].x
			pb := pBase[label.origin]
			pl := .05
			if x&(1<<2) != 0 {
				pl = .95
			}
			likelihoodRatio := pl / pb
			if !label.y {
				likelihoodRatio = (1 - pl) / (1 - pb)
			}
			logOdds = min(8, max(-8, .95*logOdds+math.Log(likelihoodRatio)))
		}
		for _, pair := range pairDue[clock] {
			if pair.origin > clock {
				return r, errors.New("future paired outcome")
			}
			if err := gate.observe(pair.origin, pair.z); err != nil {
				return r, err
			}
			r.DeliveredPairs++
		}
		gate.check(clock)
	}
	if r.Nominated != contrastBudget || r.RequestedLabels != 2*contrastBudget ||
		r.RefReads != 4*contrastBudget || r.ArrivedLabels < 2*r.DeliveredPairs ||
		r.DeliveredPairs > r.ClassMatched {
		return r, errors.New("acquisition accounting invariant")
	}
	r.FlagAt = gate.flagAt
	r.FactorUpdates = gate.factorUpdates
	r.FullBrier /= contrastSteps
	r.PostBrier /= contrastSteps - 256
	r.EarlyBrier /= 64
	return r, nil
}

type contrastCell struct {
	Scenario, Policy                                 string
	Trials                                           int
	Flags, Prechange                                 int
	MedianFlag, P95Flag                              int
	MeanFull, MeanPost, MeanEarly                    float64
	MeanMatched, MeanPairs, MeanArrived, MeanFactors float64
}

type contrastComparison struct {
	Scenario, Control                                                  string
	PostGain, PostLower, EarlyGain, EarlyLower, StableHarm, ArrivedGap float64
	Pass                                                               bool
}

func contrastQuantile(values []int, p float64) int {
	if len(values) == 0 {
		return -1
	}
	sort.Ints(values)
	return values[int(math.Ceil(float64(len(values))*p))-1]
}

func contrastGain(rows []contrastRow, scenario, control string, field func(contrastArm) float64) (float64, float64) {
	var sum, sumSquare float64
	for _, row := range rows {
		if row.Scenario != scenario {
			continue
		}
		var candidate, baseline contrastArm
		for _, arm := range row.Arms {
			if arm.Policy == "contrast" {
				candidate = arm
			}
			if arm.Policy == control {
				baseline = arm
			}
		}
		d := field(baseline) - field(candidate)
		sum += d
		sumSquare += d * d
	}
	mean := sum / contrastTrials
	variance := math.Max(0, (sumSquare-float64(contrastTrials)*mean*mean)/float64(contrastTrials-1))
	return mean, mean - 3.5*math.Sqrt(variance/contrastTrials)
}

func contrastSummarize(rows []contrastRow) ([]contrastCell, []contrastComparison, bool) {
	var cells []contrastCell
	pass := true
	for _, scenario := range contrastCases {
		for _, policy := range contrastPolicies {
			c := contrastCell{Scenario: scenario, Policy: policy, MedianFlag: -1, P95Flag: -1}
			var flags []int
			for _, row := range rows {
				if row.Scenario != scenario {
					continue
				}
				for _, arm := range row.Arms {
					if arm.Policy != policy {
						continue
					}
					c.Trials++
					c.MeanFull += arm.FullBrier
					c.MeanPost += arm.PostBrier
					c.MeanEarly += arm.EarlyBrier
					c.MeanMatched += float64(arm.ClassMatched)
					c.MeanPairs += float64(arm.DeliveredPairs)
					c.MeanArrived += float64(arm.ArrivedLabels)
					c.MeanFactors += float64(arm.FactorUpdates)
					if arm.FlagAt >= 0 {
						c.Flags++
						flags = append(flags, arm.FlagAt)
						if scenario == "member_shift" && arm.FlagAt < 256 {
							c.Prechange++
						}
					}
				}
			}
			if c.Trials != contrastTrials {
				panic("missing contrast trial")
			}
			c.MedianFlag, c.P95Flag = contrastQuantile(flags, .5), contrastQuantile(flags, .95)
			c.MeanFull /= contrastTrials
			c.MeanPost /= contrastTrials
			c.MeanEarly /= contrastTrials
			c.MeanMatched /= contrastTrials
			c.MeanPairs /= contrastTrials
			c.MeanArrived /= contrastTrials
			c.MeanFactors /= contrastTrials
			if policy == "contrast" {
				if (scenario == "stable" || scenario == "common_shift" || scenario == "null") && c.Flags > 20 {
					pass = false
				}
				if scenario == "member_shift" && (c.Flags < 800 || c.Prechange != 0) {
					pass = false
				}
			}
			cells = append(cells, c)
		}
	}
	var comparisons []contrastComparison
	for _, scenario := range contrastCases {
		for _, control := range [2]string{"random", "uncertainty"} {
			post, postLower := contrastGain(rows, scenario, control, func(a contrastArm) float64 { return a.PostBrier })
			early, earlyLower := contrastGain(rows, scenario, control, func(a contrastArm) float64 { return a.EarlyBrier })
			full, _ := contrastGain(rows, scenario, control, func(a contrastArm) float64 { return a.FullBrier })
			arrived, _ := contrastGain(rows, scenario, control, func(a contrastArm) float64 { return float64(a.ArrivedLabels) })
			c := contrastComparison{Scenario: scenario, Control: control, PostGain: post,
				PostLower: postLower, EarlyGain: early, EarlyLower: earlyLower,
				StableHarm: -full, ArrivedGap: math.Abs(arrived), Pass: true}
			if scenario == "member_shift" {
				c.Pass = post >= .02 && postLower > 0 && early >= .02 && earlyLower > 0
			}
			if scenario == "stable" && c.StableHarm > .01 {
				c.Pass = false
			}
			if c.ArrivedGap > 2 {
				c.Pass = false
			}
			pass = pass && c.Pass
			comparisons = append(comparisons, c)
		}
	}
	return cells, comparisons, pass
}

func contrastHash(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func TestContrastAcquisitionV1(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_CONTRAST_OUT"), os.Getenv("EVENTFRAME_CONTRAST_SPLIT")
	if path == "" && split == "" {
		t.Skip("opt-in research study")
	}
	seed := int64(2026100801)
	if split == "confirmation" {
		seed = 2026100802
	} else if split != "design" || path == "" {
		t.Fatal("design|confirmation split and new output path required")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	sources := []string{"docs/experiments/mmm-contrast-acquisition-v1-protocol.md",
		"internal/observationgate/contrast_acquisition_v1_test.go",
		"internal/observationgate/member_integration_test.go",
		"internal/observationpreserved/experiment.go", "internal/observation/controller.go",
		"internal/observation/forecast.go"}
	hashes := map[string]string{}
	for _, name := range sources {
		hashes[name], err = contrastHash(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
	}
	start := time.Now()
	rows := make([]contrastRow, 0, len(contrastCases)*contrastTrials)
	fitStart := time.Now()
	var bases [2]*observation.Model
	for i := range bases {
		bases[i], err = observationpreserved.Base(i == 1)
		if err != nil {
			t.Fatal(err)
		}
	}
	fitNS := time.Since(fitStart).Nanoseconds()
	var probability [2][512]float64
	var meanU [2]float64
	for group, base := range bases {
		for x := range probability[group] {
			p, e := base.ForecastObserved(511, uint16(x))
			if e != nil || p <= 0 || p >= 1 {
				t.Fatalf("baseline probability %d: %v", x, e)
			}
			probability[group][x] = p
			meanU[group] += 4 * p * (1 - p) / 512
		}
	}
	for scenarioIndex, scenario := range contrastCases {
		group := 0
		if scenario == "null" {
			group = 1
		}
		for trial := 0; trial < contrastTrials; trial++ {
			tape := contrastTape(seed, scenarioIndex, trial)
			r := contrastRow{Kind: "trial", Split: split, Scenario: scenario, Trial: trial}
			var p [contrastSteps]float64
			for clock, tick := range tape {
				p[clock] = probability[group][tick.x]
			}
			for policyIndex, policy := range contrastPolicies {
				selector := rand.New(rand.NewSource(observationpreserved.Seed(seed, scenarioIndex, trial, 10+policyIndex)))
				arm, e := contrastRunArm(tape, p, meanU[group], policy, selector)
				if e != nil {
					t.Fatalf("%s/%s/%d: %v", scenario, policy, trial, e)
				}
				r.Arms[policyIndex] = arm
			}
			rows = append(rows, r)
		}
	}
	cells, comparisons, pass := contrastSummarize(rows)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(map[string]any{"kind": "header", "split": split, "seedBase": seed,
		"trials": contrastTrials, "steps": contrastSteps, "fitNS": fitNS, "hashes": hashes}); err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if err := enc.Encode(r); err != nil {
			t.Fatal(err)
		}
	}
	if err := enc.Encode(map[string]any{"kind": "summary", "cells": cells,
		"comparisons": comparisons, "pass": pass,
		"elapsedMs": float64(time.Since(start).Microseconds()) / 1000}); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s pass=%t trajectories=%d elapsed=%v", split, pass, len(rows), time.Since(start))
}
