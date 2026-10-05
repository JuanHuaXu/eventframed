package observationlearners

import (
	"errors"
	"math"
	"math/bits"
	"sync"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// A bounded research snapshot, not a persistent arrival journal. Slice position
// identifies the origin left+i. Arrives=-1 means no label; current-clock labels
// can enter only for strictly earlier origins. No query outcome is accepted.
type segmentPacket struct {
	Bits    uint16
	Outcome bool
	Arrives int
}

type segmentPosterior struct {
	predictions [512]float64
	lastStart   []float64
	origins     []int
	left, clock int
	logEvidence float64
}

// Precompute finite mask indices, avoiding nine ternary-index operations in
// each inner prediction loop. Allocate only when this research fitter is used;
// importing the package must not initialize its 512 KiB lookup table.
var segmentIndices = sync.OnceValue(func() *[512][512]uint16 {
	table := new([512][512]uint16)
	for x := range table {
		for mask := range table[x] {
			table[x][mask] = uint16(partialIndex(uint16(mask), uint16(x&mask)))
		}
	}
	return table
})

func segmentLogAdd(a, b float64) float64 {
	if math.IsInf(a, -1) {
		return b
	}
	if math.IsInf(b, -1) {
		return a
	}
	if a < b {
		a, b = b, a
	}
	return a + math.Log1p(math.Exp(b-a))
}

var segmentPrior = sync.OnceValue(func() *[512]float64 {
	p := new([512]float64)
	z := math.Inf(-1)
	for mask := range p {
		k := float64(bits.OnesCount(uint(mask)))
		p[mask] = k*math.Log(1./3) + (9-k)*math.Log(2./3)
		z = segmentLogAdd(z, p[mask])
	}
	for mask := range p {
		p[mask] -= z
	}
	return p
})

type segmentLikelihoods struct {
	logM [65][65]float64 // half-open interval in eligible-label order
	tail [65][512]float64
}

// Build each distinct observed-label interval once. Unobserved time gaps enter
// the hazard recurrence, not fabricated labels or repeated likelihood factors.
func buildSegmentLikelihoods(samples []observation.Sample, genericMass float64) (*segmentLikelihoods, error) {
	if len(samples) > 64 || math.IsNaN(genericMass) || genericMass < 0 || genericMass > 1 {
		return nil, errors.New("invalid segment likelihood bounds")
	}
	for _, s := range samples {
		if s.Bits >= 512 {
			return nil, errors.New("invalid segment input")
		}
	}
	out := new(segmentLikelihoods)
	for x := range out.tail[len(samples)] {
		out.tail[len(samples)][x] = .5
	}
	if len(samples) == 0 {
		return out, nil
	}
	indices, prior := segmentIndices(), segmentPrior()
	var counts [19683]struct{ n, yes uint16 }
	var agreement [512]uint16
	var gLog, bLog [512]float64
	lg, lb := math.Log(genericMass), math.Log1p(-genericMass)
	for start := range samples {
		clear(counts[:])
		clear(agreement[:])
		clear(gLog[:])
		clear(bLog[:])
		for end := start; end < len(samples); end++ {
			s := samples[end]
			gz, bz := math.Inf(-1), math.Inf(-1)
			for mask := range gLog {
				c := &counts[indices[s.Bits][mask]]
				p := (float64(c.yes) + .5) / (float64(c.n) + 1)
				if !s.Outcome {
					p = 1 - p
				}
				gLog[mask] += math.Log(p)
				c.n++
				if s.Outcome {
					c.yes++
				}
				p = (float64(agreement[mask]) + .5) / float64(end-start+1)
				agrees := (bits.OnesCount16(s.Bits&uint16(mask))%2 == 1) == s.Outcome
				if !agrees {
					p = 1 - p
				}
				bLog[mask] += math.Log(p)
				if agrees {
					agreement[mask]++
				}
				gz = segmentLogAdd(gz, prior[mask]+gLog[mask])
				bz = segmentLogAdd(bz, prior[mask]+bLog[mask])
			}
			out.logM[start][end+1] = segmentLogAdd(lg+gz, lb+bz)
			if end != len(samples)-1 {
				continue
			}
			// Family weights and within-family weights derive from the SAME
			// likelihood. A convenient but unrelated predictor cannot be attached.
			wg := math.Exp(lg + gz - out.logM[start][end+1])
			wb := math.Exp(lb + bz - out.logM[start][end+1])
			for mask := range gLog {
				gw := wg * math.Exp(prior[mask]+gLog[mask]-gz)
				bw := wb * math.Exp(prior[mask]+bLog[mask]-bz)
				bp := (float64(agreement[mask]) + .5) / float64(end-start+2)
				for x := range out.tail[start] {
					c := counts[indices[x][mask]]
					p := bp
					if bits.OnesCount(uint(x&mask))%2 == 0 {
						p = 1 - p
					}
					out.tail[start][x] += gw*(float64(c.yes)+.5)/(float64(c.n)+1) + bw*p
				}
			}
		}
	}
	return out, nil
}

func fitSegmentPosterior(left, clock int, history []segmentPacket, labelCap int, hazard, genericMass float64) (*segmentPosterior, error) {
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
	likelihoods, err := buildSegmentLikelihoods(samples, genericMass)
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
	return m, nil
}
