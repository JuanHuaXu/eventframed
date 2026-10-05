# Four-Shard Public Capture Pilot

## Verdict

Original one-line candidate: **FAILED preflight**. Temporary repaired V2:
**PASSED the frozen finite DESIGN pilot**, not whole Goal 6 or a deployment.
All seven whole goals remain OPEN. Original failures are retained, not replaced.

**Subsequent promotion blocker:** V2's complete candidate/200 race run reported
a training-vector slice race despite completing every required functional row.
That command FAILED and V2 must not be promoted. V3 is a separate synchronized
training rescue with separately recorded validation/timing; V2's ordinary
performance observations remain historical evidence, not production readiness.

Three reproducible integration defects were isolated and repaired in a temporary
LibraVDB copy: duplicate admission permits for shard views sharing one controller,
missing logical parents during lazy discovery, and missing durable quantization
declarations. The quantization regression also fails on ordinary collections;
the fix therefore covers both creation/reload paths and recovery-index creation.
None/SQ8/FSQ declarations and two reopens were tested; this is not a general
quantizer-training, memory-compression, PQ or downgrade guarantee.

## Preflight

All 11 repaired validation commands completed successfully: four new regression
suites normally and under race; 78 storage suites; 67 adjacent library suites;
48 active control store suites and 49 active candidate store suites; vet in both
arms; all three named service guards under race in each arm; and the original
candidate topology test under race. Each store arm also has 119 existing opt-in
SKIPs, explicitly unexecuted rather than passing coverage. Full library/module
tests outside the recorded paths/regex are not claimed. Original six-minute
candidate admission timeout and both reopen failures remain in the archive.

Both load arms use the SAME repaired fork. Their frozen store, service,
extractor, public corpus and timing fixture differ only by the event collection's
four-shard option. This is not a mask-cache optimization or daemon default change.

## Ordinary Load

64 recalls per cell, 1,000 replicas from 288 public DESIGN capture templates,
256 concurrent future captures, all 50/200 candidates before PackK=10, explicit
ten-worker limit, fresh stores, two opposite run orders. No private data,
confirmation answers or real agent utility labels.

| Frontier | Pair | Control Recall p99 ms | Candidate Recall p99 ms | Ratio | Candidate Live Age p99 ms |
| --- | --- | --- | --- | --- | --- |
| 50 | control then candidate | 112.381 | 45.729 | 0.4069 | 11.389 |
| 50 | candidate then control | 124.052 | 65.394 | 0.5271 | 11.333 |
| 200 | control then candidate | 158.087 | 62.058 | 0.3926 | 15.027 |
| 200 | candidate then control | 157.200 | 60.929 | 0.3876 | 12.391 |

Every candidate cell passes recall <100 ms, offered-to-live age <250 ms and
the additional ratio <=0.90 screen. All ordinary control cells retain nonzero
timing-failure exits. Recall p99 reductions are 47.29-61.24%; these are sample
maximum comparisons, not statistical confidence limits on production tails.

Every run completes 256 captures, 64 recalls and 64 live worker labels before
Close; 128 durable admission/feedback rows and 64 same-epoch replay completions
are checked. Initialization bytes 182,269 and capture bytes 46,766 match across
all arms; actual overlap is 204-213 captures. Candidate capture p99 is
16.06-22.40 ms vs control 38.95-46.00 ms. One candidate capture maximum is
40.95 ms, so occasional slower writes are retained. Initialization falls from
15.23-15.41 s to 7.37-7.52 s. A closed-loop writer is NOT equal offered
open-loop load, even with identical total records/bytes and similar overlap.

## Remaining Requirements

This demonstrates useful smaller-index publication under this loaded fixture,
not constant-cost ingestion, ranking-quality equivalence, retained-RAM bounds,
large-corpus ANN performance, sustained open-loop backlog, cross-epoch learning
transfer, process-kill recovery, network serving or untouched agent outcomes.
Public fixture worker labels are synthetic mechanics labels. The seven-goal
science gates, .698246 TV approximation defect and observation-quality evidence
are unchanged. See AUDIT.md and the frozen protocols for scope and falsifiers.

The supplemental full candidate/200 race run and missing-child negative control
have their own protocol and artifacts; their verdicts are reported separately
and do not rewrite ordinary timing results.

The missing-child control PASSED normally and under race: parent discovery does
not make an incomplete shard set loadable, and failed load/ensure calls leave
the remaining physical collection names unchanged. The small four-index
training regression PASSED even before V3, so it did NOT independently reproduce
the V2 failure. The complete loaded race stack is the failing witness; do not
present the targeted pre/post unit comparison as a reproduced failing regression.
