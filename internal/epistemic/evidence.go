// Package epistemic derives bounded, content-free identifiers used to keep
// repeated records from masquerading as independent evidence.
package epistemic

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"unicode"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const DistinctOccurrenceAttribute = "epistemic_distinct_occurrence"

// Descriptor keeps predictive compatibility separate from evidence
// independence. Anti-Pigeon may still split two records with the same
// EvidenceGroupKey into different predictive buckets.
type Descriptor struct {
	ClaimKey           string
	LineageKey         string
	EvidenceGroupKey   string
	DistinctOccurrence bool
	claimTerms         map[string]struct{}
}

// Describe derives opaque keys from the canonical 5W1H representation and
// recorded provenance. Run and session IDs are deliberately excluded: a new
// conversation must not turn a repeated assertion into fresh corroboration.
func Describe(event model.Event) Descriptor {
	claim := strings.Join([]string{
		normalize(event.Who.Value),
		normalize(event.What.Value),
		normalizeSemanticWhere(event.Where.Value),
		normalize(event.Why.Value),
		normalize(event.How.Value),
	}, "\x00")
	claimKey := digest("claim\x00" + claim)
	lineageKey := lineage(event)
	distinct := !isConversation(event) && strings.EqualFold(strings.TrimSpace(event.Attributes[DistinctOccurrenceAttribute]), "true") &&
		event.What.Source == model.SourceObserved && event.When.Source == model.SourceObserved
	groupMaterial := claimKey + "\x00" + lineageKey
	if distinct {
		// A declared occurrence remains independent only when its semantic time is
		// observed. This preserves repeated real events without trusting ordinary
		// chat timestamps as evidence of independence.
		groupMaterial += "\x00occurrence\x00" + normalize(event.When.Value)
	}
	return Descriptor{ClaimKey: claimKey, LineageKey: lineageKey, EvidenceGroupKey: digest("group\x00" + groupMaterial), DistinctOccurrence: distinct, claimTerms: terms(claim)}
}

// Correlated recognizes exact canonical repetitions and conservative
// near-duplicates from the same lineage. It never treats explicitly distinct
// occurrences or different provenance lineages as the same evidence.
func Correlated(left, right Descriptor, minimumSimilarity float64) bool {
	if left.EvidenceGroupKey == right.EvidenceGroupKey {
		return true
	}
	if left.DistinctOccurrence || right.DistinctOccurrence || left.LineageKey != right.LineageKey {
		return false
	}
	return jaccard(left.claimTerms, right.claimTerms) >= minimumSimilarity
}

func lineage(event model.Event) string {
	parts := []string{normalize(event.Provenance.Producer)}
	if isConversation(event) {
		// Generated conversation is never a fresh authority merely because it
		// cites a different recalled record. Its claim remains correlated with
		// other claims emitted by the same producer.
		return digest("lineage\x00" + strings.Join(parts, "\x00"))
	}
	if len(event.Provenance.SourceEventIDs) > 0 {
		ids := append([]string(nil), event.Provenance.SourceEventIDs...)
		sort.Strings(ids)
		parts = append(parts, "sources", strings.Join(ids, "\x00"))
	} else if strings.TrimSpace(event.Provenance.ToolCallID) != "" {
		parts = append(parts, "tool", strings.TrimSpace(event.Provenance.ToolCallID))
	}
	return digest("lineage\x00" + strings.Join(parts, "\x00"))
}

func isConversation(event model.Event) bool {
	return event.Kind == "agent_turn" || event.Kind == "conversation_message" || hasTag(event.Tags, "conversation") || hasTag(event.Tags, "agent-turn")
}

func hasTag(tags []string, target string) bool {
	for _, tag := range tags {
		if strings.EqualFold(strings.TrimSpace(tag), target) {
			return true
		}
	}
	return false
}

func normalizeSemanticWhere(value string) string {
	normalized := normalize(value)
	if strings.HasPrefix(normalized, "session ") {
		return ""
	}
	return normalized
}

func normalize(value string) string {
	return strings.Join(strings.FieldsFunc(strings.ToLower(value), func(current rune) bool {
		return !unicode.IsLetter(current) && !unicode.IsDigit(current)
	}), " ")
}

func terms(value string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, term := range strings.Fields(normalize(value)) {
		if len(term) >= 2 || allDigits(term) {
			result[term] = struct{}{}
		}
	}
	return result
}

func allDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, current := range value {
		if !unicode.IsDigit(current) {
			return false
		}
	}
	return true
}

func jaccard(left, right map[string]struct{}) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	intersection := 0
	for term := range left {
		if _, ok := right[term]; ok {
			intersection++
		}
	}
	return float64(intersection) / float64(len(left)+len(right)-intersection)
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
