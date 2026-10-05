package researchpublicpool_test

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchpublicframe"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicpool"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

var now = time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)

func cfg() researchpublicframe.Config {
	return researchpublicframe.Config{TenantID: "research-scifact", SessionID: "public-import", ImportedAt: now}
}
func build(t testing.TB, records []researchpublicframe.Record) *researchpublicpool.Registry {
	t.Helper()
	r, err := researchpublicpool.New(context.Background(), records, cfg(), "public-scientific")
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func records() []researchpublicframe.Record {
	return []researchpublicframe.Record{{ID: "one", Title: "A source title", Text: strings.Repeat("A retained observation. ", 220) + "Final source tail."}, {ID: "two", Title: "Second source", Text: "A different source observation."}, {ID: "three", Title: "Third source", Text: "Another source observation."}}
}

func TestPublicPoolV1CoverageOwnershipAndIdentity(t *testing.T) {
	r := build(t, records())
	e := r.Entries()
	if len(e) != 3 {
		t.Fatal("source duplication")
	}
	if len(e[0].SpanIDs) < 3 || !strings.Contains(e[0].Candidate.Text, "Final source tail.") {
		t.Fatal("lost source spans")
	}
	if strings.Contains(e[0].Candidate.Text, e[0].Candidate.ID) || strings.Contains(e[0].Candidate.Text, e[0].SourceSHA256) {
		t.Fatal("identity leaked into scored corpus")
	}
	e[0].Candidate.Metadata[0] = 'X'
	e[0].SpanIDs[0] = "altered"
	if reflect.DeepEqual(e, r.Entries()) {
		t.Fatal("mutable import ownership")
	}
	a := records()
	a[0], a[2] = a[2], a[0]
	other := build(t, a).Entries()
	own := r.Entries()
	if !reflect.DeepEqual(own[0], other[2]) || !reflect.DeepEqual(own[2], other[0]) {
		t.Fatal("source order affected pooled identity")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if out, err := researchpublicpool.New(ctx, records(), cfg(), "public-scientific"); err == nil || out != nil {
		t.Fatal("canceled registry")
	}
	bad := []researchpublicframe.Record{{ID: "large", Title: "Title", Text: strings.Repeat("x", 16300)}}
	if out, err := researchpublicpool.New(context.Background(), bad, cfg(), "public-scientific"); err == nil || out != nil {
		t.Fatal("silent pooled input truncation")
	}
}

func TestPublicPoolV1ResponsesAndWholeFrontier(t *testing.T) {
	r := build(t, records())
	e := r.Entries()
	rows := []retrieval.Candidate{e[0].Candidate, e[1].Candidate}
	rows[0].Score = .2
	rows[1].Score = .9
	bound, err := r.BindRanked(context.Background(), rows, []retrieval.Candidate{rows[1], rows[0]}, now, 50)
	if err != nil || len(bound) != 2 || bound[0].SourceID != "two" {
		t.Fatal(bound, err)
	}
	// Equal-score order is owned by the native ranker; the binder must not sort.
	rows[0].Score = .9
	out, err := r.Bind(context.Background(), rows, now, 50)
	if err != nil || out[0].SourceID != "one" {
		t.Fatal("tie reordered", err)
	}
	mutations := []func([]retrieval.Candidate){
		func(x []retrieval.Candidate) { x[0].ID = "unknown" }, func(x []retrieval.Candidate) { x[0].Text += " hidden" },
		func(x []retrieval.Candidate) { x[0].Score = math.NaN() }, func(x []retrieval.Candidate) { x[0].Score = math.Inf(1) },
		func(x []retrieval.Candidate) { x[1] = x[0] }, func(x []retrieval.Candidate) { x[0].Metadata = []byte(`{}`) },
	}
	for _, mutate := range mutations {
		copyRows := append([]retrieval.Candidate(nil), rows...)
		mutate(copyRows)
		if out, err := r.Bind(context.Background(), copyRows, now, 50); err == nil || out != nil {
			t.Fatal("invalid row accepted")
		}
	}
	for _, key := range []string{"source_document_id", "source_sha256", "collection", "available_at", "frame_contract", "pool_contract", "ts", "span_ids"} {
		copyRows := append([]retrieval.Candidate(nil), rows...)
		var meta map[string]any
		json.Unmarshal(copyRows[0].Metadata, &meta)
		delete(meta, key)
		copyRows[0].Metadata, _ = json.Marshal(meta)
		if out, err := r.Bind(context.Background(), copyRows, now, 50); err == nil || out != nil {
			t.Fatal("changed metadata accepted", key)
		}
	}
	if out, err := r.BindRanked(context.Background(), rows, rows[:1], now, 50); err == nil || out != nil {
		t.Fatal("ranker truncated before correction")
	}
	if out, err := r.BindRanked(context.Background(), rows, []retrieval.Candidate{rows[0], e[2].Candidate}, now, 50); err == nil || out != nil {
		t.Fatal("outside-frontier ranking")
	}
	if out, err := r.Bind(context.Background(), rows, now, 1); err == nil || out != nil {
		t.Fatal("oversized frontier")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if out, err := r.Bind(ctx, nil, now, 50); err == nil || out != nil {
		t.Fatal("empty canceled binding")
	}
	t.Log("6 row corruptions, 8 metadata corruptions, shortened/outside/cap/canceled boundaries executed")
}

func TestPublicPoolV1AvailabilityAndMetadataOnlyPoisoning(t *testing.T) {
	r := build(t, records())
	before, err := r.Exclusions(context.Background(), now.Add(-time.Nanosecond))
	if err != nil || len(before[r.Collection()]) != 3 {
		t.Fatal(before, err)
	}
	after, err := r.Exclusions(context.Background(), now)
	if err != nil || len(after[r.Collection()]) != 0 {
		t.Fatal(after, err)
	}
	e := r.Entries()[0]
	if out, err := r.Bind(context.Background(), []retrieval.Candidate{e.Candidate}, now.Add(-time.Nanosecond), 50); err == nil || out != nil {
		t.Fatal("future row accepted")
	}
	var m map[string]any
	json.Unmarshal(e.Candidate.Metadata, &m)
	m["untrusted_annotation"] = "SUPPORT"
	m["access_count"] = 500
	e.Candidate.Metadata, _ = json.Marshal(m)
	b, err := r.Bind(context.Background(), []retrieval.Candidate{e.Candidate}, now, 50)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b[0].Entry.Candidate.Metadata), "SUPPORT") || strings.Contains(b[0].Entry.Candidate.Text, "SUPPORT") {
		t.Fatal("metadata decoration reached scorer input")
	}
}

func corpus(t testing.TB) []researchpublicframe.Record {
	t.Helper()
	p := os.Getenv("EVENTFRAME_PUBLIC_CORPUS")
	if p == "" {
		t.Fatal("actual corpus path required")
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	r, err := researchpublicframe.Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func TestPublicPoolV1FullCorpus(t *testing.T) {
	r := build(t, corpus(t))
	e := r.Entries()
	if len(e) != 5183 {
		t.Fatal("document count")
	}
	n := 0
	max := 0
	for _, v := range e {
		n += len(v.SpanIDs)
		if len(v.Candidate.Text) > max {
			max = len(v.Candidate.Text)
		}
	}
	if n != 10869 {
		t.Fatal("lost source frames")
	}
	t.Logf("documents=%d frames=%d max pooled bytes=%d", len(e), n, max)
}

var saved *researchpublicpool.Registry
var bound []researchpublicpool.Bound

func BenchmarkPublicPoolV1Build(b *testing.B) {
	r := corpus(b)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		saved = build(b, r)
	}
}
func BenchmarkPublicPoolV1Bind50(b *testing.B)  { benchBind(b, 50) }
func BenchmarkPublicPoolV1Bind200(b *testing.B) { benchBind(b, 200) }
func benchBind(b *testing.B, n int) {
	r := build(b, corpus(b))
	e := r.Entries()
	rows := make([]retrieval.Candidate, n)
	for i := range rows {
		rows[i] = e[i].Candidate
		rows[i].Score = float64(n-i) / float64(n)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var err error
		bound, err = r.BindRanked(context.Background(), rows, rows, now, n)
		if err != nil {
			b.Fatal(err)
		}
	}
}
