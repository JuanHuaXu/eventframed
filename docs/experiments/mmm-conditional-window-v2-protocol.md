# Conditional window power v2: frozen rescue protocol

The v1 cumulative gate failed to flag a conditional swap introduced at clock
256. This fresh-seed study tests whether old-prefix dilution, rather than
the audit budget alone, caused that failure. The method is a research-only
binary-context certificate, not an integrated MMM split path.

Keep v1's two fair contexts, reference law `(0.9,0.1)`, 512-clock horizon,
`epsilon=0.10`, `delta=0.02`, 25% outcome-independent audit nominations,
batch-512 versus online reference, immediate versus 20%-missing/delay-0..31
schedules, and stable/start-swap/clock-256-swap cases. Use independent seed
bases `2026100401` design and `2026100402` confirmation, 1,000 trajectories
per strategy/schedule/case/split. Pair stream roles across gate arms.

Three gate arms see identical arrived nominated evidence:

1. `cumulative`: v1's live prefix and cumulative reference.
2. `window256`: cumulative reference, but only live audits whose *origin*
   lies in `[clock-255, clock]`. A late audit outside that window is ignored.
   This fixed deterministic window requires no change-point knowledge.
3. `oracle_reset`: cumulative reference, live evidence only from origins at
   or after the hidden clock-256 swap. It exists **only** in the clock-256
   case and is an unattainable upper bound, not a candidate policy.

Each arm checks the same certificate as v1 after every clock, using only
arrived evidence. The repeated-test radius remains
`sqrt(log(8*(T+1)/delta)/(2*n))` for each reference/live context mean.
For a fixed origin window, the selected sample set is independent of
outcomes under the stationary null, so the same finite-horizon union bound
applies. Context labels, reference samples, audit selections and arrivals
are all independent of outcomes by construction. The oracle reset is not
eligible for a deployable claim. Missing, future, duplicate and un-nominated
audits must fail closed.

Primary screen for each reference/schedule cell: `window256` flags <=20/1,000
stable trajectories, flags >=800/1,000 start-swaps, and flags >=800/1,000
clock-256 swaps with zero prechange flags. Report detection median/p95,
reference/live resource counts, and the cumulative/oracle results even if the
candidate fails. A pass would only establish certificate power in this
finite iid setting, not an MMM split, downstream Brier gain, source validity
or loaded latency. Preserve every trial row, an independent summary verifier,
and deterministic replay. Do not tune the window width on these seeds.
