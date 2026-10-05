# Bound worker v12: one held source/as-of proof

Status: frozen before implementation, 2026-10-01. Research-only; production
unchanged. Motivation: [v11 failure](mmm-bound-worker-load-v11-results.md).

The v11 writer arm failed because `ResearchEventUnchanged` uses fail-fast
`TryAcquire` outside the later, deadline-bound as-of admission guard. A retry
of only that check would still leave a check-to-commit interval. The v12
invariant is: **one acquired writer permit must cover as-of compatibility,
source-lineage continuity, journal/feature/witness validation, and durable
admission or feedback.** The permit cannot be reentered. Require a caller
deadline; cancellation before acquisition must not invoke the callback.
Source deletion or backfill must reject. A future-only writer may delay the
operation but cannot silently weaken its source proof. Existing fail-fast
APIs and legacy bound-worker behavior remain unchanged.

Tests: callback exclusion under a held writer, deadline expiry, future-only
wait and successful admission/feedback, source mutation rejection, focused
race checks, and full affected-package suites. Then run a separately frozen
finite rate matrix: one, two, four, and eight milliseconds between 192
Recall offers, four probe workers, 64 guarded labels, and 256 future-only
writes, three paired quiet/writer trials per rate. This matrix is motivated
by v11's observed queue saturation and is not untouched confirmation.
Record offer-to-return, call, queue, label age, guard failures, and complete
replay. At the predeclared 4 ms and 8 ms rates, writer-arm offer p99 must be
below 100 ms and label-age p99 below 250 ms with all labels complete. The
1 ms and 2 ms cells diagnose the capacity boundary; failures remain visible.

This does not change the v11 failure or establish population tails, hardware
power-loss recovery, or agent-answer improvement.
