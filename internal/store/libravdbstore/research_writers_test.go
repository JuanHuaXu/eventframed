package libravdbstore_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
)

func researchJournal(snapshot model.Snapshot, i int) model.BayesianJournalEntry {
	id := fmt.Sprintf("public-journal-%d", i)
	r := model.BayesianJournalEntry{ID: id, TenantID: "tenant-a", SessionID: id, AsOf: time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC), QueryDigest: id, Snapshot: snapshot, Report: model.BayesianShadowReport{Mode: "shadow", JournalID: id, JournalDurable: true, Nominated: 50}}
	// Storage-shaped public payloads, not scored forecasts or learning evidence.
	for j := 0; j < 50; j++ {
		r.Report.Decisions = append(r.Report.Decisions, model.BayesianDecision{EventID: fmt.Sprintf("public-event-%d", j), ActivationScore: .6})
	}
	return r
}

type researchWriterResult struct {
	Four                    bool
	Readers, Trial, Records int
	CallsNS                 []int64
	ElapsedNS               int64
	PayloadBytes            int
	Digest                  string
}

func researchWriterCell(t *testing.T, four bool, readers, trial, n int) researchWriterResult {
	t.Helper()
	ctx := context.Background()
	path := t.TempDir() + "/writers.libravdb"
	open := libravdbstore.Open
	if four {
		open = libravdbstore.OpenResearchFourWriters
	}
	s, err := open(testConfig(path, "public:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if s != nil {
			s.Close()
		}
	}()
	wanted := make([]model.BayesianJournalEntry, n)
	r := researchWriterResult{Four: four, Readers: readers, Trial: trial, Records: n, CallsNS: make([]int64, n)}
	for i := range wanted {
		wanted[i] = researchJournal(s.Snapshot(ctx), i)
		b, e := json.Marshal(wanted[i])
		if e != nil {
			t.Fatal(e)
		}
		r.PayloadBytes += len(b)
	}
	var wg sync.WaitGroup
	start := make(chan struct{})
	errors := make(chan error, n)
	for lane := 0; lane < readers; lane++ {
		wg.Add(1)
		go func(lane int) {
			defer wg.Done()
			<-start
			for i := lane; i < n; i += readers {
				begin := time.Now()
				e := s.PutBayesianJournal(ctx, wanted[i])
				r.CallsNS[i] = time.Since(begin).Nanoseconds()
				if e != nil {
					errors <- e
				}
			}
		}(lane)
	}
	begin := time.Now()
	close(start)
	wg.Wait()
	r.ElapsedNS = time.Since(begin).Nanoseconds()
	close(errors)
	for e := range errors {
		t.Fatal(e)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	// Reopen via the default constructor to check format and receipt compatibility.
	s, err = libravdbstore.Open(testConfig(path, "public:d4"))
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	enc := json.NewEncoder(h)
	for _, want := range wanted {
		got, e := s.GetBayesianJournal(ctx, want.TenantID, want.ID)
		if e != nil || !reflect.DeepEqual(got, want) {
			t.Fatal("journal mismatch", e)
		}
		if e = enc.Encode(got); e != nil {
			t.Fatal(e)
		}
	}
	r.Digest = hex.EncodeToString(h.Sum(nil))
	if err = s.PutBayesianJournal(ctx, wanted[0]); err != nil {
		t.Fatal("retry", err)
	}
	bad := wanted[0]
	bad.QueryDigest = "conflicting"
	if err = s.PutBayesianJournal(ctx, bad); err == nil {
		t.Fatal("conflict accepted")
	}
	return r
}

func TestResearchFourWritersIntegrity(t *testing.T) {
	a := researchWriterCell(t, false, 4, 0, 32)
	b := researchWriterCell(t, true, 4, 0, 32)
	if a.Digest != b.Digest || a.PayloadBytes != b.PayloadBytes {
		t.Fatal("different journal contract")
	}
}

func TestResearchFourWritersAbruptExit(t *testing.T) {
	if path := os.Getenv("EVENTFRAME_FOUR_WRITER_CHILD"); path != "" {
		s, e := libravdbstore.OpenResearchFourWriters(testConfig(path, "public:d4"))
		if e != nil {
			t.Fatal(e)
		}
		snapshot := s.Snapshot(context.Background())
		var wg sync.WaitGroup
		for i := 0; i < 16; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if e := s.PutBayesianJournal(context.Background(), researchJournal(snapshot, i)); e != nil {
					t.Error(e)
				}
			}(i)
		}
		wg.Wait()
		if t.Failed() {
			os.Exit(1)
		}
		os.Exit(0) // Deliberately no Close/checkpoint.
	}
	path := t.TempDir() + "/abrupt.libravdb"
	bin, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	cmd := exec.Command(bin, "-test.run=^TestResearchFourWritersAbruptExit$")
	cmd.Env = append(os.Environ(), "EVENTFRAME_FOUR_WRITER_CHILD="+path)
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("child: %v %s", e, out)
	}
	s, e := libravdbstore.Open(testConfig(path, "public:d4"))
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	for i := 0; i < 16; i++ {
		want := researchJournal(s.Snapshot(context.Background()), i)
		got, e := s.GetBayesianJournal(context.Background(), want.TenantID, want.ID)
		if e != nil || !reflect.DeepEqual(got, want) {
			t.Fatal("ack lost after exit", e)
		}
	}
}

func TestResearchWriterCostExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_WRITER_COST_ARTIFACT")
	if path == "" {
		t.Skip("opt-in writer cost diagnostic")
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range []string{"go.mod", "go.sum", "internal/model/api.go", "internal/store/libravdbstore/store.go", "internal/store/libravdbstore/store_test.go", "internal/store/libravdbstore/research_writers.go", "internal/store/libravdbstore/research_writers_test.go", "docs/experiments/mmm-writer-cost-v61-protocol.md"} {
		b, e := os.ReadFile("../../../" + name)
		if e != nil {
			t.Fatal(e)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if e = enc.Encode(map[string]any{"Kind": "header", "Cells": 12, "Go": runtime.Version(), "GOMAXPROCS": runtime.GOMAXPROCS(0), "Sources": sources, "Hashes": hashes}); e != nil {
		t.Fatal(e)
	}
	for _, readers := range []int{1, 4} {
		for trial := 0; trial < 3; trial++ {
			var digest string
			for order := 0; order < 2; order++ {
				r := researchWriterCell(t, (trial+order)%2 == 1, readers, trial, 128)
				if e = enc.Encode(r); e != nil {
					t.Fatal(e)
				}
				if e = f.Sync(); e != nil {
					t.Fatal(e)
				}
				if digest != "" && r.Digest != digest {
					t.Fatal("paired records differ")
				}
				digest = r.Digest
			}
		}
	}
}
