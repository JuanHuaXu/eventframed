# Historical Archive V29-V32 Results

2026-10-03. Technical/lifecycle evidence only. All seven whole goals OPEN.
Production, private corpora, whitepaper and old source freezes unchanged.

## V29 Owned Handoff

Frozen1140files, race-enabled five cases PASS: future insert, visible insert,
ordinary outcome, outcome plus visible insert, canceled capture plus outcome.
Each mutation acknowledges while the historical journal remains behind an
explicit barrier; no read admission is held and no caller acknowledges early.
Exactly150recorded decisions/case, ten packed for non-cancel calls;142425-142455
byte journals preserve original snapshot/wire/laws. Cancel persists its accepted
capture and gives no packet. Zero post-handoff service-facing EventStore calls
in the tested default path; all31methods instrumented and AST coverage checked.

Ordinary current journal gate admits ONLY the future-insert control; historical
archival preserves the other four without pretending their dependency versions
are current. Full-stream feedback from all five archives is accepted; selected
feedback from visible/mixed stale-epoch archives is rejected. Native/witness
reopen preserves byte commitments. Six live authority corruptions and seven
independent raw-audit corruptions rejected. Inverse-AST check proves native
receipt/readback/marker logic unchanged beyond declared admission predicate.

Evidence: `research/archive-boundary-v29/{freeze,run,raw,audit}.json`, `test.log`.
Raw SHA256: `75a0a1ec8017972253879498d9cede86a69e04602fdcd9c146660208fea28b44`.
Timing is race/barrier diagnostic, not adoption.

## Preserved Failures And Repair

V30 frozen1140files: interruption points/ownership PASS, vet PASS, scheduler/
validity race checks three repeats PASS;16concurrent archives FAIL with native
marker context deadline. Confirmed code defect: the500ms operation timer starts
before owner wait, unlike sealed V25. V31 research wrapper puts the SAME timer
after owner acquisition; end-to-end queue time is not excluded from future gates.

V31 frozen1144files: interruption/ownership, vet and repeated core race checks
PASS; concurrent control reaches `lost/duplicate archive`. Original control
assumes one distinct journal per clock-sampled call without recording request
multiplicity. V30/V31 failure runs and sources remain unchanged, not relabeled
as passing confirmations. No captured V31 request trace exists, so its exact
duplicate pair cannot be retrospectively established.

## V32 Request-Identity Diagnostic

New frozen sources; V31 runtime unchanged.48submitted requests across three
race-enabled scenarios all acknowledge and preserve original wires/laws/reopen.
Two outcomes/one visible insert per scenario complete behind handoff barriers.

| Requests | Unique Archives | Duplicate Requests | Result |
|---|---:|---:|---|
| Actual concurrent clock samples |14|2|Identical timestamps and full wires; valid deduplication|
| Distinct past AsOf nanoseconds |16|0|All distinct requests preserve distinct journals|
| Intentional identical retries |1|15|All retry acknowledgments bind to the one immutable archive|

V32 explains why the earlier uniqueness assertion is unsafe; it does not prove
the unrecorded V31 pair. Every returned packet binds to its own AsOf, snapshot
and full report. Post-run supplementary independent audit checks all inputs/
captures/outputs, including nanosecond timestamp multiplicity, and seven
corruption controls. Supplement is explicitly post-hoc, not a predeclared
adoption screen. Artificial barriers and race instrumentation make diagnostic
elapsed durations unsuitable for the original full-serving latency gates.

V32 raw SHA256: `a5f36e3e6b773103f3da14c7cd91352256e8dd6e62c8cbaa85ca6849c602d8d9`.
Supplement auditor first assumed the journal JSON key was `journal_id`; actual
schema uses `id`. Corrected that checker-local field before any supplement was
written; original raw/runtime sources were not changed or replayed. Pre-freeze
compilation also caught a duplicated generated V25 type and a local name shadow;
both fixed before any V29 scenario run. Neither is a runtime or quality finding.

## Remaining Work

This is a new historical-record contract, not a change to current evidence
validity. Process-local capture authority is not adversary-resistant signing;
reopen validates original witness/native state, with no invented repair.
Cross-database failure remains fail closed; no cross-store atomicity theorem.

Full ancestry validation is linear in retained transitions; per-archive JSON
copy/sealing, native byte readback and one marker/witness transaction still cost
work. Queue wait lacks an overload bound. Next test batched archival publication
and explicit bounded admission with the original independently offered mixed
workload; measure complete call/offer tails and useful learning freshness. Do
not hide queue time, lower the offered rate, or relax100/250ms gates. Optional
async service configs and concurrent close remain outside present evidence.
