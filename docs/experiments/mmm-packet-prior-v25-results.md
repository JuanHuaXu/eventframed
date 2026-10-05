# Packet prior v25: contextual forecast passes; ranking remains partial

Date: 2026-10-02. The [frozen protocol](mmm-packet-prior-v25-protocol.md)
separates forecast and packet-ranking screens. The baseline-centred
Beta-Bernoulli predictive passes the forecast screen on both fresh splits
and both vector geometries. Its bounded rank correction passes the tight
geometry but fails the wide geometry's recovery requirement. These are
finite synthetic component results; all seven whole goals remain OPEN.
Production and the whitepaper are unchanged.

## Evidence and audit

- [Design](mmm-packet-prior-v25-design.jsonl): 256 worlds, 8,192 accepted
  ordinary outcomes; SHA256
  `98798a15ac4878ae853a65ac75a52309aab67405014f714622ad5664218c557b`.
- [Confirmation](mmm-packet-prior-v25-confirmation.jsonl): 256 different
  worlds, 8,192 outcomes; SHA256
  `9335b859b57d65391034404918acfa49aa3d5ccf9183ab77d2f3e8aa040cb89b`.
- [Independent summary](mmm-packet-prior-v25-summary.json) and
  [verifier](../../research/packet-prior-v25-verify.mjs) reconstruct all 76,800
  nominee laws, probabilities, packets, ordinary counts, as-of geometry,
  epoch bindings, source hashes, paired intervals and frozen decisions.
  `node research/packet-prior-v25-verify.mjs` passes.
- Pure-module tests pass under `-race`. Six actual-Service/LibraVDB/SQLite
  correctness worlds pass under `-race`; [output](mmm-packet-prior-v25-race.txt).
  They add no independent statistical sample. Package vet passes.

The prior is anchored to the actual pre-feedback baseline in a durable
journal, with strength two. Each of 32 monitored nominees receives one
ordinary Bernoulli outcome. The contextual predictive is `(2*b_ref+y)/3`;
the flat predictive is `(1+y)/3`. Both arise from an explicit common joint
model of past evidence and future outcome. Proposed laws are recorded
offline, not served or inserted into the current service's scored bundle.
The actual service remains the control; exact packet reconstruction is
required. No hidden utility enters the model or service.

## Forecast Screen

Expected whole-frontier Brier is lower-is-better. Each row averages 32
independent label worlds. Intervals are paired mean +/-3.5 SE, not
simultaneous confidence sequences. Improvement cases need gain >=.005
and positive lower endpoint; protection cases permit harm upper <=.01.

| Confirmation geometry / regime | Baseline Brier | Context Brier | Gain or harm [lower, upper] | Result |
| --- | ---: | ---: | --- | --- |
| Tight / independent | .407707 | .381320 | gain .026387 [.022248,.030526] | PASS |
| Tight / reversed | .419258 | .355849 | gain .063409 [.060362,.066456] | PASS |
| Tight / aligned | .396077 | .397961 | harm .001884 [.001520,.002247] | PASS within slack |
| Tight / calibrated | .092332 | .093785 | harm .001453 [.000918,.001987] | PASS within slack |
| Wide / independent | .316745 | .289990 | gain .026755 [.022451,.031058] | PASS |
| Wide / reversed | .412207 | .351562 | gain .060645 [.057727,.063562] | PASS |
| Wide / aligned | .213309 | .215676 | harm .002367 [.002021,.002714] | PASS within slack |
| Wide / calibrated | .187242 | .188914 | harm .001672 [.001152,.002191] | PASS within slack |

Design agrees: contextual gains .029174/.025289 on independent and
.061393/.059240 on reversed tight/wide. All design protection upper
endpoints are <=.002872. The forecast does not preserve the good baseline
exactly: it harms aligned and calibrated controls slightly but within the
frozen tolerance. Calibration is not universally improved.

Flat-prior prediction FAILS safety on both splits/geometries. Its calibrated
harm upper endpoints are .020043/.019754 design and .019655/.018227
confirmation, above .01. Larger gains in bad-baseline regimes do not excuse
this failure.

## Packet-Ranking Screen

These rank controls retain the actual service's scored law. The contextual
rank is `clip(b+.1*(context_predictive-b_ref))`, not the full posterior law.

| Confirmation geometry / regime | Baseline usefulness | Context usefulness | Gain or harm [lower, upper] | Result |
| --- | ---: | ---: | --- | --- |
| Tight / independent | .479375 | .674375 | gain .195000 [.127005,.262995] | PASS |
| Tight / reversed | .124161 | .237953 | gain .113792 [.101620,.125963] | PASS |
| Tight / aligned | .875839 | .870805 | harm .005034 upper .008495 | PASS within slack |
| Tight / calibrated | .924917 | .924895 | harm .000022 upper .000040 | PASS within slack |
| Wide / independent | .511250 | .651875 | gain .140625 [.089316,.191934] | PASS |
| Wide / reversed | .124161 | .140403 | gain .016242 [.009707,.022776] | FAIL: mean below .02 |
| Wide / aligned | .875839 | .871846 | harm .003993 upper .007211 | PASS within slack |
| Wide / calibrated | .923676 | .923380 | harm .000296 upper .000513 | PASS within slack |

Design wide reversed gain .015839 also fails. Both rank proposals fail that
gate. Their packets differ in 11/128 wide design and 15/128 wide confirmation
worlds; tight packets are identical in both splits. The geometry therefore
matters. Current learning packs zero monitored nominees in tight worlds,
but often packs monitored nominees in wide worlds; the zero-exposure finding
is not a universal serving property.

## Cost and Remaining Work

The pure [benchmark](mmm-packet-prior-v25-benchmarks.txt) computes 32 posterior
predictions in 69.70-70.48 ns per operation with zero allocations on Apple M4.
It excludes storage, selection, sorting, packing and full Recall. Maximum
observed sequential Recall times are 21.32 ms design and 15.68 ms confirmation;
outcome calls 12.61/12.46 ms. These are not loaded p99 or freshness guarantees.

This static per-event, one-label study does not establish shifts, source
independence, delayed evidence, nonlinear generalization, sharing, prospective
agent-task improvement or loaded learned-state continuity. Synthetic selection
and omission certificates assume coverage, rather than prove it.

Next test a query-scoped immutable pre-feedback prior through the actual
served/scored law and durable journal, with residuals disabled initially.
Reusing an event posterior across queries or silently moving its anchor would
invalidate this model. Full-posterior ordering is a distinct successor to the
failed bounded delta, requiring fresh cohorts and unchanged protection gates.
