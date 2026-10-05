# Copied-source ties are structural

[Protocol](COPY_TRIGGER_PROTOCOL.md), [audit](copy-trigger-audit.json),
[independent path enumeration](copy-trigger-verification.json).

The hypothesis is supported for the frozen risk-budget policy: all32 reachable
states where it differs from normalized entropy contain a source0 renewal
opposite that source's initial report. Eight differences occur after four
renewals and24 after five. No earlier difference is reachable. No counterexample
to the proposed trigger condition was found in the exhaustive finite-state audit.
This describes an emergent policy property, not an explicit source0 conditional
in the runtime or a recommended trigger to add.

Under each odd copy mask (source0 copied), both policies choose the same action
at every reachable state, and all seven occupancy maps are identical. Their
complete ordered action/outcome trace sets also match under separate enumeration.
Thus terminal and area scores are equal for this family for any common noise
strictly between0 and.5, not just the eleven previously evaluated noise points.
This follows from identical policy behavior and the same frozen forecasts,
rather than a floating-point lower-bound sign. It is not a guarantee under a
different source mechanism or observation contract.

As a negative control, all even masks produce different occupancy maps at
step5 (eight root patterns) and step6 (all16 root patterns). These are counts of
root patterns with unequal maps, not probabilities or independent experiments.

## Verification

The main audit checked1792 root/mask/layer combinations and20768 integer count
values. Counts remain exactly representable safe integers. Full byte-exact
replay passes. A separate verifier enumerated32480 individual ordered prefixes,
reconstructed all112 aggregated mask/layer mismatch counts, and verified128
root/odd-mask complete trace equalities. It does not use the main audit's merged
occupancy propagation. Frozen model/compiler hashes and statistics are checked.

## What this changes

More samples or finer noise grids cannot turn these exact ties into gains for
the frozen policy. Adding more lookahead to the existing prior objective also
does not address the diagnosed trigger condition. The guard's permitted changes
only become effective after observing evidence that rules out copying source0;
an actually copied source can never provide that contradiction.

This is not a universal impossibility of useful acquisition when source0 is
copied. The conditional guard is stronger than the original population-risk
requirement. Next test global allocation of the allowed risk across histories,
with explicit population accounting and no need for a copied source to contradict
itself. Retain copied-source masks and normalized entropy as controls. Do not
silently drop plausible regimes to manufacture an improvement.

All seven research directions stay open. No runtime, production, paper or
archived policy changes were made.
