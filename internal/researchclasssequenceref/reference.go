// Package researchclasssequenceref independently reconstructs the finite
// challenger using dense rate transitions, enumerated emissions and urn priors.
package researchclasssequenceref

import (
	"errors"
	"math"
)

type local [22]float64
type field [48]float64
type belief struct {
	p   [3]local
	log [3]float64
}
type global struct {
	p   field
	log [3]float64
}
type record struct {
	member                         int
	first, second, requested, a, b bool
}
type task struct {
	prior   local
	matrix  [22][22]float64
	indices []int
	latest  belief
}
type Reference struct {
	base       []float64
	mode       string
	hazard     float64
	tasks      []task
	rates      [][16]float64
	weights    [16]float64
	transition [16][16]float64
	records    []record
	forward    []global
	latest     global
}

var eta = [3]float64{0, .1, .2}
var hp = [3]float64{.8, .1, .1}

func logistic(x float64) float64 {
	if x < 0 {
		e := math.Exp(x)
		return e / (1 + e)
	}
	return 1 / (1 + math.Exp(-x))
}
func contribution(f int, b float64) float64 {
	if f == 0 {
		return 1
	}
	x := 2*b - 1
	if f == 1 {
		return x
	}
	return 2*x*x - 1
}
func New(base []float64, mode string, hazard float64) (*Reference, error) {
	if len(base) < 2 || len(base) > 200 || math.IsNaN(hazard) || math.IsInf(hazard, 0) || hazard < 0 || hazard > 1 {
		return nil, errors.New("reference constructor")
	}
	if mode != "local" && mode != "noise" && mode != "shared" && mode != "hybrid" {
		return nil, errors.New("reference mode")
	}
	r := &Reference{base: append([]float64(nil), base...), mode: mode, hazard: hazard, tasks: make([]task, len(base)), rates: make([][16]float64, len(base))}
	grid := []float64{-8, -4, 0, 4, 8}
	prob := []float64{.01, .09, .8, .09, .01}
	r.weights[0] = .72
	for f, w := range []float64{.12, .08, .08} {
		for j, p := range prob {
			r.weights[1+5*f+j] = w * p
		}
	}
	for i, b := range base {
		if b < .25 || b > .925+1e-12 || math.IsNaN(b) {
			return nil, errors.New("reference base")
		}
		t := &r.tasks[i]
		urn := local{1}
		for n := 0; n < 20; n++ {
			var next local
			for z := 0; z <= n; z++ {
				next[z+1] += urn[z] * (b + float64(z)) / (1 + float64(n))
				next[z] += urn[z] * (1 - b + float64(n-z)) / (1 + float64(n))
			}
			urn = next
		}
		t.prior[0] = .8
		for z := 0; z < 21; z++ {
			t.prior[z+1] = .2 * urn[z]
		}
		for from := 0; from < 22; from++ {
			for to, w := range t.prior {
				t.matrix[from][to] = hazard * w
				if from == to {
					t.matrix[from][to] += 1 - hazard
				}
			}
		}
		for h := 0; h < 3; h++ {
			t.latest.p[h] = t.prior
		}
		r.rates[i][0] = b
		for f := 0; f < 3; f++ {
			lo, hi := -100., 100.
			c := math.Log(b / (1 - b))
			a := contribution(f, b)
			for n := 0; n < 100; n++ {
				mean, derivative := 0., 0.
				for j, g := range grid {
					p := logistic(c + a*g)
					mean += prob[j] * p
					derivative += prob[j] * p * (1 - p)
				}
				if mean < b {
					lo = c
				} else {
					hi = c
				}
				step := c
				if derivative > 0 {
					step = c + (b-mean)/derivative
				}
				if !(step > lo && step < hi) {
					step = (lo + hi) / 2
				}
				c = step
			}
			for j, g := range grid {
				r.rates[i][1+5*f+j] = logistic(c + a*g)
			}
		}
	}
	gh := hazard / float64(len(base))
	for from := 0; from < 16; from++ {
		for to, w := range r.weights {
			r.transition[from][to] = gh * w
			if from == to {
				r.transition[from][to] += 1 - gh
			}
		}
	}
	for h := 0; h < 3; h++ {
		for z, w := range r.weights {
			r.latest.p[h*16+z] = w
		}
	}
	return r, nil
}
func rate(b float64, z int) float64 {
	if z == 0 {
		return b
	}
	return float64(z-1) / 20
}
func emission(p, e float64, x record) float64 {
	if !x.first {
		return 1
	}
	sum := 0.
	for y := 0; y < 2; y++ {
		v := p
		if y == 0 {
			v = 1 - p
		}
		a := e
		if (y == 1) == x.a {
			a = 1 - e
		}
		v *= a
		if x.second {
			b := e
			if (y == 1) == x.b {
				b = 1 - e
			}
			v *= b
		}
		sum += v
	}
	return sum
}
func normalize(p local) (local, float64, error) {
	s := 0.
	for _, v := range p {
		if v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
			return p, 0, errors.New("reference mass")
		}
		s += v
	}
	if s == 0 {
		return local{}, 0, nil
	}
	for z := range p {
		p[z] /= s
	}
	return p, s, nil
}
func probabilities(logs, prior [3]float64) (out [3]float64, total float64) {
	max := math.Inf(-1)
	for h, w := range prior {
		if w > 0 {
			max = math.Max(max, math.Log(w)+logs[h])
		}
	}
	if math.IsInf(max, -1) {
		return out, max
	}
	s := 0.
	for h, w := range prior {
		if w > 0 {
			out[h] = math.Exp(math.Log(w) + logs[h] - max)
			s += out[h]
		}
	}
	for h := range out {
		out[h] /= s
	}
	return out, max + math.Log(s)
}
func (t *task) move(p local) (out local) {
	for from, w := range p {
		for to := 0; to < 22; to++ {
			out[to] += w * t.matrix[from][to]
		}
	}
	return
}
func (r *Reference) rebuildLocal(i, slot int, replacement record) (out belief, err error) {
	t := &r.tasks[i]
	for h := 0; h < 3; h++ {
		p := t.prior
		log := 0.
		for n, k := range t.indices {
			if n > 0 {
				p = t.move(p)
			}
			x := r.records[k]
			if k == slot {
				x = replacement
			}
			for z := range p {
				p[z] *= emission(rate(r.base[i], z), eta[h], x)
			}
			var s float64
			p, s, err = normalize(p)
			if err != nil {
				return out, err
			}
			if s == 0 {
				log = math.Inf(-1)
				break
			}
			if x.first {
				log += math.Log(s)
			}
		}
		out.p[h], out.log[h] = p, log
	}
	return
}
func (r *Reference) move(p field) (out field) {
	for h := 0; h < 3; h++ {
		for from := 0; from < 16; from++ {
			for to := 0; to < 16; to++ {
				out[h*16+to] += p[h*16+from] * r.transition[from][to]
			}
		}
	}
	return
}
func (r *Reference) step(p global, n int, x record) (out global, err error) {
	out = p
	if n > 0 {
		out.p = r.move(p.p)
	}
	for h := 0; h < 3; h++ {
		v := local{}
		for z := 0; z < 16; z++ {
			v[z] = out.p[h*16+z] * emission(r.rates[x.member][z], eta[h], x)
		}
		q, s, e := normalize(v)
		if e != nil {
			return out, e
		}
		for z := 0; z < 16; z++ {
			out.p[h*16+z] = q[z]
		}
		if s == 0 {
			out.log[h] = math.Inf(-1)
		} else if x.first {
			out.log[h] += math.Log(s)
		}
	}
	return
}
func (r *Reference) static(slot int, replacement record) (out global) {
	for h := 0; h < 3; h++ {
		logs := [16]float64{}
		for z, w := range r.weights {
			logs[z] = math.Log(w)
		}
		for n, x := range r.records {
			if n == slot {
				x = replacement
			}
			if !x.first {
				continue
			}
			for z := range logs {
				logs[z] += math.Log(emission(r.rates[x.member][z], eta[h], x))
			}
		}
		max := math.Inf(-1)
		for _, v := range logs {
			max = math.Max(max, v)
		}
		if math.IsInf(max, -1) {
			out.log[h] = max
			continue
		}
		s := 0.
		for z, v := range logs {
			out.p[h*16+z] = math.Exp(v - max)
			s += out.p[h*16+z]
		}
		for z := range logs {
			out.p[h*16+z] /= s
		}
		out.log[h] = max + math.Log(s)
	}
	return
}
func (r *Reference) rebuildShared(slot int, replacement record, commit bool) (out global, err error) {
	if r.hazard == 0 {
		return r.static(slot, replacement), nil
	}
	if slot > 0 {
		out = r.forward[slot-1]
	} else {
		for h := 0; h < 3; h++ {
			for z, w := range r.weights {
				out.p[h*16+z] = w
			}
		}
	}
	for n := slot; n < len(r.records); n++ {
		x := r.records[n]
		if n == slot {
			x = replacement
		}
		out, err = r.step(out, n, x)
		if err != nil {
			return out, err
		}
		if commit {
			r.forward[n] = out
		}
	}
	return
}
func (r *Reference) Issue(i int) error {
	if i < 0 || i >= len(r.base) || len(r.tasks[i].indices) >= 64 {
		return errors.New("reference issue")
	}
	n := len(r.records)
	r.records = append(r.records, record{member: i})
	r.tasks[i].indices = append(r.tasks[i].indices, n)
	if r.mode != "shared" {
		p, e := r.rebuildLocal(i, -1, record{})
		if e != nil {
			return e
		}
		r.tasks[i].latest = p
	}
	if r.mode == "shared" || r.mode == "hybrid" {
		p, e := r.step(r.latest, n, r.records[n])
		if e != nil {
			return e
		}
		r.latest = p
	}
	r.forward = append(r.forward, r.latest)
	return nil
}
func (r *Reference) find(i, n int) (int, error) {
	if i < 0 || i >= len(r.tasks) || n < 1 || n > len(r.tasks[i].indices) {
		return 0, errors.New("reference origin")
	}
	return r.tasks[i].indices[n-1], nil
}
func (r *Reference) Observe(i, n, measurement int, value bool) error {
	slot, e := r.find(i, n)
	if e != nil {
		return e
	}
	x := r.records[slot]
	if measurement == 1 {
		if x.first {
			return errors.New("reference first replay")
		}
		x.first, x.a = true, value
	} else if measurement == 2 {
		if !x.first || !x.requested || x.second {
			return errors.New("reference second phase")
		}
		x.second, x.b = true, value
	} else {
		return errors.New("reference measurement")
	}
	if r.mode != "shared" {
		p, e := r.rebuildLocal(i, slot, x)
		if e != nil {
			return e
		}
		r.tasks[i].latest = p
	}
	if r.mode == "shared" || r.mode == "hybrid" {
		p, e := r.rebuildShared(slot, x, true)
		if e != nil {
			return e
		}
		r.latest = p
	}
	r.records[slot] = x
	return nil
}
func (r *Reference) Audit(i, n int) error {
	slot, e := r.find(i, n)
	if e != nil {
		return e
	}
	x := &r.records[slot]
	if !x.first || x.requested {
		return errors.New("reference audit")
	}
	x.requested = true
	return nil
}

