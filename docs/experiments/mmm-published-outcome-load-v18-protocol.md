# Loaded published-outcome v18: frozen research screen

Date: 2026-10-02. This follows the single-owner v17 component and
the eight-worker v16 no-feedback load result. The hypothesis is that
serializing outcome commits with admitted writes and read-to-journal
Recalls preserves both correctness and <100 ms offered Recall p99.
The falsifier is stale or misattributed feedback, future-data use,
unpublished acknowledged outcomes, or a latency/freshness miss.

Use two matched arms per trial: no feedback and 16 outcome deliveries.
Both arms use the same 256D 200-eligible/17-future corpus,
128 event writes and 128 full `Service.Recall` offers at 4 ms nominal
cadence, cap-16/16-ms event batches, eight Recall workers, native
cap-four/1-ms journal batching, `RecallK=50`, `PackK=10`, and
published-LSN Search. Use the v17 gated outcome adapter only in the
feedback arm. Feed outcome jobs from committed Recall journals,
one per distinct nominated event, at 16 ms nominal cadence after
the first committed journals are available. Each full-stream
outcome gets its real wall-clock `ObservedAt`/`AvailableAt` at
delivery; all forecasts are emitted before their outcomes.
Candidate events available only after the test horizon must never
be nominated or packed. The feedback arm may use synthetic, test-only
selection/omitted-influence certificates solely to exercise the
scored law; this is not empirical certification.

Predeclared pass criteria in each of two normal trials:

- 128/128 acknowledged event writes and successful Recalls in both
  arms; 16/16 new, durable, attributed outcomes in the feedback arm.
- No stale snapshot, future nomination, journal/packet mismatch,
  acknowledged-before-offer omission, direct-bypass or marker mismatch.
- Each acknowledged outcome is published before its response returns;
  any later offered Recall uses a snapshot at least as new. A later
  scored Recall must visibly use at least one updated posterior;
  an earlier as-of Recall must not use later feedback.
- Recall offer-to-done and occupied-call p99 <100 ms, event-write
  offer-to-ack p99 <250 ms, outcome offer-to-published p99 <100 ms,
  outcome maximum <250 ms, and published-view maximum age <250 ms.

Record offered gaps, journal batching, owner/admission waits, write,
Recall and outcome tails, posterior use count, stale attempts and
final durable READY. Run a correctness-only race replay; do not
reinterpret instrumented race timings as production latency.
Freeze this protocol before data collection. A pass remains a finite
synthetic single-owner component, not Goal 6 completion: external
writers, crash recovery under load, larger corpus and real-agent
outcomes still need separate evidence. Production remains untouched.
