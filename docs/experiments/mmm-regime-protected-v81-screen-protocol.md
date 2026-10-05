# V81 protected support: public development screen

Frozen before execution. This is a lead screen, NOT the original full120-cell
native Full/Adaptive comparison, fresh confirmation, agent outcome test,
loaded-serving test, or equal-TOTAL-cost acquisition test. All seven goals OPEN.

## Fixed design

Two development seeds 2026105501 and2026105503 (not reserved confirmation seeds).
50 members,16 rounds,800 original issue slots per trajectory. Member baseline
repeats .30,.45,.65,.85. Four generators: stationary; abrupt inversion at round8;
gradual inversion from round4 through round12; recurring inversion every4 rounds.
Latent labels are independent Bernoulli draws at the generator's current rate.
This is intentional model misspecification, not sampling the learner's joint law.
Noise0,.1,.2 flips two independent measurements of the SAME latent label.
Immediate arrival: first at issue, second at issue+1. Delayed arrival: first at
issue+50, second at issue+100. Exactly the first25 members of each round receive
a second measurement, independent of every label, forecast and query result.
All queued packets are drained after issue800 without introducing new slots.
48 trajectories. No case is removed or tuned after seeing results.

Three arms: V80 top-weight cap36; V81 protected cap36; V81 protected cap9
reset-only negative control. All use global reset1/16, member hazard.25, the
same baseline vector, and the same issue/measurement/arrival journal. Arm order
rotates by case. A static baseline-rate reference is scored offline and is NOT
the standing native Full or Adaptive control. Every arm receives every scheduled
packet, even if its hypothetical query abstains or a prior reveal was rejected.
Second queries occur immediately before the actual second reveal. A failed
first reveal can make the second phase unavailable; record both failures rather
than fabricating a first measurement. No posterior refresh or retry is allowed.

## Boundaries, scoring, costs and falsifiers

Issue sees only member and current clock. Future labels/true probabilities are
held by the harness, never passed to the learner. First/second values cross the
boundary only at declared arrival. Score issued receipts before feedback with
expected clean Brier p(1-p)+(q-p)^2 and expected observed-first Brier using
p_obs=eta+(1-2eta)p. Record realized first Brier too. Aggregate over ALL800
issued forecasts, not only accepted evidence. Report stationary paired harm
and shifted paired gains, arrival/rejection/query counts and per-case costs.
No confidence or calibration claim is inferred from this small development set.
Later native confirmation remains required regardless of screen outcome.

Measure constructor plus ALL public Issue, Snapshot/token access, Pending,
Reveal and queue traversal operations in each arm's core timer; total Go
allocation is measured across that whole arm. Scoring, file serialization and
generator work are outside core and measured separately by the runner's wall
clock. End Snapshot copy is inside core. These are serial, not loaded latency,
peak RSS, background freshness or equal-cost acquisition. Record every arm's
timing including rejection/abstention; never call query-only speed whole-core.

Falsifiers: issue errors, different input journals across arms, future packets
delivered early, output ranges violated, mutated source, missing case/arm/row,
or independently recomputed loss/counters differing. Unit tests cover schedule,
common journal, deterministic rerun and future-suffix forks. A bounded prefix
fork proves only that fixture, not whole-code future-independence. Preserve
negative results and all failures. No acceptance threshold is loosened.

Decision: reject a quality-rescue interpretation if cap36 stationary clean harm
exceeds.01 in any case or if mean shifted gain is non-positive. Passing this
screen alone is still insufficient for goals1/2/4/7. Cap9 is expected to destroy
memory; its outcome cannot be substituted for a useful adaptive rescue.

No production/private/sealed/reserved labels, whitepaper or publication changes.
