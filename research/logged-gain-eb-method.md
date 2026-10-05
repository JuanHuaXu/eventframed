# Time-varying empirical-Bernstein gain bound

## Attribution

[Waudby-Smith et al., Anytime-valid off-policy inference for contextual bandits](https://arxiv.org/html/2210.10768v3),
section3, AppendixA.1 and AppendixA.3, provides the empirical-Bernstein
supermartingale and permits any fixed mixing distribution. We use a discrete
seven-point distribution instead of the paper's truncated-gamma mixture.
This is not its hypergeometric closed form, a claim of equal tightness, or a
constant-mean betting confidence sequence. The paper's boundedness and
predictability requirements remain essential.

## Our specialization

Reuse the exact gain pseudo-outcome X from logged-gain.mjs. This experiment's
collapsed logging probabilities are at least1/3. Enforce its predictable support
within[-3,3] before observation. Define c_i as the prior sample mean clipped
to[-1,1], with c_1=0. Let mu_i=E[X_i | pre-action information]. Then both
(X_i-c_i)/4 and (-X_i+c_i)/4 are at least-1.

Let psi(lambda)=-log(1-lambda)-lambda and

    V_n = sum_i ((X_i-c_i)/4)^2
    S_n(g) = (sum_i X_i - n*g)/4.

For each frozen lambda in(0,1), the two processes
exp(+/-lambda*S_n(g_n)-psi(lambda)*V_n) are nonnegative supermartingales at
g_n=sum_i mu_i/n. This follows from the source's Fan-inequality argument with
Z_i=+/-X_i/4 and predictable center+/-c_i/4. The factor4 is fixed: neither
conditional means nor the target average silently lose this normalization.

Average the processes over rates[1/128,1/64,1/32,1/16,1/8,1/4,1/2]. Solve
mean_lambda exp(lambda*b-psi(lambda)*V_n)=2/alpha. The two-sided interval is
sampleMean +/-4*b/n, intersected with[-1,1]. Ville plus the two-sided allocation
gives simultaneous coverage of the running average conditional gain. Do not
intersect successive intervals, since g_n can change. Exactly known singleton
increments permit an exact interval only when all increments so far were
singleton; observed zero residual variance does not imply certainty.

The endpoint restriction is checked, not presumed from whichever action was
observed. The center is computed before adding the current observation. The
existing logged outcome/provenance contract remains the caller's obligation.
An observed sequence cannot establish its own logging probabilities or missing
outcome mechanism. This is a bounded three-strategy specialization, not the
paper's more general treatment of unbounded importance weights.

## Checks

3456 finite conditional MGF checks cover both signs and three predictable
centers. An adaptive six-step tree checks4096 leaves. A1000-step positive
control reaches a lower bound0.94406 at true gain1 while retaining nonzero
uncertainty. Zero-residual/non-singleton, invalid range, duplicate/order and
failed-update atomicity controls pass. These checks accompany the argument;
they are not its substitute.

The independent audit uses a power series for psi rather than log1p, independently
decodes actions, reconstructs prior-center variance and checks every prefix's
coverage/positive-bound decision. No policy, data, rate grid or regression was
retuned after observing this experiment.
