package frame

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const (
	maxFieldBytes    = 320
	maxEvidenceBytes = 160
)

type sourceText struct {
	name   string
	text   string
	source model.FieldSource
}

type match struct {
	value      string
	evidence   string
	source     model.FieldSource
	confidence float64
}

type pattern struct {
	expression *regexp.Regexp
	group      int
}

var (
	whoPatterns = []pattern{
		{regexp.MustCompile(`\b(?i:I am|I'm|my name is)\s+(\p{Lu}[\p{L}'-]*(?:\s+\p{Lu}[\p{L}'-]*){0,3})\b`), 1},
		{regexp.MustCompile(`\b(\p{Lu}[\p{L}'-]*(?:\s+\p{Lu}[\p{L}'-]*){0,3})\s+(?:will|should|must|needs?|wants?|decided|reported|said|owns|created|implemented|deployed)\b`), 1},
	}
	wherePatterns = []pattern{
		{regexp.MustCompile(`(?i)\b(?:in|at|inside|within|under|from|on)\s+((?:https?://|[~/])[^\s,;!?]+)`), 1},
		{regexp.MustCompile(`(?i)\b(?:in|at|inside|within|under|from|on)\s+((?:the\s+)?[[:alpha:]\d_.:-]+(?:\s+[[:alpha:]\d_.:-]+){0,3}?)(?:\s+(?:today|tomorrow|yesterday|tonight|next|last|this|because|since|due|using|via|through|with|by|before|after|during|until)|[,;.!?]|$)`), 1},
	}
	whenPatterns = []pattern{
		{regexp.MustCompile(`(?i)\b(\d{4}-\d{2}-\d{2}(?:[T ]\d{2}:\d{2}(?::\d{2})?(?:Z|[+-]\d{2}:?\d{2})?)?)\b`), 1},
		{regexp.MustCompile(`(?i)\b(today|tomorrow|yesterday|tonight|now|immediately|later|eventually|currently|previously)\b`), 1},
		{regexp.MustCompile(`(?i)\b((?:next|last|this)\s+(?:minute|hour|day|week|month|quarter|year|morning|afternoon|evening))\b`), 1},
		{regexp.MustCompile(`(?i)\b((?:before|after|during|until|since)\s+[^,;.!?]{1,80})`), 1},
		{regexp.MustCompile(`(?i)\b(at\s+\d{1,2}(?::\d{2})?\s*(?:am|pm|UTC|GMT)?)\b`), 1},
	}
	whyPatterns = []pattern{
		{regexp.MustCompile(`(?i)\b(?:because|since|as a result of|due to)\s+([^,;.!?]{1,160}?)(?:\s+(?:using|via|through|with|by)|[,;.!?]|$)`), 1},
		{regexp.MustCompile(`(?i)\b((?:so that|in order to|for the purpose of)\s+[^,;.!?]{1,160}?)(?:\s+(?:using|via|through|with|by)|[,;.!?]|$)`), 1},
	}
	howPatterns = []pattern{
		{regexp.MustCompile(`(?i)\b((?:using|via|through|with)\s+[^,;.!?]{1,160})`), 1},
		{regexp.MustCompile(`(?i)\b(by\s+[[:alpha:]\d_-]+ing\b[^,;.!?]{0,140})`), 1},
	}
	collectivePattern        = regexp.MustCompile(`(?i)\b(?:we|us|our|ours)\b`)
	firstPersonPattern       = regexp.MustCompile(`(?i)\b(?:I|me|my|mine)\b`)
	secondPersonPattern      = regexp.MustCompile(`(?i)\b(?:you|your|yours)\b`)
	statementBoundaryPattern = regexp.MustCompile(`[.!?\n](?:\s|$)`)
)

