# Published-LSN full Recall load v9: frozen private protocol

Date: 2026-10-02. Test-only adapter; production remains unchanged.
V8 established deterministic event-only Search/Snapshot/journal ordering.
V7 established loaded backend reads without Service Recall or frontier
journal writes. This test combines those costs and consistency rules.

## Fixture and ownership

Use the private 256D declared-payload, normalized, journaled publication
gate. Create the service and bind its Bayesian policy while the gate is
PENDING; publish READY afterward. The adapter owns every subsequent
event batch, frontier-journal commit, and `(LSN,snapshot,published_at)`
pointer update under one lock. Each Recall gets a private context pin.
Search decodes full EventFrames from its exact LSN; Snapshot returns the
pin. Before routing a journal into the gate, reject stale snapshots under
the owner lock without poisoning an unchanged gate. No bypass writer,
graph mutation, posterior update, or external retriever participates.

Seed two aligned eligible genesis rows, 198 eligible dense rows at
`theta_i=0.005(i+1)`, one aligned future sentinel, and 16 future dense
rows at `theta_i=0.001(i+1)`. Query as-of is 120 ms after
2026-10-02T00:00:00Z; future availability is 125 ms. Use an explicit
raw `2q` query vector from v6. Fix `RecallK=50`, `PackK=10`, overfetch
factor 3 and token budget 10,000. Thus the nominated frontier target
is 150 eligible rows.

## Offered load and oracle

Run two fresh trials. Offer 128 visible writes and 128 Recalls at
nominal 4 ms cadence, with four Recall workers and a single writer.
The writer uses groups of at most 16 and dwell at most 16 ms. New row i
has dense angle `0.00213(i+1)` and availability equal to as-of.
Every journal has a unique session ID. An acknowledgement follows the
durable batch marker and pointer publication. Per runtime version,
record an independent sorted eligible-ID/angle oracle. After the run,
compare each successful Recall's 150 nominated journal decisions to
that version's exact top-150 set; compare packet and durable journal
snapshots to the per-call pin; require packet candidates and decisions
to exclude future rows. Each successful packet must be at least as new
as the latest event-batch acknowledgement before its offer. Verify the
final gate is READY and the last published snapshot includes all 128
new events. Report retries/stale rejects, read/journal errors, write
acknowledgements, actual offer gaps, response call and offer-to-done
latencies, write offer-to-ack age, and published-view age.

## Frozen gates

For each normal trial: all 128 writes acknowledged, all 128 Recalls
succeed, zero oracle/snapshot/journal/future/freshness violations,
published-view maximum age <250 ms, write age p99 <250 ms, Recall
call p99 <100 ms and Recall offer-to-done p99 <100 ms. Use nearest-rank
percentiles. Do not adjust rates, workers, caps or gates after seeing
the result.

Run the same two trials under `-race` with
`EVENTFRAME_RACE_CORRECTNESS_ONLY=1`: counts and all semantic gates
still apply, while timing gates are reported but not enforced because
instrumented execution is a different throughput regime. This mode is
declared before the first run. Ordinary package tests and vet also run.
Even a pass is not evidence of sustained large-corpus performance,
changing semantic state isolation, label-to-forecast freshness,
OpenClaw latency, or Goal 6 completion.
