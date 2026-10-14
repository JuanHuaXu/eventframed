package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTauJSONRequiresExplicitDeclarations(t *testing.T) {
	for _, field := range []string{"tv_radius", "no_event", "horizon_ms"} {
		value := map[string]any{"tv_radius": 0, "no_event": 1, "horizon_ms": 1000}
		delete(value, field)
		data, _ := json.Marshal(value)
		var f TauForecast
		if err := json.Unmarshal(data, &f); err == nil {
			t.Fatalf("omitted %s accepted", field)
		}
		value[field] = nil
		data, _ = json.Marshal(value)
		if err := json.Unmarshal(data, &f); err == nil {
			t.Fatalf("null %s accepted", field)
		}
	}
	var f TauForecast
	if err := json.Unmarshal([]byte(`{"tv_radius":0,"no_event":1,"horizon_ms":1000}`), &f); err != nil {
		t.Fatal(err)
	}
	var point TauPoint
	if err := json.Unmarshal([]byte(`{"mark":"complete","delay_ms":1000}`), &point); err == nil {
		t.Fatal("missing atom probability accepted")
	}
	p := TauTurnPolicy{}
	data, _ := json.Marshal(p)
	for _, field := range []string{"false_cost", "turn_cost", "benefit", "margin"} {
		var raw map[string]any
		_ = json.Unmarshal(data, &raw)
		delete(raw, field)
		missing, _ := json.Marshal(raw)
		if err := json.Unmarshal(missing, &p); err == nil || !strings.Contains(err.Error(), field) {
			t.Fatalf("missing %s: %v", field, err)
		}
	}
}
