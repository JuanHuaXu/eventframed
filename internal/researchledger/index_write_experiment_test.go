package researchledger

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// The ablation is intentionally unsafe: it measures a required constraint's
// cost, not an alternative implementation of the admission contract.
func TestIndexWriteAblationBoundary(t *testing.T) {
	ctx := context.Background()
	for _, indexed := range []bool{false, true} {
		l, err := Open(t.TempDir() + "/boundary.sqlite")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { l.Close() })
		if indexed {
			if err := l.EnableServiceIdentity(ctx); err != nil {
				t.Fatal(err)
			}
		}
		req := phaseFixture(1, 2)
		req[1].Payload = req[0].Payload
		_, err = l.AppendBatchPrepared(ctx, req)
		if (err != nil) != indexed {
			t.Fatal("unexpected constraint behavior", indexed, err)
		}
		rows, err := l.ReadAfter(ctx, 0, 8)
		want := 2
		if indexed {
			want = 0
		}
		if err != nil || len(rows) != want {
			t.Fatal("atomic boundary", err, len(rows))
		}
	}
}

func TestIndexWriteExperiment(t *testing.T) {
	output := os.Getenv("EVENTFRAME_INDEX_WRITE")
	if output == "" {
		t.Skip("opt-in unsafe-index ablation")
	}
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	names := []string{"go.mod", "go.sum", "docs/experiments/mmm-index-write-v1-contract.md"}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range files {
		names = append(names, "internal/researchledger/"+name)
	}
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range names {
		b, err := os.ReadFile("../../" + name)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		sources[name], hashes[name] = string(b), hex.EncodeToString(h[:])
	}
	if err := enc.Encode(map[string]any{"Sources": sources, "Hashes": hashes, "GoVersion": runtime.Version(), "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for trial := 0; trial < 3; trial++ {
		for _, size := range []int{50, 200} {
			for order := 0; order < 2; order++ {
				indexed := (trial+order)%2 == 1
				path := t.TempDir() + "/index.sqlite"
				l, err := Open(path)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { l.Close() })
				if indexed {
					if err := l.EnableServiceIdentity(ctx); err != nil {
						t.Fatal(err)
					}
				}
				var syncMode int
				var journal string
				if err := l.db.QueryRow("PRAGMA synchronous").Scan(&syncMode); err != nil || syncMode != 2 {
					t.Fatal("sync", err)
				}
				if err := l.db.QueryRow("PRAGMA journal_mode").Scan(&journal); err != nil || journal != "wal" {
					t.Fatal("journal", err)
				}
				samples := []appendPhases{}
				for batch := 0; batch < 32; batch++ {
					req := phaseFixture(batch*size+1, size)
					var p appendPhases
					ack, err := timedPreparedAppend(l, ctx, req, &p, nil)
					if err != nil || len(ack) != size {
						t.Fatal("append", err)
					}
					for i, a := range ack {
						if a.Retry || a.Sequence != int64(batch*size+i+1) {
							t.Fatal("ack")
						}
					}
					if p.TotalNS < p.PreflightNS+p.BeginNS+p.PrepareNS+p.RowsNS+p.CommitNS {
						t.Fatal("phase overlap")
					}
					samples = append(samples, p)
				}
				if err := l.Close(); err != nil {
					t.Fatal(err)
				}
				l, err = Open(path)
				if err != nil {
					t.Fatal(err)
				}
				verified := 0
				for batch := 0; batch < 32; batch++ {
					rows, err := l.ReadAfter(ctx, int64(batch*size), size)
					if err != nil || len(rows) != size {
						t.Fatal("readback", err)
					}
					for i, req := range phaseFixture(batch*size+1, size) {
						e := rows[i]
						if e.Sequence != int64(batch*size+i+1) || e.Key != req.Key || e.Kind != req.Kind || string(e.Payload) != string(req.Payload) {
							t.Fatal("reopened mismatch")
						}
						verified++
					}
				}
				if err := l.Close(); err != nil {
					t.Fatal(err)
				}
				if err := enc.Encode(map[string]any{"Trial": trial, "Size": size, "Indexed": indexed, "Verified": verified, "Samples": samples}); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
