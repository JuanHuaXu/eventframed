# Repaired-control mask-reuse results

2026-10-05. Component preflight **PASS**, not production adoption or goal 6
completion. The control is the already repaired `f3231fa` extractor, not the
unsafe pre-quotation-admission code.

| Workflow | Geometric-mean median latency reduction | Worst cell latency change | Candidate component range | Allocations saved per call |
| --- | ---: | ---: | --- | --- |
| Turn | 11.495% | -3.621% | 37.269 us to 9.883 ms | 16 to 47 |
| Text import | 10.283% | -3.298% | 18.667 us to 4.075 ms | 12 to 15 |
| Query | 12.020% | -3.743% | 23.137 us to 4.511 ms | 16 to 20 |

All 45 cells improved in these descriptive matched measurements. Each cell has
six samples per arm (three forward and three reverse). The unchanged preflight
required at least 5% geometric-mean improvement in each workflow and no cell
regression above 10%. Raw per-cell timing and allocations remain in
`benchmarks.txt` and `results.json`; these are not confidence intervals or loaded
serving percentiles. Balanced cell weights are not real-traffic weights.

The complete repaired-control frame suite passed. The candidate passed all 36
selected semantic tests, including 8,192 full-event/query differential comparisons
and 384 parallel comparisons; race and vet passed. The old exact-source fidelity
test runs on the control only, as preregistered: its pins identify that immutable
old source. Candidate fidelity instead uses the source-hashed new patch and a
separately compiled complete control. No semantic test or scientific threshold
was weakened. Parent patch review found no confirmed P1/P2 in this scoped change;
it is not represented as a new independent review.

Five service integration test families also passed in a separate isolated
checkout: post-contract enrichment/authored-field preservation, incomplete
roster lookup/cancellation, durable reopen plus retry indexing, immutable
unresolved-reference retries, and context/card availability boundaries. These
use public synthetic inputs and ephemeral stores; future context/cards and
cross-session/cross-tenant context remain excluded. They do not measure serving
latency. Reproduce that boundary check separately with
`node research/frame-mask-cache-v1/integration.mjs`.

Masking remains byte-aligned, spans crossing quoted content still abstain, and
raw/summary text, confidence, identity and provenance outputs are unchanged in
the comparisons. Masks are call-local and use only the provided envelope. No
private chats, corpora, next-packet labels or production services were used.

## Seven-goal accounting

This removes a measured component cost introduced by the safer extractor. It
does not repair the joint TV defect, establish valid useful Anti-Pigeon splitting,
give new untouched agent-task outcomes, or establish a better observation policy.
Freshness and serving latency still require a loaded mixed-workload experiment;
the earlier failing loaded p99 is not relabeled. All seven whole goals remain
open. The candidate is retained as an isolated patch for that next integration,
not silently applied to the live checkout or daemon defaults.

Reproduce with `node research/frame-mask-cache-v1/run.mjs`. The manifest freezes
control revision, candidate patch, protocol, tests and source bytes before runs.
Historical failures, controls and original research success criteria are intact.
