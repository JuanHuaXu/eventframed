# Background fuzz queue mechanism check

Date: 2026-08-31

## Scope

This check validates the control flow of the background fuzz incubation queue. It
does not measure proposal usefulness or establish a production latency bound.

## Frozen behavior under test

- A successful low-certainty recall may nominate one bounded fuzz job.
- A high-certainty recall does not nominate a job.
- Nomination is nonblocking and a full queue drops the new nomination.
- Equivalent candidate sets are deduplicated during the cooldown.
- One worker executes only while no recall is active.
- A job whose snapshot is stale is discarded without retry.
- Queue status contains aggregate counters and no query text or event IDs.
- Fuzz output is proposal-only and cannot rewrite episodic memory or graph state.

## Commands and results

```text
go test -run 'TestLowCertaintyRecallEnqueuesAndExecutesBackgroundFuzz|TestBackgroundFuzzDropsStaleSnapshotWithoutRetry|TestBackgroundFuzzQueueAppliesNonblockingBackpressure|TestCertainRecallDoesNotEnqueueBackgroundFuzz' -v ./internal/service
PASS

go test -race ./...
PASS

go vet ./...
PASS

go test -count=20 ./internal/fuzzing ./internal/service ./internal/api ./internal/config
PASS

make build
PASS
```

The focused service test completed in 0.368 seconds on the local development
host. That elapsed time is a test-run duration, not per-recall latency.

## Audit fixes made before confirmation

1. Candidate-set fingerprints were canonicalized so tied retrieval order cannot
   bypass cooldown deduplication.
2. Synthetic replacement fields inherit source confidence instead of receiving
   certainty 1 by construction.
3. Public worker errors are reduced to fixed categories so status cannot expose
   internal identifiers.
4. The default minimum-trial count is clamped to the configured perturbation
   cap.

## Remaining evidence gap

A mixed-load confirmation must compare disabled, nomination-only, and active
worker configurations and report recall p50/p95/p99, enqueue time, queue delay,
drop/dedup/stale rates, trigger prevalence, and externally reviewed proposal
yield. A randomized or exhaustive audit stream is also required before making a
population claim about cases excluded by the certainty trigger.
