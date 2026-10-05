# Logged acquisition gain: finite audit protocol

This tests an evidence-measurement path, not a new acquisition policy. Existing
random, entropy and joint8 strategies are unchanged. All consumed2688 source
trajectories remain accounted for. Only delayed schedules have paid candidate
actions; complete schedules are identity controls, not extra independent data.

Project each strategy's actual query branch to realized mean Brier over future
frames161..191: use recorded branch Predictions and realized source Y, never
teacher Q or the branch Sample/Population risk. Previous actual-answer sampled
risks integrated future label noise; they are not logged observed losses.
Keep that distinction explicit rather than retroactively relabeling old results.

For each of64 predeclared logging replicates, sample one of the three strategies
uniformly using SHA256-based seeded rejection sampling. Collapse strategies
that nominate the exact same query origin: their logged action probabilities
are multiplicity/3 and their loss is identical. This is exact action identity,
not inferred abstraction sharing. Log only the chosen action's realized loss.
The full action table remains a separate simulator audit oracle.

Train three constant loss predictions on only phase0 delayed logged outcomes
in that replicate, using strategy membership of the selected exact action.
Use(.5+lossSum)/(1+count); combine colliding strategies' constants by their mean.
Freeze these before phase1. This is a deliberately simple regression control,
not a claimed optimal model. Never use phase1 outcomes to fit the predictor.

Compare joint8 minus entropy utility, i.e. entropy loss minus joint8 loss,
with both inverse propensity (zero regression) and doubly robust estimates.
The logged probability is known by randomization, not inferred from model
confidence. Also retain random versus entropy. For each comparison and method,
use the bounded Hoeffding-mixture CS implemented in research/logged-gain.mjs,
alpha=.05/4, fixed rates[.01,.02,.05,.1,.2,.5,1,2]. Report time-varying running
average conditional gain, not a constant-mean or future-deployment value.
Target differences are simulator-auditable from the full table. Report exact
anytime coverage across64 random logging assignments as a diagnostic, not64
new datasets, nor a proof of the theorem. Count positive lower bounds without
calling one of the previously failed acquisition policies successful.

Report per-case and aggregate realized full-information gains, logged estimator
errors/interval widths, distinct-action counts and query costs. Averages do not
replace prior all-case policy advancement gates. Record how many delayed
phase1 episodes make identical candidate/control choices. Future streams after
the one query remain fixed; this is not off-policy evaluation of a full
continuously adapting memory policy with endogenous future state.

Require exact unbiasedness/MGF and adaptive-tree contracts, source hashes,
realized-loss projection checks, hidden-teacher isolation, deterministic replay
and independent estimator reconstruction. No production, paper or deployment.
