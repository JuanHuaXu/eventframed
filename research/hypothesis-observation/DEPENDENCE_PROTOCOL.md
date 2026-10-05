# Explicit source-dependence model v3

Frozen before running dependence_v3.py. The source-bank and 16-hypothesis task
are as in v2. Four sources per test, sixteen acquisitions per arm. Every
acquisition uses a previously unread pair. This does not authenticate sources.

Joint model: prior P(H)=1/16. Each test has an independent latent mode M_t with
P(M_t=independent)=P(M_t=copy)=1/2 conditional on H. In independent mode, fresh
source outcomes are iid Bernoulli(likelihood_t(H)); in copy mode the first
Bernoulli result is repeated by every source. Conditional on H, different tests
factor. Maintain P(H|history) and P(M_t=independent|H,history). Fresh-source
predictive probability is m_t(H)*likelihood_t(H)+(1-m_t(H))*first_t after a first
observation; before that it is likelihood_t(H). Update the mode AND H from this
same joint model. Disagreement makes copy mode impossible for that test.

The acquisition target is expected reduction in target-class Gini under the
joint model, not entropy of source modes. Predictive entropy and random pair
selection share the same model as controls. Independent-only Gini is the v2
comparator. A conservative comparator reads each test once and then stops;
pad unchanged forecasts to16 decision slots and report real acquisition count.

Cases: independent05, independent20, copied05, copied20, mixed20. Mixed copies
even-indexed tests and independently samples odd-indexed tests. The observer
never receives case/mode/truth. Noise tables are declared correctly; arbitrary
source bias, cross-test dependence and unavailable true hypotheses are outside
this family. These remain future boundary tests, not implied successes.

128 episodes/case/split, two splits. Seed=2026102201*1000000+split*100000+
case_index*1000+episode; truth role seed*10, 4x8 outcome tape +1, random policy +2.
No design-to-confirmation tuning. Report all arm forecasts, posterior summaries,
accuracy, pre-query curve Brier, final Brier, confident-wrong and acquisition count.

Screen: for copied05/copied20/mixed20 confirmation, joint-Gini final Brier gain
over independent-only >=.02 with paired z3.3 lower>0. On both independent cases,
joint-Gini curve-Brier harm over independent-only <=.02 (point-estimate screen,
not equivalence proof). On independent20, require positive paired lower curve
gain over joint-random and joint-entropy. Fail any condition => overall FAILED.
No suppression or threshold adjustments after results. Report stop-once control.

Verify joint posterior against direct enumeration of H and all256 source-mode
assignments, zero mode mass after disagreement, normalized forecasts, unique
requests, full source-hash and trace replay. No production or latency claim.
