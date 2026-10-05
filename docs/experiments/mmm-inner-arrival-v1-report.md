# Inner arrival updates: small additional benefit

## Result

Updating the inner selector from arrived issued advice improves mean post-change
Brier by .00017-.00241 beyond outer-only arrival learning in all eight changed
case/cohort cells. The main improvement therefore remains the outer policy.
Both cohorts are consumed; no fresh confirmation or significance claim.

Second-cohort delayed means:

| Case | Original | Outer only | Outer and inner | Full-arm mean foreground cost |
| --- | --- | --- | --- | --- |
| Copied bit | .279590 | .258504 | .257142 | 4.659 |
| Copied XOR2 | .279574 | .259656 | .258355 | 4.532 |
| Noisy-copy bit | .280556 | .259164 | .258637 | 4.632 |
| Reversing XOR2 | .273903 | .237662 | .237491 | 4.521 |
| Stable | .047330 | .047629 | .047633 | 4.003 |
| Null | .255241 | .253748 | .253581 | 5.999 |

Brier uses the last256 frames; cost is over512. The first three changed cases
still exceed the neutral forecast's .25 Brier. Stable harm persists in the second
cohort. Original endpoints and every immediate forecast remain unchanged.

## Mechanism and checks

The new fourth arm retains its own inner and outer selector weights. It updates
the inner only when the issued journal record says that expert was available,
using the exact PRE-outcome inner advice rather than regenerating a prediction.
Arrivals precede publication, and labels missing from the stream never update.
External structural revocation caps the long outer slot and rejects outstanding
pre-split credits. Experimental role weights are not claimed as model-specific
Bayesian posteriors. Existing generation-scoped journals remain bookkeeping,
not the source of the separate arrival selectors' update counts.

[Contract](mmm-inner-arrival-v1-contract.md),
[raw artifact](mmm-inner-arrival-v1.jsonl),
[summary](mmm-inner-arrival-v1-summary.json).
SHA-256:
`b97d825011f9abab815f3150816df5be613dc5f119c8e771c82b654bd941bd4d`.

192trajectories/384paired schedules. Collection checks all three old endpoints
exactly: scores, split times, forecasts, masks, inputs, outcomes, delay/missing,
audits and fitted-origin sets. Independent evaluator reconstructs all four
arms' scores/costs and matches the fixed-view endpoint to earlier JS.865 captured
source snapshots verified; repeated summary byte-identical. The reported
meanChangedMasks field continues to describe the OUTER-ONLY arm, not the full
arm. All full-arm masks are available in the raw artifact.

Targeted race contracts2.563s, vetPASS, collection58.23s. Full Go experiment was
not rerun; endpoint checks and independent score reconstruction are the evidence.
Benchmark file added after source capture. Whole512-frame four-arm fixtures:
immediate183.28/152.48/153.90ms,71.55MB;
delayed131.48/131.93/130.70ms,63.12MB. Includes fits/observations/flush, excludes
initial base fitting. These are not serving latency or per-arm incremental cost.

## Decision

Inner credit is not the source of a reversal of the apparent gain in these
generators. Preserve the modest component benefit, but do not infer robustness
from it. The next test must cover the broader established dependence suite
(noise, appearing/disappearing/reversing dependencies) and earlier switch
families. Keep original quality/cost criteria and old v104 failures intact.
No parameter sweep or deployment justified. All seven goals remain open;
production, whitepaper and remotes untouched.
