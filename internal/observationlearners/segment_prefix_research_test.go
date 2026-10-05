package observationlearners

// Research-only cooperative cancellation at interval and recurrence boundaries.
import (
	"context"
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"math/bits"
)

func buildPrefixLikelihoods(ctx context.Context, samples []observation.Sample, genericMass float64, prepared, capture *segmentPrefix) (*segmentLikelihoods, error) {
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
	if prepared != nil {
		if len(samples) != len(prepared.samples)+1 || genericMass != prepared.mass {
			return nil, errors.New("prefix contract mismatch")
		}
		for i, v := range prepared.samples {
			if samples[i] != v {
				return nil, errors.New("prefix evidence changed")
			}
		}
	}
	out := new(segmentLikelihoods)
	if prepared != nil {
		out.logM = prepared.logM
	}
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
		first := start
		if prepared != nil && start < len(prepared.states) {
			state := &prepared.states[start]
			counts, agreement, gLog, bLog = state.counts, state.agreement, state.gLog, state.bLog
			first = len(prepared.samples)
		}
		for end := first; end < len(samples); end++ {
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
			if capture != nil {
				capture.states[start] = prefixInterval{counts: counts, agreement: agreement, gLog: gLog, bLog: bLog}
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

type prefixInterval struct {
	counts     [19683]struct{ n, yes uint16 }
	agreement  [512]uint16
	gLog, bLog [512]float64
}
type segmentPrefix struct {
	samples []observation.Sample
	mass    float64
	states  []prefixInterval
	logM    [65][65]float64
}

// Prepared statistics are immutable and only cover an ordered eligible-label
// prefix. The caller still recomputes origin/hazard/as-of chronology separately.
func prepareSegmentPrefix(ctx context.Context, samples []observation.Sample, mass float64) (*segmentPrefix, error) {
	if len(samples) > 63 {
		return nil, errors.New("prefix too long")
	}
	p := &segmentPrefix{samples: append([]observation.Sample(nil), samples...), mass: mass, states: make([]prefixInterval, len(samples))}
	out, e := buildPrefixLikelihoods(ctx, samples, mass, nil, p)
	if e != nil {
		return nil, e
	}
	p.logM = out.logM
	return p, nil
}
