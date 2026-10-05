package observationgate

import (
	"encoding/json"
	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"os"
	"testing"
)

func TestTrackingMixReference(t *testing.T) {
	path := os.Getenv("EVENTFRAME_TRACKING_REFERENCE")
	if path == "" {
		t.Skip("opt-in primitive reference")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	s := bayes.ForecastMix{Weights: [4]float64{.7, .1, .1, .1}}
	for i := 0; i < 128; i++ {
		var experts [4]float64
		for j := range experts {
			experts[j] = float64((i*17+j*29)%101) / 100
		}
		y := i%3 != 0
		before := s.Weights
		p := s.Forecast(experts)
		s = s.Observe(experts, y, 1)
		if err := enc.Encode(map[string]any{"Experts": experts, "Before": before, "P": p, "After": s.Weights, "Y": y}); err != nil {
			t.Fatal(err)
		}
	}
}
