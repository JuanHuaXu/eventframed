package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

// Research fixture only: drain already-ready observations, never wait to fill
// a group. The same consumer implements limit=1 as the preparation-order control.
func researchGroupedLoadConsumer(ctx context.Context, s *Service, tap *ResearchFrontierTap, now time.Time, limit int, durable *researchmemory.Durable, result *researchGuardLoadResult, fail func(error)) {
	researchGroupedSourceLoadConsumer(ctx, s, tap, now, limit, durable, nil, result, fail)
}

func researchGroupedSourceLoadConsumer(ctx context.Context, s *Service, tap *ResearchFrontierTap, now time.Time, limit int, durable *researchmemory.Durable, source *researchmemory.SourceOwner, result *researchGuardLoadResult, fail func(error)) {
	a := researchmemory.New(1, 42)
	guard := s.store.(interface {
		WithResearchAsOfSnapshotWait(context.Context, model.Snapshot, time.Time, func() error) error
	})
	for {
		first, err := tap.Take(ctx)
		if err != nil {
			if ctx.Err() != nil {
				fail(err)
			}
			return
		}
		group := []ResearchFrontierObservation{first}
	drain:
		for len(group) < limit {
			select {
			case next, ok := <-tap.queue:
				if !ok {
					break drain
				}
				group = append(group, next)
			default:
				break drain
			}
		}
		result.Attempts += uint64(len(group))
		result.GroupSizes = append(result.GroupSizes, len(group))
		phase := researchGuardPhase{QueueNS: time.Since(first.queuedAt).Nanoseconds()}
		prepare := time.Now()
		records := make([][]researchmemory.RecordedPrediction, len(group))
		var previewIDs []uint64
		requests := make([]model.RecallRequest, len(group))
		for i, in := range group {
			if len(in.Candidates) != 50 || !in.AsOf.Equal(now) {
				fail(fmt.Errorf("invalid grouped fixture frontier"))
				return
			}
			j, err := s.store.GetBayesianJournal(ctx, "tenant-a", in.JournalID)
			if err != nil {
				fail(err)
				return
			}
			requests[i] = model.RecallRequest{TenantID: "tenant-a", SessionID: j.SessionID, Query: "public load fixture", AsOf: now}
			for _, c := range in.Candidates {
				p, err := a.Predict(c.Features, c.Baseline, 1, in.AsOf)
				if err != nil {
					fail(err)
					return
				}
				r, err := a.Record(p.ID)
				if err != nil {
					fail(err)
					return
				}
				r.Binding = &researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: c.EventID, Snapshot: in.Snapshot}
				previewIDs = append(previewIDs, p.ID)
				if durable != nil {
					// Only valid for this cold, unlabeled fixture. Rejected previews
					// consume no durable ID; accepted originals must compare exactly.
					r.Prediction.ID = result.Admits + uint64(len(previewIDs))
				}
				records[i] = append(records[i], r)
			}
		}
		// In grouped modes PrefetchNS includes preparation of original forecasts.
		phase.PrefetchNS = time.Since(prepare).Nanoseconds()
		start := time.Now()
		waitCtx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
		err = guard.WithResearchAsOfSnapshotWait(waitCtx, first.Snapshot, now, func() error {
			phase.Entered = true
			phase.EntryNS = time.Since(start).Nanoseconds()
			callbackStart := time.Now()
			defer func() { phase.CallbackNS = time.Since(callbackStart).Nanoseconds() }()
			if result.UnionReads {
				if err := s.validateResearchAdmissionGroups(ctx, records, requests); err != nil {
					return err
				}
			} else {
				for i := range group {
					if err := s.ValidateResearchAdmissionBatch(ctx, records[i], requests[i]); err != nil {
						return err
					}
				}
			}
			if source != nil {
				return researchPersistSourceGroup(ctx, source, records, &phase, result)
			}
			if durable != nil && (result.Mode == "group4transactions" || result.Mode == "group4reads" || result.Mode == "group4postdiscard" || result.Mode == "group4postverify") {
				return researchPersistGroupBatch(ctx, durable, records, &phase, result)
			}
			if durable != nil {
				for _, batch := range records {
					for _, record := range batch {
						admitStart := time.Now()
						p, retry, err := durable.AdmitBound(ctx, record.Prediction.ID, record.Prediction.Features, record.Outer[0], record.At, *record.Binding)
						phase.DurableAdmitNS += time.Since(admitStart).Nanoseconds()
						if err != nil {
							return err
						}
						result.Admits++
						if retry || !reflect.DeepEqual(p, record.Prediction) {
							return errors.New("durable original prediction mismatch")
						}
						verifyStart := time.Now()
						saved, err := durable.Admission(ctx, p.ID)
						if err != nil {
							return err
						}
						if !reflect.DeepEqual(saved, record) {
							return errors.New("durable original record mismatch")
						}
						phase.DurableVerifyNS += time.Since(verifyStart).Nanoseconds()
						discardStart := time.Now()
						_, err = durable.Discard(ctx, p.ID, record.At)
						phase.DurableDiscardNS += time.Since(discardStart).Nanoseconds()
						if err != nil {
							return err
						}
						result.Discards++
					}
				}
			}
			return nil
		})
		cancel()
		guardNS := time.Since(start).Nanoseconds()
		if err == nil && (result.Mode == "group4postdiscard" || result.Mode == "group4postverify") {
			postStart := time.Now()
			if source != nil {
				err = researchFinishSourceGroup(ctx, source, records, &phase, result)
			} else {
				err = researchFinishGroup(ctx, durable, records, &phase, result)
			}
			phase.PostGuardNS = time.Since(postStart).Nanoseconds()
			// Post-admission failures are fatal, not expired/stale admission attempts.
			if err != nil {
				fail(fmt.Errorf("post-guard verification/cleanup failed: %w", err))
				return
			}
		}
		phase.TotalNS = time.Since(start).Nanoseconds()
		phase.Accepted = err == nil
		result.Phases = append(result.Phases, phase)
		result.GuardNS = append(result.GuardNS, guardNS)
		for _, id := range previewIDs {
			a.Discard(id)
		}
		switch {
		case err == nil:
			result.Accepted += uint64(len(group))
			result.Validated += uint64(50 * len(group))
			for _, in := range group {
				result.AgeNS = append(result.AgeNS, time.Since(in.queuedAt).Nanoseconds())
			}
		case errors.Is(err, context.DeadlineExceeded) && ctx.Err() == nil:
			result.WaitExpired += uint64(len(group))
		case strings.Contains(err.Error(), "stale or quarantined"):
			result.Stale += uint64(len(group))
		default:
			fail(err)
			return
		}
	}
}

