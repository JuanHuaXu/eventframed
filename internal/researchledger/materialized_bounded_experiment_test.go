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
	"time"
)

func TestMaterializedBoundedExperiment(t *testing.T) {
	output := os.Getenv("EVENTFRAME_MATERIALIZED_BOUNDED")
	if output == "" {
		t.Skip("opt-in materialized admission experiment")
	}
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	names := []string{"go.mod", "go.sum", "docs/experiments/mmm-materialized-bounded-v1-contract.md"}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		names = append(names, "internal/researchledger/"+file)
	}
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range names {
		b, err := os.ReadFile("../../" + name)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		sources[name] = string(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	if err := enc.Encode(map[string]any{"Sources": sources, "Hashes": hashes, "GoVersion": runtime.Version(), "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for trial := 0; trial < 3; trial++ {
		for _, size := range []int{50, 200} {
			for order := 0; order < 2; order++ {
				candidate := (trial+order)%2 == 1
				path := t.TempDir() + "/source.sqlite"
				l, err := Open(path)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { l.Close() })
				if candidate {
					err = materializedSourceSchema(l)
				} else {
					err = l.EnableServiceIdentity(ctx)
				}
				if err != nil {
					t.Fatal(err)
				}
				var syncMode int
				var journal string
				if err := l.db.QueryRow("PRAGMA synchronous").Scan(&syncMode); err != nil || syncMode != 2 {
					t.Fatal("sync", err)
				}
				if err := l.db.QueryRow("PRAGMA journal_mode").Scan(&journal); err != nil || journal != "wal" {
					t.Fatal("journal", err)
				}
				appends, reads := []int64{}, []int64{}
				for batch := 0; batch < 32; batch++ {
					req := phaseFixture(batch*size+1, size)
					start := time.Now()
					var ack []AppendResult
					if candidate {
						ack, err = materializedSourceAppend(ctx, l, req, nil)
					} else {
						ack, err = l.AppendBatchPrepared(ctx, req)
					}
					appends = append(appends, time.Since(start).Nanoseconds())
					if err != nil || len(ack) != size {
						t.Fatal("append", err)
					}
					for i, a := range ack {
						if a.Retry || a.Sequence != int64(batch*size+i+1) {
							t.Fatal("ack")
						}
					}
				}
				if err := l.Close(); err != nil {
					t.Fatal(err)
				}
				l, err = Open(path)
				if err != nil {
					t.Fatal(err)
				}
				for batch := 0; batch < 32; batch++ {
					for i, req := range phaseFixture(batch*size+1, size) {
						_, source, err := envelopeReferenceSource(req)
						if err != nil {
							t.Fatal(err)
						}
						var id ServiceIdentity
						if err := json.Unmarshal([]byte(source), &id); err != nil {
							t.Fatal(err)
						}
						var e Entry
						start := time.Now()
						if candidate {
							e, err = materializedSourceReadBounded(ctx, l, id)
						} else {
							e, err = l.GetServiceAdmission(ctx, id)
						}
						reads = append(reads, time.Since(start).Nanoseconds())
						if err != nil || e.Sequence != int64(batch*size+i+1) || e.Key != req.Key || string(e.Payload) != string(req.Payload) {
							t.Fatal("readback", err)
						}
					}
				}
				if err := l.Close(); err != nil {
					t.Fatal(err)
				}
				if err := enc.Encode(map[string]any{"Trial": trial, "Size": size, "Candidate": candidate, "AppendNS": appends, "ReadNS": reads, "Verified": len(reads)}); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
