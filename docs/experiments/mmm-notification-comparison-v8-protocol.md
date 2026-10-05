# Polling versus notification v8

Frozen before execution:36 arms, queue16 and64, six trials each, rotating
off/notification/polling. All use the current worker; enabled modes use the same
general adapter. Reuse frozen workload functions for192 recalls,96 future writes,
four readers and50 fixture labels per frontier. Only completion waiting differs.

Retain gates: no errors, overlap, >=154 admissions, every label complete, no
worker failures, age p95<=250ms and serving p99<=1.10 paired off. Both modes are
evaluated, and a control failure remains recorded even if notification passes.
Exclusive-create raw JSONL includes sources/hashes. Report paired differences;
six trials do not establish population tail guarantees. No tuning after results.

This includes repeat notification trials without source changes since v7, but is
not independent real-task evidence or mixed-mutation/recovery validation.
