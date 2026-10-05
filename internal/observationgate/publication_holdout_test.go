package observationgate

import (
	"fmt"
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

type publicationAudit struct {
	origin, arrival int
	live, reference observation.Sample
	mask, values    uint16
}
type publicationHoldout struct {
	models                          observationpreserved.Models
	subset                          *observationlearners.ConditionalForest
	mix                             bayes.ForecastMix
	trainOrigins, validationOrigins []int
	validationExperts               [][4]float64
}

// Fixed recent holdout: never refit on calibration labels before publication.
// The short expert is a fixed .7 count/.3 subset mixture during calibration;
// its inner selector starts at the same prior. This avoids fitting both levels
// of the mixture on the same small validation sample.
func fitPublicationHoldout(base *observation.Model, audits []publicationAudit, clock, version int, split, forest bool) (publicationHoldout, error) {
	var out publicationHoldout
	if base == nil || len(audits) < 32 || len(audits) > 256 || version < 1 || clock < 0 {
		return out, fmt.Errorf("invalid publication sample")
	}
	last := -1
	for _, a := range audits {
		if a.origin <= last || a.arrival < a.origin || a.arrival > clock || a.live.Bits > 511 || a.reference.Bits > 511 || a.mask > 511 || a.values&^a.mask != 0 || a.values != a.live.Bits&a.mask {
			return out, fmt.Errorf("invalid publication provenance")
		}
		last = a.origin
	}
	n := len(audits) - 16
	ls, rs := make([]observation.Sample, n), make([]observation.Sample, n)
	for i, a := range audits[:n] {
		ls[i], rs[i] = a.live, a.reference
		out.trainOrigins = append(out.trainOrigins, a.origin)
	}
	var err error
	out.models.Short, err = observation.Fit(ls[max(0, n-64):])
	if err != nil {
		return out, err
	}
	out.models.Local, err = observation.Fit(ls)
	if err != nil {
		return out, err
	}
	pool := append(append([]observation.Sample{}, ls[max(0, n-128):]...), rs[max(0, n-128):]...)
	out.models.Pooled, err = observation.Fit(pool)
	if err != nil {
		return out, err
	}
	out.models.Version = version
	trial := newSubsetTrial(base, 1)
	if forest {
		err = forestInputFit(trial, ls[max(0, n-64):])
	} else {
		err = trial.fit(ls[max(0, n-64):])
	}
	if err != nil {
		return out, err
	}
	out.subset = trial.model
	long := out.models.Pooled
	if split {
		long = out.models.Local
	}
	for _, a := range audits[n:] {
		var e [4]float64
		for i, m := range []*observation.Model{base, out.models.Short, long} {
			e[i], err = m.ForecastObserved(a.mask, a.values)
			if err != nil {
				return out, err
			}
		}
		q, err := out.subset.Forecast(a.mask, a.values)
		if err != nil {
			return out, err
		}
		e[1] = .7*e[1] + .3*q
		e[3] = .5
		out.validationExperts = append(out.validationExperts, e)
		out.validationOrigins = append(out.validationOrigins, a.origin)
	}
	y := make([]bool, 16)
	for i, a := range audits[n:] {
		y[i] = a.live.Outcome
	}
	out.mix.Weights, err = publicationWeights(out.validationExperts, y)
	return out, err
}

// Ordinary finite-expert Bernoulli likelihood weighting on a declared holdout.
// It is a selector initialization, not an error-control certificate under drift.
func publicationWeights(experts [][4]float64, outcomes []bool) ([4]float64, error) {
	prior := [4]float64{.7, .1, .1, .1}
	if len(experts) != len(outcomes) || len(experts) == 0 || len(experts) > 16 {
		return [4]float64{}, fmt.Errorf("invalid calibration length")
	}
	var logs [4]float64
	for j := range logs {
		logs[j] = math.Log(prior[j])
	}
	for i, e := range experts {
		for j, p := range e {
			if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
				return [4]float64{}, fmt.Errorf("invalid calibration probability")
			}
			p = math.Max(1e-6, math.Min(1-1e-6, p))
			if outcomes[i] {
				logs[j] += math.Log(p)
			} else {
				logs[j] += math.Log1p(-p)
			}
		}
	}
	peak := logs[0]
	for _, v := range logs {
		peak = math.Max(peak, v)
	}
	var out [4]float64
	sum := 0.
	for j, v := range logs {
		out[j] = math.Exp(v - peak)
		sum += out[j]
	}
	for j := range out {
		out[j] /= sum
	}
	return out, nil
}

