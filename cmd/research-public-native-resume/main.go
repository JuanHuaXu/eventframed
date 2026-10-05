// This fit diagnostic owns one explicit public-data socket, never discovers a
// service, and cannot read outcome labels or mixed official query files.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchpublicframe"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicpool"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicresume"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

type query struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type fitInput struct {
	Partition string  `json:"partition"`
	Queries   []query `json:"queries"`
}

func validateFit(f fitInput) error {
	if f.Partition != "fit" || len(f.Queries) != 351 {
		return errors.New("requires exact fit projection")
	}
	seen := make(map[string]bool)
	for _, q := range f.Queries {
		if q.ID == "" || strings.TrimSpace(q.Text) == "" || seen[q.ID] || len(q.Text) > 16384 {
			return errors.New("invalid fit query")
		}
		seen[q.ID] = true
	}
	return nil
}

func ownedRoot(root, cwd string) (string, error) {
	a, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(a, filepath.Join(cwd, "research")+string(os.PathSeparator)) {
		return "", errors.New("native root outside research")
	}
	return a, nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 6 {
		return errors.New("usage: research-public-native-resume CORPUS.json FIT_ONLY.json OWNED_ROOT TRACE.ndjson PREFIX.json")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err := ownedRoot(os.Args[3], cwd)
	if err != nil {
		return err
	}
	// The fourth argument is a fixed output basename to match the launch template.
	if os.Args[4] != filepath.Join(root, "trace.ndjson") {
		return errors.New("unowned output")
	}
	info, err := os.Stat(filepath.Join(root, "native.sock"))
	if err != nil || info.Mode()&os.ModeSocket == 0 {
		return errors.New("owned socket missing")
	}
	fb, err := os.ReadFile(os.Args[2])
	if err != nil {
		return err
	}
	var input fitInput
	if err = json.Unmarshal(fb, &input); err != nil {
		return err
	}
	if err = validateFit(input); err != nil {
		return err
	}
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		return err
	}
	records, err := researchpublicframe.Decode(b)
	if err != nil {
		return err
	}
	if len(records) != 5183 {
		return errors.New("incomplete corpus")
	}
	clock := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	r, err := researchpublicpool.New(context.Background(), records, researchpublicframe.Config{
		TenantID: "research-scifact", SessionID: "public-import", ImportedAt: clock}, "public-scientific")
	if err != nil {
		return err
	}
	var prefix struct {
		IDs []string `json:"ids"`
	}
	pb, err := os.ReadFile(os.Args[5])
	if err != nil {
		return err
	}
	if err = json.Unmarshal(pb, &prefix); err != nil {
		return err
	}
	if err = researchpublicresume.ValidatePlan(r, prefix.IDs); err != nil {
		return err
	}
	reader, err := researchpublicresume.Open(root, cwd)
	if err != nil {
		return err
	}
	defer reader.Close()
	client, err := retrieval.OpenLibraVDBContractsWithConfig(retrieval.ContractClientConfig{
		Endpoint: "unix:" + filepath.Join(root, "native.sock"), TLSMode: "insecure", MaxConcurrent: 1,
		RequestTimeout: 30 * time.Second, MaxAttempts: 1})
	if err != nil {
		return err
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	if err = client.Ready(ctx); err != nil {
		return err
	}
	f, err := os.OpenFile(os.Args[4], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	write := func(v any) error {
		if err := enc.Encode(v); err != nil {
			return err
		}
		return f.Sync()
	}
	started := time.Now()
	if err = write(map[string]any{"kind": "start", "documents": len(records), "queries": len(input.Queries), "partition": "fit",
		"searchContract": client.RetrievalContractName(), "rankContract": client.ContractName()}); err != nil {
		return err
	}
	if err = researchpublicresume.Verify(ctx, r, prefix.IDs, clock, reader, func(i int, e researchpublicpool.Entry, rows []retrieval.Candidate, ns int64) error {
		return write(map[string]any{"kind": "verify", "index": i, "id": e.Candidate.ID, "source": e.SourceID, "rows": rows, "ns": ns, "ok": true})
	}); err != nil {
		return err
	}
	for i, e := range r.Entries() {
		if i < len(prefix.IDs) {
			continue
		}
		t := time.Now()
		err = client.EnsureText(ctx, r.Collection(), e.Candidate, "source_sha256", e.SourceSHA256)
		row := map[string]any{"kind": "import", "index": i, "id": e.Candidate.ID, "source": e.SourceID, "ns": time.Since(t).Nanoseconds(), "ok": err == nil}
		if err != nil {
			row["error"] = err.Error()
		}
		if werr := write(row); werr != nil {
			return werr
		}
		if err != nil {
			return err
		}
		if (i+1)%100 == 0 {
			fmt.Printf("imported %d/%d\n", i+1, len(records))
		}
	}
	importNS := time.Since(started).Nanoseconds()
	for i, q := range input.Queries {
		t := time.Now()
		rows, err := client.SearchTextCollections(ctx, retrieval.SearchRequest{Collections: []string{r.Collection()}, QueryText: q.Text, K: 200})
		if err != nil {
			return err
		}
		searchNS := time.Since(t).Nanoseconds()
		t = time.Now()
		bound, err := r.Bind(ctx, rows, clock, 200)
		if err != nil {
			return err
		}
		bindNS := time.Since(t).Nanoseconds()
		ranked := []retrieval.Candidate{}
		rankNS := int64(0)
		rankBindNS := int64(0)
		if len(rows) > 0 {
			t = time.Now()
			ranked, err = client.RankCandidates(ctx, retrieval.RankRequest{Candidates: rows, QueryText: q.Text,
				SessionID: "public-native-fit-v1", UserID: "research-scifact", K1: len(rows), K2: len(rows)})
			if err != nil {
				return err
			}
			rankNS = time.Since(t).Nanoseconds()
			t = time.Now()
			bound, err = r.BindRanked(ctx, rows, ranked, clock, 200)
			if err != nil {
				return err
			}
			rankBindNS = time.Since(t).Nanoseconds()
		}
		sources := make([]string, len(bound))
		for j, e := range bound {
			sources[j] = e.SourceID
		}
		if err = write(map[string]any{"kind": "query", "index": i, "id": q.ID, "text": q.Text, "k": 200, "k1": len(rows), "k2": len(rows),
			"search": rows, "ranked": ranked, "sources": sources, "searchNS": searchNS, "bindNS": bindNS, "rankNS": rankNS, "rankBindNS": rankBindNS}); err != nil {
			return err
		}
		if (i+1)%25 == 0 {
			fmt.Printf("queried %d/%d\n", i+1, len(input.Queries))
		}
	}
	if err = write(map[string]any{"kind": "complete", "documents": len(records), "queries": len(input.Queries), "importNS": importNS,
		"totalNS": time.Since(started).Nanoseconds(), "labelsRead": false, "confirmationPredictions": 0, "verifiedPrefix": len(prefix.IDs)}); err != nil {
		return err
	}
	fmt.Printf("complete documents=%d fit_queries=%d no_labels=true\n", len(records), len(input.Queries))
	return nil
}
