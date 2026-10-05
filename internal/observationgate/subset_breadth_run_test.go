package observationgate

// v80 control workflow plus an independently updated retained-subset arm.
// Hooks cannot modify the control states or their shared audit samples.
import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
	"math/rand"
)

func subsetBreadthRun(base *observation.Model, split string, scenario, index int, seed int64, trial *subsetTrial, cfg subsetBreadthCase) (memberRecord, error) {
	name := cfg.Mode
	r := memberRecord{Split: split, Scenario: cfg.Name, Index: index, GateFirst: [2]int{-1, -1}}
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
		x, rx := breadthInput(uint16(live.Intn(512)), cfg.Dependent), breadthInput(uint16(ref.Intn(512)), cfg.Dependent)
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
		if e := trial.predict(reader, models, step, randSeed); e != nil {
			return r, e
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
		y, ry := breadthLabel(x, local, name == "null", live, cfg.Target), breadthLabel(rx, name == "common_shift" && step >= 256, name == "null", ref, cfg.Target)
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
		if e := trial.feedback(step, y, nomination && g1); e != nil {
			return r, e
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
				if e := trial.fit(ls[max(0, n-64):]); e != nil {
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
