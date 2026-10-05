# Return evidence routing to adaptive observation

Executed: [v103 adaptive results](../docs/experiments/mmm-adaptive-v103-results.md).
The lookahead component passes literal-reference tests, but the full experiment
FAILS its two incremental-lookahead gain gates (106/108 pass). Retain coherent
one-step as the lower-cost candidate; the plan below records the original scope.

Component progress: [coherent snapshot and observer checks](joint-observation-component-results.md)
pass. These were the prerequisite checks before the completed v103 run.

Research integration plan. A passing fixed-view replication cannot establish adaptive MMM
performance. This is the next integration boundary for directions1,2,4 and7,
not production promotion or a replacement for delayed-feedback validation.

## One immutable law during acquisition

The current [conditional observer](../internal/observationlearners/conditional_observer.go)
uses the same joint table for its forecasts, prospective observation values and
confidence stopping. Preserve that property when introducing routed weights.
Do not choose observations with one model and silently score another law.

At the start of a frame, after any scheduled fit, freeze the raw bank weights,
rejection states, routing credits and four immutable fitted models. Compute the
effective five-component weights before reading the current hidden fields.
For the first prototype require the same declared input law in all components:

    Q_t(x,y) = mu(x) * [v_0 Bernoulli(y;0.5)
                       + sum_{j=1..4} v_j Bernoulli(y;p_j(x))].

Use conditional marginals of this law throughout that frame. With a common
input law, its partial forecast equals the weighted component partial forecasts.
With different input laws, that identity need not hold: do not accept that case
without the correct input-conditioned mixture weights and a separate contract.

Hypothetical forecast queries must be read-only. They cannot consume issuance
IDs, create feedback records, update a posterior or add evidence. Commit only
the final actually observed mask and its raw/component forecast bundle; later
feedback updates it once. A failed reader call must not partly commit an issue.
No current outcome or unrequested field may enter the observer snapshot.

An O(K) virtual joint-cell view over the existing immutable component tables can
avoid constructing a new3^9 table for every changing weight vector. Verify it
against literal joint enumeration before claiming an optimization. K=4 and the
nine-bit space remain research bounds, not arbitrary-domain scalability.

## Controls that separate the changes

Freeze a fresh paired protocol before new outcomes. Include at least:

- Generic64 with its existing one-step conditional observer.
- Routed forecasts on those exact acquired masks: isolates the forecast change.
- Routed forecasts with their own coherent one-step observer: tests acquisition.
- Routed forecasts with bounded joint lookahead: a separate planner candidate.
- Routed forecasts with random affordable views and with fixed mask63.

The fixed-mask arm must reproduce the relevant v102 logic on a shared consumed
input tape before any new confirmation. Keep uniform input assumptions explicit,
the six-coordinate foreground cap, forced initial view and all attempted-read
costs. Count any full-frame training audits separately and equally for all arms;
do not hide complete training inputs behind the foreground cap. Both observed
quality and total acquisition cost belong in the new gates. Retain partial-
reader errors rather than inventing absent values. Later missing-field support
requires its own observation/marginalization contract.

## Why joint lookahead is a distinct hypothesis

For uniform independent parity, any remaining unobserved required bit leaves
the conditional outcome probability0.5. If every single available view leaves
at least one required bit hidden, every one-step information gain is zero even
when a sequence of affordable views reveals the rule. This is a structural
counterexample, not a claim that every existing trajectory follows that path.

In the fixed small domain, a candidate can enumerate affordable future view
sequences using the fitted joint law, memoizing by observed mask/values and
remaining budget. Its terminal value is the same forecast-entropy objective as
the one-step observer; each branch is weighted by its prospective observation
probability. Do not use true simulator rules or outcomes to rank branches.
Keep the existing confidence-stop rule fixed initially to isolate planning.
Count planning nodes, latency and memory, not only paid coordinates. Bounded
lookahead is our proposed research adaptation, not a claim of automatic causal
discovery or of a theorem imported from another architecture.

## Required tests before claims

Validate all partial-state marginal identities, one-component and neutral-only
reductions, future-hidden-field independence, reader epoch changes, failed-read
atomicity, exact cost accounting and final-forecast/feedback binding. Compare
lookahead with an independent literal tiny-tree enumerator, including zero
one-step gain but positive joint gain. Freeze success criteria before the fresh
adaptive run; never weaken the existing fixed-view result's106 gates afterward.

Only after that test should the new learner return to the older member-split and
delayed-feedback harness. The [scope audit](scope-audit-v102.md) lists the other
requirements; none is satisfied merely by adding this observer adapter.
