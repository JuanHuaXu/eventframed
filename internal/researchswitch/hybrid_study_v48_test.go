package researchswitch

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchhybridref"
	"github.com/JuanHuaXu/eventframed/internal/researchswitchref"
)

var hybridModesV48 = []string{"static", "slow", "round"}

type hybridSnapshotV48 struct {
	Tick, Arrived, Pending int
	Forecast               []float64
	Advice                 [][3]float64
	Weights                [][3]float64
	GlobalWeights          [3]float64
	ScopeWeights           [2]float64
	Heads                  [][2]float64
}
type hybridArmV48 struct {
	Mode, Schedule string
	Issued         []float64
	Advice         [][3]float64
	Heads          [][2]float64
	Receipts       []PoolReceipt
	Snapshots      []hybridSnapshotV48
	PeakPending    int
	Costs          studyCosts
}
type hybridRecordV48 struct {
	Seed int64
	Arms []hybridArmV48
}

func hybridHazardV48(mode string) (float64, error) {
	switch mode {
	case "static":
		return 0, nil
	case "slow":
		return 1. / 2400, nil
	case "round":
		return 1. / 150, nil
	}
	return 0, fmt.Errorf("unknown hybrid mode")
}

func collectHybridV48(f studyFixture, mode string, s int) (hybridArmV48, error) {
	a := hybridArmV48{Mode: mode, Schedule: studySchedules[s], Issued: make([]float64, 2400), Advice: make([][3]float64, 2400), Heads: make([][2]float64, 2400)}
	alpha, err := hybridHazardV48(mode)
	if err != nil {
		return a, err
	}
	started, phase := time.Now(), time.Now()
	p, err := NewHybridPool(f.World.Population.Base, Config{Prior: []float64{.8, .1, .1}, Hazard: 1. / 16, Trials: 2400, Pending: 2400}, 16, 1./2400, alpha, 1)
	a.Costs.SetupNS = time.Since(phase).Nanoseconds()
	if err != nil {
		return a, err
	}
	tickets := make([]HybridTicket, 2400)
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
			copy(a.Advice[tick][:], p.global.rows[tick].advice[:3])
			copy(a.Heads[tick][:], p.scope.rows[tick].advice[:2])
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
			snap := hybridSnapshotV48{Tick: tick, Arrived: len(a.Receipts), Pending: p.Pending(), Forecast: make([]float64, 150), Advice: make([][3]float64, 150), Weights: make([][3]float64, 150), Heads: make([][2]float64, 150)}
			gw := p.global.probabilities(p.global.nextLogs())
			sw := p.scope.probabilities(p.scope.nextLogs())
			copy(snap.GlobalWeights[:], gw[:3])
			copy(snap.ScopeWeights[:], sw[:2])
			for m := range snap.Forecast {
				q, heads, e := p.advice(m)
				if e != nil {
					return a, e
				}
				snap.Advice[m] = q
				snap.Heads[m] = heads
				weights := p.local.mixes[m].probabilities(p.local.mixes[m].nextLogs())
				copy(snap.Weights[m][:], weights[:3])
				snap.Forecast[m], e = p.scope.Predict(heads[:])
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

// Reference visits each task's full nomination tape through independent dense
// multiplication; neither candidate caches nor its unit-span code are used.
func auditHybridV48(f studyFixture, a hybridArmV48, s int) error {
	alpha, err := hybridHazardV48(a.Mode)
	if err != nil {
		return err
	}
	if s < 0 || s >= 3 || len(f.Due) != 3 || len(f.World.Arms) != 9 || len(f.World.Population.Base) != 150 || len(f.World.Population.Outcomes) != 16 || a.Schedule != studySchedules[s] || len(a.Issued) != 2400 || len(a.Advice) != 2400 || len(a.Heads) != 2400 {
		return fmt.Errorf("local study domain")
	}
	for _, y := range f.World.Population.Outcomes {
		if len(y) != 150 {
			return fmt.Errorf("outcome domain")
		}
	}
	for h, mode := range []string{"full", "adaptive", "rich_moment2"} {
		c := f.World.Arms[s*3+h]
		if c.Mode != mode || c.Schedule != a.Schedule || len(c.Issued) != 2400 {
			return fmt.Errorf("control domain")
		}
	}
	c := a.Costs
	if c.SetupNS <= 0 || c.IssueNS <= 0 || c.ResolveNS <= 0 || c.SnapshotNS <= 0 || c.AccountedNS != c.SetupNS+c.IssueNS+c.ResolveNS+c.SnapshotNS || c.ElapsedNS < c.AccountedNS {
		return fmt.Errorf("cost conservation")
	}
	tapes := make([][][3]float64, 150)
	known := make([]map[int]bool, 150)
	for m := range known {
		known[m] = map[int]bool{}
	}
	prior := []float64{.8, .1, .1}
	weights := func(m int) ([3]float64, error) {
		w, err := researchswitchref.EndWeights(prior, 1./16, tapes[m], known[m])
		if err == nil && len(tapes[m]) > 0 {
			for h := range w {
				w[h] = (1-1./16)*w[h] + (1./16)*prior[h]
			}
		}
		return w, err
	}
	globalKnown := map[int]bool{}
	var globalTape [][3]float64
	var headTape [][]float64
	globalWeights := func() ([3]float64, error) {
		w, err := researchswitchref.EndWeights(prior, 1./2400, globalTape, globalKnown)
		if err == nil && len(globalTape) > 0 {
			for h := range w {
				w[h] = (1-1./2400)*w[h] + (1./2400)*prior[h]
			}
		}
		return w, err
	}
	scopeWeights := func() ([]float64, error) {
		prior := []float64{.9, .1}
		w, err := researchhybridref.End(prior, alpha, headTape, globalKnown)
		if err == nil && len(headTape) > 0 {
			for h := range w {
				w[h] = (1-alpha)*w[h] + alpha*prior[h]
			}
		}
		return w, err
	}
	position, snapshot, peak := 0, 0, 0
	due := f.Due[s]
	if len(due) < 2400 {
		return fmt.Errorf("arrival domain")
	}
	last := len(due) - 1
	for last >= 0 && len(due[last]) == 0 {
		last--
	}
	for tick := 0; tick <= last; tick++ {
		if tick < 2400 {
			m := tick % 150
			w, err := weights(m)
			if err != nil {
				return err
			}
			gw, err := globalWeights()
			if err != nil {
				return err
			}
			sw, err := scopeWeights()
			if err != nil {
				return err
			}
			var heads [2]float64
			for h, v := range a.Advice[tick] {
				if !studyClose(v, f.World.Arms[s*3+h].Issued[tick]) {
					return fmt.Errorf("original expert advice")
				}
				heads[0] += gw[h] * v
				heads[1] += w[h] * v
			}
			for h := range heads {
				if !studyClose(heads[h], a.Heads[tick][h]) {
					return fmt.Errorf("original head advice")
				}
			}
			q := sw[0]*heads[0] + sw[1]*heads[1]
			if !studyClose(q, a.Issued[tick]) {
				return fmt.Errorf("independent issued mixture")
			}
			tapes[m] = append(tapes[m], a.Advice[tick])
			globalTape = append(globalTape, a.Advice[tick])
			headTape = append(headTape, a.Heads[tick][:])
			peak = max(peak, tick+1-position)
		}
		for _, trial := range due[tick] {
			if trial < 0 || trial >= min(tick+1, 2400) || position >= len(a.Receipts) {
				return fmt.Errorf("future or missing receipt")
			}
			r := a.Receipts[position]
			if r.Epoch != 1 || r.TrialOrdinal != trial+1 || r.Member != trial%150 || r.MemberOrdinal != trial/150+1 || r.IssuedAt != int64(trial) || r.ArrivedAt != int64(tick) || r.Forecast != a.Issued[trial] || r.Useful != f.World.Population.Outcomes[trial/150][trial%150] {
				return fmt.Errorf("original receipt identity")
			}
			if _, ok := known[trial%150][trial/150]; ok {
				return fmt.Errorf("duplicate receipt")
			}
			known[trial%150][trial/150] = r.Useful
			globalKnown[trial] = r.Useful
			position++
		}
		if (tick < 2400 && tick%150 == 149) || tick == last {
			if snapshot >= len(a.Snapshots) {
				return fmt.Errorf("missing snapshot")
			}
			snap := a.Snapshots[snapshot]
			if snap.Tick != tick || snap.Arrived != position || snap.Pending != min(tick+1, 2400)-position || len(snap.Advice) != 150 || len(snap.Forecast) != 150 || len(snap.Weights) != 150 || len(snap.Heads) != 150 {
				return fmt.Errorf("snapshot coverage")
			}
			gw, err := globalWeights()
			if err != nil {
				return err
			}
			sw, err := scopeWeights()
			if err != nil {
				return err
			}
			for h := range gw {
				if !studyClose(gw[h], snap.GlobalWeights[h]) {
					return fmt.Errorf("global snapshot weights")
				}
			}
			for h := range sw {
				if !studyClose(sw[h], snap.ScopeWeights[h]) {
					return fmt.Errorf("scope snapshot weights")
				}
			}
			for m, got := range snap.Forecast {
				w, err := weights(m)
				if err != nil {
					return err
				}
				var heads [2]float64
				for h, v := range snap.Advice[m] {
					control := f.World.Arms[s*3+h]
					if snapshot >= len(control.Snapshots) || control.Snapshots[snapshot].Tick != tick || len(control.Snapshots[snapshot].Forecast) != 150 || !studyClose(v, control.Snapshots[snapshot].Forecast[m]) || !studyClose(w[h], snap.Weights[m][h]) {
						return fmt.Errorf("snapshot expert or local weights")
					}
					heads[0] += gw[h] * v
					heads[1] += w[h] * v
				}
				for h := range heads {
					if !studyClose(heads[h], snap.Heads[m][h]) {
						return fmt.Errorf("snapshot head law")
					}
				}
				q := sw[0]*heads[0] + sw[1]*heads[1]
				if !studyClose(q, got) {
					return fmt.Errorf("independent snapshot law")
				}
			}
			snapshot++
		}
	}
	if position != 2400 || position != len(a.Receipts) || snapshot != len(a.Snapshots) || peak != a.PeakPending {
		return fmt.Errorf("complete tape coverage")
	}
	return nil
}

func TestHybridV48Study(t *testing.T) {
	input, out := os.Getenv("EVENTFRAME_HYBRID_V48_FIXTURE"), os.Getenv("EVENTFRAME_HYBRID_V48_OUT")
	if out == "" {
		t.Skip("explicit study output required")
	}
	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	d := json.NewDecoder(bufio.NewReader(f))
	b := bufio.NewWriter(w)
	e := json.NewEncoder(b)
	var manifest map[string]any
	if err = d.Decode(&manifest); err != nil {
		t.Fatal(err)
	}
	manifest["Kind"] = "hybrid_study_manifest"
	if err = e.Encode(manifest); err != nil {
		t.Fatal(err)
	}
	for {
		var fixture studyFixture
		if err = d.Decode(&fixture); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		r := hybridRecordV48{Seed: fixture.World.Population.Seed}
		for s := 0; s < 3; s++ {
			for _, mode := range hybridModesV48 {
				a, err := collectHybridV48(fixture, mode, s)
				if err != nil {
					t.Fatal(err)
				}
				r.Arms = append(r.Arms, a)
			}
		}
		if err = e.Encode(r); err != nil {
			t.Fatal(err)
		}
		t.Log(fixture.World.Population.Geometry + "/" + fixture.World.Population.Regime)
	}
	if err = b.Flush(); err != nil {
		t.Fatal(err)
	}
	if err = w.Sync(); err != nil {
		t.Fatal(err)
	}
}

func TestHybridV48StudyAudit(t *testing.T) {
	input, raw := os.Getenv("EVENTFRAME_HYBRID_V48_FIXTURE"), os.Getenv("EVENTFRAME_HYBRID_V48_AUDIT")
	if raw == "" {
		t.Skip("explicit audit input required")
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
	if err = rd.Decode(&rm); err != nil || rm["Kind"] != "hybrid_study_manifest" {
		t.Fatal("manifest kind", err)
	}
	rm["Kind"] = fm["Kind"]
	if !reflect.DeepEqual(fm, rm) {
		t.Fatal("manifest identity")
	}
	count := 0
	for {
		var fixture studyFixture
		if err = fd.Decode(&fixture); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		var r hybridRecordV48
		if err = rd.Decode(&r); err != nil {
			t.Fatal(err)
		}
		if r.Seed != fixture.World.Population.Seed || len(r.Arms) != 9 {
			t.Fatal("record identity")
		}
		for j, a := range r.Arms {
			if a.Mode != hybridModesV48[j%3] {
				t.Fatal("arm order")
			}
			if err = auditHybridV48(fixture, a, j/3); err != nil {
				t.Fatal(err)
			}
		}
		count++
		t.Log(fixture.World.Population.Geometry + "/" + fixture.World.Population.Regime)
	}
	if float64(count) != fm["Worlds"] {
		t.Fatal("world coverage")
	}
	var extra any
	if err = rd.Decode(&extra); err != io.EOF {
		t.Fatal("extra record", err)
	}
}

func TestHybridV48CostScreen(t *testing.T) {
	path := os.Getenv("EVENTFRAME_HYBRID_V48_COST")
	if path == "" {
		t.Skip("explicit cost artifact required")
	}
	base, cfg := benchmarkBase150(), Config{Prior: []float64{.8, .1, .1}, Hazard: 1. / 16, Trials: 2400, Pending: 2400}
	var maximum uint64
	for repeat := 0; repeat < 3; repeat++ {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		p, err := NewHybridPool(base, cfg, 16, 1./2400, 1./2400, 1)
		if err != nil {
			t.Fatal(err)
		}
		runtime.ReadMemStats(&after)
		runtime.KeepAlive(p)
		maximum = max(maximum, after.TotalAlloc-before.TotalAlloc)
	}
	loops := []map[string]any{}
	pass := maximum <= 8<<20
	f := testStudyFixture()
	for s := 0; s < 3; s++ {
		for _, mode := range hybridModesV48 {
			a, err := collectHybridV48(f, mode, s)
			if err != nil {
				t.Fatal(err)
			}
			loops = append(loops, map[string]any{"mode": mode, "schedule": studySchedules[s], "costs": a.Costs})
			pass = pass && a.Costs.ElapsedNS <= 400000000
		}
	}
	w, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	if err = json.NewEncoder(w).Encode(map[string]any{"constructorMaxBytes": maximum, "loops": loops, "preliminaryAllocationPass": maximum <= 8<<20, "preliminaryLoopPass": pass, "adoption": false}); err != nil {
		t.Fatal(err)
	}
	if err = w.Sync(); err != nil {
		t.Fatal(err)
	}
	// Timing FAIL is a valid scientific result, not concealed as execution FAIL.
	t.Log("allocation", maximum, "cost screen pass", pass)
}

func TestHybridV48FuturePrefix(t *testing.T) {
	for s := 0; s < 3; s++ {
		for _, mode := range hybridModesV48 {
			f, g := testStudyFixture(), testStudyFixture()
			for r := 8; r < 16; r++ {
				for m := 0; m < 150; m++ {
					g.World.Population.Outcomes[r][m] = !g.World.Population.Outcomes[r][m]
				}
			}
			a, err := collectHybridV48(f, mode, s)
			if err != nil {
				t.Fatal(err)
			}
			b, err := collectHybridV48(g, mode, s)
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

func TestHybridV48AuditRejectsSemanticCorruptions(t *testing.T) {
	f := testStudyFixture()
	a, err := collectHybridV48(f, "slow", 2)
	if err != nil {
		t.Fatal(err)
	}
	// Only an auditor boundary fixture, NOT independently audited scientific
	// controls. The full study uses the original generator/control exporter.
	f.World.Arms = make([]studyControl, 9)
	for h, mode := range []string{"full", "adaptive", "rich_moment2"} {
		c := studyControl{Mode: mode, Schedule: "uniform299", Issued: make([]float64, 2400)}
		for j := range c.Issued {
			c.Issued[j] = a.Advice[j][h]
		}
		for _, snap := range a.Snapshots {
			s := controlSnapshot{Tick: snap.Tick, Forecast: make([]float64, 150)}
			for m := range s.Forecast {
				s.Forecast[m] = snap.Advice[m][h]
			}
			c.Snapshots = append(c.Snapshots, s)
		}
		f.World.Arms[6+h] = c
	}
	if err = auditHybridV48(f, a, 2); err != nil {
		t.Fatal("positive auditor boundary", err)
	}
	mutations := []struct {
		name string
		f    func(*studyFixture, *hybridArmV48)
	}{
		{"original-advice", func(f *studyFixture, a *hybridArmV48) { a.Advice[0][0] += .1 }},
		{"original-global-head", func(f *studyFixture, a *hybridArmV48) { a.Heads[0][0] += .1 }},
		{"original-local-head", func(f *studyFixture, a *hybridArmV48) { a.Heads[0][1] += .1 }},
		{"issued-law", func(f *studyFixture, a *hybridArmV48) { a.Issued[0] += .1 }},
		{"epoch", func(f *studyFixture, a *hybridArmV48) { a.Receipts[0].Epoch++ }},
		{"arrival", func(f *studyFixture, a *hybridArmV48) { a.Receipts[0].ArrivedAt++ }},
		{"global-ordinal", func(f *studyFixture, a *hybridArmV48) { a.Receipts[0].TrialOrdinal++ }},
		{"local-ordinal", func(f *studyFixture, a *hybridArmV48) { a.Receipts[0].MemberOrdinal++ }},
		{"member", func(f *studyFixture, a *hybridArmV48) { a.Receipts[0].Member++ }},
		{"original-score", func(f *studyFixture, a *hybridArmV48) { a.Receipts[0].Forecast += .1 }},
		{"snapshot-law", func(f *studyFixture, a *hybridArmV48) { a.Snapshots[0].Forecast[0] += .1 }},
		{"snapshot-advice", func(f *studyFixture, a *hybridArmV48) { a.Snapshots[0].Advice[0][0] += .1 }},
		{"task-weights", func(f *studyFixture, a *hybridArmV48) { a.Snapshots[0].Weights[0][0] += .1 }},
		{"global-weights", func(f *studyFixture, a *hybridArmV48) { a.Snapshots[0].GlobalWeights[0] += .1 }},
		{"scope-weights", func(f *studyFixture, a *hybridArmV48) { a.Snapshots[0].ScopeWeights[0] += .1 }},
		{"snapshot-global-head", func(f *studyFixture, a *hybridArmV48) { a.Snapshots[0].Heads[0][0] += .1 }},
		{"snapshot-local-head", func(f *studyFixture, a *hybridArmV48) { a.Snapshots[0].Heads[0][1] += .1 }},
		{"missing-task", func(f *studyFixture, a *hybridArmV48) { a.Snapshots[0].Weights = a.Snapshots[0].Weights[:149] }},
		{"pending", func(f *studyFixture, a *hybridArmV48) { a.PeakPending++ }},
		{"cost", func(f *studyFixture, a *hybridArmV48) { a.Costs.AccountedNS++ }},
		{"negative-cost", func(f *studyFixture, a *hybridArmV48) { a.Costs.SetupNS = -1 }},
		{"label", func(f *studyFixture, a *hybridArmV48) {
			f.World.Population.Outcomes[0][0] = !f.World.Population.Outcomes[0][0]
		}},
		{"missing-control", func(f *studyFixture, a *hybridArmV48) { f.World.Arms = f.World.Arms[:8] }},
		{"missing-receipt", func(f *studyFixture, a *hybridArmV48) { a.Receipts = a.Receipts[:2399] }},
		{"future-arrival", func(f *studyFixture, a *hybridArmV48) { f.Due[2][0] = append(f.Due[2][0], 2399) }},
	}
	fb, _ := json.Marshal(f)
	ab, _ := json.Marshal(a)
	for _, m := range mutations {
		t.Run(m.name, func(t *testing.T) {
			var cf studyFixture
			var ca hybridArmV48
			if err := json.Unmarshal(fb, &cf); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(ab, &ca); err != nil {
				t.Fatal(err)
			}
			m.f(&cf, &ca)
			if err := auditHybridV48(cf, ca, 2); err == nil {
				t.Fatal("semantic corruption accepted")
			}
		})
	}
}
