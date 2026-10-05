# Fixed-model label-budget curves V33: completed study

2026-10-03. Both frozen scarcity-only rescue screens **FAIL**: 19/24
design and16/24 confirmation cells fail at the predeclared150 labels.
More evidence substantially repairs curved-case calibration, but does not
establish all protection and improvement requirements. Production and the
whitepaper remain untouched; all seven whole goals remain OPEN.

## Frozen Study

The [protocol](mmm-curve-v33-protocol.md) and
[preflight](mmm-curve-v33-preflight.md) precede both fresh splits. Seed
bases2026103303/2026103304 produce768 worlds each, with a100million
geometry stride avoiding cross-cell seed collisions. Six unchanged
learners share the SAME label-independent150-member nomination tape;
the static baseline is a seventh forecast at every checkpoint.

Budgets16,32,64,128,150 produce53,760 dependent snapshots across1,536
independent worlds, not53,760 independent trajectories. Each member
provides one label. Known latent rates enter only evaluator future-risk
calculations, never fitting or nomination. At150 every member has been
observed: unobserved-member risk is null, NOT zero or an extrapolation
success. Future risk is for new trials at these same fixed member rates.

- [Design tape](mmm-curve-v33-design.jsonl), SHA256
  `aed241b285debe13a58d2eb090c1378015b9d8d2bc13c844da2ee6369d9820a2`.
- [Confirmation tape](mmm-curve-v33-confirmation.jsonl), SHA256
  `8172dc78e017a3e51a62b9deee4db88e102bb1ea3eabec8eada32c2ee34886e1`.
- [Summary](mmm-curve-v33-summary.json): every model, checkpoint, metric,
  interval, cost, gate and19 frozen source hashes.

## Learning Curves

Confirmation means for exact-crossfit on randomized-stratum evidence:

| Case | Labels | Whole Brier | Top10 expected usefulness | Packed signed bias |
| --- | --- | --- | --- | --- |
| Tight curved | 32 | .239778 | .719800 | .117484 |
| Tight curved | 64 | .205504 | .812279 | .058829 |
| Tight curved | 150 | .198710 | .849504 | .040497 |
| Wide curved | 32 | .250363 | .511159 | .295435 |
| Wide curved | 64 | .208752 | .769879 | .066663 |
| Wide curved | 150 | .197697 | .840704 | .041070 |
| Wide independent | 32 | .272114 | .612500 | .200282 |
| Wide independent | 150 | .219824 | .681875 | .039759 |
| Wide reversed | 32 | .223629 | .759195 | .044256 |
| Wide reversed | 150 | .219493 | .872282 | .015364 |

Lower Brier and higher usefulness are better. These are synthetic future
expectations, NOT agent-answer accuracy. Intervals are mean+/-3.5SE
over32 worlds per cell, not time-uniform or simultaneous AP certificates.

Wide-curved bias magnitude upper at150 is .069467 design and .076142
confirmation, passing .10. This is a substantive repair of the large
small-budget defect. Its overall gate STILL fails: the simple blend at150
has Brier .197001 and usefulness .848124, versus crossfit .197697/.840704.
The predeclared requirement was positive .01/.02 improvement over blend,
not merely convergence toward a similar forecast. Those stricter gates
remain unchanged even when more labels let both methods perform well.

Tight-curved150 strongly beats affine (.223855 Brier/.302018 usefulness),
but blend/stack usefulness protection lower endpoints -.027105/-.015420
still fail -.01. Some failures reflect uncertainty rather than a proved
large mean loss; the gates do not distinguish these as passes.

Wide-independent150 bias upper .132604 design/.103403 confirmation fails
.10, and required improvement over blend fails. Narrow peaks and several
coordinate-irrelevant/alternating cases still fail. Wide-reversed150 now
matches affine usefulness .872282 in confirmation, but its design blend
protection lower endpoint -.015940 fails -.01. Only eight confirmation
cells and five design cells pass all frozen requirements.

More labels are not uniformly beneficial. Tight-aligned confirmation affine
whole Brier reaches .207629 at64 before rising to .217619 at150;
tight-aligned reaches .207739 design at64 before .218549 at150. The
fixed Beta2 member hierarchy updates individual rates from one noisy label
even when a deterministic mean function describes the generator. That is
a plausible variance-model issue, not an arithmetic bug established here.
The learning curves motivate a test of repeated outcomes and learnable
dispersion, not retrospective selection of64 as an adoption threshold.

## Audit and Cost

Independent JS passes all53,760 snapshot reconstructions:19 source hashes,
nomination eligibility/probability, six pre-outcome laws, original stack
rows, exact current-prefix LOO integrals, analytic simplex solves, full
laws/weights, static baselines, packets, risks, observed/unobserved
denominators and timing sums. Exact BOTH-split Go replay agrees on all
non-timing fields, including raw generated latent draws and nomination RNG.
Only explicitly elapsed fields are excluded.

[Eight corrupted-tape controls](mmm-curve-v33-negative-controls.json)
reject duplicates, nomination probabilities, issued laws, original stack
rows, LOO rows, fabricated empty-population risk, cost sums and source
hashes without editing original tapes. Future-label flips preserve earlier
checkpoints, and disabling diagnostic snapshots preserves later learning.
Sampler/exhaustion tests through200 members, all three module race tests
and vet pass.

Maximum per-model AccountedNS is2.698744ms design and2.463460ms confirmation,
below the frozen10ms component-work limit. This is the sum of measured
setup, probes, updates, current snapshot and earlier snapshots, NOT
standalone wall or daemon latency. Shared nomination is recorded separately;
world elapsed covers six interleaved learners and diagnostics. InputNS
records the legacy generator's additional discarded control-arm work.
True-rate evaluation is outside all learner timings. No loaded serving,
external acquisition or equal-total-cost observation benefit follows.

## Interpretation

The frozen scarcity-only rescue is rejected on this finite population.
This does NOT mean additional evidence is useless, nor identify one sole
cause: curved-case calibration clearly improves, some methods converge,
while protection, improvement and small residual bias requirements remain
unmet. More repeated outcomes per member, a different mean family, source
dependence, delayed feedback and temporal changes remain separate questions.

Next isolate between-member heterogeneity from Bernoulli trial noise. With
one trial/member, Beta(kappa*p,kappa*(1-p)) marginalizes to Bernoulli(p)
for EVERY kappa; those likelihoods cannot learn kappa. Repeated independent
trials can distinguish the models. A new bounded joint-model experiment
must not treat copied evidence or repeated processing as new observations.
No Anti-Pigeon authority, adaptive recovery, actual agent benefit, loaded
freshness or equal-total-cost Goal7 success is established by V33.
