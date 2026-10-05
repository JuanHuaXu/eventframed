package observationlearners

// Recompute chronology and hazard; reuse only checked ordered-label statistics.
import (
	"context"
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
)

func fitSegmentPosteriorPrefix(ctx context.Context, prepared *segmentPrefix, left, clock int, history []segmentPacket, labelCap int, hazard, genericMass float64) (*segmentPosterior, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if left < -16 || clock < 0 || clock > 256 || left > clock || len(history) != clock-left || labelCap < 1 || labelCap > 64 || math.IsNaN(hazard) || hazard <= 0 || hazard >= 1 {
		return nil, errors.New("invalid segment snapshot bounds")
	}
	origins := make([]int, 0, len(history))
	for i, p := range history {
		origin := left + i
		if p.Bits >= 512 || p.Arrives < -1 || p.Arrives > 287 || (p.Arrives >= 0 && p.Arrives < max(0, origin)) {
			return nil, errors.New("invalid segment packet")
		}
		if p.Arrives >= 0 && p.Arrives <= clock {
			origins = append(origins, origin)
		}
	}
	if len(origins) > labelCap {
		origins = origins[len(origins)-labelCap:]
	}
	samples := make([]observation.Sample, len(origins))
	for j, origin := range origins {
		p := history[origin-left]
		samples[j] = observation.Sample{Bits: p.Bits, Outcome: p.Outcome}
	}
	likelihoods, err := buildPrefixLikelihoods(ctx, samples, genericMass, prepared, nil)
	if err != nil {
		return nil, err
	}
	m := &segmentPosterior{left: left, clock: clock, origins: append([]int(nil), origins...), lastStart: make([]float64, len(history))}
	if len(history) == 0 {
		for x := range m.predictions {
			m.predictions[x] = .5
		}
		return m, nil
	}
	var before [273]int
	n := 0
	for i := 0; i <= len(history); i++ {
		for n < len(origins) && origins[n] < left+i {
			n++
		}
		before[i] = n
	}
	var prefix [273]float64
	lh, ls := math.Log(hazard), math.Log1p(-hazard)
	for end := 1; end <= len(history); end++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		z := math.Inf(-1)
		for start := 0; start < end; start++ {
			term := prefix[start] + float64(end-start-1)*ls + likelihoods.logM[before[start]][before[end]]
			if start > 0 {
				term += lh
			}
			z = segmentLogAdd(z, term)
			if end == len(history) {
				m.lastStart[start] = term
			}
		}
		prefix[end] = z
	}
	m.logEvidence = prefix[len(history)]
	total := 0.
	for i := range m.lastStart {
		m.lastStart[i] = math.Exp(m.lastStart[i] - m.logEvidence)
		total += m.lastStart[i]
	}
	for i := range m.lastStart {
		m.lastStart[i] /= total
	}
	for x := range m.predictions {
		p := 0.
		for start, w := range m.lastStart {
			p += w * likelihoods.tail[before[start]][x]
		}
		m.predictions[x] = hazard*.5 + (1-hazard)*p
		if len(origins) == 0 {
			m.predictions[x] = .5
		}
		if math.IsNaN(m.predictions[x]) || m.predictions[x] <= 0 || m.predictions[x] >= 1 {
			return nil, errors.New("invalid segment forecast")
		}
	}
	if len(origins) == 0 {
		m.logEvidence = 0
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return m, nil
}
