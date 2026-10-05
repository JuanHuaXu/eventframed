package researchdispersion

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

var regimesV35 = append(append([]string(nil), regimesV34...), "off_grid", "asymmetric", "triple", "bounded_uniform")
var filesV35 = []string{"internal/researchdispersion/model.go", "internal/researchdispersion/model_test.go", "internal/researchdispersion/experiment_test.go", "internal/researchdispersion/shape.go", "internal/researchdispersion/shape_test.go", "internal/researchdispersion/shape_experiment_test.go", "docs/experiments/mmm-shape-v35-preflight.md", "docs/experiments/mmm-shape-v35-protocol.md"}

type shapeSnapshotV35 struct {
	snapV34
	Shapes [ShapeKernels]float64
}
type worldV35 struct {
	worldV34
	ShapeTrace     []float64
	ShapeSnapshots []shapeSnapshotV35
	ShapeElapsedNS int64
}

func makeV35(seed int64, g, r, id int) worldV35 {
	if r < len(regimesV34) {
		return worldV35{worldV34: makeV34(seed, g, r, id)}
	}
	w := worldV35{worldV34: makeV34(seed+int64((r-4)*1000000), g, 4, id)}
	w.Regime = regimesV35[r]
	truth, draw := rand.New(rand.NewSource(w.Seed+101)), rand.New(rand.NewSource(w.Seed+202))
	perm := truth.Perm(150)
	for i := range w.Rates {
		switch w.Regime {
		case "off_grid":
			w.Rates[i] = .27
			if perm[i] < 75 {
				w.Rates[i] = .73
			}
		case "asymmetric":
			w.Rates[i] = .1
			if perm[i] < 100 {
				w.Rates[i] = .8
			}
		case "triple":
			w.Rates[i] = []float64{.15, .5, .85}[perm[i]/50]
		case "bounded_uniform":
			w.Rates[i] = .15 + .7*truth.Float64()
		}
	}
	for r := range w.Outcomes {
		for i, p := range w.Rates {
			w.Outcomes[r][i] = draw.Float64() < p
		}
	}
	return w
}

func runShapeV35(w *worldV35) error {
	if len(w.Trace) < 1 || len(w.Trace) > 2400 {
		return fmt.Errorf("unsupported evidence tape")
	}
	start := time.Now()
	m, err := NewShape(w.Base)
	if err != nil {
		return err
	}
	cost := costV34{SetupNS: time.Since(start).Nanoseconds()}
	for step := range w.Trace {
		i, ordinal := w.Trace[step].Index, w.Trace[step].Ordinal
		s := time.Now()
		q, err := m.Predict(i)
		cost.ProbeNS += time.Since(s).Nanoseconds()
		if err != nil {
			return err
		}
		w.ShapeTrace = append(w.ShapeTrace, q)
		useful := w.Trace[step].Useful
		s = time.Now()
		if err := m.Observe(i, ordinal, useful); err != nil {
			return err
		}
		cost.UpdateNS += time.Since(s).Nanoseconds()
		for _, cut := range cutsV34 {
			if step+1 != cut {
				continue
			}
			s = time.Now()
			ss := shapeSnapshotV35{snapV34: snapV34{Model: "shape", Budget: cut, Forecast: make([]float64, 150)}}
			for i := range ss.Forecast {
				ss.Forecast[i], err = m.Predict(i)
				if err != nil {
					return err
				}
			}
			ss.Shapes = m.Shapes()
			ss.Costs = cost
			ss.Costs.SnapshotNS = time.Since(s).Nanoseconds()
			ss.Costs.AccountedNS = cost.SetupNS + cost.ProbeNS + cost.UpdateNS + ss.Costs.SnapshotNS + cost.EarlierSnapshotsNS
			cost.EarlierSnapshotsNS += ss.Costs.SnapshotNS
			w.ShapeSnapshots = append(w.ShapeSnapshots, ss)
		}
	}
	w.ShapeElapsedNS = time.Since(start).Nanoseconds()
	return nil
}

func TestExperimentV35(t *testing.T) {
	out, split := os.Getenv("EVENTFRAME_SHAPE_V35_OUT"), os.Getenv("EVENTFRAME_SHAPE_V35_SPLIT")
	if out == "" {
		t.Skip("explicit output required")
	}
	if !filepath.IsAbs(out) || split != "design" && split != "confirmation" {
		t.Fatal("invalid output/split")
	}
	seed := int64(2026103503)
	if split == "confirmation" {
		seed++
	}
	manifest := manifestV34{"manifest", split, seed, 1024, map[string]string{}}
	for _, p := range filesV35 {
		b, err := os.ReadFile(filepath.Join(rootV34(t), p))
		if err != nil {
			t.Fatal(err)
		}
		manifest.Sources[p] = hashV34(b)
	}
	f, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	enc := json.NewEncoder(writer)
	if err := enc.Encode(manifest); err != nil {
		t.Fatal(err)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesV35 {
			for id := 0; id < 32; id++ {
				w := makeV35(seed, g, r, id)
				if err := runV34(&w.worldV34, 2400); err != nil {
					t.Fatal(err)
				}
				if err := runShapeV35(&w); err != nil {
					t.Fatal(err)
				}
				for j := range w.Snapshots {
					scoreV34(&w.Snapshots[j], w.Rates)
				}
				for j := range w.ShapeSnapshots {
					scoreV34(&w.ShapeSnapshots[j].snapV34, w.Rates)
				}
				if err := enc.Encode(w); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if err := writer.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}

func clearV35(w *worldV35) {
	clearV34(&w.worldV34)
	w.ShapeElapsedNS = 0
	for j := range w.ShapeSnapshots {
		w.ShapeSnapshots[j].Costs = costV34{}
	}
}

func TestShapeCollectorNoFutureV35(t *testing.T) {
	for _, regime := range []int{3, 4, 12, 13, 14, 15} {
		full := makeV35(2026103599, 1, regime, 0)
		if err := runV34(&full.worldV34, 2400); err != nil {
			t.Fatal(err)
		}
		if err := runShapeV35(&full); err != nil {
			t.Fatal(err)
		}
		clearV35(&full)
		for _, cut := range cutsV34[:4] {
			v := makeV35(2026103599, 1, regime, 0)
			seen := map[[2]int]bool{}
			for _, tr := range full.Trace[:cut] {
				seen[[2]int{tr.Ordinal - 1, tr.Index}] = true
			}
			for r, row := range v.Outcomes {
				for i := range row {
					if !seen[[2]int{r, i}] {
						v.Outcomes[r][i] = !row[i]
					}
				}
			}
			if err := runV34(&v.worldV34, cut); err != nil {
				t.Fatal(err)
			}
			if err := runShapeV35(&v); err != nil {
				t.Fatal(err)
			}
			clearV35(&v)
			if !reflect.DeepEqual(v.Trace, full.Trace[:cut]) || !reflect.DeepEqual(v.Snapshots, full.Snapshots[:len(v.Snapshots)]) || !reflect.DeepEqual(v.ShapeTrace, full.ShapeTrace[:cut]) || !reflect.DeepEqual(v.ShapeSnapshots, full.ShapeSnapshots[:len(v.ShapeSnapshots)]) {
				t.Fatal("unarrived trials changed earlier record", regime, cut)
			}
		}
	}
}