type view struct {
	model, noise [3]float64
	member       int
	replacement  *belief
	shared       global
	total        float64
}

func (r *Reference) view(i int, c *belief, s *global) (v view) {
	v.member, v.replacement, v.shared = i, c, r.latest
	if s != nil {
		v.shared = *s
	}
	localLog := 0.
	noiseLog := [3]float64{}
	for j, t := range r.tasks {
		p := t.latest
		if c != nil && i == j {
			p = *c
		}
		_, l := probabilities(p.log, hp)
		localLog += l
		for h := 0; h < 3; h++ {
			noiseLog[h] += p.log[h]
		}
	}
	var nl, sl float64
	v.noise, nl = probabilities(noiseLog, hp)
	_, sl = probabilities(v.shared.log, hp)
	prior := [3]float64{}
	switch r.mode {
	case "local":
		prior[0] = 1
	case "noise":
		prior[1] = 1
	case "shared":
		prior[2] = 1
	default:
		prior = [3]float64{1. / 3, 1. / 3, 1. / 3}
	}
	v.model, v.total = probabilities([3]float64{localLog, nl, sl}, prior)
	return
}
func (r *Reference) predict(i int, v view) (q, o float64) {
	p := r.tasks[i].latest
	if v.replacement != nil && v.member == i {
		p = *v.replacement
	}
	w, _ := probabilities(p.log, hp)
	for h := 0; h < 3; h++ {
		weight := v.model[0]*w[h] + v.model[1]*v.noise[h]
		if weight == 0 {
			continue
		}
		next, _, _ := normalize(r.tasks[i].move(p.p[h]))
		for z, x := range next {
			y := rate(r.base[i], z)
			q += weight * x * y
			o += weight * x * (eta[h] + (1-2*eta[h])*y)
		}
	}
	if v.model[2] > 0 {
		next := r.move(v.shared.p)
		w, _ := probabilities(v.shared.log, hp)
		for h := 0; h < 3; h++ {
			sum := 0.
			for z := 0; z < 16; z++ {
				sum += next[h*16+z]
			}
			if sum == 0 {
				continue
			}
			for z := 0; z < 16; z++ {
				weight := v.model[2] * w[h] * next[h*16+z] / sum
				y := r.rates[i][z]
				q += weight * y
				o += weight * (eta[h] + (1-2*eta[h])*y)
			}
		}
	}
	return
}
func (r *Reference) Predict(i int) (float64, float64, error) {
	if i < 0 || i >= len(r.base) {
		return 0, 0, errors.New("reference predict")
	}
	q, o := r.predict(i, r.view(-1, nil, nil))
	return q, o, nil
}
func (r *Reference) ModelWeights() [3]float64 { return r.view(-1, nil, nil).model }

