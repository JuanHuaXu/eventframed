# Feedback bridge temporal reuse v2

2026-09-12. Opt-in research extension, not enabled in production. Exact-snapshot
bridge behavior remains the default and is retained as the paired control.

## Change and invariant

The v1 bridge rejected every store-version change, including arrivals later than
all admitted queries. The new constructor `NewResearchTemporalFeedbackBridge`
uses the existing store `ResearchSnapshotCompatible` proof. The store must prove
that intervening changes are strictly future ingestion relative to the cutoff;
unknown history, missing optional interface and relevant changes fail closed.

The cutoff is the maximum as-of time across ALL admitted queries and the proposed
new query/score. Checking only the oldest pending label's query would be unsound:
a write could be future to that query but relevant to a later admitted one.
Both the bridge's anchor snapshot and the incoming journal snapshot must pass.
Journal tenant/time/member/baseline binding and bounded duplicate tracking remain
unchanged. Label availability time is not substituted for query as-of time.

When a formerly future arrival becomes visible at a later scoring time, the
bridge rejects that score. The extension does not automatically advance epochs,
retrain on a changed corpus, or bypass model feedback-time checks.

## Tests run

Memory and persistent LibraVDB-store cases exercise:

- Admit a committed frontier, append future-only data, then complete an explicit
  delayed worker update. A research score before that arrival becomes visible
  succeeds, while scoring after visibility rejects.
- A newer physical journal snapshot with unchanged as-of data can be admitted.
- A subsequent backfill invalidates feedback and research scoring.
- The paired default exact-snapshot bridge still rejects the same future write.
- An arrival between two admitted query times invalidates feedback for the older
  query as well; the learner cannot ignore its later admitted dependency.
- A wrapper without the optional compatibility interface fails closed.

Targeted race tests pass three repetitions. Broader research service/store
regressions are also run; these are integration invariants, not throughput tests.
The first fixture omitted mandatory positive recall/packing/token defaults and
failed construction. Only fixture configuration was corrected before rerunning;
the service validation contract was not relaxed.

## Not established

No accuracy rescue, persistent feedback ledger, full-request p99, or automatic
publication authority follows. The persistent store is exercised, but bridge
state and duplicate history are still in memory. Bounded history exhaustion can
still reject reuse. Externally verified labels remain required. Score-time checks
do not hold a database transaction through an asynchronous fit; no result is
connected to serving. Concurrent read/write load with actual learning remains the
next separate experiment for roadmap 6.
