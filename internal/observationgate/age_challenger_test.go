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

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

type ageRecord struct {
	delayedRecord
	AgeFits, AgeSamples int
}

func fitAgeChallenger(step int, audits []receivedAudit) (*observationlearners.ConditionalForest, int, error) {
	if step < 0 || len(audits) > 256 {
		return nil, 0, fmt.Errorf("invalid age fit")
	}
	selected := make([]observation.Sample, 0, 64)
	previous := -1
	for _, a := range audits {
		if a.Origin < 0 || a.Origin <= previous || a.Origin > step {
			return nil, 0, fmt.Errorf("unordered or future audit")
		}
		previous = a.Origin
		if a.Origin > step-128 {
			selected = append(selected, a.Live)
		}
	}
	if len(selected) > 64 {
		selected = selected[len(selected)-64:]
	}
	if len(selected) < 16 {
		return nil, len(selected), nil
	}
	var weights [512]float64
	for i := range weights {
		weights[i] = 1
	}
	m, e := observationlearners.NewSubsetConditional(selected, weights)
	return m, len(selected), e
}

func TestAgeChallengerContracts(t *testing.T) {
	base, trial, models := journalModels(t)
	old := forecastJournal{base: base, enabled: true}
	g := ageJournal{base: base, enabled: true}
	if e := old.publish(models, trial.model); e != nil {
		t.Fatal(e)
	}
	if e := g.publish(models, trial.model, trial.model); e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 128; i++ {
		rd := observationexperiment.Frames(uint16(i*7%512), "age-parity")
		a, e := old.predict(rd, i, 0)
		if e != nil {
			t.Fatal(e)
		}
		b, e := g.predict(rd, i, 0)
		if e != nil || a != b {
			t.Fatal("disabled forecast parity", e)
		}
		x, e := old.deliver(i, i%3 == 0, i == 40)
		if e != nil {
			t.Fatal(e)
		}
		y, e := g.deliver(i, i%3 == 0, i == 40)
		if e != nil || x != y || g.mix != old.mix || g.inner != old.inner {
			t.Fatal("disabled update parity", e)
		}
	}
	audits := make([]receivedAudit, 256)
	for i := range audits {
		audits[i] = receivedAudit{Origin: i, Live: observation.Sample{Bits: uint16(i), Outcome: i%2 == 0}}
	}
	m, n, e := fitAgeChallenger(255, audits)
	if e != nil || m == nil || n != 64 {
		t.Fatal("cap", n, e)
	}
	m, n, e = fitAgeChallenger(383, audits)
	if e != nil || m != nil || n != 0 {
		t.Fatal("age boundary", n, e)
	}
	m, n, e = fitAgeChallenger(256, audits[240:])
	if e != nil || m == nil || n != 16 {
		t.Fatal("minimum support", n, e)
	}
	m, n, e = fitAgeChallenger(256, audits[241:])
	if e != nil || m != nil || n != 15 {
		t.Fatal("support fallback", n, e)
	}
	for _, bad := range [][]receivedAudit{{{Origin: 257}}, {{Origin: 2}, {Origin: 2}}, {{Origin: 2}, {Origin: 1}}} {
		if _, _, e := fitAgeChallenger(256, bad); e == nil {
			t.Fatal("invalid origins accepted")
		}
	}
	age, _, e := fitAgeChallenger(256, audits[240:])
	if e != nil {
		t.Fatal(e)
	}
	if chooseAgeObserver(trial.model, age, [4]float64{.1, .1, .4, .4}) != age || chooseAgeObserver(trial.model, age, [4]float64{.1, .8, .05, .05}) != trial.model || chooseAgeObserver(trial.model, nil, [4]float64{.1, .1, .4, .4}) != trial.model {
		t.Fatal("observer selection")
	}
	g.ageEnabled = true
	models.Version++
	if e := g.publish(models, trial.model, age); e != nil {
		t.Fatal(e)
	}
	g.mix.Weights = [4]float64{.01, .97, .01, .01}
	g.inner.Weights = [4]float64{.05, .05, .45, .45}
	p, e := g.predict(observationexperiment.Frames(241, "active-age"), 128, 0)
	if e != nil {
		t.Fatal(e)
	}
	wantAge, e := age.Forecast(p.Mask, p.Values)
	if e != nil {
		t.Fatal(e)
	}
	if g.entries[0].forecast.inner[2] != wantAge || !g.entries[0].forecast.subsetGuide || p.Cost > 6 {
		t.Fatal("age forecast not wired")
	}
	models.Version++
	if e := g.publish(models, trial.model, age); e != nil {
		t.Fatal(e)
	}
	before := g.inner
	status, e := g.deliver(128, true, false)
	if e != nil || status != "stale" || g.inner != before {
		t.Fatal("stale age update", e)
	}
	for _, schedule := range learningSchedules {
		want, e := delayedLearningRun(base, "unit", 1, 0, 2026118799, schedule)
		if e != nil {
			t.Fatal(e)
		}
		got, e := ageLearningRun(base, "unit", 1, 0, 2026118799, schedule)
		if e != nil {
			t.Fatal(e)
		}
		arm := got.Arms[0]
		arm.Arm = want.Arms[1].Arm
		if arm != want.Arms[1] || got.LatentTape != want.LatentTape || got.Stats != want.Stats || got.AgeFits > got.SubsetFits {
			t.Fatal("control integration parity")
		}
	}
}