func researchPersistGroupBatch(ctx context.Context, d *researchmemory.Durable, groups [][]researchmemory.RecordedPrediction, phase *researchGuardPhase, result *researchGuardLoadResult) error {
	var requests []researchmemory.AdmissionRequest
	var originals []researchmemory.RecordedPrediction
	var discards []researchmemory.DiscardRequest
	for _, group := range groups {
		for _, r := range group {
			requests = append(requests, researchmemory.AdmissionRequest{ID: r.Prediction.ID, Features: r.Prediction.Features, Baseline: r.Outer[0], At: r.At, Binding: r.Binding})
			originals = append(originals, r)
			discards = append(discards, researchmemory.DiscardRequest{ID: r.Prediction.ID, Available: r.At})
		}
	}
	start := time.Now()
	admit := d.AdmitBatch
	discard := d.DiscardBatch
	if result.Mode == "group4reads" || result.Mode == "group4postdiscard" || result.Mode == "group4postverify" {
		admit = d.AdmitBatchWithSnapshotReads
		discard = d.DiscardBatchWithSnapshotReads
	}
	admitted, err := admit(ctx, requests)
	phase.DurableAdmitNS += time.Since(start).Nanoseconds()
	if err != nil {
		return err
	}
	result.Admits += uint64(len(admitted))
	if len(admitted) != len(originals) {
		return errors.New("batch admission count mismatch")
	}
	if result.Mode == "group4postverify" {
		// Verify the actual owned-worker return while still guarded. Only the
		// immutable ledger reread moves; these previews are cold-fixture-only.
		for i, r := range admitted {
			if r.Retry || !reflect.DeepEqual(r.Record, originals[i]) {
				return errors.New("batch original prediction mismatch")
			}
		}
		return nil
	}
	start = time.Now()
	var readback []researchmemory.RecordedPrediction
	if result.Mode == "group4reads" || result.Mode == "group4postdiscard" {
		ids := make([]uint64, len(admitted))
		for i, r := range admitted {
			ids[i] = r.Record.Prediction.ID
		}
		readback, err = d.Admissions(ctx, ids)
		if err != nil {
			return err
		}
		if len(readback) != len(admitted) {
			return errors.New("batch readback count mismatch")
		}
	}
	for i, r := range admitted {
		if r.Retry || !reflect.DeepEqual(r.Record, originals[i]) {
			return errors.New("batch original prediction mismatch")
		}
		var saved researchmemory.RecordedPrediction
		var err error
		if readback != nil {
			saved = readback[i]
		} else {
			saved, err = d.Admission(ctx, r.Record.Prediction.ID)
		}
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(saved, originals[i]) {
			return errors.New("batch stored original mismatch")
		}
	}
	phase.DurableVerifyNS += time.Since(start).Nanoseconds()
	if result.Mode == "group4postdiscard" {
		return nil
	}
	start = time.Now()
	retries, err := discard(ctx, discards)
	phase.DurableDiscardNS += time.Since(start).Nanoseconds()
	if err != nil {
		return err
	}
	result.Discards += uint64(len(retries))
	if len(retries) != len(discards) {
		return errors.New("batch discard count mismatch")
	}
	for _, retry := range retries {
		if retry {
			return errors.New("unexpected batch discard retry")
		}
	}
	return nil
}

