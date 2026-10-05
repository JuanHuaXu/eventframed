package observationgate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
)

func TestPairedInvestigatorContracts(t *testing.T) {
	for _, direction := range []bool{false, true} {
		var g pairedInvestigator
		for i := 0; i < 512; i++ {
			r, l := true, true
			if i >= 128 {
				r, l = direction, !direction
			}
			if _, e := g.observe(i, r, l); e != nil {
				t.Fatal(e)
			}
			if g.next != i+1 || g.evidence.components[0].base.next != i+1 {
				t.Fatal("replicated evidence")
			}
			before := g
			if _, e := g.observe(i, r, l); e == nil || g != before {
				t.Fatal("duplicate support")
			}
			if _, e := g.observe(i+2, r, l); e == nil || g != before {
				t.Fatal("out-of-order support")
			}
		}
		if !g.evidence.alert {
			t.Fatal("missing scalar divergence")
		}
	}
	var g pairedInvestigator
	for i := 0; i < 512; i++ {
		a, e := g.observe(i, i%2 == 0, i%2 == 0)
		if e != nil || a {
			t.Fatal("identical members")
		}
	}
}

func TestPairedScalarRounding(t *testing.T) {
	for seed := int64(0); seed < 20; seed++ {
		rng := rand.New(rand.NewSource(seed))
		var g pairedInvestigator
		for i := 0; i < 512; i++ {
			if _, e := g.observe(i, rng.Intn(2) == 1, rng.Intn(2) == 1); e != nil {
				s, _ := preparePredictiveBet(g.history)
				t.Fatalf("seed=%d step=%d q=%.18g c=%.18g error=%v", seed, i, s.q, s.c, e)
			}
		}
	}
	var g pairedInvestigator
	g.history.values[1][0] = 1
	before := g
	if _, e := g.observe(0, true, false); e == nil || g != before {
		t.Fatal("non-scalar history silently normalized")
	}
}

func integrationLabel(x uint16, local, null bool, rng *rand.Rand) bool {
	if null {
		return rng.Intn(2) == 1
	}
	y := (x&(1<<6) != 0) != (x&(1<<7) != 0)
	y = y != (x&(1<<8) != 0)
	if local {
		y = x&(1<<2) != 0
	}
	if rng.Float64() < .05 {
		y = !y
	}
	return y
}

type integrationMetric struct {
	N, Correct, Cost int
	Brier, LogLoss   float64
}

func (m *integrationMetric) add(p observationpreserved.Prediction, y bool) {
	t := 0.
	if y {
		t = 1
		m.LogLoss -= math.Log(p.P)
	} else {
		m.LogLoss -= math.Log1p(-p.P)
	}
	m.Brier += (p.P - t) * (p.P - t)
	m.N++
	m.Cost += p.Cost
	if (p.P >= .5) == y {
		m.Correct++
	}
}

type integrationArm struct {
	Arm        string
	Full, Post integrationMetric
	SplitAt    int
}
type memberRecord struct {
	Split, Scenario              string
	Index                        int
	Arms                         [5]integrationArm
	GateFirst                    [2]int
	MonitorCost, AuditCost, Fits int
	Tape                         string
}

