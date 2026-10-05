package observationgate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
)

type feedbackSchedule struct {
	Name    string
	Delay   int
	Jitter  bool
	Missing float64
}

var learningSchedules = []feedbackSchedule{{"immediate", 0, false, 0}, {"delay16", 16, false, 0}, {"missing20", 0, false, .2}, {"jitter31_missing20", 0, true, .2}}

type learningPacket struct {
	Origin, Due     int
	Y, RY, Audit    bool
	LiveP, RefP     float64
	Live, Reference observation.Sample
}
type receivedAudit struct {
	Origin          int
	Live, Reference observation.Sample
}
type delayedRecord struct {
	Split, Scenario                                                                                              string
	Schedule                                                                                                     feedbackSchedule
	Index                                                                                                        int
	StreamBase                                                                                                   int64
	Arms                                                                                                         [2]integrationArm
	Stats                                                                                                        [2]forecastJournalStats
	Received, ReceivedAudits, Missing, PendingPackets, MaxPending, CountFits, SubsetFits, MonitorCost, AuditCost int
	GateFirst                                                                                                    int
	Tape, LatentTape                                                                                             string
}

func delayedLearningRun(base *observation.Model, split string, scenario, index int, seed int64, schedule feedbackSchedule) (delayedRecord, error) {
	name := observationpreserved.Scenarios[scenario]
	r := delayedRecord{Split: split, Scenario: name, Schedule: schedule, Index: index, StreamBase: seed, GateFirst: -1}
	r.Arms = [2]integrationArm{{Arm: "journal_count", SplitAt: -1}, {Arm: "journal_retained_subset", SplitAt: -1}}
	g := [2]forecastJournal{{base: base}, {base: base, enabled: true}}
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
					models.Version++
					r.CountFits += 3
					r.SubsetFits++
					for i := range g {
						if e := g[i].publish(models, fit.model); e != nil {
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

func TestDelayedLearningContracts(t *testing.T) {
	base, e := observationpreserved.Base(false)
	if e != nil {
		t.Fatal(e)
	}
	seed := int64(2026118599)
	r, e := delayedLearningRun(base, "unit", 1, 0, seed, learningSchedules[0])
	if e != nil {
		t.Fatal(e)
	}
	trial := newSubsetTrial(base, 1)
	old, e := subsetIntegrationRun(base, "unit", 1, 0, seed, trial)
	if e != nil {
		t.Fatal(e)
	}
	for i, want := range []integrationArm{old.Arms[3], trial.metrics} {
		got := r.Arms[i]
		got.Arm = want.Arm
		if got != want {
			t.Fatal("immediate quality parity", i, got, want)
		}
	}
	if r.Stats[0].Applied != 512 || r.Stats[0].Stale != 0 || r.Stats[0].Censored != 0 || r.CountFits != old.Fits || r.SubsetFits != trial.fits {
		t.Fatal("immediate accounting")
	}
	for _, schedule := range learningSchedules[1:] {
		a, e := delayedLearningRun(base, "unit", 1, 0, seed, schedule)
		if e != nil {
			t.Fatal(e)
		}
		b, e := delayedLearningRun(base, "unit", 1, 0, seed, schedule)
		if e != nil || a != b {
			t.Fatal("delayed replay", e)
		}
		if a.LatentTape != r.LatentTape {
			t.Fatal("timing changed latent data")
		}
	}
	missing, e := delayedLearningRun(base, "unit", 1, 0, seed, feedbackSchedule{Name: "all_missing", Missing: 1})
	if e != nil {
		t.Fatal(e)
	}
	if missing.Received != 0 || missing.ReceivedAudits != 0 || missing.SubsetFits != 0 || missing.Stats[0].Censored != 512 || missing.Arms[0].SplitAt != -1 || missing.Arms[1].SplitAt != -1 || math.Abs(missing.Arms[0].Full.Brier-missing.Arms[1].Full.Brier) > 1e-12 {
		t.Fatal("learned from absent labels")
	}
}

func TestDelayedLearningV85(t *testing.T) {
	output, replay := os.Getenv("EVENTFRAME_DELAYED_OUT"), os.Getenv("EVENTFRAME_DELAYED_REPLAY")
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
		if e := dec.Decode(&header); e != nil {
			t.Fatal(e)
		}
		if header.Version != "v85" {
			t.Fatal("version")
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
	} else {
		header.Version = "v85"
		header.Hashes = map[string]string{}
		var paths []string
		for _, dir := range []string{"observationgate", "observation", "observationlearners", "observationpreserved", "observationrescue", "observationexperiment", "bayes", "model"} {
			ps, e := filepath.Glob(filepath.Join(root, "internal", dir, "*.go"))
			if e != nil {
				t.Fatal(e)
			}
			paths = append(paths, ps...)
		}
		for _, p := range []string{"docs/experiments/mmm-delayed-learning-v85-protocol.md", "research/delayed-learning-v85-summary.mjs", "go.mod", "go.sum"} {
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
					r, e := delayedLearningRun(base, split, scenario, i, int64(2026118501+phase), schedule)
					if e != nil {
						t.Fatal(e)
					}
					if latent[i] == "" {
						latent[i] = r.LatentTape
					} else if latent[i] != r.LatentTape {
						t.Fatal("schedule changed trajectory")
					}
					if dec != nil {
						var want delayedRecord
						if e := dec.Decode(&want); e != nil {
							t.Fatal(e)
						}
						if r != want {
							t.Fatal("replay mismatch", split, name, schedule.Name, i)
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
			t.Fatal("extra rows", e)
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
