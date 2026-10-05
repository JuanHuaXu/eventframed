package researchdispersion

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	noise "github.com/JuanHuaXu/eventframed/internal/researchnoisemoment"
	nref "github.com/JuanHuaXu/eventframed/internal/researchnoisemomentref"
)

var regimesNoiseV53 = append(append([]string(nil), regimesV39...), "stationary_noise20", "partial_noise20", "stationary_round_noise10", "unrelated_noise10")
var modesNoiseV53 = []string{"full", "adaptive", "rich_moment2", "eta0", "eta10", "eta20", "learned_eta"}

type noiseArmV53 struct {
	orientationArmV38
	ObservedIssued    []float64
	NoiseIssued       [][3]float64
	ObservedSnapshots [][]float64
	NoiseSnapshots    [][3]float64
}
type noiseWorldV53 struct {
	Population windowWorldV37
	Arms       []noiseArmV53
}

func makeNoiseV53(seed int64, g, r, id int) windowWorldV37 {
	if r < 14 {
		return makeV39(seed, g, r, id)
	}
	old := 1
	if r == 15 {
		old = 8
	}
	if r == 17 {
		old = 9
	}
	// Disjoint namespaces for extra stress cases; actual seed uniqueness is tested.
	w := makeV38(seed+int64(r)*1000000000000, g, old, id)
	w.Regime = regimesNoiseV53[r]
	rng := rand.New(rand.NewSource(w.Seed + 707))
	for _, row := range w.Outcomes {
		if r == 16 {
			flip := rng.Float64() < .1
			if flip {
				for i := range row {
					row[i] = !row[i]
				}
			}
			continue
		}
		rate := .2
		if r == 17 {
			rate = .1
		}
		for i := range row {
			if rng.Float64() < rate {
				row[i] = !row[i]
			}
		}
	}
	return w
}
func noiseConfigV53(mode string) (noise.Config, error) {
	c := noise.Config{Family: "rich", Prior: "moment", Strength: 2, Hazard: 1. / 16, Shared: true}
	switch mode {
	case "eta0", "learned_eta":
	case "eta10":
		c.Noise = .1
	case "eta20":
		c.Noise = .2
	default:
		return c, fmt.Errorf("noise mode")
	}
	return c, nil
}

type noiseDriverV53 struct {
	scalar  *noise.Model
	mixture *noise.Mixture
	single  []noise.Ticket
	joint   []noise.MixtureTicket
	weights [3]float64
}

func newNoiseDriverV53(base []float64, mode string) (*noiseDriverV53, error) {
	c, e := noiseConfigV53(mode)
	if e != nil {
		return nil, e
	}
	d := &noiseDriverV53{}
	if mode == "learned_eta" {
		d.mixture, e = noise.NewMixture(base, 1, 2400, c)
		d.joint = make([]noise.MixtureTicket, 2400)
	} else {
		d.scalar, e = noise.NewMemoV49(base, 1, 2400, c)
		d.single = make([]noise.Ticket, 2400)
		j := 0
		if mode == "eta10" {
			j = 1
		}
		if mode == "eta20" {
			j = 2
		}
		d.weights[j] = 1
	}
	return d, e
}
func (d *noiseDriverV53) forecast(i int) (float64, float64, [3]float64, error) {
	if d.mixture != nil {
		q, e := d.mixture.Predict(i)
		if e != nil {
			return 0, 0, [3]float64{}, e
		}
		o, e := d.mixture.ObservedPredict(i)
		if e != nil {
			return 0, 0, [3]float64{}, e
		}
		w, e := d.mixture.Weights()
		return q, o, w, e
	}
	q, e := d.scalar.Predict(i)
	if e != nil {
		return 0, 0, [3]float64{}, e
	}
	o, e := d.scalar.ObservedPredict(i)
	return q, o, d.weights, e
}
func (d *noiseDriverV53) issue(k int) (float64, error) {
	if d.mixture != nil {
		t, e := d.mixture.Issue(k%150, int64(k))
		d.joint[k] = t
		return t.Forecast(), e
	}
	t, e := d.scalar.Issue(k%150, int64(k))
	d.single[k] = t
	return t.Forecast(), e
}
func (d *noiseDriverV53) resolve(k int, y bool, at int64) (noise.Receipt, error) {
	if d.mixture != nil {
		return d.mixture.Resolve(d.joint[k], y, at)
	}
	return d.scalar.Resolve(d.single[k], y, at)
}
func (d *noiseDriverV53) pending() int {
	if d.mixture != nil {
		return d.mixture.Pending()
	}
	return d.scalar.Pending()
}

