# Joined Journal And Witness V25: Results

2026-10-03. **Not adopted: all eight original screens FAIL.** The technical
publication component passes, and offered Recall improves in this cohort,
but background outcome freshness still misses the original100/250ms gates.
All seven whole research goals remain OPEN.

[Prospective protocol](mmm-joined-witness-v25-protocol.md),
[frozen audit](../../research/joined-witness-v25/audit.json),
[raw tape](../../research/joined-witness-v25/raw.ndjson),
[run metadata](../../research/joined-witness-v25/run.json),
[technical checks](../../research/joined-witness-v25/technical-prefreeze-repaired/checks.json).
1130 source/protocol/runner/checker files frozen before collection. Two
repetitions x joined/control x future/visible:1024writes,1024Recalls,
128distinct mixed outcomes,153600 issued candidate laws. Normal load on
Apple M4/10logical CPUs/16GiB; no benchmark-device or robot claim.

## Measured Gates

| Rep | Motion | Publication | Call p99 | Offered Recall p99 | Write p99 | Outcome p99/max | View max | Screen |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | Future | Control | 48.920 | 103.602 | 113.658 | 507.964 | 37.522 | FAIL |
| 1 | Future | Joined | 50.691 | 75.217 | 112.962 | 489.071 | 38.026 | FAIL |
| 1 | Visible | Control | 50.969 | 94.781 | 112.256 | 509.002 | 29.212 | FAIL |
| 1 | Visible | Joined | 53.833 | 53.833 | 95.450 | 459.262 | 34.952 | FAIL |
| 2 | Future | Control | 49.129 | 90.599 | 117.020 | 506.060 | 37.281 | FAIL |
| 2 | Future | Joined | 48.187 | 61.894 | 122.701 | 463.271 | 38.150 | FAIL |
| 2 | Visible | Control | 49.582 | 86.968 | 111.561 | 501.581 | 29.308 | FAIL |
| 2 | Visible | Joined | 50.663 | 67.901 | 101.436 | 473.349 | 37.201 | FAIL |

Values ms. Original gates: Recall call/offer p99<100; write offer p99<250;
outcome offer p99<100,max<250; view max<250. Outcome p99=max with n=16.
Two repetitions describe this run; no general tail-probability confidence bound.
Collector test PASS means collection completed, NOT experimental adoption.

Joined journals publish129bindings in42/39/39/38 batches (including prime),
instead of129 separate witness publications. Complete witness chain lengths
are70/68/67/66 versus156/157/157/157 control, with all144runtime mutations
accounted per trial. No acknowledgement or audit evidence is removed. Native
journal commits remain separate from SQLite: only the marker and witness share
the final transaction. This does not prove cross-database atomicity.

Future joined trials serve320/332 learned decisions,282/278 cross-epoch,
versus300/309 learned and260/270 cross-epoch control. Ordinary Beta-law
arithmetic error zero. Joined uses5/16 and6/16 label sources;11/10 censored.
Control uses5/16 each,11 censored. Observed maximum availability-to-first-use
lags423.313/434.505ms joined,479.975/455.651ms control. Censoring and differing
drain times preclude a claim that all labels became useful promptly. Visible
changes still block all belief transport; stored source epochs are unchanged.

## Audits And Repairs

Independent prospectively frozen checker passes exact150nominee/as-of oracle,
packed/journal law agreement, source/query/vector/frontier provenance, immutable
epochs, Beta arithmetic, runtime/commit continuity,128wire proofs per trial,
129binding completeness, bounded batch IDs, batch-to-chain links and
acknowledgment chronology. Twelve corrupted-tape controls reject missing
work/bindings, future events/labels, pin/source/chain violations, altered wire,
duplicate batch IDs, premature ack and fabricated timing/gates. Original V23
and V24 freezes remain byte-identical.

Race controls pass valid future/visible reopen, changed request/future evidence,
cancel-during-handoff durability, four-entry wire/witness batches, duplicate
idempotence, conflicting binding/wire rejection without partial writes,
runtime bypass/gap, closed/missing-scope calls and injected interruptions.
Native-before-final-SQLite interruption allows inspection-open but prevents
READY serving. Fully committed joint SQLite publication survives actual reopen
even when its caller never received an ack. No actual power-loss/hung-device
claim. Scheduler/validity race suites pass3repetitions; three-package vet passes.

Pre-freeze local repairs: copied-method name collision, an incorrect generic
not-found sentinel, and a test incorrectly equating inspection-open with READY.
Only new research files were repaired. The first technical-check directory
retains the compile failure; repaired exclusive evidence is linked above.
No normal tape, frozen checker, protocol, gate or consumed label was retuned.

## Consequence

Batching eliminates duplicate witness transactions and helps offered Recall,
but publication alone is **not a freshness rescue**: outcome459-489ms remains
far above100/250ms. No accuracy, agent-answer, general certificate coverage or
whole-goal success claim. Production, private sessions and whitepaper untouched.

Next prospective candidates: combine the independently tested ordering and
publication mechanisms under new unchanged-gate trials; then examine a complete
immutable law/dependency snapshot that could safely shorten read-to-ack writer
exclusion. Never release the existing lease early without that proof. Keep new
untouched-domain retrieval and richer adaptive models active research directions,
not just serving optimization. All seven original requirements remain intact.
