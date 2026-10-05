// Package researchpublicpool binds public source spans to one retrieval record
// per document. It is an experiment boundary, not a production memory plugin.
package researchpublicpool

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchpublicframe"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

const Contract = "public-source-document-pool-v1"
const MaxTextBytes = 16384
const MaxFrontier = 200

type metadata struct {
	Collection    string   `json:"collection"`
	Timestamp     int64    `json:"ts"`
	Authored      bool     `json:"authored"`
	AccessCount   int      `json:"access_count"`
	Authority     float64  `json:"authority"`
	Salience      float64  `json:"salience"`
	SourceID      string   `json:"source_document_id"`
	SourceSHA256  string   `json:"source_sha256"`
	AvailableAt   string   `json:"available_at"`
	FrameContract string   `json:"frame_contract"`
	PoolContract  string   `json:"pool_contract"`
	SpanIDs       []string `json:"span_ids"`
}

type Entry struct {
	SourceID     string              `json:"source_id"`
	SourceSHA256 string              `json:"source_sha256"`
	AvailableAt  time.Time           `json:"available_at"`
	SpanIDs      []string            `json:"span_ids"`
	Candidate    retrieval.Candidate `json:"candidate"`
}

type Registry struct {
	collection string
	entries    []Entry
	byID       map[string]int
}

// New performs the entire source conversion before exposing the registry. No
// caller-provided EventFrame or annotation-bearing metadata is trusted here.
func New(ctx context.Context, records []researchpublicframe.Record, c researchpublicframe.Config, collection string) (*Registry, error) {
	if strings.TrimSpace(collection) == "" || len(collection) > 128 {
		return nil, errors.New("invalid public collection")
	}
	docs, err := researchpublicframe.Convert(ctx, records, c)
	if err != nil {
		return nil, err
	}
	r := &Registry{collection: collection, byID: make(map[string]int, len(docs))}
	for _, d := range docs {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		texts := make([]string, 0, len(d.Spans))
		ids := make([]string, 0, len(d.Spans))
		for _, s := range d.Spans {
			texts = append(texts, s.Event.FrameText())
			ids = append(ids, s.Event.ID)
		}
		text := strings.Join(texts, "\n\n")
		if len(text) > MaxTextBytes {
			return nil, errors.New("pooled public document exceeds embedding bound")
		}
		h := sha256.Sum256([]byte(Contract + "\x00" + ids[0] + "\x00" + d.SourceSHA256))
		id := "public-document-" + hex.EncodeToString(h[:])
		m := metadata{Collection: collection, Timestamp: c.ImportedAt.UnixMilli(), SourceID: d.ID, SourceSHA256: d.SourceSHA256,
			AvailableAt: c.ImportedAt.UTC().Format(time.RFC3339Nano), FrameContract: researchpublicframe.Contract, PoolContract: Contract, SpanIDs: ids}
		bytes, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		if _, ok := r.byID[id]; ok {
			return nil, errors.New("duplicate pooled public identity")
		}
		r.byID[id] = len(r.entries)
		r.entries = append(r.entries, Entry{SourceID: d.ID, SourceSHA256: d.SourceSHA256, AvailableAt: c.ImportedAt.UTC(), SpanIDs: ids,
			Candidate: retrieval.Candidate{ID: id, Text: text, Metadata: bytes}})
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return r, nil
}

func own(e Entry) Entry {
	e.SpanIDs = append([]string(nil), e.SpanIDs...)
	e.Candidate.Metadata = append([]byte(nil), e.Candidate.Metadata...)
	return e
}
func (r *Registry) Entries() []Entry {
	out := make([]Entry, len(r.entries))
	for i, e := range r.entries {
		out[i] = own(e)
	}
	return out
}
func (r *Registry) Collection() string { return r.collection }

func (r *Registry) Exclusions(ctx context.Context, asOf time.Time) (map[string][]string, error) {
	if ctx == nil || asOf.IsZero() {
		return nil, errors.New("invalid pool as-of context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ids := []string{}
	for _, e := range r.entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if e.AvailableAt.After(asOf) {
			ids = append(ids, e.Candidate.ID)
		}
	}
	return map[string][]string{r.collection: ids}, nil
}

type Bound struct {
	SourceID string
	Entry    Entry
	Score    float64
}

// Bind validates every row before returning anything. The registry supplies
// source identities; native metadata is not allowed to manufacture a new source.
func (r *Registry) Bind(ctx context.Context, rows []retrieval.Candidate, asOf time.Time, cap int) ([]Bound, error) {
	if ctx == nil || asOf.IsZero() || cap < 1 || cap > MaxFrontier || len(rows) > cap {
		return nil, errors.New("invalid public frontier")
	}
	seen := map[string]bool{}
	out := make([]Bound, 0, len(rows))
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		i, ok := r.byID[row.ID]
		if !ok || seen[row.ID] || math.IsNaN(row.Score) || math.IsInf(row.Score, 0) {
			return nil, errors.New("invalid, duplicate or nonfinite native row")
		}
		seen[row.ID] = true
		e := r.entries[i]
		if e.AvailableAt.After(asOf) || row.Text != e.Candidate.Text {
			return nil, errors.New("future or altered native row")
		}
		var got, want metadata
		if json.Unmarshal(row.Metadata, &got) != nil || json.Unmarshal(e.Candidate.Metadata, &want) != nil {
			return nil, errors.New("invalid native metadata")
		}
		// Ranking is allowed to decorate unrelated metadata; source/clock bindings
		// cannot change. Access statistics are ignored as evidence, not trusted.
		if got.SourceID != want.SourceID || got.SourceSHA256 != want.SourceSHA256 || got.Collection != want.Collection ||
			got.AvailableAt != want.AvailableAt || got.Timestamp != want.Timestamp || got.FrameContract != want.FrameContract || got.PoolContract != want.PoolContract ||
			len(got.SpanIDs) != len(want.SpanIDs) {
			return nil, errors.New("changed native source metadata")
		}
		for j, id := range want.SpanIDs {
			if got.SpanIDs[j] != id {
				return nil, errors.New("changed native source span")
			}
		}
		out = append(out, Bound{SourceID: e.SourceID, Entry: own(e), Score: row.Score})
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// BindRanked requires the whole nomination set before local correction/packing.
// No tail may disappear just because the RPC accepts a smaller K2.
func (r *Registry) BindRanked(ctx context.Context, nominated, ranked []retrieval.Candidate, asOf time.Time, cap int) ([]Bound, error) {
	if len(nominated) != len(ranked) {
		return nil, errors.New("native ranker shortened public frontier")
	}
	if _, err := r.Bind(ctx, nominated, asOf, cap); err != nil {
		return nil, err
	}
	out, err := r.Bind(ctx, ranked, asOf, cap)
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(nominated))
	for _, row := range nominated {
		set[row.ID] = true
	}
	for _, row := range ranked {
		if !set[row.ID] {
			return nil, errors.New("ranked public row outside nominated frontier")
		}
	}
	return out, nil
}
