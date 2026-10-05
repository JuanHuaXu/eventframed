package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"testing"
)

// This audits tapes and as-of provenance, not an independent forecast replay.
func TestSpikeIndependentSourceAudit(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SPIKE_INDEPENDENT_AUDIT")
	if path == "" {
		t.Skip("explicit completed source required")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var header struct {
		softV120Artifact
		Cohort string
	}
	if err := d.Decode(&header); err != nil {
		t.Fatal(err)
	}
	if header.Cohort != "spike-independent-v1" || header.Version != "soft-learners-v120" || header.TransferBase != spikeIndependentTransfer || header.BooleanBase != spikeIndependentBoolean || header.Workers != 1 {
		t.Fatal("header")
	}
	for path, want := range header.Hashes {
		raw, err := os.ReadFile("../../" + path)
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(raw)) != want {
			t.Fatal("source changed since collection", path)
		}
	}
	n := 0
	for {
		var r struct {
			softV120Record
			Change     int
			ChangeSeed int64
		}
		err := d.Decode(&r)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if !spikeIndependentIdentity(n, r.softV120Record) {
			t.Fatal("order", n)
		}
		data, change, seed, err := spikeIndependentData(r.Phase, r.Case, r.Index)
		if err != nil {
			t.Fatal(err)
		}
		if r.Change != change || r.ChangeSeed != seed || r.Seeds != data.Seeds || r.Rules != data.Rules || !reflect.DeepEqual(r.Teacher, data.Teacher) || r.Initial != data.Initial || len(r.Steps) != 256 || len(r.Fits) != 8 {
			t.Fatal("generator metadata", n)
		}
		for i, s := range r.Steps {
			f := data.Frames[i]
			delay, missing := 0, false
			if r.Schedule == 1 {
				delay, missing = int(f.Delay), f.Missing
			}
			if s.X != f.X || s.Y != f.Y || s.Q != data.Q[i] || s.Delay != delay || s.Missing != missing {
				t.Fatal("source tape", n, i)
			}
			for _, p := range s.P {
				if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
					t.Fatal("forecast range", n, i)
				}
			}
		}
		for b, fit := range r.Fits {
			if fit.Clock != b*32 {
				t.Fatal("fit clock", n, b)
			}
			for window, cap := range []int{64, 32} {
				var want []int
				for i := -16; i < fit.Clock; i++ {
					if i < 0 || !r.Steps[i].Missing && i+r.Steps[i].Delay <= fit.Clock {
						want = append(want, i)
					}
				}
				if len(want) > cap {
					want = want[len(want)-cap:]
				}
				if !reflect.DeepEqual(want, fit.Origins[window]) {
					t.Fatal("fit origins", n, b, window)
				}
			}
		}
		n++
	}
	if n != 672 {
		t.Fatal("source count", n)
	}
	t.Log("audited", n, "records; source hashes, metadata, generated tapes, forecast ranges and as-of origins; forecasts not replayed")
}
