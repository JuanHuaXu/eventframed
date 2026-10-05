package frame_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

func unresolvedReferences(t *testing.T, event model.Event) frame.UnresolvedReferences {
	t.Helper()
	var report frame.UnresolvedReferences
	if err := json.Unmarshal([]byte(event.Attributes["unresolved_references"]), &report); err != nil {
		t.Fatal(err)
	}
	return report
}

func TestIdentityUnresolvedReferencesAreExplicit(t *testing.T) {
	turn := capture("I will deploy because they need it using her console.", "I will check their status.")
	turn.UserID = "account:operator"
	event := frame.FromTurnWithIdentities(turn, nil, nil, true)
	report := unresolvedReferences(t, event)
	if report.Status != "needs-resolution" || report.Scope != "extracted-field-inputs" || report.Count != len(report.Entries) || report.Truncated {
		t.Fatalf("report = %+v", report)
	}
	wanted := map[string]string{
		"what.request:they":  "unsupported-group-or-object-reference",
		"why:it":             "unsupported-group-or-object-reference",
		"how:her":            "ambiguous-object-or-possessive-grammar",
		"what.outcome:their": "unsupported-group-or-object-reference",
	}
	for _, entry := range report.Entries {
		if entry.Mention == "I" {
			t.Fatal("resolved speaker was marked unresolved")
		}
		key := entry.Field + ":" + entry.Mention
		if reason, ok := wanted[key]; ok {
			if entry.Reason != reason || (entry.Field == "what.outcome") != (entry.Role == "assistant") {
				t.Fatalf("incorrect annotation: %+v", entry)
			}
			delete(wanted, key)
		}
	}
	if len(wanted) != 0 || !strings.Contains(event.What.Value, "they need it") || !strings.Contains(event.Content, turn.UserText) {
		t.Fatalf("missing annotations or altered evidence: %v / %+v", wanted, event)
	}
}

func TestIdentityUnresolvedReferenceReasonsAndNegativeControls(t *testing.T) {
	for _, test := range []struct{ text, mention, reason string }{
		{"I will deploy.", "I", "missing-speaker-identity"},
		{"You will deploy.", "You", "missing-addressee-identity"},
		{"He will deploy.", "He", "missing-or-ambiguous-antecedent"},
		{"We will deploy.", "We", "unsupported-group-or-object-reference"},
		{"Alex will deploy by himself.", "himself", "unsupported-reflexive-reference"},
	} {
		t.Run(test.mention, func(t *testing.T) {
			turn := capture(test.text, "Ready.")
			turn.UserID, turn.AgentID = "", ""
			report := unresolvedReferences(t, frame.FromTurnWithIdentities(turn, nil, nil, true))
			for _, entry := range report.Entries {
				if entry.Field == "what.request" && entry.Mention == test.mention && entry.Reason == test.reason {
					return
				}
			}
			t.Fatalf("unmarked pronoun: %+v", report)
		})
	}
	for _, text := range []string{
		"Alex will deploy.",
		`Alex will quote "they need her help".`,
		"Alex will inspect `it needs them`.",
		"> They will deploy.\nAlex will review.",
	} {
		report := unresolvedReferences(t, frame.FromTurnWithIdentities(capture(text, "Ready."), nil, nil, true))
		if report.Count != 0 || report.Status != "no-unresolved-detected" || len(report.Entries) != 0 {
			t.Fatalf("quoted/non-pronoun material flagged: %s / %+v", text, report)
		}
	}
	previous := frame.FromTurn(capture("Alex will deploy.", "Ready."))
	report := unresolvedReferences(t, frame.FromTurnWithIdentities(capture("He will use his console.", "Ready."), nil, &previous, true))
	if report.Count != 0 {
		t.Fatalf("resolved context marked unresolved: %+v", report)
	}
	previous = frame.FromTurn(capture("Alex will meet Blair.", "Ready."))
	report = unresolvedReferences(t, frame.FromTurnWithIdentities(capture("He will deploy.", "Ready."), nil, &previous, true))
	if report.Count == 0 || report.Entries[0].Field != "who" || report.Entries[0].Reason != "missing-or-ambiguous-antecedent" {
		t.Fatalf("ambiguous subject lacks a marker: %+v", report)
	}
	turn := capture("I will deploy.", "Ready.")
	turn.UserID = "account:a"
	report = unresolvedReferences(t, frame.FromTurnWithIdentities(turn, []frame.IdentityCard{{UserID: turn.UserID, Name: "They"}}, nil, true))
	if report.Count != 0 {
		t.Fatalf("resolved display name rescanned as a pronoun: %+v", report)
	}
}

