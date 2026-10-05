# Window competition with observation feedback

Frozen next test, not run yet. Retain all128 consumed switch trajectories and
both immediate/delayed schedules. Requested-view result is primary. Do not tune
window length, weights, share rate or quality thresholds from previous failures.

Two candidate arms:

- Fixed mask: replay the exact parent requested mask; reproduce the verified
  window-competition candidate, not merely its final average score.
- Coupled mask: use `windowGuidePredict` with that arm's own pre-outcome old-
  inner, window-inner and outer states. Its predictions decide which proxy
  observes the next frame; its own acquired advice receives subsequent feedback.

Reconstruct base from the frozen generator and validate setup. At original fit
clocks, rebuild label64 count/subset, event64 count/subset, local and pooled models
from the original arrived random-audit origins. Publication happens after the
clock's prediction and feedback. No model sees the current outcome beforehand.
Both arms may share immutable fitted objects, but never weight/journal state.

Learning timing matches the previous experiment: predict first, process arrived
nonmissing outcomes in origin order, then apply the independently monitored
parent split at the end of that clock. Update old-inner only for advice issued
when its models were available; update window-inner and outer from their issued
vectors. After split, reject origins<=split clock, preserve persistent inner
roles, and use the existing outer revoke. Do not regenerate delayed advice from
current models or use the simulator's true change point as a learner signal.

Monitoring/auditing is exogenous to these candidate weights in the parent.
Its as-of paid mask may seed a PRIVATE per-arm reader cache, validated against
the captured monitor mask/values. Other experimental arms' reads cannot leak
into it. The scored requested mask contains only bits that the chosen observer
actually requested; the cache must not enlarge a Reader response.

Record observer-requested cost and incremental unique-coordinate cost beyond
already-paid monitoring separately. Include unchanged monitoring/audit costs
when comparing TOTAL observation work. Do not label unchanged audit schedules
as proof of unchanged total cost. Report fit counts, support, runtime, peak or
retained model memory, and allocation costs separately from forecast arithmetic.

Required checks before scoring: fixed-arm per-frame reproduction to1e-12;
same X/Y/missing/arrival/audit sequences; exact fit origin sets/publication clocks;
issued journal binding; no future/missing label use; all masks/values/caps; cache
isolation; correct split ordering; deterministic replay and source hashes.

Frozen exploratory gates versus ORIGINAL requested control: reverse post-Brier
mean gain>=.005 and mean-3.5SE>0 in allfour cohort/schedule cells; full/post upper
harm<=.01 in all16 cells. Also report coupled-versus-fixed differences with
trajectory intervals, accuracy, neutral comparison, changed masks and total cost.
Neither fixed nor coupled control may be weakened after observing results.

This tests a declared proxy acquisition heuristic with a separately scored
ensemble. It is not coherent joint-law value-of-information optimization and
does not inherit the joint observer's identities. A pass still requires broader
generators, untouched confirmation and real agent utility/serving evidence.
