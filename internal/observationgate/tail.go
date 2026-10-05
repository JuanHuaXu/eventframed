package observationgate

import (
	"math"
	"math/rand"
)

// TailGate preserves fixed-rate wealth inside a fixed convex mixture. It does
// not combine independent alarm decisions or reset the monitoring budget.
type TailGate struct {
	base  Gate
	Alert [4]bool
}

func (g *TailGate) Observe(d float64) ([4]bool, error) {
	a, e := g.base.Observe(d)
	if e != nil {
		return g.Alert, e
	}
	g.Alert[0], g.Alert[1] = a[0], a[1]
	for j := range g.base.Starts {
		if g.base.Next-1 < j*64 {
			continue
		}
		for arm, weight := range []float64{.9, .5} {
			var signs [2]float64
			for k := range signs {
				s := g.base.Starts[j][k]
				signs[k] = logMean([]float64{s.Log[2] + math.Log(2*weight), logMean(s.Log[:]) + math.Log(2*(1-weight))})
			}
			if logMean(signs[:]) >= math.Log(800) {
				g.Alert[arm+2] = true
			}
		}
	}
	return g.Alert, nil
}

type TailBound struct {
	Split, Scenario, Arm     string
	Difference, Lower, Upper float64
}
type TailOutput struct {
	Output
	MissBounds []TailBound
}

func SummarizeTail(o *TailOutput) {
	Summarize(&o.Output)
	o.MissBounds = nil
	names := []string{"fixed", "grid", "anchored90", "anchored50"}
	for i := range o.Summary {
		o.Summary[i].Arm = names[i%4]
	}
	for _, split := range []string{"design", "confirmation"} {
		for j := 5; j < len(Names); j++ {
			for arm := 1; arm < 4; arm++ {
				var ds []float64
				mean := 0.
				for _, r := range o.Records {
					if r.Split != split || r.Scenario != Names[j] {
						continue
					}
					d := 0.
					if r.First[arm] < change(j) {
						d++
					}
					if r.First[0] < change(j) {
						d--
					}
					ds = append(ds, d)
					mean += d
				}
				mean /= float64(len(ds))
				v := 0.
				for _, d := range ds {
					v += (d - mean) * (d - mean)
				}
				rad := 3.3 * math.Sqrt(v/float64(len(ds)-1)/float64(len(ds)))
				o.MissBounds = append(o.MissBounds, TailBound{split, Names[j], names[arm], mean, mean - rad, mean + rad})
				if split == "confirmation" {
					o.Pass[arm-1] = o.Pass[arm-1] && mean <= .01 && mean+rad <= .02
				}
			}
		}
	}
}

func RunTail() TailOutput {
	o := TailOutput{}
	for k, split := range []string{"design", "confirmation"} {
		for j, name := range Names {
			for stream := 0; stream < 512; stream++ {
				seed := int64(2026103201+k)*1000000 + int64(j*1000+stream)
				rng := rand.New(rand.NewSource(seed))
				g := TailGate{}
				r := Record{Split: split, Scenario: name, Seed: seed, First: [4]int{-1, -1, -1, -1}}
				previous := 0
				for t := 0; t < 512; t++ {
					d := difference(j, t, previous, rng)
					previous = d
					r.Differences = append(r.Differences, d)
					a, _ := g.Observe(float64(d))
					for arm, v := range a {
						if v && r.First[arm] < 0 {
							r.First[arm] = t
						}
					}
				}
				o.Records = append(o.Records, r)
			}
		}
	}
	SummarizeTail(&o)
	return o
}
