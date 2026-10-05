# Concurrent ingestion: genuine regression plus measurement correction

## Confirmed result

512 timed reads and256 writes completed. All writes succeeded. The first raw
checker reports232 errors, but this combines two different conditions:

- 215 are a **bad harness assertion**, not failed Recall calls. Nominated counts
  the overfetched pre-truncation journal frontier. Visible ingestion legitimately
  increases it from50 to66 or200 to216. The service truncates later for packing.
  The read-only test happened to have onlyN total records and hid this distinction.
- 17 are **genuine service failures**, all experimental in-window calls, caused
  by exhausting stale-journal retries. No ordinary-control call failed.

The raw artifact and erroneous harness remain unchanged for audit. The separate
classification script verifies that the215 cardinality-mismatch cases actually
returned packets with matching journals, no future evidence, and nominated counts
within the observed corpus bound. They must not be presented as service failures
or silently removed without explaining the incorrect assertion.

## By condition

Future-available ingestion:256/256 reads succeeded, no stale journal rejections.
For200 candidates, experimental maximum was30.089ms across both repetitions.
All successfully returned full-frontier records were checked against request as-of;
none were future-available. Some50-item arms finish before all writes, but each
arm recorded positive writer/read overlap;200-item arms overlap all16 writes.

In-window ingestion: ordinary128/128 reads succeeded. Experimental111/128
succeeded (17 failures). The experimental50-item arms had4 and5 failures;
200-item arms had4 each. Experimental200-item p95 was157.222/157.620ms and
maximum161.838ms, versus ordinary maxima44.073/46.544ms. Thus the encouraging
read-only timings do NOT extend to this visible-write pattern.

The journal correctly refuses stale publication; failed requests return empty
packets. This is availability/tail-latency degradation, not observed acceptance
of an invalid journal. All495 successful returns passed future-evidence checks.

## Mechanism and next rescue

The experimental hook performs task planning, lexical ranking, diversity packing
and explanation encoding BEFORE the optimistic journal commit. The normal path
does expensive packing after the commit. Longer pre-commit work expands exposure
to concurrent visible writes; the retry loop repeats that work up to five times.
Future ingestion has a separately accounted compatibility exception and did not
trigger this problem. Do not extend that exception to visible writes merely to
make these tests pass.

First correct the next harness variant to distinguish nomination count from the
post-truncation candidate cap. Then test an explicit request-level read lease
against ingestion, measuring BOTH reader admission wait and writer wait. This
would move contention rather than erase it; its costs and cancellation behavior
must be reported. A scheduling prototype is not a production fix. An immutable
snapshot/MVCC alternative remains a larger engineering lead, with publication
semantics requiring explicit agreement rather than weakened freshness checks.

## Audit

Protocol TASK_LEXICAL_CHURN_PROTOCOL.md preceded dispatch. Raw observations and
source hashes: task-lexical-churn-results.json. The original checker preserves
the raw count; the classification script reports17 genuine failures,215 harness
misclassifications and495 successful returns.

```sh
node research/public-task-pilot/check-task-lexical-churn.mjs
node research/public-task-pilot/classify-task-lexical-churn.mjs
```

The fixture is closed-loop, memory-store, hash-embedded and contains repeated
public facts. It does not test durable persistence or semantic-policy updates.
No production code or earlier artifacts were modified. All seven whole goals
remain open; direction6 has a newly demonstrated failure requiring rescue.
