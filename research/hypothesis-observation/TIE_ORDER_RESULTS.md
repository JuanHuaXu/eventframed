# Observation order audit

The existing ordered-history selectors are not action-equivalent to the
count-state selectors, even where their costs agree numerically.

Protocol: [frozen diagnostic](TIE_ORDER_PROTOCOL.md).
Artifact: [results and source hashes](tie-order-audit.json).

Verification: a complete independent process replay matched the artifact byte
for byte. This verifies reproducibility, not independent model correctness.

Across 752 states (16 roots times 47 count vectors), reversing the renewal
order changed the selected action as follows:

| Selector | Order changes | Canonical vs count disagreements | Reversed vs count disagreements | Maximum canonical cost excess |
| --- | ---: | ---: | ---: | ---: |
| One-step | 2 | 2 | 2 | 1.11e-16 |
| Two-step | 2 | 1 | 1 | 2.22e-16 |
| Entropy | 11 | 3 | 14 | 2.22e-16 |

All 3760 posterior/evidence-mass comparisons agree within 2.23e-16. No
selected-action excess exceeds the predeclared diagnostic tolerance 1e-12.
This is consistent with numerical near-ties, not a substantive posterior
exchangeability failure. It is not a symbolic proof that every tie is exact.
The scope includes every count state through two renewals and two selected
longer vectors, not every six-renewal ordered history.

Root cause supported by the diagnostic: strict floating-point argmin depends
on report multiplication/summation order and, between implementations, grouping
of the planning arithmetic. Nearly equal objective values do not imply identical
action choices. No frozen implementation was changed.

Consequences:

- The exact count-policy population counterexamples remain valid for those
  policies. They must not be described as exact replays of the older policies.
- Model-prior indifference does not establish indifference under each external
  source regime. This audit does not quantify the external risk effect of ties.
- Next controlled experiment should explicitly freeze canonical count-state
  evaluation and a numerical tie rule, for every candidate and control. Compare
  exact per-regime risks against the archived strict-argmin policies before
  attributing a benefit to a source-prior or objective change.
- A tolerance-based tie rule alone is not a demonstrated quality rescue. No
  completion, production promotion, or new latency claim follows.

Reproduce with `node research/hypothesis-observation/tie-order-audit.mjs`.
The stdout artifact is deterministic; omit an output path when replaying to
avoid overwriting the archived JSON.
