package observationlearners

// Research-only transform candidate. The frozen fitter remains unchanged.
// Likelihood recurrence copied verbatim; only posterior tail summation differs.
import (
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"math/bits"
)

func buildSegmentLikelihoodsTransform(samples []observation.Sample, genericMass float64) (*segmentLikelihoods, error) {
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
	// Each tail overwrites every cell before transforming; reuse is fit-local.
	var cells [19683]float64
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
	return out, nil
}
