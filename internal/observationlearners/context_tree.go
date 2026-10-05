package observationlearners

import (
	"errors"
	"math"
	"math/bits"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// contextTree is a research-only exact mixture over depth-three Boolean trees.
// It compiles historical evidence to a fixed prediction table, not a live update
// to any served forecast. Bit positions carry no special prior significance.
type contextTree struct{ predictions [512]float64 }

type contextTreeNode struct {
	n, yes               int
	logLeaf, logEvidence float64
	weights              [10]float64 // leaf, then the nine possible split coordinates
}

func fitContextTree(samples []observation.Sample) (*contextTree, error) {
	if len(samples) == 0 || len(samples) > 256 {
		return nil, errors.New("invalid context-tree sample count")
	}
	nodes := make(map[int]*contextTreeNode, 835)
	for mask := uint16(0); mask < 512; mask++ {
		if bits.OnesCount16(mask) > 3 {
			continue
		}
		for values := mask; ; values = (values - 1) & mask {
			nodes[partialIndex(mask, values)] = new(contextTreeNode)
			if values == 0 {
				break
			}
		}
	}
	for _, s := range samples {
		if s.Bits >= 512 {
			return nil, errors.New("invalid context-tree input")
		}
		for mask := uint16(0); mask < 512; mask++ {
			if bits.OnesCount16(mask) > 3 {
				continue
			}
			n := nodes[partialIndex(mask, s.Bits&mask)]
			p := (float64(n.yes) + .5) / (float64(n.n) + 1)
			if !s.Outcome {
				p = 1 - p
			}
			n.logLeaf += math.Log(p)
			n.n++
			if s.Outcome {
				n.yes++
			}
		}
	}
	// Child contexts have one more observed coordinate. Summing their log
	// evidences integrates both branches, including the unvisited branch.
	for depth := 3; depth >= 0; depth-- {
		for mask := uint16(0); mask < 512; mask++ {
			if bits.OnesCount16(mask) != depth {
				continue
			}
			for values := mask; ; values = (values - 1) & mask {
				n := nodes[partialIndex(mask, values)]
				n.logEvidence, n.weights[0] = n.logLeaf, 1
				if depth < 3 {
					var logs [10]float64
					logs[0] = n.logLeaf + math.Log(.5)
					maximum := logs[0]
					for b := 0; b < 9; b++ {
						bit := uint16(1) << b
						logs[b+1] = math.Inf(-1)
						if mask&bit != 0 {
							continue
						}
						a := nodes[partialIndex(mask|bit, values)]
						c := nodes[partialIndex(mask|bit, values|bit)]
						logs[b+1] = math.Log(.5/float64(9-depth)) + a.logEvidence + c.logEvidence
						maximum = math.Max(maximum, logs[b+1])
					}
					total := 0.
					for i, v := range logs {
						n.weights[i] = math.Exp(v - maximum)
						total += n.weights[i]
					}
					for i := range n.weights {
						n.weights[i] /= total
					}
					n.logEvidence = maximum + math.Log(total)
				}
				if values == 0 {
					break
				}
			}
		}
	}
	var predict func(uint16, uint16) float64
	predict = func(mask, x uint16) float64 {
		n := nodes[partialIndex(mask, x&mask)]
		p := n.weights[0] * (float64(n.yes) + .5) / (float64(n.n) + 1)
		if bits.OnesCount16(mask) < 3 {
			for b := 0; b < 9; b++ {
				bit := uint16(1) << b
				if mask&bit == 0 {
					p += n.weights[b+1] * predict(mask|bit, x)
				}
			}
		}
		return p
	}
	m := new(contextTree)
	for x := range m.predictions {
		m.predictions[x] = predict(0, uint16(x))
	}
	return m, nil
}