func TestIdentityUnresolvedReferenceMetadataIsBounded(t *testing.T) {
	report := unresolvedReferences(t, frame.FromTurnWithIdentities(capture(strings.Repeat("it ", 200), "Ready."), nil, nil, true))
	if !report.Truncated || len(report.Entries) != 64 || report.Count <= 64 || report.Status != "needs-resolution" {
		t.Fatalf("unbounded or silently truncated report: %+v", report)
	}
}

func TestIdentityAuditSubjectAndTokenBoundaries(t *testing.T) {
	t.Run("contracted addressee", func(t *testing.T) {
		turn := capture("You'd deploy after approval.", "Ready.")
		event := frame.FromTurnWithIdentities(turn, nil, nil, true)
		if event.Who.Value != "agent:main" || unresolvedReferences(t, event).Count != 0 {
			t.Fatalf("subject and reference resolution disagree: %+v", event)
		}
	})
	t.Run("noninitial pronoun", func(t *testing.T) {
		event := frame.FromTurnWithIdentities(capture("Please proceed; He will deploy.", "Ready."), nil, nil, true)
		if event.Who.Value == "He" || event.Who.Confidence != 0 {
			t.Fatalf("pronoun treated as an explicit named subject: %+v", event.Who)
		}
	})
	t.Run("structured identifiers", func(t *testing.T) {
		turn := capture("I will deploy in us-east-1 using https://you.example/my/tool?owner=you#it.", "Ready.")
		turn.UserID = "account:a"
		event := frame.FromTurnWithIdentities(turn, nil, nil, true)
		if !strings.Contains(event.What.Value, "https://you.example/my/tool?owner=you#it") || !strings.Contains(event.Where.Value, "us-east-1") || unresolvedReferences(t, event).Count != 0 {
			t.Fatalf("identifier parts treated as pronouns: %+v", event)
		}
	})
	t.Run("unicode word boundary", func(t *testing.T) {
		event := frame.FromTurnWithIdentities(capture("Alex will inspect éit and 汉they.", "Ready."), nil, nil, true)
		if unresolvedReferences(t, event).Count != 0 {
			t.Fatalf("partial Unicode words flagged as pronouns: %+v", event)
		}
	})
}

func TestIdentityAuditQuotedMultilineField(t *testing.T) {
	turn := capture("Alex will deploy using a console\n> they need me\nand a script.", "Ready.")
	turn.UserID = "account:a"
	event := frame.FromTurnWithIdentities(turn, nil, nil, true)
	if !strings.Contains(event.How.Value, "> they need me") || unresolvedReferences(t, event).Count != 0 {
		t.Fatalf("normalized quote became active reference evidence: %+v", event)
	}
}

func TestIdentityAuditPriorFocusDoesNotBorrowFutureNames(t *testing.T) {
	previous := frame.FromTurn(capture("Alex will deploy.", "Ready."))
	turn := capture("He will deploy. Blair will inspect Casey.", "Ready.")
	event := frame.FromTurnWithIdentities(turn, nil, &previous, true)
	if event.Who.Value != "Alex" || !strings.Contains(event.What.Value, "Alex will deploy") || unresolvedReferences(t, event).Count != 0 {
		t.Fatalf("future names changed an earlier reference decision: %+v", event)
	}
}

