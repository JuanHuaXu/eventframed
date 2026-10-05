# Fresh member-integration false-revocation coverage

The v79 mixture and v80 integration code are unchanged. Do not introduce another
rate-adaptation method already tested in v76-v79. This experiment addresses the
remaining finite-sample false-revocation gap, not a model-quality rescue.

Freeze512 trajectories per scenario per phase,512 steps each, all five original
scenarios (stable, member_shift, common_shift, recurring, null) and all five
paired arms. Total5120 trajectories. Seed bases2026091507 (design) and2026091508
(confirmation) were searched in current source/protocols with no matches before
this run. Use unchanged observationpreserved.Seed derivation and memberRun.
The fitted4096-label base is unchanged, so this does not test new fitting samples.

Keep original foreground/audit budgets, forecasting, nominations, split gates,
fitting cadence, loss windows and trajectory hashes. Charge missing/premature
splits the full remaining delay. No premature label, threshold or model change.

Primary missing evidence: false revocation in stable, common_shift and null for
both old and mixture gates in both phases. Report twelve separate one-sided
exact Clopper-Pearson upper bounds with alpha=.05/12 each. Bonferroni gives
simultaneous95% coverage of these finite-horizon rates under independent
trajectory sampling within each declared generator. Require every upper bound
<=.02. Do not pool scenarios, arms or phases to hide a failure. This is not an
anytime guarantee or a real-world target-law diameter certificate.

Retain every v80 pilot gate: member-shift restricted delay >=10% faster with no
increase in premature splits, post-Brier mean gain>=.005 and positive paired
lower bound; other scenarios full/post lower gain>=-.01. Paired descriptive
intervals remain mean +/-3.5SE, now over512 trajectories. Do not mark the whole
pilot passed merely because false-revocation coverage passes. Report counts,
misses, costs, detected-only and restricted delays separately.

Require focused race tests before collection, deterministic full replay,
source hashes unchanged during each run, independent binomial-bound checks,
record-shape/budget checks and independent summary validation. Retain failures.
All runs are isolated research, not OpenClaw/production/LLM tests. No model
download, dependency update, source deployment, whitepaper edit, commit or push.
