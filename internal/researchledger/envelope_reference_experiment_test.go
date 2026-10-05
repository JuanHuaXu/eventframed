package researchledger

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type envelopeReferenceResult struct {
	Trial, Size, Verified int
	Envelope              bool
	AppendNS              []int64
	LookupNS              int64
}

func TestEnvelopeReferenceExperiment(t *testing.T) {
	artifact := os.Getenv("EVENTFRAME_ENVELOPE_REFERENCE_ARTIFACT")
	if artifact == "" {
		t.Skip("opt-in normalized envelope experiment")
	}
	f, err := os.OpenFile(artifact, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	names := []string{"go.mod", "go.sum", "docs/experiments/mmm-envelope-reference-v1-contract.md"}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		names = append(names, "internal/researchledger/"+file)
	}
	for _, name := range names {
		b, err := os.ReadFile("../../" + name)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		sources[name] = string(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(map[string]any{"Kind": "header", "Sources": sources, "Hashes": hashes, "GoVersion": runtime.Version(), "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "GOMAXPROCS": runtime.GOMAXPROCS(0)}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for trial := 0; trial < 3; trial++ {
		for _, size := range []int{50, 200} {
			for order := 0; order < 2; order++ {
				envelope := (trial+order)%2 == 1
				path := t.TempDir() + "/reference.sqlite"
				l, err := Open(path)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { l.Close() })
				if envelope {
					err = envelopeReferenceSchema(ctx, l)
				} else {
					err = l.EnableServiceIdentity(ctx)
				}
				if err != nil {
					t.Fatal(err)
				}
				var syncMode int
				var journal string
				if err := l.db.QueryRow("PRAGMA synchronous").Scan(&syncMode); err != nil || syncMode != 2 {
					t.Fatal(syncMode, err)
				}
				if err := l.db.QueryRow("PRAGMA journal_mode").Scan(&journal); err != nil || journal != "wal" {
					t.Fatal(journal, err)
				}
				r := envelopeReferenceResult{Trial: trial, Size: size, Envelope: envelope}
				var originals []AppendRequest
				for batch := 0; batch < 32; batch++ {
					requests := make([]AppendRequest, size)
					for i := range requests {
						requests[i] = envelopeReferenceFixture(batch*size+i+1, batch*size+i+1)
						var body map[string]any
						if err := json.Unmarshal(requests[i].Payload, &body); err != nil {
							t.Fatal(err)
						}
						body["Padding"] = strings.Repeat("x", 1024)
						requests[i].Payload, err = json.Marshal(body)
						if err != nil {
							t.Fatal(err)
						}
					}
					start := time.Now()
					var acks []AppendResult
					if envelope {
						acks, err = envelopeReferenceAppend(ctx, l, requests, nil)
					} else {
						acks, err = l.AppendBatchPrepared(ctx, requests)
					}
					r.AppendNS = append(r.AppendNS, time.Since(start).Nanoseconds())
					if err != nil || len(acks) != size {
						t.Fatal("append", err)
					}
					for i, a := range acks {
						if a.Retry || a.Sequence != int64(batch*size+i+1) {
							t.Fatal("ack", a)
						}
					}
					originals = append(originals, requests...)
				}
				if err := l.Close(); err != nil {
					t.Fatal(err)
				}
				l, err = Open(path)
				if err != nil {
					t.Fatal(err)
				}
				// Lookup each original by its source, not by a conveniently known offset.
				for i, original := range originals {
					_, source, err := envelopeReferenceSource(original)
					if err != nil {
						t.Fatal(err)
					}
					var sourceKey ServiceIdentity
					if err := json.Unmarshal([]byte(source), &sourceKey); err != nil {
						t.Fatal(err)
					}
					start := time.Now()
					var seq int64
					var payload []byte
					if envelope {
						lookup := strings.Replace(envelopeReferenceLookup, "WHERE r.identity=?", "WHERE r.source=?", 1)
						var storedSource string
						seq, storedSource, payload, err = envelopeReferenceRead(l.db.QueryRowContext(ctx, lookup, source))
						if storedSource != source {
							t.Fatal("source mismatch")
						}
					} else {
						var entry Entry
						entry, err = l.GetServiceAdmission(ctx, sourceKey)
						seq, payload = entry.Sequence, entry.Payload
					}
					r.LookupNS += time.Since(start).Nanoseconds()
					if err != nil || seq != int64(i+1) || !bytes.Equal(payload, original.Payload) {
						t.Fatal("reopen", i, err)
					}
					r.Verified++
				}
				var ack []AppendResult
				if envelope {
					ack, err = envelopeReferenceAppend(ctx, l, originals[:size], nil)
				} else {
					ack, err = l.AppendBatchPrepared(ctx, originals[:size])
				}
				if err != nil || len(ack) != size {
					t.Fatal("retry", err)
				}
				for i, a := range ack {
					if !a.Retry || a.Sequence != int64(i+1) {
						t.Fatal("retry changed", a)
					}
				}
				if err := l.Close(); err != nil {
					t.Fatal(err)
				}
				if err := enc.Encode(r); err != nil {
					t.Fatal(err)
				}
				if err := f.Sync(); err != nil {
					t.Fatal(err)
				}
				t.Logf("trial%d size%d envelope%t verified%d", trial, size, envelope, r.Verified)
			}
		}
	}
}
