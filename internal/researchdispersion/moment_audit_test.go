package researchdispersion

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/researchmomentref"
)

func independentMomentV41(w windowWorldV37, arm orientationArmV38) error {
	a := arm.windowArmV37
	cfg, e := configMomentV41(a.Mode)
	if e != nil {
		return e
	}
	ref, e := researchmomentref.New(w.Base, cfg)
	if e != nil {
		return e
	}
	if arm.ReadyAt != -1 || len(arm.Template) != 0 {
		return fmt.Errorf("moment claims frozen template")
	}
	due, e := scheduleV36(delayedWorldV36{Seed: w.Seed, Outcomes: w.Outcomes}, a.Schedule)
	if e != nil {
		return e
	}
	last := len(due) - 1
	for len(due[last]) == 0 {
		last--
	}
	position, snapshot, peak := 0, 0, 0
	for tick := 0; tick <= last; tick++ {
		if tick < 2400 {
			q, e := ref.Predict(tick % 150)
			if e != nil || !closeV36(q, a.Issued[tick]) || a.ExpertIssued[tick] != ([4]float64{}) {
				return fmt.Errorf("independent moment issued law %d", tick)
			}
			if e = ref.Issue(tick % 150); e != nil {
				return e
			}
			peak = max(peak, tick+1-position)
		}
		for _, trial := range due[tick] {
			if position >= len(a.Receipts) {
				return fmt.Errorf("missing moment receipt")
			}
			i, j := trial%150, trial/150
			r := a.Receipts[position]
			if r.Member != i || r.TrialOrdinal != j+1 || r.Epoch != 1 || r.IssuedAt != int64(trial) || r.ArrivedAt != int64(tick) || r.Forecast != a.Issued[trial] || r.Useful != w.Outcomes[j][i] {
				return fmt.Errorf("moment original receipt identity")
			}
			if e = ref.Resolve(i, j+1, r.Useful); e != nil {
				return e
			}
			position++
		}
		if (tick < 2400 && tick%150 == 149) || tick == last {
			if snapshot >= len(a.Snapshots) {
				return fmt.Errorf("missing moment snapshot")
			}
			s := a.Snapshots[snapshot]
			if s.Tick != tick || s.Arrived != position || s.Pending != min(tick+1, 2400)-position || len(s.Forecast) != 150 || s.Weights != ([4]float64{}) {
				return fmt.Errorf("moment as-of coverage")
			}
			for i, p := range s.Forecast {
				q, e := ref.Predict(i)
				if e != nil || !closeV36(q, p) {
					return fmt.Errorf("independent moment batch forecast %d", i)
				}
			}
			if e = independentMetricsV38(w, s); e != nil {
				return e
			}
			snapshot++
		}
	}
	if position != 2400 || snapshot != len(a.Snapshots) || peak != a.PeakPending {
		return fmt.Errorf("moment drain coverage")
	}
	return nil
}
func auditMomentV41(w orientationWorldV38, seed int64, worlds int) error {
	g, r := -1, -1
	for i, name := range []string{"tight", "wide"} {
		if w.Population.Geometry == name {
			g = i
		}
	}
	for i, name := range regimesV39 {
		if w.Population.Regime == name {
			r = i
		}
	}
	if g < 0 || r < 0 || w.Population.World < 0 || w.Population.World >= worlds || len(w.Arms) != 33 || !reflect.DeepEqual(w.Population, makeV39(seed, g, r, w.Population.World)) {
		return fmt.Errorf("moment population/domain")
	}
	for j, arm := range w.Arms {
		a := arm.windowArmV37
		if len(a.Issued) != 2400 || len(a.ExpertIssued) != 2400 || a.Mode != modesMomentV41[j%11] || a.Schedule != schedulesV39[j/11] {
			return fmt.Errorf("moment arm shape/order")
		}
		c := a.Costs
		if c.SetupNS < 0 || c.ScheduleNS < 0 || c.IssueNS < 0 || c.ResolveNS < 0 || c.SnapshotNS < 0 || c.AccountedNS != c.SetupNS+c.IssueNS+c.ResolveNS+c.SnapshotNS || c.ElapsedNS < c.AccountedNS+c.ScheduleNS {
			return fmt.Errorf("moment time accounting")
		}
		replay, e := runMomentV41(w.Population, a.Mode, a.Schedule)
		if e != nil {
			return e
		}
		x, y := arm, replay
		x.Costs, y.Costs = delayedCostsV36{}, delayedCostsV36{}
		scoreWindowV37(w.Population, &y.windowArmV37)
		if !reflect.DeepEqual(x, y) {
			return fmt.Errorf("moment auxiliary full arithmetic replay")
		}
		if j%11 < 2 {
			e = independentOrientationV38(w.Population, arm)
		} else if j%11 == 2 {
			control := arm
			control.Mode = "raw2_shared"
			e = independentPriorV40(w.Population, control)
		} else {
			e = independentMomentV41(w.Population, arm)
		}
		if e != nil {
			return e
		}
		risk, priority := 0., 0.
		for trial, q := range a.Issued {
			p := w.Population.Rates[trial/150][trial%150]
			weight := 1.
			if trial%150 < 10 {
				weight = 3
			}
			loss := q*q - 2*q*p + p
			risk += loss / 2400
			priority += weight * loss / (16 * 170)
		}
		if !closeV36(risk, a.IssuedBrier) || !closeV36(priority, a.IssuedPriority) {
			return fmt.Errorf("moment independent issued risk")
		}
		recovery := 0.
		for phase, start := range w.Population.Changes {
			end := 16
			if phase+1 < len(w.Population.Changes) {
				end = w.Population.Changes[phase+1]
			}
			delay := end - start + 1
			for round := start + 1; round < end; round++ {
				one, two := windowAtV37(a, (round-1)*150+149), windowAtV37(a, round*150+149)
				if one.Brier <= .20 && two.Brier <= .20 && one.PacketUsefulness >= .75 && two.PacketUsefulness >= .75 {
					delay = round - start + 1
					break
				}
			}
			recovery += float64(delay) / float64(len(w.Population.Changes))
		}
		if !closeV36(recovery, a.Recovery) {
			return fmt.Errorf("moment independent phase recovery")
		}
	}
	return nil
}
func TestMomentStudyAuditV41(t *testing.T) {
	input := os.Getenv("EVENTFRAME_MOMENT_V41_AUDIT")
	if input == "" {
		t.Skip("explicit audit required")
	}
	f, e := os.Open(input)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	hash := sha256.New()
	d := json.NewDecoder(bufio.NewReader(io.TeeReader(f, hash)))
	d.DisallowUnknownFields()
	var manifest manifestV34
	if e = d.Decode(&manifest); e != nil {
		t.Fatal(e)
	}
	seed, worlds, e := momentCohortV41(manifest.Split)
	if e != nil || manifest.Kind != "manifest" || manifest.SeedBase != seed || manifest.Worlds != 28*worlds {
		t.Fatal("moment manifest", e)
	}
	if len(manifest.Sources) != len(filesMomentV41) {
		t.Fatal("moment source count")
	}
	for _, p := range filesMomentV41 {
		b, e := os.ReadFile(filepath.Join(rootV34(t), p))
		if e != nil || hashV34(b) != manifest.Sources[p] {
			t.Fatal("moment source changed", p, e)
		}
	}
	groups := map[string][]orientationWorldV38{}
	seen := map[int64]bool{}
	count, snapshots := 0, 0
	for {
		var w orientationWorldV38
		e = d.Decode(&w)
		if e == io.EOF {
			break
		}
		if e != nil || seen[w.Population.Seed] {
			t.Fatal("moment decode/duplicate", e)
		}
		if e = auditMomentV41(w, seed, worlds); e != nil {
			t.Fatal(w.Population.Geometry, w.Population.Regime, w.Population.World, e)
		}
		seen[w.Population.Seed] = true
		count++
		for _, a := range w.Arms {
			snapshots += len(a.Snapshots)
		}
		key := w.Population.Geometry + "/" + w.Population.Regime
		groups[key] = append(groups[key], w)
	}
	if count != 28*worlds || len(groups) != 28 || snapshots != count*550 {
		t.Fatal("moment total coverage", count, snapshots)
	}
	benchmark, e := os.ReadFile(os.Getenv("EVENTFRAME_MOMENT_V41_BENCHMARK"))
	if e != nil {
		t.Fatal(e)
	}
	allocation, e := allocationMomentV41(benchmark)
	if e != nil {
		t.Fatal(e)
	}
	result := reportMomentV41(groups, worlds, allocation)
	result["Study"], result["Split"], result["SHA256"], result["Sources"] = "moment-v41", manifest.Split, hex.EncodeToString(hash.Sum(nil)), manifest.Sources
	result["BenchmarkSHA256"] = hashV34(benchmark)
	result["Worlds"], result["Arms"], result["Snapshots"], result["DistinctTrials"] = count, count*33, snapshots, count*2400
	result["WholeGoalsComplete"], result["Goal7EqualTotalCost"] = false, false
	b, e := json.MarshalIndent(result, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	out, e := os.OpenFile(strings.TrimSuffix(input, ".jsonl")+"-audit.json", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer out.Close()
	if _, e = out.Write(append(b, '\n')); e != nil {
		t.Fatal(e)
	}
	if e = out.Sync(); e != nil {
		t.Fatal(e)
	}
	t.Logf("verified %d moment worlds/%d snapshots", count, snapshots)
}
func TestMomentAuditorNegativeV41(t *testing.T) {
	seed := int64(740007)
	w := orientationWorldV38{Population: makeV39(seed, 1, 13, 0)}
	for _, schedule := range schedulesV39 {
		for _, mode := range modesMomentV41 {
			a, e := runMomentV41(w.Population, mode, schedule)
			if e != nil {
				t.Fatal(e)
			}
			scoreWindowV37(w.Population, &a.windowArmV37)
			w.Arms = append(w.Arms, a)
		}
	}
	if e := auditMomentV41(w, seed, 1); e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(w)
	mutations := []func(*orientationWorldV38){func(x *orientationWorldV38) { x.Population.Rates[0][0] += .01 }, func(x *orientationWorldV38) { x.Arms[3].Issued[1000] += .01 }, func(x *orientationWorldV38) { x.Arms[4].Receipts[1000].TrialOrdinal++ }, func(x *orientationWorldV38) { x.Arms[5].Snapshots[8].Forecast[0] += .01 }, func(x *orientationWorldV38) { x.Arms[6].ExpertIssued[0][0] = .5 }, func(x *orientationWorldV38) { x.Arms[7].IssuedPriority += .01 }, func(x *orientationWorldV38) { x.Arms[8].Recovery++ }, func(x *orientationWorldV38) { x.Arms[9].Costs.AccountedNS++ }, func(x *orientationWorldV38) { x.Arms[10].PeakPending++ }, func(x *orientationWorldV38) { x.Arms[3].ReadyAt = 100 }, func(x *orientationWorldV38) { x.Arms[4].Schedule = "fixed150" }, func(x *orientationWorldV38) { x.Arms[5].Receipts[1] = x.Arms[5].Receipts[0] }, func(x *orientationWorldV38) { x.Arms[3].Mode = "rich_moment4" }}
	for i, mutate := range mutations {
		var x orientationWorldV38
		if e := json.Unmarshal(raw, &x); e != nil {
			t.Fatal(e)
		}
		mutate(&x)
		if reflect.DeepEqual(w, x) {
			t.Fatal("no-op corruption", i)
		}
		if e := auditMomentV41(x, seed, 1); e == nil {
			t.Fatal("accepted corruption", i)
		}
	}
	if _, e := allocationMomentV41([]byte("incomplete")); e == nil {
		t.Fatal("missing allocation accepted")
	}
	t.Log("13 nonidentity corruptions and missing allocation rejected")
}
