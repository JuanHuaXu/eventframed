package researchswitch

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchswitchref"
)

var studyModes = []string{"static", "slow", "round"}
var studySchedules = []string{"immediate", "fixed150", "uniform299"}

type studyPopulation struct {
	Kind, Geometry, Regime string
	World                  int
	Seed                   int64
	Base                   []float64
	Rates                  [][]float64
	Outcomes               [][]bool
	Changes                []int
}
type controlSnapshot struct {
	Tick, Arrived, Pending                             int
	Forecast                                           []float64
	Brier, PriorityBrier, PacketUsefulness, PacketBias float64
}
type studyControl struct {
	Mode, Schedule                        string
	Issued                                []float64
	Snapshots                             []controlSnapshot
	IssuedBrier, IssuedPriority, Recovery float64
}
type studyFixture struct {
	World struct {
		Population studyPopulation
		Arms       []studyControl
	}
	Due [][][]int
}
type studySnapshot struct {
	Tick, Arrived, Pending int
	Forecast               []float64
	Advice                 [][3]float64
	Weights                [3]float64
}
type studyCosts struct{ SetupNS, IssueNS, ResolveNS, SnapshotNS, AccountedNS, ElapsedNS int64 }
type studyArm struct {
	Mode, Schedule string
	Issued         []float64
	Advice         [][3]float64
	Receipts       []PoolReceipt
	Snapshots      []studySnapshot
	PeakPending    int
	Costs          studyCosts
}
type studyRecord struct {
	Seed int64
	Arms []studyArm
}

