package researchbatch

import (
	"context"
	"errors"
	"sync/atomic"
)

type notEnteredError struct{ cause error }

func (e *notEnteredError) Error() string { return "batch admission did not enter: " + e.cause.Error() }
func (e *notEnteredError) Unwrap() error { return e.cause }

// RunAdmitted records a local control-flow fact, not an error-text inference.
// admit must be synchronous: it cannot retain callback or invoke it after return.
// It must call callback at most once and propagate its error (as Gate.Write does).
func RunAdmitted(ctx context.Context, admit func(context.Context, func(context.Context) error) error, fn func(context.Context) error) error {
	var entered atomic.Bool
	err := admit(ctx, func(c context.Context) error {
		entered.Store(true)
		return fn(c)
	})
	if err != nil && !entered.Load() {
		return &notEnteredError{cause: err}
	}
	return err
}

// NeverEntered is only meaningful for errors from the trusted local adapter.
// Remote strings, context errors and backend transaction errors cannot prove it.
func NeverEntered(err error) bool {
	var e *notEnteredError
	return errors.As(err, &e)
}
