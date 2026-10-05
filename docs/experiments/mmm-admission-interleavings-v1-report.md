# Admission redesign boundary audit

This is a bounded specification experiment, not a runtime rescue. All seven
research goals remain open. It follows the direct gate-timing measurement.

## Evidence and model

Inspected `internal/researchmemory/source_owner.go`,
`internal/researchmemory/durable_batch_admit.go`, and
`internal/researchpublicationstore/guard.go`. SourceOwner requires the service
guard; admission stages the actual worker forecast and commits its canonical
record atomically to the ledger. Staging or commit uncertainty stops the owner
until close and replay. The guard protects compatibility through that work.

`research/admission-interleavings.mjs` enumerates every ordering of one
incompatible publication and one admission for four protocols. Its invariant
is compatibility at durable visibility. A publication is excluded while the
guard is held. This intentionally models an incompatible policy mutation,
not a future ingestion already tolerated by the as-of compatibility rules.

Results are saved in `mmm-admission-interleavings-v1.json`:

| Protocol | Terminal schedules | Invalid commits | Rejections | Valid commits |
| --- | --- | --- | --- | --- |
| Guard held through commit | 2 | 0 | 0 | 2 |
| Release before commit | 3 | 1 | 0 | 2 |
| Release, then recheck before commit | 4 | 1 | 1 | 2 |
| Reacquire, recheck, retain guard through commit | 3 | 0 | 1 | 2 |

These are exhaustive counts within the small model, not probabilities.
Negative controls explicitly require unsafe protocols to fail; positive
controls require successful admissions, so rejecting everything cannot pass.

The key counterexample is:

`capture -> release -> successful recheck -> incompatible publication -> commit`.

A recheck outside a shared atomic boundary is therefore insufficient. Retaining
the guard restores this modeled invariant but does not move durable append out
of the critical section. The previous measurement places most admission time
inside append, so this alone is not a credible performance rescue.

## Runtime cross-check and limits

Ran:

```sh
go test ./internal/service -run '^TestResearch(SnapshotInterleavings|AdmissionRejectsInterleavedPolicyChange)$' -race -count=1
```

PASS, 1.904 seconds. The admission test checks both memory and persistent
backends and rejects policy changes during event validation. This supports
the relevance of version validation but does not execute the modeled early-
release alternatives or prove their crash behavior. The model has no retries,
partial writes, worker fitting, cancellation, cross-process mutation or crashes.
No runtime implementation changed, and no performance improvement is claimed.

## Consequence for next work

CONFIRMED in this model: early release and unguarded final recheck violate the
current admission invariant. NEEDS INVESTIGATION: whether a durable prepared
record with a separate acceptance marker can preserve the full contract while
shortening guarded work. Such a design needs new replay rules: prepared records
must not become evidence, reserve an accepted source, or become acknowledged
originals before compatible activation. A crash between activation and its
durable marker must not resurrect a stale record. Merely renaming staging to
admission would change the contract, not solve it.

Before implementing that alternative, define its linearization and recovery
rules, test duplicate-source and uncertain-commit histories, and compare it
against the unchanged guarded baseline. If durable activation still requires
the same serialized append, the expected bottleneck remains. No criterion,
original result, production configuration or whitepaper has been changed.
