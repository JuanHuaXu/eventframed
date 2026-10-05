package observationlearners

import (
	"encoding/json"
	"io"
	"math"
	"os"
	"testing"
)

func TestPopulationQueryKernel(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("explicit source")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var header softV120Artifact
	if err := d.Decode(&header); err != nil {
		t.Fatal(err)
	}
	fixtures, checks := 0, 0
	maxError := 0.
	for {
		var input softV120Record
		err := d.Decode(&input)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if input.Index != 0 {
			continue
		}
		o, err := populationOracle(input)
		if err != nil {
			t.Fatal(err)
		}
		k, err := makePopulationKernel(o)
		if err != nil {
			t.Fatal(err)
		}
		for j, s := range input.Steps {
			q, err := o.truth(s.X, j)
			if err != nil || math.Abs(q-s.Q) > 1e-14 {
				t.Fatal("stored truth mismatch", input.Case, j)
			}
		}
		for variant := 0; variant < 4; variant++ {
			var p [512]float64
			for x := range p {
				switch variant {
				case 0:
					p[x] = .5
				case 1:
					p[x] = float64(x) / 511
				case 2:
					p[x] = float64((x*73+19)%512) / 511
				case 3:
					p[x] = float64(x % 2)
				}
			}
			got, err := k.risk(p)
			if err != nil {
				t.Fatal(err)
			}
			want := 0.
			for i := 0; i < 31; i++ {
				for _, x := range o.inputMap {
					q, err := o.truth(x, 161+i)
					if err != nil {
						t.Fatal(err)
					}
					v := .5 + math.Pow(.99, float64(i))*(p[x]-.5)
					want += ((v-q)*(v-q) + q*(1-q)) / (31 * 512)
				}
			}
			e := math.Abs(got - want)
			if e > 1e-11 {
				t.Fatal("direct integration mismatch", input.Case, variant, e)
			}
			maxError = math.Max(maxError, e)
			if variant == 0 && math.Abs(got-.25) > 1e-12 {
				t.Fatal("constant half invariant")
			}
			checks++
		}
		fixtures++
	}
	if fixtures != 84 || checks != 336 {
		t.Fatal("coverage", fixtures, checks)
	}
	// A deliberately nonuniform pushforward: half of raw inputs map to zero.
	o := populationQueryOracle{truth: func(uint16, int) (float64, error) { return 0, nil }}
	for i := 256; i < 512; i++ {
		o.inputMap[i] = uint16(i)
	}
	k, err := makePopulationKernel(o)
	if err != nil {
		t.Fatal(err)
	}
	var p [512]float64
	p[0] = 1
	got, err := k.risk(p)
	if err != nil {
		t.Fatal(err)
	}
	want := 0.
	for i := 0; i < 31; i++ {
		a := math.Pow(.99, float64(i))
		want += (.5*math.Pow(.5+.5*a, 2) + .5*math.Pow(.5-.5*a, 2)) / 31
	}
	if math.Abs(got-want) > 1e-12 {
		t.Fatal("duplicate mass")
	}
	p[0] = math.NaN()
	if _, err := k.risk(p); err == nil {
		t.Fatal("NaN accepted")
	}
	if _, err := populationOracle(softV120Record{Case: 22}); err == nil {
		t.Fatal("identity accepted")
	}
	if _, err := populationOracle(softV120Record{Case: 9, Rules: [2]uint16{511, 511}}); err == nil {
		t.Fatal("forged rule accepted")
	}
	t.Logf("84 identity/schedule fixtures,336 direct integrals,duplicate weighting and source truth; max discrepancy %.3g", maxError)
}
