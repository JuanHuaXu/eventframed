# Researchstats utility handoff

These generated records accompany [the short JSON report](../../docs/experiments/inference-utility-2026-10-05-report.json)
and [the inference contract](../../docs/experiments/inference-contract-2026-10-05.md).
They are deterministic synthetic utility checks, test/benchmark records, and
retrospective planning from archived aggregate contrasts, NOT scientific
confirmation. No old verdicts or paper source were edited. Parent integration
and copying into the paper's evidence tree are separate.

- `synthetic-check.txt`: actual verbose test output, including the Appendix C
  formula match, synthetic paired Brier interval, exact archived AP member
  contrast (137-unit approximate plan), and the negative recurring AP contrast.
- `tests.txt`: focused tests/example, race/coverage and vet command records.
- `planning-regression-before-fix.txt`: captured failing subnormal cancellation
  regression (512 instead of 604), retained as numeric defect evidence only.
- `benchmarks.txt`: three serial one-CPU repetitions, with observed iterations,
  ns/op, bytes/op and allocations/op; excludes forecasting/training/acquisition.
- `source-sha256.txt`: implementation, tests, contract and archived aggregate
  summary digests. It contains hashes, not a copy of the summary or private data.

Run from the repository root to check source provenance:

```sh
shasum -a 256 -c evidence/researchstats-2026-10-05/source-sha256.txt
```

The utility requires a fixed conditional mean and frozen alpha/comparison family.
One unit is a completed stream mean or an independent fit-cluster mean, not a
clock. The single-fit AP pilot's observed variance omits random-fit uncertainty.
Both old AP intervals include zero; 137 is an approximate normal planning
calculation, not AP confirmation, guaranteed power, or an anytime stopping budget.
The refreshed report also records the corrected 604-unit equal-subnormal case.
The planning formula remains approximate statistically and is evaluated in
float64, not formally outward-rounded; unsupported intermediates are rejected.
