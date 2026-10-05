package observationlearners

// Research-only cooperative cancellation at interval and recurrence boundaries.
import (
	"context"
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"math/bits"
)

func buildSegmentLikelihoodsTable(ctx context.Context, samples []observation.Sample, genericMass float64) (*segmentLikelihoods, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
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
	var logTerms [64][64][2]float64
	for n := 0; n < len(samples); n++ {
		for yes := 0; yes <= n; yes++ {
			p := (float64(yes) + .5) / float64(n+1)
			logTerms[n][yes][1] = math.Log(p)
			logTerms[n][yes][0] = math.Log(1 - p)
		}
	}
	indices, prior := segmentIndices(), segmentPrior()
	var counts [19683]struct{ n, yes uint16 }
	var agreement [512]uint16
	var gLog, bLog [512]float64
	// Each tail overwrites every cell before transforming; reuse is fit-local.
	var cells [19683]float64
	lg, lb := math.Log(genericMass), math.Log1p(-genericMass)
	for start := range samples {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		clear(counts[:])
		clear(agreement[:])
		clear(gLog[:])
		clear(bLog[:])
		for end := start; end < len(samples); end++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			s := samples[end]
			outcome := 0
			if s.Outcome {
				outcome = 1
			}
			gz, bz := math.Inf(-1), math.Inf(-1)
			for mask := range gLog {
				c := &counts[indices[s.Bits][mask]]
				gLog[mask] += logTerms[c.n][c.yes][outcome]
				c.n++
				if s.Outcome {
					c.yes++
				}
				agrees := (bits.OnesCount16(s.Bits&uint16(mask))%2 == 1) == s.Outcome
				outcomeB := 0
				if agrees {
					outcomeB = 1
				}
				bLog[mask] += logTerms[end-start][agreement[mask]][outcomeB]
				if agrees {
					agreement[mask]++
				}
				gz = math.Max(gz, prior[mask]+gLog[mask])
				bz = math.Max(bz, prior[mask]+bLog[mask])
			}
			// Max-shifted batch sums avoid one logarithm per hypothesis.
			gs, bs := 0., 0.
			for mask := range gLog {
				gs += math.Exp(prior[mask] + gLog[mask] - gz)
				bs += math.Exp(prior[mask] + bLog[mask] - bz)
			}
			gz += math.Log(gs)
			bz += math.Log(bs)
			out.logM[start][end+1] = segmentLogAdd(lg+gz, lb+bz)
			if end != len(samples)-1 {
				continue
			}
			// Family weights and within-family weights derive from the SAME
			// likelihood. A convenient but unrelated predictor cannot be attached.
			wg := math.Exp(lg + gz - out.logM[start][end+1])
			wb := math.Exp(lb + bz - out.logM[start][end+1])
			var walsh [512]float64
			var weights [512]float64
			constant := 0.
			for mask := range gLog {
				weights[mask] = wg * math.Exp(prior[mask]+gLog[mask]-gz)
				bw := wb * math.Exp(prior[mask]+bLog[mask]-bz)
				bp := (float64(agreement[mask]) + .5) / float64(end-start+2)
				constant += bw / 2
				walsh[mask] = bw * (.5 - bp)
			}
			// Decode mask metadata on every tail; no uncounted lookup initialization.
			for c := range cells {
				mask, v := 0, c
				for bit := 0; bit < 9; bit++ {
					if v%3 != 0 {
						mask |= 1 << bit
					}
					v /= 3
				}
				cells[c] = weights[mask] * (float64(counts[c].yes) + .5) / (float64(counts[c].n) + 1)
			}
			for stride := 1; stride < len(cells); stride *= 3 {
				for base := 0; base < len(cells); base += 3 * stride {
					for j := 0; j < stride; j++ {
						cells[base+j+stride] += cells[base+j]
						cells[base+j+2*stride] += cells[base+j]
					}
				}
			}
			for width := 1; width < len(walsh); width *= 2 {
				for base := 0; base < len(walsh); base += 2 * width {
					for j := 0; j < width; j++ {
						a, b := walsh[base+j], walsh[base+j+width]
						walsh[base+j], walsh[base+j+width] = a+b, a-b
					}
				}
			}
			for x := range out.tail[start] {
				out.tail[start][x] = cells[indices[x][511]] + constant + walsh[x]
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func fitSegmentPosteriorTable(ctx context.Context, left, clock int, history []segmentPacket, labelCap int, hazard, genericMass float64) (*segmentPosterior, error) {
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
	likelihoods, err := buildSegmentLikelihoodsTable(ctx, samples, genericMass)
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
