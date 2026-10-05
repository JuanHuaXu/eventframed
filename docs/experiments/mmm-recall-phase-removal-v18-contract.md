# Recall phase-removal v18: frozen diagnostic

Date: 2026-10-01. Goal 6 research-only; production remains unchanged.

Question: in the 200-event workload that fails the 8 ms offered Recall gate,
which of graph reads and journal persistence are necessary for queue growth?
The prior v17 store trace names several writer-sensitive operations but is
not causal attribution.

## Fixed factorial

- 200 as-of-visible events, `RecallK=200`, `PackK=10`, four probe workers,
  192 Recall offers at 8 ms intervals, and three trials for each arm. Pair
  each arm with quiet and 256 future-only writer operations. Rotate arm order
  by trial and reverse quiet/writer order on the middle trial.
- Native: unmodified persistent `Service.Recall`.
- Graph-cache diagnostic: return a preloaded graph from the same persistent
  store. Require graph version to remain unchanged during the trial.
- Journal-sink diagnostic: encode every journal and keep it in a synchronized
  in-memory map with exact conflict checking, but do not persist it. Verify
  all 192 distinct journals reached the sink. This arm has **no durable
  journal** and cannot satisfy normal serving/learning semantics.
- Combined diagnostic: both graph cache and in-memory journal sink.
- Other store operations, nomination, residual lookups, ranking, and future
  event writes remain unchanged. Validate the exact 200 nominated IDs, no
  packed future events, all completed offers, and all writer operations.
- Report offer, call, queue p99, plus traced Search, graph, and journal p99.
  Keep the native arm in this run. No after-the-fact gate adjustment.

## Interpretation

If neither single removal changes overloaded p99, a combined effect or another
phase remains. If either removes queueing, it shows a necessary contribution
under this synthetic workload, not a production fix. A durable redesign must
retain journal correctness and retry semantics and be re-tested on the full
service/learner gate. Marginal p99s are not additive.

Run with `EVENTFRAME_RUN_RECALL_REMOVAL_V18=1 go test ./internal/service -run
'^TestResearchRecallPhaseRemovalV18$' -count=1 -v`.
