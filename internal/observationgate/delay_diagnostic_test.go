package observationgate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

type delayWindow struct {
	N, Available, SubsetGuide               int
	Guide                                   [3]int
	MixtureBrier                            float64
	ExpertBrier, Weights                    [4]float64
	InnerBrier, InnerWeights                [2]float64
	ModelAge, NewestAge, OldestAge, Support int
	CurrentRegimeFraction                   float64
	Applied, Stale, AppliedAge, StaleAge    int
}
type delayDiagnostic struct {
	Windows       [8]delayWindow
	next, lastFit int
	origins       []int
}

func (d *delayDiagnostic) fit(step int, audits []receivedAudit) {
	d.lastFit = step
	d.origins = make([]int, len(audits))
	for i, a := range audits {
		d.origins[i] = a.Origin
	}
}
func (d *delayDiagnostic) capture(step int, name string, y bool, p subsetPending) error {
	if step != d.next || step < 0 || step >= 512 || p.p.Guide < 0 || p.p.Guide > 2 {
		return fmt.Errorf("diagnostic order")
	}
	d.next++
	w := &d.Windows[step/64]
	w.N++
	w.Guide[p.p.Guide]++
	target := 0.
	if y {
		target = 1
	}
	square := func(x float64) float64 { return (x - target) * (x - target) }
	w.MixtureBrier += square(p.p.P)
	for i := range w.Weights {
		w.Weights[i] += p.p.Weights[i]
		w.ExpertBrier[i] += square(p.p.Experts[i])
	}
	if p.available {
		if len(d.origins) == 0 || d.lastFit >= step {
			return fmt.Errorf("invalid fitting history")
		}
		w.Available++
		if p.subsetGuide {
			w.SubsetGuide++
		}
		w.InnerBrier[0] += square(p.inner[0])
		w.InnerBrier[1] += square(p.inner[1])
		w.InnerWeights[0] += p.innerWeights[0]
		w.InnerWeights[1] += p.innerWeights[1] + p.innerWeights[2] + p.innerWeights[3]
		w.ModelAge += step - d.lastFit
		w.NewestAge += step - d.origins[len(d.origins)-1]
		w.OldestAge += step - d.origins[0]
		w.Support += len(d.origins)
		// Simulator regime membership is reporting-only and never enters a predictor.
		regime := func(origin int) int {
			if name == "recurring" {
				return origin / 128 % 2
			}
			if (name == "member_shift" || name == "common_shift") && origin >= 256 {
				return 1
			}
			return 0
		}
		same := 0
		for _, origin := range d.origins {
			if origin >= step {
				return fmt.Errorf("future training audit")
			}
			if regime(origin) == regime(step) {
				same++
			}
		}
		w.CurrentRegimeFraction += float64(same) / float64(len(d.origins))
	}
	return nil
}
func (d *delayDiagnostic) feedback(step, origin int, status string) error {
	if step < 0 || step >= 512 || origin < 0 || origin > step {
		return fmt.Errorf("invalid feedback age")
	}
	w := &d.Windows[step/64]
	switch status {
	case "applied":
		w.Applied++
		w.AppliedAge += step - origin
	case "stale":
		w.Stale++
		w.StaleAge += step - origin
	default:
		return fmt.Errorf("invalid feedback status")
	}
	return nil
}

type delayDiagnosticRecord struct {
	Original delayedRecord
	Windows  [8]delayWindow
}

func TestDelayDiagnosticContracts(t *testing.T) {
	base, e := observationpreserved.Base(false)
	if e != nil {
		t.Fatal(e)
	}
	for _, schedule := range learningSchedules {
		want, e := delayedLearningRun(base, "unit", 1, 0, 2026118699, schedule)
		if e != nil {
			t.Fatal(e)
		}
		var d delayDiagnostic
		got, e := delayedLearningDiagnosticRun(base, "unit", 1, 0, 2026118699, schedule, &d)
		if e != nil || got != want {
			t.Fatal("diagnostic changed run", e)
		}
		n, applied, stale := 0, 0, 0
		for _, w := range d.Windows {
			n += w.N
			applied += w.Applied
			stale += w.Stale
			if w.N != 64 || w.Available > w.N || w.CurrentRegimeFraction < 0 || w.CurrentRegimeFraction > float64(w.Available) || w.SubsetGuide > w.Available {
				t.Fatal("invalid diagnostic counts")
			}
		}
		if n != 512 || applied != got.Stats[1].Applied || stale != got.Stats[1].Stale {
			t.Fatal("feedback accounting")
		}
	}
}

