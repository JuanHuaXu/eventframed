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
)

func sameMemoArmV49(a, b hybridArmV48) bool {
	a.Costs, b.Costs = studyCosts{}, studyCosts{}
	return reflect.DeepEqual(a, b)
}

func validMemoCostsV49(c studyCosts) bool {
	return c.SetupNS > 0 && c.IssueNS > 0 && c.ResolveNS > 0 && c.SnapshotNS > 0 && c.AccountedNS == c.SetupNS+c.IssueNS+c.ResolveNS+c.SnapshotNS && c.ElapsedNS >= c.AccountedNS
}

func TestMemoV49FuturePrefix(t *testing.T) {
	for s := 0; s < 3; s++ {
		for _, mode := range hybridModesV48 {
			f, g := testStudyFixture(), testStudyFixture()
			for r := 8; r < 16; r++ {
				for m := 0; m < 150; m++ {
					g.World.Population.Outcomes[r][m] = !g.World.Population.Outcomes[r][m]
				}
			}
			a, e := collectMemoV49(f, mode, s)
			if e != nil {
				t.Fatal(e)
			}
			b, e := collectMemoV49(g, mode, s)
			if e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(a.Issued[:1201], b.Issued[:1201]) || !reflect.DeepEqual(a.Advice[:1201], b.Advice[:1201]) || !reflect.DeepEqual(a.Heads[:1201], b.Heads[:1201]) {
				t.Fatal("future altered prefix", mode, s)
			}
			for j, snap := range a.Snapshots {
				if snap.Tick < 1200 && !reflect.DeepEqual(snap, b.Snapshots[j]) {
					t.Fatal("future altered snapshot", mode, s)
				}
			}
		}
	}
}

func TestMemoV49HybridOwnership(t *testing.T) {
	cfg := Config{Prior: []float64{.8, .1, .1}, Hazard: 1. / 16, Trials: 12, Pending: 12}
	p, e := NewMemoHybridV49([]float64{.25, .5, .925}, cfg, 4, 1./2400, 1./150, 1)
	if e != nil {
		t.Fatal(e)
	}
	other, e := NewHybridPool([]float64{.25, .5, .925}, cfg, 4, 1./2400, 1./150, 1)
	if e != nil {
		t.Fatal(e)
	}
	x, e := p.Issue(0, 1)
	if e != nil {
		t.Fatal(e)
	}
	y, e := other.Issue(0, 1)
	if e != nil {
		t.Fatal(e)
	}
	if x.Forecast() != y.Forecast() {
		t.Fatal("forecast differs")
	}
	if _, e = p.Resolve(y, true, 2); e == nil {
		t.Fatal("foreign owner accepted")
	}
	if e = p.BeginEpoch(2, 3); e != nil {
		t.Fatal(e)
	}
	if e = other.BeginEpoch(2, 3); e != nil {
		t.Fatal(e)
	}
	if _, e = p.Resolve(x, true, 4); e == nil {
		t.Fatal("stale epoch accepted")
	}
	for m := 0; m < 3; m++ {
		a, e := p.Predict(m)
		if e != nil {
			t.Fatal(e)
		}
		b, e := other.Predict(m)
		if e != nil || a != b {
			t.Fatal("epoch differs", e)
		}
	}
}

func TestMemoV49Cost(t *testing.T) {
	path := os.Getenv("EVENTFRAME_MEMO_V49_COST")
	if path == "" {
		t.Skip("explicit cost artifact required")
	}
	base, cfg := benchmarkBase150(), Config{Prior: []float64{.8, .1, .1}, Hazard: 1. / 16, Trials: 2400, Pending: 2400}
	alloc := map[string]uint64{}
	for _, name := range []string{"original", "memo"} {
		for r := 0; r < 3; r++ {
			runtime.GC()
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			var p *HybridPool
			var e error
			if name == "original" {
				p, e = NewHybridPool(base, cfg, 16, 1./2400, 1./2400, 1)
			} else {
				p, e = NewMemoHybridV49(base, cfg, 16, 1./2400, 1./2400, 1)
			}
			if e != nil {
				t.Fatal(e)
			}
			runtime.ReadMemStats(&after)
			runtime.KeepAlive(p)
			alloc[name] = max(alloc[name], after.TotalAlloc-before.TotalAlloc)
		}
	}
	var loops []map[string]any
	f := testStudyFixture()
	for s := 0; s < 3; s++ {
		for j, mode := range hybridModesV48 {
			var a, b hybridArmV48
			var e error
			if (s*3+j)%2 == 0 {
				a, e = collectHybridV48(f, mode, s)
				if e == nil {
					b, e = collectMemoV49(f, mode, s)
				}
			} else {
				b, e = collectMemoV49(f, mode, s)
				if e == nil {
					a, e = collectHybridV48(f, mode, s)
				}
			}
			if e != nil {
				t.Fatal(e)
			}
			if !sameMemoArmV49(a, b) || !validMemoCostsV49(a.Costs) || !validMemoCostsV49(b.Costs) {
				t.Fatal("paired equivalence/cost failed")
			}
			loops = append(loops, map[string]any{"mode": mode, "schedule": studySchedules[s], "original": a.Costs, "memo": b.Costs, "memoFirst": (s*3+j)%2 != 0})
		}
	}
	w, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer w.Close()
	if e = json.NewEncoder(w).Encode(map[string]any{"constructorMaxBytes": alloc, "loops": loops, "exactPairEquality": true, "adoption": false}); e != nil {
		t.Fatal(e)
	}
	if e = w.Sync(); e != nil {
		t.Fatal(e)
	}
}

