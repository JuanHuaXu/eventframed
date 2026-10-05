# Tracking-preserving delayed selector diagnostic

## What this isolates

Replace the failed static Brier base with the existing fixed-share likelihood
primitive, without tuning its .002 share rate or .7/.1/.1/.1 prior. The new
script checks128 forecasts and updates against actual Go `bayes.ForecastMix`
output, including extreme expert probabilities. It also reconstructs every
original immediate-feedback forecast on all192 recorded trajectories within
1e-12, not merely their mean scores.

Revocation at the recorded Anti-Pigeon split time caps the long-slot weight in
all replicas; pre-revocation pending credits cannot update post-revocation
weights. Slots remain origin-bound, one-use and age32-expiring. Publication
alone does not invalidate algorithm-role loss; that is an experimental policy,
not a calibrated posterior interpretation. Original model fits and acquisition
remain fixed. All records are consumed, including the cohort named confirmation.

## Results

Second-cohort delayed post-change Brier:

| Case | Original | Single arrival selector | Replicated selector |
| --- | --- | --- | --- |
| Copied bit | .279590 | .260005 | .349289 |
| Copied XOR2 | .279574 | .259701 | .350944 |
| Noisy-copy bit | .280556 | .259742 | .356538 |
| Reversing XOR2 | .273903 | .253957 | .324355 |
| Stable | .047330 | .047604 | .048003 |
| Null | .255241 | .253376 | .256702 |

Immediate feedback matches original for both variants. Replication still fails
to rescue delayed recovery despite retaining the stronger primitive. The single
arrival selector has roughly .020 lower Brier in the four changed cases on
these fixed tapes, but stable Brier worsens slightly. No adoption gate or
population guarantee is inferred from these means.

## Evidence and verification

Implementation: `research/delay-tracking-selector.mjs`.
[All results](mmm-delay-tracking-selector-v1.json),
[Go primitive reference](mmm-tracking-mix-reference-v1.jsonl),
test driver `internal/observationgate/tracking_mix_reference_test.go`.
384 schedule-runs, same parent SHA as the previous replica screen:
`1533a049c2146cc7839799e9a394691d96c3fdf4a1a8e8d1c504db86053bf8e6`.
Duplicate/stale-token and detached-advice checks remain; original-immediate
parity and Go primitive comparisons pass. Repeat artifact byte-identical.
Go reference test passed .225s; JS whole diagnostic .79-.83s, not serving cost.
The previous static Brier regret bound is deliberately not asserted for this
fixed-share likelihood variant.

## Why the apparent arrival gain is not a rescue

[v104](mmm-delayed-v104-results.md) already tested stale role-credit carry in
full learning/acquisition and failed both switching cases despite stable-parity
improvement. That is stronger contrary evidence than this fixed-advice replay
can erase. The present result differs in generators, learners, observation masks,
and gate behavior. It does not establish which difference explains the sign.

Do not rebrand stale carry as new or promote it on these tapes. The useful next
question is whether its apparent benefit disappears when it controls acquisition
or when tested on v104's changing distributions. A consumed factorial comparison
of selector-only versus coupled acquisition would distinguish this, before
spending fresh confirmation data. Preserve all old failures and original gates.

All seven goals remain open. No production, whitepaper or remote changes.
