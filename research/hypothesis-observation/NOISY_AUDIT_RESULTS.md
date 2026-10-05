# Equal-cost noisy source audit: greedy pilot fails

[Protocol](NOISY_AUDIT_PROTOCOL.md), [full model and scores](noisy-audit-evaluation.json),
[verification](noisy-audit-verification.json).

All arms receive the same mandatory source0-copy audit with assumed sensitivity
and specificity.8. It costs2 credits, leaving five renewals instead of six.
Total cost remains16 credits. The audit is a hypothetical observation channel,
not an implemented provenance assessor or a measured80% real-world reliability.
It is noisy evidence about source behavior, not a target label or actual mask.

## Findings

The one-step target-Brier selector fails the full screen:

| Control | Final nonharm /80 | Positive area /75 | Worst final gain | Worst area gain |
| --- | ---: | ---: | ---: | ---: |
| Random renewals | 80 | 72 | -.0023382996 | -.0048670414 |
| Report entropy | 72 | 33 | -.0164269590 | -.0047236191 |

There are152/160 final passes and105/150 area passes. Eight final comparisons
against entropy exceed the.01 harm allowance. Source masks2/6 fail at noise.25;
2/3/6/7/10/14 fail at.30. Maximum final harm is .0164269590. These are exact
finite population scores from ordinary numerical arithmetic, not sample rates.

The result rejects this particular one-step pilot. It does not show that the
new information model is infeasible: no optimal joint-policy bound has yet been
computed for it. Nor does it measure the causal benefit of an audit in isolation;
every current arm has an audit and sacrifices a renewal. The old no-audit
screen is a different information contract, not a same-arm matched ablation.

## Integration and checks

The model conditions its joint evidence on the audit before normalization.
Both audit outcomes are integrated in evaluation. The six pre-action losses
include the pre-audit forecast and five pre-renewal forecasts. Audit information
does not enter the pre-audit forecast. It is charged even when unhelpful.

There are41184 compiled states and1440 probability normalization checks.
Verification includes2816 posterior marginalization/neutral-channel checks,
1408 independently integrated quadrature forecasts (maximum error3.34e-16),
36 separate backward-score comparisons (maximum error5.22e-15), and full
byte-exact replay. A reliability.5 audit leaves the posterior unchanged, and
marginalizing the two outcomes recovers the original evidence law.

## Next decision

Determine whether a joint-policy oracle can exploit this channel before trying
more greedy heuristics or assuming provenance accuracy solves the problem.
The pre-audit state must have one common forecast across both future audit
outcomes; conditioning it on the later result would leak future information.
Any optimality check must account for the mandatory audit cost and both outcomes.
Real provenance reliability remains unvalidated. All seven whole research
directions stay open; no production, paper or archived implementation changes.
