# V71 Dynamic Dispersion Results

**PASS mathematical/implementation audits; FAIL the full scientific rescue.**
All seven whole research goals remain open. No production, paper, publication,
commit or push change. [Frozen protocol](mmm-dynvariance-v71-protocol.md).

## Complete Screen

All 18 declared arms, 40 consumed worlds, 20 regimes, two geometries and three
delay schedules: 2,160 arms including 1,920 model arms. This is a development
screen with one trajectory per cell, not untouched confirmation. Each paired
arm requests 400 second measurements of original outcomes, not 400 new outcomes.
Every original failed arm/cell and unchanged gate is retained.

| Arm | Issued Brier | Final Brier | Final top-10 usefulness | Worst complete core ms | Cells over 400 ms |
| --- | ---: | ---: | ---: | ---: | ---: |
| Full | .224699096 | .208151654 | .727150389 | 67.184 | 0 |
| Adaptive | .214747855 | .181598705 | .815037852 | 207.908 | 0 |
| baseline_no_pair | .348212362 | .348220407 | .509863671 | 105.052 | 0 |
| current_no_pair | .291882828 | .234268194 | .755230148 | 109.668 | 0 |
| current_uncertainty | .277368100 | .211769474 | .782570774 | 153.671 | 0 |
| free_no_pair | .271358559 | .213083716 | .786537442 | 101.853 | 0 |
| individual_no_pair | .262436406 | .211789714 | .793231859 | 346.270 | 0 |
| individual_uncertainty | .258671153 | .205783415 | .804750211 | 465.477 | 120 |
| local_no_pair | .248644065 | .202624691 | .803417124 | 676.329 | 120 |
| local_uncertainty | .245903739 | .199800776 | .807328674 | 907.963 | 120 |
| learn_no_pair | .271856086 | .212957450 | .786658247 | 105.139 | 0 |
| learn_random | .261542227 | .201641118 | .808033372 | 125.747 | 0 |
| learn_uncertainty | .256949823 | .200258541 | .807033372 | 157.597 | 0 |
| learn_information | .255968083 | .199801080 | .808533372 | 167.798 | 0 |
| learn_falsification | .255240816 | .199844019 | .808033372 | 165.890 | 0 |
| learn_predictive | .257241842 | .200312332 | .807033372 | 871.549 | 120 |
| learn_model_class | .261468372 | .201896948 | .809110553 | 161.095 | 0 |
| learn_noise_class | .258068777 | .199766381 | .810033372 | 161.173 | 0 |

ALL 16 model arms fail the mean-gain, per-cell Adaptive-harm and faster-recovery
gates. The best new quality arm, local_uncertainty, remains .031155884 worse
than Adaptive on mean issued Brier, has 111/120 harm cells over .01, and recovers
1.75 rounds slower on shifted worlds. It also fails the runtime gate in every
cell. No selected stationary subset or final-packet improvement replaces the
whole criterion. Its final top-10 usefulness still trails Adaptive.

## What The Controlled Comparisons Establish

Under common noise, learned dispersion improves uncertainty issued Brier
.277368100 -> .256949823 (absolute .020418277) relative to the fixed-current
family. This does not meet the strong-control gates. Without pairs, learning
.271856086 is slightly worse than the fixed-free family .271358559; learning
dispersion is not uniformly better than declaring a wider prior.

With independent member noise, sharing only the family improves uncertainty
.258671153 -> .245903739 (absolute .012767414) versus fully independent families
and noise. Replacing common noise by independent noise improves .256949823 ->
.245903739 (absolute .011046084) with the same shared-family alternatives.
These contrasts identify useful component differences on this cohort, not a
unique diagnosis of all harm or permission for Anti-Pigeon to merge members.

Common-noise falsification improves issued Brier by .006301411 over random and
.001709007 over uncertainty, but uses 1.338480803x and 1.012372214x total measured
core time. Equal 400 requests is NOT equal TOTAL cost. Goal 7 remains open.
Family-class sampling loses to uncertainty by .004518549; family/noise-class
sampling loses by .001118954. More direct class targeting did not rescue quality.

## Cost Evidence

