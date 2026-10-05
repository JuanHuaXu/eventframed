# R-IDeA primary-source review and direction-aware alternative

## Source Inspection

Tang, Sloman and Kaski (2026),
[Representative, Informative, and De-Amplifying](https://proceedings.mlr.press/v300/tang26d.html).
Read sections3-4 and AppendicesC-D; visually verified printed PDF pages6,7,20.
The official download succeeded via ordinary HTTP retrieval after the browsing
tool rejected its application/octet-stream content type. No access challenge
was bypassed. PDF SHA256:
`f81df76828de572f322cee3231f859409de0a1c802f03bd9e7c40345e54024c5`.
The paper is not copied into the distributable repository.

Equation10 multiplies information gain by a factor involving the ratio of
post-addition to current MMD. Equations11-12 fit an observation-matching but
disagreeing proxy and weight acquisition by a sigmoid of disagreement. Theorem1
assumes finite models, bounded outcomes, empirical risk minimization and
training support. Its transfer to Bayesian prediction is motivational, not a
direct theorem for our learner. Theorem2 claims an absolute-disagreement region
lies inside a signed-product region. The proxy lemma only propagates a uniform
approximation assumption; the proxy fitting objective does not prove it.

## Confirmed Limitation of the Stated Theorem2

Let m be the best-in-class predictor, h the learned predictor and t the true
regression function. The claimed antecedent is
|h-m| >= tau/B + cB, where B=sup|m-t|, c>=2 and tau>=0.
The claimed conclusion is (h-m)(m-t)>=tau.

Two finite bounded examples disprove that implication as stated:

1. One input, F={m=.4,h=.8}, t=.5. Then m is uniquely best in F,
   B=.1. With tau=.01,c=2, the threshold is.3 and |h-m|=.4 passes.
   But (h-m)(m-t)=-.04, below.01. Under Bernoulli(.5) observations,
   a sample of ones makes h the empirical risk minimizer with positive
   probability. Both input distributions can put probability1 on this input.
2. Two equiprobable inputs, t=(.5,.5), m=(.501,.6), h=(.9,.6).
   F again contains just m,h, with m uniquely best; B=.1. At the first
   input, the same threshold.3 is passed by gap.399, but the product is
   .000399, below.01 even though its sign is positive. A sample of ones
   at both inputs again makes h the empirical minimizer.

The first example shows missing sign information. The second shows why a
global upper bias bound cannot replace the needed local magnitude information.
AppendixC.1 relation(iii) reasons from a signed product toward a weaker absolute
condition, but that does not prove the displayed reverse subset implication.
We verified the signs visually rather than relying on extracted PDF characters.

`research/deamplification-source-check.mjs` checks both examples with integer
arithmetic in BigInt, including best-in-class risk and empirical minimizers.
A separate finite grid has218 qualifying triples and140 violations. Full replay
is exact. Artifacts: `mmm-deamplification-source-check-v1.json` and `-replay.json`.
This does not refute the paper's empirical comparisons or imply the acquisition
heuristic cannot work. Do not inherit Theorem2 as an EventFrame certificate.

## A Direct, Conditional Brier Bound

Our alternative avoids identifying the best-in-class function. Let b be the
baseline forecast, f0/f1 its corrected forecasts for query answers0/1, p the
true query-answer probability, and q0/q1 the true target probabilities
conditional on each query answer. The expected Brier improvement is exactly

```math
G(p,q_0,q_1)
=(1-p)(b-f_0)(b+f_0-2q_0)
+p(b-f_1)(b+f_1-2q_1).
```

This expression is affine in each uncertain probability separately. Over
rectangular intervals for p,q0,q1 its minimum and maximum occur at the eight
corners. Thus valid simultaneous intervals give an exact worst-case bound over
their rectangular envelope. The envelope may be conservative if probabilities
are dependent; the calculation does not assume query/target independence.
For weighted target inputs, the weighted sum of pairwise lower bounds remains
a valid lower bound when the joint coverage event holds.

`query-directional-bound.mjs` implements this calculation, not interval
estimation. Tests use independent direct Brier evaluation:46875 interior points
across375 boxes, exact corner extrema, relabeling, ownership, invalid input and
a genuinely dependent-target example pass. See
`mmm-query-directional-bound-v1-contracts.json`.

No evidence coverage is created by this algebra. Model posterior intervals
cannot certify themselves against misspecification. Repeated candidate choices
require simultaneous/time-valid coverage; as-of availability, source dependence,
support and acquisition costs remain explicit unresolved contracts. Do not
enable a production gate based on this calculation alone.

## Next Experiment

Test the independently usable representativeness component: a frozen MMD-based
candidate factor using only observed input support and visible target inputs,
with unchanged acquisition budgets and actual publication. Specify the kernel,
MMD versus squared-MMD convention, zero-denominator behavior and signed factor
before scoring. This is an R-I-inspired Bernoulli adaptation, not R-IDeA or an
inherited risk guarantee. A proxy-disagreement experiment remains possible as
a heuristic, but requires its own controls and cannot bypass Anti-Pigeon.

All seven goals remain open. No empirical rescue, serving performance success,
production change, whitepaper promotion or author contact occurs in this review.
