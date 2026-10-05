# Completion notification v7: finite screen passed

Research worker WaitProcessed waits on an absolute completed+failed counter.
The counter check and lazy broadcast registration share the completion lock,
preventing missed wakeups. Cancellation stops only waiting; it does not retract
accepted labels. Close wakes unfinished waiters. Callers still inspect failures.
Polling-only users allocate no notification channels. No fitting/journal or
forecast-publication semantics were intentionally changed.

Three full researchmemory race runs passed, including existing synchronous
parity and new overlapping waiters, completed-before-wait, cancellation, close
and failed-counter tests. Package vet passed. This is unconfigured research code.

The18-arm load experiment PASSED every frozen criterion. Raw source and data:
mmm-notified-v7.jsonl. Independent audit verified source/current hashes, unique
arms, request/write measurements, queue/label conservation and all gates.

For192 requests, queue16 admitted170/168/167 with age p95
98.679/96.660/97.331ms. Queue64 admitted192 in every trial with age p95
164.833/170.178/186.535ms. All serving p99 ratios remained<=1.10; no errors or
worker failures occurred, writes overlapped reads, and all admitted labels
completed. Short queue16 admitted64/64/61; short queue64 admitted all64.

This separate-run pass is provisional. Next compare unchanged polling and
notification under one rotated workload, then replicate without retuning. Old
artifacts retain their embedded pre-notification worker source: current worker
hashes no longer match them, and their historical outcomes are not rewritten.
Repeated fixtures establish load behavior only. Mixed mutations, durable
recovery, realistic workload robustness and real-task accuracy remain open.
