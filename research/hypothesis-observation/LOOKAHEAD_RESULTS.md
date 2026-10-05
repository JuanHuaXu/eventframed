# Two-step diagnostic v8: fresh rollout warranted

Inspected448 pre-outcome states from112 design episodes: first16 seeds per case
in v7 split0, prefixes0,4,8,12. This is reused design data, not confirmation.
`lookahead-v8.json` records all candidate scores, hashes and the screening result.

Independent20 mean model-implied two-step value gain was0.011422, with meaningful
action changes in28/64 states (43.75%). Misleading20 mean gain was0.014285, with
changes in30/64 states (46.875%). Both exceed the frozen>=0.002 and>=10% rules
for warranting a fresh rollout. These are NOT empirical Brier improvements.

The value compares two first actions with the SAME two-report budget and optimal
one-step continuation, not two observations against one. Differences concentrate
at prefixes4/8: independent mean gains0.019779/0.025909, misleading
0.020669/0.036472. Neither target case changed meaningfully at prefixes0 or12.
The four states from an episode are dependent; no independent-sample interval is
claimed. Conditional model value may fail under misspecification or changed
state occupancy in a real rollout.

All three tests in `python3 -m unittest test_lookahead_v8 -v` passed, including
complete diagnostic replay, parent/source hashes, prefix-forecast agreement,
empty/single-action boundaries and direct four-leaf terminal-value enumeration.
Selection does not mutate the model. The diagnostic took1.079 seconds on this
host, including prefix reconstruction; that is not a serving latency benchmark.

Proceed with a fresh receding-horizon rollout,16 total reports and a one-step
last decision. Freeze empirical success gates before execution and retain all
previous failed experiments. No production behavior changed.
