# V79 canonical frozen-support ledger preflight

All seven WHOLE goals remain OPEN. This addresses the confirmed integration gap
in V78, not its .318 approximation defect, weak error envelopes or failed quality
gates. Parent checkpoint: V77/V78, SHA256
`5aa836d99096fe6c843edc2a8d855e261d3e0a0a526b2d1d7c4d4bc75cd05249`.

## Patch Reasoning Gate

Confirmed: V78 conditional helper holds historical supports fixed, but its
journal runs adaptive pruning after every reveal. Query branches then differ
from the accepted update. Other candidates are floating-point drift, incorrect
same-Y likelihood and unsupported evidence. Dense/path/tower controls reject
the first two as explanations for the measured re-pruning defect; zero support
remains possible and must fail closed. No upstream daemon fix applies: these
are isolated, unadopted research packages, not a production repair.

Fork inference mechanically, leaving V77/V78 frozen sources unchanged. A
single-owner ledger publishes one evidence/support/law bundle. Evidence updates
condition on its support without re-ranking. An explicit refresh recomputes
support and increments the support epoch; it is an approximation/publication
step, not ordinary Bayes. An issue appends exactly one transition and selects
only the new support mask before either observation is available, retaining
historical masks. Its scored receipt must use that new restricted law, not the
old unrestricted next-event forecast. Return BOTH clean-Y and first-sensor
probabilities, with the distinction named. Queries bind the whole publication
token; stale tokens reject. Reveal increments the law version, not support
epoch; arrivals do not insert latent transitions. Unsupported updates and all
other invalid operations publish nothing. Zero-probability query branches may
cause explicit abstention; never interpret arbitrary errors as probability 0.

Falsifiers: mismatch with an independent constrained dense trajectory oracle;
branch-sum or tower defect over 2e-11; query/failed-update mutation; historical
support changes during reveal/issue; hidden epoch refresh; future data seen
before reveal; or issued probabilities differing from independent masked-state
clean/sensor marginals. Require nonvacuous comparison with re-pruning controls.
Test older/latest first and same-Y second evidence, unknown gaps, equal/reverse
arrival clocks, isolated hypothetical branches, ownership, stale bindings,
configuration limits and explicit refresh. Do not weaken previous gates.

Cost remains cold full replay and allocations; this preflight does not promise
incremental operation, full-cap resources or loaded performance. Measure serial
32/64-row, 2/150/200-member public operations and constructor costs, with timing
scope explicit. Repeated growing-history cost must not be labeled hot serving.
Freeze all candidate/tests/compiler closure before unit/race/vet/benchmark.
No production, private/sealed labels, reserved seeds, prior dirty files,
whitepaper, publication, commit/push/install/deploy changes.

Next remains equivalent bounded incremental/rewind, full memory constraints,
omitted-mass safeguards, then the original broad quality/recovery/harm and
equal-TOTAL-cost experiments. Component coherence is not goal completion.
