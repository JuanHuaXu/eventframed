# Predictable evidence allocation v72

Frozen before running. Directions3/7 prerequisite, not a complete MMM trigger
integration. Targeted observation can change the evidence population. Before
connecting a warning to acquisition, test whether correcting that selection can
retain the declared null and yield a worthwhile equal-query detection gain.

## Target and candidate

Four known channels define a uniform target population p_i=1/4. Each query
returns a fresh paired difference D in {-1,0,1}; only one selected channel is
observed. Under the null the target mean satisfies |sum_i p_i E[D_i|past]|<=.15.
Individual channels may violate that bound. This does not assert that the
channel populations are causally identified or that arbitrary missing evidence
has been corrected.

Maintain the last32 observed D^2 values per channel. Estimate v_i=(1+sum D^2)/
(2+count), then before the next query use q_i=.1+.6*sqrt(v_i)/sum_j sqrt(v_j).
The .1 floor guarantees coverage and caps p_i/q_i at2.5. All other outcomes are
unseen by the selector. Estimate update occurs after selection and feedback.

Arms: uniform observation + pooled fixed gate; adaptive observation + naive
unweighted pooled gate (deliberately invalid control); identical adaptive
observation + importance-corrected pooled gate. The latter uses Z=(.25/q_I)D.
Then E[Z|past]=sum_i p_i E[D_i|past]. Signed wealth factors
1+.25*(s Z-.15) are >=.3375 and conditional expectation<=1 under the null.
Use the v71 eight-start arithmetic wealth mean and threshold100. Thus the
corrected arm retains the per-arm1% anytime bound; naive targeting does not.
No estimated propensity is substituted for the known randomized q.

For fixed known v_i=E[D_i^2], minimizing sum_i p_i^2 v_i/q_i without a floor
gives q_i proportional to p_i sqrt(v_i). The bounded rolling estimate above
is our heuristic, NOT a finite-sample optimality or convergence guarantee.
Source motivation: [Ryu & Boyd, Adaptive Importance Sampling via Stochastic
Convex Programming](https://web.stanford.edu/~boyd/papers/adaMC.html). Our
discrete second-moment rule is not their exponential-family optimizer. Gate
validity follows from the explicit conditional-expectation calculation here,
not a claimed inherited theorem about our acquisition heuristic.

## Frozen experiment

512 streams per phase/scenario;512 query steps each. Design base2026117201,
confirmation2026117202; seed=base*1e6+scenario*1000+stream. Common randomized
selection/outcome tapes pair arms; latent outcomes are generated in the simulator
and passed only after the arm chooses its channel. Naive/corrected arms have
identical selected observations. Sampling continues to horizon for comparisons;
one paid query per arm per step, no all-channel observations in the learner.

Ten scenarios, channel entries are (P(+1),P(-1)):
1. symmetric null: all(.5,.5).
2. sparse null: all(.05,.05).
3. heterogeneous boundary null: channel0(1,0), others(0,2/15). Target mean.15.
4. cancelling null: channels0/1(.9,.1)/(.1,.9), others(0,0).
5. homogeneous128: after128 all(.5,.1).
6. sparse128: after128 channels0/1(.85,.05), others(0,0).
7. sparse256: same change at256.
8. negative256: after256 channels0/1(.05,.85), others(0,0).
9. sparse384: sparse change at384.
10. weak256: after256 channel0(.8,0), others(0,0).
Alternatives use all(.05,.05) before change. Different streams are independent;
source dependence/spoofed propensities and real-agent task mapping remain open.

Primary corrected-arm screen in confirmation: sparse128 and sparse256 require
>=10% restricted-mean delay gain against uniform, paired z=3.3 lower gain>0,
and no increase in premature count. Every null requires Wilson95 alert upper
<=.02. Other alternatives allow at most10 steps restricted-delay harm. All six
alternatives require excess misses<=.01 and a harmful paired-discordance
Clopper-Pearson upper<=.02, using alpha=.05/12 over six alternatives/two phases.
Premature and absent alarms count as misses and receive full remaining delay.
Naive arm is diagnostic and cannot pass a validity claim by finite simulation.
Freeze all cells and report failure rather than tuning sampling floors or gates.

Persist every seed, first alarm, selected-tape SHA256, channel query counts and
observed moment summaries; full deterministic replay reconstructs q/channel/D
per step and verifies the hash. Hash source/protocol files and use exclusive
output creation. Verify weighted conditional expectation on heterogeneous nulls,
positivity, uniform parity, input rejection and pre-outcome selector timing.
Measure bounded policy/gate computation separately from research replay cost.
