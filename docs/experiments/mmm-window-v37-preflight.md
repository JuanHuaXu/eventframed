# Rolling/adaptive shape windows V37: isolated reasoning gate

2026-10-03. Previous goal turn was progress: V36 delivered bound original
forecasts and delayed identity correctness, but full-history reversal risk
remained near .25 rather than .16. All seven whole goals OPEN. Existing
protocols/models/tapes stay frozen; production and whitepaper untouched.

## Diagnosis and Model

Confirmed symptom, not an implementation bug: full-history stationary
inference pools contradictory old/new rates. Candidate causes include stale
influence, model shape mismatch, scarce post-shift evidence and delayed
selection. V35 stationary successes plus V36 opposite-rate cancellation
support testing explicit retention, not declaring a universal root cause.
Falsifier: failure to improve fresh shifts while protecting stationary cases.

Retain the newest W ARRIVED ISSUE ORDINALS per member, W in{4,8,16,64}.
Network arrival order does not determine evidence age. An old late outcome
outside a full window is acknowledged exactly once without evicting newer
trials. Cancellation is not a negative label. Event-time retention does not
repair nonignorable missingness; pending old labels remain unknown.

For each mean/shape hypothesis z let M_z(S) be its integrated member
likelihood. Remove an old outcome y_old from set S, yielding S_minus. Then:

    M_z(S_minus + y_new) / M_z(S)
      = P_z(y_new | S_minus) / P_z(y_old | S_minus).

Use the SAME remaining set for numerator/divisor; using the full-set mean
to remove evidence is incorrect. This yields O(378) member replacement,
without full-history refits. Under each retained set it is an exact
truncated-evidence WORKING posterior, not ordinary Bayes on full history or
a model of how a real latent rate evolves. Retention chosen from as-of
identities, not outcome values, preserves the declared fixed-width contract.

Adaptive observer maintains all four fixed-window children. Start weights
pi=(.85,.05,.05,.05), in order full64/4/8/16. On each resolved label y:

    v_j = w_j * P_j^issued(y) / sum_l w_l * P_l^issued(y)
    w_j(next) = (1-alpha)*v_j + alpha*pi_j, alpha=1/600.

Scores use privately retained ORIGINAL expert forecasts, not today's
forecasts. The emitted law is the predictable weighted mixture of current
child forecasts. These are predictive expert weights, not family posteriors,
an ADWIN detector, valid false-cut certificates or AP/source authority.
Under delays this arrival-order aggregation does not inherit no-delay regret
bounds. Children remain current-event-time windows as of arrived evidence.

The V36 identity ledger is reused without altering its sealed source. Its
baseline forecast is replaced privately before Issue returns; no baseline
update or baseline scored law is emitted. All child plans are validated
before any is committed. Owner/epoch/replay/cancellation/time/cap rules
stay intact. Epoch replacement invalidates old owners as well as old epochs.

N<=200,64 unique trials/member/epoch, single-owner only. Storage O(N*378*J
+N*64*J), J<=4, including the fixed ledger. Predict/Resolve O(J*378);
no I/O or database access. An unused ledger baseline dot product is measured
in Issue rather than omitted. Costs, atom/Beta window batch checks, atomicity,
no-future controls and fresh outcome experiments precede any adoption.

## Research Sources and Boundaries

[Bifet & Gavalda's ADWIN preprint](https://www.cs.upc.edu/~gavalda/papers/adwin06.pdf)
motivates the accuracy/stale-influence tradeoff. This implementation instead
aggregates a finite window bank; it is NOT ADWIN and inherits no detector
false-positive/negative guarantee.

[Mourtada & Maillard (ALT 2017), section5/Corollary11](https://proceedings.mlr.press/v76/mourtada17a/mourtada17a.pdf)
gives MarkovHedge/fixed-share expert tracking. We use a frozen restart-to-prior
transition and log-likelihood loss. Neither their growing-expert guarantees
nor a no-delay theorem is claimed for our delayed adapter. Generalization,
stationary protection and shifted usefulness must be tested independently.

## Precollection Verification

Independent batch Beta/atom laws and newest-arrived identity reconstruction
agree within3e-10 across2,11,150,200 members, widths1/4/8/16/64 and random/
reverse arrival through64 trials/member. Exact old-label eviction, late-label
non-eviction, invalid/replay and failed-normalization atomicity pass. Delayed
original expert forecasts, all-child commit atomicity, owner/epoch/cancel,
and prospective no-future prefixes pass. Fixed-share arithmetic is checked
both for identical issued rows and changing expert performance.

Apple M4 Go1.27.1: replacement18.176-18.753us, adaptive Issue1.141-1.149us,
first Resolve43.494-43.937us, all zero allocations. Benchmarks include state
restoration. Adaptive construction347.770-349.648us/3071454-3071460 bytes/
55 allocations, below the declared8MiB cap. These are not service/queue/
acquisition costs; the400ms complete-arm cap is declared separately.

Warm Resolve after16 rounds/member, when the three short windows replace
old labels, is75.585-76.072us with zero allocations. Do not substitute the
~44us first-update timing for this sustained replacement cost.
