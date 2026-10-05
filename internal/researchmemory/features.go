// Package researchmemory is an experimental real-field adapter, not serving.
package researchmemory

import (
	"errors"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const FeatureContract = "research-5w1h-overlap-v1"

var stops = map[string]bool{"a": true, "an": true, "the": true, "is": true, "was": true, "of": true, "to": true, "at": true, "in": true, "on": true, "and": true, "what": true, "which": true, "when": true, "according": true, "retained": true, "records": true, "date": true, "s": true}

func tokens(s string) map[string]bool {
	m := map[string]bool{}
	for _, v := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		if !stops[v] {
			m[v] = true
		}
	}
	return m
}
func numeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return s != ""
}

// Extract deliberately ignores Content, IDs, attributes, outcomes and source
// oracle metadata. All six compressed fields are observed; no independent-bit
// missing-field assumption is imported from the synthetic MMM experiment.
func Extract(query string, event model.Event, baseline float64) (uint16, error) {
	if len(query) > 1024 || !utf8.ValidString(query) || math.IsNaN(baseline) || math.IsInf(baseline, 0) || baseline < 0 || baseline > 1 {
		return 0, errors.New("invalid research feature input")
	}
	q := tokens(query)
	if len(q) == 0 {
		return 0, errors.New("no usable query terms")
	}
	fields := []model.Field{event.Who, event.What, event.Where, event.When, event.Why, event.How}
	union := map[string]bool{}
	var bits uint16
	for i, f := range fields {
		if len(f.Value) > 2048 || !utf8.ValidString(f.Value) {
			return 0, errors.New("invalid compressed field")
		}
		for v := range tokens(f.Value) {
			union[v] = true
			if q[v] {
				bits |= 1 << i
			}
		}
	}
	covered, numbers, numberMatches := 0, 0, 0
	for v := range q {
		if union[v] {
			covered++
		}
		if numeric(v) {
			numbers++
			if union[v] {
				numberMatches++
			}
		}
	}
	if 2*covered >= len(q) {
		bits |= 1 << 6
	}
	if numbers > 0 && numbers == numberMatches {
		bits |= 1 << 7
	}
	if baseline >= .5 {
		bits |= 1 << 8
	}
	return bits, nil
}
