# Augmented evidence v74

Frozen before fresh data. Directions3/7 research only. v72's importance-corrected
targeting improved stopping by ~5%; test whether a predictable control variate
reduces noise further without changing the target, paid queries or false-alarm
budget. Source inspiration: [Dudik, Langford & Li (2011), Doubly Robust Policy
Evaluation and Learning](https://icml.cc/2011/papers/554_icmlpaper.pdf).
Our bounded control-variate betting process is an adaptation, not their policy
optimizer. Actual randomized propensities remain required; no robustness to
wrong propensity metadata is claimed, especially after coefficient clipping.

Four channels, uniform target p=1/4. Keep last32 signed observations per channel.
Before selecting a query compute m_i=sum D/(n_i+2), v_i=(1+sum D^2)/(n_i+2).
New proposal q_i=.1+.6*sqrt(v_i-m_i^2)/sum_j sqrt(v_j-m_j^2).
These smoothed moments ensure positive estimated variance. q is computed before
the next outcome, and only the selected channel updates its history.

Let w_i=.25/q_i, mbar=sum_i m_i/4, c_i=mbar-w_i*m_i.
Choose one predictable eta=min(1,min_{i:|c_i|>0}(3.53-w_i)/|c_i|).
Use Z=w_I D_I+eta*c_I. Then E[Z|past]=the uniform target mean because
sum_i q_i*c_i=0. Also |Z|<=3.53, so signed factors1+.25*(s Z-.15)
are >=.08 and retain null conditional expectation<=1. eta depends on all
pre-query q,m, NOT on selected outcomes. No outcome-dependent clipping of Z.
Retain the eight-start pooled wealth threshold100 and null margin.15.

Four arms: uniform/pool; v72 moment-proposal/IPW; new variance-proposal/IPW;
new variance-proposal/augmented. The last two share identical selected outcomes;
this separates acquisition from control-variate effects. All arms use one query
per step. New policy does not access channel truth, future labels or scorer state.

Use all ten v72 scenarios unchanged,512 streams x512 steps per phase/scenario.
Fresh design base2026117401 and confirmation2026117402, with the existing
base*1e6+scenario*1000+stream seed convention. No tuning between phases.
Store per-stream first alarms, query counts, augmentation clipping count and
the SHA256 of q/m/eta/channel/outcome tapes; full replay must match them.

Primary adoption candidate is augmented arm only. Confirmation sparse128 and
sparse256 require >=10% restricted-mean gain against uniform, paired z3.3 lower
gain>0 and no increase in premature count. All null Wilson95 alarm upper<=.02.
Other alternatives permit at most10 steps mean delay harm. All alternatives
require observed excess misses<=.01 and the v73 paired rectangular upper<=.02
using alpha=.05/12 over six alternatives/two phases. Thus the improved bound
is declared before these fresh outcomes; do not reinterpret older v72 results.
Other arms are frozen ablations, not post-hoc substitute primary candidates.

Test unbiasedness under incorrect m, all-channel factor positivity, zero-m
parity, proposal normalization, invalid inputs/state isolation, current versus
future observation ordering and full replay. Benchmark bounded policy/gate only.
All source, protocol and evaluator hashes must be recorded. Passing this finite
screen would still require independent families and actual MMM/Anti-Pigeon
integration. The null remains population mean, not the full diameter certificate.
