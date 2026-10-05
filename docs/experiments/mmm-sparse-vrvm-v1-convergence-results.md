# Sparse variational convergence: optimization improves, quality not rescued

## Frozen comparison

Both phases, all 21 scenarios, both feedback schedules, index 0, clock 128,
latest eligible 64 labels: 84 fits and 2,688 forecasts. Only the cap changes
from 64 to 1,024 iterations; priors, initialization, equations, evidence and
stopping tolerance are unchanged. The default wrapper still uses 64 steps.

Collection passed in 367.05 seconds (367.415 seconds package time). This is
total parsing/fitting/prediction/serialization time for the diagnostic, not
serving latency. Full replay passed in 370.83 seconds (371.158 seconds package
time); `cmp` confirms byte-identical raw output. Isolated fits on the same
64-label fixture cost 4.167-4.251 seconds at the extended budget versus
295.259-296.917 ms at 64 steps. Allocated bytes are about 128.68 MB versus
9.05 MB per fit, not peak resident memory. Three one-fit repetitions are
retained in `mmm-sparse-vrvm-v1-convergence-benchmarks.txt`. Neither is an
end-to-end performance result.

## Result

57/84 fits meet the original numerical bound-change criterion; 27 remain
capped. Iterations range from 688 to 1,024, median 949.5. Every final bound
improves on its original 64-step state, by 2.60-5.66 (mean 3.63). The largest
last-step increment is still 8.84e-5. Numerical stopping is not a proof of a
stationary point or global optimum.

Expected Brier, lower is better. Each row averages 21 cases over forecasts
128-159. Means and case counts are descriptive, not independent trajectory
confidence intervals or full recovery/stationary-protection gates.

| Phase | Feedback | 64 steps | Up to 1,024 | Generic64 | Boolean64 | Markov | Improved / worsened cases |
|---|---|---:|---:|---:|---:|---:|---:|
| 0 | Immediate | 0.26452 | 0.27255 | 0.18434 | 0.18974 | 0.17890 | 10 / 11 |
| 0 | Delayed/missing | 0.26000 | 0.25843 | 0.18255 | 0.18975 | 0.17932 | 12 / 9 |
| 1 | Immediate | 0.26527 | 0.26340 | 0.19528 | 0.18305 | 0.17616 | 12 / 9 |
| 1 | Delayed/missing | 0.27064 | 0.25812 | 0.18921 | 0.18791 | 0.18154 | 18 / 3 |

The mean plug-in ablation is retained in the JSON and is worse than the
integrated law in all four aggregates. Generic/Boolean forecasts use the same
publication clock and eligible window; Markov additionally updates mixture
weights with subsequently arriving evidence. It is a system control, not an
identical update-schedule ablation.

Convergence progress is real but does not rescue forecast quality. The cap
alone cannot explain the weakness, including cases that meet the numerical
stop. No promotion, full broad-quality run, or production change is justified.

## Audit and tests

All 84 extended traces begin with the EXACT original 64-step trace. Origins,
query features, evaluator outcomes and all 15 original controls match the
pilot. Independent unscaled Gauss-Jordan reconstruction verifies every final
Gaussian, Gamma/xi state, moments and log determinant. All 2,688 integrated
probabilities match independent 32,768-panel midpoint checks within 2.05e-10.
The audit verifies recorded bound monotonicity, not every intermediate bound
by independent recomputation; earlier algebra tests remain separate evidence.

Budget/default-equivalence, prefix, extended future-data isolation, and the
existing numerical/as-of tests pass under race (22.404 seconds package time).
An initial new test file had a missing closing brace; it was repaired before
any experiment ran. No fit equations or protocol values changed in that repair.

Artifacts: `mmm-sparse-vrvm-v1-convergence.jsonl`, `-audit.json`, `-summary.json`;
auditor `research/sparse-vrvm-v1-pilot-audit.mjs` with mode `convergence`;
summary builder `research/sparse-vrvm-v1-convergence-summary.mjs`.
The auditor's default mode still supports the original 504-fit pilot.
Its compatibility rerun reproduces the original audit JSON byte-for-byte
(`mmm-sparse-vrvm-v1-pilot-audit-compatibility.json`). Final complete sparse
race suite passes in 22.551 seconds. No experimental processes remain running.

SHA-256:
- Raw and replay: `4ebfa32d309395cd85467c7c8cbc9f0517e1e13af6ebeae71d2c045204469e95`.
- Audit: `ca7147c8b8330a92148a2221f694c8c8468368ff318fda7a0cb87663b09284cc`.
- Summary: `eab65afd0b03a5048f618b87ec8250ff16cfc1a4346ad116a83c42018f9f9377`.
- Extended auditor: `7420a61e4f0596e18ea7d054174768640ff6c9f099ba57331bf3713f8d98e7e6`.

## Next direction

Do not run a cap grid or treat faster optimization as a prediction rescue.
`research/sparse-vrvm-acceleration-source-note.md` records SQUAREM as a future
numerical tool, not an adopted learner. The more relevant unresolved causes
are prior/approximation mismatch, representation and regime dynamics.

`research/sparse-vrvm-prior-diagnostic.md` derives a prior-only check for the
existing Gamma precision prior and records a finite-slab/global-local source
lead. That check passes against a Cauchy special case, independent panel counts,
and an analytic upper bound. Only 0.00145% of the original coefficient prior
lies in [-1,1], exposing how extreme the supposed weak prior actually is.
This does not establish that the posterior has the same behavior or prove
the cause of all forecast errors. Replacing the prior requires its own frozen
contract and inference derivation, not a post-hoc probability clamp. No goal
is complete. Production, dependencies and whitepaper remain untouched.
