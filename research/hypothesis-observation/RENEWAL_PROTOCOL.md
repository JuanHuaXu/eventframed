# Independently renewed measurement pilot

Frozen before outcomes. This changes the observation family, not the truth
labels or report noise. It cannot retroactively rescue prior failed screens.

All policies start from the existing exact Joint model with uniform16-hypothesis
prior and independent copy-mode prior .5 per type. Eight original likelihood
types and four ordinary slots per type remain. An ordinary slot costs1 credit;
a renewed measurement costs2. Renewals use a separate iid conditional draw and
update only hypothesis weights, not the ordinary source's first report/mode.
That conditional independence is a simulator assumption, not an implemented
provenance certificate. Include false-renewal20, where the fresh-labelled path
actually repeats each type's ordinary first draw and the model is misspecified.

Cases: independent20, copied20, mixed20 (even types copied), copied05,
false_renewal20 (all ordinary and renewal reports copied). Every episode gives
all policies the same latent target, ordinary tape and separate renewal tape.
Policy randomness is separate. Noise/likelihood is known, copy mode is not.

Four policies: regular-only two-credit lookahead, mixed ordinary/renewal
two-credit lookahead, random over feasible mixed actions, and maximum report
entropy per credit over mixed actions. Receding planner minimizes expected
terminal target-class Brier over min(2,remaining credits), executes one action
then replans. It charges every action, including hypothetical branches, according
to declared credit cost. Planning CPU is separate from acquisition credits.
No actual future report or target is passed to selection. Fixed lowest-index
ties at1e-14 tolerance. One next ordinary slot per type avoids duplicating
equivalent actions; up to8 renewal slots/type suffice for the16-credit budget.

Two splits,64 episodes/case/split, base2027110701*1000000 plus split*100000,
case*1000 and index. All effective seeds must be recorded. No prior matching
seed was found in this experiment directory. This is a frozen synthetic pilot,
not real-task validation. Keep all640 episodes and all four policy traces.

Every arm spends exactly16 credits. Report final multiclass Brier, Brier area
over credit time (pre-action forecast held during the action's cost), accuracy,
confident-wrong fraction, renewal count and calls. Primary gains: mixed planner
vs regular-only in copied20 and mixed20, mean>=.02 and paired lower>0. Primary
nonharm: mixed vs regular-only in independent20 and copied05, mean harm<=.01.
Also require mixed-planner Brier gains vs random and entropy controls on copied20
and mixed20 with positive paired lower bound. Both splits required; intervals
mean +/-3.3 SE over64 episodes are descriptive, not simultaneous certificates.
False-renewal is a separately reported negative control: any harm/confidence
inflation prevents claims of robustness to unverified freshness.

Verify credit and slot uniqueness, separate ordinary/fresh state effects,
posterior normalization, original controls, causality of selection, complete
deterministic replay, and no omitted failed episodes. No production, external
measurement service, privacy-sensitive data, or whitepaper changes.