// FromTurn performs deterministic post-contract enrichment inside eventframed.
// The 5W1H frame is an internal retrieval corpus; full turn text remains
// metadata. Pattern matches are bounded evidence-backed extractions, while
// fallbacks are explicitly marked inferred or synthetic rather than treated as
// ground truth.
func FromTurn(turn model.TurnCapture) model.Event {
	user := sourceText{name: "user", text: turn.UserText, source: model.SourceObserved}
	assistant := sourceText{name: "assistant", text: turn.AssistantText, source: model.SourceSynthetic}
	sources := []sourceText{user, assistant}
	request := firstStatement(user)
	outcome := firstStatement(assistant)

	return model.Event{
		ID: turn.ID, TenantID: turn.TenantID, SessionID: turn.SessionID, Sequence: turn.Sequence,
		Kind: "agent_turn", Content: "User: " + turn.UserText + "\n\nAssistant: " + turn.AssistantText,
		OccurredAt: turn.OccurredAt, ObservedAt: turn.ObservedAt, AvailableAt: turn.AvailableAt,
		Who:      firstField(sources, whoPatterns, .88, nil, participantFallback(turn, user)),
		What:     field("request: "+request.value+"; outcome: "+outcome.value, model.SourceInferred, .82, request.evidence+"; "+outcome.evidence),
		Where:    firstField(sources, wherePatterns, .86, validLocation, field("session:"+turn.SessionID, model.SourceObserved, 1, "session metadata")),
		When:     firstField(sources, whenPatterns, .90, nil, field(turn.OccurredAt.Format("2006-01-02T15:04:05.999999999Z07:00"), model.SourceObserved, 1, "turn timestamp")),
		Why:      firstField(sources, whyPatterns, .90, nil, field("to address the user request: "+request.value, model.SourceInferred, .62, request.evidence)),
		How:      firstField(sources, howPatterns, .86, nil, field("through the agent response: "+outcome.value, model.SourceInferred, .62, outcome.evidence)),
		Priority: .5, Tags: []string{"conversation", "agent-turn"},
		Provenance: model.Provenance{Producer: "openclaw-eventframe-memory", RetrievedIDs: append([]string(nil), turn.RetrievedIDs...), RunID: turn.RunID},
		Attributes: map[string]string{"user_source": "observed", "assistant_source": "synthetic", "semantic_extractor": "fivew1h-deterministic-v1"},
	}
}

func firstField(sources []sourceText, patterns []pattern, confidence float64, accept func(string) bool, fallback model.Field) model.Field {
	for _, source := range sources {
		for _, candidate := range patterns {
			indices := candidate.expression.FindStringSubmatchIndex(source.text)
			group := candidate.group * 2
			if len(indices) <= group+1 || indices[group] < 0 {
				continue
			}
			start, end := indices[group], indices[group+1]
			value := strings.TrimSpace(source.text[start:end])
			if value == "" || (accept != nil && !accept(value)) {
				continue
			}
			if source.source == model.SourceSynthetic {
				confidence = max(0, confidence-.12)
			}
			return field(value, source.source, confidence, span(source.name, start, end))
		}
	}
	return fallback
}

func participantFallback(turn model.TurnCapture, user sourceText) model.Field {
	if indices := collectivePattern.FindStringIndex(user.text); indices != nil {
		return field("user and agent", model.SourceInferred, .72, span(user.name, indices[0], indices[1]))
	}
	if indices := firstPersonPattern.FindStringIndex(user.text); indices != nil {
		return field("user", model.SourceInferred, .72, span(user.name, indices[0], indices[1]))
	}
	if turn.AgentID != "" {
		if indices := secondPersonPattern.FindStringIndex(user.text); indices != nil {
			return field("agent:"+turn.AgentID, model.SourceInferred, .70, span(user.name, indices[0], indices[1]))
		}
		return field("user and agent:"+turn.AgentID, model.SourceInferred, .55, "turn participants")
	}
	return field("user and agent", model.SourceInferred, .55, "turn participants")
}

func firstStatement(source sourceText) match {
	start := len(source.text) - len(strings.TrimLeft(source.text, " \t\r\n"))
	if start == len(source.text) {
		return match{value: "empty turn", evidence: "no textual evidence", source: model.SourceInferred}
	}
	rest := source.text[start:]
	end := len(rest)
	if boundary := statementBoundaryPattern.FindStringIndex(rest); boundary != nil {
		end = boundary[0] + 1
	}
	return match{value: bounded(strings.TrimSpace(rest[:end]), maxFieldBytes), evidence: span(source.name, start, start+end), source: source.source, confidence: 1}
}

func validLocation(value string) bool {
	normalized := strings.TrimPrefix(strings.ToLower(value), "the ")
	switch normalized {
	case "case", "event", "fact", "order", "response", "same way", "such a way", "time":
		return false
	default:
		return true
	}
}

func field(value string, source model.FieldSource, confidence float64, evidence string) model.Field {
	return model.Field{Value: bounded(value, maxFieldBytes), Source: source, Confidence: min(1, max(0, confidence)), Evidence: bounded(evidence, maxEvidenceBytes)}
}

func span(source string, start, end int) string {
	return fmt.Sprintf("%s[bytes:%d:%d]", source, start, end)
}

func bounded(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= limit {
		return value
	}
	end := limit - 3
	for end > 0 && !utf8.RuneStart(value[end]) {
		end--
	}
	return value[:end] + "..."
}
