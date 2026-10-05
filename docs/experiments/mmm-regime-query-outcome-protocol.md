# Joint-query actual-outcome diagnostic

Frozen before implementation/results. Use all2688 consumed v120 schedule runs:
21 cases, two phases,32 indices, both schedules. This is an isolated one-decision
diagnostic, not a completed full-stream policy or fresh confirmation. It tests
whether the verified query-value calculation predicts useful actual learning.
All previous full-direction success criteria remain unchanged.

## Decision and controls

At clock160, construct the existing as-of63-label view and unknown pool[152,160).
Use eight visible probes[153,160], never future inputs, hidden labels or Q.
Compute the exact batched joint values. Four arms: no query, pseudorandom query,
maximum query-label entropy, and maximum joint predictive-risk reduction.
For nonempty pools all three paid arms spend exactly one unit, even if natural
feedback subsequently makes the purchase redundant. Complete schedule pools
must be empty, with no paid costs and identical forecasts across all four arms.

Random selection uses ascending SHA256 priority over
regime-query-outcome-v1:phase:case:index:clock:origin, lowest priority selected,
origin tie-break. Omit schedule for paired latent priorities. Entropy is binary
entropy of the query's posterior label mass, not a current-event stand-in.
For maximum utility/entropy, choose lowest origin within1e-10 of the maximum.
No fitted thresholds, rank scaling, prior sweep or teacher-aware selection.

## Real update and scoring

Record all decisions before reading any paid Y. At clock161, reveal the chosen
past outcome and admit all naturally arrived labels with origin<161, including
clock160 zero-delay feedback. Each arm fits the same existing full segment model
using its latest64 known labels. Keep hazard.01 and generic prior.95. The paid
label enters the actual likelihood; it does not merely update expert weights.

Natural feedback can be newly available and the64-label cap can change support.
This post-reveal publication is therefore NOT claimed to equal the frozen
63-plus-one hypothetical conditional law used during acquisition. Record both
the query prediction and the actual outcome effect. Count purchases that were
already naturally available at161 and actual training-origin changes per arm.

From that publication, emit forecasts for161..191, aging by the declared hazard
between frames without further refitting in this isolated diagnostic:
p_t(x)=.5+.99^(t-161)*(Q_161(x)-.5). Future inputs are applied only after choosing
the query and fitting. Q and actual future Y are evaluator-only data. The next
full-stream experiment must reintroduce regular updates and repeated decisions.

Retain every identity, base support, pool, query masses/values, selected origins,
charged costs, redundant purchases, actual support sets, and all four forecast
sequences. Compare expected Brier against no-query, random and entropy, separately
by phase/case/schedule. Mean +/-3.5SE over32 trajectories is exploratory, not
simultaneous/anytime coverage. Report all cells and paired contrasts; do not
convert isolated one-label gains or null results into whole-goal pass/fail.

## Verification and boundaries

Before collection: test decision-time poisoning of all hidden Y/Q/future X,
post-reveal forecasts independent of not-yet-available outcomes, complete-delivery
identity, real paid-label inclusion, same-unit paid counts, redundant-purchase
accounting, deterministic tie/random selection, detached replay and ownership
under race. Separate selection from outcome publication to make the boundary
auditable. Candidate scores must remain stable if only unavailable outcomes
change. After publication, the queried label may legitimately change forecasts.

Independently score all retained forecasts and accounting, verify source/raw/code
hashes, and replay summaries. Record collection wall/CPU and the number of fitted
models separately from the existing component benchmark. A shared batch can
provide all selectors in the experiment; report that amortization, not three
independent deployed query-pool costs. Perfect one-tick paid evidence is still an
experimental assumption. No production, private data, installation or publishing.