func BenchmarkAgeChallenger(b *testing.B) {
	audits := make([]receivedAudit, 64)
	for i := range audits {
		audits[i] = receivedAudit{Origin: i, Live: observation.Sample{Bits: uint16(i * 7 % 512), Outcome: i%2 == 0}}
	}
	for _, n := range []int{16, 64} {
		b.Run(fmt.Sprintf("fit%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, _, e := fitAgeChallenger(64, audits[64-n:]); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
	base, e := observationpreserved.Base(false)
	if e != nil {
		b.Fatal(e)
	}
	retained, _, e := fitAgeChallenger(64, audits)
	if e != nil {
		b.Fatal(e)
	}
	age, _, e := fitAgeChallenger(64, audits[48:])
	if e != nil {
		b.Fatal(e)
	}
	samples := make([]observation.Sample, 64)
	for i, a := range audits {
		samples[i] = a.Live
	}
	count, e := observation.Fit(samples)
	if e != nil {
		b.Fatal(e)
	}
	models := observationpreserved.Models{Short: count, Local: count, Pooled: count, Version: 1}
	for _, enabled := range []bool{false, true} {
		b.Run(fmt.Sprintf("active_age_%t", enabled), func(b *testing.B) {
			g := ageJournal{base: base, enabled: true, ageEnabled: enabled}
			if e := g.publish(models, retained, age); e != nil {
				b.Fatal(e)
			}
			rd := observationexperiment.Frames(241, "age-benchmark")
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g.mix.Weights = [4]float64{.01, .97, .01, .01}
				g.inner.Weights = [4]float64{.05, .05, .45, .45}
				if _, e := g.predict(rd, i, 0); e != nil {
					b.Fatal(e)
				}
				if _, e := g.deliver(i, i%2 == 0, false); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}

func TestAgeChallengerV87(t *testing.T) {
	out, replay := os.Getenv("EVENTFRAME_AGE_OUT"), os.Getenv("EVENTFRAME_AGE_REPLAY")
	if out == "" && replay == "" {
		t.Skip("opt-in")
	}
	if out != "" && replay != "" {
		t.Fatal("choose one")
	}
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	var header struct {
		Version string
		Hashes  map[string]string
	}
	var dec *json.Decoder
	if replay != "" {
		f, e := os.Open(replay)
		if e != nil {
			t.Fatal(e)
		}
		defer f.Close()
		dec = json.NewDecoder(f)
		dec.DisallowUnknownFields()
		if e := dec.Decode(&header); e != nil || header.Version != "v87" {
			t.Fatal("header", e)
		}
	} else {
		header.Version = "v87"
		header.Hashes = map[string]string{}
		var paths []string
		for _, dir := range []string{"observationgate", "observation", "observationlearners", "observationpreserved", "observationrescue", "observationexperiment", "bayes", "model"} {
			ps, e := filepath.Glob(filepath.Join(root, "internal", dir, "*.go"))
			if e != nil {
				t.Fatal(e)
			}
			paths = append(paths, ps...)
		}
		for _, p := range []string{"docs/experiments/mmm-age-challenger-v87-protocol.md", "research/generate-age-challenger-v87.mjs", "research/age-challenger-v87-summary.mjs", "go.mod", "go.sum"} {
			paths = append(paths, filepath.Join(root, p))
		}
		for _, p := range paths {
			b, e := os.ReadFile(p)
			if e != nil {
				t.Fatal(e)
			}
			h := sha256.Sum256(b)
			rel, e := filepath.Rel(root, p)
			if e != nil {
				t.Fatal(e)
			}
			header.Hashes[rel] = hex.EncodeToString(h[:])
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
			var latent [64]string
			for _, schedule := range learningSchedules {
				for i := 0; i < 64; i++ {
					r, e := ageLearningRun(base, split, scenario, i, int64(2026118701+phase), schedule)
					if e != nil {
						t.Fatal(e)
					}
					if latent[i] == "" {
						latent[i] = r.LatentTape
					} else if latent[i] != r.LatentTape {
						t.Fatal("latent mismatch")
					}
					if dec != nil {
						var want ageRecord
						if e := dec.Decode(&want); e != nil || want != r {
							t.Fatal("replay", split, name, schedule.Name, i, e)
						}
					} else {
						if e := enc.Encode(r); e != nil {
							t.Fatal(e)
						}
					}
				}
				t.Log(split, name, schedule.Name, "complete")
			}
		}
	}
	if dec != nil {
		var extra any
		if e := dec.Decode(&extra); e != io.EOF {
			t.Fatal("trailing rows", e)
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
