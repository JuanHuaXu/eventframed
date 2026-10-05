package libravdbstore

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
)

// The returned validator is PRIVATE to one owner-held batch. It is never cached
// across publication, and checks exactly the original chain, not only a suffix.
func (s *batchArchiveV34) validatorV34(ctx context.Context, current model.Snapshot) (func(*archiveCaptureV29) error, error) {
	if !s.sharedAncestry {
		return func(c *archiveCaptureV29) error { s.scanPasses++; return s.validate(ctx, c, current) }, nil
	}
	s.scanPasses++
	head := s.state.Base
	root := witnessGenesisV23(head).Hash
	stateHash := root
	roots := map[string]model.Snapshot{root: head}
	rows, err := s.gate.sidecar.QueryContext(ctx, "SELECT payload,prior,digest FROM witness_v23 ORDER BY seq")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
		if count > 11000 {
			return nil, errors.New("archive history cap")
		}
		var payload, prior, digest string
		if err := rows.Scan(&payload, &prior, &digest); err != nil {
			return nil, err
		}
		var tr witnessTransitionV23
		if err := json.Unmarshal([]byte(payload), &tr); err != nil {
			return nil, err
		}
		if prior != root || tr.Before != head || digest != witnessHashV23([]string{prior, payload}) {
			return nil, errors.New("archive chain gap")
		}
		root, head, stateHash = digest, tr.After, tr.StateHash
		roots[root] = head
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if head != current || root != s.state.Hash || stateHash != witnessStateHashV23(*s.state) {
		return nil, errors.New("archive ancestry mismatch")
	}
	return func(c *archiveCaptureV29) error {
		if c == nil || c.owner != s.archiveWitnessV29 || c.root == "" || c.wire != witnessHashV23(c.entry) || c.seal != witnessHashV23([]any{c.entry, c.binding, c.root}) || c.binding.Snapshot != c.entry.Snapshot || c.binding.SourceAt != c.entry.AsOf || s.poison.Load() || current != s.state.Head || current != s.gate.store.Snapshot(ctx) {
			return errors.New("invalid archive authority")
		}
		if snap, ok := roots[c.root]; !ok || snap != c.entry.Snapshot {
			return errors.New("archive ancestry mismatch")
		}
		return nil
	}, nil
}

type loadArchiveV34 struct {
	*joinedWitnessV25
	archive *batchArchiveV34
	mode    int
}

func attachLoadArchiveV34(t *testing.T, f *witnessFixtureV23, mode int) *loadArchiveV34 {
	if mode < -1 || mode > 3 {
		t.Fatal("invalid factorial mode")
	}
	if mode == -1 {
		return &loadArchiveV34{joinedWitnessV25: attachCombinedV26(t, f, true), mode: mode}
	}
	a := attachBatchArchiveV34(t, f, 128, mode&1 != 0, mode&2 != 0)
	return &loadArchiveV34{joinedWitnessV25: &joinedWitnessV25{scheduledWitnessV24: a.scheduledWitnessV24, joined: true}, archive: a, mode: mode}
}
func (s *loadArchiveV34) recall(ctx context.Context, svc *service.Service, r model.RecallRequest) (model.ContextPacket, *publishedRecallViewV8, error) {
	if s.archive != nil {
		return s.archive.recall(ctx, svc, r)
	}
	return s.joinedWitnessV25.recall(ctx, svc, r)
}
func (s *loadArchiveV34) closeV34() error {
	if s.archive != nil {
		return s.archive.Close()
	}
	return s.joinedWitnessV25.Close()
}

type archiveTrialV34 struct {
	joinedTrialV25
	Mode                                     int
	Archived, SharedAncestry, DeferredProofs bool
	Captures                                 []archiveProofV34
	ArchiveQueue                             archiveQueueStatsV34
	AncestryScans, ProofBeforeAck            int
	PostDrainProofNS                         int64
}

// Caller has joined ALL serving calls before entering this diagnostic function.
// Ownership copies, seals, native readback and marker commit remain before ack.
func finishArchiveTrialV34(s *loadArchiveV34, r joinedTrialV25) archiveTrialV34 {
	if s.archive == nil {
		return archiveTrialV34{joinedTrialV25: r, Mode: s.mode}
	}
	a := s.archive
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.stats.Active != 0 || a.stats.Accepted != a.stats.Finished {
		panic("proof materialization before serving drain")
	}
	start := time.Now()
	for _, job := range a.pendingProofs {
		c := job.capture
		encoded, err := json.Marshal(c.entry)
		if err != nil {
			panic(err)
		}
		a.proofs = append(a.proofs, archiveProofV34{ID: c.entry.ID, Root: c.root, Wire: c.wire, Encoded: string(encoded), Snapshot: c.entry.Snapshot, CapturedAt: job.capturedAt})
	}
	a.pendingProofs = nil
	proofNS := time.Since(start).Nanoseconds()
	r.Batches = append([]joinedBatchV25(nil), a.batches...)
	a.owner.Lock()
	scans := a.scanPasses
	a.owner.Unlock()
	return archiveTrialV34{joinedTrialV25: r, Mode: s.mode, Archived: true, SharedAncestry: a.sharedAncestry, DeferredProofs: a.deferredProofs, Captures: append([]archiveProofV34(nil), a.proofs...), ArchiveQueue: a.stats, AncestryScans: scans, ProofBeforeAck: a.proofBeforeAck, PostDrainProofNS: proofNS}
}
