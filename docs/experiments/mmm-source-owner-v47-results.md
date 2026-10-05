# Source owner v47: identity integration

Status: PASS for the tested identity/lifecycle contracts only. All seven research
directions remain in progress. No production consumer, feedback authorization,
model-serving authority, or learning-quality improvement is established here.

## Evidence and patch reasoning

The v44 audit reproduced one service journal/event receiving two distinct learner
IDs in the raw durable API, then replaying two labels. That API's key is the
learner ID; it does not claim source deduplication. The older service bridge
refused the duplicate, and exact-ID retries were correct. This local composition
gap is confirmed across single/batch and restart cases, not inferred from corrupt
production rows. These experimental files are local/unpublished work; no upstream
patch was relied upon or production artifact repaired.

The v46 unique source index prevents duplicate bound originals in one transaction,
but rejection after staging a second learner ID would unnecessarily stop the
worker. v47 therefore resolves the logical source before staging. The falsifier
is any repeated journal/event returning a fresh ID, or a late input conflict
advancing the worker or creating a partial batch.

## Change

`OpenSourceOwner` privately owns the durable worker and ledger. Callers supply
features, baseline, observation time and a service binding, never a learner ID.
Under a serialized owner lock, every source is resolved through the five-field
index. Exact retries return the complete persisted original; changed immutable
inputs conflict. New source IDs are allocated consecutively only after preflight.
The existing prepared atomic batch path writes the original and reserves source
identity in the same SQLite transaction.

Startup first replays canonical, bound history, then enables the unique index.
Unbound or duplicate-bearing legacy history fails without silently rewriting
records. Existing raw durable constructors and permissive replay remain controls.
Terminal discard retains the source identity and never supplies a negative label.
An uncertain commit stops admissions and readback until close/reopen/replay.

## Verification

- `go test -race ./internal/researchmemory -run TestSourceOwner -count=3`: PASS.
- `go test -race ./internal/service -run '^TestResearchSourceOwnerGuardedRetry$' -count=3`: PASS.
- Full race suites for `internal/researchledger`, `internal/researchmemory` and
  `internal/service`: PASS. `go vet` for those packages: PASS.
- Cases cover exact originals across restart and discard, mixed new/retry order,
  separate journals/events, detached output bindings, and eight concurrent retries.
- Late snapshot, feature, baseline, time and tenant conflicts; duplicate sources;
  invalid probabilities, invalid UTF-8 and cancellation reject without staging.
- Legacy unbound/duplicate histories refuse activation and retain both rows, with
  no usable index left by failed activation.
- Injected acknowledgment failure before and after a real commit stops the owner;
  reopen correctly distinguishes a new admission from the persisted exact retry.
- The real service-journal fixture validates admission under its publication guard,
  retries across three restarts, rejects a stale-policy retry, and distinguishes
  immutable original lookup from current permission. Synthetic public fixture only.

Process-exit atomicity was tested for the underlying index and durable batch in
v46 and earlier; v47's new wrapper tests acknowledgment uncertainty, not a new
process-kill campaign. This is not a claim of exhaustive failure coverage.

## Remaining boundaries

The caller must still hold the real service admission guard. The owner is not
installed in a service path and intentionally exposes neither raw feedback nor a
scoring API. Replaying old labeled rows grants no fresh history authority. A
separate trusted feedback path and history invalidation mechanism remain required.

The owner performs one indexed lookup per source (at most 256), followed by the
existing batch preflight/write. Memory is bounded by the request and fixed learner
caps; disk history/index and startup replay are not bounded in lifetime size.
No new performance experiment was run for this wrapper. v46 point-lookup timings
must not be reported as v47 loaded latency. Next test batched source admission
cost and wire explicit feedback/history authority before any serving claim.
