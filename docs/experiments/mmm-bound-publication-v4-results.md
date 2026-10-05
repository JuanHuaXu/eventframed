# Guarded bound-model publication v4: finite result

2026-10-01. Frozen contract:
[mmm-bound-publication-v4-contract.md](mmm-bound-publication-v4-contract.md).
The opt-in research publication component PASSES its frozen local boundary
checks. It does not complete Goal 6 or enter the production serving path.

The publication input now comes from a complete scan of one exclusively
owned durable stream through a declared cutoff, not caller-selected label
IDs. The scan distinguishes explicit discard from `Useful=false`, ignores
unresolved predictions, includes every terminal feedback available by the
cutoff, and rejects unwitnessed or malformed included records. An appended
admission or terminal changes a log-sequence token. Preparation validates
each included source's original journal, keyed witness, event and continuity
against one exact event-store snapshot, then freezes the new-epoch model.

Publication holds the store's exact-snapshot mutation guard and, inside it,
the durable stream's unchanged-sequence guard while comparing and swapping
the immutable slot. A competing slot revision, equal epoch, new log entry, or
source deletion between preparation and publication rejected without
replacing the visible model in the memory and temporary persistent LibraVDB
tests. The two-backend integration replayed two guarded labels from SQLite,
scanned both terminal records, and retained only the one surviving source.
Deleting that source after publication made guarded scoring fail, even though
the old pointer remained in memory.

The reader uses an as-of guard at score time. A future-only ingestion **after
publication** permitted an earlier as-of score and rejected a score at that
event's availability; scoring before the model's last admitted feedback also
failed. An event already present in the publication target is not a
post-publication transition, so this guard alone does not invalidate the
model when that preexisting future-dated event later becomes available. That
distinction was caught by an initially overstrict test and is retained as an
explicit scope limit, not counted as a stale-source bug.

**Finite component benchmark**, Apple M4 ordinary build, three repetitions
of 10,000 iterations with one witnessed label and an in-memory event store:

| Operation | Observed mean range |
| --- | ---: |
| Complete-stream preparation and source validation | 28.9-30.4 us/op |
| Exact-store plus durable-sequence publication | 11.7-11.8 us/op |
| Guarded cold baseline score | 53.6-54.9 ns/op |
| Guarded fitted 32-label component score | 85.7-88.6 ns/op |

The fitted-score benchmark substitutes a separately constructed synthetic
32-label `Frozen` model to measure computation through the same guard; it is
**not** a 32-source-authenticated service publication. Existing
`RebuildFromBoundLabels` unit tests independently verify 32 retained samples
produce fitted component models and rejected-source labels do not influence
their scores. All numbers are uncontended component means, not loaded p95/p99
or agent answer quality. Focused race tests, vet, and full affected-package
suites passed.

**Still open:** a durable new-epoch worker that can continue admitting and
learning from fresh feedback, publication recovery after process/power loss,
handling availability transitions for future-dated events already inside the
publication target when required by the serving contract, loaded tail latency
and freshness, and untouched outcome-labeled agent tasks. The slot remains
research-only; callers must explicitly handle a rejected score, and no
production OpenClaw instance or user data was used.
