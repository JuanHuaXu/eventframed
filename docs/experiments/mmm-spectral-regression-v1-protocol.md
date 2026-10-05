# Bounded interaction regression, frozen v1 protocol

Frozen 2026-09-15 before collecting forecasts. Exploratory comparison on all
2688 consumed v120 records (21 cases, two phases, two schedules, 32 trajectories).
Both phases are consumed; the historical phase label is not fresh confirmation.
No production changes. Addresses representation, not nomination or gate tuning.

Three signed-label least-squares ridge learners, each with 32/64 most recent
arrived labels, refit from zero every 32 frames. All retain intercept and nine
main effects. Linear has no interactions. Screened retains at most16 degree2-4
products passing the previously frozen union-Hoeffding threshold. Full retains
all246 degree2-4 products without screening. All minimize sum squared error plus
sum squared coefficients (lambda1, including intercept). This is deliberately
NOT logistic regression: dual ridge permits exact bounded-size solves for the
full256-feature alternative, and a linear least-squares control isolates that
loss change. Compare additionally to archived logistic ridge, generic, Boolean,
and Markov predictions. No fit uses teacher identity, Q, future Y, or case ID.

Prediction: clip((1+f(x))/2, 1e-12, 1-1e-12), fixed until the next publication,
matching the source learner comparison. No claim that clipping creates a calibrated
Bayesian posterior. Evidence origins are rebuilt independently, require i<clock,
not missing, i+delay<=clock; initial16 labels are available at clock0. Reject
any mismatch with archived fitting origins. Cap64; no asynchronous delivery or
serving integration is simulated here.

Numerics: solve (Phi Phi^T + I) alpha=y by Cholesky, recover beta=Phi^T alpha.
Require finite outputs; component tests check dual residuals and equivalence to
analytic exhaustive-cube coefficients. Tests must include parity, main effects,
constant labels, duplicate inputs, and invalid inputs. Feature enumeration is
teacher-independent. Screening uses training labels, not held-out outcomes.

Report expected Brier, realized Brier, and expected accuracy over all256 and
terminal64 frames. Preserve each forecast. Gate each candidate/window against
its same-window linear-L2 and archived logistic, generic, Boolean controls plus
Markov: non-harm in every phase/case/schedule/window uses paired loss increase
upper endpoint<=.01. Recovery on cases1/2/4/5/7/8/19/20, terminal64, against
linear-L2, logistic, generic and Markov requires mean gain>=.005 and positive
lower endpoint. Endpoints mean +/-3.5*SE over32 trajectories match the prior
exploratory screen; they are not an anytime or multiplicity-corrected guarantee.
No success declared from aggregate average or a strong-parity subset. Record
all gates and negatives; no tuning from these outcomes. All seven goals remain
open regardless of isolated component gains.

Cost: fitting O(N^2*d+N^3), coefficient recovery O(N*d), prediction O(d),
N<=64, d<=256. Fixed full/linear feature Gram tables may be precomputed; screened
Gram construction is per-fit. Record experiment wall time and separate component
fit/predict benchmarks; neither is a loaded serving latency benchmark.

Source inspiration: [Heidari and Szpankowski2023](https://proceedings.mlr.press/v206/heidari23b.html).
This bounded ridge variant and screening are not their agnostic PAC learner.
