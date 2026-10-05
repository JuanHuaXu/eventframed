# Interval-aware evidence acquisition: random beaten, full screen fails

Under the [frozen protocol](INTERVAL_ACQUISITION_PROTOCOL.md),1536 fresh-seed
episodes compare target-class one-step Gini selection, uniform random,
next-report entropy, and a fixed diagnostic schedule. Every arm uses the same
joint noise/source inference and spends16 evidence credits: four initial
reports and six renewals. The fixed-observation900-gate failure is NOT changed.

Overall19/84 checks pass (FAIL):16/48 final-loss nonharm checks and3/36
learning-area gain checks. The claim of protected faster learning against
both random and uncertainty sampling across every tested cell is not validated.

## Pooled descriptive results

Lower Brier and learning-area Brier are better. Area covers the12 post-initial
credits, not the common first4. The following aggregates pool the declared
noise/mask/split mixture; they do not replace the per-cell acceptance gates.

| Policy | Final Brier | Learning-area Brier | Accuracy | Confidently wrong |
| --- | ---: | ---: | ---: | ---: |
| Target Gini | .333168 | .399890 | 77.21% | 20/1536 |
| Random | .405275 | .439408 | 70.38% | 1/1536 |
| Report entropy | .332866 | .404401 | 76.95% | 13/1536 |
| Fixed schedule | .384336 | .427590 | 73.11% | 0/1536 |

Target-focused observation clearly improves the pooled point estimates over
random. Against entropy, final Brier is slightly worse (.000302 difference),
while learning-area Brier is slightly better (.004511). That is not a broad
claim of superiority or calibrated real-world confidence. The higher count of
confidently wrong predictions must be retained alongside higher accuracy; the
count alone does not distinguish increased confidence from miscalibration.

## Heterogeneity and uncertainty

With all genuine sources, pooled target final Brier is .224280 versus entropy
.239810 and random .355789. With mask5 (types0 and2 copied), target .261523
is worse than entropy .245168, though better than random .375293.
With all renewals copied, all four final Briers are close (.464711-.467540);
no new independent draws are available in that condition.

Of65 failed gates,14 fail even their point-mean threshold (six final nonharm,
eight area gains);51 have acceptable point means but insufficient lower bounds.
Thus neither 'just collect more samples' nor 'the method is uniformly harmful'
is supported. The paired mean +/-3.3 SE intervals are descriptive normal
approximations, not simultaneous coverage, confidence sequences or exact tests.

## Verification

Full1536-episode replay is byte-exact. All6144 arm costs and12288 loss/area
recalculations pass. All17224 copied-report checks match the declared root.
For two episodes per cell,864 deterministic choices and1152 forecast prefixes
are reconstructed from observed history alone, without hidden-world fields.
Random-policy traces are covered by full replay and budget/slot checks.

Independent direct quadrature agrees with448 joint-model forecast values
within8.89e-16;192 cached/uncached choice checks agree without history mutation.
These are finite model and implementation checks, not an independent generator
validation or proof of real-world source authenticity.

[Full traces and gates](interval-acquisition-experiment.json),
[verification](interval-acquisition-verification.json),
[observer](interval-acquisition.mjs).

## Next useful research

Investigate whether the one-step target criterion neglects observations that
clarify source reliability and become useful over subsequent queries. A bounded
multi-step criterion is a live lead, but must preserve equal acquisition costs,
as-of information boundaries and strong entropy controls on new seeds.
Retain every current failure; do not remove mask5 or silently change priors.

The measurement-model interval, finite hypothesis family and source-mask model
remain assumed. Acquisition credits exclude planner CPU; this run makes no
serving-latency or compute-efficiency claim. No production/private-data access,
whitepaper promotion or publication. All seven directions remain open.

