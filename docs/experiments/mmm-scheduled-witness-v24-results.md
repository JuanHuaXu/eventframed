# Scheduled Durable Witness V24: Results

2026-10-03. **Not adopted. All eight normal screens fail.** All seven whole
research goals remain OPEN. The scheduler passes its technical controls but
does not meet the unchanged offered-load latency/freshness requirements.

## Prospective Evidence

[Protocol](mmm-scheduled-witness-v24-protocol.md),
[audit](../../research/scheduled-witness-v24/audit.json),
[run](../../research/scheduled-witness-v24/run.json),
[raw tape](../../research/scheduled-witness-v24/raw.ndjson).
1127 source/protocol/runner/checker files frozen before collection. Two
repetitions x control/scheduled x future/visible: 1024 writes, 1024 full
Recalls, 128 distinct mixed outcomes, 153600 candidate laws. Both arms enable
the V23 witness mechanism; only external admission order differs.

| Rep | Motion | Admission | Recall call p99 | Recall offered p99 | Write offered p99 | Outcome offered p99/max | View max | Screen |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | Future | Control | 49.010 | 82.894 | 108.308 | 493.937 | 38.444 | FAIL |
| 1 | Future | Scheduled | 56.531 | 188.085 | 86.811 | 397.108 | 39.243 | FAIL |
| 1 | Visible | Control | 53.023 | 69.377 | 120.349 | 469.124 | 36.224 | FAIL |
| 1 | Visible | Scheduled | 59.052 | 191.908 | 89.802 | 406.903 | 36.428 | FAIL |
| 2 | Future | Control | 50.896 | 82.663 | 107.303 | 487.943 | 30.271 | FAIL |
| 2 | Future | Scheduled | 55.360 | 191.280 | 89.491 | 407.214 | 38.337 | FAIL |
| 2 | Visible | Control | 45.918 | 71.465 | 109.230 | 476.785 | 40.522 | FAIL |
| 2 | Visible | Scheduled | 51.136 | 172.989 | 86.058 | 391.032 | 38.413 | FAIL |

All timing values ms. Gates: both Recall p99<100; write p99<250; outcome
p99<100 and max<250; view max<250. Outcome p99 equals max because n=16.
These are two descriptive repetitions, not confidence-interval tail guarantees.

Future controls use 5/16 label sources, with 11 censored per trial. Scheduled
future runs use 16/16, but the read backlog drains later: maximum observed
availability-to-first-scored-use lags are 409.425/419.206ms, versus
458.954/433.821ms for control's observed subset. Learned decision counts are
1037/1007 scheduled versus 300/319 control; cross-epoch counts 897/871 versus
260/280. Later drain permits more laws to encounter labels, so neither count
establishes improved accuracy, fixed-window freshness, or increased throughput.
Visible-motion arms transport zero beliefs. Ordinary Beta arithmetic error zero.

## Checks And Limitations

Independent sealed-source checker recomputes exact top150 nomination, as-of
admission, packed/decision law agreement, posterior arithmetic, source/epoch
identity, commit continuity, journal wire hashes and all timing gates. Eleven
corrupted-tape controls reject missing work, future events/labels, mismatched
pins, broken chains, retagged sources, altered wire hashes, grant counts,
overlap and fabricated timing/gates. Scheduler queues drain; peak queued9;
at most8 readers, exclusive writers, conserved grants. Lifecycle/race controls
cover cancellation during durable handoff, actual reopen, gaps, bypasses,
interruption and changed query/vector/settings. No live production paths changed.

Three retained follow-up uncontended scheduler microbenchmarks:
66.79-71.27ns/op,184B/op,4allocs, in
[technical evidence](../../research/joined-witness-v25/technical-prefreeze-repaired/checks.json).
This is not database, loaded serving, or end-to-end performance.

The hypothesis that admission rotation alone would rescue the loaded target
is falsified on this workload. Outcome latency improves about17-20%, but
remains391-407ms; offered Recall regresses to173-192ms. Next prospectively
test joined native-journal/marker/witness publication, not slot-weight tuning
on this consumed cohort or weaker gates. Synthetic certificates remain assumed;
visible low-impact bounds, general corpus scaling, genuine task quality and
the other six research goals are not established here.
