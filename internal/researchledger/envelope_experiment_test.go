package researchledger

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Storage-only surrogate. These generic rows are never passed to a learner;
// they intentionally do not implement source uniqueness or acceptance policy.
type envelopeStorageResult struct {
	Trial, Size, Batches, Verified int
	Mode                           string
	PrepareNS, FinalNS             []int64
}

func envelopeStorageCell(t *testing.T, trial, size, batches int, mode string) envelopeStorageResult {
	t.Helper()
	ctx := context.Background()
	path := t.TempDir() + "/envelope.sqlite"
	l, err := Open(path)
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
	r := envelopeStorageResult{Trial: trial, Size: size, Batches: batches, Mode: mode}
	var originals []AppendRequest
	var sequences []int64
	for batch := 0; batch < batches; batch++ {
		payloads := make([]json.RawMessage, size)
		for i := range payloads {
			label := fmt.Sprintf("event-%06d-", batch*size+i)
			payloads[i] = json.RawMessage(`"` + label + strings.Repeat("x", 1022-len(label)) + `"`)
			if len(payloads[i]) != 1024 {
				t.Fatal("payload length")
			}
		}
		envelope, err := json.Marshal(payloads)
		if err != nil {
			t.Fatal(err)
		}
		key := Key{"public-storage-fixture", "envelope", fmt.Sprint(batch), "envelope-storage-v1"}
		var prepare, final []AppendRequest
		switch mode {
		case "rows":
			for i, payload := range payloads {
				k := key
				k.Event = fmt.Sprint(batch*size + i)
				final = append(final, AppendRequest{k, "admit", payload})
			}
		case "envelope":
			final = []AppendRequest{{key, "admit", envelope}}
		case "staged":
			prepare = []AppendRequest{{key, "admit", envelope}}
			digest := sha256.Sum256(envelope)
			marker, err := json.Marshal(struct {
				Digest string
				Count  int
			}{hex.EncodeToString(digest[:]), size})
			if err != nil {
				t.Fatal(err)
			}
			key.Journal = "marker"
			final = []AppendRequest{{key, "admit", marker}}
		default:
			t.Fatal("unknown mode")
		}
		appendTimed := func(requests []AppendRequest) int64 {
			start := time.Now()
			acks, err := l.AppendBatchPrepared(ctx, requests)
			elapsed := time.Since(start).Nanoseconds()
			if err != nil || len(acks) != len(requests) {
				t.Fatal("append", err)
			}
			for i, ack := range acks {
				if ack.Retry || ack.Sequence != int64(len(originals)+1) {
					t.Fatal("ack sequence", ack)
				}
				originals = append(originals, requests[i])
				sequences = append(sequences, ack.Sequence)
			}
			return elapsed
		}
		var prepareNS int64
		if len(prepare) > 0 {
			prepareNS = appendTimed(prepare)
		}
		r.PrepareNS = append(r.PrepareNS, prepareNS)
		r.FinalNS = append(r.FinalNS, appendTimed(final))
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	for i, request := range originals {
		entry, err := reopened.Get(ctx, request.Key, request.Kind)
		if err != nil || entry.Sequence != sequences[i] || !bytes.Equal(entry.Payload, request.Payload) {
			t.Fatal("reopen", i, err)
		}
		r.Verified++
	}
	last := originals[len(originals)-1]
	ack, err := reopened.AppendBatchPrepared(ctx, []AppendRequest{last})
	if err != nil || len(ack) != 1 || !ack[0].Retry || ack[0].Sequence != sequences[len(sequences)-1] {
		t.Fatal("retry", ack, err)
	}
	last.Payload = json.RawMessage(`{"changed":true}`)
	if _, err := reopened.AppendBatchPrepared(ctx, []AppendRequest{last}); err == nil {
		t.Fatal("conflicting bytes accepted")
	}
	return r
}

func TestEnvelopeStorageAccounting(t *testing.T) {
	for _, mode := range []string{"rows", "envelope", "staged"} {
		t.Run(mode, func(t *testing.T) { envelopeStorageCell(t, 0, 3, 2, mode) })
	}
}

func TestEnvelopeStorageExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ENVELOPE_ARTIFACT")
	if path == "" {
		t.Skip("opt-in storage feasibility experiment")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	names := []string{"go.mod", "go.sum", "docs/experiments/mmm-admission-envelope-v1-contract.md"}
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
	modes := []string{"rows", "envelope", "staged"}
	for trial := 0; trial < 3; trial++ {
		for _, size := range []int{50, 200} {
			for order := 0; order < 3; order++ {
				r := envelopeStorageCell(t, trial, size, 32, modes[(trial+order)%3])
				if err := enc.Encode(r); err != nil {
					t.Fatal(err)
				}
				if err := f.Sync(); err != nil {
					t.Fatal(err)
				}
				t.Logf("trial%d size%d mode%s verified%d", trial, size, r.Mode, r.Verified)
			}
		}
	}
}