Serial 150-member/16-trial microbenchmarks on Apple M4 / Go 1.27.1:
prediction 8.472-8.516 us; family/noise class query 27.490-28.678 us; all-target
predictive query 332.889-335.049 us. All report zero allocations per operation.
Constructor allocations are 1,508,832 and 2,009,056 bytes at 150/200 members,
both under 8 MiB. These are allocations, not RSS, and neither benchmark nor
complete synthetic core time proves loaded serving, persistence or freshness.

The separate phase audit shows local_uncertainty spending 35.66% on issue,
19.63% on first replay, 19.57% on proposal and 18.50% on snapshots. Its mean
complete core is 851.827 ms. The proof implementation recomputes member-noise
marginalization/global family odds, and the independent-member control validates
all local beliefs on reads. These are observed code paths plus measured phase
costs, not a unique profiler attribution. Global-weight/conditional-moment
caching is a performance lead, not a forecast-quality rescue.

All-target predictive nomination consumes 85.79% of its elapsed core time.
It neither wins the quality comparison nor passes 400 ms. Do not hide that
cost or adopt it based on its model-conditional expected value.

## Audits And Reproducibility

- Independent urn DP, dense transitions, explicit latent-Y likelihoods and
  separate branch-risk formulas agree in 108 configurations: 83,808 counted
  scalar comparisons, including hazards 0/1/16/1 and delays 0/3/11.
- Static full joint enumeration and actual branch posterior updates agree over
  64 histories, with 1,856 counted checks. Prior/emission checks add 1,911.
- Distinct-outcome identifiability, nonlocal family influence and wholly
  independent-member no-borrowing controls pass. Same-Y repeat measurements
  alone cannot identify dispersion with only one original outcome per member.
- Current-family limiting laws match the previous model within the unchanged
  numerical tolerance. No bitwise claim for different model arithmetic.
- Ownership, receipt, cancellation, cap, epoch, atomic-failure, zero-support,
  missing-response and future-boundary tests pass under race instrumentation.
  Three fixture future forks are nonvacuous; 29 artifact corruptions are rejected.
- An external full-journal API audit adds 42,840 counted scalar comparisons in
  36 configurations, including queries of old trials after 64 distinct member
  issues and late paired disagreements. Unit, race and vet pass. It ran during
  untimed independent replay, AFTER all measured collection/benchmarks; it did
  not change experiment source or any measured workload.
- ALL 1,920 model arms are independently replayed: 4,608,000 issued packets with
  clean and noisy laws, 9,216,000 scalar issued-law comparisons, plus query
  scores, choices, receipts, snapshots, metrics and cost arithmetic.
- All 240 Full/Adaptive arms are bitwise identical excluding measured time to
  V69; all 40 populations are identical. All eight frozen commands terminate 0.
  Frozen closure has 128 compiler/support files and saved generated test mains.
- Complete collection wall time 571.119 s; independent replay 1,851.335 s;
  readback 8.194 s. These are research-job times, not serving latency.

Structured artifacts: `research/dynvariance-v71-diagnostic/` holds freeze,
commands/logs, raw JSONL, independent readback and separate cost audit.
`research/dynvariance-v71-{initial,joint,independent,long}/` retains all preflight
attempts and source snapshots. Their earlier successful source versions remain
immutable; adding the independent-member alternative did not rewrite them.
The final chained checkpoint validates artifact hashes and protected tracked files.

## Next Action, Not Adoption

[Next lead](../../research/dynvariance-v71-next.md): retain the positive
independent-noise/shared-dispersion component; independently verify a cached
equivalent implementation rather than replacing its failed quality gate with
a timing success. In parallel, test whether shared mean calibration can exploit
the earlier V34 mean/variance components that dispersion-only integration omits.
Mean mismatch, sparse observations, noise sharing and static hyperparameters
remain competing explanations; this study does not select a unique cause.

Any expanded model must define a joint law, include nonlocal effects, record
approximation and total cost, pass the original whole gates, and retain all
negative controls. Useful valid splits, untouched outcome-labeled agent tasks
and durable loaded freshness still need their own evidence. No seven-goal
completion, exhaustion, fresh confirmation or scientific adoption is declared.
