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

	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

const pairFeasSteps = 512
const pairFeasTrials = 1000

var pairFeasStarts = [8]int{0, 64, 128, 192, 256, 320, 384, 448}
var pairFeasCases = [5]string{"stable", "member_shift", "common_shift", "recurring", "null"}
var pairFeasModes = [2]string{"passive", "targeted"}

type pairFeasEvidence struct {
	origin int
	cell   int
	d      int
}

type pairFeasGate struct {
	wealth        [8][2][2]float64
	first         int
	factorUpdates int
}

func newPairFeasGate() pairFeasGate {
	g := pairFeasGate{first: -1}
	for i := range g.wealth {
		for x := range g.wealth[i] {
			g.wealth[i][x] = [2]float64{1, 1}
		}
	}
	return g
}

func (g *pairFeasGate) observe(e pairFeasEvidence) error {
	if e.origin < 0 || e.origin >= pairFeasSteps || e.cell < 0 || e.cell > 1 || e.d < -1 || e.d > 1 {
		return errors.New("invalid complete matched pair")
	}
	for j, start := range pairFeasStarts {
		if e.origin < start {
			continue
		}
		for k, sign := range [2]int{-1, 1} {
			factor := 1 + .8*(float64(sign*e.d)-.1)
			if factor < .12-1e-12 || factor > 1.72+1e-12 {
				return errors.New("unsafe betting factor")
			}
			g.wealth[j][e.cell][k] *= factor
			g.factorUpdates++
		}
	}
	return nil
}

func (g *pairFeasGate) check(clock int) {
	if g.first >= 0 {
		return
	}
	sum := 0.
	for _, start := range g.wealth {
		for _, cell := range start {
			for _, wealth := range cell {
				sum += wealth
			}
		}
	}
	if sum/32 >= 50 {
		g.first = clock
	}
}

type pairFeasRow struct {
	Kind, Split, Mode, Scenario string
	Trial                       int
	FlagAt                      int
	Nominated, Matched          int
	Delivered, ObservedLabels   int
	LiveContexts, RefContexts   int
	FactorUpdates               int
}

func pairFeasRun(split, mode, scenario string, scenarioIndex, modeIndex, trial int, seedBase int64) (pairFeasRow, error) {
	r := pairFeasRow{Kind: "trial", Split: split, Mode: mode, Scenario: scenario, Trial: trial, FlagAt: -1}
	var rng [8]*rand.Rand
	for role := range rng {
		rng[role] = rand.New(rand.NewSource(observationpreserved.Seed(seedBase, scenarioIndex, trial, role+10*modeIndex)))
	}
	liveContext, refContext := rng[0], rng[1]
	liveOutcome, refOutcome := rng[2], rng[3]
	nomination, missingLive, missingRef := rng[4], rng[5], rng[6]
	delay := rng[7]
	var arrivals [pairFeasSteps][]pairFeasEvidence
	var labels [pairFeasSteps]int
	gate := newPairFeasGate()
	for clock := 0; clock < pairFeasSteps; clock++ {
		x := uint16(liveContext.Intn(512))
		r.LiveContexts++
		if nomination.Float64() < .25 {
			r.Nominated++
			var rx uint16
			matched := false
			for attempts := 0; attempts < 512; attempts++ {
				rx = uint16(refContext.Intn(512))
				r.RefContexts++
				matched = x&(1<<2) == rx&(1<<2)
				if mode == "passive" || matched {
					break
				}
			}
			if mode == "targeted" && !matched {
				return r, errors.New("targeted match exceeded 512 attempts")
			}
			if matched {
				r.Matched++
			}
			local := (scenario == "member_shift" || scenario == "common_shift") && clock >= 256
			if scenario == "recurring" {
				local = (clock/128)%2 == 1
			}
			y := integrationLabel(x, local, scenario == "null", liveOutcome)
			ry := integrationLabel(rx, scenario == "common_shift" && clock >= 256, scenario == "null", refOutcome)
			lm, rm := missingLive.Float64() < .2, missingRef.Float64() < .2
			liveDue, refDue := -1, -1
			if !lm {
				liveDue = clock + delay.Intn(32)
				if liveDue < pairFeasSteps {
					labels[liveDue]++
				}
			}
			if !rm {
				refDue = clock + delay.Intn(32)
				if refDue < pairFeasSteps {
					labels[refDue]++
				}
			}
			if matched && liveDue >= 0 && refDue >= 0 && liveDue < pairFeasSteps && refDue < pairFeasSteps {
				arrival := max(liveDue, refDue)
				d := 0
				if ry {
					d++
				}
				if y {
					d--
				}
				arrivals[arrival] = append(arrivals[arrival], pairFeasEvidence{origin: clock, cell: int((x >> 2) & 1), d: d})
			}
		}
		r.ObservedLabels += labels[clock]
		for _, pair := range arrivals[clock] {
			if pair.origin > clock {
				return r, errors.New("future pair delivered")
			}
			if err := gate.observe(pair); err != nil {
				return r, err
			}
			r.Delivered++
		}
		gate.check(clock)
	}
	r.FlagAt, r.FactorUpdates = gate.first, gate.factorUpdates
	if r.ObservedLabels < 2*r.Delivered || r.Delivered > r.Matched || r.RefContexts < r.Nominated {
		return r, errors.New("resource accounting invariant")
	}
	return r, nil
}

