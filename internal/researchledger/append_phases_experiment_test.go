package researchledger

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func phaseFixture(start, size int) []AppendRequest {
	r := make([]AppendRequest, size)
	for i := range r {
		r[i] = envelopeReferenceFixture(start+i, start+i)
		var body map[string]any
		if err := json.Unmarshal(r[i].Payload, &body); err != nil {
			panic(err)
		}
		body["Padding"] = strings.Repeat("x", 1024)
		var err error
		r[i].Payload, err = json.Marshal(body)
		if err != nil {
			panic(err)
		}
	}
	return r
}

func TestAppendPhasesParity(t *testing.T) {
	ctx := context.Background()
	var logs [2]*Ledger
	for i := range logs {
		l, err := Open(t.TempDir() + "/phases.sqlite")
		if err != nil {
			t.Fatal(err)
		}
		logs[i] = l
		t.Cleanup(func() { l.Close() })
		if err = l.EnableServiceIdentity(ctx); err != nil {
			t.Fatal(err)
		}
	}
	for _, req := range [][]AppendRequest{phaseFixture(1, 4), phaseFixture(1, 4), phaseFixture(5, 3)} {
		a, err := logs[0].AppendBatchPrepared(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		var trace appendPhases
		b, err := timedPreparedAppend(logs[1], ctx, req, &trace, nil)
		if err != nil || !reflect.DeepEqual(a, b) {
			t.Fatal("ack parity", err)
		}
		if trace.TotalNS < trace.PreflightNS+trace.BeginNS+trace.PrepareNS+trace.RowsNS+trace.CommitNS {
			t.Fatal("overlapping phase clocks")
		}
	}
	for i, l := range logs {
		cancelCtx, cancel := context.WithCancel(ctx)
		var err error
		if i == 0 {
			_, err = l.appendBatchPlan(cancelCtx, phaseFixture(8, 2), true, false, false, cancel)
		} else {
			var trace appendPhases
			_, err = timedPreparedAppend(l, cancelCtx, phaseFixture(8, 2), &trace, cancel)
		}
		cancel()
		if err == nil {
			t.Fatal("canceled commit")
		}
	}
	bad := phaseFixture(8, 2)
	bad[1].Payload = phaseFixture(1, 1)[0].Payload
	for i, l := range logs {
		var err error
		if i == 0 {
			_, err = l.AppendBatchPrepared(ctx, bad)
		} else {
			var trace appendPhases
			_, err = timedPreparedAppend(l, ctx, bad, &trace, nil)
		}
		if err == nil {
			t.Fatal("source conflict accepted")
		}
	}
	a, err := logs[0].ReadAfter(ctx, 0, 32)
	if err != nil {
		t.Fatal(err)
	}
	b, err := logs[1].ReadAfter(ctx, 0, 32)
	if err != nil || !reflect.DeepEqual(a, b) || len(a) != 7 {
		t.Fatal("rollback parity", err)
	}
}

type appendPhaseRecord struct {
	Trial, Size, Verified int
	Measured              bool
	Samples               []appendPhases
}

func TestAppendPhasesExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_APPEND_PHASES")
	if path == "" {
		t.Skip("opt-in admission phase timing")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	names := []string{"go.mod", "go.sum", "docs/experiments/mmm-append-phases-v1-contract.md", "research/append-phases-summary.mjs"}
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
	if err := enc.Encode(map[string]any{"Sources": sources, "Hashes": hashes, "GoVersion": runtime.Version(), "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for trial := 0; trial < 3; trial++ {
		for _, size := range []int{50, 200} {
			for order := 0; order < 2; order++ {
				measured := (trial+order)%2 == 1
				path := t.TempDir() + "/measure.sqlite"
				l, err := Open(path)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { l.Close() })
				if err = l.EnableServiceIdentity(ctx); err != nil {
					t.Fatal(err)
				}
				var syncMode int
				var journal string
				if err = l.db.QueryRow("PRAGMA synchronous").Scan(&syncMode); err != nil || syncMode != 2 {
					t.Fatal("sync", err)
				}
				if err = l.db.QueryRow("PRAGMA journal_mode").Scan(&journal); err != nil || journal != "wal" {
					t.Fatal("journal", err)
				}
				r := appendPhaseRecord{Trial: trial, Size: size, Measured: measured}
				for batch := 0; batch < 32; batch++ {
					req := phaseFixture(batch*size+1, size)
					var p appendPhases
					var ack []AppendResult
					if measured {
						ack, err = timedPreparedAppend(l, ctx, req, &p, nil)
					} else {
						start := time.Now()
						ack, err = l.AppendBatchPrepared(ctx, req)
						p.TotalNS = time.Since(start).Nanoseconds()
					}
					if err != nil || len(ack) != size {
						t.Fatal("append", err)
					}
					for i, a := range ack {
						if a.Retry || a.Sequence != int64(batch*size+i+1) {
							t.Fatal("sequence")
						}
					}
					r.Samples = append(r.Samples, p)
				}
				if err = l.Close(); err != nil {
					t.Fatal(err)
				}
				l, err = Open(path)
				if err != nil {
					t.Fatal(err)
				}
				for batch := 0; batch < 32; batch++ {
					for _, req := range phaseFixture(batch*size+1, size) {
						_, source, err := envelopeReferenceSource(req)
						if err != nil {
							t.Fatal(err)
						}
						var id ServiceIdentity
						if err := json.Unmarshal([]byte(source), &id); err != nil {
							t.Fatal(err)
						}
						e, err := l.GetServiceAdmission(ctx, id)
						if err != nil || e.Key != req.Key || string(e.Payload) != string(req.Payload) {
							t.Fatal("reopened source", err)
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
			}
		}
	}
}
