# Paired betting certificate v4: fixed-model component pass

The [frozen protocol](mmm-paired-bet-v4-protocol.md) **passes its declared
component screen** on independent design and confirmation seeds. A bounded
betting e-process, averaged over eight predeclared starts, two contexts and
two signs, recovers strong-shift detection from the same sparse-delayed
matched pairs on which the v3 Hoeffding window was silent. This is not
actual MMM sharing authority, a target-law diameter certificate, or proof
of better downstream forecasts.

## Method and verification

For each contemporaneous same-context pair, `D=Y_ref-Y_live` enters only
after both labels arrive. Each signed, context-specific process multiplies
by `1+0.8*(sign*D-0.1)` for pairs at or after its fixed origin start.
Future starts remain at wealth one; the 32-process arithmetic mean flags
at 50. Under the declared per-clock null and outcome-blind, independent
arrival model, each factor is positive and has conditional expectation at
most one. Averaging, rather than multiplying correlated starts, retains
the one-pair e-process bound. The design is inspired by
[Waudby-Smith and Ramdas](https://arxiv.org/abs/2010.09686), but the fixed
start schedule and threshold are this experiment's own construction.

- [Simulator](../../research/paired-bet-v4.mjs) and [independent summary verifier](../../research/paired-bet-v4-verify.mjs).
- [Design JSONL](mmm-paired-bet-v4-design.jsonl): 14,000 trials, SHA256 `8abc363710a22907374883be0213a3db8af204589d873abab6051e41b706844d`.
- [Confirmation JSONL](mmm-paired-bet-v4-confirmation.jsonl): 14,000 fresh trials, SHA256 `075cfd53ab7452d1cfc60307a3a8151e2ef917bcf7103c630bd1c97fbc4d07b9`.
- Both artifacts pass independent row-key, resource-count, source/protocol
  hash, and summary checks. A fresh confirmation replay matches all 14,002
  JSONL entries after excluding elapsed time. Replaying the same simulator
  is not an independent implementation of the candidate statistic.
- Embedded negative controls reject future, partial, missing, unnominated
  and duplicate pair updates without changing gate wealth.

## Confirmation results

Flags are counts per 1,000 trajectories. Median clocks are conditional on a
flag, so missed trajectories are not hidden inside the reported delay.
The matched conditional Hoeffding window and context-blind window see the
same usable pairs and acquisition cost.

| Schedule and case | Betting flags | Hoeffding flags | Scalar flags | Betting median clock |
| --- | ---: | ---: | ---: | ---: |
| Complete, stable | 0 | 0 | 0 | none |
| Complete, null boundary | 4 | 0 | 0 | 317 |
| Complete, common co-drift at 256 | 0 | 0 | 0 | none |
| Complete, live-only shift at 256 | 1,000 | 1,000 | 0 | 286 |
| Complete, reference-only shift at 256 | 1,000 | 1,000 | 0 | 285 |
| Complete, live-only shift at 224 | 1,000 | 1,000 | 0 | 265 |
| Complete, live shift from start | 1,000 | 1,000 | 0 | 29 |
| Sparse-delayed, stable | 0 | 0 | 0 | none |
| Sparse-delayed, null boundary | 4 | 0 | 0 | 337 |
| Sparse-delayed, common co-drift at 256 | 0 | 0 | 0 | none |
| Sparse-delayed, live-only shift at 256 | **936** | 0 | 0 | 435 |
| Sparse-delayed, reference-only shift at 256 | **928** | 0 | 0 | 435 |
| Sparse-delayed, live-only shift at 224 | **985** | 0 | 0 | 410 |
| Sparse-delayed, live shift from start | 1,000 | 0 | 0 | 203 |

All clock-224/256 shift cells have zero prechange flags. The corresponding
complete-evidence Hoeffding medians are 329/361; betting reaches 265/285-286.
In sparse-delayed confirmation, the clock-256 median delay is 179 clocks
among detections and the p95 flag clock is 495-496, close to the 512-clock
horizon. The 64/1,000 and 72/1,000 misses matter; this is not a universal
fast-split result. Per-cell Wilson 95% intervals are approximately
91.9%-95.0% for 936/1,000 and 91.0%-94.2% for 928/1,000. The 4/1,000
boundary false-flag rate has a per-cell Wilson upper endpoint about 1.02%.
These empirical intervals are not the sequential theorem and are not
simultaneous across cells.

The sparse-delayed confirmation cases average roughly 128 nominations,
256 context readings, 198-199 observed labels, and 78-79 complete pairs
over all 512 clocks. The candidate performs about 680-688 signed factor
updates per trajectory, versus 4,608 with complete acquisition. The
14,000-trial split took about 4.7 seconds in the isolated Node simulator;
that includes generation, controls, and artifact construction. It is not
a loaded service p99 or a benchmark of an optimized Go implementation.

## Proof boundary and next decision

The supermartingale claim applies to the synthetic generator's independent
outcomes across clocks, outcome-independent nomination/missingness/delays,
and a gate filtration that **cannot inspect one-sided labels**. The protocol's
origin-time conditional-mean statement alone would not suffice for a
correlated delayed stream: later observations could reveal information about
a pending pair before its update. Operational validity requires the
conditional mean bound at *delivery time relative to everything the gate
or selector has seen*, or a different proof accounting for that information.
This post-run scope clarification does not alter the frozen synthetic data
or thresholds; it limits what may be inferred from the pass.

Also, a latched flag certifies evidence of historical divergence since a
predeclared start. It does not prove the members remain divergent now,
does not handle unmatched contexts, and gives no simultaneous protection
over multiple groups or externally identified causal effects. No scored-law,
member-share, answer-quality, or loaded-latency result was measured.
Keep production untouched and Goal 3 open. The next necessary test is an
actual MMM member integration with realistic match availability, downstream
proper scores, explicit multi-bucket error allocation, and a delayed-feedback
filtration audit. Only after that should serving-cost work begin.
