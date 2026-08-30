package frame

import (
	"strings"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// FromText creates a single-message EventFrame for offline imports and other
// daemon-owned ingestion paths that do not have a paired turn envelope.
func FromText(text, participant, sessionID string, occurredAt time.Time, source model.FieldSource) model.Event {
	text = strings.TrimSpace(text)
	input := sourceText{name: "message", text: text, source: source}
	sources := []sourceText{input}
	statement := firstStatement(input)
	return model.Event{
		Content: text,
		Who:     field(participant, source, 1, "message role"),
		What:    field(statement.value, source, 1, statement.evidence),
		Where:   firstField(sources, wherePatterns, .86, validLocation, field("session:"+sessionID, model.SourceObserved, 1, "session metadata")),
		When:    firstField(sources, whenPatterns, .90, nil, field(occurredAt.Format(time.RFC3339Nano), model.SourceObserved, 1, "message timestamp")),
		Why:     firstField(sources, whyPatterns, .90, nil, model.Field{}),
		How:     firstField(sources, howPatterns, .86, nil, model.Field{}),
	}
}
