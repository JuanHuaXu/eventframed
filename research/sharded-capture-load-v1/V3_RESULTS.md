# V3 Final Repaired Pilot

**PASSED the unchanged finite DESIGN timing/functional pilot**, including a
complete candidate/200 functional race recheck. NOT whole Goal 6, production
readiness, an untouched confirmation or completion of any of the seven goals.

## Integration

V3 retains the three V2 temporary dependency repairs and adds synchronization
of completed quantizer samples. V2 remains blocked by its full-load training
slice race. The small pre-patch unit test did not reproduce that race, and its
passing result is preserved. The full V2 failure is the authoritative witness.

All 12 V3 preflight commands pass their declared FUNCTIONAL criteria. Five
named component suites run normally and under race; 78 storage and 67 adjacent
library suites run; both full store suites, vet and three service guards run;
candidate topology/reopen runs under race; and the COMPLETE candidate/200
fixture runs under race without a reported data race. Each store arm retains
119 unexecuted opt-in skips. No source-pin test is silently dropped.

The instrumented full-load race command exits **1**, solely on the original
recall <100 ms timing screen. Its 64 recalls, 256 captures, 64 live worker labels,
128 durable rows and replay complete. This is explicitly a functional race
pass, NOT a passing timing command or a gate relaxation for the ordinary runs.
All-arm/frontier race coverage, crash and adversarial lifecycle are still open.

## Ordinary Performance

Same public fixture, ten workers, both arms using the SAME V3 fork, no concurrent
benchmark, all candidates scored before packing. Two opposite execution orders.

| Frontier | Pair | Control Recall p99 ms | Candidate Recall p99 ms | Ratio | Candidate Live Age p99 ms |
| --- | --- | --- | --- | --- | --- |
| 50 | control then candidate | 116.764 | 48.512 | 0.4155 | 11.483 |
| 50 | candidate then control | 115.110 | 46.578 | 0.4046 | 11.562 |
| 200 | control then candidate | 172.096 | 88.679 | 0.5153 | 18.641 |
| 200 | candidate then control | 164.590 | 61.667 | 0.3747 | 12.876 |

All candidate cells pass recall <100 ms, live publication age <250 ms and paired
ratio <=0.90. All control cells preserve their ordinary timing-failure exit.
Recall p99 reduction is 48.47-62.53%, with worst candidate sample 88.679 ms.
At 64 recalls p99 is the maximum; no population-tail or confidence claim.

All arms complete the identical required records/labels and check same-epoch
replay. Initialization bytes 182,269, capture bytes 46,766; overlap counts
204 at frontier 50 and 213 at 200 in both arms. Candidate capture p99
15.92-23.95 ms vs control 40.93-49.16 ms. Candidate initialization 7.30-7.71 s
vs control 15.39-15.77 s. Closed-loop offered rate still changes with cost.

V2 versus V3 timings are NOT a paired training-mutex overhead estimate: the
separate runs have scheduling/index-order variance and different matched-control
tails. The practical V3 claim is that its own frozen gates pass, not zero repair
cost or a measured causal performance penalty relative to V2.

## Next

Keep V3 isolated. Larger-corpus scaling, sustained open-loop load/queue failure,
memory retention, visible mutations/crash/cross-epoch transfer, network contracts,
retrieval/scored-law quality and untouched agent outcomes remain required. See
NEXT.md. No live cache, production config, private corpus or whitepaper changed.
The .698246 approximation defect and the other six whole-goal science gaps are
not repaired by this storage result. All seven whole goals stay OPEN/ACTIVE.
