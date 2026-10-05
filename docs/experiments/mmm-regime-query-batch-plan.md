# Exact batched regime-query plan

Frozen before implementation or timing. The reference implementation is correct
but evaluates an eight-query pool using16 conditional refits, about0.65s excluding
the base fit. This plan changes computation, not the model or research goal.
No speedup, efficacy or error-control claim is yet established.

## Shared interval moments

Use the same sanitized history positions0..T-1, same retained S of at most63
known labels, and the same.01 hazard/.95 generic model. A query q identifies an
unobserved position in this history. Let M(a,b) be the existing integrated
likelihood of retained labels within the half-open time interval[a,b).

For every relevant interval, compute posterior mean mu(a,b,x) and joint
Bernoulli-parameter moment nu(a,b,xq,x) from its SAME family/hypothesis posterior.
Do not substitute prior family weights or multiply separately averaged means.
The query is an unobserved outcome; no true query Y is used in these moments.

Within a generic mask, different cells have independent Beta parameters. Equal
cells add Var(theta)=p(1-p)/(n_cell+2) to the product of posterior means. Within a
Boolean mask, all inputs share the agreement parameter: add s_q*s_x*Var(theta),
where s is+1 for odd parity and-1 for even parity. Finally average these moments
using the posterior weights of masks AND families. Empty intervals use the
actual prior moments, not independence by fiat (same-cell prior variance1/8).

## Forward identity

Set pi(a,b)=(a==0 ? 1 : h)*(1-h)^(b-a-1), for a<b. The ordinary prefix sum is
F(0)=1 and F(b)=sum_{a<b} F(a)*pi(a,b)*M(a,b).

For each query q, construct a positive-outcome tilted prefix Fq, with Fq(0)=1.
In the same recurrence, if a<=q<b, use F(a)*mu(a,b,xq); otherwise use Fq(a).
Multiply that term by pi(a,b)*M(a,b). Before q is encountered, Fq(b)=F(b).
Then m=P(Yq=1|S)=Fq(T)/F(T). Use log-domain sums for numerical stability.

The unnormalized joint numerator for query success and a virtual probe in the
last segment is a sum over its start a. If a<=q, use
F(a)*pi(a,T)*M(a,T)*nu(a,T,xq,x). Otherwise use
Fq(a)*pi(a,T)*M(a,T)*mu(a,T,x). Divide the sum by F(T), obtaining J(x).

For a target at clock+1, two unobserved hazard transitions follow the last
historical frame. Let lambda=(1-h)^2. The joint query/target-success probability
is (1-lambda)*m/2+lambda*J(x). The ordinary probe law p0(x) uses the analogous
base tail mean. Derive p1(x)=joint/m and p0conditional(x)=(p0(x)-joint)/(1-m).
These must match the reference's two explicit conditional refits and its extra
hazard propagation. Compute the existing V(q) without changing the objective.

All candidate queries share the base interval calculation. Time intervals with
the same retained-label endpoints can share M/moment computations. Do not merge
query origins themselves: different positions in an unobserved gap still have
different boundary/survival weights.

## Required gates

1. Tiny exhaustive partition and same/different-cell moment tests, including
   empty tails and queries before/after the last observed label.
2. Compare all query masses, conditional probe laws and values against the slow
   refit reference at full63-label support, with multiple histories, up to eight
   queries/probes, gaps, both regime directions and stationary replay fixtures.
   Absolute tolerance1e-10; no clipping or renormalizing mismatches into agreement.
3. Preserve erased outcomes, as-of nomination, cancellation, capacity, input
   ownership and concurrent-read contracts under race. Cross-check query order
   under ties rather than relying solely on close utility means.
4. Paired isolated timing at identical support/history/query/probe counts. The
   cold batch must include its base evidence work and must not use a free slow
   base fit. Compare with the reference's base plus16 refits. Three repetitions;
   require at least4x lower median time and at most half the allocation volume.
   Preserve failures; no post-hoc threshold relaxation. This is not a serving
   deadline or whole-goal success criterion.

The prototype should remain a research-only function. Do not remove the slower
oracle/reference, alter existing likelihood builders, retune priors or change
quality gates. A later whole-stream test must still establish equal-outcome-cost
learning benefits against random and entropy acquisition, with all21 cases,
delayed/missing outcomes and actual posterior refitting after paid evidence.
