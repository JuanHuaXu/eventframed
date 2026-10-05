package observationgate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

type forecastWindow struct {
	N, Split, Bit2, Parity int
	Guide                  [3]int
	Weights, ExpertBrier   [4]float64
	MixtureBrier           float64
	Available, Support     [4]int
	ModelBrier             [4][4]float64 // base/short/local/pooled by actual/all/bit2/parity
}
type forecastPending struct {
	p         observationpreserved.Prediction
	available [4]bool
	support   [4]int
	probs     [4][4]float64
}
type forecastDiagnostic struct {
	Windows [8]forecastWindow
	next    int
	pending *forecastPending
}

func (d *forecastDiagnostic) prepare(step int, x uint16, p observationpreserved.Prediction, m observationpreserved.Models, base *observation.Model) error {
	if d.pending != nil || step != d.next || step < 0 || step >= 512 || x >= 512 || p.Values != x&p.Mask {
		return fmt.Errorf("invalid diagnostic preparation")
	}
	next := forecastPending{p: p}
	models := [4]*observation.Model{base, m.Short, m.Local, m.Pooled}
	masks := [4]uint16{p.Mask, 511, 1 << 2, (1 << 6) | (1 << 7) | (1 << 8)}
	for i, model := range models {
		if model == nil {
			continue
		}
		next.available[i] = true
		next.support[i] = int(model.Support())
		for j, mask := range masks {
			v, e := model.ForecastObserved(mask, x&mask)
			if e != nil {
				return e
			}
			next.probs[i][j] = v
		}
	}
	expected := [4]float64{.5, .5, .5, .5}
	expected[0] = next.probs[0][0]
	if next.available[1] {
		expected[1] = next.probs[1][0]
	}
	long := 3
	if p.Split {
		long = 2
	}
	if next.available[long] {
		expected[2] = next.probs[long][0]
	}
	for j, v := range expected {
		if math.Abs(v-p.Experts[j]) > 1e-14 {
			return fmt.Errorf("journaled expert mismatch")
		}
	}
	if p.Guide < 0 || p.Guide >= 3 {
		return fmt.Errorf("invalid guide")
	}
	d.pending = &next
	return nil
}

func (d *forecastDiagnostic) observe(step int, y bool) error {
	if d.pending == nil || step != d.next {
		return fmt.Errorf("missing or duplicate diagnostic feedback")
	}
	s := d.pending
	p := s.p
	w := &d.Windows[step/64]
	w.N++
	if p.Split {
		w.Split++
	}
	if p.Mask&(1<<2) != 0 {
		w.Bit2++
	}
	if p.Mask&448 == 448 {
		w.Parity++
	}
	w.Guide[p.Guide]++
	t := 0.
	if y {
		t = 1
	}
	w.MixtureBrier += (p.P - t) * (p.P - t)
	for j := range w.Weights {
		w.Weights[j] += p.Weights[j]
		w.ExpertBrier[j] += (p.Experts[j] - t) * (p.Experts[j] - t)
	}
	for i, available := range s.available {
		if !available {
			continue
		}
		w.Available[i]++
		w.Support[i] += s.support[i]
		for j, v := range s.probs[i] {
			w.ModelBrier[i][j] += (v - t) * (v - t)
		}
	}
	d.pending = nil
	d.next++
	return nil
}

