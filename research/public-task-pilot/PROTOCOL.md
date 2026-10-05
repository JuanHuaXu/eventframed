# Public outcome-verifiable task pilot

Prepared2026-09-12; model runs NOT yet conducted. This first dataset is a wiring
pilot, not sufficient population evidence for recommendation5. It contains13
verified public facts across3 mission clusters; confirmation is only1 cluster.
Do not multiply paraphrases into independent samples or call this enough for
clustered superiority inference. Broader sources/tasks remain necessary.

Sources checked directly2026-09-12:
[NASA Mariner10](https://science.nasa.gov/mission/mariner-10/) and
[NASA Voyager planetary voyage](https://science.nasa.gov/mission/voyager/planetary-voyage/).
Only historical encounter/launch/contact dates are used. Mutable mission-status
prose and conflicting range/altitude conventions are deliberately excluded.
No private transcript, invented scientific fact or unpublished research is used.
Dataset IDs are local labels, not people or external accounts.

## Prospective execution boundary

Export corpus text separately from evaluation oracle. Never ingest expected
answers, target IDs or scoring keys as instruction context. Runtime observation
IDs must be recorded in a private manifest mapping to public fixture labels;
do not put that mapping in the query. Legitimate memory context may expose its
own event IDs via the actual plugin formatter.

Use isolated seed/query states, fresh workspace and session for EVERY query,
disable capture and all alternate memory hooks, and no browser/tool fallback.
Model returns JSON with answer (ISO date, UNKNOWN or NEEDS_CLARIFICATION) and
evidence_ids. Positive credit requires correct date AND a valid retrieved event
ID supporting the relation. Score task outcome and evidence separately. Record
provider-visible context and packet target survival; answer correctness alone
cannot distinguish prior knowledge from memory use.

Arms planned: no memory; existing daemon/control; research adapter only after a
real-field feature/feedback mapping is frozen. The synthetic9-bit learner is
NOT to be applied to retrieval diagnostics by relabeling coordinates. Until
that adapter exists, no result here validates retained-learner transfer.

Design uses Mariner10/Voyager1 tasks. Confirmation uses Voyager2 and must receive
no feedback until all answers are sealed. Include one design ambiguity and
one absent-record task in each split. At most48 generation calls for this pilot
(16 tasks*3 arms); no automatic retries to obtain better answers. Failed/absent
responses count as failures unless a recorded infrastructure fault invalidates
the arm. Do not launch against production or reuse production sessions.

Lab launcher was found read-only at the configured lab home LaunchAgents
location; node and entrypoint exist. The old harness contains state/auth copying
and migrations and must NOT be executed unchanged. New isolated paths, clean
query-state audit, adapter contract and provider configuration must be checked
before model dispatch. No auth values were read or copied during preparation.

This dataset has not been used to choose research parameters. Existing public
testdata were searched for Mariner/Voyager overlap with no match, but broader
pretraining overlap is expected and explicitly controlled via evidence IDs and
the no-memory arm. Freshly authored prompts do not imply facts unknown to an LLM.
