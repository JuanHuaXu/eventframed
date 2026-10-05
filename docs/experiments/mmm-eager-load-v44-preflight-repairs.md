# V44 preflight repairs

2026-10-03. Preserve each exclusive artifact root and its exact source snapshot.
No test failure changes the workload, native durability, seeds or quality gates.

`research/eager-load-v44-preflight-initial` generates the complete workload,
then compilation rejects six authority selectors in the new control fixture.
The load wrapper embeds `loadWarmV37`, not `warmStoreV37`; authority is reached
through `s.eager.authority`. Repair only those six fixture selectors. The eager
store's own `s.authority` selectors are already valid and remain unchanged.
This is a fixture compilation failure, not a measured runtime failure/pass.
