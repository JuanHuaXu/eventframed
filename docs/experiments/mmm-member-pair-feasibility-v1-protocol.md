# MMM member-pair feasibility v1: frozen protocol

This isolated research audit asks whether the paired-betting component
from [v4](mmm-paired-bet-v4-results.md) receives enough valid evidence
under the existing `integrationLabel` MMM member generators. It does
not change the actual member harness, models, serving path, or paper.

Use `integrationLabel` and the existing five generator cases `stable`,
`member_shift`, `common_shift`, `recurring`, and `null` with 512 clocks.
Draw independent live and reference 9-bit contexts and independent label
noise. `member_shift` changes only the live label rule at clock 256;
`common_shift` changes both; `recurring` changes live over clocks
128..255 and 384..511. The declared context cell is bit 2 of the
9-bit frame. A valid pair has the same bit-2 cell and the same origin
clock; no full-frame equality or cross-clock joining is inferred.

Compare two acquisition modes on separate, fresh random streams:

- `passive`: for each nomination, take one independently drawn reference
  context. Use the pair only if bit 2 matches the live context.
- `targeted`: for each nomination, draw reference contexts until bit 2
  matches. Count **every** reference-context proposal as an acquisition
  read. Draws occur before revealing outcomes; cap attempts at 512 and
  fail the trial if the cap is reached without a match.

Both modes nominate outcome-independently with probability 0.25 per
clock. Per nominated side, independently drop labels with probability
0.20 and delay surviving labels by an integer-uniform 0..31 clocks.
Count all arrived labels as cost, even for unmatched or incomplete
pairs. The gate sees only complete, matched pairs at the maximum arrival
clock; partial labels never enter its state or later nomination.

Freeze the v4 candidate unchanged: starts `0,64,...,448`; for each
start, bit-2 cell and sign, multiply wealth by
`1+0.8*(sign*(Y_ref-Y_live)-0.10)` when a complete matched pair with
origin at or after that start arrives. Start wealth is one and future
starts stay one. Flag when the arithmetic mean of 32 processes reaches
50. No threshold, rate, restart or cell is fitted to this generator.
Report flags and prechange flags, median/p95 flag clock, matched and
delivered pair counts, observed labels, live and reference context reads,
nomination counts and isolated experiment wall time. No forecast is
scored here: this is a prerequisite boundary test, not Goal 3 completion.

Use independent seed bases `2026100701` design and `2026100702`
confirmation, 1,000 trajectories per case/mode/split. Preserve all
trial records and a source/protocol hash. The primary transfer screen
requires <=20/1,000 flags in `stable`, `common_shift`, and `null` for
each mode, zero prechange flags in `member_shift`, and >=800/1,000
member-shift flags by clock 511 for targeted acquisition. A mean of
more than 2.1 reference-context reads per targeted nomination fails
the cost screen. Passive power and recurring behavior are reported
without a pass floor. Do not retune against design or confirmation.
An independent verifier must reconstruct summaries, and confirmation
must replay exactly apart from timing.

The conditional-mean/e-process claim is restricted to this generator's
independent per-clock noise and outcome-blind acquisition. Bit-2
matching identifies an observational *conditional marginal* contrast,
not an intervention effect. A historical latch does not certify current
divergence after a recurring regime reverts. If targeted acquisition
fails, the next candidate must change the observation policy or the
declared hypothesis representation, not silently lower the evidence
threshold.