func TestForecastDiagnosticContracts(t *testing.T) {
	base, e := observationpreserved.Base(false)
	if e != nil {
		t.Fatal(e)
	}
	state, e := observationpreserved.New(base, "mix_mmm_ap")
	if e != nil {
		t.Fatal(e)
	}
	p, e := state.Predict(observationexperiment.Frames(511, "diagnostic"), observationpreserved.Models{}, 0, 0)
	if e != nil {
		t.Fatal(e)
	}
	var d forecastDiagnostic
	if e := d.observe(0, true); e == nil {
		t.Fatal("feedback without forecast")
	}
	if e := d.prepare(0, 511, p, observationpreserved.Models{}, base); e != nil {
		t.Fatal(e)
	}
	if e := d.prepare(0, 511, p, observationpreserved.Models{}, base); e == nil {
		t.Fatal("duplicate preparation")
	}
	if e := d.observe(1, true); e == nil {
		t.Fatal("out of order feedback")
	}
	if e := d.observe(0, true); e != nil {
		t.Fatal(e)
	}
	if e := d.observe(0, true); e == nil {
		t.Fatal("duplicate feedback")
	}
	if d.Windows[0].Available != [4]int{1, 0, 0, 0} || d.Windows[0].MixtureBrier != (p.P-1)*(p.P-1) {
		t.Fatal("availability/score")
	}
	a, e := memberRun(base, "unit", 1, 0, 2026118199)
	if e != nil {
		t.Fatal(e)
	}
	var diag forecastDiagnostic
	b, e := memberDiagnosticRun(base, "unit", 1, 0, 2026118199, &diag)
	if e != nil || a != b || diag.next != 512 || diag.pending != nil {
		t.Fatal("observer changed experiment", e)
	}
	for _, w := range diag.Windows {
		if w.N != 64 {
			t.Fatal("window count")
		}
	}
}

type forecastRecord struct {
	Original memberRecord
	Windows  [8]forecastWindow
}

func TestForecastDiagnosticV81(t *testing.T) {
	output, replay := os.Getenv("EVENTFRAME_FORECAST_DIAGNOSTIC_OUT"), os.Getenv("EVENTFRAME_FORECAST_DIAGNOSTIC_REPLAY")
	if output == "" && replay == "" {
		t.Skip("opt-in")
	}
	if output != "" && replay != "" {
		t.Fatal("choose write or replay")
	}
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(filepath.Join(root, "docs/experiments/mmm-member-integration-v80.jsonl"))
	if e != nil {
		t.Fatal(e)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	var oldHeader struct {
		Version string
		Hashes  map[string]string
	}
	if e := dec.Decode(&oldHeader); e != nil {
		t.Fatal(e)
	}
	if oldHeader.Version != "v80" {
		t.Fatal("version")
	}
	hashes := oldHeader.Hashes
	for p, want := range hashes {
		b, e := os.ReadFile(filepath.Join(root, p))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			t.Fatal("v80 source drift", p)
		}
	}
	for _, p := range []string{"internal/observationgate/forecast_diagnostic_test.go", "internal/observationgate/member_diagnostic_replay_test.go", "docs/experiments/mmm-post-split-v81-protocol.md", "research/post-split-v81-summary.mjs", "docs/experiments/mmm-member-integration-v80.jsonl"} {
		b, e := os.ReadFile(filepath.Join(root, p))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if e := enc.Encode(struct {
		Version string
		Hashes  map[string]string
	}{"v81-posthoc", hashes}); e != nil {
		t.Fatal(e)
	}
	for phase, split := range []string{"design", "confirmation"} {
		for scenario, name := range observationpreserved.Scenarios {
			base, e := observationpreserved.Base(name == "null")
			if e != nil {
				t.Fatal(e)
			}
			for i := 0; i < 64; i++ {
				var want memberRecord
				if e := dec.Decode(&want); e != nil {
					t.Fatal(e)
				}
				var diag forecastDiagnostic
				got, e := memberDiagnosticRun(base, split, scenario, i, int64(2026118001+phase), &diag)
				if e != nil {
					t.Fatal(e)
				}
				if got != want || diag.next != 512 || diag.pending != nil {
					t.Fatal("v80 mismatch", split, name, i)
				}
				if e := enc.Encode(forecastRecord{got, diag.Windows}); e != nil {
					t.Fatal(e)
				}
			}
			t.Log(split, name, "complete")
		}
	}
	var extra any
	if e := dec.Decode(&extra); e != io.EOF {
		t.Fatal("extra source rows", e)
	}
	if replay != "" {
		want, e := os.ReadFile(replay)
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(want, buf.Bytes()) {
			t.Fatal("diagnostic replay/source mismatch")
		}
		return
	}
	f, e := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
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