// Only explicit unlabeled cleanup is allowed here. No admission, feedback or
// fitting authority is carried outside the service guard. Completion waits for
// the durable terminal; failures leave recovery to the owner/replay contract.
func researchDiscardGroup(ctx context.Context, d *researchmemory.Durable, groups [][]researchmemory.RecordedPrediction, phase *researchGuardPhase, result *researchGuardLoadResult) error {
	var requests []researchmemory.DiscardRequest
	for _, group := range groups {
		for _, r := range group {
			requests = append(requests, researchmemory.DiscardRequest{ID: r.Prediction.ID, Available: r.At})
		}
	}
	start := time.Now()
	retries, err := d.DiscardBatchWithSnapshotReads(ctx, requests)
	phase.DurableDiscardNS += time.Since(start).Nanoseconds()
	if err != nil {
		return err
	}
	if len(retries) != len(requests) {
		return errors.New("post-guard discard count mismatch")
	}
	for _, retry := range retries {
		if retry {
			return errors.New("unexpected post-guard discard retry")
		}
	}
	result.Discards += uint64(len(retries))
	return nil
}

// Ledger identity/integrity is checked without asserting fresh service validity.
// Any read error or mismatch prevents terminal cleanup and completion counting.
func researchFinishGroup(ctx context.Context, d *researchmemory.Durable, groups [][]researchmemory.RecordedPrediction, phase *researchGuardPhase, result *researchGuardLoadResult) error {
	if result.Mode == "group4postverify" {
		start := time.Now()
		var ids []uint64
		var originals []researchmemory.RecordedPrediction
		for _, group := range groups {
			for _, r := range group {
				ids = append(ids, r.Prediction.ID)
				originals = append(originals, r)
			}
		}
		saved, err := d.Admissions(ctx, ids)
		if err == nil && len(saved) != len(originals) {
			err = errors.New("post-guard readback count mismatch")
		}
		if err == nil {
			for i, r := range saved {
				if !reflect.DeepEqual(r, originals[i]) {
					err = errors.New("post-guard original mismatch")
					break
				}
			}
		}
		phase.DurableVerifyNS += time.Since(start).Nanoseconds()
		if err != nil {
			return err
		}
	}
	return researchDiscardGroup(ctx, d, groups, phase, result)
}
