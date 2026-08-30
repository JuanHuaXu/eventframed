package retrieval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func BenchmarkLibraVDBRankCandidates100(b *testing.B) {
	endpoint := os.Getenv("EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT")
	if endpoint == "" {
		b.Skip("EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT is not set")
	}
	ranker, err := OpenLibraVDBRanker(endpoint)
	if err != nil {
		b.Fatal(err)
	}
	defer ranker.Close()
	candidates := make([]Candidate, 100)
	for index := range candidates {
		metadata, marshalErr := json.Marshal(map[string]any{
			"collection": "session:benchmark", "ts": time.Now().Add(-time.Duration(index) * time.Minute).UnixMilli(),
			"authored": false, "access_count": index % 8, "authority": float64(index%10) / 10, "salience": float64((index+3)%10) / 10,
		})
		if marshalErr != nil {
			b.Fatal(marshalErr)
		}
		candidates[index] = Candidate{ID: string(rune(index + 32)), Text: "bounded candidate retrieval text", Score: float64(100-index) / 100, Metadata: metadata}
	}
	request := RankRequest{Candidates: candidates, QueryText: "bounded candidate retrieval", SessionID: "benchmark", UserID: "benchmark", K1: 100, K2: 50}
	b.ResetTimer()
	for b.Loop() {
		if _, err := ranker.RankCandidates(context.Background(), request); err != nil {
			b.Fatal(err)
		}
	}
}

