package researchledger

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

type preparedResult struct {
	Trial, Size         int
	Prepared            bool
	AdmitNS, TerminalNS []int64
	Verified            int
}

func TestPreparedWriteExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_PREPARED_ARTIFACT")
	if path == "" {
		t.Skip("opt-in isolated ledger experiment")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range []string{"go.mod", "go.sum", "internal/researchledger/batch.go", "internal/researchledger/ledger.go", "internal/researchledger/read_batch.go", "internal/researchledger/prepared_experiment_test.go", "docs/experiments/mmm-prepared-v40-protocol.md"} {
		b, err := os.ReadFile("../../" + name)
		if err != nil {
			t.Fatal(err)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if err = enc.Encode(map[string]any{"kind": "header", "expected_arms": 12, "Sources": sources, "Hashes": hashes, "GoVersion": runtime.Version(), "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "GOMAXPROCS": runtime.GOMAXPROCS(0)}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for trial := 0; trial < 3; trial++ {
		for _, size := range []int{50, 200} {
			for order := 0; order < 2; order++ {
				prepared := (trial+order)%2 == 1
				l, err := Open(t.TempDir() + "/experiment.sqlite")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { l.Close() })
				var synchronous int
				var journal string
				if err = l.db.QueryRow("PRAGMA synchronous").Scan(&synchronous); err != nil || synchronous != 2 {
					t.Fatal("FULL", synchronous, err)
				}
				if err = l.db.QueryRow("PRAGMA journal_mode").Scan(&journal); err != nil || journal != "wal" {
					t.Fatal("WAL", journal, err)
				}
				r := preparedResult{Trial: trial, Size: size, Prepared: prepared}
				for batch := 0; batch < 32; batch++ {
					admissions, terminals := make([]AppendRequest, size), make([]AppendRequest, size)
					for i := 0; i < size; i++ {
						key := Key{"public-tenant", "stream", fmt.Sprint(batch*size + i + 1), "storage-v40"}
						admissions[i] = AppendRequest{key, "admit", json.RawMessage(`"` + strings.Repeat("x", 1022) + `"`)}
						terminals[i] = AppendRequest{key, "feedback", json.RawMessage(`{"discard":true}`)}
					}
					write := l.AppendBatch
					if prepared {
						write = l.AppendBatchPrepared
					}
					start := time.Now()
					a, err := write(ctx, admissions)
					r.AdmitNS = append(r.AdmitNS, time.Since(start).Nanoseconds())
					if err != nil {
						t.Fatal(err)
					}
					start = time.Now()
					b, err := write(ctx, terminals)
					r.TerminalNS = append(r.TerminalNS, time.Since(start).Nanoseconds())
					if err != nil {
						t.Fatal(err)
					}
					if len(a) != size || len(b) != size {
						t.Fatal("ack count")
					}
					var lookups []LookupRequest
					requests := append(admissions, terminals...)
					for _, request := range requests {
						lookups = append(lookups, LookupRequest{Key: request.Key, Kind: request.Kind})
					}
					rows, err := l.GetBatch(ctx, lookups)
					if err != nil || len(rows) != 2*size {
						t.Fatal("readback", err)
					}
					acks := append(a, b...)
					for i, row := range rows {
						seq := int64(batch*size*2 + i + 1)
						if !row.Found || row.Entry.Key != requests[i].Key || row.Entry.Kind != requests[i].Kind || !bytes.Equal(row.Entry.Payload, requests[i].Payload) || row.Entry.Sequence != seq || acks[i] != (AppendResult{Sequence: seq}) {
							t.Fatal("original/ack mismatch", i)
						}
						r.Verified++
					}
				}
				if err = l.Close(); err != nil {
					t.Fatal(err)
				}
				if err = enc.Encode(r); err != nil {
					t.Fatal(err)
				}
				if err = f.Sync(); err != nil {
					t.Fatal(err)
				}
				t.Logf("trial%d size%d prepared%t verified%d", trial, size, prepared, r.Verified)
			}
		}
	}
}
