package frame_test

import (
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestFromTurnExtractsContentSpecificFieldsAfterCapture(t *testing.T) {
	turn := capture("Example Operator will deploy EventFrame in Toronto tomorrow because retrieval is stale using the OpenClaw plugin.", "I prepared the release configuration and validation tests.")
	event := frame.FromTurn(turn)

	want := map[string]string{
		"who": "Example Operator", "where": "Toronto", "when": "tomorrow",
		"why": "retrieval is stale", "how": "using the OpenClaw plugin",
	}
	got := map[string]model.Field{"who": event.Who, "where": event.Where, "when": event.When, "why": event.Why, "how": event.How}
	for name, expected := range want {
		if got[name].Value != expected || got[name].Source != model.SourceObserved || !strings.HasPrefix(got[name].Evidence, "user[bytes:") {
			t.Errorf("%s = %+v, want observed %q with user evidence", name, got[name], expected)
		}
	}
	if !strings.Contains(event.What.Value, "deploy EventFrame") || event.Attributes["semantic_extractor"] != "fivew1h-deterministic-v1" {
		t.Fatalf("enriched event = %+v", event)
	}
}

func TestFromTurnLabelsAssistantEvidenceSyntheticAndBoundsFields(t *testing.T) {
	turn := capture("Please handle the deployment.", "I deployed it because the stale cache failed using "+strings.Repeat("verified ", 100)+"script.")
	event := frame.FromTurn(turn)
	if event.Why.Source != model.SourceSynthetic || event.How.Source != model.SourceSynthetic || event.Why.Confidence >= .9 {
		t.Fatalf("assistant provenance = why %+v, how %+v", event.Why, event.How)
	}
	for _, field := range []model.Field{event.Who, event.What, event.Where, event.When, event.Why, event.How} {
		if len(field.Value) > 320 || len(field.Evidence) > 160 {
			t.Fatalf("unbounded field = %+v", field)
		}
	}
}

func TestFromTurnStopsLocationBeforeTemporalPhrase(t *testing.T) {
	event := frame.FromTurn(capture("Update the cache in the production database next week using the migration tool.", "I will prepare the migration."))
	if event.Where.Value != "the production database" || event.When.Value != "next week" {
		t.Fatalf("where=%q when=%q", event.Where.Value, event.When.Value)
	}
}

func TestQueryTextUsesOnlyPartialEventFrame(t *testing.T) {
	raw := "Find the deployment in Toronto tomorrow because retrieval is stale using graph search. RAW_QUERY_SENTINEL"
	got := frame.QueryText(raw)
	for _, want := range []string{
		"representation: " + model.SemanticRepresentationVersion,
		"what: request: Find the deployment in Toronto tomorrow because retrieval is stale using graph search.",
		"where: Toronto", "when: tomorrow", "why: retrieval is stale", "how: using graph search",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("query frame %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "RAW_QUERY_SENTINEL") {
		t.Fatalf("query frame leaked raw tail: %q", got)
	}
}

func TestFromTextCreatesContentBearingEventFrame(t *testing.T) {
	now := time.Now().UTC()
	event := frame.FromText("Deploy EventFrame in Toronto tomorrow because retrieval is stale using graph search.", "user", "session-a", now, model.SourceObserved)
	if event.What.Value == "" || event.Where.Value != "Toronto" || event.When.Value != "tomorrow" || event.Why.Value != "retrieval is stale" || event.How.Value != "using graph search" {
		t.Fatalf("text EventFrame = %+v", event)
	}
}

func capture(user, assistant string) model.TurnCapture {
	now := time.Date(2026, 8, 30, 13, 0, 0, 0, time.UTC)
	return model.TurnCapture{
		ID: "turn-1", TenantID: "tenant-a", SessionID: "session-a", Sequence: uint64(now.UnixMilli()), AgentID: "main",
		UserText: user, AssistantText: assistant, OccurredAt: now, ObservedAt: now.Add(time.Second), AvailableAt: now.Add(time.Second),
	}
}