func runNoiseV53(w windowWorldV37, mode, schedule string) (noiseArmV53, error) {
	if mode == "full" || mode == "adaptive" || mode == "rich_moment2" {
		a, e := runMomentV41(w, mode, schedule)
		return noiseArmV53{orientationArmV38: a}, e
	}
	started := time.Now()
	a := noiseArmV53{orientationArmV38: orientationArmV38{ReadyAt: -1, windowArmV37: windowArmV37{Mode: mode, Schedule: schedule, Issued: make([]float64, 2400), ExpertIssued: make([][4]float64, 2400)}}, ObservedIssued: make([]float64, 2400), NoiseIssued: make([][3]float64, 2400)}
	phase := time.Now()
	due, e := scheduleV36(delayedWorldV36{Seed: w.Seed, Outcomes: w.Outcomes}, schedule)
	a.Costs.ScheduleNS = time.Since(phase).Nanoseconds()
	if e != nil {
		return a, e
	}
	phase = time.Now()
	d, e := newNoiseDriverV53(w.Base, mode)
	a.Costs.SetupNS = time.Since(phase).Nanoseconds()
	if e != nil {
		return a, e
	}
	last := len(due) - 1
	for len(due[last]) == 0 {
		last--
	}
	for tick := 0; tick <= last; tick++ {
		if tick < 2400 {
			phase = time.Now()
			q, o, weights, e := d.forecast(tick % 150)
			if e != nil {
				return a, e
			}
			issued, e := d.issue(tick)
			a.Costs.IssueNS += time.Since(phase).Nanoseconds()
			if e != nil || issued != q {
				return a, fmt.Errorf("issue forecast mismatch %v", e)
			}
			a.Issued[tick] = q
			a.ObservedIssued[tick] = o
			a.NoiseIssued[tick] = weights
			a.PeakPending = max(a.PeakPending, d.pending())
		}
		for _, k := range due[tick] {
			phase = time.Now()
			r, e := d.resolve(k, w.Outcomes[k/150][k%150], int64(tick))
			a.Costs.ResolveNS += time.Since(phase).Nanoseconds()
			if e != nil {
				return a, e
			}
			a.Receipts = append(a.Receipts, DelayedReceipt{Member: r.Member, TrialOrdinal: r.TrialOrdinal, Epoch: r.Epoch, IssuedAt: r.IssuedAt, ArrivedAt: r.ArrivedAt, Forecast: r.Forecast, Useful: r.Useful})
		}
		if (tick < 2400 && tick%150 == 149) || tick == last {
			phase = time.Now()
			s := windowSnapshotV37{delayedSnapshotV36: delayedSnapshotV36{Tick: tick, Arrived: len(a.Receipts), Pending: d.pending(), Forecast: make([]float64, 150)}}
			obs := make([]float64, 150)
			weights := [3]float64{}
			for i := range s.Forecast {
				q, o, w, e := d.forecast(i)
				if e != nil {
					return a, e
				}
				s.Forecast[i] = q
				obs[i] = o
				weights = w
			}
			a.Costs.SnapshotNS += time.Since(phase).Nanoseconds()
			a.Snapshots = append(a.Snapshots, s)
			a.ObservedSnapshots = append(a.ObservedSnapshots, obs)
			a.NoiseSnapshots = append(a.NoiseSnapshots, weights)
		}
	}
	a.Costs.AccountedNS = a.Costs.SetupNS + a.Costs.IssueNS + a.Costs.ResolveNS + a.Costs.SnapshotNS
	a.Costs.ElapsedNS = time.Since(started).Nanoseconds()
	return a, nil
}