func memberRun(base *observation.Model, split string, scenario, index int, seed int64) (memberRecord, error) {
	name := observationpreserved.Scenarios[scenario]
	r := memberRecord{Split: split, Scenario: name, Index: index, GateFirst: [2]int{-1, -1}}
	var streams [4]*rand.Rand
	for role := range streams {
		streams[role] = rand.New(rand.NewSource(observationpreserved.Seed(seed, scenario, index, role)))
	}
	live, ref, audit, random := streams[0], streams[1], streams[2], streams[3]
	arms := [5]string{"frozen_mmm", "mix_mmm_no_ap", "mix_mmm_ap", "mix_mmm_ap", "mix_mmm_local"}
	var states [5]*observationpreserved.State
	for i, a := range arms {
		s, e := observationpreserved.New(base, a)
		if e != nil {
			return r, e
		}
		states[i] = s
		r.Arms[i] = integrationArm{Arm: a, SplitAt: -1}
	}
	r.Arms[2].Arm = "old_gate_ap"
	r.Arms[3].Arm = "mixture_gate_ap"
	var models observationpreserved.Models
	var monitor observationrescue.Monitor
	var old observationpreserved.Evidence
	var next pairedInvestigator
	var ls, rs []observation.Sample
	audits := 0
	hash := sha256.New()
	enc := json.NewEncoder(hash)
	post := 256
	if name == "recurring" {
		post = 128
	}
	for step := 0; step < 512; step++ {
		auditNow := audit.Float64() < .25
		x, rx := uint16(live.Intn(512)), uint16(ref.Intn(512))
		reader := observationexperiment.Frames(x, fmt.Sprintf("v80-live-%s-%s-%d-%d", split, name, index, step))
		reference := observationexperiment.Frames(rx, fmt.Sprintf("v80-ref-%s-%s-%d-%d", split, name, index, step))
		f, e := observation.Run(base, reader, 1, "mmm", 0)
		if e != nil {
			return r, e
		}
		rf, e := observation.Run(base, reference, 1, "mmm", 0)
		if e != nil {
			return r, e
		}
		r.MonitorCost += f.Cost + rf.Cost
		randSeed := random.Int63()
		var predictions [5]observationpreserved.Prediction
		for i, s := range states {
			p, e := s.Predict(reader, models, step, randSeed)
			if e != nil {
				return r, e
			}
			if p.Cost > 6 || p.Version != models.Version || p.Values&^p.Mask != 0 || p.P <= 0 || p.P >= 1 {
				return r, fmt.Errorf("invalid pre-label forecast")
			}
			predictions[i] = p
		}
		var a, b observation.Sample
		if auditNow {
			r.AuditCost += 18
			for j, rd := range []observation.Reader{reader, reference} {
				var mask, values uint16
				for scope := 0; scope < 3; scope++ {
					m, v, e := rd.Read(observation.View{Scope: scope, Depth: 2})
					if e != nil {
						return r, e
					}
					mask |= m
					values |= v
				}
				if mask != 511 {
					return r, fmt.Errorf("incomplete audit")
				}
				if j == 0 {
					a.Bits = values
				} else {
					b.Bits = values
				}
			}
		}
		// No prediction or audit acquisition sees the revealing labels above.
		local := (name == "member_shift" || name == "common_shift") && step >= 256
		if name == "recurring" {
			local = (step/128)%2 == 1
		}
		y, ry := integrationLabel(x, local, name == "null", live), integrationLabel(rx, name == "common_shift" && step >= 256, name == "null", ref)
		rc, lc := (rf.Probability >= .5) == ry, (f.Probability >= .5) == y
		revision, _ := monitor.Observe(rc, lc, true)
		nomination := bayes.RevisionSplits(revision.Action)
		g0 := old.Observe(rc, lc)
		g1, e := next.observe(step, rc, lc)
		if e != nil {
			return r, e
		}
		gates := [2]bool{g0, g1}
		for i, v := range gates {
			if v && r.GateFirst[i] < 0 {
				r.GateFirst[i] = step
			}
		}
		for i, s := range states {
			authorized := false
			if i == 2 {
				authorized = nomination && g0
			}
			if i == 3 {
				authorized = nomination && g1
			}
			if e := s.Observe(step, y, authorized); e != nil {
				return r, e
			}
			if authorized && r.Arms[i].SplitAt < 0 {
				r.Arms[i].SplitAt = step
			}
			r.Arms[i].Full.add(predictions[i], y)
			if step >= post {
				r.Arms[i].Post.add(predictions[i], y)
			}
		}
		if e := enc.Encode(struct {
			Step                     int
			Predictions              [5]observationpreserved.Prediction
			Y, RY, Audit, Nomination bool
			Gates                    [2]bool
		}{step, predictions, y, ry, auditNow, nomination, gates}); e != nil {
			return r, e
		}
		if auditNow {
			a.Outcome = y
			b.Outcome = ry
			ls = append(ls, a)
			rs = append(rs, b)
			audits++
			if len(ls) > 256 {
				ls = ls[1:]
				rs = rs[1:]
			}
			if audits >= 32 && (audits-32)%16 == 0 {
				n := len(ls)
				var e error
				models.Short, e = observation.Fit(ls[max(0, n-64):])
				if e != nil {
					return r, e
				}
				models.Local, e = observation.Fit(ls)
				if e != nil {
					return r, e
				}
				pooled := append([]observation.Sample{}, ls[max(0, n-128):]...)
				pooled = append(pooled, rs[max(0, n-128):]...)
				models.Pooled, e = observation.Fit(pooled)
				if e != nil {
					return r, e
				}
				models.Version++
				r.Fits += 3
			}
		}
	}
	r.Tape = hex.EncodeToString(hash.Sum(nil))
	return r, nil
}

