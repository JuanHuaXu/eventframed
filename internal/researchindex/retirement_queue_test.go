package researchindex

import (
	"errors"
	"testing"
)

func TestRetirementQueueBoundsAndJoins(t *testing.T) {
	q := newRetirementQueue(1)
	entered, release := make(chan struct{}), make(chan struct{})
	failure := errors.New("cleanup failed")
	if err := q.submit(func() error { close(entered); <-release; return failure }, nil); err != nil {
		t.Fatal(err)
	}
	<-entered
	if err := q.submit(func() error { return nil }, nil); !errors.Is(err, ErrServingBusy) {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- q.close() }()
	close(release)
	if err := <-done; !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if err := q.close(); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if err := q.submit(func() error { return nil }, nil); !errors.Is(err, ErrServingClosed) {
		t.Fatal(err)
	}
}

func TestRetirementCompletionCanReuseQueueSlot(t *testing.T) {
	q := newRetirementQueue(1)
	submitted := make(chan error, 1)
	ran := make(chan struct{})
	if err := q.submit(func() error { return nil }, func(error) { submitted <- q.submit(func() error { close(ran); return nil }, nil) }); err != nil {
		t.Fatal(err)
	}
	if err := <-submitted; err != nil {
		t.Fatal("completion saw stale capacity", err)
	}
	<-ran
	if err := q.close(); err != nil {
		t.Fatal(err)
	}
}
