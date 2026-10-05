# Already-pending event batching: durable arrival screen

Continue the32-arm atomic-durable screen with the same facts, arrival schedules,
deadlines, packing variants, read admission and reopen validation. New output:
batch-queue-results.json. Normal modules and atomic-overlay-v1, researchpriority
build tag. Production remains unchanged.

Change only timed event persistence: CaptureTurn still extracts5w1h, validates
and embeds. Its Put enters a64-pending queue with one worker, draining at most16
already-pending same-tenant events. No accumulation timer. Each transaction
obtains the original exclusive admission permit when the arm enables admission.
Reads retain their original shared permit. Seeds still use ordinary Put.

Snapshot publication uses the final committed snapshot for every batch member.
Context deadline is the earlier of100ms from batch construction and every active
member's deadline. Already-canceled jobs do not enter the transaction. Once
submitted, callers wait for settlement: manual cancellation of one member does
not abandon or cancel another's shared transaction. Count late settlements as
latency failures. This is an explicit semantic change, not transparent cancel
parity. Any callback error conservatively stops further queue commits until
reconciliation; this includes safe admission errors in this first prototype.
Record batch sizes including failed admission attempts. Store files retained.

Require no errors, stale rejections, >100ms responses, warmup failures or reopen
failures to pass an arm, exactly as before. Count all1024 reads and512 writes.
Report false starts and all failures; event batching alone is rejected as a
complete rescue if any high-rate arm fails. Do not reinterpret synchronous
journal commits, reduce workload or drop arrivals to pass. Cross-run comparison
with atomic-durable is descriptive, not randomized confirmation.

Before running: race-enabled queue tests ten repetitions, including settlement,
cancellation, capacity, shutdown, deadlines and unknown-outcome stop behavior.
Original store remains unchanged. No fault-injected durable commits yet, so no
claim of complete crash/partial-failure coverage follows.