func TestMemoV49Study(t *testing.T) {
	root := os.Getenv("EVENTFRAME_MEMO_V49_ROOT")
	if root == "" {
		t.Skip("explicit ablation root required")
	}
	input, e := os.Open(os.Getenv("EVENTFRAME_MEMO_V49_FIXTURE"))
	if e != nil {
		t.Fatal(e)
	}
	defer input.Close()
	ref, e := os.Open(os.Getenv("EVENTFRAME_MEMO_V49_REFERENCE"))
	if e != nil {
		t.Fatal(e)
	}
	defer ref.Close()
	fd, rd := json.NewDecoder(bufio.NewReader(input)), json.NewDecoder(bufio.NewReader(ref))
	rd.DisallowUnknownFields()
	var fm, rm map[string]any
	if e = fd.Decode(&fm); e != nil {
		t.Fatal(e)
	}
	if e = rd.Decode(&rm); e != nil {
		t.Fatal(e)
	}
	fm["Kind"] = "hybrid_study_manifest"
	if !reflect.DeepEqual(fm, rm) {
		t.Fatal("reference manifest differs")
	}
	var files [2]*os.File
	var buffers [2]*bufio.Writer
	var enc [2]*json.Encoder
	for i, name := range []string{"original", "memo"} {
		files[i], e = os.OpenFile(root+"/"+name+".jsonl", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal(e)
		}
		defer files[i].Close()
		buffers[i] = bufio.NewWriter(files[i])
		enc[i] = json.NewEncoder(buffers[i])
		if e = enc[i].Encode(fm); e != nil {
			t.Fatal(e)
		}
	}
	count := 0
	for {
		var f studyFixture
		e = fd.Decode(&f)
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		var old hybridRecordV48
		if e = rd.Decode(&old); e != nil {
			t.Fatal(e)
		}
		if old.Seed != f.World.Population.Seed || len(old.Arms) != 9 {
			t.Fatal("reference identity")
		}
		a, b := hybridRecordV48{Seed: old.Seed}, hybridRecordV48{Seed: old.Seed}
		for j, x := range old.Arms {
			if x.Mode != hybridModesV48[j%3] || x.Schedule != studySchedules[j/3] {
				t.Fatal("arm order")
			}
			var original, memo hybridArmV48
			if (count*9+j)%2 == 0 {
				original, e = collectHybridV48(f, x.Mode, j/3)
				if e == nil {
					memo, e = collectMemoV49(f, x.Mode, j/3)
				}
			} else {
				memo, e = collectMemoV49(f, x.Mode, j/3)
				if e == nil {
					original, e = collectHybridV48(f, x.Mode, j/3)
				}
			}
			if e != nil {
				t.Fatal(e)
			}
			if !sameMemoArmV49(original, x) || !sameMemoArmV49(memo, x) || !validMemoCostsV49(original.Costs) || !validMemoCostsV49(memo.Costs) {
				t.Fatal("exact reference equivalence/cost failed", count, j)
			}
			a.Arms = append(a.Arms, original)
			b.Arms = append(b.Arms, memo)
		}
		if e = enc[0].Encode(a); e != nil {
			t.Fatal(e)
		}
		if e = enc[1].Encode(b); e != nil {
			t.Fatal(e)
		}
		count++
		t.Log(f.World.Population.Geometry + "/" + f.World.Population.Regime)
	}
	if float64(count) != fm["Worlds"] || count != 28 {
		t.Fatal("diagnostic coverage", count)
	}
	var extra any
	if e = rd.Decode(&extra); e != io.EOF {
		t.Fatal("extra reference", e)
	}
	for i := range files {
		if e = buffers[i].Flush(); e != nil {
			t.Fatal(e)
		}
		if e = files[i].Sync(); e != nil {
			t.Fatal(e)
		}
	}
	t.Log(fmt.Sprintf("exact prior/forecast/receipt/snapshot equivalence: %d worlds, %d paired arms", count, count*9))
}
