# MMM + Anti-Pigeon rescue v2 protocol

Frozen before running the rescue, 2026-09-12. New fitting seed 2026091401,
design 2026091402, confirmation 2026091403. No tuning between splits. Existing
v1 artifacts/source stay unchanged. This is a Go primitive-level integration,
not a service/store certificate or full X3 implementation.

## Hypothesis and source of evidence

Use existing bayes.ApplyOutcome and bayes.AssessRevision, not a new detector named
Anti-Pigeon. The monitored sharing assumption concerns reliability of the ORIGINAL
attention policy in two initially compatible contexts: independent reference and
live streams. Each supplies its own subsequently observed correctness outcome.
The reference stream is a synthetic extra source, not an oracle. Both streams
are fully monitored; selection cannot conceal an outcome. A prior sharing
relationship is an explicit fixture assumption, NOT a validated target-diameter
certificate produced by this harness. The primitive can revoke that assumption;
no new sharing certificate is issued.

Production-default change policy: hazard .05, threshold .30, run cap 64, fast/slow
rates .25/.025, drift threshold .30, persistence 12, warm-up 20, CUSUM slack .10,
boundary 8, cooldown 20. Group policy: split prior .5, decision threshold .95,
member support 8, max members 64, equivalence width .15, threshold .80, uncertain
borrowing cap .10, shared weight .5. Group evidence is a sliding 64-pair window;
decisions start only after 64 pairs. These are working heuristics under repeated
monitoring, not proven confidence sequences. Observe empirical false revocations.

## Arms

All use the v1 reader, conditional-count model and six-coordinate foreground
budget. Fitting uses 4096 independent simulator examples per known task type.

| Arm | Invalidation | Recovery |
| --- | --- | --- |
| frozen | none | original MMM |
| audit_learning | none | always fit recent independent full-view audits |
| cp_fallback | production changepoint | breadth inspection, uniform forecast |
| ap_fallback | production AP split/reset decision | breadth inspection, uniform forecast |
| cp_audit | production changepoint | independent audit relearning |
| ap_audit | production AP split/reset decision | independent audit relearning |

An audit flag is an independent Bernoulli(.25) draw before each frame/outcome.
ALL arms receive the same audit opportunities, including unused control data.
After forecasting, selected audits inspect all nine coordinates, charged nine
additional units without assuming free reuse. Audit-only learning uses at most
128 recent full observations. On initial invalidation the two gated learners
clear their audit fitting buffer, retain the old model for rollback, and begin
with uniform forecasts. At 32 new complete audits they fit a local working model;
then refit every 16 audits on the most recent 128. This is research-only model
replacement, NOT a held-out production-publication certificate. No labels or
unread values are inferred to complete a sample.

The old sharing assumption is revoked at most once per stream. Replacement
models remain context-local and are never automatically re-shared. The monitor
keeps checking the old policy, not the adaptively improving replacement. All
arms pay for independent reference and frozen-live shadow inspections; report
those costs separately from foreground and audit work. Full-stream correctness
outcomes and reference observations are important assumptions, not free sensors
in a real deployment. Foreground prediction is journaled before feedback; the
revealing outcome can affect only the NEXT forecast. Duplicate/out-of-order
feedback, epoch mismatch, or missing pending prediction is rejected.

## Scenarios and budgets

32 independent streams per split/scenario, 512 frames each. Binary nine-field
simulator; same supplied views as v1. Five percent label flips, except null's
independent fair outcomes. First 256 frames are old-regime for single shifts.

- stable: reference/live both process-bit XOR throughout.
- member_shift: live changes to local third bit; reference stays process XOR.
- common_shift: both change to local third bit; reset need not mean split.
- recurring: live alternates process XOR/local third bit every 128 frames;
  reference stays process XOR. Single revocation plus rolling relearning is tested.
- null: both outcomes independent fair draws; fitting uses the same null law.

Seed-allocation correction before the authoritative rerun: evaluation RNG seed
is split_base * 1000000 + scenario_index * 100000 + stream_index * 100 + role,
where role is 0 (live), 1 (reference), or 2 (audit). These ranges are disjoint
from each other and the fitting seed. The superseded initial run used adjacent
split bases without scaling, causing cross-role design/confirmation reuse. Its
artifacts are retained in mmm-antipigeon-v2-superseded-seed and are not confirmation
evidence. Its aggregate pass/fail output was seen before this correction; no
model, thresholds, arm, scenario, or acceptance criterion was changed.

Separate RNG streams for reference, live and audit coins. The controller and
monitor do not receive scenario names, true change times or generator formulas.
Held-out trajectories may revisit known feature combinations: no grokking or
novel-rule discovery claim.

## Criteria

Primary: ap_audit post-change Brier gain >= .05 over frozen, with simultaneous
paired lower bound >0, in BOTH member_shift and common_shift confirmation.
Protection: stable mean Brier excess <= .01 over frozen with simultaneous upper
bound <= .01; stable false-invalidation fraction <= .05. Also flag >.01 mean
harm in every scenario. No overall pass if stationary protection fails.
Report recurring and null without treating them as supported by a pooled win.
Compare ap_audit against cp_audit explicitly: a tie does not identify a unique
Anti-Pigeon benefit. Compare relearning with fallback to distinguish reduced
overconfidence from recovered useful prediction.

Per-stream metrics, full forecast/action traces, training counts, costs, delay
over detected streams only, and missed invalidations are preserved. Use z=3.5
paired normal intervals over 32 independent streams (at most 100 reported
intervals; approximate, conditional on the frozen fitting model) and Wilson 95%
intervals for revocation/miss proportions. No interim outcome-based stopping.
All forecasts, including recovery warm-up, count in the post-change score.

Run tests/race checks first, then design and confirmation together, then serial
microbenchmarks. No production services or private data; no GitHub push implied.
