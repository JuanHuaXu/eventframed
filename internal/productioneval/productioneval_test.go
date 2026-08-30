package productioneval

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

type countingRetriever struct{ calls int }
type countingRanker struct{ calls int }

func (r *countingRetriever) RetrievalContractName() string { return "test/SearchTextCollections" }
func (r *countingRetriever) SearchTextCollections(_ context.Context, _ retrieval.SearchRequest) ([]retrieval.Candidate, error) {
	r.calls++
	return []retrieval.Candidate{{ID: "candidate"}}, nil
}

func (r *countingRanker) ContractName() string { return "test/RankCandidates" }
func (r *countingRanker) RankCandidates(_ context.Context, request retrieval.RankRequest) ([]retrieval.Candidate, error) {
	r.calls++
	return append([]retrieval.Candidate(nil), request.Candidates...), nil
}

func TestCachedCandidateRetrieverCallsContractOncePerRequest(t *testing.T) {
	upstream := &countingRetriever{}
	cached := newCachedCandidateRetriever(upstream)
	request := retrieval.SearchRequest{Collections: []string{"user:test"}, QueryText: "query", K: 50}
	for range 9 {
		if _, err := cached.SearchTextCollections(context.Background(), request); err != nil {
			t.Fatal(err)
		}
	}
	if upstream.calls != 1 {
		t.Fatalf("contract calls = %d, want 1", upstream.calls)
	}
}

func TestCachedCandidateRankerCallsContractOncePerDistinctScoreVector(t *testing.T) {
	upstream := &countingRanker{}
	cached := newCachedCandidateRanker(upstream)
	request := retrieval.RankRequest{Candidates: []retrieval.Candidate{{ID: "candidate", Score: .5}}, QueryText: "query", K1: 1, K2: 1}
	for range 9 {
		if _, err := cached.RankCandidates(context.Background(), request); err != nil {
			t.Fatal(err)
		}
	}
	request.Candidates[0].Score = .6
	if _, err := cached.RankCandidates(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if upstream.calls != 2 {
		t.Fatalf("contract calls = %d, want 2", upstream.calls)
	}
}

func TestExtractAnchors(t *testing.T) {
	anchors := ExtractAnchors("See https://Example.com/a, /tmp/project/paper.md and foo_bar.baz")
	for _, want := range []string{"url:https://example.com/a", "path:/tmp/project/paper.md", "code:foo_bar.baz"} {
		if _, ok := anchors[want]; !ok {
			t.Fatalf("missing anchor %q in %#v", want, anchors)
		}
	}
}

func TestRunDoesNotExportTextAndRespectsCutoff(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "11111111-1111-1111-1111-111111111111.jsonl")
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 12; index++ {
		role, text := "assistant", "ordinary response"
		if index%2 == 0 {
			role = "user"
		}
		if index == 0 || index == 10 {
			text = "inspect /tmp/private/project/paper.md"
		}
		_, _ = file.WriteString(recordJSON(index, role, text, start.Add(time.Duration(index)*time.Minute)) + "\n")
	}
	_ = file.Close()
	result, err := Run(context.Background(), Config{SessionDir: dir, ConfirmationStart: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), DataEnd: time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC), RuleFrozenAt: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	if got := len(result.Design.Cases); got != 1 {
		t.Fatalf("design cases = %d, want 1", got)
	}
	payload, _ := os.ReadFile(path)
	_ = payload
	encoded, _ := jsonMarshal(result)
	if contains(encoded, "/tmp/private") || contains(encoded, "paper.md") {
		t.Fatal("result leaked transcript text or anchors")
	}
	if contains(encoded, "2026-07-01") {
		t.Fatal("result leaked a production session timestamp")
	}
}

func recordJSON(index int, role, text string, timestamp time.Time) string {
	return `{"type":"message","id":"m` + itoa(index) + `","timestamp":"` + timestamp.Format(time.RFC3339Nano) + `","message":{"role":"` + role + `","content":[{"type":"text","text":"` + text + `"}]}}`
}

func jsonMarshal(value any) (string, error) {
	payload, err := json.Marshal(value)
	return string(payload), err
}

func contains(value, needle string) bool { return strings.Contains(value, needle) }
func itoa(value int) string              { return strconv.Itoa(value) }
