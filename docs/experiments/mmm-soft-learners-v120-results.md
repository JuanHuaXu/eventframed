# Segment-membership comparison v120

Status: **FAIL** for both segment candidates under the frozen broad criteria.
Generation, independent verification and full exact replay complete.
No production adoption, whitepaper promotion or completed research direction.

## Evidence

- [Frozen protocol](mmm-soft-learners-v120-protocol.md)
- [Raw JSONL](mmm-soft-learners-v120.jsonl)
- [Independent summary and every gate](mmm-soft-learners-v120-summary.json)
- Generation: 2688 schedule runs, 1344 latent trajectories, 688128 frames,
  6720 distinct effective seeds, 176 source hashes; 620.69 seconds. This is
  parallel research wall time, not serving latency.
- Exact replay: all2688 records and176 hashes match, 628.04 seconds. The
  independent summary, static-atom diagnostic, bound audit and composition
  diagnostic each reproduce byte-for-byte in separate executions.
- Raw SHA256: `5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f`.

The independent verifier reconstructs every score, 1376256 MAP forecasts,
1376256 variational forecasts and 688128 Markov forecasts. It checks 43008
segment weight normalizations. A separate direct-probability implementation
reconstructs 504 stratified fits and 16128 forecasts each for the segment and
no-change models, not every segment forecast. Maximum errors: log evidence
1.85e-13, boundary weight 1.37e-14, segment forecast 1.20e-13, no-change
forecast 9.44e-15. Numerical agreement is not quality validation.

## Frozen outcomes

| Candidate | Non-harm | Recovery gains | Overall |
| --- | ---: | ---: | --- |
| Segment64 | 493/672 | 19/96 | FAIL, 512/768 |
| Segment32 | 432/672 | 6/96 | FAIL, 438/768 |

All 4184 retained and new checks yield 2034 passes. No incumbent challenger
passes its own complete criteria either. Counts are not independent trials or
probabilities of success. Non-harm permits an upper loss increase of .01; it
does not certify zero harm. Gain requires mean improvement at least .005 and
a positive lower endpoint. Intervals use the frozen mean +/-3.5 SE over 32
trajectories, not an anytime or population-wide guarantee.

Segment64's non-harm counts against generic64, Boolean64, Markov and no-change64
are respectively 140/168, 119/168, 111/168 and 123/168. Its recovery gain counts
against generic64, Markov and no-change64 are 8/32, 5/32 and 6/32. Segment32
passes zero of 32 gain requirements against Markov.

## Where segmentation helps and hurts

Confirmation, delayed/missing schedule, terminal64 expected Brier (lower better):

| Case | Generic64 | Markov | No-change64 | Segment64 |
| --- | ---: | ---: | ---: | ---: |
| Additive stationary | .221845 | .221755 | .221954 | .222753 |
| Additive gradual | .239102 | .239065 | .238771 | .232713 |
| Hierarchy gradual | .248073 | .244456 | .247004 | .240660 |
| Local-table gradual | .254006 | .250387 | .253292 | .247081 |
| Parity4 stationary | .070807 | .049066 | .048143 | .057144 |
| Majority3 stationary | .054958 | .054983 | .054959 | .065497 |
| Majority to parity | .127403 | .060262 | .096959 | .055906 |
| Parity to majority | .108775 | .102272 | .109095 | .078300 |

In confirmation delayed parity-to-majority, Segment64 improves on Markov by
.023973, interval [.010325,.037620], and on no-change64 by .030796,
[.010572,.051019]. In delayed majority-to-parity, its gain over no-change64 is
.041053, [.018786,.063319]. Those two no-change comparisons also pass in the
design phase. This supports a narrow benefit of segment membership inference,
not a broad rescue. Transfer-family gains do not consistently pass both phases
and every control.

The stable-pattern cost is real: confirmation delayed parity4 all256 loses
.020111 against Boolean64, interval [.012596,.027625]. The corresponding
design interval is also strictly harmful. The displayed terminal averages
must not be confused with that all256 comparison.

## Interpretation and next falsifier

The component is numerically consistent with the declared model. Quality
failure is not evidence of an arithmetic bug. Plausible contributors include
unnecessary segment boundaries, limited post-boundary evidence, family-prior
cost and stale 32-frame publications. These are hypotheses, not established
root causes.

A bounded next diagnostic is an explicit no-change model atom competing with
the segment model through their marginal likelihoods on exactly the same
eligible evidence. It differs from the earlier expert-switch hazard mixture:
the latent variable here changes which observations share model parameters.
It must not count shared labels twice or accumulate overlapping-window Bayes
factors as independent evidence. A useful falsifier is whether stationary
protection improves while the two-direction switch gains survive. Consumed
v120 data may diagnose that tradeoff but cannot confirm a selected rescue;
any resulting candidate needs fresh seeds and unchanged broad comparisons.

