# Resolved source admission v55 results

**Overall: FAIL the unchanged loaded screen.** Resolved admission completes569/576
observations versus566/576 for combined-cleanup control, but only one of three
age p95 values passes250ms. The250.051833ms value fails; rounding cannot change
that decision. Read tails pass all three; writer tails remain elevated. This is
an opt-in research implementation, not a default, deployment or completed lead.

## Implementation and audit

`OpenSourceOwnerResolvedAdmissions` keeps canonical source resolution under the
private owner lock, then reuses those call-local decoded results in private
Durable admission. Existing source constructors and raw Durable APIs retain
their original preflight reads. No public trusted hints, retained cache or new
evidence/serving authority. Source resolution binds tenant, stream, contract,
journal, event, learner ID, seed, epoch and canonical original before reuse.

The common admission path still checks input equality, order and pending caps,
stages actual owned-worker forecasts, and atomically appends with exact retry
and acknowledgment checks. An error or panic after staging stops the owner.
The unique source index remains enforced. Asynchronous fitting may change the
snapshot used between new records; no synthetic preview replaces an original.

Tests count zero learner-ID preflight reads in the new path versus two batch
reads for two control admissions, while both retain two atomic append calls.
Matched training/publication barriers preserve warm originals across point,
batch and resolved owners, and across reopening through a control constructor.
Mixed retry/new and timezone-equivalent retries preserve original records.
Conflicts, duplicate sources, cancellation, capacity, concurrent retries and
before/after-commit errors/panics retain rejection and replay behavior.

An intentionally unexpected private-log write between resolution and append
conflicts at the final atomic boundary: no candidate original is acknowledged,
no candidate batch member is persisted, the owner stops, and replay restores
the correct next ID. This deliberately violates ownership for a negative test;
it is not support for external raw writers or authentication of their evidence.

Verification before measurement:

```sh
go test -race ./internal/researchledger ./internal/researchmemory ./internal/service -count=1
go vet ./internal/researchledger ./internal/researchmemory ./internal/service
```

PASS: ledger3.721s, learner13.693s, service56.373s; vet clean. Earlier targeted
resolved/warm checks passed three repetitions. One test-wrapper compile error
was corrected before the successful full run and before measurement.

## Frozen load comparison

[Protocol](mmm-resolved-admission-v55-protocol.md),
[raw artifact](mmm-resolved-admission-v55.jsonl). SHA-256:
`f4b11e36a0d89fd7a2ed3466ca395c8b58c2f58311bf5780d750e57c74e969eb`.

```sh
EVENTFRAME_RESOLVED_ADMISSION_ARTIFACT=<LOCAL_ROOT>/docs/experiments/mmm-resolved-admission-v55.jsonl go test ./internal/service -run '^TestResearchResolvedAdmissionExperiment$' -count=1 -v
```

Go1.27.1 darwin/arm64, GOMAXPROCS10. Go PASS21.355s means accounting/integrity,
not performance success. Twelve rotated cells record2304 Recall calls and1152
future writes in fresh isolated public fixtures. No labels, fitting, private
data or production access. Every recorded source hash matches its captured
source and the measured local file. Raw sample lengths, counts and phase
containment were independently checked after execution.

Nearest-rank quantiles, milliseconds:

| Trial | Raw / combined / resolved completed | Resolved drops | Combined / resolved age p95 | Off / resolved read p99 | Off / resolved write p99 |
| --- | --- | --- | --- | --- | --- |
| 0 | 192 / 182 / 192 | 0 | 275.855 / 235.352 | 36.040 / 26.138 | 22.465 / 37.886 |
| 1 | 192 / 192 / 192 | 0 | 246.892 / 250.052 | 41.031 / 22.430 | 18.930 / 32.565 |
| 2 | 188 / 192 / 185 | 7 | 238.092 / 276.898 | 34.978 / 27.991 | 18.732 / 37.517 |

Resolved age passes only trial0. Its read p99 stays within1.10x same-trial off
in all trials. Resolved writes exceed both off and combined control in all
three: combined write p99 is35.995,28.023,29.039ms. No writer-nonharm claim.
The raw control loses four observations to entry expiry in trial2 and also
fails its read screen there:42.088ms versus34.978ms off. Retain that control
failure, rather than using earlier raw-control passes as this run's result.

Resolved has seven drops and zero entry expiry; combined has ten drops and zero
expiry. Accepted original/readback/terminal checks cover28450 resolved records,
28300 combined and28600 raw, total85350. There are zero unexpected errors or
original mismatches. Rejected/expired/dropped observations are not evidence.

Mean admission time per accepted group, control versus resolved:
6.820/6.145,6.196/5.807,5.835/6.477ms. Eliminating the extra reads is confirmed
structurally, but loaded admission time improves in only two trials. Mean
verified cleanup time is7.256/6.362,6.717/6.476,7.127/6.323ms even though cleanup
code is unchanged between these arms. These are variable scheduling/load cells,
not pure isolated admission-cost estimates or proof of a uniform speedup.

## Interpretation and next lead

Do not adopt the optimization as a reliable latency rescue. Current tests use
closed-loop concurrent readers: each arm's serving speed also changes when
observations enter its bounded queue. Host variation, arrival timing and actual
per-group storage cost are competing explanations for the mixed outcome.
Separate them with a frozen isolated admission-cost comparison and/or a matched
arrival-schedule diagnostic before another optimization. A changed workload
would be a separately labeled test, not a retroactive rescue of this screen.

No thresholds, stored history or original forecast semantics were changed.
Full feedback/history authority, fitting under load, wider overlap/arrival
regimes and real outcome validation remain open. All seven research directions
remain IN PROGRESS; nothing was pushed or deployed.
