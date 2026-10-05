package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchpublication"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

// ONLY for the closed ingestion-only load experiment. The embedded store has
// other mutation methods that this adapter does not wrap, so this type MUST NOT
// be promoted into runtime code or used as a general coherent store wrapper.
// Installation occurs after initial policy binding and seed ingestion.
type ingestionPublicationExperiment struct {
	store.EventStore
	writer      sync.Mutex
	publication *researchpublication.Publisher
}

type publicationPutFixture struct {
	store.EventStore
	entered, release chan struct{}
	result           store.PutResult
	err              error
}

func (s *publicationPutFixture) Put(context.Context, model.Event, []float32, string) (store.PutResult, error) {
	close(s.entered)
	<-s.release
	return s.result, s.err
}

func TestResearchPublicationPendingAndUncertain(t *testing.T) {
	now := time.Now()
	base := model.Snapshot{RuntimeVersion: 1, EvidenceEpoch: 1}
	next := base
	next.RuntimeVersion++
	next.EvidenceEpoch++
	for _, fail := range []bool{false, true} {
		backend := &publicationPutFixture{entered: make(chan struct{}), release: make(chan struct{}), result: store.PutResult{Snapshot: next}}
		if fail {
			backend.err = errors.New("uncertain commit")
		}
		adapter := &ingestionPublicationExperiment{EventStore: backend, publication: researchpublication.New(base)}
		done := make(chan error, 1)
		go func() {
			_, e := adapter.Put(context.Background(), model.Event{AvailableAt: now.Add(time.Hour)}, nil, "fixture")
			done <- e
		}()
		<-backend.entered
		if !adapter.ResearchPublicationCompatible(context.Background(), base, now) {
			t.Error("future pending state rejected")
		}
		if adapter.ResearchPublicationCompatible(context.Background(), base, now.Add(2*time.Hour)) {
			t.Error("visible pending state accepted")
		}
		close(backend.release)
		e := <-done
		if (e != nil) != fail {
			t.Fatal("unexpected transaction outcome", e)
		}
		if adapter.ResearchPublicationCompatible(context.Background(), base, now) == fail {
			t.Fatal("wrong post-transaction compatibility")
		}
	}
}

type publicationProofFixture struct {
	store.EventStore
	result bool
}

func (s publicationProofFixture) Snapshot(context.Context) model.Snapshot {
	panic("ordinary lock path must not be called")
}
func (s publicationProofFixture) ResearchPublicationCompatible(context.Context, model.Snapshot, time.Time) bool {
	return s.result
}

func TestResearchPublicationBridgeProofRouting(t *testing.T) {
	for _, accept := range []bool{false, true} {
		b := &ResearchFeedbackBridge{temporal: true, service: &Service{store: publicationProofFixture{result: accept}}}
		if b.compatible(context.Background(), model.Snapshot{}, time.Now()) != accept {
			t.Fatal("publication proof ignored")
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if b.compatible(ctx, model.Snapshot{}, time.Now()) {
			t.Fatal("cancelled proof accepted")
		}
	}
}

func (s *ingestionPublicationExperiment) Put(ctx context.Context, event model.Event, vector []float32, digest string) (store.PutResult, error) {
	s.writer.Lock()
	defer s.writer.Unlock()
	ticket, e := s.publication.Begin(researchpublication.Ingestion, event.AvailableAt)
	if e != nil {
		return store.PutResult{}, e
	}
	r, e := s.EventStore.Put(ctx, event, vector, digest)
	if e != nil || r.Duplicate {
		// No optimistic rollback on ambiguous commits or duplicate resolution.
		_ = s.publication.Abort(ticket, true)
		if e == nil {
			e = errors.New("research duplicate requires authoritative resolution")
		}
		return r, e
	}
	if e = s.publication.Commit(ticket, r.Snapshot); e != nil {
		return r, e
	}
	return r, nil
}

func (s *ingestionPublicationExperiment) ResearchPublicationCompatible(ctx context.Context, captured model.Snapshot, asOf time.Time) bool {
	return ctx.Err() == nil && s.publication.Compatible(captured, asOf)
}
