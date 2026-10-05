package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// Hindsight boundary is diagnostic-only. Availability remains as-of fit.Clock.
func cleanFitIDs(frames []innerArrivalFrame, fit forestDelayFit, boundary int, seed int64) ([3][]int, error) {
	var out [3][]int
	actual, _, err := eventWindowIDs(frames, fit)
	if err != nil {
		return out, err
	}
	out[0] = actual
	for _, i := range fit.Origins {
		if i >= boundary {
			out[2] = append(out[2], i)
		}
	}
	out[2] = out[2][max(0, len(out[2])-64):]
	if len(out[2]) > len(actual) {
		return out, fmt.Errorf("count mismatch")
	}
	if len(out[2]) == len(actual) {
		out[1] = append([]int(nil), actual...)
	} else {
		perm := rand.New(rand.NewSource(seed)).Perm(len(actual))
		for _, j := range perm[:len(out[2])] {
			out[1] = append(out[1], actual[j])
		}
		sort.Ints(out[1])
	}
	return out, nil
}

func TestCleanFitContracts(t *testing.T) {
	f := make([]innerArrivalFrame, 20)
	ids := make([]int, 20)
	for i := range f {
		f[i] = innerArrivalFrame{Audit: true, Arrival: i}
		ids[i] = i
	}
	fit := forestDelayFit{Clock: 19, Origins: ids}
	got, e := cleanFitIDs(f, fit, 12, 41)
	if e != nil || len(got[1]) != 8 || len(got[2]) != 8 || got[2][0] != 12 {
		t.Fatal(got, e)
	}
	again, e := cleanFitIDs(f, fit, 12, 41)
	if e != nil || !reflect.DeepEqual(got, again) {
		t.Fatal("replay")
	}
	stable, e := cleanFitIDs(f, fit, 0, 41)
	if e != nil || !reflect.DeepEqual(stable[0], stable[1]) || !reflect.DeepEqual(stable[0], stable[2]) {
		t.Fatal("stable control")
	}
	empty, e := cleanFitIDs(f, fit, 20, 41)
	if e != nil || len(empty[1])+len(empty[2]) != 0 {
		t.Fatal("empty")
	}
	f[19].Arrival = 20
	if _, e = cleanFitIDs(f, fit, 12, 41); e == nil {
		t.Fatal("future label accepted")
	}
}

func TestCleanFitExperiment(t *testing.T) {
	output := os.Getenv("EVENTFRAME_CLEAN_OUTPUT")
	if output == "" {
		t.Skip("explicit diagnostic output required")
	}
	parentPath := os.Getenv("EVENTFRAME_CLEAN_PARENT")
	raw, e := os.ReadFile(parentPath)
	if e != nil {
		t.Fatal(e)
	}
	digest := sha256.Sum256(raw)
	const expected = "4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655"
	if hex.EncodeToString(digest[:]) != expected {
		t.Fatal("parent hash")
	}
	f, e := os.Open(parentPath)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var header json.RawMessage
	if e = dec.Decode(&header); e != nil {
		t.Fatal(e)
	}
	out, e := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	hashes := map[string]string{}
	e = filepath.WalkDir(filepath.Join(root, "internal"), func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(p) != ".go" {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		s := sha256.Sum256(b)
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		hashes[rel] = hex.EncodeToString(s[:])
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	for _, p := range []string{"go.mod", "go.sum", "docs/experiments/mmm-clean-fit-v1-contract.md"} {
		b, err := os.ReadFile(filepath.Join(root, p))
		if err != nil {
			t.Fatal(err)
		}
		s := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(s[:])
	}
	if e = enc.Encode(map[string]any{"ParentSHA256": expected, "Hashes": hashes, "Order": []string{"actual", "matched", "clean"}, "DiagnosticOnly": true}); e != nil {
		t.Fatal(e)
	}
	rows := 0
	for {
		var p creditLearningRecord
		e = dec.Decode(&p)
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		scenario := -1
		for i, n := range arrivalSwitchNames {
			if n == p.Case {
				scenario = i
			}
		}
		if scenario < 0 {
			t.Fatal("scenario")
		}
		phase := 0
		if p.Phase == "cohort2" {
			phase = 1
		} else if p.Phase != "cohort1" {
			t.Fatal("phase")
		}
		for si, tape := range []innerArrivalResult{p.Immediate, p.Delayed} {
			for _, clock := range []int{128, 256, 384, 480} {
				var fit *forestDelayFit
				for j := range tape.Fits {
					if tape.Fits[j].Clock < clock {
						fit = &tape.Fits[j]
					}
				}
				if fit == nil {
					// No snapshot exists yet; all diagnostic arms remain neutral.
					fit = &forestDelayFit{Clock: -1}
				}
				boundary := 0
				if scenario >= 2 && clock >= 256 {
					boundary = 256
				}
				seed := int64(2026092207 + phase*1000000 + scenario*100000 + p.Index*1000 + si*500 + clock)
				var ids [3][]int
				if fit.Clock >= 0 {
					var err error
					ids, err = cleanFitIDs(tape.Frames, *fit, boundary, seed)
					if err != nil {
						t.Fatal(err)
					}
				}
				var predictions [3][][2]float64
				for j := range ids {
					m, err := fitEventWindow(tape.Frames, ids[j])
					if err != nil {
						t.Fatal(err)
					}
					for x := uint16(0); x < 512; x++ {
						v, err := m.forecast(511, x)
						if err != nil {
							t.Fatal(err)
						}
						predictions[j] = append(predictions[j], v)
					}
				}
				row := map[string]any{"Phase": p.Phase, "Case": p.Case, "Index": p.Index, "Schedule": []string{"Immediate", "Delayed"}[si], "Checkpoint": clock, "FitClock": fit.Clock, "Boundary": boundary, "Seed": seed, "Origins": ids, "Predictions": predictions}
				if e = enc.Encode(row); e != nil {
					t.Fatal(e)
				}
				rows++
			}
		}
		if p.Index == 15 {
			t.Log(p.Phase, p.Case, "complete")
		}
	}
	if rows != 1024 {
		t.Fatal("record count", rows)
	}
}
