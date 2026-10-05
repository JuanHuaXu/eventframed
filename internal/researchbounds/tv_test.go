package researchbounds

import (
	"math"
	"math/big"
	"math/rand"
	"testing"
)

func rat(x float64) *big.Rat     { return new(big.Rat).SetFloat64(x) }
func add(a, b *big.Rat) *big.Rat { return new(big.Rat).Add(a, b) }
func mul(a, b *big.Rat) *big.Rat { return new(big.Rat).Mul(a, b) }
func sub(a, b *big.Rat) *big.Rat { return new(big.Rat).Sub(a, b) }
func sum(p []*big.Rat) *big.Rat {
	z := new(big.Rat)
	for _, x := range p {
		z.Add(z, x)
	}
	return z
}
func normalize(p []*big.Rat) []*big.Rat {
	z := sum(p)
	out := make([]*big.Rat, len(p))
	for i, x := range p {
		out[i] = new(big.Rat).Quo(x, z)
	}
	return out
}
func distance(p, q []*big.Rat) *big.Rat {
	z := new(big.Rat)
	for i := range p {
		z.Add(z, new(big.Rat).Abs(sub(p[i], q[i])))
	}
	return z.Quo(z, big.NewRat(2, 1))
}
func floatBound(x *big.Rat, upper bool) float64 {
	f, _ := x.Float64()
	cmp := rat(f).Cmp(x)
	if upper && cmp < 0 {
		return math.Nextafter(f, math.Inf(1))
	}
	if !upper && cmp > 0 {
		return math.Nextafter(f, math.Inf(-1))
	}
	return f
}

func TestStepAgainstExactRationalFiltering(t *testing.T) {
	rng := rand.New(rand.NewSource(2026100501))
	for trial := 0; trial < 4096; trial++ {
		p, q, g := make([]*big.Rat, 8), make([]*big.Rat, 8), make([]*big.Rat, 8)
		for i := range p {
			p[i] = big.NewRat(int64(1+rng.Intn(100)), 1)
			q[i] = big.NewRat(int64(1+rng.Intn(100)), 1)
			g[i] = big.NewRat(int64(rng.Intn(16)), 16)
		}
		p, q = normalize(p), normalize(q)
		before := distance(p, q)
		reset := float64(rng.Intn(5)) / 4
		for i := range p {
			p[i] = add(mul(rat(1-reset), p[i]), big.NewRat(int64(reset*8), 64))
			q[i] = add(mul(rat(1-reset), q[i]), big.NewRat(int64(reset*8), 64))
		}
		// reset/8 is exactly representable here; use the same common kernel.
		minimum, maximum := 1., 0.
		for i := range p {
			f, _ := g[i].Float64()
			minimum = math.Min(minimum, f)
			maximum = math.Max(maximum, f)
			p[i] = mul(p[i], g[i])
			q[i] = mul(q[i], g[i])
		}
		if sum(q).Sign() == 0 {
			continue
		}
		z := sum(q)
		p, q = normalize(p), normalize(q)
		kept := make([]*big.Rat, len(q))
		discard := new(big.Rat)
		for i, x := range q {
			if i%2 == 0 {
				kept[i] = new(big.Rat).Set(x)
			} else {
				kept[i] = new(big.Rat)
				discard.Add(discard, x)
			}
		}
		if sum(kept).Sign() == 0 {
			continue
		}
		kept = normalize(kept)
		bound, err := StepTV(floatBound(before, true), reset, minimum, maximum, floatBound(z, false), floatBound(discard, true))
		if err != nil {
			t.Fatal(err)
		}
		if distance(p, kept).Cmp(rat(bound)) > 0 {
			t.Fatalf("trial %d exact defect exceeds %.17g", trial, bound)
		}
	}
}

func TestEvidenceAwareBoundAndSafeRegion(t *testing.T) {
	x, err := StepTV(.001, .25, 0, 1, .5, 0)
	if err != nil || x > .002 {
		t.Fatalf("zero minimum need not imply vacuity: %g %v", x, err)
	}
	x, err = StepTV(.7, .25, 0, 1, .1, 0)
	if err != nil || x != 1 {
		t.Fatalf("rare unsupported evidence: %g %v", x, err)
	}
	d, err := SafeDiscard(.75, .5, 1, .01)
	if err != nil || d <= 0 || d > .005 {
		t.Fatalf("safe mass %g %v", d, err)
	}
	e := 0.
	for i := 0; i < 1000; i++ {
		e, err = StepTV(e, .75, .5, 1, 0, d)
		if err != nil || e > .01 {
			t.Fatalf("safe region escaped: %.17g %v", e, err)
		}
	}
	if _, err = SafeDiscard(.1, .02, 1, .01); err == nil {
		t.Fatal("unsafe global contraction accepted")
	}
	for _, v := range []float64{-1, math.NaN(), math.Inf(1), 1.1} {
		if _, err = StepTV(v, .5, .1, 1, .2, 0); err == nil {
			t.Fatal("invalid prior bound accepted")
		}
		if _, err = BinaryBrierShift(v); err == nil {
			t.Fatal("invalid Brier bound accepted")
		}
	}
	if _, err = StepTV(0, .5, 0, 0, 0, 0); err == nil {
		t.Fatal("impossible all-zero likelihood accepted")
	}
}

func TestPositiveUnderflowNeverBecomesZeroError(t *testing.T) {
	s := math.SmallestNonzeroFloat64
	bound, err := StepTV(1./8, 0, s, 2*s, s, 0)
	if err != nil || rat(bound).Cmp(big.NewRat(2, 9)) < 0 {
		t.Fatalf("lost positive underflow: %.17g %v", bound, err)
	}
	d, err := SafeDiscard(1-math.Ldexp(1, -53), s, math.Ldexp(1, -1023), .1)
	if err == nil && d > .075 {
		t.Fatalf("underflow enlarged safe mass: %.17g", d)
	}
	if zero, err := StepTV(0, 0, s, 2*s, s, 0); err != nil || zero != 0 {
		t.Fatal("exact equality acquired phantom error", zero, err)
	}
}

func TestBrierShiftForCommonBinaryMap(t *testing.T) {
	for i := 0; i <= 100; i++ {
		for j := 0; j <= 100; j++ {
			p, q := float64(i)/100, float64(j)/100
			bound, _ := BinaryBrierShift(math.Abs(p - q))
			for _, y := range []float64{0, 1} {
				if math.Abs((p-y)*(p-y)-(q-y)*(q-y)) > bound+1e-15 {
					t.Fatal("Brier bound violated")
				}
			}
		}
	}
	x, _ := BinaryBrierShift(.6982459966723692)
	if x != 1 {
		t.Fatal("large joint defect reported as usable loss protection")
	}
}

func BenchmarkStepTV(b *testing.B) {
	for b.Loop() {
		_, _ = StepTV(.01, .25, .02, .98, .3, .001)
	}
}
