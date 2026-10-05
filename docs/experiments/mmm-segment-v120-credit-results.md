# Feedback and expert-credit diagnosis

Consumed v120 data only. These studies do not create an untouched confirmation,
finish a research direction, change serving defaults, or promote the whitepaper.

## Is delayed feedback the whole problem?

The [credit diagnostic](mmm-segment-v120-credit-diagnostic.json) retains all2688
schedule runs. It holds each issued expert forecast fixed and compares the
segment-augmented Markov mixer under actual feedback, immediate nonmissing
feedback, and immediate complete feedback. The latter two are counterfactual
diagnostics, not permissible forecasts under the actual observation schedule.
They do not retrain the expert forecasts under that counterfactual schedule.

For each frame let p0,p1,p2 be those three forecasts, q the teacher law, and
h=clip(q,min_i p_i,max_i p_i) the oracle convex-hull projection. For expected
Brier L(p)=(p-q)^2+q(1-q), the exact decomposition is

`L(p0)-q(1-q) = [L(p0)-L(p1)] + [L(p1)-L(p2)] + [L(p2)-L(h)] + [L(h)-q(1-q)]`.

The four terms describe fixed-tape delay cost, missing-feedback cost, remaining
selection headroom, and convex-hull approximation cost. The first two may be
negative; more feedback need not improve every realization. The last two are
nonnegative pointwise. Oracle headroom is not necessarily learnable because q
is unavailable to the algorithm. It is not proof of an implementation bug.

All688128 identities pass, maximum error1.11e-16. All168 corresponding actual
mixer means match the earlier composition artifact exactly. There are1008
fixed-tape as-of checks, including future-label guards for counterfactual modes.
All210 immediate-feedback group controls have exactly zero delay/missing cost.

Confirmation delayed terminal64 means:

| Case | Actual | Delay removed | All past feedback | Oracle hull | Noise floor |
| --- | ---: | ---: | ---: | ---: | ---: |
| Additive stationary | .221635 | .221419 | .221347 | .193862 | .173275 |
| Additive gradual | .235990 | .233970 | .233572 | .199086 | .173021 |
| Majority to parity | .056658 | .054380 | .053959 | .049016 | .047500 |
| Parity to majority | .098801 | .090903 | .088783 | .063034 | .047500 |

For the reverse switch, delay cost is .007898 [.001766,.014031], missing cost
.002120 [-.002141,.006381], remaining selection headroom .025748
[.015596,.035901], and hull cost .015534 [.007027,.024041]. Thus changing only
feedback delivery cannot close this fixed-tape gap. These approximate paired
mean +/-3.5SE intervals use32 trajectories; they are not simultaneous or anytime
certificates. The negative interval endpoint on a pointwise nonnegative term in
other cells reflects the approximate interval construction, not negative cost.

## A tested second-order alternative

[Koolen and van Erven (2015), Section2](https://proceedings.mlr.press/v40/Koolen15a.pdf)
construct Squint weights from cumulative excess loss and its squared sum,
averaged over learning rates. Their expert protocol reveals the loss vector
after prediction. This motivates the finite-grid component here; it does not
establish a delayed/missing-feedback guarantee for this adaptation.

The [Squint diagnostic](mmm-segment-v120-squint-diagnostic.json) uses rates
{.5,.25,.125,.0625} with uniform rate prior, generic64 prior.95 and uniform
remaining expert mass. It tests the original four experts and eight experts
adding variational64/32 and segment64/32. No transition, forgetting parameter,
post-outcome rate sweep or case-specific filter is added. When feedback arrives,
excess loss uses the ORIGINAL issued expert weights and predictions. The
randomized expert loss defines the update; the convex probability average is
the properly scored forecast. No teacher q enters the update.

There are1376256 forecast evaluations and672 as-of prefix checks. The component
also enumerates256 binary eight-forecast paths, checking the immediate-feedback
prior-potential bound through the first seven delivered labels (the final label
is not flushed). Maximum potential is1. That small numerical check is not a
general theorem or predictive validation.

[Paired comparison against archived Markov12](mmm-segment-v120-squint-comparison.json):

| Candidate | Non-harm cells | Meaningful-gain cells | Decision |
| --- | ---: | ---: | --- |
| Squint four experts | 126/168 | 2/168 | Reject as broad rescue |
| Squint eight experts | 117/168 | 2/168 | Reject as broad rescue |

The same .01 upper harm tolerance and .005 mean gain/positive lower endpoint
are used. These cell comparisons are not the complete fresh standalone gates.
Confirmation delayed terminal majority-to-parity is .121633 for eight-expert
Squint versus .060262 for Markov. The paired loss increase is .061371,
[.041262,.081480]. A passing potential check clearly does not imply that this
method adapts quickly enough to the workload.

## Next lead, with explicit boundaries

[Neuteboom and van Erven (2022), Sections4-5](https://arxiv.org/pdf/2209.06826)
develop Squint-CE using interval-active learners and a surrogate-loss meta
algorithm. This is a distinct changing-environment method, not ordinary Squint
with an arbitrary forgetting coefficient. The source's full-feedback setting
does not settle our delayed-label integration.

Before implementing that lead, resolve the normalized mixture consistently:
each interval learner's distribution on rate/expert pairs has its own partition
function. A mixture of these distributions must retain those normalizers.
Use the explicit mixture definition in Section5.1 as the reference and check
any collapsed implementation against it on unequal-evidence intervals; do not
assume cancellations in a displayed shortcut. Also retain each issued forecast
and active interval set when accepting late evidence. No source theorem should
be claimed for a changed update order without a separate argument.

The current code is plain finite-grid Squint, NOT Squint-CE. Remaining leads
include a verified changing-environment surrogate learner and conditional
expert selection. Both must retain all21 cases, stationary protection, both
switch directions and delayed/missing controls. Neither is validated here.

## Reproduction

All three JSON outputs record input/source hashes. The credit decomposition,
Squint diagnostic and paired comparison are separately rerun and required to
reproduce byte-for-byte. These are offline reference computations, not serving
performance measurements. No external/private data or production services are
used.
