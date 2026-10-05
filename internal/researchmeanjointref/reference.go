// Package researchmeanjointref reconstructs the isolated hierarchy using
// urn dynamic programming, dense transitions and explicit latent-Y emissions.
// It imports no candidate-model code and is an offline checker, not a service.
package researchmeanjointref

import (
	"errors"
	"math"
)

type vector [22]float64
type belief struct {
	p   [243]vector
	log [243]float64
}
type record struct {
	member                         int
	first, second, requested, a, b bool
}
type task struct {
	mean    [27]float64
	prior   [27][3]vector
	matrix  [27][3][22][22]float64
	indices []int
	latest  belief
}
type Reference struct {
	base    []float64
	mode    string
	family  [3]float64
	means   [27]float64
	tasks   []task
	records []record
}
type Option struct{ Observed, Uncertainty, Information, EdgeCut, Value, ClassGain float64 }

var eta = [3]float64{0, .1, .2}
var hp = [3]float64{.8, .1, .1}

func New(base []float64, mode, family, means string, hazard float64) (*Reference, error) {
	if len(base) < 2 || len(base) > 200 || math.IsNaN(hazard) || math.IsInf(hazard, 0) || hazard < 0 || hazard > 1 || (mode != "local" && mode != "noise" && mode != "individual") {
		return nil, errors.New("mean reference constructor")
	}
	r := &Reference{base: append([]float64(nil), base...), mode: mode, tasks: make([]task, len(base))}
	switch family {
	case "learn":
		r.family = [3]float64{1. / 3, 1. / 3, 1. / 3}
	case "baseline":
		r.family[0] = 1
	case "current":
		r.family[1] = 1
	case "free":
		r.family[2] = 1
	default:
		return nil, errors.New("mean reference family")
	}
	switch means {
	case "learn":
		r.means[0], r.means[1] = .1, .8
		for k := 2; k < 27; k++ {
			r.means[k] = .004
		}
	case "baseline":
		r.means[0] = 1
	default:
		return nil, errors.New("mean reference means")
	}
	for i, b := range base {
		if math.IsNaN(b) || math.IsInf(b, 0) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("mean reference baseline")
		}
		task := &r.tasks[i]
		task.mean[0], task.mean[1] = b, (.9*b-.05)/.8
		k := 2
		for _, intercept := range []float64{.1, .3, .5, .7, .9} {
			for _, slope := range []float64{-.8, -.4, 0, .4, .8} {
				task.mean[k] = math.Max(.02, math.Min(.98, intercept+slope*float64(i)/float64(len(base)-1)))
				k++
			}
		}
		for theta, p := range task.mean {
			for a := 0; a < 3; a++ {
				task.prior[theta][a][0] = 1
				if a > 0 {
					strength, spike := 1., .8
					if a == 2 {
						strength, spike = 2, 0
					}
					urn := vector{1}
					for n := 0; n < 20; n++ {
						var next vector
						for z := 0; z <= n; z++ {
							next[z+1] += urn[z] * (strength*p + float64(z)) / (strength + float64(n))
							next[z] += urn[z] * (strength*(1-p) + float64(n-z)) / (strength + float64(n))
						}
						urn = next
					}
					task.prior[theta][a][0] = spike
					for z := 0; z < 21; z++ {
						task.prior[theta][a][z+1] = (1 - spike) * urn[z]
					}
				}
				for from := 0; from < 22; from++ {
					for to, mass := range task.prior[theta][a] {
						task.matrix[theta][a][from][to] = hazard * mass
						if from == to {
							task.matrix[theta][a][from][to] += 1 - hazard
						}
					}
				}
				for h := 0; h < 3; h++ {
					task.latest.p[9*theta+3*a+h] = task.prior[theta][a]
				}
			}
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
func likelihood(p, e float64, x record) float64 {
	if !x.first {
		return 1
	}
	s := 0.
	for y := 0; y < 2; y++ {
		q := p
		if y == 0 {
			q = 1 - p
		}
		v := e
		if (y == 1) == x.a {
			v = 1 - e
		}
		q *= v
		if x.second {
			v = e
			if (y == 1) == x.b {
				v = 1 - e
			}
			q *= v
		}
		s += q
	}
	return s
}
func unit(p vector) (vector, float64, error) {
	s := 0.
	for _, q := range p {
		if math.IsNaN(q) || math.IsInf(q, 0) || q < 0 {
			return p, 0, errors.New("mean reference mass")
		}
		s += q
	}
	if s == 0 {
		return vector{}, 0, nil
	}
	for z := range p {
		p[z] /= s
	}
	return p, s, nil
}
func probabilities(logs, prior [243]float64) (out [243]float64, total float64) {
	max := math.Inf(-1)
	for k, w := range prior {
		if w > 0 {
			max = math.Max(max, logs[k]+math.Log(w))
		}
	}
	if math.IsInf(max, -1) {
		return out, max
	}
	s := 0.
	for k, w := range prior {
		if w > 0 {
			out[k] = math.Exp(logs[k] + math.Log(w) - max)
			s += out[k]
		}
	}
	for k := range out {
		out[k] /= s
	}
	return out, max + math.Log(s)
}
func (t *task) move(k int, p vector) (out vector) {
	theta, a := k/9, (k/3)%3
	for from, q := range p {
		for to := 0; to < 22; to++ {
			out[to] += q * t.matrix[theta][a][from][to]
		}
	}
	return
}
func (r *Reference) rebuild(i, slot int, replacement record) (out belief, err error) {
	t := &r.tasks[i]
	for k := 0; k < 243; k++ {
		a, h := (k/3)%3, k%3
		if r.means[k/9]*r.family[a] == 0 {
			continue
		}
		p, log := t.prior[k/9][a], 0.
		for n, index := range t.indices {
			if n > 0 {
				p = t.move(k, p)
			}
			x := r.records[index]
			if index == slot {
				x = replacement
			}
			for z := range p {
				p[z] *= likelihood(rate(r.tasks[i].mean[k/9], z), eta[h], x)
			}
			var z float64
			p, z, err = unit(p)
			if err != nil {
				return out, err
			}
			if z == 0 {
				log = math.Inf(-1)
				break
			}
			if x.first {
				log += math.Log(z)
			}
		}
		out.p[k], out.log[k] = p, log
	}
	return
}
func (r *Reference) weights(i int, replacement *belief) [243]float64 {
	var logs, prior [243]float64
	for theta, p := range r.means {
		for a, q := range r.family {
			if p*q == 0 {
				continue
			}
			k := 9*theta + 3*a
			if r.mode == "noise" {
				for h, w := range hp {
					prior[k+h] = p * q * w
				}
			} else if r.mode == "local" {
				prior[k] = p * q
			}
			for j := range r.tasks {
				b := &r.tasks[j].latest
				if j == i && replacement != nil {
					b = replacement
				}
				if r.mode == "noise" {
					for h := 0; h < 3; h++ {
						logs[k+h] += b.log[k+h]
					}
				} else {
					var l, q [243]float64
					for h := 0; h < 3; h++ {
						l[h], q[h] = b.log[k+h], hp[h]
					}
					_, total := probabilities(l, q)
					logs[k] += total
				}
			}
		}
	}
	w, _ := probabilities(logs, prior)
	return w
}
func (r *Reference) memberWeights(i int, w [243]float64, replacement *belief) [243]float64 {
	if r.mode == "noise" {
		return w
	}
	b := &r.tasks[i].latest
	if replacement != nil {
		b = replacement
	}
	if r.mode == "individual" {
		var prior [243]float64
		for theta, p := range r.means {
			for a, q := range r.family {
				for h, u := range hp {
					prior[9*theta+3*a+h] = p * q * u
				}
			}
		}
		v, _ := probabilities(b.log, prior)
		return v
	}
	var out [243]float64
	for theta := 0; theta < 27; theta++ {
		for a := 0; a < 3; a++ {
			k := 9*theta + 3*a
			var l, q [243]float64
			for h := 0; h < 3; h++ {
				l[h], q[h] = b.log[k+h], hp[h]
			}
			v, _ := probabilities(l, q)
			for h := 0; h < 3; h++ {
				out[k+h] = w[k] * v[h]
			}
		}
	}
	return out
}
func (r *Reference) forecast(i, changed int, b *belief, w [243]float64) (clean, observed float64) {
	p := r.tasks[i].latest
	var override *belief
	if i == changed && b != nil {
		p, override = *b, b
	}
	w = r.memberWeights(i, w, override)
	for k, mass := range w {
		x, _, _ := unit(r.tasks[i].move(k, p.p[k]))
		for z, q := range x {
			p := rate(r.tasks[i].mean[k/9], z)
			clean += mass * q * p
			observed += mass * q * likelihood(p, eta[k%3], record{first: true, a: true})
		}
	}
	return
}
func (r *Reference) Predict(i int) (float64, float64, error) {
	if i < 0 || i >= len(r.tasks) {
		return 0, 0, errors.New("variance reference member")
	}
	q, o := r.forecast(i, -1, nil, r.weights(-1, nil))
	return q, o, nil
}
func (r *Reference) ModelWeights() (out [3]float64) {
	if r.mode == "individual" {
		for i := range r.tasks {
			for k, p := range r.memberWeights(i, [243]float64{}, nil) {
				out[(k/3)%3] += p / float64(len(r.tasks))
			}
		}
		return out
	}
	for k, p := range r.weights(-1, nil) {
		out[(k/3)%3] += p
	}
	return
}

func (r *Reference) Posterior(i int) ([243]float64, error) {
	if i < 0 || i >= len(r.tasks) {
		return [243]float64{}, errors.New("mean reference posterior member")
	}
	return r.memberWeights(i, r.weights(-1, nil), nil), nil
}
func (r *Reference) Issue(i int) error {
	if i < 0 || i >= len(r.tasks) || len(r.tasks[i].indices) >= 64 {
		return errors.New("variance reference issue")
	}
	t := &r.tasks[i]
	if len(t.indices) > 0 {
		for k, p := range t.latest.p {
			if r.means[k/9]*r.family[(k/3)%3] == 0 {
				continue
			}
			t.latest.p[k], _, _ = unit(t.move(k, p))
		}
	}
	t.indices = append(t.indices, len(r.records))
	r.records = append(r.records, record{member: i})
	return nil
}
func (r *Reference) Observe(i, ordinal, measurement int, value bool) error {
	if i < 0 || i >= len(r.tasks) || ordinal < 1 || ordinal > len(r.tasks[i].indices) || (measurement != 1 && measurement != 2) {
		return errors.New("variance reference observation")
	}
	slot := r.tasks[i].indices[ordinal-1]
	x := r.records[slot]
	if (measurement == 1 && x.first) || (measurement == 2 && (!x.first || !x.requested || x.second)) {
		return errors.New("variance reference replay")
	}
	if measurement == 1 {
		x.first, x.a = true, value
	} else {
		x.second, x.b = true, value
	}
	p, e := r.rebuild(i, slot, x)
	if e != nil {
		return e
	}
	r.records[slot], r.tasks[i].latest = x, p
	return nil
}
func (r *Reference) Audit(i, ordinal int) error {
	if i < 0 || i >= len(r.tasks) || ordinal < 1 || ordinal > len(r.tasks[i].indices) {
		return errors.New("variance reference audit")
	}
	x := &r.records[r.tasks[i].indices[ordinal-1]]
	if !x.first || x.requested {
		return errors.New("variance reference audit phase")
	}
	x.requested = true
	return nil
}
func (r *Reference) origin(i, ordinal int) (out [243]vector, err error) {
	t := &r.tasks[i]
	for k := 0; k < 243; k++ {
		a, h := (k/3)%3, k%3
		if r.means[k/9]*r.family[a] == 0 {
			continue
		}
		p := t.prior[k/9][a]
		var at vector
		supported := true
		for n, index := range t.indices {
			if n > 0 {
				p = t.move(k, p)
			}
			for z := range p {
				p[z] *= likelihood(rate(r.tasks[i].mean[k/9], z), eta[h], r.records[index])
			}
			var z float64
			p, z, err = unit(p)
			if err != nil {
				return out, err
			}
			if z == 0 {
				supported = false
				break
			}
			if n+1 == ordinal {
				at = p
			}
		}
		if !supported {
			continue
		}
		back := vector{}
		for z := range back {
			back[z] = 1
		}
		for n := len(t.indices) - 1; n >= ordinal; n-- {
			var next vector
			for from := 0; from < 22; from++ {
				for to := 0; to < 22; to++ {
					next[from] += t.matrix[k/9][a][from][to] * likelihood(rate(r.tasks[i].mean[k/9], to), eta[h], r.records[t.indices[n]]) * back[to]
				}
			}
			back, _, err = unit(next)
			if err != nil {
				return out, err
			}
		}
		for z := range at {
			at[z] *= back[z]
		}
		out[k], _, err = unit(at)
		if err != nil {
			return out, err
		}
	}
	return
}
func binaryEntropy(p float64) float64 {
	if p == 0 || p == 1 {
		return 0
	}
	return -p*math.Log(p) - (1-p)*math.Log(1-p)
}
func (r *Reference) QueryMode(i, ordinal int, mode string) (out Option, err error) {
	if i < 0 || i >= len(r.tasks) || ordinal < 1 || ordinal > len(r.tasks[i].indices) {
		return out, errors.New("variance reference query")
	}
	x := r.records[r.tasks[i].indices[ordinal-1]]
	if !x.first || x.requested {
		return out, errors.New("variance reference query phase")
	}
	w := r.weights(-1, nil)
	mw := r.memberWeights(i, w, nil)
	o, e := r.origin(i, ordinal)
	if e != nil {
		return out, e
	}
	var masses, yes [5346]float64
	var classes, joint [243]float64
	for k, p := range o {
		for z, q := range p {
			j := 22*k + z
			masses[j] = mw[k] * q
			d := likelihood(rate(r.tasks[i].mean[k/9], z), eta[k%3], x)
			paired := x
			paired.second, paired.b = true, true
			if d > 0 {
				yes[j] = likelihood(rate(r.tasks[i].mean[k/9], z), eta[k%3], paired) / d
			}
			out.Observed += masses[j] * yes[j]
			c := k
			if mode == "model_class" {
				c /= 3
			}
			classes[c] += masses[j]
			joint[c] += masses[j] * yes[j]
		}
	}
	out.Observed = math.Min(1, out.Observed)
	if mode == "forecast" {
		return
	}
	out.Uncertainty = binaryEntropy(out.Observed)
	if mode == "uncertainty" {
		return
	}
	if mode == "model_class" || mode == "noise_class" {
		for c, p := range classes {
			out.ClassGain -= p * p
			if out.Observed > 0 {
				out.ClassGain += joint[c] * joint[c] / out.Observed
			}
			if out.Observed < 1 {
				out.ClassGain += (p - joint[c]) * (p - joint[c]) / (1 - out.Observed)
			}
		}
		out.ClassGain = math.Max(0, out.ClassGain)
		return
	}
	if mode == "information" || mode == "falsification" {
		out.Information = out.Uncertainty
		for j, p := range masses {
			out.Information -= p * binaryEntropy(yes[j])
			out.EdgeCut -= p * p
			if out.Observed > 0 {
				out.EdgeCut += p * p * yes[j] * yes[j] / out.Observed
			}
			if out.Observed < 1 {
				out.EdgeCut += p * p * (1 - yes[j]) * (1 - yes[j]) / (1 - out.Observed)
			}
		}
		out.Information, out.EdgeCut = math.Max(0, out.Information), math.Max(0, out.EdgeCut)
		return
	}
	if mode != "predictive" {
		return out, errors.New("variance reference query mode")
	}
	var before [200]float64
	for j := range r.base {
		before[j], _ = r.forecast(j, -1, nil, w)
		out.Value += before[j] * (1 - before[j]) / float64(len(r.base))
	}
	for _, y := range []bool{false, true} {
		p := out.Observed
		if !y {
			p = 1 - p
		}
		if p == 0 {
			continue
		}
		paired := x
		paired.second, paired.b = true, y
		b, e := r.rebuild(i, r.tasks[i].indices[ordinal-1], paired)
		if e != nil {
			return out, e
		}
		v := r.weights(i, &b)
		for j := range r.base {
			q, _ := r.forecast(j, i, &b, v)
			out.Value -= p * q * (1 - q) / float64(len(r.base))
		}
	}
	if out.Value < -2e-10 {
		return out, errors.New("variance reference negative value")
	}
	out.Value = math.Max(0, out.Value)
	return
}
