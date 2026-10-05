package observationlearners

import (
	"math"
	"testing"
)

func TestMixtureOracleAnalytic(t *testing.T) {
	p, err := makeMixtureProfile([][5]float64{{.1, .9, .1, .9, .5}}, []float64{.7})
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{2, 4, 5} {
		x, err := p.minimum(n)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(x.Risk-.21) > 1e-12 || x.Lower > .21+1e-12 || x.Gap > 1e-10 {
			t.Fatal("analytic mixture", x)
		}
	}
	// All identical forecasts leave singular interior faces. A vertex is enough.
	p, _ = makeMixtureProfile([][5]float64{{.3, .3, .3, .3, .3}}, []float64{.9})
	x, err := p.minimum(5)
	if err != nil || math.Abs(x.Risk-.45) > 1e-12 {
		t.Fatal("singular profile", x, err)
	}
}

func TestMixtureOracleInterior(t *testing.T) {
	w := [5]float64{.1, .15, .2, .25, .3}
	rows := make([][5]float64, 32)
	truth := make([]float64, 32)
	floor := 0.
	for x := range rows {
		for j := range w {
			rows[x][j] = .1
			if x&(1<<j) != 0 {
				rows[x][j] = .9
			}
			truth[x] += w[j] * rows[x][j]
		}
		floor += truth[x] * (1 - truth[x]) / 32
	}
	p, err := makeMixtureProfile(rows, truth)
	if err != nil {
		t.Fatal(err)
	}
	got, err := p.minimum(5)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got.Risk-floor) > 1e-12 {
		t.Fatal("Bayes floor", got, floor)
	}
	for j := range w {
		if math.Abs(got.Weights[j]-w[j]) > 1e-10 {
			t.Fatal("interior optimizer")
		}
	}
}

func TestMixtureOracleBoundaryAndGrid(t *testing.T) {
	rows := [][5]float64{{.1, .3, .7, .9, .5}, {.4, .8, .2, .6, .5}}
	truth := []float64{.9, .6}
	p, _ := makeMixtureProfile(rows, truth)
	got, err := p.minimum(4)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got.Weights[3]-1) > 1e-10 {
		t.Fatal("vertex optimum", got)
	}
	for i := 0; i <= 50; i++ {
		for j := 0; j <= 50-i; j++ {
			for k := 0; k <= 50-i-j; k++ {
				w := [5]float64{float64(i) / 50, float64(j) / 50, float64(k) / 50, float64(50-i-j-k) / 50, 0}
				direct := 0.
				for row, q := range truth {
					forecast := 0.
					for m := range w {
						forecast += w[m] * rows[row][m]
					}
					direct += ((forecast-q)*(forecast-q) + q*(1-q)) / 2
				}
				if direct < got.Risk-1e-12 || direct < got.Lower-1e-12 {
					t.Fatal("grid beats optimizer")
				}
			}
		}
	}
	if _, err := makeMixtureProfile(rows, []float64{math.NaN(), .5}); err == nil {
		t.Fatal("NaN admitted")
	}
	if _, err := p.minimum(6); err == nil {
		t.Fatal("unbounded dimension")
	}
}
