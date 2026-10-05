package researchblend

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchcalibration"
	"github.com/JuanHuaXu/eventframed/internal/researchpartition"
)

var budgetsV33 = []int{16, 32, 64, 128, 150}
var modelsV33 = []string{"local", "old", "partition", "blend", "stack", "crossfit"}

type nominationV33 struct {
	seen  []bool
	order []int
	count int
	rng   *rand.Rand
}

func newNominationV33(n int, seed int64) (*nominationV33, error) {
	if n < 2 || n > 200 {
		return nil, errors.New("unsupported nomination frontier")
	}
	m := &nominationV33{seen: make([]bool, n), rng: rand.New(rand.NewSource(seed + 303))}
	buckets, size, bits := min(32, n), 1, 0
	for size < buckets {
		size *= 2
		bits++
	}
	for k := 0; k < size; k++ {
		j, x := 0, k
		for bit := 0; bit < bits; bit++ {
			j = j*2 + (x & 1)
			x >>= 1
		}
		if j < buckets {
			m.order = append(m.order, j)
		}
	}
	return m, nil
}

func (m *nominationV33) next() (Selection, error) {
	if m.count == len(m.seen) {
		return Selection{Index: -1}, errors.New("nomination exhausted")
	}
	for attempt := 0; attempt < len(m.order); attempt++ {
		bucket := m.order[(m.count+attempt)%len(m.order)]
		lo, hi := bucket*len(m.seen)/len(m.order), (bucket+1)*len(m.seen)/len(m.order)
		left := 0
		for i := lo; i < hi; i++ {
			if !m.seen[i] {
				left++
			}
		}
		if left == 0 {
			continue
		}
		ordinal := m.rng.Intn(left)
		for i := lo; i < hi; i++ {
			if !m.seen[i] {
				if ordinal == 0 {
					m.seen[i] = true
					m.count++
					return Selection{Index: i, Probability: 1 / float64(left)}, nil
				}
				ordinal--
			}
		}
	}
	panic("inconsistent nomination state")
}

type costsV33 struct {
	SetupNS, ProbeNS, UpdateNS, SnapshotNS, EarlierSnapshotsNS, AccountedNS int64
}
type traceV33 struct {
	Index       int
	Useful      bool
	Probability float64
	Q           [6]float64
	Issued      [3]float64
}
type snapshotV33 struct {
	Model                                                           string
	Budget                                                          int
	Forecast                                                        []float64
	Weights                                                         [3]float64
	LOO                                                             [][3]float64
	Packet                                                          []int
	Brier, PriorityBrier, PacketUsefulness, PacketBrier, PacketBias float64
	ObservedBrier, UnobservedBrier                                  *float64
	NominationNS                                                    int64
	Costs                                                           costsV33
}
type curveV33 struct {
	Trace                   []traceV33
	Snapshots               []snapshotV33
	NominationNS, ElapsedNS int64
}
type worldV33 struct {
	Kind, Geometry, Regime              string
	World, Partition                    int
	Seed                                int64
	Base, Coordinates, Rates, LeafMeans []float64
	Labels                              []bool
	InputNS                             int64
	curveV33
}

func runCurveV33(base, coordinate []float64, labels []bool, seed int64, stop int) curveV33 {
	return runCurveWithCheckpointsV33(base, coordinate, labels, seed, stop, budgetsV33)
}

