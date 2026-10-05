package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"testing"
)

// This reconstruction checks the archived score boundary used by the offline
// localization study. It does not validate a change detector or confidence set.
func TestResearchAuditScoreReconstruction(t *testing.T) {
	input := os.Getenv("EVENTFRAME_AUDIT_SCORE_INPUT")
	output := os.Getenv("EVENTFRAME_AUDIT_SCORE_OUTPUT")
	if input == "" || output == "" {
		t.Skip("set both audit-score artifact paths")
	}
	raw, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	const expected = "4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655"
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != expected {
		t.Fatal("input hash mismatch")
	}
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var header json.RawMessage
	if err := dec.Decode(&header); err != nil {
		t.Fatal(err)
	}
	counts := [2]int{}
	maxDelta := 0.0
	rows := 0
	for {
		var row creditLearningRecord
		err := dec.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		phase := 0
		if row.Phase == "cohort2" {
			phase = 1
		} else if row.Phase != "cohort1" {
			t.Fatal("unknown phase")
		}
		scenario := -1
		for i, name := range arrivalSwitchNames {
			if row.Case == name {
				scenario = i
				break
			}
		}
		if scenario < 0 {
			t.Fatal("unknown scenario")
		}
		base, masks, err := arrivalSwitchSetup(phase, scenario, row.Index)
		if err != nil || masks != row.Masks {
			t.Fatal("frozen base setup", err)
		}
		for schedule, tape := range []innerArrivalResult{row.Immediate, row.Delayed} {
			if len(tape.Frames) != 512 || len(row.Acquisition[schedule]) != 512 {
				t.Fatal("incomplete schedule")
			}
			for origin, frame := range tape.Frames {
				if !frame.Audit || frame.Missing {
					continue
				}
				if !row.Acquisition[schedule][origin].Full || frame.Arrival < origin || frame.Arrival > origin+31 {
					t.Fatalf("invalid audit record %s/%d/%d", row.Case, schedule, origin)
				}
				live, err := base.ForecastObserved(511, frame.X)
				if err != nil {
					t.Fatal(err)
				}
				reference, err := base.ForecastObserved(511, frame.RX)
				if err != nil {
					t.Fatal(err)
				}
				for _, pair := range [][2]float64{{live, frame.Baseline}, {reference, frame.Reference}} {
					delta := math.Abs(pair[0] - pair[1])
					maxDelta = math.Max(maxDelta, delta)
					if delta > 1e-12 || !isFiniteProbability(pair[0]) || !isFiniteProbability(pair[1]) {
						t.Fatalf("score mismatch %s/%d/%d: %v", row.Case, schedule, origin, pair)
					}
					counts[schedule]++
				}
			}
		}
		rows++
	}
	if rows != 128 || counts[0] == 0 || counts[1] == 0 {
		t.Fatal("incomplete reconstruction", rows, counts)
	}
	artifact := struct {
		InputSHA256 string  `json:"inputSHA256"`
		Rows        int     `json:"rows"`
		Scores      [2]int  `json:"scoresBySchedule"`
		MaxDelta    float64 `json:"maxDelta"`
		Limit       string  `json:"limit"`
	}{expected, rows, counts, maxDelta, "Frozen full-input score reconstruction only; no localization or coverage result."}
	out, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	if err := json.NewEncoder(out).Encode(artifact); err != nil {
		t.Fatal(err)
	}
	t.Log(fmt.Sprintf("reconstructed %d/%d scores, max delta %.3g", counts[0], counts[1], maxDelta))
}

func isFiniteProbability(p float64) bool {
	return !math.IsNaN(p) && !math.IsInf(p, 0) && p >= 0 && p <= 1
}
