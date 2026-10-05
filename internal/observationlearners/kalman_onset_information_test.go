package observationlearners

import (
	"math"
	"math/rand"
	"testing"
)

type onsetPacket struct {
	origin, due int
	y           [2]bool
	audit       bool
	base        float64
	features    [kalmanDim]float64
}

// Matched worlds differ only in the hidden post-change outcome rule. The
// filter may use a label only after its packet becomes due.
func TestKalmanDelayedOnsetInformation(t *testing.T) {
	const change, delay, streams = 256, 16, 32
	stable := Scenario{Name: "shift256", Noise: .05, Change: 512, Delay: delay, Missing: .25}
	shifted := stable
	shifted.Change = change
	base, err := Base(stable, 7, 20)
	if err != nil {
		t.Fatal(err)
	}
	firstAvailableMin, firstAuditedMin := 512, 512
	preEvidenceForecasts, differingPreDeliveryContexts := 0, 0
	for stream := 0; stream < streams; stream++ {
		seed := Seed(2026100101, 7, 20, stream, 0)
		rng := [2]*rand.Rand{rand.New(rand.NewSource(seed)), rand.New(rand.NewSource(seed))}
		auditRNG := rand.New(rand.NewSource(Seed(2026100101, 7, 20, stream, 1)))
		missingRNG := rand.New(rand.NewSource(Seed(2026100101, 7, 20, stream, 2)))
		filters := [2]KalmanResidual{NewKalmanResidual(), NewKalmanResidual()}
		var queue []onsetPacket
		firstAvailable, firstAudited := -1, -1
		auditedWorldsDiffer, forecastsDiverged := false, false
		for clock := 0; clock < 512; clock++ {
			for i := range filters {
				filters[i].Advance()
			}
			x := uint16(rng[0].Intn(512))
			if other := uint16(rng[1].Intn(512)); other != x {
				t.Fatalf("stream %d clock %d: contexts differ", stream, clock)
			}
			pb := forecast(base, x)
			h := kalmanFeatures(x, pb)
			p0, p1 := filters[0].Predict(pb, h), filters[1].Predict(pb, h)
			if auditedWorldsDiffer && math.Float64bits(p0) != math.Float64bits(p1) {
				forecastsDiverged = true
			}
			if !auditedWorldsDiffer && math.Float64bits(p0) != math.Float64bits(p1) {
				t.Fatalf("stream %d clock %d: forecasts differ before distinguishing audited evidence", stream, clock)
			}
			if clock >= change && clock < change+delay {
				preEvidenceForecasts++
				if math.Float64bits(p0) != math.Float64bits(p1) {
					t.Fatalf("stream %d clock %d: pre-delivery forecasts differ", stream, clock)
				}
			}
			audit, missing := auditRNG.Float64() < .25, missingRNG.Float64() < stable.Missing
			y := [2]bool{truth(x, clock, stable, rng[0]), truth(x, clock, shifted, rng[1])}
			if clock >= change && clock < change+delay && y[0] != y[1] {
				differingPreDeliveryContexts++
			}
			if clock < change && y[0] != y[1] {
				t.Fatalf("stream %d clock %d: pre-change outcomes differ", stream, clock)
			}
			if !missing {
				queue = append(queue, onsetPacket{origin: clock, due: clock + delay, y: y,
					audit: audit, base: pb, features: h})
			}
			for len(queue) > 0 && queue[0].due <= clock {
				packet := queue[0]
				queue = queue[1:]
				if packet.origin >= change {
					if firstAvailable < 0 {
						firstAvailable = clock
					}
					if packet.audit && firstAudited < 0 {
						firstAudited = clock
					}
				}
				if packet.audit {
					for i := range filters {
						if err := filters[i].Observe(packet.base, packet.features, packet.y[i]); err != nil {
							t.Fatal(err)
						}
					}
					if packet.y[0] != packet.y[1] {
						auditedWorldsDiffer = true
					}
				}
			}
		}
		if firstAvailable < change+delay || firstAudited < change+delay || !auditedWorldsDiffer || !forecastsDiverged {
			t.Fatalf("stream %d: available=%d audited=%d distinguishing audited evidence=%v later forecast divergence=%v",
				stream, firstAvailable, firstAudited, auditedWorldsDiffer, forecastsDiverged)
		}
		firstAvailableMin = min(firstAvailableMin, firstAvailable)
		firstAuditedMin = min(firstAuditedMin, firstAudited)
	}
	if preEvidenceForecasts != streams*delay {
		t.Fatalf("pre-evidence forecasts=%d, want %d", preEvidenceForecasts, streams*delay)
	}
	if differingPreDeliveryContexts == 0 {
		t.Fatal("matched worlds contain no differing pre-delivery contexts")
	}
	t.Logf("matched streams=%d; identical pre-delivery forecasts=%d; differing contexts=%d; first changed-label availability >=%d; first audited availability >=%d; later filters diverged in every stream",
		streams, preEvidenceForecasts, differingPreDeliveryContexts, firstAvailableMin, firstAuditedMin)
}
