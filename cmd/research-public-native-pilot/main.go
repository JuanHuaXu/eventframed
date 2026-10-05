// research-public-native-pilot only connects to its explicit owned Unix socket.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicframe"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicpool"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type queryResult struct {
	Query     string
	Excluded  []string
	Search    []retrieval.Candidate
	Ranked    []retrieval.Candidate
	SourceIDs []string
	SearchNS  int64
	RankNS    int64
}

func main() {
	if len(os.Args) != 4 {
		panic("usage: research-public-native-pilot CORPUS.json OWNED_ROOT NEW.json")
	}
	root, err := filepath.Abs(os.Args[2])
	if err != nil {
		panic(err)
	}
	if !strings.HasPrefix(root, filepath.Join(mustCwd(), "research")+string(os.PathSeparator)) {
		panic("native root outside research")
	}
	socket := filepath.Join(root, "native.sock")
	info, err := os.Stat(socket)
	if err != nil || info.Mode()&os.ModeSocket == 0 {
		panic("owned socket missing")
	}
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	records, err := researchpublicframe.Decode(b)
	if err != nil {
		panic(err)
	}
	records = records[:8]
	clock := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	cfg := researchpublicframe.Config{TenantID: "research-scifact", SessionID: "public-import", ImportedAt: clock}
	r, err := researchpublicpool.New(context.Background(), records, cfg, "public-scientific")
	if err != nil {
		panic(err)
	}
	future := records[0]
	future.ID = "future-only-native-control"
	cfg.ImportedAt = clock.Add(time.Hour)
	later, err := researchpublicpool.New(context.Background(), []researchpublicframe.Record{future}, cfg, r.Collection())
	if err != nil {
		panic(err)
	}
	entries := append(r.Entries(), later.Entries()...)
	futureID := later.Entries()[0].Candidate.ID
	client, err := retrieval.OpenLibraVDBContractsWithConfig(retrieval.ContractClientConfig{Endpoint: "unix:" + socket, TLSMode: "insecure", MaxConcurrent: 1, RequestTimeout: 30 * time.Second, MaxAttempts: 1})
	if err != nil {
		panic(err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if err = client.Ready(ctx); err != nil {
		panic(err)
	}
	start := time.Now()
	for _, e := range entries {
		if err = client.EnsureText(ctx, r.Collection(), e.Candidate, "source_sha256", e.SourceSHA256); err != nil {
			panic(err)
		}
	}
	importNS := time.Since(start).Nanoseconds()
	results := []queryResult{}
	for _, query := range []string{records[0].Title, records[1].Title} {
		req := retrieval.SearchRequest{Collections: []string{r.Collection()}, QueryText: query, K: 50, ExcludeByCollection: map[string][]string{r.Collection(): {futureID}}}
		start = time.Now()
		rows, err := client.SearchTextCollections(ctx, req)
		if err != nil {
			panic(err)
		}
		searchNS := time.Since(start).Nanoseconds()
		if len(rows) == 0 {
			panic("empty native pilot control")
		}
		if _, err = r.Bind(ctx, rows, clock, 50); err != nil {
			panic(err)
		}
		start = time.Now()
		ranked, err := client.RankCandidates(ctx, retrieval.RankRequest{Candidates: rows, QueryText: query, SessionID: "public-native-pilot", UserID: "research-scifact", K1: len(rows), K2: len(rows)})
		if err != nil {
			panic(err)
		}
		rankNS := time.Since(start).Nanoseconds()
		bound, err := r.BindRanked(ctx, rows, ranked, clock, 50)
		if err != nil {
			panic(err)
		}
		ids := []string{}
		for _, v := range bound {
			ids = append(ids, v.SourceID)
		}
		results = append(results, queryResult{Query: query, Excluded: []string{futureID}, Search: rows, Ranked: ranked, SourceIDs: ids, SearchNS: searchNS, RankNS: rankNS})
	}
	// The excluded copy must actually be retrievable without that exclusion;
	// otherwise the availability test would be a vacuous empty-data check.
	raw, err := client.SearchTextCollections(ctx, retrieval.SearchRequest{Collections: []string{r.Collection()}, QueryText: records[0].Title, K: 50})
	if err != nil {
		panic(err)
	}
	found := false
	for _, v := range raw {
		if v.ID == futureID {
			found = true
		}
	}
	if !found {
		panic("future positive control unreachable")
	}
	out := struct {
		Entries        []researchpublicpool.Entry
		Results        []queryResult
		Unexcluded     []retrieval.Candidate
		FutureID       string
		ImportNS       int64
		SearchContract string
		RankContract   string
	}{entries, results, raw, futureID, importNS, client.RetrievalContractName(), client.ContractName()}
	f, err := os.OpenFile(os.Args[3], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(f).Encode(out); err != nil {
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
	fmt.Printf("native documents=%d queries=%d import_ns=%d future_reachable_excluded=true\n", len(entries), len(results), importNS)
}
func mustCwd() string {
	s, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return s
}
