package researchadmission

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
)

func TestQueuedCancellation(t *testing.T) {
	for _, writer := range []bool{false, true} {
		t.Run(map[bool]string{false: "read", true: "write"}[writer], func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				g, _ := New(2)
				release := make(chan struct{})
				holder := make(chan error, 1)
				go func() { holder <- g.Write(t.Context(), func(context.Context) error { <-release; return nil }) }()
				synctest.Wait()
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				done := make(chan error, 1)
				called := false
				admit := g.Read
				if writer {
					admit = g.Write
				}
				go func() { done <- admit(ctx, func(context.Context) error { called = true; return nil }) }()
				synctest.Wait()
				select {
				case err := <-done:
					t.Fatalf("did not queue: %v", err)
				default:
				}
				cancel()
				if err := <-done; !errors.Is(err, context.Canceled) {
					t.Fatalf("cancel: %v", err)
				}
				if called {
					t.Fatal("canceled waiter executed")
				}
				close(release)
				if err := <-holder; err != nil {
					t.Fatal(err)
				}
				if err := g.Write(t.Context(), func(context.Context) error { return nil }); err != nil {
					t.Fatal(err)
				}
			})
		})
	}
}

func TestReleaseOnErrorAndPanic(t *testing.T) {
	g, _ := New(2)
	sentinel := errors.New("callback failure")
	for _, admit := range []func(context.Context, func(context.Context) error) error{g.Read, g.Write} {
		if err := admit(t.Context(), func(context.Context) error { return sentinel }); err != sentinel {
			t.Fatal(err)
		}
		func() {
			defer func() {
				if recover() != sentinel {
					t.Error("panic not propagated")
				}
			}()
			_ = admit(t.Context(), func(context.Context) error { panic(sentinel) })
		}()
		// TryAcquire is a nonblocking leak check, not part of the public API.
		if !g.permits.TryAcquire(2) {
			t.Fatal("permits leaked")
		}
		g.permits.Release(2)
	}
}

func TestParallelReadsExcludeWriter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		g, _ := New(2)
		release := make(chan struct{})
		entered := make(chan struct{}, 2)
		done := make(chan error, 3)
		for i := 0; i < 2; i++ {
			go func() {
				done <- g.Read(t.Context(), func(context.Context) error { entered <- struct{}{}; <-release; return nil })
			}()
		}
		synctest.Wait()
		if len(entered) != 2 {
			t.Fatal("reads not concurrent")
		}
		writerEntered := make(chan struct{}, 1)
		go func() {
			done <- g.Write(t.Context(), func(context.Context) error { writerEntered <- struct{}{}; return nil })
		}()
		synctest.Wait()
		if len(writerEntered) != 0 {
			t.Fatal("writer overlapped readers")
		}
		close(release)
		for i := 0; i < 3; i++ {
			if err := <-done; err != nil {
				t.Fatal(err)
			}
		}
		if len(writerEntered) != 1 {
			t.Fatal("writer never ran")
		}
	})
}

func TestInvalidAndAlreadyCanceled(t *testing.T) {
	for _, capacity := range []int64{0, -1} {
		if _, err := New(capacity); err == nil {
			t.Fatal("invalid capacity accepted")
		}
	}
	g, _ := New(1)
	if err := g.Read(t.Context(), nil); err == nil {
		t.Fatal("nil callback accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	for _, admit := range []func(context.Context, func(context.Context) error) error{g.Read, g.Write} {
		if err := admit(ctx, func(context.Context) error { t.Fatal("canceled callback executed"); return nil }); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	}
}

func TestActiveCancellationDoesNotReleaseEarly(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		g, _ := New(2)
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		release := make(chan struct{})
		done := make(chan error, 1)
		go func() {
			done <- g.Read(ctx, func(context.Context) error {
				<-release
				return ctx.Err()
			})
		}()
		synctest.Wait()
		cancel()
		synctest.Wait()
		if g.permits.TryAcquire(2) {
			g.permits.Release(2)
			t.Fatal("cancellation released a still-running callback")
		}
		close(release)
		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if !g.permits.TryAcquire(2) {
			t.Fatal("finished callback leaked admission")
		}
		g.permits.Release(2)
	})
}

func TestSeparateGatesDoNotProtectSharedState(t *testing.T) {
	// Negative control: creating one gate per request/tenant cannot protect a
	// shared version domain. No data race is needed to demonstrate the overlap.
	synctest.Test(t, func(t *testing.T) {
		readerGate, _ := New(1)
		writerGate, _ := New(1)
		release := make(chan struct{})
		done := make(chan error, 1)
		go func() { done <- readerGate.Read(t.Context(), func(context.Context) error { <-release; return nil }) }()
		synctest.Wait()
		called := false
		if err := writerGate.Write(t.Context(), func(context.Context) error { called = true; return nil }); err != nil {
			t.Fatal(err)
		}
		if !called {
			t.Fatal("negative control did not demonstrate bypass")
		}
		close(release)
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	})
}
