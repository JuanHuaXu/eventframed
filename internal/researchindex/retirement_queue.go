package researchindex

import (
	"errors"
	"sync"
)

// retirementQueue schedules cleanup only. Its bound includes the running job;
// the serving owner separately retains failed resources in its retirement count.
type retirementQueue struct {
	mu              sync.Mutex
	jobs            chan retirementJob
	done            chan struct{}
	count, capacity int
	closed          bool
	err             error
}
type retirementJob struct {
	run  func() error
	done func(error)
}

func newRetirementQueue(capacity int) *retirementQueue {
	q := &retirementQueue{jobs: make(chan retirementJob, capacity), done: make(chan struct{}), capacity: capacity}
	go func() {
		defer close(q.done)
		for job := range q.jobs {
			err := job.run()
			q.mu.Lock()
			q.err = errors.Join(q.err, err)
			q.count--
			q.mu.Unlock()
			// Release the queue slot before making the owner's retirement slot
			// reusable; reversing this order creates a spurious queue-full race.
			if job.done != nil {
				job.done(err)
			}
		}
	}()
	return q
}
func (q *retirementQueue) submit(job func() error, done func(error)) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return ErrServingClosed
	}
	if job == nil || q.count >= q.capacity {
		return ErrServingBusy
	}
	q.count++
	q.jobs <- retirementJob{run: job, done: done}
	return nil
}

// close stops admission and joins every admitted cleanup. There is deliberately
// no timeout that would report success while a cleanup still owns resources.
func (q *retirementQueue) close() error {
	q.mu.Lock()
	if !q.closed {
		q.closed = true
		close(q.jobs)
	}
	q.mu.Unlock()
	<-q.done
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.err
}
