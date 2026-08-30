package claimsexperiment

import (
	"math"
	"math/rand"
	"sort"
)

func wilsonInterval(successes, trials int) ProportionInterval {
	interval := ProportionInterval{Method: IntervalMethod, ConfidenceLevel: defaultConfidenceLevel}
	if trials <= 0 {
		return interval
	}
	const z = 1.959963984540054
	p := float64(successes) / float64(trials)
	z2 := z * z
	denominator := 1 + z2/float64(trials)
	center := (p + z2/(2*float64(trials))) / denominator
	halfWidth := z * math.Sqrt(p*(1-p)/float64(trials)+z2/(4*float64(trials*trials))) / denominator
	interval.Lower = math.Max(0, center-halfWidth)
	interval.Upper = math.Min(1, center+halfWidth)
	return interval
}

func bootstrapMeanInterval(values []float64, seed int64, samples int) ProportionInterval {
	interval := ProportionInterval{Method: "trajectory_bootstrap_percentile", ConfidenceLevel: defaultConfidenceLevel}
	if len(values) == 0 || samples <= 0 {
		return interval
	}
	rng := rand.New(rand.NewSource(seed))
	means := make([]float64, samples)
	for sample := range means {
		for draw := 0; draw < len(values); draw++ {
			means[sample] += values[rng.Intn(len(values))]
		}
		means[sample] /= float64(len(values))
	}
	sort.Float64s(means)
	interval.Lower = means[int(.025*float64(samples))]
	interval.Upper = means[int(.975*float64(samples))]
	return interval
}
