package observationgate

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"unsafe"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

func BenchmarkWindowCouplingRun(b *testing.B) {
	// Benchmark one existing reverse-shift trajectory, not a fresh quality test.
	load := func(name string, dst any) *json.Decoder {
		f, e := os.Open("../../docs/experiments/" + name)
		if e != nil {
			b.Fatal(e)
		}
		b.Cleanup(func() { f.Close() })
		d := json.NewDecoder(f)
		var h json.RawMessage
		if e = d.Decode(&h); e != nil {
			b.Fatal(e)
		}
		return d
	}
	pd := load("mmm-credit-learning-v1.jsonl", nil)
	dd := load("mmm-available-evidence-v1-verified.jsonl", nil)
	var p creditLearningRecord
	var d availableEvidenceRecord
	for {
		if e := pd.Decode(&p); e != nil {
			b.Fatal(e)
		}
		if e := dd.Decode(&d); e != nil {
			b.Fatal(e)
		}
		if p.Case == "parity_to_majority" && p.Index == 0 {
			break
		}
	}
	base, _, e := arrivalSwitchSetup(0, 3, 0)
	if e != nil {
		b.Fatal(e)
	}
	for s := 0; s < 2; s++ {
		b.Run(fmt.Sprint(s), func(b *testing.B) {
			parent, advice := p.Immediate, d.Immediate
			if s == 1 {
				parent, advice = p.Delayed, d.Delayed
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r, e := windowCouplingRun(base, parent, advice, nil)
				if e != nil || len(r.Frames) != 512 {
					b.Fatal(e)
				}
			}
		})
	}
}

func TestWindowCouplingRetainedTables(t *testing.T) {
	// These immutable tables can be shared by both arms. This is table storage,
	// not process RSS, peak transient allocation or publication/journal overhead.
	bytes := 5*unsafe.Sizeof(observation.Model{}) + 2*unsafe.Sizeof(observationlearners.ConditionalForest{})
	t.Logf("retained base/label/event/local/pool plus two subset tables: %d bytes", bytes)
}
