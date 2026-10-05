# Guarded segmentation: gains, limits, and an oracle falsifier

Status: research screen, not promotion. All seven goals remain open.
The source is the already-consumed independent-v1 cohort: 672 trajectories,
256 forecasts each, eight indices, both phases and both feedback schedules.
See the [frozen contract](mmm-segment-guard-v1-contract.md).

## Measured comparison

Lower expected binary Brier is better. Harms count non-overlapping 32-frame
windows whose expected loss exceeds Markov by more than .01 (1e-12 numerical
slack), out of 5,376 windows. Zero such harms does not mean zero regression.

| Arm | Whole Brier | Terminal 64 | Harm windows |
| --- | ---: | ---: | ---: |
| Segment, pointwise guard | .157227373 | .145340987 | 0 |
| Static, pointwise guard | .157562187 | .145459802 | 0 |
| Markov | .157516545 | .145520129 | reference |
| Segment, local ledger | .156743582 | .144783147 | 17 |
| Static, local ledger | .157600524 | .145354393 | 17 |
| Raw segment | .158618743 | .146867537 | 690 |
| Raw static | .160344189 | .147972831 | 446 |

Within the pointwise guard, segment minus static whole Brier is -.000334814,
with descriptive pointwise 95% bootstrap interval [-.000398052,-.000275015].
However, the stationary difference is +.000368115
[.000303890,.000425758], while the changing difference is -.001477073
[-.001584978,-.001370320]. The local ledger has the same directional tradeoff.
Intervals resample eight index clusters, not individual frames; they are not
simultaneous guarantees, confidence sequences, or fresh confirmation.

Only 4 of 128 changing-cell terminal comparisons meet the original .005 mean
gain threshold. Those 128 comparisons cross four arm/reference pairs; they
are not 128 independent trajectories. This does not run or satisfy the full
original 32-index, 768-requirement protocol.

## Verification and compute

- Screen replay is byte-identical, including after extracting the unchanged
  proposal/guard helper into `research/guarded-head.mjs`.
- Independent direct-probability reconstruction covers index 0, clocks
  0/128/224, every phase/case/schedule: 84 rows, 252 fitted states, and
  8,064 predictions per head. Maximum segment prediction error is 9.54e-14;
  maximum log-evidence error is 1.92e-13. Perturbed weights and forecasts
  are rejected. This is stratified reconstruction, not a full fitter replay.
- Warm combination of the delayed proposal and BOTH guards takes
  1.34-1.40 microseconds per forecast amortized. This excludes fitting, I/O,
  loaded serving, and inference queue delay; the reference filter is O(T^2).
- All source-model collection cost remains charged: 545.32 seconds for the
  original 15-head source collection. No new per-fitter or serving benchmark
  was performed here.

## Oracle feasibility audit

After inspecting the screen, an explicitly post-outcome diagnostic asked
whether better weighting can meet the recovery target within the fixed guard.
It uses generator probability Q only for evaluation. No oracle output enters
the learner, source tapes, runtime, or a validation claim.

For baseline b, candidate c, and maximum admissible weight u from the existing
pointwise guard, the optimal conditional-risk weight is
`clip((Q-b)/(c-b), 0, u)`; equal forecasts use weight zero. The noise term
`Q*(1-Q)` cancels when computing gain. The audit also lets the oracle choose
either head independently each frame.

The stronger diagnostic removes the head restriction entirely. Both endpoint
constraints on excess binary Brier are equivalent to the interval

```text
lo = max(0, 1 - sqrt((1-b)^2 + .01))
hi = min(1, sqrt(b^2 + .01))
p_oracle = clip(Q, lo, hi)
```

Indeed the outcome-zero constraint is `p^2-b^2 <= .01`; the outcome-one
constraint is `(1-p)^2-(1-b)^2 <= .01`. Their intersection is exactly that
interval, and squared distance to Q is minimized by projection onto it.
Thus this is an upper bound for ANY binary forecast satisfying this particular
pointwise guard relative to the fixed Markov tape, not just a selected mixer.

| Oracle class | Whole mean gain vs Markov | Changing terminal cells below .005 |
| --- | ---: | ---: |
| Segment head | .001883964 | 31 / 32 |
| Static head | .000704716 | 32 / 32 |
| Either head | .002096266 | 31 / 32 |
| Any admissible binary forecast | .003185839 | 26 / 32 |

The universal diagnostic checks 10,838,016 grid competitors, endpoint
admissibility, dominance over both head oracles, all 672 unique trajectory
indices, and the existing 15,972 guard tests. Its replay is byte-identical.
These are exact conditional-risk optimization checks up to numerical slack,
not statistical evidence that generator Q is observable or transferable.

This rules out attaining all measured cell targets by changing only weights
under the current pointwise guard. It does not prove an impossibility for
other baselines, other cohorts, the original expected-risk criteria, or other
guards. The original goal did not require this stronger per-outcome bound.

## Avoid repeating an earlier failed lead

The proposed equal-prior no-change/segment evidence mixture was already
tested in [v120](mmm-soft-learners-v120-results.md), including a retained-window
weight lower-bound audit. It failed the broad gates. Reimplementing that exact
mixture would repeat an existing result, not produce a new rescue.

[Wilson, Nassar and Gold (2010)](https://pmc.ncbi.nlm.nih.gov/articles/PMC2966286/),
sections 2-3, explicitly model hazard uncertainty. This supports considering
uncertain change rates, but supplies no guarantee for our guarded, delayed,
bounded-history Brier experiment. The older expert-switch hazard mixture also
failed; it is distinct from segment parameter sharing and must remain a
negative control rather than be forgotten.

Next investigate risk control that matches the original statistical non-harm
requirement, with valid delayed-evidence uncertainty, rather than another
weight search under an empirically unattainable pointwise restriction. First
check existing interval/credit diagnostics for duplicates. No threshold
relaxation, case removal, or fresh-confirmation relabeling is justified.

## Artifacts and replay

The `mmm-segment-guard-v1-*` screen, audit, summary, benchmark, oracle, and
replay JSON files retain all results. The first oracle JSON contains only
the three head-restricted arms; the universal artifact supersedes its scope
without deleting it. Current script reproduces the universal artifact:

```sh
node research/segment-guard-oracle.mjs docs/experiments/mmm-spike-independent-v1-source.jsonl /tmp/segment-guard-oracle-new.json
cmp /tmp/segment-guard-oracle-new.json docs/experiments/mmm-segment-guard-v1-oracle-universal.json
```

The output path must not exist. SHA256:

```text
script: ee9d48b9959849c101988fd8d5c8241454d14d0f1346e14420087125b092c6a5
screen: fdbc9fb2d8b3469ab7889b3d467862573714c9c97a02d200dab7d92dd33755f7
universal oracle: 52f2d73d57f86f9f159e5bebb132054a22ed5ed7e9a16797523c5e8534b55646
```

No production, Go runtime, whitepaper, commit, or push changes in this round.
