//go:build researchpriority

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"github.com/JuanHuaXu/eventframed/internal/researchcalendar"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
)

type sourceArm struct {
	DatabasePath, Mode     string
	Enabled, Lease         bool
	ReadIntervalMS, Repeat int
	Writes                 []struct {
		Error   string
		Entered bool
	}
}
type writeCheck struct {
	ID, OriginalError, LookupError, RetryError                     string
	Entered, Present, PayloadMatches, RetryDuplicate, PresentAfter bool
}
type result struct {
	Source, Copy, SourceHash                         string
	SourceUnchanged                                  bool
	Mode                                             string
	Lease, Enabled                                   bool
	IntervalMS, Repeat                               int
	Before, After                                    model.Snapshot
	Writes                                           []writeCheck
	DirectCount, SearchCount                         int
	DirectMissingFromSearch, SearchMissingFromDirect []string
	Errors                                           []string
	RecallError                                      string
	RecallNominated                                  int
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func fail(err error) {
	if err != nil {
		panic(err)
	}
}
func read(path string, v any) { b, e := os.ReadFile(path); fail(e); fail(json.Unmarshal(b, v)) }
func main() {
	if len(os.Args) != 3 {
		panic("input durable JSON and NEW output JSON required")
	}
	if _, e := os.Stat(os.Args[2]); !os.IsNotExist(e) {
		panic("output exists/inaccessible")
	}
	dir := os.Args[2] + ".stores"
	fail(os.Mkdir(dir, 0700))
	var input struct{ Arms []sourceArm }
	read(os.Args[1], &input)
	var facts []struct{ Text string }
	read("research/public-task-pilot/landing-transfer-v1/corpus.json", &facts)
	if len(input.Arms) != 32 || len(facts) != 4 {
		panic("unexpected fixture")
	}
	em, e := embed.NewHashEmbedder(32)
	fail(e)
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	query := "Which record refutes the claim that Curiosity landed after 2020?"
	vector, e := embed.Query(em, frame.QueryText(query))
	fail(e)
	rows := []result{}
	for _, a := range input.Arms {
		if len(a.Writes) != 16 {
			panic("incomplete writes")
		}
		b, e := os.ReadFile(a.DatabasePath)
		fail(e)
		copyPath := filepath.Join(dir, filepath.Base(a.DatabasePath))
		f, e := os.OpenFile(copyPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		fail(e)
		_, e = f.Write(b)
		fail(e)
		fail(f.Close())
		r := result{Source: a.DatabasePath, Copy: copyPath, SourceHash: digest(b), Mode: a.Mode, Lease: a.Lease, Enabled: a.Enabled, IntervalMS: a.ReadIntervalMS, Repeat: a.Repeat}
		db, e := libravdbstore.Open(libravdbstore.Config{Path: copyPath, Dimension: 32, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
		fail(e)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		r.Before = db.Snapshot(ctx)
		direct := map[string]bool{}
		for i := 0; i < 200; i++ {
			id := fmt.Sprintf("public-%03d", i)
			events, err := db.GetEvents(ctx, "research", []string{id}, now.Add(10*time.Minute))
			if err != nil || len(events) != 1 {
				r.Errors = append(r.Errors, "seed missing: "+id)
			} else {
				direct[id] = true
			}
		}
		requests := make([]model.CaptureTurnRequest, 16)
		for i, w := range a.Writes {
			id := fmt.Sprintf("writer-%03d", i)
			stamp := now
			if a.Mode == "future" {
				stamp = now.Add(2 * time.Minute)
			}
			turn := model.TurnCapture{ID: id, TenantID: "research", SessionID: id, Sequence: 1, UserText: facts[i%4].Text, AssistantText: "Recorded.", OccurredAt: stamp, ObservedAt: stamp, AvailableAt: stamp}
			requests[i] = model.CaptureTurnRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Turn: turn}
			check := writeCheck{ID: id, OriginalError: w.Error, Entered: w.Entered}
			events, err := db.GetEvents(ctx, "research", []string{id}, now.Add(10*time.Minute))
			if err != nil {
				check.LookupError = err.Error()
				if !errors.Is(err, store.ErrEventNotFound) {
					r.Errors = append(r.Errors, "lookup: "+err.Error())
				}
			} else if len(events) == 1 {
				check.Present = true
				direct[id] = true
				want := frame.FromTurn(turn)
				check.PayloadMatches = events[0].FrameText() == want.FrameText() && events[0].Content == want.Content && events[0].AvailableAt.Equal(stamp)
			}
			r.Writes = append(r.Writes, check)
		}
		r.DirectCount = len(direct)
		search, err := db.Search(ctx, "research", vector, now.Add(10*time.Minute), 300)
		if err != nil {
			r.Errors = append(r.Errors, "search: "+err.Error())
		}
		indexed := map[string]bool{}
		for _, v := range search {
			indexed[v.Event.ID] = true
			if !direct[v.Event.ID] {
				r.SearchMissingFromDirect = append(r.SearchMissingFromDirect, v.Event.ID)
			}
		}
		r.SearchCount = len(indexed)
		for id := range direct {
			if !indexed[id] {
				r.DirectMissingFromSearch = append(r.DirectMissingFromSearch, id)
			}
		}
		p := packing.DefaultPolicy()
		p.AdaptiveEnabled = true
		p.DiversityEnabled = true
		s, err := service.New(db, em, service.Config{DefaultRecallK: 200, DefaultPackK: 10, DefaultTokenBudget: 10000, PackingPolicy: p, ResearchTemporalPriority: a.Enabled})
		fail(err)
		packet, err := s.Recall(researchcalendar.Bind(ctx, "research", query), model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "research", SessionID: "recovery-audit", Query: query, AsOf: now.Add(time.Minute), RecallK: 200, PackK: 10, TokenBudget: 10000})
		if err != nil {
			r.RecallError = err.Error()
		}
		r.RecallNominated = packet.BayesianShadow.Nominated
		// Reconcile only after measuring recovered state. Original stores remain untouched.
		for i := range r.Writes {
			w := &r.Writes[i]
			if w.OriginalError != "" {
				response, err := s.CaptureTurn(ctx, requests[i])
				w.RetryDuplicate = response.Duplicate
				if err != nil {
					w.RetryError = err.Error()
				}
			}
			events, err := db.GetEvents(ctx, "research", []string{w.ID}, now.Add(10*time.Minute))
			w.PresentAfter = err == nil && len(events) == 1 && events[0].FrameText() == frame.FromTurn(requests[i].Turn).FrameText()
		}
		r.After = db.Snapshot(ctx)
		fail(s.Close())
		cancel()
		b, e = os.ReadFile(a.DatabasePath)
		fail(e)
		r.SourceUnchanged = digest(b) == r.SourceHash
		rows = append(rows, r)
		fmt.Println("audited", filepath.Base(a.DatabasePath))
	}
	hashes := map[string]string{}
	for _, p := range []string{os.Args[1], "cmd/research-durable-reconcile/main.go", "research/public-task-pilot/DURABLE_RECONCILE_PROTOCOL.md", "internal/store/libravdbstore/store.go", "internal/frame/turn.go", "research/public-task-pilot/incremental-overlay-v1/overlay.json", "research/public-task-pilot/incremental-overlay-v1/source-0.go.txt", "research/public-task-pilot/incremental-overlay-v1/source-3.go.txt"} {
		b, e := os.ReadFile(p)
		fail(e)
		hashes[p] = digest(b)
	}
	f, e := os.OpenFile(os.Args[2], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	fail(e)
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	fail(enc.Encode(struct {
		Hashes  map[string]string
		Results []result
	}{hashes, rows}))
	fail(f.Close())
}
