# Frozen segment-membership comparison v120

Predeclared before fresh outcomes. No hazard, family prior, sample count, window,
fit cadence or acceptance change after generation. All21 v119 cases remain:
nine transfer family/mode cells and12 Boolean regimes. Keep every incumbent.

## Data and computation

Use transfer base2200112200 and Boolean base2204112300, retaining the existing
phase/case/index/role offsets. All6720 effective seeds must be distinct and
disjoint from archived blocks including v119. There are1344 latent trajectories
and2688 paired schedule runs:2 phases*21 cases*32 indices*2 schedules. Each run
has256 scored frames and16 initial labeled examples. All9 input coordinates are
visible to all arms,2448 coordinate reads per trajectory including initial data.

Retain immediate and delayed/missing schedules, delay0..31, missing.2, eligible
64/32-label caps and32-frame publication cadence. Earlier feedback arrives before
fitting; a current zero-delay label follows every forecast. Never treat missing
labels as negatives. Schedules of one latent trajectory are not independent.

Four workers run independent trajectories, with per-seed RNGs, detached state
and deterministic output order. Serial/parallel equality and race checks precede
generation. No dropped failures. Wall time includes parallel computation and is
not a serving-latency measurement.

## Fifteen arms

0..9 retain generic64,Boolean64,generic32,Boolean32,MAP64,MAP32,tree64,tree32,
variational64,variational32. Add:

-10 segment64 and11 segment32: constant per-frame boundary hazard.01; per-segment
 family prior.95 generic,.05 Boolean. Use the same labeled observations as the
 matched cap controls. The retained history starts at-16; excluded labels have
 unit likelihood, not recycled observations. Freeze the existing segment fitter.
-12 full-input four-expert Markov control: the existing.001 transition and prior
 .95 generic64 with the remaining.05 divided among the other three experts.
 Delayed labels refilter the issued-forecast journal, not refitted forecasts.
-13 no-change64 and14 no-change32: posterior predictive of the SAME generic/
 Boolean joint family model and prior as the segment arms, but with one segment.
 This isolates segmentation from simply changing the family combination.

Segment and no-change laws are fitted at the shared publication clocks and reused
until the next publication, like the other fitted experts. No between-publication
hazard aging is added. This is a stale published predictive, not exact inference
conditioned on every newly arriving label at each intervening clock.

## Criteria

Retain all2648 v119 screens on arms4..9. Score expected Brier on all256 and
terminal64 frames; retain expected accuracy/log loss, realized loss and teacher
floors. Paired intervals are mean +/-3.5SE over32 trajectories. They are
approximate fixed-sample screens, not anytime population certificates.
Non-harm permits upper loss increase<=.01; gains require mean>=.005 and lower>0.

Each segment candidate adds:
-672 non-harm screens against matched generic, matched Boolean, Markov12 and
 matched no-change13/14:21 cases*2 phases*2 schedules*2 segments*4 controls.
-96 terminal recovery gains:all6 changed transfer cases plus both Boolean
 switches, both phases/schedules, against matched generic, Markov and no-change.

Thus768 new requirements per candidate,1536 combined,4184 total. There are3696
non-harm and488 gain requirements. No stationary-gain claim is required for this
recovery hypothesis; all stationary protection controls remain. A broad segment
pass requires every one of its768 screens. Any individual pass is not a completed
seven-direction goal, partial-observer integration or agent validation.

## Artifact and audit

Write exclusive0600 JSONL:one header with source hashes and frozen constants,
then2688 records in identity order. JSONL avoids whole-file string-size limits.
Retain packets, forecasts, fitted-state diagnostics, segment boundary weights
and segment marginal log evidence. Require full exact replay of every record.

Independently reconstruct every score, teacher probability, seed, evidence origin,
MAP forecast, variational forecast and Markov forecast. Check all segment weight
normalizations. Independently reconstruct segment/no-change evidence, boundaries
and forecasts at index0, clocks0/128/224 in EVERY case, phase and schedule:504
cap-specific fitted states and16128 forecasts for each of segment/no-change.
Use a separate direct-probability interval implementation; do not call the Go
fitter. This is stratified numerical verification plus full replay, NOT a claim
of independent reconstruction of every segment forecast. Keep that limitation
visible in the results. No solver failure may become a skipped row or fallback.

No production/OpenClaw changes, new private data, deployment, push or whitepaper
promotion. The existing component performance artifact remains unchanged.
