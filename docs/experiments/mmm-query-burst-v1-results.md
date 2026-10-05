# Evidence-triggered query bursts: failed primary rescue

Status: completed on all2688 consumed trajectories. Primary burst/disagreement
passes663/672 non-harm checks and0/64 required gain checks. No production
promotion. All seven full research directions remain open.

Every paid arm spends exactly31 queries on each delayed trajectory and zero on
complete-feedback trajectories. There are no paired cost mismatches. Failure
is therefore not explained by comparing unequal numbers of acquired labels.

Phase1, delayed/missing feedback, terminal64 served Brier (lower is better):

| Case | Natural | Periodic random | Periodic entropy | Periodic disagreement | Burst disagreement |
| --- | ---: | ---: | ---: | ---: | ---: |
| Additive stationary | .221755 | .222576 | .221988 | .221755 | .223971 |
| Additive gradual | .239065 | .241200 | .238236 | .239820 | .239924 |
| Parity4 | .049066 | .048921 | .048608 | .048788 | .048678 |
| Null | .258525 | .259627 | .258925 | .257876 | .258197 |
| Majority to parity | .060262 | .059603 | .059126 | .056520 | .061135 |
| Parity to majority | .102272 | .099687 | .098791 | .098032 | .100872 |

For parity-to-majority, burst/disagreement gain versus natural is.001400 with
paired interval[-.008740,.011540]. Versus matched periodic/disagreement the
gain is-.002840[-.012279,.006599]. Neither establishes a benefit. All nine
non-harm failures are phase1 delayed terminal comparisons in cases0,19,20.
Their intervals cross zero: failure to certify non-harm is not proof of harm.
Both phases were already consumed, not fresh confirmation.

Burst/random has a better point estimate on case20 (.096207), but it is not
the frozen primary candidate and a selected case cannot rescue the broad claim.
All seven arms are retained in the summary for secondary comparisons.

## Mechanism diagnostics

Across1344 delayed trajectories, each paid arm buys41664 labels. For the matched
disagreement pair, the fraction adding missing training evidence or reaching
the first64-window fit before natural delivery improves41.23% to61.90%.
Mean wait from paid reveal to the first included fit falls18.56 to9.27 frames.
Yet prediction quality does not reliably improve. Faster useful delivery is
not sufficient evidence that the learned forecast improves.

Of41664 burst/disagreement queries,19567 occur inside the surprise-triggered
window and22097 use the forced-spend fallback. Thus more than half the queries
are scheduled by budget completion, not a revealing surprise. The stationary
parity4 case still averages19.75 surprise-window queries out of31 in phase1.
This trigger is not a calibrated changepoint detector; ordinary surprising
outcomes can activate it without any regime change.

Four queries per trajectory are after the final fit and do not enter a new fit
within the horizon; they can still affect the mixer. There are319 late acquired
labels outside the retained mixer journal for burst/disagreement, still admitted
to training/observation. These are explicitly recorded, not silently dropped.
Diagnostics concern the64-window expert; the32-window fit sets are also audited.

## Verification and cost

Collection completed in590.38s wall (587.903s Go package),2351.97s user CPU,
6.71s system CPU, with four workers. This is an offline seven-arm experiment,
not serving latency or evidence of production throughput improvement.

The independent audit checked24,084,480 probabilities, every fit origin list,
arrival/trigger, query clock/pool, exact seeded random choice, and maximal
entropy/disagreement utility within1e-10. A direct-probability Markov recursion
also verifies the served mixture independently of Go's log-space journal.
The expert fitters themselves are not independently reimplemented here.
Four deliberately corrupted records were rejected. Full scoring replay is
byte-identical (`cmp` exit0), not a second collection. Prior repaired race tests
and their initially caught expired-journal failure remain documented.

Artifacts: [protocol](mmm-query-burst-protocol.md),
[raw](mmm-query-burst-v1.jsonl), [summary](mmm-query-burst-v1-summary.json),
[replay](mmm-query-burst-v1-summary-replay.json), [run](mmm-query-burst-v1-run.txt),
[integration](mmm-query-burst-contracts.md), [audit contract](mmm-query-burst-scoring-contract.md).

## Next lead

Do not tune this surprise threshold to the consumed outcomes. The intervention
improved training lead without a reliable quality gain. Next isolate expert
publication cadence from label availability in the current four-expert learner,
before claiming another acquisition score will help. Earlier cadence v6 used
different forest learners, so it does not settle this particular boundary.
A faster-refit diagnostic must report its extra compute and is not an equal-cost
rescue or performance success. No new cadence policy is implemented here.