func runCurveWithCheckpointsV33(base, coordinate []float64, labels []bool, seed int64, stop int, cuts []int) curveV33 {
	if len(base) != 150 || len(coordinate) != 150 || len(labels) != 150 || stop < 1 || stop > 150 {
		panic("unsupported curve inputs")
	}
	start := time.Now()
	result := curveV33{Trace: make([]traceV33, 0, stop), Snapshots: make([]snapshotV33, 0, 30)}
	var p [6]predictorV30
	var costs [6]costsV33
	var blend *Model
	var stack *Stack
	var crossfit *Crossfit
	var err error
	var baselineSnapshotsNS int64
	for k := range p {
		s := time.Now()
		switch k {
		case 0:
			p[k] = &localV30{base: append([]float64(nil), base...), seen: make([]bool, 150), useful: make([]bool, 150)}
		case 1:
			p[k], err = researchcalibration.New(base)
		case 2:
			p[k], err = researchpartition.New(base, coordinate)
		case 3:
			blend, err = New(base, coordinate)
			p[k] = blend
		case 4:
			stack, err = NewStack(base, coordinate)
			p[k] = stack
		case 5:
			crossfit, err = NewCrossfit(base, coordinate)
			p[k] = crossfit
		}
		if err != nil {
			panic(err)
		}
		costs[k].SetupNS = time.Since(s).Nanoseconds()
	}
	nominationStart := time.Now()
	nominee, err := newNominationV33(150, seed)
	if err != nil {
		panic(err)
	}
	result.NominationNS = time.Since(nominationStart).Nanoseconds()
	for step := 0; step < stop; step++ {
		s := time.Now()
		choice, e := nominee.next()
		if e != nil {
			panic(e)
		}
		result.NominationNS += time.Since(s).Nanoseconds()
		row := traceV33{Index: choice.Index, Probability: choice.Probability}
		var ticket Ticket
		for k := range p {
			s = time.Now()
			if k == 4 {
				ticket, e = stack.Issue(choice.Index)
				row.Q[k] = ticket.Forecast()
				row.Issued = stack.pending[choice.Index].p
			} else {
				row.Q[k], e = p[k].Predict(choice.Index)
			}
			if e != nil {
				panic(e)
			}
			costs[k].ProbeNS += time.Since(s).Nanoseconds()
		}
		// All six laws and the original stack row precede access to this label.
		row.Useful = labels[choice.Index]
		result.Trace = append(result.Trace, row)
		for k := range p {
			s = time.Now()
			if k == 4 {
				e = stack.Resolve(ticket, row.Useful)
			} else {
				e = p[k].Observe(choice.Index, row.Useful)
			}
			if e != nil {
				panic(e)
			}
			costs[k].UpdateNS += time.Since(s).Nanoseconds()
		}
		for _, budget := range cuts {
			if budget != step+1 {
				continue
			}
			for k, name := range modelsV33 {
				s = time.Now()
				snap := snapshotV33{Model: name, Budget: budget, Forecast: make([]float64, 150), LOO: make([][3]float64, 0), NominationNS: result.NominationNS}
				for i := range base {
					snap.Forecast[i], e = p[k].Predict(i)
					if e != nil {
						panic(e)
					}
				}
				switch k {
				case 3:
					snap.Weights = blend.Weights()
				case 4:
					snap.Weights = stack.Weights()
				case 5:
					snap.Weights = crossfit.Weights()
					for _, v := range result.Trace {
						x, e := crossfit.trainingRow(v.Index)
						if e != nil {
							panic(e)
						}
						snap.LOO = append(snap.LOO, x)
					}
				}
				snap.Costs = costs[k]
				snap.Costs.SnapshotNS = time.Since(s).Nanoseconds()
				snap.Costs.AccountedNS = snap.Costs.SetupNS + snap.Costs.ProbeNS + snap.Costs.UpdateNS + snap.Costs.SnapshotNS + snap.Costs.EarlierSnapshotsNS
				costs[k].EarlierSnapshotsNS += snap.Costs.SnapshotNS
				result.Snapshots = append(result.Snapshots, snap)
			}
			s = time.Now()
			baseline := snapshotV33{Model: "baseline", Budget: budget, Forecast: append([]float64(nil), base...), LOO: make([][3]float64, 0), NominationNS: result.NominationNS}
			baseline.Costs.SnapshotNS = time.Since(s).Nanoseconds()
			baseline.Costs.EarlierSnapshotsNS = baselineSnapshotsNS
			baseline.Costs.AccountedNS = baseline.Costs.SnapshotNS + baselineSnapshotsNS
			baselineSnapshotsNS += baseline.Costs.SnapshotNS
			result.Snapshots = append(result.Snapshots, baseline)
		}
	}
	result.ElapsedNS = time.Since(start).Nanoseconds()
	return result
}

