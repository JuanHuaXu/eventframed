# Repaired Four-Shard Pilot V2

Frozen before repaired integration validation and load runs, 2026-10-05.
This is a post-failure DESIGN experiment, not untouched confirmation.

The original one-line sharding candidate FAILED admission and reopen preflight;
those failures and the discovery-only quantization failure remain immutable.
The temporary dependency fork repairs shared-controller admission, lazy parent
discovery and new-database quantization declaration persistence. Both repaired
control and candidate use the SAME fork, so the timing contrast is the one-line
four-shard collection option, not sharding mixed with dependency repairs.
The candidate's strengthened topology fixture is unchanged and control does
not run a fixture whose assertions require sharding. No runtime mask optimization.

All seven original whole-goal criteria stay open. Production and private data
are untouched. This does not enable sharding in the live store, migrate existing
data or claim retroactive recovery of missing quantization declarations.

## Preflight

Record all modified dependency sources and regression fixtures plus eventframed
store/service/extractor/module/corpus hashes BEFORE execution. Explicitly require
all four new library/codec regressions, both ordinary and race; full singlefile
storage suite; adjacent transaction/sharding/config/persistence/lifecycle tests;
both full eventframed store suites; candidate topology under race; both vet and
all three non-optional persistent-admission/batch-parity/temporal service guards
under race. Record existing opt-in skips as NOT EXECUTED, never green coverage.
Source checks run before and after every command. Any failure blocks promotion.
Full load race, abrupt process recovery and adversarial mutation are not covered.

## Load

Reuse EXACT original formatted public capture fixture and public DESIGN corpus:
1,000 replicas of 288 capture templates, 256 concurrent future writes, all 50 or
200 selected candidates before PackK=10, 64 recalls and 64 guarded worker labels.
Future exclusion, complete durable ledger/live publication and same-epoch replay
assertions are unchanged. GOMAXPROCS=10 for both arms. Eight sequential commands:
frontiers 50/200, control-candidate then candidate-control in each frontier.

Keep all failed commands; no early truncation or ignored elapsed requests.
Candidate finite-sample recall p99 must be <100 ms and publication age p99 <250
ms in every cell. Candidate/control recall p99 must be <=0.90 in every paired
cell. Functional/preflight failure cannot be overruled by better fresh timings.
At 64 samples nearest-rank p99 equals maximum, NOT a population-tail guarantee.
Serial writer arrival is closed loop; report overlap, counts, bytes and capture
duration, without inferring equal offered open-loop load. The corpus is replicated
public DESIGN material, not 1,000 independent facts or untouched agent tasks.
