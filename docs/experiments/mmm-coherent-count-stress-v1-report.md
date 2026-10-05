# Coherent count stress: reject unconditional replacement

The unchanged coherent-count candidate fails its intended robustness check.
384 fresh training fits at noise.25 and.5, two input laws, two nominal targets,
three sample sizes and four fixed masks. At noise.5 the target names label
independent null replications, not distinct learnable rules. All results retained.

Examples of exact population Brier (lower is better):

| Noise | Input | Target label | n | Mask | Old | Coherent |
|---|---|---|---:|---:|---:|---:|
| .25 | Uniform | Bit |64|31|.228000|.279369|
| .25 | Uniform | Parity4 |64|31|.228211|.274497|
| .50 | Uniform | Parity4/null |64|31|.276029|.366596|
| .50 | Dependent | Bit/null |128|511|.261501|.338128|
| .50 | Uniform | Parity4/null |4096|511|.270753|.286821|

The null prior's Bayes risk is .25. The candidate's lower fine-context smoothing
learns accidental label imbalance too strongly. Mathematical coherence and
the earlier low-noise gains remain true, but neither justifies integration.

## Exact Explanation

Condition on m training inputs in a fixed context with independent fair labels.
For K~Binomial(m,1/2) and symmetric smoothing mass a:

    p = (K+a/2)/(m+a)
    E[(p-1/2)^2] = m/[4(m+a)^2]

This is excess expected Brier above .25. Old per-mask smoothing uses a=2.
The coherent uniform full-cell prior gives a=2*2^(-d) after observing d bits.
For every m>0,d>0 the latter has strictly greater conditional null variance.
Thus this regression is not merely an unlucky seed. Literal binomial sums
confirm the identity on650 combinations, maximum discrepancy2.776e-17.
The statement concerns fixed contexts with independent null labels; it is not
a theorem about adaptive acquisition or arbitrary real-world streams.

## Decision and Next Boundary

Do not promote total-mass2 coherent counting or tune its concentration on these
consumed fixtures. A stronger prior would trade null protection against the
earlier low-noise gains; coherence alone cannot select that tradeoff.
Any new candidate needs structural pooling or separately validated model
comparison, not simply redistributing the old pseudo-counts. Existing subset
and hierarchical-model experiments must be inspected before another proposal.
Retain the old estimator and all failed artifacts. No serving path changed.

## Verification

- Race-enabled stress PASS:1.28s experiment,2.732s package; vet PASS.
- All384 unique fit seeds, cells, four masks, finite expected risks, null/noise
  floors and exact root parity checked by the summary.
- Summary with eight captured source files replays byte-identically; capture
  occurs after collection and does not establish independent preregistration.
- No additional performance campaign for a quality-rejected candidate. Prior
  component72-74us fitting measurements remain only component measurements.

Raw:`mmm-coherent-count-stress-v1.json`.
SHA256:`7de597256a093d39a6433a863bc45149caef4b4b492cafe521c2fcc2af3a4ffe`.
Contract:`mmm-coherent-count-stress-v1-contract.md`.
Identity:`mmm-coherent-count-null-identity-v1.json` and
`research/coherent-count-null-identity.mjs`.
All seven goals remain open. No production, whitepaper or remote changes.
