// Mechanically derived v85 driver with the declared age challenger only.
package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
)

func ageLearningRun(base *observation.Model, split string, scenario, index int, seed int64, schedule feedbackSchedule) (ageRecord, error) {
	name := observationpreserved.Scenarios[scenario]
	r := ageRecord{delayedRecord: delayedRecord{Split: split, Scenario: name, Schedule: schedule, Index: index, StreamBase: seed, GateFirst: -1}}
	r.Arms = [2]integrationArm{{Arm: "retained_control", SplitAt: -1}, {Arm: "age_challenger", SplitAt: -1}}
	g := [2]ageJournal{{base: base, enabled: true}, {base: base, enabled: true, ageEnabled: true}}
	var streams [5]*rand.Rand
	for role := range streams {
		streams[role] = rand.New(rand.NewSource(observationpreserved.Seed(seed, scenario, index, role)))
	}
	live, ref, audit, random, timing := streams[0], streams[1], streams[2], streams[3], streams[4]
	var packets []learningPacket
	var audits []receivedAudit
	var monitor observationrescue.Monitor
	var gate pairedInvestigator
	var models observationpreserved.Models
	fit := newSubsetTrial(base, 1)
	hash, latentHash := sha256.New(), sha256.New()
	enc, latentEnc := json.NewEncoder(hash), json.NewEncoder(latentHash)
	post := 256
	if name == "recurring" {
		post = 128
	}
	for step := 0; step < 512; step++ {
		for i := range g {
			if _, e := g[i].expireBefore(max(0, step-48)); e != nil {
				return r, e
			}
		}
		auditNow := audit.Float64() < .25
		x, rx := uint16(live.Intn(512)), uint16(ref.Intn(512))
		reader := observationexperiment.Frames(x, fmt.Sprintf("v85-live-%s-%s-%d-%d", split, name, index, step))
		reference := observationexperiment.Frames(rx, fmt.Sprintf("v85-ref-%s-%s-%d-%d", split, name, index, step))
		bp, e := observation.Run(base, reader, 1, "mmm", 0)
		if e != nil {
			return r, e
		}
		rp, e := observation.Run(base, reference, 1, "mmm", 0)
		if e != nil {
			return r, e
		}
		r.MonitorCost += bp.Cost + rp.Cost
		randSeed := random.Int63()
		var predictions [2]observationpreserved.Prediction
		for i := range g {
			p, e := g[i].predict(reader, step, randSeed)
			if e != nil {
				return r, e
			}
			if p.Cost > 6 || p.Version != models.Version || p.P <= 0 || p.P >= 1 {
				return r, fmt.Errorf("invalid delayed prediction")
			}
			predictions[i] = p
			r.MaxPending = max(r.MaxPending, g[i].stats.Pending)
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
		local := (name == "member_shift" || name == "common_shift") && step >= 256
		if name == "recurring" {
			local = (step/128)%2 == 1
		}
		// The simulator owns truth; learning receives it only through due packets.
		y, ry := integrationLabel(x, local, name == "null", live), integrationLabel(rx, name == "common_shift" && step >= 256, name == "null", ref)
		for i, p := range predictions {
			r.Arms[i].Full.add(p, y)
			if step >= post {
				r.Arms[i].Post.add(p, y)
			}
		}
		if e := latentEnc.Encode(struct {
			Step         int
			X, RX        uint16
			Y, RY, Audit bool
		}{step, x, rx, y, ry, auditNow}); e != nil {
			return r, e
		}
		if e := enc.Encode(struct {
			Step         int
			Predictions  [2]observationpreserved.Prediction
			Y, RY, Audit bool
		}{step, predictions, y, ry, auditNow}); e != nil {
			return r, e
		}
		missingDraw, delayDraw := timing.Float64(), timing.Intn(32)
		if missingDraw < schedule.Missing {
			r.Missing++
		} else {
			delay := schedule.Delay
			if schedule.Jitter {
				delay = delayDraw
			}
			a.Outcome = y
			b.Outcome = ry
			packets = append(packets, learningPacket{step, step + delay, y, ry, auditNow, bp.Probability, rp.Probability, a, b})
		}
		for q := 0; q < len(packets); {
			p := packets[q]
			if p.Due > step {
				q++
				continue
			}
			if p.Origin > step || p.Due > step {
				return r, fmt.Errorf("future packet")
			}
			rc, lc := (p.RefP >= .5) == p.RY, (p.LiveP >= .5) == p.Y
			revision, _ := monitor.Observe(rc, lc, true)
			request, e := gate.observe(r.Received, rc, lc)
			if e != nil {
				return r, e
			}
			authorized := request && bayes.RevisionSplits(revision.Action)
			if request && r.GateFirst < 0 {
				r.GateFirst = step
			}
			var statuses [2]string
			for i := range g {
				status, e := g[i].deliver(p.Origin, p.Y, authorized)
				if e != nil {
					return r, e
				}
				statuses[i] = status
				if g[i].split && r.Arms[i].SplitAt < 0 {
					r.Arms[i].SplitAt = step
				}
			}
			if statuses[0] != statuses[1] {
				return r, fmt.Errorf("unequal feedback policy")
			}
			r.Received++
			if e := enc.Encode(struct {
				Step, Origin int
				Status       [2]string
				Authorized   bool
				Stats        [2]forecastJournalStats
			}{step, p.Origin, statuses, authorized, [2]forecastJournalStats{g[0].stats, g[1].stats}}); e != nil {
				return r, e
			}
			if p.Audit {
				audits = append(audits, receivedAudit{p.Origin, p.Live, p.Reference})
				sort.Slice(audits, func(i, j int) bool { return audits[i].Origin < audits[j].Origin })
				if len(audits) > 256 {
					audits = audits[len(audits)-256:]
				}
				r.ReceivedAudits++
				if r.ReceivedAudits >= 32 && (r.ReceivedAudits-32)%16 == 0 {
					ls, rs := make([]observation.Sample, len(audits)), make([]observation.Sample, len(audits))
					for i, a := range audits {
						ls[i], rs[i] = a.Live, a.Reference
					}
					n := len(ls)
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
					if e := fit.fit(ls[max(0, n-64):]); e != nil {
						return r, e
					}
					age, support, e := fitAgeChallenger(step, audits)
					if e != nil {
						return r, e
					}
					if age != nil {
						r.AgeFits++
						r.AgeSamples += support
					}
					models.Version++
					r.CountFits += 3
					r.SubsetFits++
					for i := range g {
						if e := g[i].publish(models, fit.model, age); e != nil {
							return r, e
						}
					}
				}
			}
			packets = append(packets[:q], packets[q+1:]...)
		}
		for i := range g {
			if g[i].issued != step+1 || g[i].stats.Applied+g[i].stats.Stale+g[i].stats.Censored+g[i].stats.Pending != g[i].issued {
				return r, fmt.Errorf("journal accounting")
			}
		}
	}
	r.PendingPackets = len(packets)
	for i := range g {
		if _, e := g[i].expireBefore(512); e != nil {
			return r, e
		}
		r.Stats[i] = g[i].stats
	}
	if r.Stats[0] != r.Stats[1] || r.Stats[0].Applied+r.Stats[0].Stale != r.Received || r.Missing+r.Received+r.PendingPackets != 512 || r.ReceivedAudits > r.Received || r.Arms[0].SplitAt != r.Arms[1].SplitAt {
		return r, fmt.Errorf("terminal accounting")
	}
	r.Tape, r.LatentTape = hex.EncodeToString(hash.Sum(nil)), hex.EncodeToString(latentHash.Sum(nil))
	return r, nil
}
