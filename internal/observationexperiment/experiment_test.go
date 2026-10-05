package observationexperiment

import (
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func TestFrameExtractionIgnoresRawAndInferredData(t *testing.T) {
	r := Frames(511, "test")
	for scope := 0; scope < 3; scope++ {
		mask, value, err := r.Read(observation.View{Scope: scope, Depth: 2})
		if err != nil {
			t.Fatal(err)
		}
		if mask != uint16(7<<(scope*3)) || value != mask {
			t.Fatal("incorrect scope extraction")
		}
	}
	r.Frames[0][1].Content = "outcome is false"
	r.Frames[0][1].Why.Value = "0"
	_, v, err := r.Read(observation.View{Scope: 0, Depth: 2})
	if err != nil || v != 7 {
		t.Fatal("raw or inferred contamination")
	}
	r.Frames[0][1].How.Source = model.SourceInferred
	mask, _, err := r.Read(observation.View{Scope: 0, Depth: 2})
	if err != nil || mask != 5 {
		t.Fatal("inferred field counted")
	}
}
func TestFutureEvidenceAndTenant(t *testing.T) {
	r := Frames(511, "test")
	r.Frames[0][1].AvailableAt = r.AsOf.Add(time.Second)
	mask, value, err := r.Read(observation.View{Scope: 0, Depth: 2})
	if err != nil || mask != 4 || value != 4 {
		t.Fatal("future evidence visible")
	}
	r.Frames[0][0].TenantID = "other"
	if _, _, err := r.Read(observation.View{Scope: 0, Depth: 2}); err == nil {
		t.Fatal("cross tenant accepted")
	}
}
func TestEndToEndControllerOnEventFrames(t *testing.T) {
	m, err := FitFamily(2)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range observation.Policies {
		r, err := observation.Run(m, Frames(127, "e2e"), 1, p, 7)
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Trace) == 0 || r.Probability <= 0 || r.Probability >= 1 {
			t.Fatal("invalid result")
		}
	}
	if m.Support() != 4096 {
		t.Fatal("support changed")
	}
}

func TestRejectsReversedAndDuplicatedFrames(t *testing.T) {
	r := Frames(511, "test")
	r.Frames[0][0], r.Frames[0][1] = r.Frames[0][1], r.Frames[0][0]
	if _, _, err := r.Read(observation.View{Scope: 0, Depth: 2}); err == nil {
		t.Fatal("reversed temporal relation accepted")
	}
	r = Frames(511, "test")
	r.Frames[1][0].ID = r.Frames[0][0].ID
	if _, _, err := r.Read(observation.View{Scope: 0, Depth: 0}); err == nil {
		t.Fatal("duplicate observation source accepted")
	}
}

func BenchmarkObservation(b *testing.B) {
	m, err := FitFamily(3)
	if err != nil {
		b.Fatal(err)
	}
	r := Frames(170, "benchmark")
	for _, policy := range []string{"fixed", "breadth", "depth_first", "mmm", "exhaustive"} {
		b.Run(policy, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := observation.Run(m, r, 1, policy, 0); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
