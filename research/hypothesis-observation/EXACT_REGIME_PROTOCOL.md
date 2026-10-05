# Frozen exact regime evaluation of compiled policies

Use count-planning-exact.json unchanged. Evaluate its six compiled policies
under each of16 true copy masks and noise .10,.15,.20,.25,.30 (80 regimes).
No optimization, changed prior, new action rule or fitted parameter. Actual
regime is visible ONLY to the scorer, never policy selection.

These are compiled count-state policies. Floating-point ties can differ from
earlier ordered-history rollout choices; do not call this an exact expectation
of those distinct implementations without proving tie equivalence.

For each root/policy, propagate path multiplicities using action weights only:
deterministic weight1 or random1/4; both hypothetical outcomes lead to children.
Summing parent multiplicities into a count-state handles multiple report orders.
Multiply by that state's actual joint class/history likelihood when scoring.
This avoids adding multinomial factors to individual history likelihoods twice.
Normalize the total probability at every paid-query stage.

Record final Brier, post-query risk sum, original pre-query learning-area risk,
accuracy and confidently-wrong rate. Compare full and two-query policies against
random and entropy for exact population final nonharm (gain>=-.01), and positive
area gain where mask!=15. This is a diagnostic against the same substantive
inequalities, not a new empirical confidence screen. Preserve rollout failures.

Validate actual likelihoods against ordered direct products and use separate
backward conditional-probability evaluation for representative regimes. Replay
the full exact table. Separate finite population harms from uncertainty-only
sample failures. Grid evaluation does not establish continuous coverage.

The compiled optimum remains model-prior optimal, not necessarily regimewise
optimal. No real-source authenticity, new private data, benchmark claim,
production change, whitepaper promotion or publication.

