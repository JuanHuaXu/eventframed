# Recall store-phase profile v17: frozen diagnostic

Date: 2026-10-01. Goal 6 research-only diagnostic; no production changes.

Question: which full-service store operations grow under the writer load that
failed v14/v15, given that raw `Store.Search` alone passed v16? This profile
does not claim that the slowest observed operation is the sole root cause.

## Fixed workload

- Use the v15 persistent LibraVDB fixture and ordinary `Service.Recall`, with
  50 or 200 as-of-visible events, 8 ms offers, four probe workers, 192 offers
  per trial, three paired trials per quiet/writer cell, and 256 future-only
  writes per writer trial. No bound learner or feedback.
- Keep `RecallK` equal to the live event count and `PackK=10`. Validate every
  nominated event ID against the exact live set and reject any future packet
  candidate. Require all offers and writes to complete.
- Trace only the probe service's `EventStore` boundary: Search, Snapshot,
  predictive graph, selection and omitted certificates, Anti-Pigeon
  certificate, posterior, residual candidate reads, and Bayesian journal put.
  The writer uses the unwrapped underlying service.
- Record offer, call, and queue latency separately; per-operation call counts,
  individual durations, and per-Recall wall span. Residual reads may overlap,
  so do not sum their durations into an exclusive Recall breakdown.
- Record the number of Search and journal calls per Recall to detect retries.

## Interpretation

A phase is a candidate bottleneck only if its timing and count change with
the writer arm in a way that can explain the service latency. If named store
spans do not account for it, profile non-store candidate work next. Do not
replace v14's loaded latency gate with this diagnostic. No parameter tuning
or production patch is authorized by this screen alone.

Run with `EVENTFRAME_RUN_RECALL_PROFILE_V17=1 go test ./internal/service -run
'^TestResearchRecallStorePhaseProfileV17$' -count=1 -v`.
