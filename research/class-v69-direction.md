# V69 Class-Level Observation Lead

This is a separate mathematical lead, not a change to the frozen V68 study.
V68's Gini score uses its 180-state probe marginal. Distinguishing latent states
can spend observations on nuisance variation. We must specify the target
explanation/decision class before assigning a concentration reward.

For a declared class C, let m_c=P(C=c|visible), s_c=P(C=c,W2=1|visible),
p=sum_c s_c. The prospective score is
sum_c [s_c^2/p + (m_c-s_c)^2/(1-p) - m_c^2], omitting zero-probability branches.
This is expected class concentration gain, not empirical falsification or an
Anti-Pigeon sharing certificate. Refine latent atoms while preserving P(C,W2):
the score stays unchanged, unlike a Gini score over individual atoms.

The finite preflight freezes an independent original-joint branch enumeration,
256 declared joint cases, boundary/independence guards and a counterexample:
four equal atoms in two classes; a class-discriminating test has response
(.9,.9,.1,.1), a nuisance test (1,0,.5,.5). Splitting both B atoms into20
indistinguishable copies reverses the latent-Gini test ordering without changing
the class/outcome joint law. This proves a representation sensitivity, not that
it explains V68 quality or that class acquisition will win on real streams.

[Golovin, Krause and Ray (2010, revised2013)](https://arxiv.org/pdf/1010.3091),
sections3-4, distinguishes target equivalence classes from nuisance hypotheses.
The proposed score has the EffECXtive concentration form applied to a declared
target class. It is NOT EC2 edge cutting: the paper's competitiveness guarantee
depends on its complete decision classes/noise model and does not transfer to
our dynamic forecast family. Its appendix also gives poor cases for myopic
posterior-based criteria. No near-optimality claim is made here.

First runtime candidate: collapse V68's probe to its three alternatives
(local/shared-noise/shared-context), with likelihoods from the same joint law.
That targets a structural explanation, not verified external truth; a model
alternative is not automatically an operational decision class. Compare its
choices and downstream risk against the existing latent score, uncertainty,
random and all-target predictive value, with total computation charged.
Keep the independent and no-sharing controls, existing gates and full seven-goal
scope. Require exact posterior branch/readback checks, delayed/future/corruption
tests, whole-loop cost and cost-matched replication before any rescue claim.
Untouched task outcomes, valid useful splits and durable loaded freshness remain
required. No production changes or fresh confirmation from this preflight.
