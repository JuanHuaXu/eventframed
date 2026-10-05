package researchindex

import (
	"context"
	"slices"
	"sync"
	"time"
)

type PrivateBatchResult struct {
	Timing    PrivateWriteTiming
	BatchID   uint64
	BatchSize int
}
type privateBatchReply struct {
	result PrivateBatchResult
	err    error
}
type privateBatchJob struct {
	ctx     context.Context
	request PrivateInsertRequest
	queued  time.Time
	reply   chan privateBatchReply
}

// PrivateBatchCollector owns at most capacity outstanding calls, including the
// active batch. One worker coalesces for at most window and commits at most four.
// Submit waits for the resolved callback outcome even after cancellation; late
// completion is observable, never disguised as a prompt uncommitted failure.
type PrivateBatchCollector struct {
	mu     sync.Mutex
	closed bool
	queue  chan *privateBatchJob
	slots  chan struct{}
	done   chan struct{}
	writer *PrivateGraphWriter
	window time.Duration
}

func NewPrivateBatchCollector(w *PrivateGraphWriter, capacity int, window time.Duration) (*PrivateBatchCollector, error) {
	if w == nil || capacity < 1 || capacity > 64 || window < 0 || window > 50*time.Millisecond {
		return nil, ErrCapacity
	}
	c := &PrivateBatchCollector{queue: make(chan *privateBatchJob, capacity), slots: make(chan struct{}, capacity), done: make(chan struct{}), writer: w, window: window}
	go c.run()
	return c, nil
}
func (c *PrivateBatchCollector) Submit(ctx context.Context, r PrivateInsertRequest) (PrivateBatchResult, error) {
	if e := ctx.Err(); e != nil {
		return PrivateBatchResult{}, e
	}
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return PrivateBatchResult{}, ErrFinished
	}
	select {
	case c.slots <- struct{}{}:
	default:
		c.mu.Unlock()
		return PrivateBatchResult{}, ErrCapacity
	}
	// Bounded dimensions prevent allocating attacker-sized vectors before the
	// writer's normal validation. Inputs must not be concurrently mutated here.
	if len(r.Vector) != c.writer.dimension {
		<-c.slots
		c.mu.Unlock()
		return PrivateBatchResult{}, ErrCapacity
	}
	r.Vector = slices.Clone(r.Vector)
	job := &privateBatchJob{ctx: ctx, request: r, queued: time.Now(), reply: make(chan privateBatchReply, 1)}
	c.queue <- job
	c.mu.Unlock()
	reply := <-job.reply
	return reply.result, reply.err
}
func (c *PrivateBatchCollector) finish(j *privateBatchJob, r PrivateBatchResult, e error) {
	<-c.slots
	j.reply <- privateBatchReply{r, e}
}
func (c *PrivateBatchCollector) run() {
	defer close(c.done)
	var batchID uint64
	for first := range c.queue {
		jobs := []*privateBatchJob{first}
		deadline := time.Now().Add(c.window)
		if d, ok := first.ctx.Deadline(); ok && d.Before(deadline) {
			deadline = d
		}
		timer := time.NewTimer(max(0, time.Until(deadline)))
		collecting := true
		for collecting && len(jobs) < 4 {
			select {
			case j, ok := <-c.queue:
				if !ok {
					collecting = false
					break
				}
				jobs = append(jobs, j)
				if d, ok := j.ctx.Deadline(); ok && d.Before(deadline) {
					deadline = d
					timer.Reset(max(0, time.Until(deadline)))
				}
			case <-timer.C:
				collecting = false
			}
		}
		timer.Stop()
		active := jobs[:0]
		for _, j := range jobs {
			if e := j.ctx.Err(); e != nil {
				c.finish(j, PrivateBatchResult{}, e)
			} else {
				active = append(active, j)
			}
		}
		if len(active) == 0 {
			continue
		}
		// Earliest deadline and explicit cancellation of ANY participant govern the
		// atomic batch. Once preparation begins, members cannot be silently dropped.
		var earliest time.Time
		for _, j := range active {
			if d, ok := j.ctx.Deadline(); ok && (earliest.IsZero() || d.Before(earliest)) {
				earliest = d
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		if !earliest.IsZero() {
			cancel()
			ctx, cancel = context.WithDeadline(context.Background(), earliest)
		}
		stops := make([]func() bool, 0, len(active))
		requests := make([]PrivateInsertRequest, len(active))
		waits := make([]time.Duration, len(active))
		for i, j := range active {
			stops = append(stops, context.AfterFunc(j.ctx, cancel))
			requests[i] = j.request
			waits[i] = time.Since(j.queued)
		}
		batchID++
		timing, e := c.writer.InsertBatch(ctx, requests)
		for _, stop := range stops {
			stop()
		}
		cancel()
		for i, j := range active {
			r := PrivateBatchResult{Timing: timing, BatchID: batchID, BatchSize: len(active)}
			r.Timing.Wait += waits[i]
			c.finish(j, r, e)
		}
	}
}

// Close rejects new calls, drains admitted jobs, and joins the worker. A timed
// out Close can be retried; it never abandons a callback. Close writer afterward.
func (c *PrivateBatchCollector) Close(ctx context.Context) error {
	c.mu.Lock()
	if !c.closed {
		c.closed = true
		close(c.queue)
	}
	c.mu.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return nil
	}
}