func (r *Reference) localOrigin(slot int) (out [3]local, err error) {
	x := r.records[slot]
	t := &r.tasks[x.member]
	ordinal := -1
	for n, k := range t.indices {
		if k == slot {
			ordinal = n
		}
	}
	for h := 0; h < 3; h++ {
		if math.IsInf(t.latest.log[h], -1) {
			continue
		}
		p := t.prior
		var origin local
		for n, k := range t.indices {
			if n > 0 {
				p = t.move(p)
			}
			for z := range p {
				p[z] *= emission(rate(r.base[x.member], z), eta[h], r.records[k])
			}
			p, _, err = normalize(p)
			if err != nil {
				return out, err
			}
			if n == ordinal {
				origin = p
			}
		}
		back := local{}
		for z := range back {
			back[z] = 1
		}
		for n := len(t.indices) - 1; n > ordinal; n-- {
			k := t.indices[n]
			var terms local
			for to := range terms {
				terms[to] = emission(rate(r.base[x.member], to), eta[h], r.records[k]) * back[to]
			}
			var next local
			for from := 0; from < 22; from++ {
				for to := 0; to < 22; to++ {
					next[from] += t.matrix[from][to] * terms[to]
				}
			}
			back, _, err = normalize(next)
			if err != nil {
				return out, err
			}
		}
		for z := range origin {
			origin[z] *= back[z]
		}
		out[h], _, err = normalize(origin)
		if err != nil {
			return out, err
		}
	}
	return
}
func (r *Reference) sharedOrigin(slot int) (out field, err error) {
	if r.hazard == 0 {
		return r.latest.p, nil
	}
	for h := 0; h < 3; h++ {
		if math.IsInf(r.latest.log[h], -1) {
			continue
		}
		back := local{}
		for z := 0; z < 16; z++ {
			back[z] = 1
		}
		for n := len(r.records) - 1; n > slot; n-- {
			x := r.records[n]
			var terms local
			for to := 0; to < 16; to++ {
				terms[to] = emission(r.rates[x.member][to], eta[h], x) * back[to]
			}
			var next local
			for from := 0; from < 16; from++ {
				for to := 0; to < 16; to++ {
					next[from] += r.transition[from][to] * terms[to]
				}
			}
			back, _, err = normalize(next)
			if err != nil {
				return out, err
			}
		}
		p := local{}
		for z := 0; z < 16; z++ {
			p[z] = r.forward[slot].p[h*16+z] * back[z]
		}
		p, _, err = normalize(p)
		if err != nil {
			return out, err
		}
		for z := 0; z < 16; z++ {
			out[h*16+z] = p[z]
		}
	}
	return
}

