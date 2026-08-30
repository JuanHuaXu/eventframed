package service

import (
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestResolutionPreferenceIsOrderIndependentForNestedCompositions(t *testing.T) {
	inner := model.Candidate{Event: model.Event{ID: "inner", Composition: &model.Composition{MemberEventIDs: []string{"a", "b"}, RepresentativeEventID: "a"}}, Score: .5}
	outer := model.Candidate{Event: model.Event{ID: "outer", Composition: &model.Composition{MemberEventIDs: []string{"inner", "c"}, RepresentativeEventID: "c"}}, Score: .5}
	a := model.Candidate{Event: model.Event{ID: "a"}, Score: .5}
	b := model.Candidate{Event: model.Event{ID: "b"}, Score: .5}
	c := model.Candidate{Event: model.Event{ID: "c"}, Score: .5}
	forward := []model.Candidate{inner, outer, a, b, c}
	reverse := []model.Candidate{outer, inner, a, b, c}
	applyResolutionPreference(forward, model.RecallResolutionCoarse)
	applyResolutionPreference(reverse, model.RecallResolutionCoarse)
	forwardDelta, reverseDelta := map[string]float64{}, map[string]float64{}
	for _, candidate := range forward {
		forwardDelta[candidate.Event.ID] = candidate.ResolutionRankDelta
	}
	for _, candidate := range reverse {
		reverseDelta[candidate.Event.ID] = candidate.ResolutionRankDelta
	}
	for id, want := range forwardDelta {
		if got := reverseDelta[id]; got != want {
			t.Fatalf("%s delta depends on candidate order: forward=%v reverse=%v", id, want, got)
		}
	}
}
