package observationlearners

import (
	"context"
	"fmt"
	"math"
	"math/bits"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type regimeQueryMoments struct {
	logM  [64][64]float64
	query [8][64][64]float64
	tail  [64][8]float64
	joint [8][64][8]float64
}

type regimeQueryBatch struct {
	BaseLogEvidence float64
	Values          []regimeQueryValue
}

// Integrate query and probe moments under each interval's existing posterior.
// A same-cell Beta covariance must survive averaging; multiplying posterior
// means would incorrectly erase the value of learning the shared parameter.
func buildRegimeQueryMoments(ctx context.Context, samples []observation.Sample, qx, probes []uint16, positions []int) (*regimeQueryMoments, error) {
	m := new(regimeQueryMoments)
	n := len(samples)
	indices, prior := segmentIndices(), segmentPrior()
	for q := range qx {
		for u := 0; u <= n; u++ {
			m.query[q][u][u] = .5
		}
	}
	for p := range probes {
		m.tail[n][p] = .5
	}
	for mask := 0; mask < 512; mask++ {
		w := math.Exp(prior[mask])
		for q, xq := range qx {
			for p, x := range probes {
				g := .25
				if xq&uint16(mask) == x&uint16(mask) {
					g += .125
				}
				b := .25
				if bits.OnesCount16((xq^x)&uint16(mask))%2 == 0 {
					b += .125
				} else {
					b -= .125
				}
				m.joint[q][n][p] += w * (.95*g + .05*b)
			}
		}
	}
	var counts [19683]struct{ n, yes uint16 }
	var agreement [512]uint16
	var gl, bl [512]float64
	lg, lb := math.Log(.95), math.Log(.05)
	for u := 0; u < n; u++ {
		clear(counts[:])
		clear(agreement[:])
		clear(gl[:])
		clear(bl[:])
		for v := u; v < n; v++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			s := samples[v]
			gz, bz := math.Inf(-1), math.Inf(-1)
			for mask := 0; mask < 512; mask++ {
				c := &counts[indices[s.Bits][mask]]
				p := (float64(c.yes) + .5) / (float64(c.n) + 1)
				if !s.Outcome {
					p = 1 - p
				}
				gl[mask] += math.Log(p)
				c.n++
				if s.Outcome {
					c.yes++
				}
				p = (float64(agreement[mask]) + .5) / float64(v-u+1)
				match := (bits.OnesCount16(s.Bits&uint16(mask))%2 == 1) == s.Outcome
				if !match {
					p = 1 - p
				}
				bl[mask] += math.Log(p)
				if match {
					agreement[mask]++
				}
				gz = math.Max(gz, prior[mask]+gl[mask])
				bz = math.Max(bz, prior[mask]+bl[mask])
			}
			gs, bs := 0., 0.
			for mask := 0; mask < 512; mask++ {
				gs += math.Exp(prior[mask] + gl[mask] - gz)
				bs += math.Exp(prior[mask] + bl[mask] - bz)
			}
			gz += math.Log(gs)
			bz += math.Log(bs)
			z := segmentLogAdd(lg+gz, lb+bz)
			m.logM[u][v+1] = z
			for mask := 0; mask < 512; mask++ {
				gw := math.Exp(lg + prior[mask] + gl[mask] - z)
				bw := math.Exp(lb + prior[mask] + bl[mask] - z)
				ap := (float64(agreement[mask]) + .5) / float64(v-u+2)
				var gp, bp [8]float64
				if v == n-1 {
					for p, x := range probes {
						c := counts[indices[x][mask]]
						gp[p] = (float64(c.yes) + .5) / (float64(c.n) + 1)
						bp[p] = ap
						if bits.OnesCount16(x&uint16(mask))%2 == 0 {
							bp[p] = 1 - ap
						}
						m.tail[u][p] += gw*gp[p] + bw*bp[p]
					}
				}
				for q, xq := range qx {
					if u > positions[q] || v+1 < positions[q] {
						continue
					}
					c := counts[indices[xq][mask]]
					gq := (float64(c.yes) + .5) / (float64(c.n) + 1)
					bq := ap
					if bits.OnesCount16(xq&uint16(mask))%2 == 0 {
						bq = 1 - ap
					}
					m.query[q][u][v+1] += gw*gq + bw*bq
					if v != n-1 {
						continue
					}
					for p, x := range probes {
						g := gq * gp[p]
						if indices[xq][mask] == indices[x][mask] {
							g += gq * (1 - gq) / (float64(c.n) + 2)
						}
						b := bq * bp[p]
						cov := ap * (1 - ap) / float64(v-u+3)
						if bits.OnesCount16((xq^x)&uint16(mask))%2 == 0 {
							b += cov
						} else {
							b -= cov
						}
						m.joint[q][u][p] += gw*g + bw*b
					}
				}
			}
		}
	}
	return m, ctx.Err()
}

