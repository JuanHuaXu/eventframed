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

	"github.com/JuanHuaXu/eventframed/internal/bayes"
)

// Separate issued advice from feedback: the selector never receives a frame's
// stored truth until its delivery clock, even though the offline tape has it.
type viewAdvice struct{ Original, Available float64 }
type viewFeedback struct {
	Origin, Arrival  int
	Outcome, Missing bool
}
type viewIssued struct {
	P       float64
	Weights [4]float64
}

func (a viewAdvice) experts() [4]float64 {
	return [4]float64{a.Original, a.Available, a.Available, a.Available}
}

func replayViewExperts(advice []viewAdvice, feedback []viewFeedback) ([]viewIssued, int, error) {
	n := len(advice)
	if n == 0 || len(feedback) != n {
		return nil, 0, fmt.Errorf("empty or mismatched tape")
	}
	arrivals := make([][]viewFeedback, n+32)
	seen := make([]bool, n)
	for _, a := range advice {
		if !bayes.ValidForecastExperts(a.experts()) {
			return nil, 0, fmt.Errorf("invalid advice")
		}
	}
	// Require origin-ordered input so simultaneous labels have one declared order.
	for i, f := range feedback {
		if f.Origin != i || seen[f.Origin] || f.Arrival < i || f.Arrival > i+31 {
			return nil, 0, fmt.Errorf("invalid feedback timing")
		}
		seen[f.Origin] = true
		if !f.Missing {
			arrivals[f.Arrival] = append(arrivals[f.Arrival], f)
		}
	}
	var mix bayes.ForecastMix
	issued := make([]viewIssued, n)
	updates := 0
	for clock, batch := range arrivals {
		if clock < n {
			issued[clock] = viewIssued{mix.Forecast(advice[clock].experts()), mix.Weights}
		}
		for _, f := range batch {
			mix = mix.Observe(advice[f.Origin].experts(), f.Outcome, 1)
			updates++
		}
	}
	return issued, updates, nil
}

func TestViewExpertsContracts(t *testing.T) {
	a := []viewAdvice{{.8, .2}, {.7, .3}, {.6, .4}, {.9, .1}}
	f := []viewFeedback{{0, 2, true, false}, {1, 2, false, false}, {2, 3, true, true}, {3, 3, true, false}}
	p, n, err := replayViewExperts(a, f)
	if err != nil || n != 3 {
		t.Fatal(n, err)
	}
	f[0].Outcome = false
	q, _, _ := replayViewExperts(a, f)
	for i := 0; i <= 2; i++ {
		if p[i] != q[i] {
			t.Fatal("future feedback affected forecast", i)
		}
	}
	if p[3] == q[3] {
		t.Fatal("delivery did not affect subsequent forecast")
	}
	f[0].Outcome = true
	f[2].Outcome = false
	r, _, _ := replayViewExperts(a, f)
	for i := range p {
		if p[i] != r[i] {
			t.Fatal("missing outcome leaked")
		}
	}
	var mix bayes.ForecastMix
	mix = mix.Observe(a[0].experts(), true, 1).Observe(a[1].experts(), false, 1)
	if p[3].Weights != mix.Weights || p[3].P != mix.Forecast(a[3].experts()) {
		t.Fatal("simultaneous delivery order")
	}
	for i := range a {
		a[i].Available = a[i].Original
	}
	r, _, err = replayViewExperts(a, f)
	if err != nil {
		t.Fatal(err)
	}
	for i := range r {
		if math.Abs(r[i].P-a[i].Original) > 1e-14 {
			t.Fatal("equal advice drift")
		}
	}
	for _, v := range []float64{-1, 2, math.NaN(), math.Inf(1)} {
		b := append([]viewAdvice(nil), a...)
		b[0].Original = v
		if _, _, err := replayViewExperts(b, f); err == nil {
			t.Fatal("invalid probability accepted")
		}
	}
	for _, bad := range []viewFeedback{{0, -1, true, false}, {0, 32, true, false}, {1, 0, true, false}} {
		b := append([]viewFeedback(nil), f...)
		b[0] = bad
		if _, _, err := replayViewExperts(a, b); err == nil {
			t.Fatal("invalid timing accepted")
		}
	}
}

