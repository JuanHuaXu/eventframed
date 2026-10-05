# Context-dependent mixture: frozen consumed-data screen

Use only observable X (nine bits), as-of arrived outcomes, and issued baseline
and challenger forecasts. No teacher, scenario, oracle q, future X, or outcomes
enter the learner. Retain the most recent64 arrived origins by origin order.

Features are phi(X)=[1,s_0/3,...,s_8/3], s_j=+1 for a set bit and -1 otherwise.
Intercept scaling matches scalar ridge; the nine added features have squared
norm1. The matched global control zeroes all nine added coordinates.
Regularization is identity, mean weight prior.5, no hyperparameter search.
This is a declared small contextual extension of the prior ridge adaptation,
not an inherited AAR regret theorem.

For d_j=c_j-b_j and m_j=(c_j+b_j)/2, fit theta by minimizing

    ||theta||^2 + sum_arrived (m_j+d_j*phi(X_j)'theta-y_j)^2.

Propose w=clip_[0,1](.5+phi(X_t)'theta) and apply the unchanged .01 pointwise
Brier guard. The ridge objective is the unprojected affine-mixture loss;
clipping weights and the safety projection do not preserve that objective.
Score both guarded and unguarded contextual/global arms, Fixed Share/pointwise,
Fixed Share/local, and Markov. Preserve all failures, not just pooled means.

Before scoring, test the normal-equation residual, a separate iterative solver,
global-control equivalence, exact retained origins, future/unavailable evidence
invariance, and a deterministic context-dependent positive control. Benchmark
the reference cost including construction/solution of the10-dimensional system.
No online constant-time, fresh confirmation, real-agent, or adaptation claim.
No changes to fitting tapes, daemon, production, or whitepaper.
