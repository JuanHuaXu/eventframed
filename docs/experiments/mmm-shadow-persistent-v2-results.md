# Persistent shadow pilot: FAILED zero-error screen

Three off/on pairs, each256 reads and64 writes against a fresh temporary embedded
LibraVDB. All64 writes in each arm completed while readers were still active.
This is persistent internal service traffic, not remote ranking, HTTP, SQLite
rank-delta integration, OpenClaw or a real learner. Callback remains diagnostic.

| Trial | Off read p99 | On read p99 | Off read errors | On read errors | Completed shadow jobs |
| --- | --- | --- | --- | --- | --- |
| 0 | 80.529 ms | 52.496 ms | 1 | 2 | 162 |
| 1 | 68.639 ms | 48.528 ms | 0 | 0 | 159 |
| 2 | 69.018 ms | 61.400 ms | 0 | 0 | 177 |

All three errors were stale runtime snapshots at Bayesian frontier journal
commit, after the service's bounded retries. They occur with shadow OFF too;
the experiment does not establish that shadow caused them. Zero-error criterion
FAILED. No request errors were discarded or retrospectively allowed.

Enabled p99 met the screen in each pair, but the tiny sequential alternating
sample cannot establish a causal speedup. All values include failed requests.
Raw durations, errors and statuses are retained in mmm-shadow-persistent-v2.json.
Stale shadow jobs91/96/78 demonstrate real version churn. Many jobs completed,
but this is not proof useful learned discoveries arrive before they become stale.

## Accounting correction found during audit

Accepted jobs exceeded terminal counters by one in each enabled trial. Source
inspection confirmed a worker can dequeue an accepted job, notice cancellation
and return before accounting for it. Close cannot drain a removed job. Fixed
that branch and gave accepted-but-cancelled jobs a distinct Cancelled counter;
Dropped now denotes refused nominations rather than a mix of lifecycle stages.
Terminal invariant: Accepted=Completed+Stale+Failed+Cancelled after Close.

A cancellation-ready queue test exercises both shutdown choices across128
iterations and verifies conservation and idempotent Close. Targeted race tests
passed, along with existing cancellation/isolation/stale checks. Raw timing
data above predates this accounting-only correction and is not a performance
retest of it. No journal retry or persistence behavior was changed.

The timing test is opt-in using EVENTFRAME_RESEARCH_ARTIFACT so normal unit runs
do not become flaky performance gates; opted-in failures still fail the test and
write raw evidence. Next work: analyze snapshot retry starvation under sustained
writes without weakening version integrity. This is an unresolved serving gap,
not a reason to relabel the persistent screen successful.
