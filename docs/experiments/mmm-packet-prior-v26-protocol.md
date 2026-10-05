# Packet prior v26: actual scored-law integration

Date: 2026-10-02. Freeze before either cohort is run. V25's contextual
predictive passed its finite forecast screen but was offline; bounded ranking
failed wide reversed recovery. Test full contextual prediction and ordering
through the actual Service/LibraVDB/SQLite boundary without modifying
production service code. This is a research adapter, not daemon deployment.

## Scope and Model

Keep V25's two geometries (.005/.020), four hidden regimes, 150 activated
nominees, first-32 one-outcome monitoring, 200 eligible/17 future unit vectors,
recall50/pack10 and same-as-of epoch unchanged. Use new seed bases2026102603
design and2026102604 confirmation, with the existing geometry/regime/world
offsets and separate hidden-truth RNG. Run32 worlds per geometry/regime per
split. Prior strength stays2 and no model parameter is tuned on these tapes.

An isolated single-owner Store adapter binds immutable prior means from the
first durable pre-feedback journal. It rejects unscoped or mismatched query,
tenant, horizon, origin, policy, epoch and as-of reads; sharing, fractional
counts, working posteriors and non-unit full-stream outcomes are unsupported.
The caller derives context scope from the actual request's framed query,
embedding and model key. Missing or stale authority returns no posterior.
This caller-scope contract is not a native Store query-context API.

The underlying store retains ordinary Beta(1,1) sufficient counts. The adapter
returns Alpha=2*b_ref+u and Beta=2*(1-b_ref)+v. The same Beta-Bernoulli joint
family from V25 then gives the evidence likelihood and forecast. Anchors may
be reconstructed from the original durable journal, not silently re-centred
after outcomes. Exclusive owned writes are assumed; raw external mutations
or a poisoned scope provider are not certified by this test adapter.

Use the existing contextual scoring policy with PosteriorWeight=1, all other
weights0, identity predictive calibration, disabled residuals, MaxRankDelta=1,
and disabled elastic scaling with min=max=1. The resulting accepted contextual
posterior must appear in BeliefLaw, PreResidualLaw, CorrectedLaw, RankScore,
durable journal and the returned packet. Unmonitored nominees keep baseline.
No rank-only callback is used. Existing synthetic selection/omission
certificates are assumptions, not empirical coverage evidence.

## Comparisons and Frozen Screens

The actual served contextual law/order is the candidate. Reconstruct baseline
ordering and V25's two bounded rank controls from the same eligible events and
labels; those are offline counterfactuals, not separate served arms. Retain
flat/context full forecast counterfactuals as in V25. Verify actual packet,
all durable nominee forecasts and expected risks independently. Report the
current unchanged service's evidence from V25 separately; do not call an
analytic counterfactual an actual control run.

Use identical V25 screening intervals (paired mean +/-3.5SE over32 worlds,
descriptive, not confidence sequences). Separately in each split/geometry:

- Actual contextual packet usefulness: independent and reversed gain>=.02
  with positive lower endpoint; aligned/calibrated harm upper<=.01.
- Actual whole-frontier expected Brier: independent/reversed gain>=.005 with
  positive lower endpoint; aligned/calibrated harm upper<=.01.

Preserve failures without changing margins, monitoring or sample size. Before
cohort generation, test no-evidence identity, durable-anchor reconstruction,
wrong-query/no-scope/future-label/epoch/policy rejection and immutable origin.
Run race controls and six diagnostic full-service worlds (both geometries and
independent/aligned/calibrated). Record sequential call cost; no loaded p99,
large-corpus or agent-task quality claim follows. All seven goals remain open
unless their full requirements are independently achieved.
