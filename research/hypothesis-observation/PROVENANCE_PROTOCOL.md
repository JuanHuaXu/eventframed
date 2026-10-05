# Source-aware hypothesis observation v2

Frozen before running provenance_v2.py. Additive follow-up, old experiment and
evidence unchanged. Same 16 hypotheses, four target classes and eight binary
tests as v1. This is known-model synthetic evidence, not factual verification.

Four cases: noise05 and noise20 have four independent source instances per test,
each instance returning the same result when reread; one_source has only one
instance per test with noise05; spoofed presents four IDs per test but secretly
reuses one actual result. Last case violates the declared source independence
assumption and must remain an explicit non-certified boundary.

Arms: naive target-Gini ignores source history and always requests source0;
source-aware target-Gini, source-aware label entropy, source-aware random.
Source-aware methods select only unobserved (test,source) pairs, applying one
likelihood update per independent pair. Known provenance is supplied by this
simulator, not inferred or authenticated by the acquisition policy. Independence
is conditional on H; distinct IDs alone do not establish it in real systems.

Use 16 decision slots. One_source exhausts eight sources then emits unchanged
forecasts for remaining slots; report actual queries separately. All methods
have the same available source bank and acquisition budget. No free extra data.
Report pre-query learning-curve Brier over all16 slots, final Brier, accuracy,
confident-wrong fraction (max probability >=.9), and actual acquisition count.

128 episodes/case/split, two splits; seed=2026101201*1000000+split*100000+
case_index*1000+episode. RNG roles seed*10 for truth, +1 for 4x8 outcome tape,
+2 for random selection. Policy sees likelihood and past requests only. No
tuning between splits. All arms share the latent truth and potential tape.

Confirmation screening for BOTH noise05/noise20: final-Brier improvement over
naive >=.02 with paired z3.3 lower>0; curve-Brier improvement over source-aware
random and label-entropy >=.02 with paired lower>0. One_source must not increase
final Brier over naive. Report spoofed without requiring pass; no general
provenance guarantee or research-roadmap completion follows from matched pass.

Tests: unique pair admission, no second likelihood update on exhausted bank,
fair-coin no-update, deterministic full replay, source/input hashes, and hidden
truth unavailable to the policy. Latency of this Python research loop is not a
serving benchmark. Retain failures, especially spoofed-source results.