func publicationFixture(t testing.TB) (*observation.Model, []publicationAudit) {
	t.Helper()
	base, err := observationpreserved.Base(false)
	if err != nil {
		t.Fatal(err)
	}
	a := make([]publicationAudit, 80)
	for i := range a {
		x := uint16(i * 13 % 512)
		a[i] = publicationAudit{origin: i * 3, arrival: i*3 + 2, live: observation.Sample{Bits: x, Outcome: x&4 != 0}, reference: observation.Sample{Bits: x ^ 17, Outcome: i%2 == 0}, mask: 511, values: x}
	}
	return base, a
}

func TestPublicationHoldoutIsolation(t *testing.T) {
	base, a := publicationFixture(t)
	x, err := fitPublicationHoldout(base, a, 250, 1, false, true)
	if err != nil {
		t.Fatal(err)
	}
	b := append([]publicationAudit{}, a...)
	for i := 64; i < len(b); i++ {
		b[i].live.Outcome = !b[i].live.Outcome
		b[i].reference.Outcome = !b[i].reference.Outcome
	}
	y, err := fitPublicationHoldout(base, b, 250, 1, false, true)
	if err != nil {
		t.Fatal(err)
	}
	for bits := uint16(0); bits < 512; bits++ {
		for j, m := range []*observation.Model{x.models.Short, x.models.Local, x.models.Pooled} {
			n := []*observation.Model{y.models.Short, y.models.Local, y.models.Pooled}[j]
			p, _ := m.ForecastObserved(511, bits)
			q, _ := n.ForecastObserved(511, bits)
			if p != q {
				t.Fatal("validation labels trained model")
			}
		}
		p, _ := x.subset.Forecast(511, bits)
		q, _ := y.subset.Forecast(511, bits)
		if p != q {
			t.Fatal("validation labels trained subset")
		}
	}
	if x.mix.Weights == y.mix.Weights {
		t.Fatal("calibration ignored labels")
	}
	if len(x.trainOrigins) != 64 || len(x.validationOrigins) != 16 || x.trainOrigins[63] >= x.validationOrigins[0] {
		t.Fatal("chronological separation")
	}
	for _, mutate := range []func([]publicationAudit){func(b []publicationAudit) { b[2].origin = b[1].origin }, func(b []publicationAudit) { b[70].arrival = 251 }, func(b []publicationAudit) { b[70].values ^= 1 }} {
		b = append([]publicationAudit{}, a...)
		mutate(b)
		if _, err := fitPublicationHoldout(base, b, 250, 1, false, true); err == nil {
			t.Fatal("invalid provenance accepted")
		}
	}
}

func TestPublicationWeightsLiteral(t *testing.T) {
	var e [][4]float64
	var y []bool
	product := [4]float64{.7, .1, .1, .1}
	for i := 0; i < 16; i++ {
		p := [4]float64{.2, .8, .3, .5}
		truth := i%3 != 0
		e = append(e, p)
		y = append(y, truth)
		for j, v := range p {
			if !truth {
				v = 1 - v
			}
			product[j] *= v
		}
	}
	w, err := publicationWeights(e, y)
	if err != nil {
		t.Fatal(err)
	}
	sum := 0.
	for _, p := range product {
		sum += p
	}
	for j := range w {
		if math.Abs(w[j]-product[j]/sum) > 1e-14 {
			t.Fatal("literal mismatch")
		}
	}
	for i := range e {
		e[i] = [4]float64{.5, .5, .5, .5}
	}
	w, err = publicationWeights(e, y)
	if err != nil {
		t.Fatal(err)
	}
	for j, p := range [4]float64{.7, .1, .1, .1} {
		if math.Abs(w[j]-p) > 1e-14 {
			t.Fatal("tie prior")
		}
	}
	e[0][0] = math.NaN()
	if _, err := publicationWeights(e, y); err == nil {
		t.Fatal("NaN accepted")
	}
}

func BenchmarkPublicationHoldout(b *testing.B) {
	base, a := publicationFixture(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := fitPublicationHoldout(base, a, 250, 1, false, true); err != nil {
			b.Fatal(err)
		}
	}
}