## Consumed static-atom diagnostic

The [separate diagnostic](mmm-segment-v120-static-atom-diagnostic.json) evaluates
one equal-prior no-change/segment mixture. At each publication, compute
`w = Z_segment / (Z_segment + Z_no_change)` on the same selected labels, and
publish `w*p_segment + (1-w)*p_no_change`. The weight is recomputed from the
original prior, never multiplied across overlapping windows. Direct sequential
Beta integrals independently reconstruct the no-change marginal and one
predictive at all43008 fits; maximum predictive error is1.98e-14.

| Diagnostic | Non-harm | Recovery gains | Overall |
| --- | ---: | ---: | --- |
| Static atom64 | 609/672 | 24/96 | FAIL |
| Static atom32 | 498/672 | 4/96 | FAIL |

This is a post-outcome diagnostic on consumed v120 data, not another untouched
confirmation. Its64-label confirmation-delayed terminal losses are .050939 on
parity4, .058016 on majority3, .053823 on majority-to-parity and .076995 on
parity-to-majority. The latter retains a gain of .025278 against Markov,
interval [.011950,.038605]. However, parity4 all256 still loses .013591 against
Boolean64, interval [.008042,.019140], and majority-to-parity improvement over
Markov still lacks a positive lower endpoint. Stable protection improves but
is not solved; do not promote it based on a larger pass count.

Average segment weight at terminal publications is around .3-.4 in stable
Boolean cases but rises above .88 at clock192 in both switch directions. This
supports investigating boundary uncertainty, but does not identify it as the
only cause. Early family learning and retained evidence remain plausible
sources of the residual cost. No post-hoc case filter or parameter sweep was
used in this diagnostic.

### A finite-window limitation

Let D be the number of possible boundaries between the first and last selected
label. Under the fixed hazard h, the probability of no boundary between them
is q=(1-h)^D. On that event, all selected labels share one parameter draw,
giving the same evidence Z0 as the no-change model. Missing and excluded
labels contribute unit likelihood. Therefore Z_segment >= q*Z0 and the
equal-prior model weight satisfies w >= q/(1+q). Boundaries outside the selected
label span integrate out; the bound uses that span, not the entire retained
history length.

The [bound audit](mmm-segment-static-atom-bound.json) checks all1344 archived
averaged fit weights against the corresponding averaged lower bounds. It does
not claim individual-weight checks or a lower bound on forecast error. In
confirmation delayed parity4, clock224, the64-label mean lower bound is .29194
and observed mean weight .34364; the32-label values are .38351 and .44372.
Thus even ideal stationary evidence within this short window cannot make the
equal-prior segment weight vanish. Its internal no-change hypotheses can still
forecast well, so weight persistence alone does not prove prediction harm.

A next research lead is a persistent prequential comparison using each issued
forecast and each later eligible outcome once, rather than accumulating
overlapping-window marginal likelihood ratios. Earlier expert-mixture failures
remain relevant controls: the new ingredient would be segment-membership
forecasts, not a claim that the old switching algorithm has become valid by
renaming it. Retain delayed/missing feedback, all stationary controls, both
switch directions, source separation and exact replay. This is an untested
composition lead, not yet a rescue. The simplest existing mixer was then tested
below; it does not deliver that rescue.

## Consumed prequential composition

The [composition diagnostic](mmm-segment-v120-composition-diagnostic.json) keeps
the existing .001 Markov transition and .95 generic64 prior. Compare the raw
four-expert mixer, its addition of variational64/32, and its addition of
segment64/32; the remaining prior mass is divided equally. Each issued forecast
is fixed before its outcome. Delayed eligible labels refilter the original
forecast tape exactly once per origin, with unknown emissions equal to one.
The reference reconstructs the full bounded prefix, not a production journal.

All2064384 forecasts run, the four-expert control matches all688128 archived
Markov predictions, and1008 current/future/unavailable-label poison checks pass
with the issued tape held fixed. This does not retest the generation of that
tape or an adaptive observation policy. Exact reproduction of the diagnostic
also passes separately.

The segment-augmented mixer passes168/168 non-harm cells versus the raw mixer,
but0/168 cells meet mean gain >=.005 with a positive lower endpoint. Against the
variational-augmented mixer it passes162/168 protection cells and again has zero
qualifying gains. These are cell comparisons, not the full standalone gates.
The six protection failures concern additive cases. Confirmation delayed
terminal reverse-switch gain versus raw Markov is only .003471,
[-.004454,.011395], despite the standalone segment's much larger gain. Thus
the existing conservative mixer does not reliably realize the specialist's
headroom. Adding the same mixer is not a successful rescue; arbitrary prior
retuning after this result would require fresh confirmation.
