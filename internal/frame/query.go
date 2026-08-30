package frame

import (
	"strings"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// QueryText converts a raw recall request into a partial EventFrame. It omits
// inferred defaults that would add agent/session noise to semantic matching.
func QueryText(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return "representation: " + model.SemanticRepresentationVersion
	}
	source := sourceText{name: "query", text: query, source: model.SourceObserved}
	sources := []sourceText{source}
	request := firstStatement(source)
	event := model.Event{
		Who:  firstField(sources, whoPatterns, .88, nil, model.Field{}),
		What: field("request: "+request.value, model.SourceObserved, 1, request.evidence),
		Where: firstField(sources, wherePatterns, .86, validLocation,
			model.Field{}),
		When: firstField(sources, whenPatterns, .90, nil, model.Field{}),
		Why:  firstField(sources, whyPatterns, .90, nil, model.Field{}),
		How:  firstField(sources, howPatterns, .86, nil, model.Field{}),
	}
	return event.FrameText()
}