// Cold batch: validates and owns its evidence and builds the base moments here.
// It neither needs nor borrows a previously fitted regimeQueryState.
func fitRegimeQueryBatch(ctx context.Context, left, clock int, history []segmentPacket, pool []int, probes []uint16) (*regimeQueryBatch, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if left < -16 || clock < 0 || clock > 256 || left > clock || len(history) != clock-left || len(pool) < 1 || len(pool) > 8 || len(probes) < 1 || len(probes) > 8 {
		return nil, fmt.Errorf("invalid batch bounds")
	}
	for _, x := range probes {
		if x >= 512 {
			return nil, fmt.Errorf("invalid batch probe")
		}
	}
	owned := append([]segmentPacket(nil), history...)
	var samples []observation.Sample
	var origins []int
	for i, p := range owned {
		if p.Bits >= 512 || p.Arrives < -1 || p.Arrives > 287 || (p.Arrives >= 0 && p.Arrives < max(0, left+i)) {
			return nil, fmt.Errorf("invalid batch packet")
		}
		if p.Arrives >= 0 && p.Arrives <= clock {
			samples = append(samples, observation.Sample{Bits: p.Bits, Outcome: p.Outcome})
			origins = append(origins, left+i)
		} else {
			owned[i].Outcome = false
			owned[i].Arrives = -1
		}
	}
	if len(samples) > 63 {
		return nil, fmt.Errorf("batch needs reserved evidence slot")
	}
	qx := make([]uint16, len(pool))
	positions := make([]int, len(pool))
	for q, j := range pool {
		if j < left || j >= clock || owned[j-left].Arrives >= 0 {
			return nil, fmt.Errorf("invalid batch query origin")
		}
		for k := 0; k < q; k++ {
			if pool[k] == j {
				return nil, fmt.Errorf("duplicate batch query")
			}
		}
		qx[q] = owned[j-left].Bits
		for _, o := range origins {
			if o < j {
				positions[q]++
			}
		}
	}
	m, err := buildRegimeQueryMoments(ctx, samples, qx, probes, positions)
	if err != nil {
		return nil, err
	}
	T := len(owned)
	before := make([]int, T+1)
	n := 0
	for t := 0; t <= T; t++ {
		for n < len(origins) && origins[n] < left+t {
			n++
		}
		before[t] = n
	}
	F := make([]float64, T+1)
	lh, ls := math.Log(.01), math.Log(.99)
	transition := func(a, b int) float64 {
		z := float64(b-a-1)*ls + m.logM[before[a]][before[b]]
		if a > 0 {
			z += lh
		}
		return z
	}
	for b := 1; b <= T; b++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		z := math.Inf(-1)
		for a := 0; a < b; a++ {
			z = segmentLogAdd(z, F[a]+transition(a, b))
		}
		F[b] = z
	}
	const survival = .99 * .99
	base := make([]float64, len(probes))
	for p := range probes {
		mean := 0.
		for a := 0; a < T; a++ {
			mean += math.Exp(F[a]+transition(a, T)-F[T]) * m.tail[before[a]][p]
		}
		base[p] = (1-survival)*.5 + survival*mean
	}
	out := &regimeQueryBatch{BaseLogEvidence: F[T]}
	for q, origin := range pool {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		qpos := origin - left
		Fq := make([]float64, T+1)
		for b := 1; b <= T; b++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if b <= qpos {
				Fq[b] = F[b]
				continue
			}
			z := math.Inf(-1)
			for a := 0; a < b; a++ {
				term := Fq[a] + transition(a, b)
				if a <= qpos {
					term = F[a] + transition(a, b) + math.Log(m.query[q][before[a]][before[b]])
				}
				z = segmentLogAdd(z, term)
			}
			Fq[b] = z
		}
		mass := math.Exp(Fq[T] - F[T])
		if !(mass > 0 && mass < 1) {
			return nil, fmt.Errorf("invalid batch query mass")
		}
		r := regimeQueryValue{Origin: origin, Mass: [2]float64{1 - mass, mass}, Base: append([]float64(nil), base...)}
		r.LogEvidence = [2]float64{F[T] + math.Log1p(-mass), Fq[T]}
		for p, p0 := range base {
			joint := 0.
			for a := 0; a < T; a++ {
				if a <= qpos {
					joint += math.Exp(F[a]+transition(a, T)-F[T]) * m.joint[q][before[a]][p]
				} else {
					joint += math.Exp(Fq[a]+transition(a, T)-F[T]) * m.tail[before[a]][p]
				}
			}
			joint = (1-survival)*.5*mass + survival*joint
			p1 := joint / mass
			pc0 := (p0 - joint) / (1 - mass)
			if !(p1 > 0 && p1 < 1 && pc0 > 0 && pc0 < 1) {
				return nil, fmt.Errorf("invalid batch conditional")
			}
			r.Conditional[0] = append(r.Conditional[0], pc0)
			r.Conditional[1] = append(r.Conditional[1], p1)
			r.Gain += ((1-mass)*(pc0-p0)*(pc0-p0) + mass*(p1-p0)*(p1-p0)) / float64(len(probes))
		}
		out.Values = append(out.Values, r)
	}
	return out, nil
}