func TestMemberIntegrationParity(t *testing.T) {
	base, e := observationpreserved.Base(false)
	if e != nil {
		t.Fatal(e)
	}
	for _, scenario := range []int{0, 1} {
		a, e := memberRun(base, "unit", scenario, 0, 2026118099)
		if e != nil {
			t.Fatal(e)
		}
		b, e := memberRun(base, "unit", scenario, 0, 2026118099)
		if e != nil || a != b {
			t.Fatal("replay", e)
		}
		old, e := observationpreserved.RunStream(base, "unit", scenario, 0, 2026118099)
		if e != nil {
			t.Fatal(e)
		}
		for i, j := range []int{0, 3, 5, 5, 4} {
			if i == 3 {
				continue
			}
			p := a.Arms[i]
			q := old.Arms[j]
			if math.Abs(p.Full.Brier/512-q.Full.Brier) > 1e-14 || p.Full.Cost != q.Foreground || p.SplitAt != q.SplitAt {
				t.Fatal("archived state parity", scenario, i)
			}
		}
	}
	if base.Support() != 4096 {
		t.Fatal("incumbent mutated")
	}
}

func TestMemberV80Experiment(t *testing.T) {
	output, replay := os.Getenv("EVENTFRAME_MEMBER_OUT"), os.Getenv("EVENTFRAME_MEMBER_REPLAY")
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
	var paths []string
	for _, dir := range []string{"observationgate", "observation", "observationpreserved", "observationrescue", "observationexperiment", "bayes", "model"} {
		ps, e := filepath.Glob(filepath.Join(root, "internal", dir, "*.go"))
		if e != nil {
			t.Fatal(e)
		}
		paths = append(paths, ps...)
	}
	for _, p := range []string{"docs/experiments/mmm-member-integration-v80-protocol.md", "research/member-integration-v80-summary.mjs", "go.mod", "go.sum"} {
		paths = append(paths, filepath.Join(root, p))
	}
	hashes := map[string]string{}
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
		hashes[rel] = hex.EncodeToString(h[:])
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if e := enc.Encode(struct {
		Version string
		Hashes  map[string]string
	}{"v80", hashes}); e != nil {
		t.Fatal(e)
	}
	for phase, split := range []string{"design", "confirmation"} {
		for scenario, name := range observationpreserved.Scenarios {
			base, e := observationpreserved.Base(name == "null")
			if e != nil {
				t.Fatal(e)
			}
			for i := 0; i < 64; i++ {
				r, e := memberRun(base, split, scenario, i, int64(2026118001+phase))
				if e != nil {
					t.Fatal(e)
				}
				if e := enc.Encode(r); e != nil {
					t.Fatal(e)
				}
			}
			t.Log(split, name, "complete")
		}
	}
	if replay != "" {
		want, e := os.ReadFile(replay)
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(want, buf.Bytes()) {
			t.Fatal("artifact/source replay mismatch")
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
