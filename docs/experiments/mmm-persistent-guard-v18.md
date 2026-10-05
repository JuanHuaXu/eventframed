# Persistent guarded admission v18

## Scope and observed gap

V17 tested memory-backed service admission with a real SQLite research ledger.
Persistent service validation existed separately, but did not establish the
combined validation/guard/ledger boundary. This was missing coverage, not proof
of a production backend defect. No production implementation changed here.

`TestResearchGuardedAdmissionPersistentBoundary` now runs four cells: memory or
LibraVDB backend, each with a concurrent policy change or future-only ingestion.
The service really observes a public fixture, recalls it, produces the research
frontier and journal, and validates the original request/event/baseline/features.

## Verified behavior

- A guarded callback commits the original service-bound prediction to SQLite.
- A concurrently started writer does not complete during the callback's20ms
  observation window; the snapshot remains unchanged at ledger commit.
- The writer resumes after guard release and advances the snapshot.
- Closing and reopening the research ledger preserves the entire original
  prediction and binding, not merely the event ID.
- The current exact-snapshot guard rejects a later operation on that old binding;
  successful persistence does not grant fresh service authority.
- With future-only ingestion, ordinary as-of admission validation still accepts
  the old query, while the exact-snapshot guarded operation rejects it. Both
  memory and persistent backends exhibit this distinction.

The last point is an availability limitation to measure, not a safety bug to
silently relax. The existing notification load experiments permit future-only
motion via temporal compatibility. Therefore their throughput/admission results
cannot be carried over to this stricter guarded durable path. A later as-of guard
would need a proof under the held mutation lock, not a check followed by an
unprotected ledger write. Feedback and whole-history validity remain separate.

## Checks run

The initial two policy cells passed three targeted race repetitions, then full
race tests and vet for service, research-memory and publication-store packages.
After adding the two future-ingestion cells, all four cells passed three targeted
race repetitions and service vet. Opt-in load experiments were not rerun.

```sh
go test -race ./internal/service -run '^TestResearchGuardedAdmissionPersistentBoundary$' -count=3
go test -race ./internal/service ./internal/researchmemory ./internal/researchpublicationstore -count=1
go vet ./internal/service ./internal/researchmemory ./internal/researchpublicationstore
```

The20ms writer observation is a scheduled concurrency fixture, not an exhaustive
scheduler proof. V17's direct mutex-ownership checks and existing wrapper-mutation
coverage support the exclusion invariant. This is ledger reopen only, not a
service crash/restart, cross-database transaction or power-loss test.

## Next work

Measure guarded admission rejection/latency under persistent writes before
claiming loaded success. Preserve the distinction between unsafe policy motion
and compatible future-only ingestion. Full durable feedback integration, unique
service-event mapping, restored-history authority and realistic agent outcomes
remain unverified; the overall research directions remain open.
