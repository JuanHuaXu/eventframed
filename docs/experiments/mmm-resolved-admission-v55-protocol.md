# Resolved source admission v55

Frozen before measurement. v54 combined verified cleanup reached570/576 accepted
observations but failed the250ms age screen in two of three trials. Admission
still resolves canonical sources then rereads records by learner ID. Investigate
that duplicate preflight, not a claim that it explains all remaining latency.
Competing costs include canonical decoding, atomic append, service validation
and writer contention. No common-case read/write or guard checks are removed.

The optional constructor reuses only call-local, validated source results under
the private SourceOwner lock. Its private Durable and exclusive ledger ownership
exclude concurrent ID allocation. Asynchronous fitting can change publications,
but actual owned-worker Predict/Record still supplies each new original. There
is no public resolution hint, retained cache or preview promotion. Existing
constructors/APIs remain controls. AppendBatch still verifies exact identities,
retry bytes and acknowledgments atomically; uncertainty stops the owner until
replay. A wrong/missing resolution or unexpected conflicting append must not
produce an accepted wrong original. Source authority is not feedback authority.

Checks: matched-publication warm parity against point and batch source owners;
mixed new/retry, equivalent timestamp encodings, reopen, preflight conflicts,
capacity, cancellation, concurrent retries, before/after-commit error and panic,
and an injected unexpected write before the final atomic append. Count the
eliminated learner-ID preflight reads while retaining atomic append calls. Run
full ledger/learner/service race and vet checks before isolated measurement.

Four rotated arms, three trials: off; raw prepared durable/union/postverify;
batch source with combined verified cleanup; same with resolved admissions.
Keep192 Recall calls/four readers,96 future writes2ms apart,K50/pack10,50 shared
visible public events, fixed as-of,queue64,groups<=4,20ms entry budget,FULL
durability,parent callback context. No labels, fitting, production or private
data. Source/flag/hash snapshots and all raw timings saved in exclusive JSONL.

Unchanged screens: each resolved trial age p95<=250ms and read p99<=1.10x
same-trial off. Report all completions,drops,expiry,writer tails,original counts
and admission/cleanup phases. No unexpected errors or original mismatches.
Go PASS is accounting/integrity, not automatic latency success. No threshold
retuning. A finite pass does not complete direction6, generalize to varied overlap
or establish feedback/history authority; all seven directions remain open.
