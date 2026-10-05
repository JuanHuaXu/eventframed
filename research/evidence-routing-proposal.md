# Evidence-attributed routing after v100

Proposal only: no implemented or validated rescue is claimed here.

## Observed limitation

v100 fails both parity-to-majority positive-recovery gates despite passing all
96 non-harm checks and rejecting more published versions than neutral-only.
On the confirmation streams, generic32 has late Brier0.155822 versus generic64's
0.178977, yet its raw weight at step160 averages only0.006381. Comparative
gating produces0.176502, essentially the neutral gate's0.176421. These are
consumed diagnostics, not a new confirmation set.

Renormalizing survivors preserves their old relative weights. Rejecting an
incumbent therefore does not necessarily promote the forecast that supplied
the contradictory evidence. Underweighting is a candidate mechanism, not a
proven sole cause: wrong forecasts before retraining and short-window variance
after the long model catches up remain competing explanations.

## Proposed routing rule

Leave the four raw learners, their raw bank, and all v100 tests unchanged.
For these equations, raw experts are indexed1..4 and index0 denotes neutrality;
this differs from the zero-based raw-expert arrays in the existing artifacts.
At the FIRST crossing for target i, record each alternative's contribution to
its test mixture, including its fixed alternative mass and all32 start terms:

    e_i,j = a_j * (A_i,j + dormant_i) / 32
    c_i,j = e_i,j / sum_k e_i,k.

Alternative0 is neutrality; the others identify the three raw experts other
than i. Compute the normalized contributions in log space and freeze them for
that tested version. They are evidence-allocation fractions, NOT probabilities
that an alternative is true or that it has smaller future Brier loss.

For a subsequent prediction, let S be the currently unrejected raw experts.
For each rejected target i, retain only its neutral credit and credits for
alternatives in S, renormalizing these surviving credits to d_i,j. Neutrality
always supplies a mathematically positive remaining term. If finite arithmetic
cannot normalize safely, fail to neutrality rather than inventing confidence.

With the current, unchanged raw-bank weights w_i, form

    v_j = w_j + sum_{i rejected} w_i d_i,j  for j in S
    v_0 = sum_{i rejected} w_i d_i,0
    p_routed = v_0 * 0.5 + sum_{j in S} v_j p_j.

This is a convex mixture. No rejected expert receives routed mass, and no
recursive routing or cycle resolution is needed. With no rejection, return the
raw-bank law. With no survivors, return0.5. At the fixed next publication,
discard all rejection credits together with the version's tests. Ordinary raw
bank updates continue on every received outcome throughout.

The revealing outcome can determine credit only AFTER its already-issued
forecast; routing begins on the next prediction. The rule must not use hidden
simulator truth, case labels, known change times or future outcomes.

## What is and is not inherited

The test-martingale interpretation comes from
[Shafer et al., author working version revised December 2010](https://www.probabilityandfinance.com/articles/33.pdf).
It supports evidence against a conditional forecast law, not the reliability of
a routing policy. The proposed credit transfer is our experimental design, not
an algorithm or performance theorem attributed to that paper. Keep the same
64-test/0.01 allocation and threshold6400; using contributions must not add an
uncharged maximum or extra rejection rule. The raw bank's cumulative-loss bound
still does not cover the served routed forecast.

## Falsifiers and next experiment

First check mass conservation, nonnegativity, exact target/alternative mapping,
no routing to rejected recipients, simultaneous rejection, all-neutral fallback,
next-outcome timing and publication reset. Replay the same inputs to verify
identical rejection times and raw-bank forecasts between v100 and routing.

Then freeze a fresh learned experiment with all106 quality gates and retained
generic, raw-bank, neutral and comparative controls. Report neutral routed mass,
effective generic32 weight during each publication, and block-level losses so
more promotion is not mistaken for useful promotion. Model fitting and evidence
volume stay unchanged. Benchmark the added O(K^2) routing for capped K=4.

Failure to obtain both positive recovery gates, new stationary harm, delayed
promotion beyond the useful short-window interval, or credits mostly sent to
the wrong alternatives would falsify this rescue. Preserve that result rather
than tuning thresholds on it. Delayed feedback, real-agent tasks and full-roadmap
validation remain separate requirements even if this finite experiment passes.
