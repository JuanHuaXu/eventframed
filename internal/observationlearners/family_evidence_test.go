package observationlearners

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type familyEvidenceFit struct {
	predictions                                      [512]float64
	genericLog, booleanLog, logEvidence, genericMass float64
}

// Research-only same-window Bayes model average. Owning the two fits here
// prevents callers from comparing evidence computed from different samples.
func fitFamilyEvidence(samples []observation.Sample, mass float64) (*familyEvidenceFit, error) {
	if math.IsNaN(mass) || mass < 0 || mass > 1 {
		return nil, errors.New("invalid family prior")
	}
	g, err := fitSubset(samples)
	if err != nil {
		return nil, err
	}
	b, err := fitBooleanSpecialist(samples)
	if err != nil {
		return nil, err
	}
	return combineFamilyEvidence(g, b, mass), nil
}

// Private aggregation assumes both fits were created from the same sample set.
func combineFamilyEvidence(g *subsetModel, b *booleanSpecialist, mass float64) *familyEvidenceFit {
	out := &familyEvidenceFit{genericLog: familyLogEvidence(g.evidence), booleanLog: familyLogEvidence(b.evidence)}
	lg, lb := math.Log(mass)+out.genericLog, math.Log1p(-mass)+out.booleanLog
	out.logEvidence = segmentLogAdd(lg, lb)
	out.genericMass = math.Exp(lg - out.logEvidence)
	wb := math.Exp(lb - out.logEvidence)
	for x := range out.predictions {
		out.predictions[x] = out.genericMass*g.predictions[x] + wb*b.predictions[x]
	}
	return out
}

func TestFamilyEvidenceReference(t *testing.T) {
	samples := ridgeTestSamples(6)
	before := append([]observation.Sample(nil), samples...)
	for _, mass := range []float64{0, .5, .95, 1} {
		got, err := fitFamilyEvidence(samples, mass)
		if err != nil {
			t.Fatal(err)
		}
		want := segmentTestMarginal(samples, mass)
		if math.Abs(math.Exp(got.logEvidence)-want) > 1e-13 {
			t.Fatal("beta evidence disagreement")
		}
		existing, weights, err := fitPriorConditional(samples, 1-mass)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(weights[0]-got.genericMass) > 1e-12 {
			t.Fatal("existing prior weight mismatch")
		}
		for x, p := range got.predictions {
			if math.Abs(existing.cells[partialIndex(511, uint16(x))].weighted-p) > 1e-12 {
				t.Fatal("existing skeptical-prior forecast mismatch")
			}
		}
		for _, x := range []uint16{0, 1, 17, 73, 511} {
			next := append(append([]observation.Sample(nil), samples...), observation.Sample{Bits: x, Outcome: true})
			predicted := segmentTestMarginal(next, mass) / want
			if math.Abs(got.predictions[x]-predicted) > 1e-12 {
				t.Fatal("predictive ratio disagreement")
			}
		}
		segment, err := buildSegmentLikelihoods(samples, mass)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(segment.logM[0][len(samples)]-got.logEvidence) > 1e-12 {
			t.Fatal("existing evidence mismatch")
		}
		for x, p := range got.predictions {
			if math.Abs(p-segment.tail[0][x]) > 1e-12 || p <= 0 || p >= 1 {
				t.Fatal("existing predictive mismatch")
			}
		}
		reversed := append([]observation.Sample(nil), samples...)
		for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
			reversed[i], reversed[j] = reversed[j], reversed[i]
		}
		other, err := fitFamilyEvidence(reversed, mass)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(other.logEvidence-got.logEvidence) > 1e-12 {
			t.Fatal("order changed evidence")
		}
		for x, p := range other.predictions {
			if math.Abs(p-got.predictions[x]) > 1e-12 {
				t.Fatal("order changed forecast")
			}
		}
	}
	if !reflect.DeepEqual(samples, before) {
		t.Fatal("sample mutation")
	}
	for _, mass := range []float64{math.NaN(), math.Inf(1), -1, 2} {
		if _, err := fitFamilyEvidence(samples, mass); err == nil {
			t.Fatal("invalid mass accepted")
		}
	}
	for _, bad := range [][]observation.Sample{nil, make([]observation.Sample, 257), {{Bits: 512}}} {
		if _, err := fitFamilyEvidence(bad, .95); err == nil {
			t.Fatal("invalid sample set accepted")
		}
	}
	t.Log("four priors, 20 beta predictive ratios, all512 existing predictions and ownership checks PASS")
}

var familyEvidenceSink *familyEvidenceFit

func BenchmarkFamilyEvidence(b *testing.B) {
	samples := ridgeTestSamples(64)
	g, err := fitSubset(samples)
	if err != nil {
		b.Fatal(err)
	}
	specialist, err := fitBooleanSpecialist(samples)
	if err != nil {
		b.Fatal(err)
	}
	b.Run("aggregate", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			familyEvidenceSink = combineFamilyEvidence(g, specialist, .95)
		}
	})
	b.Run("fit64", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var err error
			familyEvidenceSink, err = fitFamilyEvidence(samples, .95)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
