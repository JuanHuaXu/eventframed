# Compatibility handoff component

Subsequent [v111 quality comparison](../docs/experiments/mmm-compatibility-v111-results.md)
FAILS both candidates. The component-only results below precede that experiment
and remain valid, but do not establish a successful quality rescue.

2026-09-13. Implemented in the research-only observation learner package.
Component checks PASS; quality remains UNTESTED. V109 remains FAIL and v110
remains a consumed diagnosis, not a rescue. No production integration,
whitepaper promotion, commit or push.

## Patch reasoning

- Confirmed observation: v110 reproduces stale short-window weighting in both
  phases of the delayed reverse switch, including at matched observation masks.
- Candidate explanations include acquisition, insufficient model capacity,
  neutral dilution and stale role evidence. Matched/full-view diagnostics and
  later useful raw forecasts support the last explanation for this block;
  acquisition still matters in the other switch direction.
- Recommendation only: forecast compatibility could preserve useful evidence
  better than unconditional age decay. Neither the diagnosis nor the geometric
  motivation proves that this rescue improves scored Brier.
- Scope: new private research component, not an upstream released bug fix.
  Frozen v108/v109/v110 kernels and artifacts remain unchanged. No data repair
  or production change is involved.
- Falsifier: fresh stationary protection or recovery comparisons fail, or
  the extra publication work outweighs the demonstrated gains. Do not tune
  the exponent, prior or thresholds on consumed v110 trajectories.

## Implemented contract

The [proposal](compatibility-handoff-proposal.md) now has two executable
variants: publication transfer alone, and publication plus pending-loss
transfer. Both start from v109 log/no-neutral and retain the same comparative
gate, acquisition law, publication cadence, expiry and fixed-share step.

Publication computes Bernoulli affinity over the declared common input law,
checks that this input law agrees across publications, and raises affinity to
the predeclared 32-frame scale. It transfers normalized log evidence relative
to the original prior. Zero-prior roles stay excluded; identical forecasts
preserve the existing policy exactly. Changing the input law is rejected
atomically rather than silently using a new measure.

The pending variant caches log products for each of at most eight issue
versions. An arriving label uses its immutable issued probabilities and a
fixed neutral-reference log likelihood ratio. It never rescores that label
with the new model. The neutral reference is not a competing expert. Old-version
evidence still cannot update the new version's comparative tests.

These are generalized evidence-transfer heuristics. Affinity is not a
posterior validity probability, and product modulation does not establish a
statistical certificate for overlapping fitted windows or delayed feedback.

## Tests and bug hunt

- Geometry: exact identity, symmetry, analytic parity-to-neutral affinity under
  uniform and nonuniform common input laws, and rejection of changed input law.
- Transfer: explicit independently calculated normalized-weight formula;
  invariance to adding 1024 to all stored log weights; recoverable underflow;
  all-one identity, all-zero prior reset and excluded-prior preservation.
- Pending update: independent probability-domain likelihood-ratio calculation
  for both outcomes; exact old-update identity; all-zero evidence still applies
  the declared share; invalid coefficients/probabilities/future origins fail
  without mutation.
- Complete unchanged-model journal parity: both variants, 256 frames, all eight
  publications, delays 0/8/31 and full settlement match the original log
  journal state exactly at every clock.
- Changing-model journal: two publication boundaries compound correctly;
  immutable old probabilities drive arrival updates; old evidence does not
  enter new tests; duplicates and expired labels fail; expiry does not duplicate
  evidence; changed-measure publication and failed acquisition roll back;
  reentrant publication/delivery fail, followed by successful normal issuance.
- Initial rollback fixture failed because it requested failure on a second read
  after this policy had legitimately stopped at the first. A trace showed
  `err=nil, calls=1`; the test now fails the guaranteed first read. No policy
  behavior was changed to satisfy that incorrect fixture assumption.

Final combined compatibility/log/age tests under the race detector PASS in
6.623s; package vet PASS. These test deterministic lifecycle and numerical
contracts, not a quality claim or a concurrent multi-owner API.

## Isolated performance

Command:

```sh
go test ./internal/observationlearners -run '^$' -bench '^BenchmarkCompatibilityJournal$' -benchmem -benchtime=300ms -count=3
```

Darwin arm64, Apple M4, Go benchmark suffix 10. No other research test or
verification process was running during this batch. All 18 measurements are
retained in [the raw output](../docs/experiments/mmm-compatibility-component-benchmarks.txt).
One operation includes 256 adaptive forecasts, delay-8 feedback, expiry scans,
eight publications and full flush. Fitted models are prebuilt outside timing.
Moving mode alternates parity and neutral tables on each publication.

| Model sequence | Selector | ms / 256-frame lifecycle | B/op | allocs/op |
| --- | --- | ---: | ---: | ---: |
| unchanged | original log | 1.553-1.574 | 151480-151483 | 1241 |
| unchanged | publication handoff | 1.964-1.976 | 165040-165041 | 1241 |
| unchanged | plus pending transfer | 1.969-1.990 | 165040-165041 | 1241 |
| alternating | original log | 1.594-1.599 | 155065-155066 | 1255 |
| alternating | publication handoff | 1.940-1.968 | 155632-155633 | 1168 |
| alternating | plus pending transfer | 1.925-1.948 | 155632-155636 | 1168 |

Identity behavior is identical, so its approximately 26-28% lifecycle increase
is extra component work in this fixture. Moving-mode observation paths differ;
its totals are not a pure same-work overhead estimate. These are neither
database query latencies nor tail-latency guarantees. Affinity costs four times
512 full-input evaluations per publication, plus at most eight times five
cached-coefficient updates; each arrival reads five cached coefficients.
Input dimension is fixed at nine here, not a scalable full-domain enumeration
claim. This component performs no persistence or network I/O.

Source SHA256:

```text
aa6499a18b007e544e9b709073320ec8d47b221e8855eb574e5162ed690a1948 compatibility_advice.go
ef68cc2dd721373c04cad73c3a9dde08c214abee4bfd15a00b0c17a768286043 compatibility_advice_test.go
06de71f24042c545f799ca9ec758e943084ece68445ed764d3b9f2887af922bb compatibility_advice_benchmark_test.go
```

Next: freeze a fresh paired quality experiment including both variants, the
original log and Brier controls, both switch directions, every stationary
scenario and both immediate/delayed schedules. Retain all protection and
recovery gates. All seven research directions remain open.
