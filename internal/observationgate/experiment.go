package observationgate

import (
	"math"
	"math/rand"
)

var Names = []string{"symmetric", "sparse", "boundary_low", "boundary_high", "dependent", "moderate128", "moderate256", "strong256", "negative256", "weak256"}
var Arms = []string{"fixed", "grid", "adaptive", "hedge"}

type Record struct {
	Split, Scenario string
	Seed            int64
	Differences     []int
	First           [4]int
}
type Summary struct {
	Split, Scenario                                string
	Arm                                            string
	N, Alerts, Premature, Detected, Missed         int
	AlertUpper, RestrictedDelay, DetectedOnlyDelay float64
	Gain, Lower, Upper                             float64
}
type Output struct {
	Hashes  map[string]string
	Records []Record
	Summary []Summary
	Pass    [3]bool
}

func change(j int) int {
	if j < 5 {
		return 512
	}
	if j == 5 {
		return 128
	}
	return 256
}
func difference(j, t, previous int, r *rand.Rand) int {
	p, m := .05, .05
	switch j {
	case 0:
		p, m = .5, .5
	case 2:
		p, m = .2, .05
	case 3:
		p, m = .575, .425
	case 4:
		if previous == 0 {
			p, m = .5, .5
		} else {
			p, m = 0, 0
		}
	default:
		if t >= change(j) {
			switch j {
			case 5, 6:
				p, m = .5, .1
			case 7:
				p, m = .75, .05
			case 8:
				p, m = .1, .5
			case 9:
				p, m = .25, .05
			}
		}
	}
	u := r.Float64()
	if u < p {
		return 1
	}
	if u < p+m {
		return -1
	}
	return 0
}

func Run() Output {
	o := Output{}
	for k, split := range []string{"design", "confirmation"} {
		for j, name := range Names {
			for stream := 0; stream < 512; stream++ {
				seed := (int64(2026092501+k))*1000000 + int64(j*1000+stream)
				r := rand.New(rand.NewSource(seed))
				g := Gate{}
				rec := Record{Split: split, Scenario: name, Seed: seed, First: [4]int{-1, -1, -1, -1}}
				previous := 0
				for t := 0; t < 512; t++ {
					d := difference(j, t, previous, r)
					previous = d
					rec.Differences = append(rec.Differences, d)
					a, _ := g.Observe(float64(d))
					for arm, v := range a {
						if v && rec.First[arm] < 0 {
							rec.First[arm] = t
						}
					}
				}
				o.Records = append(o.Records, rec)
			}
		}
	}
	Summarize(&o)
	return o
}

func restricted(first, at int) float64 {
	if first < at {
		return float64(512 - at)
	}
	return float64(first - at)
}
func wilson(k, n int) float64 {
	z := 1.959963984540054
	nn := float64(n)
	p := float64(k) / nn
	return (p + z*z/(2*nn) + z*math.Sqrt(p*(1-p)/nn+z*z/(4*nn*nn))) / (1 + z*z/nn)
}

func Summarize(o *Output) {
	o.Summary = nil
	o.Pass = [3]bool{true, true, true}
	for _, split := range []string{"design", "confirmation"} {
		for j, name := range Names {
			var rows [4]Summary
			for arm := range Arms {
				s := Summary{Split: split, Scenario: name, Arm: Arms[arm]}
				var gains []float64
				for _, r := range o.Records {
					if r.Split != split || r.Scenario != name {
						continue
					}
					s.N++
					f := r.First[arm]
					at := change(j)
					if f >= 0 {
						s.Alerts++
					}
					if j >= 5 {
						if f >= 0 && f < at {
							s.Premature++
						}
						if f >= at {
							s.Detected++
							s.DetectedOnlyDelay += float64(f - at)
						} else {
							s.Missed++
						}
						d := restricted(f, at)
						s.RestrictedDelay += d
						g := restricted(r.First[0], at) - d
						gains = append(gains, g)
						s.Gain += g
					}
				}
				s.AlertUpper = wilson(s.Alerts, s.N)
				s.RestrictedDelay /= float64(s.N)
				s.Gain /= float64(s.N)
				if s.Detected > 0 {
					s.DetectedOnlyDelay /= float64(s.Detected)
				} else {
					s.DetectedOnlyDelay = -1
				}
				v := 0.
				for _, g := range gains {
					v += (g - s.Gain) * (g - s.Gain)
				}
				margin := 3.3 * math.Sqrt(v/float64(s.N-1)/float64(s.N))
				s.Lower, s.Upper = s.Gain-margin, s.Gain+margin
				rows[arm] = s
				o.Summary = append(o.Summary, s)
			}
			if split == "confirmation" {
				for arm := 1; arm < 4; arm++ {
					s := rows[arm]
					ok := true
					if j < 5 {
						ok = s.AlertUpper <= .02
					} else if j == 5 || j == 6 {
						ok = s.Gain >= .1*rows[0].RestrictedDelay && s.Lower > 0 && s.Premature <= rows[0].Premature
					} else {
						ok = s.Gain >= -10
					}
					o.Pass[arm-1] = o.Pass[arm-1] && ok
				}
			}
		}
	}
}
