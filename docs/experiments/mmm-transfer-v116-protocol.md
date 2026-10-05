# Frozen independent-generator transfer v116

Freeze before quality generation. No tuning, sample extension, early winner
selection, changed comparison windows, or retroactive threshold relaxation.
Generator QA uses a different consumed seed base and is not confirmation.

## Design and policies

Use the construction in research/independent-generator-contract.md unchanged.
Quality seed base2180111700 plus phase*1000000 +family*100000 +mode*10000
+index*10, with independent RNG role offsets0..4. Audit effective seeds modulo
2147483647 against the archived blocks and generator QA base2176111600.
Use both phases, three families, three modes and32 indices per cell:576 latent
trajectories, each run with immediate and delayed/missing feedback,1152 runs.
Both phases are evaluated with no intervening policy changes. The confirmation
phase is a separate sample, not a new generator family.

Arms:0 unchanged generic64;1 unchanged arrival-log/no-neutral;2 unchanged fixed
rate delayed Markov. Preserve their existing priors, gates, fitting cadence,
four-model publication, acquisition controller and journal lifecycle. This is a
transfer test of existing research policies, not a new tuned mechanism.

Keep16 initial full-audit packets,256 scored ticks, as-of64/32 fits every32,
six-coordinate forecast acquisition cap, delays0..31 and missing probability.2.
Full packet audits for fitting are additional equal evidence, not free acquisition
within the six-coordinate cap. Initial audit cost is16*9; later audit cost is9
per arrived packet. Count acquisition reads separately and retain fit origins.
All arms receive the same full fitting publication, including unused specialists
for generic64, so whole-fixture timings are not standalone arm timings.

Earlier eligible feedback is released before fitting; the current zero-delay
outcome only after forecasting. Flush through clock287. A missing packet never
contributes a label. Policies cannot see the teacher or future feedback schedule.

## Mandatory fixed-sample quality screens

Primary score: expected Brier, averaging over all256 ticks and separately over
terminal ticks192..255. The final gradual transition can finish at190. This
terminal64 recovery window is not the archived late128 metric.

All348 screens must pass for an overall transfer pass:

-288 non-harm: each of arms1/2 against both other arms, in each of2 phases,
 9 family/mode cells,2 schedules and2 segments. Upper paired harm bound <=.01.
-48 gains: each of arms1/2 against generic0 in each of2 phases,6 changed
 family/mode cells and2 schedules, on terminal64. Mean gain >=.005 and lower
 paired bound >0.
-12 additional gains: Markov2 against arrival-log1 on delayed terminal64,
 for all6 changed cells and both phases, using the same gain rule.

Arm1 therefore has168 requirements; arm2 has180. Report each arm separately as
well as the joint result. Failing gain does not mean statistically proven harm;
passing non-harm does not establish useful recovery. No favorable subset may
replace the full predeclared requirements.

Use paired mean +/-3.5 standard errors across32 latent trajectories in each
cell. These are approximate fixed-sample screens, not anytime confidence
sequences, exact simultaneous guarantees, or correction for the full adaptive
research history. Immediate/delayed runs of one trajectory are not independent
replicates. Retain all means and bounds, including failures.

Expected accuracy, expected log loss, realized scores, mean acquisition cost,
full-input Bayes Brier floor and Bayes accuracy ceiling are descriptive only.
They cannot replace primary gates or drive the forecast. The new soft teachers
do not have the old near-deterministic95% accuracy ceiling.

## Evidence and boundaries

Write an exclusive0600 JSON artifact with every step, teacher parameters/seeds,
fit origin set, journal accounting and hashes of all source files in the three
relevant packages, go.mod/go.sum and this protocol. Recompute all trajectories
and require exact replay equality. Independently verify probability formulas,
scores, as-of fits, acquisition bounds, pairing, terminal counts and all gate
totals from JSON. Source hashes must still match when verification runs.

Do not edit frozen sources after generation; a necessary correction requires a
separately named artifact and an honest consumed-seed designation. Record actual
generation/replay times but do not treat them as serving latency. A separate
benchmark, if run, uses consumed QA data without concurrent heavy computation.

No production access, private data, pushing, whitepaper promotion or deployment.
A transfer pass alone cannot close all seven research directions or erase
earlier recovery failures. A failure must remain in the research record.
