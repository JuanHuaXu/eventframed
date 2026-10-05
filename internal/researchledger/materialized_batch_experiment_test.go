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

type materializedBatchCell struct {
	Trial, Size, Verified int
	Mode, Layout          string
	ReadNS                []int64
}

func TestMaterializedBatchExperiment(t *testing.T) {
	artifact := os.Getenv("EVENTFRAME_MATERIALIZED_BATCH")
	if artifact == "" {
		t.Skip("opt-in batch read experiment")
	}
	f, err := os.OpenFile(artifact, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	names := []string{"go.mod", "go.sum", "docs/experiments/mmm-materialized-batch-v1-contract.md"}
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
	enc := json.NewEncoder(f)
	if err := enc.Encode(map[string]any{"Kind": "header", "Sources": sources, "Hashes": hashes, "GoVersion": runtime.Version(), "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "GOMAXPROCS": runtime.GOMAXPROCS(0)}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	var originals []AppendRequest
	var keys []ServiceIdentity
	for i := 0; i < 6400; i++ {
		r := envelopeReferenceFixture(i+1, i+1)
		var body map[string]any
		json.Unmarshal(r.Payload, &body)
		body["Padding"] = strings.Repeat("x", 1024)
		r.Payload, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		originals = append(originals, r)
		_, s, err := envelopeReferenceSource(r)
		if err != nil {
			t.Fatal(err)
		}
		var k ServiceIdentity
		if err := json.Unmarshal([]byte(s), &k); err != nil {
			t.Fatal(err)
		}
		keys = append(keys, k)
	}
	modes := []string{"indexed", "materialized"}
	for trial := 0; trial < 3; trial++ {
		for order := 0; order < 2; order++ {
			mode := modes[(trial+order)%2]
			l, err := Open(t.TempDir() + "/read.sqlite")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { l.Close() })
			if mode == "indexed" {
				err = l.EnableServiceIdentity(ctx)
			} else {
				err = materializedSourceSchema(l)
			}
			if err != nil {
				t.Fatal(err)
			}
			for batch := 0; batch < 32; batch++ {
				rs := originals[batch*200 : (batch+1)*200]
				if mode == "indexed" {
					_, err = l.AppendBatchPrepared(ctx, rs)
				} else {
					_, err = materializedSourceAppend(ctx, l, rs, nil)
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			for _, size := range []int{50, 200} {
				for _, layout := range []string{"grouped", "scattered"} {
					cell := materializedBatchCell{Trial: trial, Size: size, Mode: mode, Layout: layout}
					for batch := 0; batch < 32; batch++ {
						ids := make([]int, size)
						requests := make([]ServiceIdentity, size)
						for i := range ids {
							if layout == "grouped" {
								ids[i] = batch*200 + i
							} else {
								ids[i] = ((batch+i)%32)*200 + (i / 32)
							}
							requests[i] = keys[ids[i]]
						}
						start := time.Now()
						var results []ServiceLookupResult
						if mode == "indexed" {
							results, err = l.GetServiceAdmissions(ctx, requests)
						} else {
							results, err = materializedSourceBatch(ctx, l, requests, nil)
						}
						cell.ReadNS = append(cell.ReadNS, time.Since(start).Nanoseconds())
						if err != nil || len(results) != size {
							t.Fatal("read", err)
						}
						for i, r := range results {
							if !r.Found || r.Source != requests[i] || r.Entry.Key != originals[ids[i]].Key || r.Entry.Sequence != int64(ids[i]+1) || !bytes.Equal(r.Entry.Payload, originals[ids[i]].Payload) {
								t.Fatal("wrong original")
							}
							cell.Verified++
						}

					}
					if err := enc.Encode(cell); err != nil {
						t.Fatal(err)
					}
					if err := f.Sync(); err != nil {
						t.Fatal(err)
					}
					t.Logf("trial%d %s size%d %s verified%d", trial, mode, size, layout, cell.Verified)
				}
			}
			if err := l.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
}
