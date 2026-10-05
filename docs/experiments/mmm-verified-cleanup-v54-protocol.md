# Combined verified source cleanup v54

Frozen before execution. v53's opt-in batch source owner still failed the250ms
age criterion, with539/576 completions. Its post-guard path reads and checks full
originals, then source discard resolves/decodes those same immutable originals
again. Test combining verified readback and atomic discard under the owner lock.

Four rotated arms per trial, three trials: off; raw prepared durable/union guard;
batch-source owner with separate LookupBatch and DiscardBatch; the same owner with
VerifyAndDiscardBatch. Keep192 Recall calls/four readers,96 future writes2ms apart,
K50/pack10,50 overlapping visible events, fixed as-of, queue64, groups<=4,20ms
entry budget, FULL writes and parent callback context unchanged. Cold isolated
public fixtures only; no labels, fitting, production or private data.

The combined operation must read canonical stored originals by logical source,
compare complete expected originals, and derive terminal IDs from storage, not
from caller assertions. Timestamp instants must match; non-persistable monotonic
clock/location representation is not identity. Every member must pass before
writing any terminal. Duplicate/missing/mismatched/invalid-time members reject;
existing labels cannot be erased. Exact terminal retries remain idempotent;
uncertain error/panic stops the owner until replay. Default APIs remain controls.

Test field/source/time mismatches, no partial cleanup, preserved labels, mixed
retry/reopen, monotonic timestamp roundtrip, concurrent retries and before/after
commit error/panic recovery. Run targeted and full race/vet checks before timing.

Pre-execution audit addition: reproduce and fix the shared batch-discard retry
encoding defect. Equivalent Available instants pass semantic preflight, but
re-encoding a different timezone can conflict with stored bytes. Retain the
original canonical terminal payload on retry; test prepared/unprepared and
point/snapshot preflight modes against the already-correct single-item control.
This affects all arms equally and the load fixture has no terminal retries, so
it is not the performance intervention or a threshold adjustment.

Record all request/read/write/guard/age samples, complete accounting, source hashes
and runtime metadata in exclusive JSONL. Combined readback/comparison/terminal
work gets one explicit VerifiedDiscardNS span inside PostGuardNS. Its separate
DurableVerifyNS and DurableDiscardNS are zero because not independently measured,
NOT because readback or persistence was skipped. Test containment/no double count.

Screens stay unchanged: every combined trial must have accepted-observation age
p95<=250ms and read p99<=1.10x its same-trial off. Report drops/expiry/completion,
write tails and all group sizes; a read-only pass does not erase missing work or
writer overhead. Zero unexpected errors or original mismatches are mandatory.
Go test PASS means integrity/accounting, not those latency screens. Retain all
trials without threshold tuning; a finite pass still leaves feedback/history
authority, broader traffic and whole-direction completion open.
