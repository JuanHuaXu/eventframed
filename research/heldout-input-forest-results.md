# Held-out forest component: verified, quality untested

Research-only implementation in `internal/observationlearners/chow_forest_test.go`.
The first chronological half fits a smoothed input tree; the second chooses
among uniform plus nine nested forest laws using input likelihood. No outcome
labels or refit on validation. This pays the half-sample estimation cost rather
than silently giving the candidate additional data.

Motivation and adaptation boundary:
https://jmlr.org/papers/volume12/liu11a/liu11a.pdf
The paper supports held-out forest selection as a research technique; it does
not establish accuracy or generalization for this smoothed discrete component.
Earlier outcome-family evidence failures remain relevant cautions but are not
the same algorithm. Pairwise pruning still cannot capture pure XOR dependence.

## Checks

Race suite PASS1.432s, vet PASS. Every candidate positive and normalized;
uniform and copied-field full-support fixtures select zero dependencies and
one dependency respectively. Selected score matches literal held-out law
likelihood, all outcome labels can flip without affecting output, inconsistent
copy validation changes likelihood, and invalid sample shapes are rejected.

An initial test wrongly demanded that a uniform model's likelihood change
when validation inputs change. Its failure was in the test expectation, not
the selector. The corrected negative control explicitly requires invariance
under uniformity and sensitivity for conflicting copied-field evidence.

## Component Cost

Apple M4/darwin arm64, fit/select on64samples, three300ms repetitions:

    20853 / 20764 / 20748 ns/op
    10184 B/op, 9 allocations/op

About20.8us, excluding outcome fitting, conditional-table compilation and all
daemon serving costs. No daemon performance or quality rescue claim.

## Next

Fresh frozen comparison: uniform, histogram, full tree and held-out forest,
same overall data budget; retain null/high-noise/higher-order cases. No split
ratio tuning or treating selected validation likelihood as a confidence bound.
Quality remains untested; all seven goals remain open.

SHA256:

    source 129ee1adca88bfdb4c02f3a818a908fdfd880f112944c0b9bd0c3a2f8542f4bf
    contract f8b32527701175dde5da6aca6618eb3ff82a7b58c5c2400128a346333d1c303e

No production, remote, whitepaper, commit or push changes.