func TestDelayDiagnosticV86(t *testing.T) {
	out, replay := os.Getenv("EVENTFRAME_DELAY_DIAG_OUT"), os.Getenv("EVENTFRAME_DELAY_DIAG_REPLAY")
	if out == "" && replay == "" {
		t.Skip("opt-in")
	}
	if out != "" && replay != "" {
		t.Fatal("choose output or replay")
	}
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(filepath.Join(root, "docs/experiments/mmm-delayed-learning-v85.jsonl"))
	if e != nil {
		t.Fatal(e)
	}
	original := json.NewDecoder(bytes.NewReader(raw))
	original.DisallowUnknownFields()
	var oldHeader struct {
		Version string
		Hashes  map[string]string
	}
	if e := original.Decode(&oldHeader); e != nil || oldHeader.Version != "v85" {
		t.Fatal("v85 header", e)
	}
	hash := sha256.Sum256(raw)
	header := struct {
		Version, ParentSHA256 string
		Hashes                map[string]string
	}{"v86", hex.EncodeToString(hash[:]), oldHeader.Hashes}
	var saved *json.Decoder
	if replay != "" {
		f, e := os.Open(replay)
		if e != nil {
			t.Fatal(e)
		}
		defer f.Close()
		saved = json.NewDecoder(f)
		saved.DisallowUnknownFields()
		var want struct {
			Version, ParentSHA256 string
			Hashes                map[string]string
		}
		if e := saved.Decode(&want); e != nil || want.Version != "v86" || want.ParentSHA256 != header.ParentSHA256 {
			t.Fatal("diagnostic header", e)
		}
		header = want
	} else {
		for _, p := range []string{"internal/observationgate/delay_diagnostic_run_test.go", "internal/observationgate/delay_diagnostic_test.go", "research/generate-delay-diagnostic-v86.mjs", "docs/experiments/mmm-delay-diagnostic-v86-protocol.md"} {
			b, e := os.ReadFile(filepath.Join(root, p))
			if e != nil {
				t.Fatal(e)
			}
			h := sha256.Sum256(b)
			header.Hashes[p] = hex.EncodeToString(h[:])
		}
	}
	for p, want := range header.Hashes {
		if !filepath.IsLocal(p) {
			t.Fatal("source path")
		}
		b, e := os.ReadFile(filepath.Join(root, p))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			t.Fatal("source changed", p)
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if e := enc.Encode(header); e != nil {
		t.Fatal(e)
	}
	for phase, split := range []string{"design", "confirmation"} {
		for scenario, name := range observationpreserved.Scenarios {
			base, e := observationpreserved.Base(name == "null")
			if e != nil {
				t.Fatal(e)
			}
			for _, schedule := range learningSchedules {
				for i := 0; i < 64; i++ {
					var want delayedRecord
					if e := original.Decode(&want); e != nil {
						t.Fatal(e)
					}
					var d delayDiagnostic
					got, e := delayedLearningDiagnosticRun(base, split, scenario, i, int64(2026118501+phase), schedule, &d)
					if e != nil || got != want {
						t.Fatal("v85 parity", split, name, schedule.Name, i, e)
					}
					record := delayDiagnosticRecord{got, d.Windows}
					if saved != nil {
						var prior delayDiagnosticRecord
						if e := saved.Decode(&prior); e != nil || prior != record {
							t.Fatal("diagnostic replay", e)
						}
					} else {
						if e := enc.Encode(record); e != nil {
							t.Fatal(e)
						}
					}
				}
				t.Log(split, name, schedule.Name, "parity complete")
			}
		}
	}
	var extra any
	if e := original.Decode(&extra); e != io.EOF {
		t.Fatal("v85 trailing data", e)
	}
	if saved != nil {
		if e := saved.Decode(&extra); e != io.EOF {
			t.Fatal("diagnostic trailing data", e)
		}
		return
	}
	f, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if _, e := f.Write(buf.Bytes()); e != nil {
		t.Fatal(e)
	}
	if e := f.Sync(); e != nil {
		t.Fatal(e)
	}
}