func studyHazard(mode string) (float64, error) {
	switch mode {
	case "static":
		return 0, nil
	case "slow":
		return 1. / 2400, nil
	case "round":
		return 1. / 150, nil
	}
	return 0, fmt.Errorf("unknown study mode")
}
func collectStudy(f studyFixture, mode string, s int) (studyArm, error) {
	a := studyArm{Mode: mode, Schedule: studySchedules[s], Issued: make([]float64, 2400), Advice: make([][3]float64, 2400)}
	alpha, err := studyHazard(mode)
	if err != nil {
		return a, err
	}
	started := time.Now()
	phase := time.Now()
	p, err := NewPool(f.World.Population.Base, Config{Prior: []float64{.8, .1, .1}, Hazard: alpha, Trials: 2400, Pending: 2400}, 1)
	a.Costs.SetupNS = time.Since(phase).Nanoseconds()
	if err != nil {
		return a, err
	}
	tickets := make([]PoolTicket, 2400)
	due := f.Due[s]
	last := len(due) - 1
	for len(due[last]) == 0 {
		last--
	}
	for tick := 0; tick <= last; tick++ {
		if tick < 2400 {
			phase = time.Now()
			ticket, e := p.Issue(tick%150, int64(tick))
			if e != nil {
				return a, e
			}
			tickets[tick], a.Issued[tick] = ticket, ticket.Forecast()
			copy(a.Advice[tick][:], p.mix.rows[tick].advice[:3])
			a.PeakPending = max(a.PeakPending, p.Pending())
			a.Costs.IssueNS += time.Since(phase).Nanoseconds()
		}
		for _, trial := range due[tick] {
			phase = time.Now()
			r, e := p.Resolve(tickets[trial], f.World.Population.Outcomes[trial/150][trial%150], int64(tick))
			if e != nil {
				return a, e
			}
			a.Receipts = append(a.Receipts, r)
			a.Costs.ResolveNS += time.Since(phase).Nanoseconds()
		}
		if (tick < 2400 && tick%150 == 149) || tick == last {
			phase = time.Now()
			snap := studySnapshot{Tick: tick, Arrived: len(a.Receipts), Pending: p.Pending(), Forecast: make([]float64, 150), Advice: make([][3]float64, 150)}
			weights := p.mix.probabilities(p.mix.nextLogs())
			copy(snap.Weights[:], weights[:3])
			for i := range snap.Forecast {
				q, e := p.advice(i)
				if e != nil {
					return a, e
				}
				snap.Advice[i] = q
				snap.Forecast[i], e = p.mix.Predict(q[:])
				if e != nil {
					return a, e
				}
			}
			a.Snapshots = append(a.Snapshots, snap)
			a.Costs.SnapshotNS += time.Since(phase).Nanoseconds()
		}
	}
	a.Costs.AccountedNS = a.Costs.SetupNS + a.Costs.IssueNS + a.Costs.ResolveNS + a.Costs.SnapshotNS
	a.Costs.ElapsedNS = time.Since(started).Nanoseconds()
	return a, nil
}
func studyClose(a, b float64) bool { return finite(a) && finite(b) && math.Abs(a-b) <= 3e-12 }
func auditStudy(f studyFixture, a studyArm, s int) error {
	alpha, err := studyHazard(a.Mode)
	if err != nil {
		return err
	}
	if a.Schedule != studySchedules[s] || len(a.Issued) != 2400 || len(a.Advice) != 2400 || len(f.World.Arms) != 9 || len(f.Due) != 3 {
		return fmt.Errorf("study domain")
	}
	for h, mode := range []string{"full", "adaptive", "rich_moment2"} {
		if f.World.Arms[s*3+h].Mode != mode || f.World.Arms[s*3+h].Schedule != a.Schedule {
			return fmt.Errorf("control domain")
		}
	}
	c := a.Costs
	if c.SetupNS < 0 || c.IssueNS < 0 || c.ResolveNS < 0 || c.SnapshotNS < 0 || c.AccountedNS != c.SetupNS+c.IssueNS+c.ResolveNS+c.SnapshotNS || c.ElapsedNS < c.AccountedNS {
		return fmt.Errorf("cost conservation")
	}
	known := map[int]bool{}
	position, snapshot, peak := 0, 0, 0
	due := f.Due[s]
	last := len(due) - 1
	for len(due[last]) == 0 {
		last--
	}
	prior := []float64{.8, .1, .1}
	for tick := 0; tick <= last; tick++ {
		weights, err := researchswitchref.EndWeights(prior, alpha, a.Advice[:min(tick, 2400)], known)
		if err != nil {
			return err
		}
		if tick > 0 {
			for h := range weights {
				weights[h] = (1-alpha)*weights[h] + alpha*prior[h]
			}
		}
		if tick < 2400 {
			q := 0.
			for h, v := range a.Advice[tick] {
				if !studyClose(v, f.World.Arms[s*3+h].Issued[tick]) {
					return fmt.Errorf("original audited expert advice %d", tick)
				}
				q += weights[h] * v
			}
			if !studyClose(q, a.Issued[tick]) {
				return fmt.Errorf("independent issued mixture %d", tick)
			}
			peak = max(peak, tick+1-position)
		}
		for _, trial := range due[tick] {
			if position >= len(a.Receipts) {
				return fmt.Errorf("missing receipt")
			}
			r := a.Receipts[position]
			if r.Epoch != 1 || r.TrialOrdinal != trial+1 || r.Member != trial%150 || r.MemberOrdinal != trial/150+1 || r.IssuedAt != int64(trial) || r.ArrivedAt != int64(tick) || r.Forecast != a.Issued[trial] || r.Useful != f.World.Population.Outcomes[trial/150][trial%150] {
				return fmt.Errorf("original receipt identity")
			}
			if _, seen := known[trial]; seen {
				return fmt.Errorf("duplicate receipt")
			}
			known[trial] = r.Useful
			position++
		}
		if (tick < 2400 && tick%150 == 149) || tick == last {
			if snapshot >= len(a.Snapshots) {
				return fmt.Errorf("missing snapshot")
			}
			snap := a.Snapshots[snapshot]
			if snap.Tick != tick || snap.Arrived != position || snap.Pending != min(tick+1, 2400)-position || len(snap.Advice) != 150 || len(snap.Forecast) != 150 {
				return fmt.Errorf("snapshot coverage")
			}
			weights, err = researchswitchref.EndWeights(prior, alpha, a.Advice[:min(tick+1, 2400)], known)
			if err != nil {
				return err
			}
			for h := range weights {
				weights[h] = (1-alpha)*weights[h] + alpha*prior[h]
				if !studyClose(weights[h], snap.Weights[h]) {
					return fmt.Errorf("independent snapshot weights")
				}
			}
			for i, q := range snap.Forecast {
				want := 0.
				for h, v := range snap.Advice[i] {
					control := f.World.Arms[s*3+h]
					if snapshot >= len(control.Snapshots) || control.Snapshots[snapshot].Tick != tick || !studyClose(v, control.Snapshots[snapshot].Forecast[i]) {
						return fmt.Errorf("audited snapshot advice")
					}
					want += weights[h] * v
				}
				if !studyClose(want, q) {
					return fmt.Errorf("independent snapshot mixture")
				}
			}
			snapshot++
		}
	}
	if position != 2400 || snapshot != len(a.Snapshots) || peak != a.PeakPending {
		return fmt.Errorf("complete drain conservation")
	}
	return nil
}
func TestSwitchStudyV43(t *testing.T) {
	input, out := os.Getenv("EVENTFRAME_SWITCH_V43_FIXTURE"), os.Getenv("EVENTFRAME_SWITCH_V43_OUT")
	if out == "" {
		t.Skip("explicit output required")
	}
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := json.NewDecoder(bufio.NewReader(f))
	var manifest map[string]any
	if err = d.Decode(&manifest); err != nil {
		t.Fatal(err)
	}
	w, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	b := bufio.NewWriter(w)
	enc := json.NewEncoder(b)
	manifest["Kind"] = "switch_study_manifest"
	if err = enc.Encode(manifest); err != nil {
		t.Fatal(err)
	}
	count := 0
	for {
		var fixture studyFixture
		err = d.Decode(&fixture)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		record := studyRecord{Seed: fixture.World.Population.Seed}
		for s := 0; s < 3; s++ {
			for _, mode := range studyModes {
				a, err := collectStudy(fixture, mode, s)
				if err != nil {
					t.Fatal(err)
				}
				record.Arms = append(record.Arms, a)
			}
		}
		if err = enc.Encode(record); err != nil {
			t.Fatal(err)
		}
		count++
		t.Log(fixture.World.Population.Geometry + "/" + fixture.World.Population.Regime)
	}
	if float64(count) != manifest["Worlds"] {
		t.Fatal("world coverage")
	}
	if err = b.Flush(); err != nil {
		t.Fatal(err)
	}
	if err = w.Sync(); err != nil {
		t.Fatal(err)
	}
}
func TestSwitchStudyAuditV43(t *testing.T) {
	input, raw := os.Getenv("EVENTFRAME_SWITCH_V43_FIXTURE"), os.Getenv("EVENTFRAME_SWITCH_V43_AUDIT")
	if raw == "" {
		t.Skip("explicit audit required")
	}
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w, err := os.Open(raw)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	fd, rd := json.NewDecoder(bufio.NewReader(f)), json.NewDecoder(bufio.NewReader(w))
	rd.DisallowUnknownFields()
	var fm, rm map[string]any
	if err = fd.Decode(&fm); err != nil {
		t.Fatal(err)
	}
	if err = rd.Decode(&rm); err != nil {
		t.Fatal(err)
	}
	if rm["Kind"] != "switch_study_manifest" {
		t.Fatal("manifest kind")
	}
	rm["Kind"] = fm["Kind"]
	if !reflect.DeepEqual(fm, rm) {
		t.Fatal("manifest mismatch")
	}
	count := 0
	for {
		var fixture studyFixture
		err = fd.Decode(&fixture)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		var record studyRecord
		if err = rd.Decode(&record); err != nil {
			t.Fatal(err)
		}
		if record.Seed != fixture.World.Population.Seed || len(record.Arms) != 9 {
			t.Fatal("record identity")
		}
		for j, a := range record.Arms {
			if a.Mode != studyModes[j%3] {
				t.Fatal("arm order")
			}
			if err = auditStudy(fixture, a, j/3); err != nil {
				t.Fatal(err)
			}
		}
		count++
		t.Log(fixture.World.Population.Geometry + "/" + fixture.World.Population.Regime)
	}
	if float64(count) != fm["Worlds"] {
		t.Fatal("audit world coverage")
	}
	var extra any
	if err = rd.Decode(&extra); err != io.EOF {
		t.Fatal("extra record", err)
	}
}