type Option struct{ Observed, Uncertainty, Information, EdgeCut, Value, ClassGain float64 }

func entropy(p float64) float64 {
	if p <= 0 || p >= 1 {
		return 0
	}
	return -p*math.Log(p) - (1-p)*math.Log(1-p)
}
func (r *Reference) Query(i, n int) (out Option, err error) {
	return r.QueryMode(i, n, "all")
}

// Audit callers may omit unused branch evaluations, never requested fields.
func (r *Reference) QueryMode(i, n int, mode string) (out Option, err error) {
	if mode != "all" && mode != "uncertainty" && mode != "information" && mode != "falsification" && mode != "predictive" && mode != "model_class" && mode != "noise_class" {
		return out, errors.New("reference query mode")
	}
	slot, err := r.find(i, n)
	if err != nil {
		return out, err
	}
	x := r.records[slot]
	if !x.first || x.requested {
		return out, errors.New("reference query phase")
	}
	v := r.view(-1, nil, nil)
	var masses, yes [180]float64
	if v.model[0] > 0 || v.model[1] > 0 {
		origin, e := r.localOrigin(slot)
		if e != nil {
			return out, e
		}
		w, _ := probabilities(r.tasks[i].latest.log, hp)
		for h := 0; h < 3; h++ {
			for z, p := range origin[h] {
				k := h*22 + z
				first := emission(rate(r.base[i], z), eta[h], x)
				second := x
				second.second, second.b = true, true
				ratio := 0.
				if first > 0 {
					ratio = emission(rate(r.base[i], z), eta[h], second) / first
				}
				masses[k] = v.model[0] * w[h] * p
				masses[66+k] = v.model[1] * v.noise[h] * p
				yes[k], yes[66+k] = ratio, ratio
			}
		}
	}
	if v.model[2] > 0 {
		origin, e := r.sharedOrigin(slot)
		if e != nil {
			return out, e
		}
		w, _ := probabilities(v.shared.log, hp)
		for k, p := range origin {
			first := emission(r.rates[i][k%16], eta[k/16], x)
			second := x
			second.second, second.b = true, true
			ratio := 0.
			if first > 0 {
				ratio = emission(r.rates[i][k%16], eta[k/16], second) / first
			}
			masses[132+k], yes[132+k] = v.model[2]*w[k/16]*p, ratio
		}
	}
	for k, p := range masses {
		out.Observed += p * yes[k]
	}
	out.Uncertainty = entropy(out.Observed)
	if mode == "model_class" || mode == "noise_class" {
		// Independently enumerate the ORIGINAL joint (class,W) and normalize
		// each branch. Do not use the candidate's collapsed variance formula.
		classes := map[int][2]float64{}
		for k, w := range masses {
			f, h := 0, k/22
			if k >= 132 {
				f, h = 2, (k-132)/16
			} else if k >= 66 {
				f, h = 1, (k-66)/22
			}
			c := f
			if mode == "noise_class" {
				c = 3*f + h
			}
			pair := classes[c]
			pair[0] += w * (1 - yes[k])
			pair[1] += w * yes[k]
			classes[c] = pair
		}
		before, after := 0., 0.
		for c := 0; c < 9; c++ {
			pair := classes[c]
			before += (pair[0] + pair[1]) * (pair[0] + pair[1])
		}
		for b := 0; b < 2; b++ {
			total := 0.
			for c := 0; c < 9; c++ {
				total += classes[c][b]
			}
			if total == 0 {
				continue
			}
			for c := 0; c < 9; c++ {
				v := classes[c][b] / total
				after += total * v * v
			}
		}
		out.ClassGain = after - before
		if out.ClassGain < -2e-10 {
			return out, errors.New("reference negative class gain")
		}
		out.ClassGain = math.Max(0, out.ClassGain)
		return out, nil
	}
	if mode == "uncertainty" {
		return out, nil
	}
	out.Information = out.Uncertainty
	for k, p := range masses {
		if p == 0 {
			continue
		}
		y := yes[k]
		out.Information -= p * entropy(y)
		gain := -1.
		if out.Observed > 0 {
			gain += y * y / out.Observed
		}
		if out.Observed < 1 {
			gain += (1 - y) * (1 - y) / (1 - out.Observed)
		}
		out.EdgeCut += p * p * gain
	}
	out.Information = math.Max(0, out.Information)
	out.EdgeCut = math.Max(0, out.EdgeCut)
	if mode == "information" || mode == "falsification" {
		return out, nil
	}
	before := make([]float64, len(r.base))
	for j := range before {
		before[j], _ = r.predict(j, v)
	}
	branch := [2][]float64{}
	prob := [2]float64{}
	for y := 0; y < 2; y++ {
		replacement := x
		replacement.second, replacement.b = true, y == 1
		var local belief
		var shared global
		var cp *belief
		var sp *global
		if r.mode != "shared" {
			local, err = r.rebuildLocal(i, slot, replacement)
			if err != nil {
				return out, err
			}
			cp = &local
		}
		if r.mode == "shared" || r.mode == "hybrid" {
			shared, err = r.rebuildShared(slot, replacement, false)
			if err != nil {
				return out, err
			}
			sp = &shared
		}
		after := r.view(i, cp, sp)
		prob[y] = math.Exp(after.total - v.total)
		branch[y] = make([]float64, len(before))
		for j := range before {
			branch[y][j], _ = r.predict(j, after)
		}
	}
	if math.Abs(prob[0]+prob[1]-1) > 2e-10 || math.Abs(prob[1]-out.Observed) > 2e-10 {
		return out, errors.New("reference joint evidence ratio")
	}
	for j, q := range before {
		if math.Abs(prob[1]*branch[1][j]+prob[0]*branch[0][j]-q) > 2e-10 {
			return out, errors.New("reference tower")
		}
		out.Value += (prob[1]*(branch[1][j]-q)*(branch[1][j]-q) + prob[0]*(branch[0][j]-q)*(branch[0][j]-q)) / float64(len(before))
	}
	return
}
