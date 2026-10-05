// Package researchpublicframe converts source descriptions after a document
// contract, without posing documents as conversations or inventing causal fields.
// This isolated adapter is not installed in the daemon or plugin.
package researchpublicframe

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const (
	Contract         = "public-source-assertion-span-v1"
	MaxInputBytes    = 16 << 20
	MaxDocuments     = 8192
	MaxDocumentBytes = 128 << 10
	MaxFrames        = 64
	MaxFieldBytes    = 2048
)

type Record struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

// UnmarshalJSON refuses two identities in one record and rejects the standard
// decoder's silent lone-surrogate replacement. RawMessage keeps that evidence
// available until validation; the standard parser still owns JSON syntax.
func (r *Record) UnmarshalJSON(input []byte) error {
	d := json.NewDecoder(bytes.NewReader(input))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return errors.New("public record must be an object")
	}
	seen := map[string]bool{}
	next := Record{}
	for d.More() {
		token, err = d.Token()
		if err != nil {
			return err
		}
		key, ok := token.(string)
		if !ok || seen[key] {
			return errors.New("duplicate or invalid public field")
		}
		seen[key] = true
		var raw json.RawMessage
		if err = d.Decode(&raw); err != nil {
			return err
		}
		value, err := scalarString(raw)
		if err != nil {
			return err
		}
		switch key {
		case "id":
			next.ID = value
		case "title":
			next.Title = value
		case "text":
			next.Text = value
		default:
			return errors.New("unknown public field")
		}
	}
	if token, err = d.Token(); err != nil || token != json.Delim('}') {
		return errors.New("invalid public object end")
	}
	var extra any
	if err = d.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("trailing public record")
	}
	*r = next
	return nil
}

func scalarString(raw []byte) (string, error) {
	var out string
	if len(raw) < 2 || raw[0] != '"' || raw[len(raw)-1] != '"' {
		return "", errors.New("public field must be a string")
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	for i := 1; i < len(raw)-1; i++ {
		if raw[i] != '\\' {
			continue
		}
		if raw[i+1] != 'u' {
			i++
			continue
		}
		v, err := strconv.ParseUint(string(raw[i+2:i+6]), 16, 16)
		if err != nil {
			return "", err
		}
		if v >= 0xdc00 && v <= 0xdfff {
			return "", errors.New("unpaired public low surrogate")
		}
		if v >= 0xd800 && v <= 0xdbff {
			if i+12 > len(raw)-1 || raw[i+6] != '\\' || raw[i+7] != 'u' {
				return "", errors.New("unpaired public high surrogate")
			}
			low, err := strconv.ParseUint(string(raw[i+8:i+12]), 16, 16)
			if err != nil || low < 0xdc00 || low > 0xdfff {
				return "", errors.New("invalid public surrogate pair")
			}
			i += 11
		} else {
			i += 5
		}
	}
	return out, nil
}

type Config struct {
	TenantID   string
	SessionID  string
	ImportedAt time.Time
}

type Span struct {
	Section string      `json:"section"`
	Start   int         `json:"start"`
	End     int         `json:"end"`
	Event   model.Event `json:"event"`
}

type Document struct {
	ID                string `json:"id"`
	SourceSHA256      string `json:"source_sha256"`
	CanonicalTitle    string `json:"canonical_title"`
	CanonicalAbstract string `json:"canonical_abstract"`
	Spans             []Span `json:"spans"`
}

func (c Config) validate() error {
	if strings.TrimSpace(c.TenantID) == "" || strings.TrimSpace(c.SessionID) == "" ||
		len(c.TenantID) > 128 || len(c.SessionID) > 128 || !utf8.ValidString(c.TenantID) ||
		!utf8.ValidString(c.SessionID) || c.ImportedAt.IsZero() {
		return errors.New("invalid public frame configuration")
	}
	return nil
}

// Decode refuses annotation-bearing records rather than discarding gold fields.
func Decode(input []byte) ([]Record, error) {
	if len(input) > MaxInputBytes || !utf8.Valid(input) {
		return nil, errors.New("public input bound or UTF-8 violation")
	}
	d := json.NewDecoder(bytes.NewReader(input))
	d.DisallowUnknownFields()
	var records []Record
	if err := d.Decode(&records); err != nil {
		return nil, err
	}
	var extra any
	if err := d.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errors.New("trailing public JSON")
	}
	if err := validateRecords(records); err != nil {
		return nil, err
	}
	return records, nil
}

