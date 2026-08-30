package model_test

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestFrameTextExcludesRawContentAndIsCanonical(t *testing.T) {
	event := model.Event{
		Content: "RAW_CONTENT_SENTINEL",
		Who:     model.Field{Value: "Juan"}, What: model.Field{Value: "deploy EventFrame"},
		Where: model.Field{Value: "Toronto"}, When: model.Field{Value: "tomorrow"},
		Why: model.Field{Value: "retrieval is stale"}, How: model.Field{Value: "using graph search"},
	}
	want := strings.Join([]string{
		"representation: " + model.SemanticRepresentationVersion,
		"who: Juan", "what: deploy EventFrame", "where: Toronto", "when: tomorrow",
		"why: retrieval is stale", "how: using graph search",
	}, "\n")
	if got := event.FrameText(); got != want {
		t.Fatalf("FrameText() = %q, want %q", got, want)
	}
	if strings.Contains(event.FrameText(), event.Content) {
		t.Fatal("FrameText leaked raw content")
	}
}

func TestCompositionValidationBindsRepresentativeAndExactProvenance(t *testing.T) {
	now := time.Now().UTC()
	event := model.Event{
		ID: "macro", TenantID: "tenant", SessionID: "session", Kind: model.HigherOrderEventKind, Content: "macro",
		OccurredAt: now, ObservedAt: now, AvailableAt: now,
		What:        model.Field{Value: "macro", Source: model.SourceInferred, Confidence: .8},
		Provenance:  model.Provenance{Producer: "invariant", SourceEventIDs: []string{"a", "b"}},
		Composition: &model.Composition{MemberEventIDs: []string{"a", "b"}, RepresentativeEventID: "a", RuleID: "rule", Resolution: "episode", Confidence: .8, AntiPigeonCertificateID: "certificate"},
	}
	if err := event.Validate(8); err != nil {
		t.Fatal(err)
	}
	event.Provenance.SourceEventIDs = []string{"a", "c"}
	if err := event.Validate(8); err == nil {
		t.Fatal("composition accepted provenance outside its member set")
	}
	event.Provenance.SourceEventIDs = []string{"a", "b"}
	event.Composition.RepresentativeEventID = "c"
	if err := event.Validate(8); err == nil {
		t.Fatal("composition accepted a representative outside its member set")
	}
	event.Composition.RepresentativeEventID = "a"
	event.Provenance.SourceEventIDs = []string{"a", "a"}
	if err := event.Validate(8); err == nil {
		t.Fatal("composition accepted duplicate provenance in place of one member")
	}
	event.Provenance.SourceEventIDs = []string{"a", "b"}
	event.Composition.Confidence = math.NaN()
	if err := event.Validate(8); err == nil {
		t.Fatal("composition accepted non-finite confidence")
	}
}

func TestFrameTextIsInvariantToRawPayload(t *testing.T) {
	left := model.Event{Content: "first raw transcript", What: model.Field{Value: "same event"}}
	right := left
	right.Content = "unrelated second transcript"
	if left.FrameText() != right.FrameText() {
		t.Fatal("raw payload changed semantic representation")
	}
}

func TestFrameTextCanonicalizesWhitespace(t *testing.T) {
	event := model.Event{What: model.Field{Value: "deploy\n\tEventFrame   now"}}
	if got := event.FrameText(); !strings.Contains(got, "what: deploy EventFrame now") || strings.Contains(got, "\n\t") {
		t.Fatalf("noncanonical frame text: %q", got)
	}
}