type pairFeasCell struct {
	Mode, Scenario string
	Trials         int
	Flags          int
	PrechangeFlags int
	MedianFlag     int
	P95Flag        int
	MeanNominated  float64
	MeanMatched    float64
	MeanDelivered  float64
	MeanLabels     float64
	MeanLiveReads  float64
	MeanRefReads   float64
	MeanFactors    float64
	Pass           bool
}

func pairFeasQuantile(v []int, p float64) int {
	if len(v) == 0 {
		return -1
	}
	sort.Ints(v)
	return v[int(math.Ceil(float64(len(v))*p))-1]
}

func pairFeasSummary(rows []pairFeasRow) ([]pairFeasCell, bool) {
	cells := make([]pairFeasCell, 0, 10)
	pass := true
	for _, mode := range pairFeasModes {
		for _, scenario := range pairFeasCases {
			c := pairFeasCell{Mode: mode, Scenario: scenario, MedianFlag: -1, P95Flag: -1}
			var flags []int
			for _, r := range rows {
				if r.Mode != mode || r.Scenario != scenario {
					continue
				}
				c.Trials++
				c.MeanNominated += float64(r.Nominated)
				c.MeanMatched += float64(r.Matched)
				c.MeanDelivered += float64(r.Delivered)
				c.MeanLabels += float64(r.ObservedLabels)
				c.MeanLiveReads += float64(r.LiveContexts)
				c.MeanRefReads += float64(r.RefContexts)
				c.MeanFactors += float64(r.FactorUpdates)
				if r.FlagAt >= 0 {
					c.Flags++
					flags = append(flags, r.FlagAt)
					if scenario == "member_shift" && r.FlagAt < 256 {
						c.PrechangeFlags++
					}
				}
			}
			if c.Trials != pairFeasTrials {
				panic("missing trial")
			}
			c.MedianFlag, c.P95Flag = pairFeasQuantile(flags, .5), pairFeasQuantile(flags, .95)
			c.MeanNominated /= float64(c.Trials)
			c.MeanMatched /= float64(c.Trials)
			c.MeanDelivered /= float64(c.Trials)
			c.MeanLabels /= float64(c.Trials)
			c.MeanLiveReads /= float64(c.Trials)
			c.MeanRefReads /= float64(c.Trials)
			c.MeanFactors /= float64(c.Trials)
			c.Pass = true
			if scenario == "stable" || scenario == "common_shift" || scenario == "null" {
				c.Pass = c.Flags <= 20
			}
			if scenario == "member_shift" {
				c.Pass = c.PrechangeFlags == 0 && (mode != "targeted" || c.Flags >= 800)
			}
			if mode == "targeted" && c.MeanRefReads > 2.1*c.MeanNominated {
				c.Pass = false
			}
			pass = pass && c.Pass
			cells = append(cells, c)
		}
	}
	return cells, pass
}

func pairFeasHash(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func TestMemberPairFeasibilityV1(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_PAIR_FEAS_OUT"), os.Getenv("EVENTFRAME_PAIR_FEAS_SPLIT")
	if path == "" && split == "" {
		t.Skip("opt-in research study")
	}
	seedBase := int64(2026100701)
	if split == "confirmation" {
		seedBase = 2026100702
	} else if split != "design" {
		t.Fatal("split must be design or confirmation")
	}
	if path == "" {
		t.Fatal("output path required")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	sourceFiles := []string{"docs/experiments/mmm-member-pair-feasibility-v1-protocol.md",
		"internal/observationgate/member_pair_feasibility_v1_test.go",
		"internal/observationgate/member_integration_test.go",
		"internal/observationpreserved/experiment.go"}
	hashes := map[string]string{}
	for _, name := range sourceFiles {
		hashes[name], err = pairFeasHash(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
	}
	start := time.Now()
	rows := make([]pairFeasRow, 0, 10000)
	for modeIndex, mode := range pairFeasModes {
		for scenarioIndex, scenario := range pairFeasCases {
			for trial := 0; trial < pairFeasTrials; trial++ {
				r, e := pairFeasRun(split, mode, scenario, scenarioIndex, modeIndex, trial, seedBase)
				if e != nil {
					t.Fatalf("%s/%s/%d: %v", mode, scenario, trial, e)
				}
				rows = append(rows, r)
			}
		}
	}
	cells, pass := pairFeasSummary(rows)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(map[string]any{"kind": "header", "split": split, "seedBase": seedBase,
		"trials": pairFeasTrials, "steps": pairFeasSteps, "hashes": hashes}); err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if err := enc.Encode(r); err != nil {
			t.Fatal(err)
		}
	}
	if err := enc.Encode(map[string]any{"kind": "summary", "cells": cells, "pass": pass,
		"elapsedMs": float64(time.Since(start).Microseconds()) / 1000}); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s: pass=%t, %d trials, %v", split, pass, len(rows), time.Since(start))
}