func TestIdentityAuditUnicodeSubjectsAndPunctuation(t *testing.T) {
	for _, text := range []string{"Youé will deploy.", "Ié will deploy.", "you:123 will deploy."} {
		turn := capture(text, "Ready.")
		turn.UserID = "account:a"
		event := frame.FromTurnWithIdentities(turn, nil, nil, true)
		if event.Who.Value == "agent:main" || event.Who.Value == "user:account:a" || unresolvedReferences(t, event).Count != 0 {
			t.Fatalf("identifier/Unicode prefix became an actor: %s / %+v", text, event)
		}
	}
	turn := capture("Alex will inspect it: check the logs.", "Ready.")
	report := unresolvedReferences(t, frame.FromTurnWithIdentities(turn, nil, nil, true))
	if report.Count == 0 {
		t.Fatal("colon punctuation hid a genuine unresolved pronoun")
	}
	turn = capture("I will inspect my.example and me@example.org.", "Ready.")
	turn.UserID = "account:a"
	event := frame.FromTurnWithIdentities(turn, nil, nil, true)
	if !strings.Contains(event.What.Value, "my.example") || !strings.Contains(event.What.Value, "me@example.org") {
		t.Fatalf("domain/email rewritten: %+v", event)
	}
}

func TestIdentityAuditTruncationDoesNotInventPronouns(t *testing.T) {
	// The 320-byte excerpt cuts the word after "it" and appends "...".
	turn := capture(strings.Repeat("x ", 157)+" itsomething", "Ready.")
	report := unresolvedReferences(t, frame.FromTurnWithIdentities(turn, nil, nil, true))
	if report.Count != 0 {
		t.Fatalf("truncation created a pronoun from part of a word: %+v", report)
	}
}

func FuzzIdentityReferenceMetadata(f *testing.F) {
	for _, pair := range [][2]string{
		{"I will deploy because they need it.", "I'll inspect your logs."},
		{"Alex will deploy using a console\n> they need me\nand a script.", "Ready."},
		{"You'd inspect https://my.example/my/tool.", "We will wait."},
		{"He will inspect éit and 汉they.", "`they need it`"},
		{strings.Repeat("it ", 200), "their tool"},
	} {
		f.Add(pair[0], pair[1])
	}
	f.Fuzz(func(t *testing.T, user, assistant string) {
		if len(user) > 4096 || len(assistant) > 4096 || !utf8.ValidString(user) || !utf8.ValidString(assistant) {
			t.Skip()
		}
		turn := capture(user, assistant)
		turn.UserID = "account:a"
		event := frame.FromTurnWithIdentities(turn, nil, nil, true)
		report := unresolvedReferences(t, event)
		if event.Content != "User: "+user+"\n\nAssistant: "+assistant || report.Count < len(report.Entries) || len(report.Entries) > 64 || report.Truncated != (report.Count > 64) || (report.Count > 0) != (report.Status == "needs-resolution") {
			t.Fatalf("metadata/raw-text invariant violated: %+v", report)
		}
		for _, entry := range report.Entries {
			if entry.Mention == "" || entry.Reason == "" || (entry.Role != "user" && entry.Role != "assistant") {
				t.Fatalf("invalid annotation: %+v", entry)
			}
		}
		for _, field := range []model.Field{event.Who, event.What, event.Where, event.When, event.Why, event.How} {
			if !utf8.ValidString(field.Value) || len(field.Value) > 320 {
				t.Fatalf("invalid bounded field: %+v", field)
			}
		}
		replay := frame.FromTurnWithIdentities(turn, nil, nil, true)
		if replay.Attributes["unresolved_references"] != event.Attributes["unresolved_references"] || replay.FrameText() != event.FrameText() {
			t.Fatal("same evidence produced nondeterministic extraction")
		}
	})
}

