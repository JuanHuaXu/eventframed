# V64 Actual-Mixture Quality Results

Compute gates PASS; scientific rescue FAILS. No adoption and no whole-goal
completion. Complete 600-arm diagnostic: 40 consumed worlds, 20 regimes,
2 generators, 3 delay schedules, 5 arms. There are 96,000 distinct latent Y
outcomes, not 1.44 million independent outcomes from repeating arms.

| Arm | Issued Brier | Terminal Brier | Top-10 expected usefulness | Max full loop |
| --- | ---: | ---: | ---: | ---: |
| Full | .224699096 | .208151654 | .727150389 | 65.91 ms |
| Adaptive | .214747855 | .181598705 | .815037852 | 205.98 ms |
| Bank, no pair | .246098185 | .223498380 | .763924395 | 172.82 ms |
| Bank, random | .243557234 | .222001034 | .770790963 | 194.57 ms |
| Bank, uncertainty | .243324222 | .221895794 | .774300636 | 196.79 ms |

These are expected CLEAN-label Brier and utility under declared generators;
they are not untouched real-agent accuracy or authentication of physical truth.

## What Improved And What Failed

Compared with V60's single-window counterparts, issued-risk gains are .005728845
(no pair), .005373346 (random) and .005093518 (uncertainty), with wins/losses
112/8, 111/9 and 109/11 across 120 cells. Retaining several windows is useful
relative to that rejected learner, but still trails both stronger controls.

The unchanged >=.01 gain-over-Full gate fails for all three bank arms. They
also fail the no->.01-Adaptive-harm gate in 104/100/101 cells respectively.
Stationary mean harms are .031559987/.029077820/.028970113, with maximum
.065322036/.061592603/.061592690. Recovery remains 2.383333 rounds slower than
Adaptive on average across 60 shift cells; there is no recovery gain over the
bank without paired observations. Keep this negative result, not merely the
relative-to-V60 improvement.

All bank complete core loops are below 400 ms. They include constructor,
first scheduling, all issued forecasts, expiry, first receipts, nomination,
request/delayed-secondary queues, cancellation, snapshots and draining. They
exclude backend, RPC, persistence and concurrent serving. V63 separately passes
the 8 MiB constructor gate at BOTH 150 and 200 members; neither result completes
loaded serving/freshness goal 6.

Random and uncertainty each request 48,000 secondary packets overall. Random
receives 46,947 and misses 1,053; uncertainty receives 46,995 and misses 1,005.
Shortest-window-expired receipts are 2,154 and 2,301. They still incur cost and
can grade the selector, but cannot resurrect expired child factors.
Total full-loop cost: random 21.893 s, uncertainty 21.955 s. Candidate nomination
cost is 1.688 ms versus 48.775 ms. The .000233012 uncertainty gain over random
is NOT a falsification or equal-TOTAL-cost goal 7 win.

Observed W1 Brier is .240918192/.241108827/.241060092 for no-pair/random/
uncertainty; observed W2 Brier is .070808210/.080071808 for random/uncertainty.
W2 measurements concern already-observed Y and are selectively nominated, so
these scores are not interchangeable with CLEAN next-Y risk or task utility.

## Audit Trail

All 240 Full/Adaptive cells are bitwise unchanged from V60 excluding timing;
all 40 populations are identical. Independent reconstruction covers all 360
bank arms and 864,000 issued packets, checking clean AND W1 forecasts per packet
(1,728,000 scalar comparisons), requests, receipts, choices, missingness,
per-window expiry, snapshots, metrics/recovery and cost accounting.

The first preflight failed because an unscored raw arm was supplied to the
metric checker. Its frozen source/log and local repair are preserved. The
subsequent original audit failed on an endpoint near-tie; original data was
NOT rerun or changed. A separately frozen checker verifies independent laws
and scores plus exact stable sorting of the verified recorded scores.
25/1,920 uncertainty decisions have ambiguous reference order. Maximum score
difference 4.1550096696596484e-14; reference cutoff regret 1.5718863968723348e-14.
Do not call those rankings bitwise equal to independent floating-point arithmetic.
All nine corrupted-field checks and nonvacuous future forks still reject/pass
as intended after that repair.

Collection: `research/retention-v64b-diagnostic/diagnostic.jsonl`.
Original failed audits: `research/retention-v64-diagnostic` and
`research/retention-v64b-diagnostic`.
Completed additional replay/readback: `research/retention-v64c-audit`.
Streaming readback preserves metric definitions without retaining both entire
corpora in RAM. This is a consumed n1-per-cell diagnostic, not fresh confirmation.
Sealed task labels and reserved design/confirmation seeds remain unopened.

The V65 counterfactual audit separately disproves transferring a single-window
tower identity to this working mixture. A calibrated/coherent observation-value
rescue remains to be implemented and tested. All seven WHOLE goals stay OPEN.
