# Contextual mixture: rejected screen

The frozen nine-bit contextual ridge extension failed to improve the matched
global combiner. All data are consumed synthetic independent-v1 trajectories;
no new fitting samples or confirmation claims.

| Method | Whole expected Brier | Terminal64 | >.01 harmful32-windows |
|---|---:|---:|---:|
| Context + pointwise guard | .156871533 | .145313551 | 0 |
| Matched global + guard | .156838173 | .145279261 | 0 |
| Fixed Share + pointwise | .156847006 | .145296192 | 0 |
| Markov | .157516545 | .145520129 | reference |
| Context unguarded | .154315335 | .145580608 | 322 |
| Global unguarded | .154180571 | .145472437 | 293 |

Context-minus-global whole difference+.000033360 has pointwise eight-index
bootstrap95% interval[.000027932,.000038986]. It also worsens changing-scenario
whole performance by .000050007, interval[.000038699,.000062030]. This is a
failed quality rescue, not support for faster recovery. The result rejects
this particular feature/regularization/retention contract, not all contextual
learning. Do not tune additional feature grids on this cohort.

## Verification and cost

The normal-equation residual was at most2.67e-15. An independent Gauss-Seidel
solution agreed within3.89e-16. As-of/current/future/missing evidence tests and
retained-set completeness passed. A deterministic context-dependent positive
control was learned: unguarded late Brier .00005917 versus .25 for the global
control. This demonstrates wiring, not performance on the research population.

The full screen replayed byte-identically. Matched scalar and other prior
controls agree within1e-14 over all672record scores. No behavioral bug found.

Warm context+guard reference replay costs13.84--13.98us/forecast amortized.
The matched zero-context ten-dimensional solver costs13.65--13.91us; the
earlier specialized scalar implementation costs~.48us. These are not serving
latencies and exclude base predictor fitting/I/O. Complexity is
O(T^2+TWp^2+Tp^3), W=64,p=10. Optimizing this rejected quality candidate is
not the next priority. No runtime changes or new Go race run were needed.

## Re-centering the next experiment

Reviewed the original seven criteria. Goals2/4 require a better shifted
predictor, not endless mixer changes. Audited all15 already-collected v120
heads, retaining every result in `mmm-independent-head-audit-v1.json`.

Markov remains the best pooled head (.157516545). Segment64 loses pooled
(.158618743) and stationary (.123564054 versus .119723242) but wins on changing
scenarios (.215582613 versus .218930662; terminal .200661476 versus .204761880).
This is post-hoc ranking, not independently validated model selection.

Next read the existing segment/static model contract and implementation, then
freeze a matched segment64-versus-static64 challenger experiment. Preserve
Markov and the unchanged guards. Test whether a guarded segment model retains
shift benefits without its stationary harm. No scenario labels or oracle
routing may enter the predictor; all model costs remain charged. Data used
for this audit are consumed and cannot serve as fresh confirmation.

Artifacts: `mmm-context-mixture-v1-screen.json`, `-analysis.json`,
`-screen-replay.json`, and `mmm-independent-head-audit-v1.json`.
All seven goals remain open. Production, daemon, whitepaper and Git untouched
apart from isolated untracked research artifacts; no commits or pushes.
