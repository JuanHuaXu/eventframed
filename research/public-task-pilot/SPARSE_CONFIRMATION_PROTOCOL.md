# Sparse pilot confirmation

Freeze after design sparse-v1 screen, before executing Voyager2 queries.
Use the unchanged sparse_v1 feature map and fitting parameters. Fit on all eight
answer-bearing design queries, then score all six confirmation retrieval packets
without updates. Same post-contract in-memory/hash-embedding retrieval as design.
No new parameter selection. This is one mission cluster, not population evidence.

Run baseline, lexical, learned and equal-weight blend. Keep the failed design
blend for transparency, not as the selected candidate. Evaluate five positive
questions with the same ranking/Brier metrics; evaluate the absent question only
as a separate all-negative support diagnostic. No LLM answers or abstention
accuracy are measured by this run. Report candidate support loss if any.

Pass screen: learned top1 at least baseline and lexical, learned candidate Brier
below 0.09 on the five positive tasks, and absent-task mean probability <=0.1.
The 0.1 absent threshold is a new prospective boundary, not a previous design
claim. No confidence interval over ten candidates as independent trials.

All queries are retrieved before the oracle is opened by the evaluation layer.
No outcome is provided to the service or learner during retrieval. Full raw
packets, source hashes and per-candidate predictions are retained. If this pilot
passes, external-source, phrasing, delayed-feedback and real-agent testing remain.