func testStudyFixture() studyFixture {
	var f studyFixture
	f.World.Population.Base = benchmarkBase150()
	f.World.Population.Outcomes = make([][]bool, 16)
	for r := range f.World.Population.Outcomes {
		f.World.Population.Outcomes[r] = make([]bool, 150)
		for i := range f.World.Population.Outcomes[r] {
			f.World.Population.Outcomes[r][i] = (i%3 == 0) != (r >= 8)
		}
	}
	f.Due = make([][][]int, 3)
	for s := 0; s < 3; s++ {
		f.Due[s] = make([][]int, 2700)
		rng := rand.New(rand.NewSource(2026104397))
		for j := 0; j < 2400; j++ {
			delay := 0
			if s == 1 {
				delay = 150
			}
			if s == 2 {
				delay = rng.Intn(300)
			}
			f.Due[s][j+delay] = append(f.Due[s][j+delay], j)
		}
	}
	return f
}

func TestStudyAuditRejectsSemanticCorruptions(t *testing.T) {
	f := testStudyFixture()
	a, err := collectStudy(f, "round", 0)
	if err != nil {
		t.Fatal(err)
	}
	// Unit-test advice tapes exercise the auditor's boundaries. These are NOT
	// the independently audited generator/control fixtures in the real study.
	f.World.Arms = make([]studyControl, 9)
	for h, mode := range []string{"full", "adaptive", "rich_moment2"} {
		c := studyControl{Mode: mode, Schedule: "immediate", Issued: make([]float64, 2400)}
		for j := range c.Issued {
			c.Issued[j] = a.Advice[j][h]
		}
		for _, snap := range a.Snapshots {
			s := controlSnapshot{Tick: snap.Tick, Forecast: make([]float64, 150)}
			for i := range s.Forecast {
				s.Forecast[i] = snap.Advice[i][h]
			}
			c.Snapshots = append(c.Snapshots, s)
		}
		f.World.Arms[h] = c
	}
	if err := auditStudy(f, a, 0); err != nil {
		t.Fatal("positive audit control", err)
	}
	mutations := []struct {
		name   string
		change func(*studyFixture, *studyArm)
	}{
		{"advice", func(f *studyFixture, a *studyArm) { a.Advice[0][0] += .1 }},
		{"issued", func(f *studyFixture, a *studyArm) { a.Issued[0] += .1 }},
		{"epoch", func(f *studyFixture, a *studyArm) { a.Receipts[0].Epoch++ }},
		{"arrival", func(f *studyFixture, a *studyArm) { a.Receipts[0].ArrivedAt++ }},
		{"member", func(f *studyFixture, a *studyArm) { a.Receipts[0].Member++ }},
		{"original-score", func(f *studyFixture, a *studyArm) { a.Receipts[0].Forecast += .1 }},
		{"snapshot-law", func(f *studyFixture, a *studyArm) { a.Snapshots[0].Forecast[0] += .1 }},
		{"snapshot-advice", func(f *studyFixture, a *studyArm) { a.Snapshots[0].Advice[0][0] += .1 }},
		{"snapshot-weight", func(f *studyFixture, a *studyArm) { a.Snapshots[0].Weights[0] += .1 }},
		{"pending", func(f *studyFixture, a *studyArm) { a.PeakPending++ }},
		{"cost", func(f *studyFixture, a *studyArm) { a.Costs.AccountedNS++ }},
		{"label", func(f *studyFixture, a *studyArm) {
			f.World.Population.Outcomes[0][0] = !f.World.Population.Outcomes[0][0]
		}},
		{"missing-control", func(f *studyFixture, a *studyArm) { f.World.Arms = f.World.Arms[:8] }},
	}
	fb, _ := json.Marshal(f)
	ab, _ := json.Marshal(a)
	for _, m := range mutations {
		t.Run(m.name, func(t *testing.T) {
			var cf studyFixture
			var ca studyArm
			if err := json.Unmarshal(fb, &cf); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(ab, &ca); err != nil {
				t.Fatal(err)
			}
			m.change(&cf, &ca)
			if err := auditStudy(cf, ca, 0); err == nil {
				t.Fatal("corruption accepted")
			}
		})
	}
}

func TestStudyCollectorFuturePrefix(t *testing.T) {
	for s := 0; s < 3; s++ {
		for _, mode := range studyModes {
			f, g := testStudyFixture(), testStudyFixture()
			for r := 8; r < 16; r++ {
				for i := 0; i < 150; i++ {
					g.World.Population.Outcomes[r][i] = !g.World.Population.Outcomes[r][i]
				}
			}
			a, err := collectStudy(f, mode, s)
			if err != nil {
				t.Fatal(err)
			}
			b, err := collectStudy(g, mode, s)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(a.Issued[:1201], b.Issued[:1201]) || !reflect.DeepEqual(a.Advice[:1201], b.Advice[:1201]) {
				t.Fatal("future altered issued prefix", mode, s)
			}
			for j, snap := range a.Snapshots {
				if snap.Tick < 1200 && !reflect.DeepEqual(snap, b.Snapshots[j]) {
					t.Fatal("future altered snapshot", mode, s)
				}
			}
		}
	}
}
