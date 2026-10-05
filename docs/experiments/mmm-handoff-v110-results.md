# v110 handoff diagnosis: reproduced weighting lag

Status: consumed diagnostic PASS for reconstruction, not a quality rescue.
All128 v109 delayed switch trajectories and all seven original arms remain
bit-identical. V109 remains FAIL; no new candidate is promoted.

## Main finding

In confirmation delayed parity-to-majority, the log/no-neutral selector keeps
about84.2% mean advice weight on generic32 during frames224-255. Yet generic64
has mean full-input Brier .061463 versus .120100 for generic32. Only14.6% of
mean advice weight returns to generic64. The evidence gate leaves essentially
the same weights in this block: it does not reject a merely inferior model
simply because a better one exists.

The error remains when fixing the current forecast's view to the generic
observer's acquired mask:

| Confirmation block | Own-view mixture Brier | Generic-mask mixture | Generic control | Generic32 weight |
| --- | ---: | ---: | ---: | ---: |
| 128-159 | .420398 | .420116 | .381003 | 1.96% |
| 160-191 | .266990 | .265478 | .299015 | 49.79% |
| 192-223 | .116871 | .128529 | .166570 | 82.81% |
| 224-255 | .106802 | .104392 | .061285 | 84.20% |

At224-255, changing only the view removes .002410 of the .045518 difference
against generic; .043108 remains at the same view. The short model is useful
at192-223, then keeps most of the weight after generic64 improves. The design
phase shows the same pattern: final generic32 weight70.75%, with raw full-input
Brier .103913 versus generic64 .062214. These are aggregate diagnostics, not a
claim that every trajectory has the same cause.

Matched-view forecasts preserve the actual policy's preceding weight history.
They do not simulate a policy that used different observations all along.
Nevertheless, the instantaneous remaining gap does not disappear with the
generic mask or full input, supporting a weighting/handoff rescue over a
view-only change for this particular failure.

## Neutrality and the other direction

The log/neutral selector retains53.89% neutral mass at192-223 in confirmation
parity-to-majority. Own-view Brier .188554 remains .193488 at the generic mask,
versus generic .166570. Useful generic32 forecasts already exist (full-input
Brier .106128), but receive only29.74% mean advice weight. Neutrality helps
earlier and becomes an impediment here; it is not an unconditional rescue.

Observation choice still matters elsewhere. In confirmation majority-to-parity
at224-255, log/no-neutral Brier .068187 falls to .050205 under the generic mask.
Do not generalize the reverse-change diagnosis into a claim that observation
attention is solved or irrelevant. Both switches and both phases are retained
in the [complete summary](mmm-handoff-v110-summary.json).

## Timing and model movement

The trace records52,526 single-use log-advice updates;20,868 arrive across a
publication boundary. Each update is reconstructed from its actual issued
probability, label and .001 share, with no new-model rescoring or future label.
This confirms version crossing is common; it does not quantify its causal
contribution by itself.

For confirmation parity-to-majority at publication224, generic64 changes by
mean absolute probability .235106 and generic32 by .150272 over the512-input
reference law. Historical role identity therefore does not imply that fitted
forecasts stay close. The next proposed rescue transfers advice evidence in
proportion to declared old/new forecast compatibility, including pending losses,
rather than applying the unconditional time decay rejected in v108.

See the [compatibility-handoff proposal](../../research/compatibility-handoff-proposal.md).
Its modulation is a new empirical candidate, not a statistical certificate or
a theorem inherited from Hellinger/Renyi geometry. It must pass fresh quality
and stationary-protection gates before being considered a rescue.

## Verification

- Two consumed switch smoke replays passed under the race detector (2.549s);
  package vet passed.
- Complete128-record reconstruction and all seven original arms match exactly;
  generation18.28s and replay18.55s.
- All1,024 publications,65,536 traced forecast rows and52,526 updates are
  independently checked. Partial forecasts reconstruct as uniform marginals
  of full-input model tables; own/matched/full mixtures reconstruct exactly
  within1e-10 numerical tolerance. All64 summary cells contain1,024 frames.
- Model risk, inter-publication movement, label timing, versions, log updates
  and37 source hashes check. The evaluator reproduces the summary byte-for-byte.
- These offline reconstruction times are not serving performance or overhead.
  No oracle input enters the issued forecast, weight update or gate. No source
  policy, production state, whitepaper claim, commit or push was changed.

The [frozen diagnostic protocol](../../research/handoff-v110-protocol.md)
states scope and falsifiers. Artifact129,728,708 bytes, created with mode0600;
SHA256 `c358e5a5740fc4d3ca6f4571f6fdadc80065a0afbc2f43446f2524f629787af8`.
All seven research directions remain open.
