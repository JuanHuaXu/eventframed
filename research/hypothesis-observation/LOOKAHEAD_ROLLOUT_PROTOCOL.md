# Two-step acquisition rollout v9

Frozen after v8's design-state diagnostic and BEFORE fresh outcome evaluation.
V8 warranted a rollout under its predeclared rule; its model-implied gains are
not empirical results. No v8 states or v7 seeds are confirmation data here.

Seven unchanged v7 cases, two splits,128 episodes per case, new seed base
202609190613 plus the existing split/case/episode offsets. Five arms: fixed,
reliable, fixed-acquisition averaging, one-step mixture acquisition and two-step
mixture acquisition. Same source tapes/signals and cost24 (eight signal checks,
16 reports). Two-step is receding-horizon: maximize exact V2 at decisions0-14;
at decision15 use one-step because only one report remains. No extra reports
or oracle reliability/target labels. Current forecast is emitted before the
chosen report arrives.

Confirmation gates:

- Mean curve/final harm <=0.01 against fixed in every case.
- Mean curve/final harm <=0.01 against reliable-only in independent20.
- Misleading20 final gain against reliable-only >=0.03 and paired z3.3 lower >0.
- Independent20 curve gain versus one-step mixture acquisition >=0.005 and
  paired z3.3 lower >0.

Report comparisons to every control and both splits. These normal bounds are
descriptive, not simultaneous certificates. Retain every failed gate. A pass
would establish finite equal-report-budget evidence, not equal computation cost,
real-source authentication, real-agent benefit or general-domain success.

Verify all prefixes reselect from recorded past only; last-step horizon rule,
unique slots, shared environmental tapes, valid distributions, source hashes and
complete replay. Keep the four-leaf v8 value test. Record whole offline batch
elapsed time separately from deterministic replay; no serving latency claim.
Exclusive-create artifacts. No production adoption or push.
