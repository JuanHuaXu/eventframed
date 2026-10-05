# Observation order and numerical ties

Diagnostic, not a fresh quality confirmation. Freeze all existing models and
selectors. Enumerate all 16 root patterns and all eight-coordinate renewal
count vectors of total at most two, plus two longer mixed count vectors.
Compare canonical order and reversed renewal order (renumbering per-type slots).
Compare one-step, two-step and entropy choices with the count-state selector.

Record posterior/mass discrepancies, action disagreement counts, and regret of
each selected action under the canonical count-state objective. A disagreement
with canonical cost excess above 1e-12 falsifies the numerical-tie explanation
at this diagnostic tolerance. This tolerance classifies evidence only; it does
not modify any selector or claim exact symbolic equality. Also count action
changes caused solely by reversing the same observed multiset.

Do not infer identical rollout laws from approximately equal action costs:
different tied actions may have different risks under an external regime.
No runtime, source model, production configuration or archived result changes.
