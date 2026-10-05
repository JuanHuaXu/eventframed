package researchpublicationstore

import (
	"errors"
	"sync"
	"time"
)

// GateSpan contains durations only, never event, tenant, query or record data.
type GateSpan struct {
	Kind           string
	WaitNS, HeldNS int64
	Entered        bool
}

// GateRecorder is bounded opt-in research instrumentation. It is not installed
// by normal constructors. Saturation is explicit instead of growing memory.
type GateRecorder struct {
	mu      sync.Mutex
	limit   int
	spans   []GateSpan
	dropped uint64
}

func NewGateRecorder(limit int) (*GateRecorder, error) {
	if limit <= 0 {
		return nil, errors.New("positive gate trace capacity required")
	}
	return &GateRecorder{limit: limit}, nil
}

func (r *GateRecorder) Snapshot() ([]GateSpan, uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]GateSpan(nil), r.spans...), r.dropped
}

func (r *GateRecorder) record(span GateSpan) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.spans) == r.limit {
		r.dropped++
		return
	}
	r.spans = append(r.spans, span)
}

func (s *Store) timingStart() time.Time {
	if s.timing == nil {
		return time.Time{}
	}
	return time.Now()
}

// Release before taking the recorder mutex: observation must not extend the
// protected callback or introduce recorder -> publication lock ordering.
func (s *Store) finishGate(kind string, start, entered time.Time) {
	if s.timing == nil {
		s.writer.Release(1)
		return
	}
	held := time.Since(entered)
	s.writer.Release(1)
	s.timing.record(GateSpan{Kind: kind, WaitNS: entered.Sub(start).Nanoseconds(), HeldNS: held.Nanoseconds(), Entered: true})
}

func (s *Store) failedGate(kind string, start time.Time) {
	if s.timing != nil {
		s.timing.record(GateSpan{Kind: kind, WaitNS: time.Since(start).Nanoseconds()})
	}
}
