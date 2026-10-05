# Approximation and Inference Correction

Date: 2026-10-05. Research-only mechanism corrections, not scientific adoption,
fresh confirmation, a production deployment, or a change to earlier verdicts.

## Mathematical and Runtime Scope

`internal/researchbounds` implements the whitepaper Appendix C current-state TV
recurrence, safe-discard condition, triangle composition and same-outcome binary
Brier envelope. `researchregimelogsummary` now uses the evidence-aware recurrence.
It does not change the retained support, likelihood, forecast model or default
cap. The large audited approximation defect remains **0.6982459966723692**.

`IssueBounded` checks the post-selection law transactionally. Failed admission
does not append history, bump epochs or corrupt a warm prefix cache.
`PendingCleanBounded` checks each conditional replay separately, returning only
clean future forecasts. Historical-query evidence-ratio weights are deliberately
not returned by the bounded API: a current-state certificate cannot cover their
historical latent paths. Legacy research calls remain uncertified.

A named numerical bound is mandatory but its name is NOT proof. The caller must
justify a whole-operation bound, including nondirected input/envelope error,
likelihood normalization, replay and predictive arithmetic. One supplied scalar
for a query must cover both conditional branches separately, not their average.
No valid external numerical witness was established by this work. Zero-bound
unit/benchmark witnesses are explicit synthetic wiring assumptions only.

## Verification and Bugs

An independent review found positive underflow in scalar bound arithmetic;
computed zero now rounds upward and exact-zero cases are handled explicitly.
Exact-rational regressions include the former 2/9 counterexample. The bound
tests also compare 4,096 finite filtering cases to rational reference arithmetic
and check the uniform safe region and single-coordinate Brier inequality.

The separate inference utility had a subnormal planning defect (512 instead of
604 for equal smallest-positive SE and effect); ratio-first evaluation repairs
it and rejects unsupported intermediates. The pilot AP planning case stays 137.
The confidence sequence is a prospective union-Hoeffding utility, not a
retroactive confidence statement for the old 32-stream screens. See the
[inference contract](inference-contract-2026-10-05.md).

The full capped-filter reference audit retained 193,613 checks, 48,780 summary
parity comparisons and 280 dense envelope comparisons; 225 envelopes were
nonvacuous. The shared forecast arithmetic and measured maximum joint defect
remain unchanged. These are development fixtures, not untouched replication.

## Matched Performance

Reproduction: run `node research/approximation-admission-2026-10-05/run.mjs`
from the software root. The runner snapshots control source from immutable
revision `7a7b9357c06ce855795b947ffcde2fd61a82d3ca`, pins candidate source hashes,
and runs six samples per arm/profile in before-after then after-before order.
It refuses empty/incomplete benchmark output. Full extraction uses the separate
immutable `1a7edb62b6be4031fd01ebeab8071b17303a7815` test-only reconstruction.

On the M4, Go 1.27.1, one Go benchmark CPU, the twelve prepared issue/query
profile medians changed by -1.92% to +0.47%, with unchanged allocation counts.
This is descriptive timing variation, not proof of zero overhead or speedup.
The scalar bound took 25.66--25.91 ns/op with zero allocations.

| Members, 64 prepared rows | Legacy issue median | Bounded issue median | Extra allocation |
| --- | ---: | ---: | ---: |
| 2 | 246.992 microseconds | 242.477 microseconds | 1 |
| 150 | 523.639 microseconds | 527.521 microseconds | 1 |
| 200 | 620.407 microseconds | 620.865 microseconds | 1 |

The bounded-issue benchmark uses maximum budgets of one and a synthetic zero
numerical witness to measure mechanics; it does not establish usable fidelity.
The allocation is the copied prefix slice. Its cost scales with retained prefix
count; this short-history fixture is not a long-history memory bound.

The final extraction measurement checks five profiles at 256 B, 2 KiB and
16 KiB per role, with three forward and three reverse repetitions. Unquoted
median latency increased 4.54--29.26%; quoted-only profiles increased
138.00--186.26%, also reflecting intentional fallback path changes. Candidate
profile medians range 39.879 microseconds to 11.294 milliseconds. Repeated quote
masking is still linear in supplied text and allocates; mask reuse is a separate
potential optimization, not implemented here. No loaded service, retrieval,
network, observation-acquisition, fitting or database latency is measured.

Raw benchmark logs, structured summaries and source hashes are under
`research/approximation-admission-2026-10-05/`. Earlier extraction audit failures
and measurements are retained separately in the
[turn fallback audit](turn-fallback-2026-10-05.md).

## Deployment Boundary

The two extraction bugs are repaired: quoted captured bytes cannot be promoted
as direct field evidence, and pronouns cannot be accepted as named actors.
The sibling query/text paths and identity enrichment have regression checks.
An old multiline identity test was updated to retain its quote/reference safety
purpose without requiring the now-prohibited direct field extraction.

New extraction changes the 5W1H corpus. Historical quality numbers are not
validation of this revised extractor. Prospective replay/agent-outcome validation,
calibration compatibility, a genuine numerical witness and useful omitted-bound
tightness remain necessary before claiming a deployable research integration.
No production process, datastore or OpenClaw installation was touched.
