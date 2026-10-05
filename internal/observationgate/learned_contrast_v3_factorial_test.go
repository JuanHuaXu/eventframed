package observationgate

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

var factorialInputHashes = [...]string{
	"0dcd2eae7c3ed64ee43844e731fba7d07c850f77cdfd7a165e59f4e30ba5fb59",
	"856b8fb095aca8b374be471ea6034d630640b3cf12e280c6837c38a839efc2ac",
}

type factorialCell struct {
	Full, Post, Early, ExpectedPost float64
	RecoveryDelay                   int
	MissedRecovery                  bool
	Arrived, Nominated              int
}

type factorialRow struct {
	Kind, Split, Scenario string
	Fit, Stream           int
	Cells                 [4]factorialCell
}

func replayFactorial(row *v3Row, base *[512]float64, model, schedule int) (factorialCell, error) {
	if model < 0 || model > 1 || (schedule != 2 && schedule != 5) {
		return factorialCell{}, errors.New("invalid factorial cell")
	}
	var old learnedState
	var current v3State
	if model == 0 {
		old = newLearnedState()
	} else {
		current = newV3State()
	}
	var due [learnedSteps][]int
	data := learnedArm{Ticks: make([]learnedArmTick, 0, learnedSteps)}
	delay := 0
	if row.Scenario == "bit2_delayed" {
		delay = 16
	}
	for clock, event := range row.Tape {
		pb := base[event.X]
		p := 0.
		if model == 0 {
			p = old.forecast(event.X, pb)
		} else {
			p = current.forecast(event.X, pb)
		}
		if p <= 0 || p >= 1 || math.IsNaN(p) {
			return factorialCell{}, errors.New("invalid replay forecast")
		}
		selected := row.Arms[schedule].Data.Ticks[clock].Selected
		tick := learnedArmTick{P: p, Selected: selected}
		loss := (p - boolFloat(event.Y)) * (p - boolFloat(event.Y))
		data.FullBrier += loss / learnedSteps
		if clock >= 256 {
			data.PostBrier += loss / 256
			if clock < 320 {
				data.EarlyBrier += loss / 64
			}
		}
		if selected {
			data.Nominated++
			if !event.Missing {
				arrival := clock + delay
				if arrival < learnedSteps {
					due[arrival] = append(due[arrival], clock)
				} else {
					data.Pending++
				}
			}
		}
		for _, origin := range due[clock] {
			if origin > clock || row.Tape[origin].Missing {
				return factorialCell{}, errors.New("future or missing replay label")
			}
			data.Arrived++
			tick.Delivered = append(tick.Delivered, origin)
			observed := row.Tape[origin]
			var err error
			if model == 0 {
				err = old.observe(observed.X, observed.Y, base)
			} else {
				err = current.observe(observed.X, observed.Y, base)
			}
			if err != nil {
				return factorialCell{}, err
			}
		}
		if schedule == 2 && model == 0 || schedule == 5 && model == 1 {
			original := row.Arms[schedule].Data.Ticks[clock]
			if math.Abs(tick.P-original.P) > 1e-12 || tick.Selected != original.Selected ||
				len(tick.Delivered) != len(original.Delivered) {
				return factorialCell{}, errors.New("diagonal replay mismatch")
			}
			for i, origin := range tick.Delivered {
				if origin != original.Delivered[i] {
					return factorialCell{}, errors.New("diagonal delivery mismatch")
				}
			}
		}
		data.Ticks = append(data.Ticks, tick)
	}
	if data.Nominated != learnedBudget || data.Arrived+data.Pending > learnedBudget {
		return factorialCell{}, errors.New("invalid replay acquisition budget")
	}
	measured := v3Measure(data, row.Tape, row.Scenario != "stable" && row.Scenario != "null")
	if schedule == 2 && model == 0 || schedule == 5 && model == 1 {
		original := row.Arms[schedule]
		if math.Abs(data.FullBrier-original.Data.FullBrier) > 1e-12 ||
			math.Abs(data.PostBrier-original.Data.PostBrier) > 1e-12 ||
			math.Abs(measured.PostExpected-original.PostExpected) > 1e-12 ||
			measured.RecoveryDelay != original.RecoveryDelay ||
			measured.MissedRecovery != original.MissedRecovery ||
			data.Arrived != original.Data.Arrived {
			return factorialCell{}, errors.New("diagonal metric mismatch")
		}
	}
	return factorialCell{Full: data.FullBrier, Post: data.PostBrier,
		Early: data.EarlyBrier, ExpectedPost: measured.PostExpected,
		RecoveryDelay: measured.RecoveryDelay, MissedRecovery: measured.MissedRecovery,
		Arrived: data.Arrived, Nominated: data.Nominated}, nil
}

