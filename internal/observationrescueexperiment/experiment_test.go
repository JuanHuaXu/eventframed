package observationrescueexperiment

import (
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
)

func TestEvaluationSeedDomainsDoNotOverlap(t *testing.T) {
	seen := map[int64]bool{FitSeed: true}
	for _, base := range []int64{DesignSeed, ConfirmationSeed} {
		for scenario := range Scenarios {
			for stream := 0; stream < Streams; stream++ {
				for role := int64(0); role < 3; role++ {
					seed := base*1000000 + int64(scenario)*100000 + int64(stream)*100 + role
					if seen[seed] {
						t.Fatalf("reused seed %d", seed)
					}
					seen[seed] = true
				}
			}
		}
	}
}

func TestCompleteStreamHasPairedPreOutcomeTraces(t *testing.T) {
	base, err := Base(false)
	if err != nil {
		t.Fatal(err)
	}
	r, err := RunStream(base, "unit", "stable", 0, 177)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Outcomes) != Steps || len(r.Policies) != len(observationrescue.Arms) {
		t.Fatal("incomplete stream")
	}
	for _, p := range r.Policies {
		if len(p.Ticks) != Steps || p.Full.N != Steps || p.Post.N != 256 {
			t.Fatal("incorrect denominator")
		}
		for _, tick := range p.Ticks {
			if tick.P <= 0 || tick.P >= 1 || tick.Cost > 6 {
				t.Fatal("invalid forecast or cost")
			}
		}
	}
}
func BenchmarkRecoveryStep(b *testing.B) {
	base, err := Base(false)
	if err != nil {
		b.Fatal(err)
	}
	reader := observationexperiment.Frames(511, "benchmark")
	for _, arm := range []string{"frozen", "cp_audit", "ap_audit"} {
		b.Run(arm, func(b *testing.B) {
			s, _ := observationrescue.New(base, arm)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f, err := s.Predict(reader, i)
				if err != nil {
					b.Fatal(err)
				}
				_, err = s.Observe(observationrescue.Feedback{Sequence: i, Epoch: f.Epoch, ReferenceCorrect: true, LiveCorrect: true, ValidationEligible: true})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
func BenchmarkAuditFit128(b *testing.B) {
	samples := make([]observation.Sample, 128)
	for i := range samples {
		samples[i] = observation.Sample{Bits: uint16(i), Outcome: i%2 == 0}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := observation.Fit(samples); err != nil {
			b.Fatal(err)
		}
	}
}
