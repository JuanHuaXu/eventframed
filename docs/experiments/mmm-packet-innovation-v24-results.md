# Packet innovation v24: useful exposure rescue, protection screen fails

Date: 2026-10-02. The normalized [frozen protocol](mmm-packet-innovation-v24-protocol.md)
completes in design and fresh confirmation: 48 worlds, 7,200 nominee laws and
1,536 ordinary event-local training outcomes. Every durable-vector, cosine,
service-baseline, journal, epoch, posterior and actual-packet reconstruction
check passes. **Both proposed rank corrections fail the primary rescue screen**
because the aligned-harm interval exceeds the .01 ceiling in both splits.
The component is promising; Goals 5/6 and all seven whole goals remain open.

## What changed in the test

Unlike v22/v23, every stored seed vector is normalized. Each nominated durable
vector is hydrated and checked for unit norm and angle-oracle agreement, and
the actual service baseline is checked against the declared cosine, confidence,
priority and recency formula. The flawed earlier tapes remain preserved with
errata. V24 uses fresh seed bases 2026102403 and 2026102404.

All 150 nominees now have a hidden Bernoulli usefulness law, including the
unmonitored replacement candidates. Thirty-two monitored events supply one
past outcome each. Their accepted Beta(1,1) updates, actual forecast laws and
epochs are checked directly. Neither proposed rank formula receives hidden p.
The baseline ordering, actual current ordering and two proposed orderings are
repacked over the same events, recall cap, token budget and provenance policy.
Replaying current order exactly reproduces each real `Service.Recall` packet.
The rank controls keep the actual learned scored-law bundle unchanged; their
packet Brier differences reflect selection, not a replacement forecast model.

## Actual packet utility

Higher utility is better. This is expected usefulness of an independent future
Bernoulli outcome, averaged over the ten packed events, not an LLM answer score.
The flat and context proposals produce identical packets in every recorded
world. Their different score magnitudes therefore cannot be ranked by this
fixture's packet outcome. Intervals are mean +/-3.5 SE across eight worlds per
regime, as frozen; they are descriptive and not simultaneous coverage claims.

| Split / regime | Baseline order | Current order | Either proposed order | Proposal gain interval |
| --- | ---: | ---: | ---: | ---: |
| Design / independent | .49250 | .52250 | .68000 | .18750 [.08681,.28819] |
| Confirmation / independent | .44750 | .53000 | .65750 | .21000 [.09775,.32225] |
| Design / aligned | .87584 | .70403 | .86852 | -.00732 [-.01786,.00323] |
| Confirmation / aligned | .87584 | .70403 | .86987 | -.00597 [-.01185,-.00010] |
| Design / reversed | .12416 | .29597 | .23255 | .10839 [.07558,.14120] |
| Confirmation / reversed | .12416 | .29597 | .23195 | .10779 [.07450,.14107] |

Both proposals clear the .02 mean-gain and positive-lower-endpoint requirements
for independent and reversed populations. The aligned mean harm is .007315
and .005973, but its upper endpoint is **.017862 and .011846**, above .01.
Do not round either into a pass. The current policy has .171812 aligned utility
harm and the same-size reversed gain; blindly displacing early nominees helps
when the initial ranking is reversed and harms when it is already useful.

Every actual current packet contains zero of the 32 monitored candidates.
The posterior/rank mismatch persists on the corrected unit-vector fixture:
their initially high retrieval baseline is blended with a flat-prior mean
of 2/3 or 1/3, so even a positive first outcome lowers its forecast and rank.
Elasticity changes the magnitude without repairing that reference mismatch.
The proposals instead centre their rank movement on the original flat prior
or on the baseline prior; positive observations can promote and negative ones
can demote. This is a proposed search-order rescue, not an accepted serving
policy or a claim that the contextual prior is calibrated.

The journaled whole-frontier expected Brier still improves in all three regimes:
design independent .406800->.399026, aligned .396077->.394835, reversed
.419258->.403061; confirmation .407306->.398559, .396077->.394836 and
.419258->.403098. Those forecast improvements coexist with the current packet's
large aligned utility loss. Whole-frontier score and delivered packet utility
must both be measured; one does not establish the other.

## Exact conditional model diagnostic

The separately recorded [post-hoc analysis](mmm-packet-innovation-v24-oracle-protocol.md)
does not change the failed primary screen. Score separation proves that either
proposal selects the first ten positive monitored events, then unmonitored
fillers if fewer than ten are positive. An exact dynamic program integrates
that decision over every possible 32-label vector. It passes twelve small
exhaustive-enumeration controls, probability-mass checks and score-separation
checks on all 48 recorded populations.

For the declared aligned law, exact expected proposal utility is .871843899,
versus baseline .875838926: **.003995027 expected harm**, or about .40 percentage
points. For reversed it is .232507834, a .108346760 gain. On the eight independent
hidden populations in each split, conditional expected proposal utility averages
.689268105 and .667561024, versus baselines .49250 and .44750. This locates the
empirical safety failure in finite outcome sampling under this known model;
it does not supply a real-world envelope or make the algorithm safe across
unknown generators, correlated evidence or shifted priorities.

## Audit and cost

- [Design tape](mmm-packet-innovation-v24-design.jsonl): SHA256
  `1db73214785085d549b5538455496d1adbc616e5693cc4b590172e2ae7b2c5da`.
- [Confirmation tape](mmm-packet-innovation-v24-confirmation.jsonl): SHA256
  `a11300023fcccd99ed569d88bcb3150c108c88785d38019c969b0f5706d1e888`.
- [Independent reconstruction](../../research/packet-innovation-v24-verify.mjs)
  checks 13 source/protocol hashes, all scored laws, innovations, utility,
  false-item counts, packet identities, Brier and frozen gates. It passes.
- The full design run under `-race` and the zero-evidence/direction controls pass;
  package `go vet` passes. The [race verifier](../../research/packet-innovation-v24-race-verify.mjs)
  confirms every non-timing field in all 24 worlds equals ordinary design.
  This replay contributes correctness evidence, not 24 independent new worlds.
- [Exact-model verifier](../../research/packet-innovation-v24-oracle.mjs) and its
  [machine-readable output](mmm-packet-innovation-v24-oracle-summary.json) pass.
- Sequential Recall maxima are 14.86/17.56 ms and outcome-publication maxima
  11.61/12.08 ms in design/confirmation. These are unloaded finite diagnostics.
- On Apple M4, both arithmetic formulas over 32 members take 186.4-187.4 ns
  per benchmark operation, zero allocations; [raw benchmark output](mmm-packet-innovation-v24-benchmarks.txt)
  is retained. This excludes posterior reads, sorting, packing and persistence.

Next use a predeclared broader geometry that separates the proposals, more
independent outcome trajectories and calibrated-prior controls before actual
serving integration under visible writes. The current epoch guard, as-of rules,
selection/provenance validity, loaded freshness and untouched agent-task outcomes
remain required. Synthetic certificates confer no external coverage. Production,
OpenClaw and the whitepaper have not changed.

Reproduce with `node research/packet-innovation-v24-verify.mjs`,
`node research/packet-innovation-v24-race-verify.mjs` and
`node research/packet-innovation-v24-oracle.mjs`.