func BenchmarkViewExpertsStep(b *testing.B) {
	var mix bayes.ForecastMix
	a := viewAdvice{.7, .8}.experts()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = mix.Forecast(a)
		mix = mix.Observe(a, i%3 != 0, 1)
	}
	if mix.Weights[0] < 0 {
		b.Fatal("invalid")
	}
}

func TestViewExpertsExperiment(t *testing.T) {
	in, parent, out := os.Getenv("EVENTFRAME_VIEW_INPUT"), os.Getenv("EVENTFRAME_VIEW_PARENT"), os.Getenv("EVENTFRAME_VIEW_OUTPUT")
	if in == "" || parent == "" || out == "" {
		t.Skip("opt-in consumed view selection")
	}
	hash := func(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
	read := func(p string) []byte {
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	ir, pr := read(in), read(parent)
	if hash(ir) != "3a5fb1fd119c7c14aaeb1864a95e3895a18781bc3e21e75ef5f7dcef485cb76e" || hash(pr) != "4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655" {
		t.Fatal("wrong parent tape")
	}
	f, e := os.Open(in)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	g, e := os.Open(parent)
	if e != nil {
		t.Fatal(e)
	}
	defer g.Close()
	d, pd := json.NewDecoder(f), json.NewDecoder(g)
	var header struct{ ParentSHA256 string }
	if e = d.Decode(&header); e != nil || header.ParentSHA256 != hash(pr) {
		t.Fatal("parent binding", e)
	}
	var ignored json.RawMessage
	if e = pd.Decode(&ignored); e != nil {
		t.Fatal(e)
	}
	o, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer o.Close()
	enc := json.NewEncoder(o)
	sources, hashes := map[string]string{}, map[string]string{}
	for _, path := range []string{"internal/observationgate/view_experts_test.go", "internal/bayes/forecast_mix.go", "go.mod", "go.sum", "docs/experiments/mmm-view-experts-v1-contract.md", "research/view-experts-summary.mjs"} {
		b := read("../../" + path)
		sources[path] = string(b)
		hashes[path] = hash(b)
	}
	if e = enc.Encode(map[string]any{"InputSHA256": hash(ir), "ParentSHA256": hash(pr), "Sources": sources, "Hashes": hashes, "Consumed": true}); e != nil {
		t.Fatal(e)
	}
	n := 0
	for {
		var row availableEvidenceRecord
		if e = d.Decode(&row); e == io.EOF {
			break
		} else if e != nil {
			t.Fatal(e)
		}
		var parent creditLearningRecord
		if e = pd.Decode(&parent); e != nil {
			t.Fatal(e)
		}
		if row.Phase != parent.Phase || row.Case != parent.Case || row.Index != parent.Index {
			t.Fatal("row mismatch")
		}
		result := struct {
			Phase, Case        string
			Index              int
			Immediate, Delayed []viewIssued
			Updates            [2]int
		}{Phase: row.Phase, Case: row.Case, Index: row.Index}
		for s, diag := range [][]availableEvidenceFrame{row.Immediate, row.Delayed} {
			frames := parent.Immediate.Frames
			if s == 1 {
				frames = parent.Delayed.Frames
			}
			if len(diag) != 512 || len(frames) != 512 {
				t.Fatal("incomplete schedule")
			}
			a, f := make([]viewAdvice, 512), make([]viewFeedback, 512)
			for i, x := range diag {
				if x.Origin != i || math.Abs(x.Original-frames[i].Predictions[2].P) > 1e-12 {
					t.Fatal("advice binding")
				}
				a[i] = viewAdvice{x.Original, x.Available}
				f[i] = viewFeedback{i, frames[i].Arrival, frames[i].Y, frames[i].Missing}
			}
			issued, updates, e := replayViewExperts(a, f)
			if e != nil {
				t.Fatal(e)
			}
			if s == 0 {
				result.Immediate = issued
			} else {
				result.Delayed = issued
			}
			result.Updates[s] = updates
		}
		if e = enc.Encode(result); e != nil {
			t.Fatal(e)
		}
		n++
	}
	if n != 128 {
		t.Fatal("incomplete rows", n)
	}
	if e = pd.Decode(&ignored); e != io.EOF {
		t.Fatal("trailing parent", e)
	}
	if e = o.Sync(); e != nil {
		t.Fatal(e)
	}
}
