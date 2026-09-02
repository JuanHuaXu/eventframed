package epistemic

import (
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestRepeatedConversationIgnoresRunSessionAndTimestamp(t *testing.T) {
	first := evidenceFixture()
	second := first
	second.ID, second.SessionID, second.Provenance.RunID = "second", "other-session", "other-run"
	first.Provenance.RetrievedIDs = []string{"source-a"}
	second.Provenance.RetrievedIDs = []string{"source-b"}
	second.When.Value = first.When.Value + " later"

	left, right := Describe(first), Describe(second)
	if left.EvidenceGroupKey != right.EvidenceGroupKey {
		t.Fatalf("repeated conversation became independent: %q != %q", left.EvidenceGroupKey, right.EvidenceGroupKey)
	}
}

func TestConversationCannotDeclareTimestampAsIndependentOccurrence(t *testing.T) {
	first := evidenceFixture()
	first.Kind = "conversation_message"
	first.What.Source = model.SourceObserved
	first.Attributes = map[string]string{DistinctOccurrenceAttribute: "true"}
	second := first
	second.ID = "second"
	second.When.Value = "2026-09-02T11:00:00Z"
	if Describe(first).EvidenceGroupKey != Describe(second).EvidenceGroupKey {
		t.Fatal("conversation timestamp bypassed evidence grouping")
	}
}

func TestSourceLineageIgnoresSourceIDOrdering(t *testing.T) {
	first := evidenceFixture()
	first.Kind = "external_observation"
	first.Provenance.SourceEventIDs = []string{"source-b", "source-a"}
	second := first
	second.ID = "second"
	second.Provenance.SourceEventIDs = []string{"source-a", "source-b"}
	if Describe(first).LineageKey != Describe(second).LineageKey {
		t.Fatal("source identifier ordering changed evidence lineage")
	}
}

func TestDeclaredObservedOccurrencesRemainDistinct(t *testing.T) {
	first := evidenceFixture()
	first.Kind = "physical_observation"
	first.What.Source = model.SourceObserved
	first.Attributes = map[string]string{DistinctOccurrenceAttribute: "true"}
	second := first
	second.ID = "second"
	second.When.Value = "2026-09-02T11:00:00Z"

	if Describe(first).EvidenceGroupKey == Describe(second).EvidenceGroupKey {
		t.Fatal("separate declared occurrences collapsed into one evidence group")
	}
}

func TestTimestampDoesNotGrantIndependenceWithoutDeclaration(t *testing.T) {
	first := evidenceFixture()
	first.Kind = "conversation_message"
	first.What.Source = model.SourceObserved
	second := first
	second.ID = "second"
	second.When.Value = "2026-09-02T11:00:00Z"

	if Describe(first).EvidenceGroupKey != Describe(second).EvidenceGroupKey {
		t.Fatal("ordinary message timestamp granted evidence independence")
	}
}

func TestNearDuplicateClaimsFromSameLineageAreCorrelated(t *testing.T) {
	first := evidenceFixture()
	second := first
	second.ID = "second"
	second.What.Value = first.What.Value + " today"
	if !Correlated(Describe(first), Describe(second), .80) {
		t.Fatal("near-duplicate claim from one lineage was treated as independent")
	}
	second.Provenance.Producer = "independent-observer"
	if Correlated(Describe(first), Describe(second), .80) {
		t.Fatal("different provenance lineages were collapsed")
	}
}

func evidenceFixture() model.Event {
	now := time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)
	field := func(value string, source model.FieldSource) model.Field {
		return model.Field{Value: value, Source: source, Confidence: 1}
	}
	return model.Event{
		ID: "first", TenantID: "tenant", SessionID: "session", Kind: "agent_turn",
		OccurredAt: now, ObservedAt: now, AvailableAt: now,
		Who: field("user and agent", model.SourceInferred), What: field("request: capital; outcome: Paris", model.SourceInferred),
		Where: field("session:session", model.SourceObserved), When: field(now.Format(time.RFC3339), model.SourceObserved),
		Why: field("answer request", model.SourceInferred), How: field("agent response", model.SourceInferred),
		Provenance: model.Provenance{Producer: "openclaw-eventframe-memory", RunID: "run-a"},
	}
}
