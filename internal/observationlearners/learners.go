// Package observationlearners contains bounded research-only challenger models.
package observationlearners

import (
	"math"
	"math/rand"
)

type lossPoint struct {
	Value float64
	Time  int
}
type Window struct {
	Points             []lossPoint
	Seen, Cutoff, Cuts int
}

// Add uses only losses from previously journaled forecasts with available labels.
// This explicit capped scan is ADWIN-inspired, not ADWIN2 or a certificate.
func (w *Window) Add(loss float64, origin int) bool {
	w.Points = append(w.Points, lossPoint{loss, origin})
	if len(w.Points) > 256 {
		w.Points = w.Points[1:]
	}
	w.Seen++
	if w.Seen%8 != 0 || len(w.Points) < 32 {
		return false
	}
	n := len(w.Points)
	total := 0.
	for _, p := range w.Points {
		total += p.Value
	}
	sum, best, cut := 0., 0., 0
	for i, p := range w.Points {
		sum += p.Value
		k := i + 1
		if k < 16 || n-k < 16 {
			continue
		}
		gap := math.Abs(sum/float64(k) - (total-sum)/float64(n-k))
		bound := math.Sqrt(.5 * (1/float64(k) + 1/float64(n-k)) * math.Log(4*float64(n)/.01))
		if gap-bound > best {
			best = gap - bound
			cut = k
		}
	}
	if cut == 0 {
		return false
	}
	w.Points = append([]lossPoint(nil), w.Points[cut:]...)
	w.Cutoff = w.Points[0].Time
	w.Cuts++
	return true
}

type node struct {
	N, Yes      int
	Counts      [9][2][2]int
	Fields      []int
	Feature     int
	Left, Right *node
	Depth       int
	Used        uint16
}
type Forest struct {
	Trees   [5]*node
	rng     *rand.Rand
	Updates int
}

func NewForest(seed int64) *Forest {
	f := &Forest{rng: rand.New(rand.NewSource(seed))}
	for i := range f.Trees {
		f.Trees[i] = f.newNode(0, 0)
	}
	return f
}
func (f *Forest) newNode(depth int, used uint16) *node {
	n := &node{Feature: -1, Depth: depth, Used: used}
	for _, j := range f.rng.Perm(9) {
		if used&(1<<j) == 0 {
			n.Fields = append(n.Fields, j)
			if len(n.Fields) == 6 {
				break
			}
		}
	}
	return n
}
func entropy(n, yes int) float64 {
	if n == 0 || yes == 0 || yes == n {
		return 0
	}
	p := float64(yes) / float64(n)
	return -p*math.Log2(p) - (1-p)*math.Log2(1-p)
}
func (f *Forest) update(n *node, x uint16, y bool) {
	if n.Feature >= 0 {
		if x&(1<<n.Feature) == 0 {
			f.update(n.Left, x, y)
		} else {
			f.update(n.Right, x, y)
		}
		return
	}
	n.N++
	c := 0
	if y {
		c = 1
		n.Yes++
	}
	for _, j := range n.Fields {
		v := 0
		if x&(1<<j) != 0 {
			v = 1
		}
		n.Counts[j][v][c]++
	}
	if n.Depth >= 4 || n.N < 32 || n.N%16 != 0 {
		return
	}
	best, second, feature := 0., 0., -1
	for _, j := range n.Fields {
		a, b := n.Counts[j][0], n.Counts[j][1]
		n0, n1 := a[0]+a[1], b[0]+b[1]
		if n0 == 0 || n1 == 0 {
			continue
		}
		gain := entropy(n.N, n.Yes) - float64(n0)/float64(n.N)*entropy(n0, a[1]) - float64(n1)/float64(n.N)*entropy(n1, b[1])
		if gain > best {
			second = best
			best = gain
			feature = j
		} else if gain > second {
			second = gain
		}
	}
	epsilon := math.Sqrt(math.Log(1/.05) / (2 * float64(n.N)))
	if feature < 0 || best <= 0 || (best-second <= epsilon && epsilon >= .1) {
		return
	}
	n.Feature = feature
	used := n.Used | uint16(1<<feature)
	n.Left = f.newNode(n.Depth+1, used)
	n.Right = f.newNode(n.Depth+1, used)
}
func (f *Forest) Update(x uint16, y bool) {
	f.Updates++
	for _, n := range f.Trees {
		product, k := 1., 0
		for product > math.Exp(-1) && k <= 8 {
			k++
			product *= f.rng.Float64()
		}
		copies := min(8, k-1)
		for i := 0; i < copies; i++ {
			f.update(n, x, y)
		}
	}
}
func (f *Forest) Predict(x uint16) float64 {
	p := 0.
	for _, n := range f.Trees {
		for n.Feature >= 0 {
			if x&(1<<n.Feature) == 0 {
				n = n.Left
			} else {
				n = n.Right
			}
		}
		p += float64(n.Yes+1) / float64(n.N+2) / 5
	}
	return p
}
func (f *Forest) Nodes() int {
	var count func(*node) int
	count = func(n *node) int {
		if n.Feature < 0 {
			return 1
		}
		return 1 + count(n.Left) + count(n.Right)
	}
	n := 0
	for _, t := range f.Trees {
		n += count(t)
	}
	return n
}