func noiseSourcesV53(t *testing.T) map[string]string {
	t.Helper()
	b, e := os.ReadFile(os.Getenv("EVENTFRAME_NOISE_V53_FREEZE"))
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Files map[string]string `json:"files"`
	}
	if e = json.Unmarshal(b, &f); e != nil || len(f.Files) == 0 {
		t.Fatal("source freeze", e)
	}
	for p, h := range f.Files {
		b, e := os.ReadFile(filepath.Join(rootV34(t), p))
		if e != nil || hashV34(b) != h {
			t.Fatal("source mismatch", p, e)
		}
	}
	return f.Files
}
func TestNoiseV53Experiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_NOISE_V53_OUT")
	if path == "" {
		t.Skip("explicit isolated output")
	}
	sources := noiseSourcesV53(t)
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	buf := bufio.NewWriter(f)
	enc := json.NewEncoder(buf)
	if e = enc.Encode(manifestV34{Kind: "manifest", Split: "diagnostic", SeedBase: 2026105307, Worlds: 36, Sources: sources}); e != nil {
		t.Fatal(e)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesNoiseV53 {
			w := noiseWorldV53{Population: makeNoiseV53(2026105307, g, r, 0)}
			for _, schedule := range schedulesV39 {
				for _, mode := range modesNoiseV53 {
					a, e := runNoiseV53(w.Population, mode, schedule)
					if e != nil {
						t.Fatal(e)
					}
					scoreWindowV37(w.Population, &a.windowArmV37)
					w.Arms = append(w.Arms, a)
				}
			}
			if e = enc.Encode(w); e != nil {
				t.Fatal(e)
			}
			t.Log(w.Population.Geometry, w.Population.Regime)
		}
	}
	if e = buf.Flush(); e != nil {
		t.Fatal(e)
	}
	if e = f.Sync(); e != nil {
		t.Fatal(e)
	}
}

