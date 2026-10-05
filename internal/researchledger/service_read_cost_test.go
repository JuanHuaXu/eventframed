package researchledger

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

type serviceReadCostResult struct {
	Kind                                                     string
	Size, Trial, History, Groups, Hits, Misses, PayloadBytes int
	Batch, Hit                                               bool
	ReadNS                                                   []int64
}

func serviceReadCostArm(t *testing.T, size, trial int, batch, hit bool) serviceReadCostResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	l, err := Open(t.TempDir() + "/cost.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if err = l.EnableServiceIdentity(ctx); err != nil {
		t.Fatal(err)
	}
	r := serviceReadCostResult{Kind: "result", Size: size, Trial: trial, History: 1000, Groups: 32, Batch: batch, Hit: hit}
	keys := make([]ServiceIdentity, r.History)
	fixtures := make([]AppendRequest, r.History)
	for i := range keys {
		k := sourceFixture()
		k.Journal = fmt.Sprintf("journal-%d", i/50)
		k.Event = fmt.Sprintf("event-%d", i)
		keys[i] = k
		record := serviceEntry(fmt.Sprint(i+1), k)
		var value map[string]any
		if err = json.Unmarshal(record.Payload, &value); err != nil {
			t.Fatal(err)
		}
		value["padding"] = ""
		raw, _ := json.Marshal(value)
		value["padding"] = strings.Repeat("x", 1024-len(raw))
		record.Payload, _ = json.Marshal(value)
		if len(record.Payload) != 1024 {
			t.Fatal("fixture size")
		}
		fixtures[i] = record
	}
	for i := 0; i < len(fixtures); i += 200 {
		if _, err = l.AppendBatchPrepared(ctx, fixtures[i:i+200]); err != nil {
			t.Fatal(err)
		}
	}
	for group := 0; group < r.Groups; group++ {
		requests := make([]ServiceIdentity, size)
		indices := make([]int, size)
		for i := range requests {
			index := (group*size + i) * 7919 % r.History
			indices[i] = index
			requests[i] = keys[index]
			if !hit {
				requests[i].Event = "missing-" + requests[i].Event
			}
		}
		start := time.Now()
		var got []ServiceLookupResult
		if batch {
			got, err = l.GetServiceAdmissions(ctx, requests)
		} else {
			got = make([]ServiceLookupResult, size)
			for i, k := range requests {
				entry, e := l.GetServiceAdmission(ctx, k)
				if errors.Is(e, sql.ErrNoRows) {
					got[i] = ServiceLookupResult{Source: k}
				} else if e != nil {
					err = e
					break
				} else {
					got[i] = ServiceLookupResult{Source: k, Entry: entry, Found: true}
				}
			}
		}
		r.ReadNS = append(r.ReadNS, time.Since(start).Nanoseconds())
		if err != nil || len(got) != size {
			t.Fatal("read failed", err)
		}
		for i, v := range got {
			if v.Source != requests[i] || v.Found != hit {
				t.Fatal("source/missing mismatch", group, i)
			}
			if hit {
				want := fixtures[indices[i]]
				if v.Entry.Sequence <= 0 || v.Entry.Key != want.Key || v.Entry.Kind != "admit" || !bytes.Equal(v.Entry.Payload, want.Payload) {
					t.Fatal("original mismatch", group, i)
				}
				r.Hits++
				r.PayloadBytes += len(v.Entry.Payload)
			} else {
				if !reflect.DeepEqual(v.Entry, Entry{}) {
					t.Fatal("nonzero miss")
				}
				r.Misses++
			}
		}
	}
	return r
}

func TestServiceReadCostStudy(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SERVICE_READ_COST_ARTIFACT")
	if path == "" {
		t.Skip("opt-in transactional source read costs")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range []string{"go.mod", "go.sum", "internal/researchledger/service_identity.go", "internal/researchledger/service_identity_test.go", "internal/researchledger/service_read_batch.go", "internal/researchledger/service_read_batch_test.go", "internal/researchledger/service_read_cost_test.go", "internal/researchledger/ledger.go", "internal/researchledger/batch.go", "docs/experiments/mmm-source-reads-v52-protocol.md"} {
		b, e := os.ReadFile("../../" + name)
		if e != nil {
			t.Fatal(e)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if err = enc.Encode(map[string]any{"Kind": "header", "ExpectedCells": 24, "Sources": sources, "Hashes": hashes, "Go": runtime.Version(), "OS": runtime.GOOS, "Arch": runtime.GOARCH, "CPUs": runtime.NumCPU(), "GOMAXPROCS": runtime.GOMAXPROCS(0)}); err != nil {
		t.Fatal(err)
	}
	for trial := 0; trial < 3; trial++ {
		for _, size := range []int{50, 200} {
			for _, hit := range []bool{true, false} {
				for order := 0; order < 2; order++ {
					r := serviceReadCostArm(t, size, trial, (trial+order)%2 == 1, hit)
					if err = enc.Encode(r); err != nil {
						t.Fatal(err)
					}
					if err = f.Sync(); err != nil {
						t.Fatal(err)
					}
					t.Logf("trial%d size%d hit%t batch%t hits%d misses%d", trial, size, hit, r.Batch, r.Hits, r.Misses)
				}
			}
		}
	}
}
