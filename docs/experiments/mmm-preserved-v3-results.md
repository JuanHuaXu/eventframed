# Preserved-incumbent MMM v3 results

## Verdict

**Passed the frozen synthetic rescue criteria.** The preserved-incumbent mixture
improved shifted forecasts while keeping stationary harm within the declared
0.01 Brier ceiling. MMM had a resolved incremental advantage over the matched
breadth-first and random observation controls. Anti-Pigeon's incremental forecast
benefit remained unresolved, even though its guarded sharing transition operated
without observed stable false splits.

This is not full recovery of old-regime accuracy, causal discovery, real-data
validation, or a production promotion. The experiment used synthetic EventFrame
envelopes with supplied scopes and binary fields, not a live OpenClaw instance.

## Protocol and artifacts

- `mmm-preserved-v3-protocol.md`: frozen models, seed domains, criteria and costs.
- `mmm-preserved-v3.json.gz`: all pre-outcome forecasts, expert weights, observed
  masks, outcomes, audit flags, fit versions, nominations, evidence and splits.
- `mmm-preserved-v3-summary.json`: per-stream scores, all 100 contrasts, hashes.
- `mmm-preserved-v3-benchmark.txt`: three serial timing repetitions.

320 streams x 512 live steps x six paired policies = 983,040 forecasts on 163,840
live steps. There are 32 independent streams per split/scenario, not 983,040
independent observations. Extra reference observations are explicitly charged.
Fit seed 2026091701; design/confirmation bases 2026091702/2026091703. Disjoint
scenario/stream/role seed ranges were tested before either evaluation split ran.
No thresholds, model settings or criteria changed after seeing either split.

## Confirmation

Lower Brier is better. Full = all 512 steps; post = last 256 for single shifts
and steps 128--511 for recurring. All warm-up and adaptation delay are included.

| Scenario / window | Frozen MMM | Matched breadth mixture | Matched random mixture | Preserved MMM + AP | Result |
| --- | ---: | ---: | ---: | ---: | --- |
| Stable / full | 0.049016 | 0.251713 | 0.176488 | 0.049409 | Stationary protection passed |
| Member shift / post | 0.441229 | 0.251649 | 0.252452 | 0.243053 | Shift gain passed |
| Common shift / post | 0.452182 | 0.251056 | 0.251059 | 0.240417 | Shift gain passed |
| Recurring / post | 0.318638 | 0.251783 | 0.235997 | 0.197875 | Gain over frozen and observation controls |
| Null / full | 0.254928 | 0.251852 | 0.252003 | 0.252166 | Uncertain; no useful signal claimed |

Gain versus frozen MMM on member/common shift was 0.198176
[0.177080, 0.219272] and 0.211764 [0.194148, 0.229381], respectively, meeting
both the 0.05 absolute-gain requirement and positive lower bound. These correspond
to about 44.9% and 46.8% relative Brier reduction.

Stationary harm was small but **not zero**: 0.000392
[0.000219, 0.000565]. Its upper bound is below the frozen 0.01 ceiling. Do not
describe this result as bit-identical, zero-regression, or universally superior.
The design split also met the same gain/protection criteria (stationary harm
upper bound 0.000411; member/common gain lower bounds 0.194328 / 0.187563).

Intervals are approximate z=3.5 paired normal intervals over 32 stream means,
conditional on the single fitting model. They are distinct from the sharing
gate's conditional sequential guarantee. No confirmation scenario had mean
Brier harm over 0.01 relative to frozen MMM, in either full or post windows.

## What contributed

MMM's gain over the otherwise matched breadth-first mixture was:

| Scenario / window | Brier gain | Approximate simultaneous interval |
| --- | ---: | --- |
| Stable / full | 0.202305 | [0.196700, 0.207909] |
| Member shift / post | 0.008596 | [0.002339, 0.014853] |
| Common shift / post | 0.010639 | [0.004168, 0.017110] |
| Recurring / post | 0.053908 | [0.046942, 0.060874] |

