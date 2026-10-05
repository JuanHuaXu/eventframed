# Query-burst scheduler contract results

2026-09-14. Research-only; no daemon serving behavior changed.

Command: `go test -race ./internal/observationlearners -run '^TestQueryBurst' -count=1 -v`

Both TestQueryBurstBudgetContracts and TestQueryBurstTimingAndExpiry passed.
Go package time1.474s, not a performance benchmark.

The tests cover256 periodic trigger patterns times3 pool-availability patterns
(always, intermittently, never), each over256 clocks. Every block respects its
cap; with always-available pools every schedule spends exactly31 labels.
Never-available pools spend zero. Explicit controls verify quiet fallback clocks,
early shocked spending, trigger expiry, and no carryover of unspent block tokens.

The frozen protocol is mmm-query-burst-protocol.md. The current implementation
does NOT yet calculate surprises, nominate origins, deliver labels, refit experts,
or collect Brier scores. Those integration boundaries need their own checks.
Race success does not imply concurrency support for this single-owner scheduler.
No efficacy, empirical latency improvement, or research-goal completion claimed.

## Learner integration

The repaired race run in mmm-query-burst-integration-repaired-run.txt passed:
- Original twelve-fixture natural replay and acquisition/as-of contracts43.82s.
- New integration48.90s: stationary case0 and shifted case20, each under both
  delivery schedules, periodic/burst timings and three selectors.
- Twelve poisoned as-of prefixes preserved forecasts and query decisions.
- Query count31 for each tested delayed paid arm; zero under complete feedback.
- Delivery deduplication, issued-forecast surprise anchoring, eligible fit labels,
  query bounds and input ownership checked.
- Scheduler contract tests passed again; package time94.200s.

The initial failure is retained in mmm-query-burst-integration-run.txt. A32-origin
pool can pay for evidence whose mixer journal entry was already censored. The
repair keeps that label for training/observation but does not reopen the expired
entry; each arrival records Mixer status. Both new timing arms share this rule.
The frozen protocol was clarified before any quality collection.

Full seven-arm collector TestQueryBurstCollect is implemented and compiled,
but NOT run. It records all2688 trajectories and does not hide unequal budgets.
Next implement independent scoring/audit, then collect and compare quality.
