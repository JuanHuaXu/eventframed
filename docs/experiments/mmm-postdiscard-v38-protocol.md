# Post-guard discard v38

Frozen before execution. Twelve rotated cells, three trials each: off,
non-durable group4, snapshot-read durable group4 with terminals inside the
guard, and the same path with explicit unlabeled discard after guard release.
Keep 192 recalls/four readers, 96 future writes spaced 2ms, 50 candidates,
four-observation ready groups, 64-slot handoff, 20ms entry deadline and fresh
persistent stores. No labels, fitting or production configuration changes.

Admission plus complete original readback stay inside the service guard.
Post-guard cleanup is synchronous in the same consumer: no next group can
increase pending state until terminal completion. Count accepted observations
and their age only after FULL durable discard succeeds. Record post-guard time
separately; guard time excludes it, total/age include it. Discard failures are
fatal with incomplete counters, never reclassified as an expired admission.
No assertion of historical admission or feedback authority follows from cleanup.

Before the full comparison, run small load accounting and publication/reopen
tests under race, plus existing batch commit-error/panic/process-exit tests.
Preserve exact originals, stale-admission rejection, typed terminal finality,
pending caps and absence of model/evidence-clock updates. Use exclusive JSONL
with raw timings, group-weighted counters and selected source hashes.

Retain 250ms accepted-age p95 and read-p99/off <= 1.10 screens; report write
tails, drops and expiries independently. This may transfer costs instead of
solving throughput: unchanged completion age or increased drops falsifies a
general latency rescue even if writer latency improves. No warm loaded learning
or real-agent claim can be established by this cold storage fixture.