// Reconstruct the learned-noise joint law using separately integrated full
// histories and independent marginal evidence, not the candidate predictor.
func independentNoiseV53(w windowWorldV37, a noiseArmV53) error {
	c, e := noiseConfigV53(a.Mode)
	if e != nil {
		return e
	}
	refs := []*nref.Reference{}
	etas := []float64{c.Noise}
	if a.Mode == "learned_eta" {
		etas = []float64{0, .1, .2}
	}
	for _, eta := range etas {
		c.Noise = eta
		r, e := nref.New(w.Base, c)
		if e != nil {
			return e
		}
		refs = append(refs, r)
	}
	forecast := func(i int) (float64, float64, [3]float64, error) {
		weights := [3]float64{}
		local := make([]float64, len(refs))
		if len(refs) == 1 {
			local[0] = 1
			j := 0
			if etas[0] == .1 {
				j = 1
			}
			if etas[0] == .2 {
				j = 2
			}
			weights[j] = 1
		} else {
			mx := math.Inf(-1)
			for j, r := range refs {
				v, e := r.LogEvidence()
				if e != nil {
					return 0, 0, weights, e
				}
				local[j] = v + math.Log([]float64{.8, .1, .1}[j])
				mx = math.Max(mx, local[j])
			}
			sum := 0.
			for j := range local {
				local[j] = math.Exp(local[j] - mx)
				sum += local[j]
			}
			for j := range local {
				local[j] /= sum
				weights[j] = local[j]
			}
		}
		q, o := 0., 0.
		for j, r := range refs {
			p, e := r.Predict(i)
			if e != nil {
				return 0, 0, weights, e
			}
			q += local[j] * p
			o += local[j] * ((1-etas[j])*p + etas[j]*(1-p))
		}
		return q, o, weights, nil
	}
	due, e := scheduleV36(delayedWorldV36{Seed: w.Seed, Outcomes: w.Outcomes}, a.Schedule)
	if e != nil {
		return e
	}
	last := len(due) - 1
	for len(due[last]) == 0 {
		last--
	}
	position, snap, peak := 0, 0, 0
	check := func(i int, q, o float64, weights [3]float64) error {
		v, u, ww, e := forecast(i)
		if e != nil || !closeV36(q, v) || !closeV36(o, u) {
			return fmt.Errorf("noise law %d %v", i, e)
		}
		for j := range ww {
			if !closeV36(ww[j], weights[j]) {
				return fmt.Errorf("joint noise evidence weights")
			}
		}
		return nil
	}
	for tick := 0; tick <= last; tick++ {
		if tick < 2400 {
			if e = check(tick%150, a.Issued[tick], a.ObservedIssued[tick], a.NoiseIssued[tick]); e != nil {
				return e
			}
			if a.ExpertIssued[tick] != ([4]float64{}) {
				return fmt.Errorf("noise expert")
			}
			for _, r := range refs {
				if e = r.Issue(tick % 150); e != nil {
					return e
				}
			}
			peak = max(peak, tick+1-position)
		}
		for _, k := range due[tick] {
			if position >= len(a.Receipts) {
				return fmt.Errorf("missing receipt")
			}
			r := a.Receipts[position]
			if r.Member != k%150 || r.TrialOrdinal != k/150+1 || r.Epoch != 1 || r.IssuedAt != int64(k) || r.ArrivedAt != int64(tick) || r.Forecast != a.Issued[k] || r.Useful != w.Outcomes[k/150][k%150] {
				return fmt.Errorf("noise receipt identity")
			}
			for _, ref := range refs {
				if e = ref.Resolve(r.Member, r.TrialOrdinal, r.Useful); e != nil {
					return e
				}
			}
			position++
		}
		if (tick < 2400 && tick%150 == 149) || tick == last {
			if snap >= len(a.Snapshots) || snap >= len(a.ObservedSnapshots) || snap >= len(a.NoiseSnapshots) {
				return fmt.Errorf("missing snapshot")
			}
			s := a.Snapshots[snap]
			if s.Tick != tick || s.Arrived != position || s.Pending != min(tick+1, 2400)-position || len(s.Forecast) != 150 || len(a.ObservedSnapshots[snap]) != 150 || s.Weights != ([4]float64{}) {
				return fmt.Errorf("snapshot coverage")
			}
			for i, q := range s.Forecast {
				if e = check(i, q, a.ObservedSnapshots[snap][i], a.NoiseSnapshots[snap]); e != nil {
					return e
				}
			}
			if e = independentMetricsV38(w, s); e != nil {
				return e
			}
			snap++
		}
	}
	if position != 2400 || snap != len(a.Snapshots) || snap != len(a.ObservedSnapshots) || snap != len(a.NoiseSnapshots) || a.PeakPending != peak || a.ReadyAt != -1 || len(a.Template) != 0 {
		return fmt.Errorf("noise drain/template")
	}
	return nil
}
func auditNoiseArmV53(w windowWorldV37, a noiseArmV53) error {
	if len(a.Issued) != 2400 || len(a.ExpertIssued) != 2400 || len(a.Receipts) != 2400 {
		return fmt.Errorf("arm shape")
	}
	c := a.Costs
	if c.SetupNS < 0 || c.ScheduleNS < 0 || c.IssueNS < 0 || c.ResolveNS < 0 || c.SnapshotNS < 0 || c.AccountedNS != c.SetupNS+c.IssueNS+c.ResolveNS+c.SnapshotNS || c.ElapsedNS < c.AccountedNS+c.ScheduleNS {
		return fmt.Errorf("cost accounting")
	}
	var e error
	if a.Mode == "full" || a.Mode == "adaptive" || a.Mode == "rich_moment2" {
		if len(a.ObservedIssued) != 0 || len(a.NoiseIssued) != 0 || len(a.ObservedSnapshots) != 0 || len(a.NoiseSnapshots) != 0 {
			return fmt.Errorf("control extras")
		}
		if a.Mode == "rich_moment2" {
			e = independentMomentV41(w, a.orientationArmV38)
		} else {
			e = independentOrientationV38(w, a.orientationArmV38)
		}
	} else {
		if len(a.ObservedIssued) != 2400 || len(a.NoiseIssued) != 2400 {
			return fmt.Errorf("noise extras shape")
		}
		e = independentNoiseV53(w, a)
	}
	if e != nil {
		return e
	}
	// Independently recompute risk and the two-consecutive-round recovery rule.
	risk, priority, recovery := 0., 0., 0.
	for trial, q := range a.Issued {
		p, weight := w.Rates[trial/150][trial%150], 1.
		if trial%150 < 10 {
			weight = 3
		}
		loss := q*q - 2*q*p + p
		risk += loss / 2400
		priority += weight * loss / (16 * 170)
	}
	for phase, start := range w.Changes {
		end := 16
		if phase+1 < len(w.Changes) {
			end = w.Changes[phase+1]
		}
		delay := end - start + 1
		for round := start + 1; round < end; round++ {
			one, two := windowAtV37(a.windowArmV37, (round-1)*150+149), windowAtV37(a.windowArmV37, round*150+149)
			if one.Brier <= .20 && two.Brier <= .20 && one.PacketUsefulness >= .75 && two.PacketUsefulness >= .75 {
				delay = round - start + 1
				break
			}
		}
		recovery += float64(delay) / float64(len(w.Changes))
	}
	if !closeV36(risk, a.IssuedBrier) || !closeV36(priority, a.IssuedPriority) || !closeV36(recovery, a.Recovery) {
		return fmt.Errorf("derived metrics")
	}
	return nil
}
func TestNoiseV53Audit(t *testing.T) {
	path := os.Getenv("EVENTFRAME_NOISE_V53_AUDIT")
	if path == "" {
		t.Skip("explicit artifact")
	}
	sources := noiseSourcesV53(t)
	f, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	d := json.NewDecoder(bufio.NewReader(f))
	d.DisallowUnknownFields()
	var manifest manifestV34
	if e = d.Decode(&manifest); e != nil || manifest.Kind != "manifest" || manifest.Split != "diagnostic" || manifest.SeedBase != 2026105307 || manifest.Worlds != 36 || !reflect.DeepEqual(manifest.Sources, sources) {
		t.Fatal("manifest", e)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesNoiseV53 {
			var w noiseWorldV53
			if e = d.Decode(&w); e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(w.Population, makeNoiseV53(2026105307, g, r, 0)) || len(w.Arms) != 21 {
				t.Fatal("population coverage")
			}
			for j, a := range w.Arms {
				if a.Mode != modesNoiseV53[j%7] || a.Schedule != schedulesV39[j/7] {
					t.Fatal("arm order")
				}
				if e = auditNoiseArmV53(w.Population, a); e != nil {
					t.Fatal(g, r, j, e)
				}
			}
			for s := 0; s < 3; s++ {
				zero, old := w.Arms[s*7+3].orientationArmV38, w.Arms[s*7+2].orientationArmV38
				zero.Costs = delayedCostsV36{}
				old.Costs = delayedCostsV36{}
				zero.Mode = old.Mode
				if !reflect.DeepEqual(zero, old) {
					t.Fatal("eta0 exact full non-cost parity")
				}
			}
			t.Log(w.Population.Geometry, w.Population.Regime)
		}
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		t.Fatal("trailing artifact", e)
	}
}
func TestNoiseV53FutureAndCorruptions(t *testing.T) {
	for _, mode := range modesNoiseV53[3:] {
		for _, schedule := range schedulesV39 {
			w, x := makeNoiseV53(740533, 1, 15, 0), makeNoiseV53(740533, 1, 15, 0)
			for round := 8; round < 16; round++ {
				for i := range x.Outcomes[round] {
					x.Outcomes[round][i] = !x.Outcomes[round][i]
					x.Rates[round][i] = 1 - x.Rates[round][i]
				}
			}
			a, e := runNoiseV53(w, mode, schedule)
			if e != nil {
				t.Fatal(e)
			}
			b, e := runNoiseV53(x, mode, schedule)
			if e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(a.Issued[:1201], b.Issued[:1201]) || !reflect.DeepEqual(a.ObservedIssued[:1201], b.ObservedIssued[:1201]) || !reflect.DeepEqual(a.NoiseIssued[:1201], b.NoiseIssued[:1201]) {
				t.Fatal("future leak")
			}
			if reflect.DeepEqual(a.Issued[1500:], b.Issued[1500:]) {
				t.Fatal("future fork vacuous")
			}
			for j, s := range a.Snapshots {
				if s.Tick < 1200 && (!reflect.DeepEqual(s, b.Snapshots[j]) || !reflect.DeepEqual(a.ObservedSnapshots[j], b.ObservedSnapshots[j]) || a.NoiseSnapshots[j] != b.NoiseSnapshots[j]) {
					t.Fatal("snapshot future leak")
				}
			}
			scoreWindowV37(w, &a.windowArmV37)
			if e = auditNoiseArmV53(w, a); e != nil {
				t.Fatal(e)
			}
			edits := []func(*noiseArmV53){func(z *noiseArmV53) { z.Issued[1] += .01 }, func(z *noiseArmV53) { z.ObservedIssued[1] += .01 }, func(z *noiseArmV53) { z.NoiseIssued[1][0] += .01 }, func(z *noiseArmV53) { z.Receipts[0].Useful = !z.Receipts[0].Useful }, func(z *noiseArmV53) { z.Receipts[0].TrialOrdinal++ }, func(z *noiseArmV53) { z.Snapshots = z.Snapshots[:len(z.Snapshots)-1] }, func(z *noiseArmV53) { z.Costs.AccountedNS++ }}
			for _, edit := range edits {
				bytes, e := json.Marshal(a)
				if e != nil {
					t.Fatal(e)
				}
				var bad noiseArmV53
				if e = json.Unmarshal(bytes, &bad); e != nil {
					t.Fatal(e)
				}
				edit(&bad)
				if e = auditNoiseArmV53(w, bad); e == nil {
					t.Fatal("corruption accepted", mode, schedule)
				}
			}
		}
	}
}
func TestNoiseV53SeedSeparation(t *testing.T) {
	seen := map[int64]bool{}
	for _, base := range []int64{2026103903, 2026103904, 2026104001, 2026104003, 2026104004, 2026104101, 2026104103, 2026104104, 2026104907, 2026104909, 2026104911, 2026105107, 2026105109, 2026105111, 2026105207, 2026105209, 2026105211, 2026105307, 2026105309, 2026105311} {
		for g := 0; g < 2; g++ {
			for r := range regimesNoiseV53 {
				for id := 0; id < 16; id++ {
					s := makeNoiseV53(base, g, r, id).Seed
					if seen[s] {
						t.Fatal("cohort collision", s)
					}
					seen[s] = true
				}
			}
		}
	}
	t.Log("unique actual seeds", len(seen))
}
func TestNoiseV53Allocation(t *testing.T) {
	path := os.Getenv("EVENTFRAME_NOISE_V53_ALLOCATION")
	if path == "" {
		t.Skip("explicit allocation output")
	}
	base := makeNoiseV53(2026105307, 0, 0, 0).Base
	results := map[string]uint64{}
	for _, mode := range modesNoiseV53[3:] {
		var maximum uint64
		for k := 0; k < 5; k++ {
			runtime.GC()
			var a, b runtime.MemStats
			runtime.ReadMemStats(&a)
			d, e := newNoiseDriverV53(base, mode)
			if e != nil {
				t.Fatal(e)
			}
			runtime.ReadMemStats(&b)
			runtime.KeepAlive(d)
			maximum = max(maximum, b.TotalAlloc-a.TotalAlloc)
		}
		results[mode] = maximum
	}
	b, e := json.MarshalIndent(results, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if _, e = f.Write(append(b, '\n')); e != nil {
		t.Fatal(e)
	}
	if e = f.Sync(); e != nil {
		t.Fatal(e)
	}
}
