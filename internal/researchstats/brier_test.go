package researchstats

import (
	"math"
	"testing"
)

func TestPairedBrierMean(t *testing.T) {
	cases := []struct {
		name  string
		pairs []BrierPair
		want  float64
	}{
		{"positive endpoint", []BrierPair{{Control: 1, Candidate: 0}}, 1},
		{"negative endpoint", []BrierPair{{Control: 0, Candidate: 1}}, -1},
		{"positive binary one", []BrierPair{{Control: 0, Candidate: 1, Outcome: true}}, 1},
		{"negative binary one", []BrierPair{{Control: 1, Candidate: 0, Outcome: true}}, -1},
		{"null", []BrierPair{{Control: .7, Candidate: .7, Outcome: true}}, 0},
		{"paired average", []BrierPair{
			{Control: .5, Candidate: .75, Outcome: true},
			{Control: .5, Candidate: .25},
		}, .1875},
		{"opposing gains", []BrierPair{{Control: 1}, {Candidate: 1}}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := PairedBrierMean(c.pairs)
			if err != nil {
				t.Fatal(err)
			}
			near(t, got, c.want)
		})
	}
}

func TestInvalidPairedStreamAtomicity(t *testing.T) {
	s := newSequence(t, .05, "gain")
	if err := s.AddStream("gain", .1); err != nil {
		t.Fatal(err)
	}
	before, _ := s.Anytime("gain")
	for _, pairs := range [][]BrierPair{nil, {}} {
		if _, err := PairedBrierMean(pairs); err == nil {
			t.Fatal("empty stream accepted")
		}
		if err := s.AddPairedStream("gain", pairs); err == nil {
			t.Fatal("empty stream added")
		}
	}
	for _, bad := range []float64{-1, math.Nextafter(0, -1), math.Nextafter(1, 2), math.NaN(), math.Inf(1), math.Inf(-1)} {
		for _, pair := range []BrierPair{{Control: bad}, {Candidate: bad}} {
			// Invalid late data must not partially add a stream.
			pairs := []BrierPair{{Control: .5, Candidate: .25}, pair}
			if _, err := PairedBrierMean(pairs); err == nil {
				t.Fatalf("accepted invalid pair: %+v", pair)
			}
			if err := s.AddPairedStream("gain", pairs); err == nil {
				t.Fatal("invalid stream added")
			}
			after, _ := s.Anytime("gain")
			if after != before {
				t.Fatal("rejected paired stream changed CS")
			}
		}
	}
	if err := s.AddPairedStream("new comparison", []BrierPair{{}}); err == nil {
		t.Fatal("paired stream expanded family")
	}
}

func TestStreamWeightingNotClockWeighting(t *testing.T) {
	s := newSequence(t, .05, "gain")
	long := make([]BrierPair, 1000)
	for i := range long {
		long[i] = BrierPair{Control: 1}
	}
	if err := s.AddPairedStream("gain", long); err != nil {
		t.Fatal(err)
	}
	if err := s.AddPairedStream("gain", []BrierPair{{Candidate: 1}}); err != nil {
		t.Fatal(err)
	}
	i, _ := s.Anytime("gain")
	if i.N != 2 || i.Mean != 0 {
		t.Fatalf("counted clocks instead of equally weighted streams: %+v", i)
	}
}

func TestSyntheticPairedBrierStreams(t *testing.T) {
	// The law is a deterministic two-clock pattern, with both forecasts and
	// outcomes declared as a fixture. Every stream mean is exactly .1875.
	// This checks arithmetic and every prefix, not coverage for arbitrary laws.
	stream := []BrierPair{
		{Control: .5, Candidate: .75, Outcome: true},
		{Control: .5, Candidate: .25, Outcome: false},
	}
	s := newSequence(t, .05, "synthetic-benefit")
	var interval Interval
	for j := uint64(1); j <= 4096; j++ {
		if err := s.AddPairedStream("synthetic-benefit", stream); err != nil {
			t.Fatal(err)
		}
		var err error
		interval, err = s.Anytime("synthetic-benefit")
		if err != nil || interval.N != j || interval.Mean != .1875 ||
			interval.Lower > .1875 || interval.Upper < .1875 {
			t.Fatalf("synthetic prefix check failed: %+v %v", interval, err)
		}
	}
	if interval.Lower <= 0 {
		t.Fatal("synthetic positive-gain endpoint did not exclude zero")
	}
	logJSONRecord(t, "synthetic_interval", map[string]any{
		"scope":        "deterministic utility check, not scientific confirmation",
		"alpha_family": .05, "family_size": 1, "stream_count": interval.N,
		"clocks_per_stream": len(stream), "known_constant_mean": .1875,
		"mean": interval.Mean, "halfwidth": interval.Radius,
		"lower": interval.Lower, "upper": interval.Upper,
		"every_prefix_contains_known_mean": true, "lower_above_zero": true,
	})
}