func TestLibraVDBRankCandidatesContract(t *testing.T) {
	endpoint := os.Getenv("EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT")
	if endpoint == "" {
		t.Skip("EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT is not set")
	}
	ranker, err := OpenLibraVDBRanker(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer ranker.Close()
	metadata := func(values map[string]any) []byte {
		encoded, marshalErr := json.Marshal(values)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		return encoded
	}
	now := time.Now().UnixMilli()
	ranked, err := ranker.RankCandidates(context.Background(), RankRequest{
		QueryText: "RecallK service.go", SessionID: "contract-test", UserID: "contract-test", K1: 3, K2: 3,
		Candidates: []Candidate{
			{ID: "semantic", Text: "general retrieval discussion", Score: .91, Metadata: metadata(map[string]any{"collection": "session:contract-test", "ts": now - 86_400_000, "access_count": 0, "authored": false, "authority": .1, "salience": .1})},
			{ID: "exact", Text: "fix RecallK in service.go", Score: .72, Metadata: metadata(map[string]any{"collection": "user:contract-test", "ts": now - 3_600_000, "access_count": 4, "authority": .4, "salience": .8})},
			{ID: "authority", Text: "Never disclose private memory search output", Score: .62, Metadata: metadata(map[string]any{"collection": "global", "ts": now - 604_800_000, "access_count": 10, "authored": true, "authority": 1, "salience": 1})},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(ranked))
	for _, candidate := range ranked {
		ids = append(ids, candidate.ID)
	}
	if len(ranked) != 3 || ranked[0].ID != "exact" || ranked[1].ID != "authority" || ranked[2].ID != "semantic" {
		t.Fatalf("unexpected LibraVDB contract order: %v", ids)
	}
}

func TestLibraVDBRankCandidatesPreservesRequestedCardinality(t *testing.T) {
	endpoint := os.Getenv("EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT")
	if endpoint == "" {
		t.Skip("EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT is not set")
	}
	ranker, err := OpenLibraVDBRanker(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer ranker.Close()

	for _, size := range []int{41, 50, 82, 99} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			candidates := make([]Candidate, size)
			for index := range candidates {
				metadata, marshalErr := json.Marshal(map[string]any{
					"collection": "session:cardinality", "ts": time.Now().Add(-time.Duration(index) * time.Minute).UnixMilli(),
					"authored": false, "access_count": index % 8, "authority": .5, "salience": .5,
				})
				if marshalErr != nil {
					t.Fatal(marshalErr)
				}
				candidates[index] = Candidate{
					ID: fmt.Sprintf("cardinality-%03d", index), Text: fmt.Sprintf("bounded candidate retrieval text %03d", index),
					Score: float64(size-index) / float64(size), Metadata: metadata,
				}
			}
			ranked, rankErr := ranker.RankCandidates(context.Background(), RankRequest{
				Candidates: candidates, QueryText: "bounded candidate retrieval", SessionID: "cardinality",
				UserID: "cardinality", K1: size, K2: size,
			})
			if rankErr != nil {
				t.Fatal(rankErr)
			}
			if len(ranked) != size {
				t.Fatalf("RankCandidates returned %d unique inputs from %d requested candidates", len(ranked), size)
			}
		})
	}
	t.Run("duplicate-text", func(t *testing.T) {
		const size = 82
		candidates := make([]Candidate, size)
		for index := range candidates {
			metadata, marshalErr := json.Marshal(map[string]any{
				"collection": "session:cardinality", "ts": time.Now().Add(-time.Duration(index) * time.Minute).UnixMilli(),
				"authored": false, "access_count": index % 8, "authority": .5, "salience": .5,
			})
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			candidates[index] = Candidate{
				ID: fmt.Sprintf("duplicate-text-%03d", index), Text: fmt.Sprintf("repeated candidate text %03d", index/2),
				Score: float64(size-index) / float64(size), Metadata: metadata,
			}
		}
		ranked, rankErr := ranker.RankCandidates(context.Background(), RankRequest{
			Candidates: candidates, QueryText: "repeated candidate text", SessionID: "cardinality",
			UserID: "cardinality", K1: size, K2: size,
		})
		if rankErr != nil {
			t.Fatal(rankErr)
		}
		if len(ranked) != size {
			t.Fatalf("RankCandidates returned %d unique-ID inputs from %d requested candidates when texts repeat", len(ranked), size)
		}
	})
}

func TestLibraVDBSearchTextCollectionsContract(t *testing.T) {
	endpoint := os.Getenv("EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT")
	if endpoint == "" {
		t.Skip("EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT is not set")
	}
	client, err := OpenLibraVDBRanker(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	collection := "eventframe-contract-search"
	metadata := func(priority float64) []byte {
		encoded, marshalErr := json.Marshal(map[string]any{
			"collection": collection,
			"ts":         time.Now().UnixMilli(),
			"authored":   false,
			"authority":  priority,
			"salience":   priority,
		})
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		return encoded
	}
	exact := Candidate{ID: "contract-search-exact", Text: "eventframe exact magnetic teapot retrieval", Metadata: metadata(.8)}
	other := Candidate{ID: "contract-search-other", Text: "bounded contextual memory candidate", Metadata: metadata(.2)}
	for _, candidate := range []Candidate{exact, other} {
		if err := client.InsertText(context.Background(), collection, candidate); err != nil {
			t.Fatal(err)
		}
	}

	search := func(exclusions map[string][]string) []Candidate {
		results, searchErr := client.SearchTextCollections(context.Background(), SearchRequest{
			Collections: []string{collection}, QueryText: exact.Text, K: 10, ExcludeByCollection: exclusions,
		})
		if searchErr != nil {
			t.Fatal(searchErr)
		}
		return results
	}
	contains := func(candidates []Candidate, id string) bool {
		for _, candidate := range candidates {
			if candidate.ID == id {
				return true
			}
		}
		return false
	}

	if results := search(nil); !contains(results, exact.ID) {
		t.Fatalf("exact contract result absent: %#v", results)
	}
	if results := search(map[string][]string{collection: {exact.ID}}); contains(results, exact.ID) {
		t.Fatalf("excluded contract result returned: %#v", results)
	}
}

func TestLibraVDBCandidateLifecycleContracts(t *testing.T) {
	endpoint := os.Getenv("EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT")
	if endpoint == "" {
		t.Skip("EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT is not set")
	}
	client, err := OpenLibraVDBRanker(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	suffix := time.Now().UnixNano()
	collection := fmt.Sprintf("eventframe-lifecycle-%d", suffix)
	metadata := func(id, digest string) []byte {
		encoded, marshalErr := json.Marshal(map[string]any{
			"collection": collection, "ts": time.Now().UnixMilli(),
			"eventframe_digest": digest, "authority": .5, "salience": .5,
		})
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		return encoded
	}
	first := Candidate{ID: fmt.Sprintf("lifecycle-first-%d", suffix), Text: "eventframe lifecycle exact candidate", Metadata: metadata("first", "digest-first")}
	if err := client.EnsureText(context.Background(), collection, first, "eventframe_digest", "digest-first"); err != nil {
		t.Fatal(err)
	}
	if err := client.EnsureText(context.Background(), collection, first, "eventframe_digest", "digest-first"); err != nil {
		t.Fatalf("idempotent ensure failed: %v", err)
	}
	if err := client.DeleteText(context.Background(), collection, first.ID); err != nil {
		t.Fatal(err)
	}
	assertIdentityAbsent(t, client, collection, "digest-first", first.ID)

	second := Candidate{ID: fmt.Sprintf("lifecycle-second-%d", suffix), Text: "eventframe lifecycle second candidate", Metadata: metadata("second", "digest-second")}
	third := Candidate{ID: fmt.Sprintf("lifecycle-third-%d", suffix), Text: "eventframe lifecycle third candidate", Metadata: metadata("third", "digest-third")}
	for _, candidate := range []Candidate{second, third} {
		if err := client.EnsureText(context.Background(), collection, candidate, "eventframe_digest", map[string]string{second.ID: "digest-second", third.ID: "digest-third"}[candidate.ID]); err != nil {
			t.Fatal(err)
		}
	}
	if err := client.DeleteTextBatch(context.Background(), collection, []string{second.ID, third.ID}); err != nil {
		t.Fatal(err)
	}
	assertIdentityAbsent(t, client, collection, "digest-second", second.ID)
	assertIdentityAbsent(t, client, collection, "digest-third", third.ID)
}

func assertIdentityAbsent(t *testing.T, client *LibraVDBRanker, collection, digest, id string) {
	t.Helper()
	response := &listByMetaResponse{}
	if err := client.connection.Invoke(context.Background(), listByMetaMethod, &listByMetaRequest{
		Collection: collection, Key: "eventframe_digest", Value: digest,
	}, response); err != nil {
		t.Fatal(err)
	}
	for _, result := range response.Results {
		if result != nil && result.ID == id {
			t.Fatalf("deleted record %q remained visible to ListByMeta", id)
		}
	}
}
