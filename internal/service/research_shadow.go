package service

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// ResearchShadowInput deliberately carries no text, identifiers or mutable
// slices. These are retrieval diagnostics, NOT the synthetic learner's fields.
type ResearchShadowInput struct {
	AsOf      time.Time
	Snapshot  model.Snapshot
	Certainty float64
	Count     int
	Scores    [64]float64
}
type ResearchShadowPolicy struct {
	TemporalReuse bool
	Enabled       bool
	Capacity      int
	MaxAge        time.Duration
	// Process must honor context cancellation. It has no service/store handle
	// supplied by this API and cannot publish a forecast into a served packet.
	Process func(context.Context, ResearchShadowInput) (float64, error)
}
type ResearchShadowStatus struct {
	AsOf                                                   time.Time
	Accepted, Dropped, Completed, Stale, Failed, Cancelled uint64
	Depth                                                  int
	Running, Closed, HasResult                             bool
	Value                                                  float64
	Snapshot                                               model.Snapshot
}
type researchShadowJob struct {
	input    ResearchShadowInput
	deadline time.Time
}
type researchShadowQueue struct {
	service *Service
	policy  ResearchShadowPolicy
	jobs    chan researchShadowJob
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{}
	mu      sync.Mutex
	status  ResearchShadowStatus
}

func (p ResearchShadowPolicy) validate() error {
	if p.Enabled && (p.Capacity < 1 || p.Capacity > 256 || p.MaxAge <= 0 || p.MaxAge > time.Minute || p.Process == nil) {
		return errors.New("invalid research shadow policy")
	}
	return nil
}
func newResearchShadow(s *Service, p ResearchShadowPolicy) *researchShadowQueue {
	ctx, cancel := context.WithCancel(context.Background())
	q := &researchShadowQueue{service: s, policy: p, jobs: make(chan researchShadowJob, p.Capacity), ctx: ctx, cancel: cancel, done: make(chan struct{})}
	go q.run()
	return q
}
func (s *Service) nominateResearchShadow(packet model.ContextPacket, asOf ...time.Time) {
	q := s.researchShadow
	if q == nil {
		return
	}
	in := ResearchShadowInput{Snapshot: packet.Snapshot, Certainty: packet.PacketAnswerCertainty, Count: min(64, len(packet.Candidates))}
	if len(asOf) > 0 {
		in.AsOf = asOf[0]
	}
	for i := 0; i < in.Count; i++ {
		in.Scores[i] = packet.Candidates[i].PredictiveScore
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.status.Closed {
		q.status.Dropped++
		return
	}
	select {
	case q.jobs <- researchShadowJob{in, time.Now().Add(q.policy.MaxAge)}:
		q.status.Accepted++
	default:
		q.status.Dropped++
	}
}
func runShadowProcess(ctx context.Context, p ResearchShadowPolicy, in ResearchShadowInput) (value float64, err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("shadow processor failed")
		}
	}()
	return p.Process(ctx, in)
}
func (q *researchShadowQueue) run() {
	defer close(q.done)
	for {
		select {
		case <-q.ctx.Done():
			return
		case job := <-q.jobs:
			if q.ctx.Err() != nil {
				// This job already left the channel, so Close cannot drain it.
				q.mu.Lock()
				q.status.Cancelled++
				q.mu.Unlock()
				return
			}
			q.mu.Lock()
			q.status.Running = true
			q.mu.Unlock()
			ctx, cancel := context.WithDeadline(q.ctx, job.deadline)
			stale := !q.compatible(ctx, job.input.Snapshot, job.input.AsOf)
			var value float64
			var err error
			if !stale {
				value, err = runShadowProcess(ctx, q.policy, job.input)
			}
			stale = stale || !q.compatible(ctx, job.input.Snapshot, job.input.AsOf)
			cancel()
			q.mu.Lock()
			q.status.Running = false
			if stale || q.status.Closed {
				q.status.Stale++
			} else if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
				q.status.Failed++
			} else {
				q.status.Completed++
				q.status.HasResult = true
				q.status.Value = value
				q.status.Snapshot = job.input.Snapshot
				q.status.AsOf = job.input.AsOf
			}
			q.mu.Unlock()
		}
	}
}

// Status is diagnostic and version stamped, never authorization for serving.
// The store may change after this read; callers must not treat it as a lease.
func (s *Service) ResearchShadowStatus() ResearchShadowStatus {
	q := s.researchShadow
	if q == nil {
		return ResearchShadowStatus{}
	}
	q.mu.Lock()
	v := q.status
	v.Depth = len(q.jobs)
	q.mu.Unlock()
	if v.HasResult && !q.compatible(context.Background(), v.Snapshot, v.AsOf) {
		v.HasResult = false
		v.Value = 0
	}
	return v
}

func (q *researchShadowQueue) compatible(ctx context.Context, snapshot model.Snapshot, asOf time.Time) bool {
	if ctx.Err() != nil {
		return false
	}
	if q.policy.TemporalReuse && !asOf.IsZero() {
		if db, ok := q.service.store.(interface {
			ResearchSnapshotCompatible(context.Context, model.Snapshot, time.Time) bool
		}); ok {
			return db.ResearchSnapshotCompatible(ctx, snapshot, asOf)
		}
	}
	return q.service.store.Snapshot(ctx) == snapshot
}
func (q *researchShadowQueue) Close() {
	q.mu.Lock()
	q.status.Closed = true
	q.status.HasResult = false
	q.status.Value = 0
	q.mu.Unlock()
	q.cancel()
	<-q.done
	q.mu.Lock()
	for len(q.jobs) > 0 {
		<-q.jobs
		q.status.Cancelled++
	}
	q.mu.Unlock()
}