Random-observation contrasts were also positive in these four windows. These
controls use the same six-coordinate foreground cap, model fitting, audits and
mixture algorithm. The supplied XOR/local-bit fixture rewards appropriate scope
and depth selection; this is not proof of benefit across arbitrary real tasks.

Removing AP while retaining the preserved MMM mixture yielded member-shift
Brier 0.243617 versus 0.243053 with AP: gain 0.000564
[-0.000375, 0.001502], unresolved. Common-shift forecasts were identical because
neither group split. Recurring AP gain over no AP was -0.000348
[-0.001846, 0.001149]. The local-only challenger also had no resolved difference
from AP in these windows. Thus the rescue cannot be attributed uniquely to AP;
preservation, forecast blending and independent audits are the principal tested
composition. No ablation here separates each of those three in isolation.

Post-change accuracy was 57.67% / 58.69%, still far below stable accuracy.
Confident errors fell from 4035/8192 to 115/8192 and from 4137/8192 to 115/8192.
Much of the gain therefore means avoiding confidently wrong predictions, not
fully learning the new relationship. The positive gain over breadth/random is
smaller than the gain over the overconfident frozen model.

## Sharing and timing

Stable, common-shift and null confirmation had zero sharing revocations. Stable
0/32 has Wilson 95% upper bound 10.72%; the sample alone does NOT establish a
population false-split rate below 5% or 1%. The frozen criterion was an empirical
rate ceiling, and the separately derived .01 anytime bound requires the stated
conditional-mean null on the observed reliability proxy.

All 32 member-shift streams eventually split, none before the true change;
mean delay was 117.78 steps (all 32 detected, Wilson detection interval
[89.28%, 100%]). This is not fast structural adaptation. Mixture weights may
respond before sharing is revoked. Recurring also split 32/32, with a 197.78-step
mean from the FIRST change; that includes later regimes and must not be read as
per-change detection. No re-sharing is implemented in this experiment.

The gate requires existing revision nomination plus a two-sided bounded betting
process with eight fixed starts and a per-stream .01 error budget. It monitors
original-policy correctness, not candidate self-consistency. It certifies neither
the full target-law diameter nor causal equivalence. A common shift does not
force a split when both contexts remain compatible.

## Costs and boundaries

Informative preserved MMM averaged 4.00--4.88 foreground coordinates per step,
versus approximately six for breadth. All experts share the SAME observed mask,
so there is no hidden six-per-expert allowance. In addition every arm was charged
eight diagnostic coordinates and about 4.45--4.52 audit coordinates per live step.
Null uses approximately six foreground + twelve diagnostic + 4.61 audit units.
These are inspection counts, not total CPU or service latency measurements.

There were 6,219 shared fitting calls across the two splits, with measured mean
349.2 microseconds per call during the replay. Each count model occupies about
2 MiB; base plus short/local/pooled versions are bounded but expensive relative
to this nine-bit toy. Sharing fitted immutable versions across arms saves test
duplication, not a claim that deployment can serve every branch for free.

The harness fits synchronously only to make next-step publication deterministic.
Serving would need background execution, dependency checks and a prospective
publication gate. No daemon service, database, production OpenClaw, or private
session corpus was changed or exercised here. v2 remains a failed historical
policy; this fresh-seed v3 pass does not rewrite that result.

Three serial Apple M4 / one-CPU steady microbenchmarks measured preserved MMM
prediction plus mixture feedback at 8.70--8.79 microseconds versus instrumented
frozen MMM at 8.77--9.18 and breadth mixture at 10.43--10.49. These small timings
do not establish a speedup. Each 256-sample fit averaged 0.333--0.335 ms and
allocated 2 MiB. The steady benchmark excludes the diagnostic reference reads,
AP nomination/evidence monitor, audit acquisition, refitting, queueing, database
and concurrency. It is not directly comparable to v2's monitor-inclusive timing,
and neither benchmark establishes end-to-end p99 or a production deadline.