func TestIdentitySpeakerIsNotSelfClaimedName(t *testing.T) {
	turn := capture("I am Another Person because my account needs access.", "I will check your account.")
	turn.UserID = "account:operator"
	card, err := frame.ParseIdentityCard(turn.UserID, `{"name":"Example Operator","aliases":["Operator"]}`, 3, turn.AvailableAt.Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	event := frame.FromTurnWithIdentities(turn, []frame.IdentityCard{card}, nil, true)
	var resolution frame.WhoResolution
	if err := json.Unmarshal([]byte(event.Attributes["who_resolution"]), &resolution); err != nil {
		t.Fatal(err)
	}
	if event.Who.Value != "Example Operator [user:account:operator]" || resolution.UserID != turn.UserID || resolution.CardVersion != 3 || len(resolution.CardHash) != 64 {
		t.Fatalf("identity = %+v / %+v", event.Who, resolution)
	}
	if !strings.Contains(event.What.Value, "Example Operator is Another Person") || !strings.Contains(event.What.Value, "agent:main will check Example Operator's account") {
		t.Fatalf("references = %s", event.What.Value)
	}
	if !strings.Contains(event.Why.Value, "Example Operator's account") || !strings.Contains(event.Content, "my account") {
		t.Fatalf("raw evidence overwritten: %+v", event)
	}
}

func TestIdentityContextResolvesButAbstainsOnCompetingSubjects(t *testing.T) {
	previous := frame.FromTurn(capture("Alex will deploy tomorrow.", "Ready."))
	turn := capture("He will use his tool because he needs it.", "Acknowledged.")
	event := frame.FromTurnWithIdentities(turn, nil, &previous, true)
	if event.Who.Value != "Alex" || !strings.Contains(event.What.Value, "Alex will use Alex's tool") || event.Who.Confidence >= 1 {
		t.Fatalf("context resolution = %+v", event)
	}
	previous = frame.FromTurn(capture("Alex said Blair will deploy tomorrow.", "Ready."))
	event = frame.FromTurnWithIdentities(turn, nil, &previous, true)
	if event.Who.Confidence != 0 || !strings.Contains(event.What.Value, "He will") {
		t.Fatalf("ambiguous reference guessed: %+v", event)
	}
	for _, text := range []string{"Alex and Blair will deploy.", "Alex will meet Blair."} {
		previous = frame.FromTurn(capture(text, "Ready."))
		event = frame.FromTurnWithIdentities(turn, nil, &previous, true)
		if event.Who.Confidence != 0 {
			t.Fatalf("competing object/coordinated antecedent guessed: %s / %+v", text, event.Who)
		}
	}
}

func TestIdentityDoesNotBorrowLaterSubjectOrQuotedNames(t *testing.T) {
	for _, text := range []string{
		`He will deploy. Alex will review it.`,
		`"Alex will deploy." He will review it.`,
		"`Alex will deploy.` He will review it.",
		"> Alex will deploy.\nHe will review it.",
		"```text\nAlex will deploy using `code`.\n```\nHe will review it.",
		`"quoted \" text Alex will deploy." He will review it.`,
		"We will deploy together.",
	} {
		event := frame.FromTurnWithIdentities(capture(text, "Alex will help."), nil, nil, true)
		if strings.Contains(event.What.Value, "Alex will deploy") && !strings.Contains(text, "Alex will deploy") {
			t.Fatalf("later reference = %s", event.What.Value)
		}
		if event.Who.Value == "Alex" && !strings.HasPrefix(text, "He will deploy. Alex") {
			t.Fatalf("quoted/assistant subject leaked: %+v", event.Who)
		}
		if strings.HasPrefix(text, "He will") && !strings.Contains(event.What.Value, "He will deploy") {
			t.Fatalf("future subject used: %s", event.What.Value)
		}
	}
}

func TestIdentityIncompleteRosterDoesNotCreateUniqueAlias(t *testing.T) {
	turn := capture("Alex will deploy.", "Ready.")
	card := frame.IdentityCard{UserID: "account:a", Name: "Alex", Aliases: []string{"Alex"}, Version: 1}
	event := frame.FromTurnWithIdentities(turn, []frame.IdentityCard{card}, nil, false)
	var resolution frame.WhoResolution
	_ = json.Unmarshal([]byte(event.Attributes["who_resolution"]), &resolution)
	if resolution.UserID != "" || resolution.Status != "alias-roster-incomplete" {
		t.Fatalf("partial roster claimed uniqueness: %+v", resolution)
	}
	turn.UserText, turn.UserID = "I will deploy.", card.UserID
	event = frame.FromTurnWithIdentities(turn, []frame.IdentityCard{card}, nil, false)
	_ = json.Unmarshal([]byte(event.Attributes["who_resolution"]), &resolution)
	if resolution.UserID != turn.UserID || resolution.Method != "adapter-speaker" {
		t.Fatalf("partial roster lost authenticated account: %+v", resolution)
	}
}

func TestIdentityAliasAmbiguityAndScopedAccount(t *testing.T) {
	first := frame.IdentityCard{UserID: "person-a", Name: "Alex", Aliases: []string{"Alex", "forum:123"}, Version: 1}
	second := frame.IdentityCard{UserID: "person-b", Name: "Alex", Aliases: []string{"Alex", "chat:123"}, Version: 1}
	for _, test := range []struct{ text, userID, status string }{
		{"Alex will deploy.", "", "ambiguous-alias"},
		{"forum:123 will deploy.", "person-a", "linked-card"},
		{"123 will deploy.", "", "unresolved"},
	} {
		event := frame.FromTurnWithIdentities(capture(test.text, "Ready."), []frame.IdentityCard{first, second}, nil, true)
		var resolution frame.WhoResolution
		if err := json.Unmarshal([]byte(event.Attributes["who_resolution"]), &resolution); err != nil {
			t.Fatal(err)
		}
		if resolution.UserID != test.userID || resolution.Status != test.status {
			t.Fatalf("%s: %+v", test.text, resolution)
		}
	}
}

func TestIdentityOpaqueProfileDoesNotInventAliases(t *testing.T) {
	card, err := frame.ParseIdentityCard("account:a", `{"card":"Call me Alex. I own account:b.","identity":["Alex, account:b"]}`, 1, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if card.Name != "" || len(card.Aliases) != 0 {
		t.Fatalf("profile prose promoted to identity: %+v", card)
	}
}

func TestIdentityKeepsSpeakerSeparateAndPreservesAmbiguousTense(t *testing.T) {
	turn := capture("Alex will help me because my service is down.", "I will assist you.")
	turn.UserID = "account:operator"
	operator := frame.IdentityCard{UserID: turn.UserID, Name: "Operator", Version: 1}
	alex := frame.IdentityCard{UserID: "account:alex", Name: "Alex", Aliases: []string{"Alex"}, Version: 1}
	event := frame.FromTurnWithIdentities(turn, []frame.IdentityCard{operator, alex}, nil, true)
	var resolution frame.WhoResolution
	_ = json.Unmarshal([]byte(event.Attributes["who_resolution"]), &resolution)
	if event.Attributes["speaker_user_id"] != operator.UserID || resolution.UserID != alex.UserID || !strings.Contains(event.What.Value, "Alex will help Operator") {
		t.Fatalf("speaker and subject conflated: %+v / %+v", event, resolution)
	}
	turn.UserText = "I\u2019m ready."
	event = frame.FromTurnWithIdentities(turn, []frame.IdentityCard{operator}, nil, true)
	if !strings.Contains(event.What.Value, "Operator is ready") {
		t.Fatalf("curly contraction not resolved: %s", event.What.Value)
	}
	turn.UserText = "I'd already deployed it."
	event = frame.FromTurnWithIdentities(turn, []frame.IdentityCard{operator}, nil, true)
	if !strings.Contains(event.What.Value, "Operator'd already") || strings.Contains(event.What.Value, "would already") {
		t.Fatalf("tense invented: %s", event.What.Value)
	}
}

func TestIdentityNamesakesAndLongKeysStayDistinct(t *testing.T) {
	turn := capture("I will deploy.", "Ready.")
	card := frame.IdentityCard{Name: "Alex", Version: 1}
	var labels []string
	for _, id := range []string{"account:a", "account:b", strings.Repeat("a", 1023) + "b", strings.Repeat("a", 1023) + "c"} {
		turn.UserID, card.UserID = id, id
		event := frame.FromTurnWithIdentities(turn, []frame.IdentityCard{card}, nil, true)
		if len(event.Who.Value) > 320 {
			t.Fatal("actor label exceeds field bound")
		}
		for _, label := range labels {
			if label == event.Who.Value {
				t.Fatal("namesakes or truncated keys collapsed")
			}
		}
		labels = append(labels, event.Who.Value)
	}
}

func BenchmarkIdentityEnrichment(b *testing.B) {
	turn := capture("Alex will deploy because his service needs a restart using the console.", "I will assist you.")
	turn.UserID = "account:a"
	card := frame.IdentityCard{UserID: turn.UserID, Name: "Operator", Aliases: []string{"Alex"}, Version: 1}
	for b.Loop() {
		frame.FromTurnWithIdentities(turn, []frame.IdentityCard{card}, nil, true)
	}
}