func scoreCurveV33(w *worldV33) {
	for k := range w.Snapshots {
		s := &w.Snapshots[k]
		a := armV30{Forecast: s.Forecast}
		scoreV30(&a, w.Rates)
		s.Packet, s.Brier, s.PriorityBrier, s.PacketUsefulness, s.PacketBrier, s.PacketBias = a.Packet, a.Brier, a.PriorityBrier, a.PacketUsefulness, a.PacketBrier, a.PacketBias
		seen := make([]bool, 150)
		for _, row := range w.Trace[:s.Budget] {
			seen[row.Index] = true
		}
		observed, unobserved := 0., 0.
		for i, p := range w.Rates {
			q := s.Forecast[i]
			risk := (q-p)*(q-p) + p*(1-p)
			if seen[i] {
				observed += risk / float64(s.Budget)
			} else {
				unobserved += risk / float64(150-s.Budget)
			}
		}
		s.ObservedBrier = &observed
		if s.Budget < 150 {
			s.UnobservedBrier = &unobserved
		}
	}
}

func makeWorldV33(seedBase int64, g, r, w int) worldV33 {
	s := time.Now()
	// Frozen input synthesis also runs discarded evaluator-side legacy arms.
	// The offset gives the new protocol a nonoverlapping geometry stride.
	old := makeWorldV30(seedBase+int64(g)*90000000, g, r, w)
	result := worldV33{Kind: old.Kind, Geometry: old.Geometry, Regime: old.Regime, World: old.World, Partition: old.Partition, Seed: old.Seed, Base: old.Base, Coordinates: old.Coordinates, Rates: old.Rates, LeafMeans: old.LeafMeans, Labels: old.Labels, InputNS: time.Since(s).Nanoseconds()}
	result.curveV33 = runCurveV33(result.Base, result.Coordinates, result.Labels, result.Seed, 150)
	scoreCurveV33(&result)
	return result
}

var sourcesV33 = []string{"internal/researchblend/model.go", "internal/researchblend/model_test.go", "internal/researchblend/experiment_test.go", "internal/researchblend/stack.go", "internal/researchblend/stack_test.go", "internal/researchblend/crossfit.go", "internal/researchblend/crossfit_test.go", "internal/researchcalibration/model.go", "internal/researchcalibration/loo.go", "internal/researchcalibration/loo_test.go", "internal/researchpartition/model.go", "internal/researchpartition/loo.go", "internal/researchpartition/loo_test.go", "internal/researchblend/curve_experiment_test.go", "internal/researchblend/curve_test.go", "research/curve-v33-math.mjs", "research/curve-v33-verify.mjs", "docs/experiments/mmm-curve-v33-preflight.md", "docs/experiments/mmm-curve-v33-protocol.md"}

func TestExperimentV33(t *testing.T) {
	out := os.Getenv("EVENTFRAME_CURVE_V33_OUT")
	if out == "" {
		t.Skip("opt-in fixed-model label-budget experiment")
	}
	split := os.Getenv("EVENTFRAME_CURVE_V33_SPLIT")
	seed := int64(2026103303)
	if split == "confirmation" {
		seed = 2026103304
	} else if split != "design" {
		t.Fatal("unknown split")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := map[string]string{}
	for _, source := range sourcesV33 {
		b, err := os.ReadFile(filepath.Join(root, source))
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		hashes[source] = hex.EncodeToString(h[:])
	}
	f, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err = enc.Encode(map[string]any{"Kind": "manifest", "Split": split, "SeedBase": seed, "Worlds": 768, "Sources": hashes}); err != nil {
		t.Fatal(err)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesV30 {
			for w := 0; w < 32; w++ {
				if err = enc.Encode(makeWorldV33(seed, g, r, w)); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if err = f.Sync(); err != nil {
		t.Fatal(err)
	}
}
