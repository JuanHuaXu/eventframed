# Two-step rollout v9: local learning gains, FAILED full rescue

Completed1,792 fresh episodes and8,960 scored trajectories under
`LOOKAHEAD_ROLLOUT_PROTOCOL.md`. All arms used16 reports plus eight signal checks;
two-step planning did not receive extra evidence. Raw artifacts and hashes are
in `lookahead-rollout-v9.json.gz` and the summary JSON. Earlier v8 values were
design diagnostics only and were not counted as empirical observations.

## Confirmation outcome

The overall screen FAILED four conditions:

- Independent20 mean curve harm versus reliable-only was0.012919 (>0.01).
- Independent20 mean final harm versus reliable-only was0.011182 (>0.01).
- Independent20 curve gain versus one-step acquisition was-0.002029, not the
  required>=0.005; its paired z3.3 interval[-0.023079,0.019021] crossed zero.
- Misleading20 final gain versus reliable-only was0.066532 (above0.03), but
  its interval[-0.007386,0.140449] crossed zero.

All mean non-harm ceilings versus fixed passed across the seven confirmation
cases. This is not confidence-certified non-harm. The independent20 harm
intervals also span zero: these are failed mean protection gates, not a claim
of established population harm at the descriptive confidence level.

## Useful partial evidence

| Case | One-step curve Brier | Two-step curve Brier | One-step final Brier | Two-step final Brier |
| --- | ---: | ---: | ---: | ---: |
| Independent20 | 0.316048 | 0.318077 | 0.164265 | 0.166516 |
| Copied20 | 0.560099 | 0.508880 | 0.441952 | 0.443450 |
| Mixed20 | 0.490998 | 0.432003 | 0.294641 | 0.262055 |
| Matched05 | 0.185945 | 0.167193 | 0.038270 | 0.032915 |
| Matched20 | 0.404343 | 0.393625 | 0.266775 | 0.257093 |
| Matched random signal20 | 0.472004 | 0.449664 | 0.360709 | 0.348656 |
| Matched misleading signal20 | 0.415232 | 0.404024 | 0.296482 | 0.292438 |

Copied20 curve gain versus one-step was0.051219 with interval[0.008466,0.093971];
mixed20 curve gain0.058995 with interval[0.008454,0.109536]. Both curve contrasts
also had positive descriptive lower bounds in split0. These are evidence of
earlier useful learning in these finite populations, not universal improvement.
Copied20 final benefit was absent; mixed20 final gain0.032587 had an interval
spanning zero. Do not turn better early forecasts into a claim of established
terminal improvement.

## Cost and verification

The offline five-arm batch took82.693 seconds on this host, including all control
calculations and summaries. That is not per-request serving latency. Receding
two-step selection evaluates branches and continuations, roughly O(T^2 R H)
under this finite model, versus O(T R H) for one-step; equal report budget is
not equal computation budget.

`python3 -m unittest test_lookahead_rollout_v9 -v` passed both tests, including
complete replay of all1,792 episodes, source hashes, shared environmental tapes,
unique slots, forecast normalization and fixed costs. Replay took83.171 seconds.
The v8 four-leaf enumeration and boundary tests also passed; the v9 final-step
test enforces one-step selection when only one report remains. These checks
validate the tested computation, not empirical generalization.

## Next lead

Do not add another lookahead depth based solely on model value. The useful
effect is now localized to earlier learning with copied/mixed sources, while the
original independent-stream rescue remains unresolved. A cost-sensitive stop
or fallback can be tested against fixed acquisition and this explicit partial
result, without hiding the failed independent/misleading gates. Any selective
lookahead trigger must use observed state, never the simulator's source-mode
label. Production integration and realistic observation families remain open.