func factorialHash(path string) (string, error) {
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

func TestLearnedContrastV3Factorial(t *testing.T) {
	out := os.Getenv("EVENTFRAME_V3_FACTORIAL_OUT")
	if out == "" {
		t.Skip("opt-in consumed-data diagnostic")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{"docs/experiments/mmm-learned-contrast-v3-factorial-protocol.md",
		"internal/observationgate/learned_contrast_v3_factorial_test.go",
		"internal/observationgate/learned_contrast_v2_test.go",
		"internal/observationgate/learned_contrast_v3_test.go"}
	hashes := make(map[string]string)
	for _, name := range paths {
		hashes[name], err = factorialHash(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
	}
	f, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	if err := encoder.Encode(map[string]any{"kind": "manifest", "inputHashes": factorialInputHashes,
		"sourceHashes": hashes, "cells": [4]string{"old/old", "old/new", "new/old", "new/new"}}); err != nil {
		t.Fatal(err)
	}
	for splitIndex, split := range []string{"design", "confirmation"} {
		input := filepath.Join(root, "docs/experiments/mmm-learned-contrast-v3-"+split+".jsonl.gz")
		actual, e := factorialHash(input)
		if e != nil || actual != factorialInputHashes[splitIndex] {
			t.Fatalf("input archive changed: %s %s %v", split, actual, e)
		}
		archive, e := os.Open(input)
		if e != nil {
			t.Fatal(e)
		}
		gz, e := gzip.NewReader(archive)
		if e != nil {
			archive.Close()
			t.Fatal(e)
		}
		decoder := json.NewDecoder(gz)
		var header map[string]any
		if e := decoder.Decode(&header); e != nil || header["kind"] != "manifest" {
			gz.Close()
			archive.Close()
			t.Fatalf("invalid input header: %v", e)
		}
		fitOffset := 600
		if split == "confirmation" {
			fitOffset = 700
		}
		lastScenario, lastFit := "", -1
		var probabilities [512]float64
		count := 0
		for {
			var source v3Row
			e := decoder.Decode(&source)
			if e == io.EOF {
				break
			}
			if e != nil {
				t.Fatal(e)
			}
			index := -1
			for i, name := range v3Cases {
				if name == source.Scenario {
					index = i
					break
				}
			}
			if source.Kind != "trial" || source.Split != split || index < 0 ||
				source.Fit < 0 || source.Fit >= 16 || source.Stream < 0 || source.Stream >= 16 {
				t.Fatal("invalid archived trial coordinates")
			}
			if source.Scenario != lastScenario || source.Fit != lastFit {
				base, e := observationlearners.Base(observationlearners.Scenarios[0], index, source.Fit+fitOffset)
				if e != nil {
					t.Fatal(e)
				}
				for x := range probabilities {
					probabilities[x], e = base.ForecastObserved(511, uint16(x))
					if e != nil {
						t.Fatal(e)
					}
				}
				lastScenario, lastFit = source.Scenario, source.Fit
			}
			r := factorialRow{Kind: "trial", Split: split, Scenario: source.Scenario,
				Fit: source.Fit, Stream: source.Stream}
			for cell, pair := range [4][2]int{{0, 2}, {0, 5}, {1, 2}, {1, 5}} {
				r.Cells[cell], e = replayFactorial(&source, &probabilities, pair[0], pair[1])
				if e != nil {
					t.Fatalf("%s/%s/%d/%d/%d: %v", split, source.Scenario,
						source.Fit, source.Stream, cell, e)
				}
			}
			if e := encoder.Encode(r); e != nil {
				t.Fatal(e)
			}
			count++
		}
		if e := gz.Close(); e != nil {
			t.Fatal(e)
		}
		if e := archive.Close(); e != nil {
			t.Fatal(e)
		}
		if count != 1792 {
			t.Fatalf("incomplete %s replay: %d", split, count)
		}
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}
