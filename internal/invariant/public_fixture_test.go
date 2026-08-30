package invariant_test

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/invariant"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestPublicApolloFixtureComposesMissionWithoutErasingStageDifferences(t *testing.T) {
	file, err := os.Open("../../testdata/invariant-public-facts/apollo11.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	type row struct {
		Stage, What, Where, Who, Why, How, OccurredAt string
	}
	rows := make([]row, 0, 5)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var item struct {
			Stage      string `json:"stage"`
			What       string `json:"what"`
			Where      string `json:"where"`
			Who        string `json:"who"`
			Why        string `json:"why"`
			How        string `json:"how"`
			OccurredAt string `json:"occurred_at"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, row(item))
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	members := make([]model.Event, 0, len(rows))
	memberIDs := make([]string, 0, len(rows))
	for index, item := range rows {
		occurredAt, err := time.Parse(time.RFC3339, item.OccurredAt)
		if err != nil {
			t.Fatal(err)
		}
		field := func(value string) model.Field {
			return model.Field{Value: value, Source: model.SourceObserved, Confidence: 1}
		}
		memberIDs = append(memberIDs, item.Stage)
		members = append(members, model.Event{ID: item.Stage, TenantID: "public", SessionID: "apollo-11", Sequence: uint64(index + 1), Kind: "public-fact", Content: item.What, OccurredAt: occurredAt, ObservedAt: occurredAt, AvailableAt: occurredAt, Who: field(item.Who), What: field(item.What), Where: field(item.Where), When: field(item.OccurredAt), Why: field(item.Why), How: field(item.How), Priority: .8, Provenance: model.Provenance{Producer: "NASA public fixture"}})
	}
	request := model.ComposeInvariantRequest{ID: "apollo-11", TenantID: "public", SessionID: "apollo-11", MemberEventIDs: memberIDs, RepresentativeEventID: "launch", Label: "Apollo 11 crewed lunar landing and safe return mission", RuleID: "apollo-mission-stages-v1", Resolution: "mission", Confidence: .95, AntiPigeonCertificateID: "fixture-audit", PublishedAt: time.Date(1969, 7, 24, 17, 0, 0, 0, time.UTC)}
	composite := invariant.Compose(request, members)
	if composite.Who.Value != "Apollo 11 crew" || composite.Why.Value != "perform a crewed lunar landing and return safely to Earth" {
		t.Fatalf("shared mission invariants were not retained: who=%+v why=%+v", composite.Who, composite.Why)
	}
	if composite.Where.Value != "" || composite.How.Value != "" {
		t.Fatalf("stage-specific differences were falsely collapsed: where=%+v how=%+v", composite.Where, composite.How)
	}
	if len(composite.Composition.MemberEventIDs) != 5 || composite.Composition.RepresentativeEventID != "launch" || composite.OccurredAt != members[0].OccurredAt {
		t.Fatalf("composition envelope/provenance = %+v", composite)
	}
}