func validateRecords(records []Record) error {
	if len(records) < 1 || len(records) > MaxDocuments {
		return errors.New("public document count violation")
	}
	seen := make(map[string]bool, len(records))
	total := 0
	for _, r := range records {
		total += len(r.ID) + len(r.Title) + len(r.Text)
		if total > MaxInputBytes {
			return errors.New("public aggregate source bound")
		}
		if len(r.ID) > 128 || strings.TrimSpace(r.ID) == "" || seen[r.ID] ||
			!utf8.ValidString(r.ID) || !utf8.ValidString(r.Title) || !utf8.ValidString(r.Text) ||
			strings.TrimSpace(r.Title) == "" || strings.TrimSpace(r.Text) == "" ||
			len(r.Title)+len(r.Text) > MaxDocumentBytes {
			return errors.New("invalid, duplicate or over-bound public source")
		}
		seen[r.ID] = true
	}
	return nil
}

func digest(value any) string {
	// Arrays have a deterministic cross-language encoding. Go's HTML escaping
	// is disabled so the separate auditor need not imitate an incidental encoder.
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if err := e.Encode(value); err != nil {
		panic(err)
	}
	h := sha256.Sum256(bytes.TrimSuffix(b.Bytes(), []byte{'\n'}))
	return hex.EncodeToString(h[:])
}

func canonical(s string) string { return strings.Join(strings.Fields(s), " ") }

// Convert validates the complete input before returning any envelopes. Source
// membership and the availability clock, not corpus order, determine identities.
func Convert(ctx context.Context, records []Record, c Config) ([]Document, error) {
	if ctx == nil {
		return nil, errors.New("nil public context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	if err := validateRecords(records); err != nil {
		return nil, err
	}
	out := make([]Document, 0, len(records))
	for _, r := range records {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		d := Document{ID: r.ID, SourceSHA256: digest([]string{r.ID, r.Title, r.Text}), CanonicalTitle: canonical(r.Title), CanonicalAbstract: canonical(r.Text)}
		content := r.Title + "\n\n" + r.Text
		for _, part := range []struct{ section, text string }{{"title", d.CanonicalTitle}, {"abstract", d.CanonicalAbstract}} {
			for start := 0; start < len(part.text); {
				end := min(len(part.text), start+MaxFieldBytes)
				for end < len(part.text) && !utf8.RuneStart(part.text[end]) {
					end--
				}
				if end < len(part.text) {
					if split := strings.LastIndexByte(part.text[start:end], ' '); split >= 0 {
						end = start + split + 1
					}
				}
				value := part.text[start:end]
				if strings.TrimSpace(value) == "" || len(d.Spans) >= MaxFrames {
					return nil, errors.New("public semantic span bound")
				}
				identity := digest([]any{Contract, c.TenantID, c.SessionID, c.ImportedAt.UTC().Format(time.RFC3339Nano), r.ID, d.SourceSHA256, part.section, start, end})
				clock := c.ImportedAt.UTC()
				e := model.Event{
					ID: "public-span-" + identity, TenantID: c.TenantID, SessionID: c.SessionID, Sequence: uint64(len(d.Spans) + 1),
					Kind: "public_source_assertion", Content: content, OccurredAt: clock, ObservedAt: clock, AvailableAt: clock,
					What:     model.Field{Value: value, Source: model.SourceObserved, Confidence: 1, Evidence: fmt.Sprintf("canonical-%s[bytes:%d:%d]", part.section, start, end)},
					Where:    model.Field{Value: "public scientific " + part.section, Source: model.SourceObserved, Confidence: 1, Evidence: "source section, not physical location"},
					Priority: .5, Provenance: model.Provenance{Producer: Contract},
					Attributes: map[string]string{"source_document_id": r.ID, "source_sha256": d.SourceSHA256, "source_section": part.section, "semantic_extractor": Contract, "clock_semantics": "import-not-publication", "confidence_semantics": "transcription-not-truth"},
				}
				if err := e.Validate(0); err != nil {
					return nil, err
				}
				d.Spans = append(d.Spans, Span{Section: part.section, Start: start, End: end, Event: e})
				start = end
			}
		}
		out = append(out, d)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Available owns mutable fields so a caller cannot mutate an imported snapshot.
// It is an import adapter helper, not a replacement for daemon as-of checks.
func Available(ctx context.Context, docs []Document, asOf time.Time) ([]model.Event, error) {
	if ctx == nil || asOf.IsZero() {
		return nil, errors.New("invalid public as-of context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := []model.Event{}
	for _, d := range docs {
		for _, span := range d.Spans {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			e := span.Event
			if e.AvailableAt.After(asOf) {
				continue
			}
			e.Attributes = make(map[string]string, len(span.Event.Attributes))
			for k, v := range span.Event.Attributes {
				e.Attributes[k] = v
			}
			e.Tags = append([]string(nil), e.Tags...)
			e.Embedding = append([]float32(nil), e.Embedding...)
			e.Provenance.SourceEventIDs = append([]string(nil), e.Provenance.SourceEventIDs...)
			e.Provenance.RetrievedIDs = append([]string(nil), e.Provenance.RetrievedIDs...)
			out = append(out, e)
		}
	}
	return out, nil
}
