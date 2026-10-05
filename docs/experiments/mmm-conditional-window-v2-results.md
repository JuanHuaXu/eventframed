# Conditional window v2: batch-reference component pass, full transfer fail

The [frozen protocol](mmm-conditional-window-v2-protocol.md) passes its
certificate-power screen **only** with a 512-label precollected reference.
The full study fails because the same 256-clock window loses too much power
when the reference itself must be learned online. This is a finite two-context
component result, not an MMM split or Goal 3 completion.

## Fresh confirmation

Each row summarizes 1,000 independent 512-clock trajectories. The cumulative
and window arms see identical streams and audit arrivals; the reset diagnostic
is available only for the clock-256 swap. `W` is the fixed 256-clock window.

| Reference / schedule | Stable W flags | Start-swap W flags | Clock-256 flags: cumulative / W / reset | W median flag clock |
| --- | ---: | ---: | ---: | ---: |
| Batch 512 / immediate | 0 | 1,000 | 6 / **997** / 996 | 463 |
| Batch 512 / sparse delayed | 0 | 1,000 | 6 / **917** / 928 | 479 |
| Online / immediate | 0 | **654** | 0 / **460** / 474 | 500 |
| Online / sparse delayed | 0 | **95** | 0 / **54** / 55 | 504 |

The fresh design split agrees: batch-reference late-swap window flags are
995/1,000 immediate and 923/1,000 delayed, versus cumulative 13 and 2;
online-reference late-swap window flags are 472 and 40, versus cumulative
zero. No clock-256 case flags before the change. The batch delayed condition
still needs a median 223 clocks after the change to flag, leaving little
time for a corrected forecast within this horizon. Online immediate/delayed
reference uses about 128/99 arrived labels by horizon, versus 512 extra
pre-live labels in the batch condition. The online window also discards
useful start-swap evidence: its confirmation detection rate falls from
cumulative 989 to 654 immediate, and from 743 to 95 sparse-delayed.

The protocol calls `oracle_reset` an "upper bound." That wording was too
strong: hidden-change access makes it an optimistic **diagnostic**, not a
samplewise mathematical upper bound on detection count or clock. In the
batch-immediate confirmation, the window flags 997 trajectories and reset
flags 996. This mismatch is preserved rather than normalized away.

## Audit and limits

- [Design rows](mmm-conditional-window-v2-design.jsonl), SHA256
  `10628185c24506976b5f2fbea4756c7f762e710fbb4d0d8f31906f940dce4f30`.
- [Confirmation rows](mmm-conditional-window-v2-confirmation.jsonl), SHA256
  `dedd16ba3222b4c16dba707728cf9b89b7d0b8fc017c0f35ea8d6f735bde73a2`.
- The [simulator](../../research/conditional-window-v2.mjs) keeps separate
  source/protocol hashes and all 24,000 design/confirmation trial rows. The
  [independent summary verifier](../../research/conditional-window-v2-verify.mjs)
  checks unique trials, evidence accounting, all 24 cell summaries and hashes.
  A fresh confirmation replay matches all 12,002 rows exactly after excluding
  elapsed wall time. Negative controls reject future, missing, un-nominated
  and duplicate evidence and check window expiry.
- The full confirmation simulation took 2.15 seconds, including generation,
  three gate arms, audit queues, assertions and JSON output. This is not a
  Go or loaded serving-latency benchmark. The research duplicate-origin sets
  retain up to one horizon of IDs; indefinite retention is not specified.

The repeated-check radius uses [Hoeffding's bounded-variable inequality](https://www.tandfonline.com/doi/abs/10.1080/01621459.1963.10500830)
and a finite union over two contexts, two streams and 513 check times. Its
false-flag interpretation requires fixed contexts and outcome-independent
audit/arrival selection under the stationary null. It does not certify
dependent sources, adaptive context construction, multiple active buckets,
or an unknown EventFrame target-law diameter. It tests neither the current
scalar Bayesian nomination boundary nor the effect of a split on MMM's scored
law. Goal 3's downstream benefit and all seven whole goals remain open.

The current research `memberRun` forms nomination from
`Monitor.Observe(referenceCorrect, liveCorrect, true)` and then requires that
nomination before either scalar gate can authorize a split. That monitor has
no context argument. In the constructed equal-marginal conditional swap,
scalar correctness alone cannot systematically nominate the change.
An integrated experiment must add a separately budgeted, independent-audit
nomination path; routing the conditional alert through the existing scalar
nomination Boolean would leave the new gate inert. The independent audit,
not a selected posterior update, must retain final Anti-Pigeon authority.

Next research should address the reference-evidence budget and window
adaptation jointly, then require a fresh integrated forecast comparison. Do
not tune this fixed window on these confirmation trajectories. Production,
the whitepaper and remote repositories are unchanged.
