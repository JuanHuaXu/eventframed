# Age challenger breadth v88

Freeze before fresh outcomes. Keep the v87 learner, 128-step fit-time age
window, support16, retained control, audit cadence, selector and journal
unchanged. Test the ten v83 generator cases: stationary uniform, bit, XOR2,
majority3, multiplexer, parity4, dependent bit, dependent XOR2, stationary
dependent, and null. All changed rules use at most the first5 coordinates,
within the cost6 acquisition cap. Keep noise0.05 and change step256.

Use immediate feedback and combined jitter0..31/missing0.2, the clean anchor
and hardest v87 schedule. This is not additional coverage of every delay law.
For each of two phases and ten cases use64 trajectories. Each trajectory gets
its own independently fitted4096-label base; reuse that base and latent stream
across the two schedules. There are1,280 underlying fit/stream pairs and2,560
schedule-runs, not2,560 independent trajectories.

Training seed =2026118800*1000000 + case*10000 + phase*1000 + index.
Stream base =2026118801 +10*case +phase, with the existing Seed(base,case,index,
role) and all five RNG roles. Verify collision freedom before generation.
Dependent input mapping is exactly the v83 mapping; the subset learner's
uniform input mass remains unchanged, hence intentionally misspecified there.

Primary: each of the seven changed cases under combined stress needs candidate
post-Brier gain >=0.005 versus retained control and paired z3.5 lower>0, in both
phases. All other cells require full/post harm upper<=0.01. Preserve all results;
any failure rejects overall breadth adoption. Also report absolute Brier,
accuracy, fits, support and acquisition. Intervals are approximate finite
trajectory screens, not time-uniform or calibration guarantees.

Require unit parity with v87 for the unchanged bit generator, generator truth
tables, deterministic replay, matching latent streams across schedules, source
hashes, bounded metrics and feedback accounting. No extra performance claim:
the candidate is unchanged and v87 microbenchmarks remain narrowly scoped.
No production, OpenClaw, commit or push changes.
