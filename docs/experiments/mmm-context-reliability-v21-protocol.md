# Context reliability v21 diagnostic

Freeze before replay of consumed v19 mixture_stop traces. This is not fresh
confirmation and changes no online observation or forecast policy.

Key each confidence stop by (acquired mask, acquired values, predicted class).
Use only earlier delivered outcomes for that exact key. Keep at most 512 keys
with least-recently-updated eviction and 64 outcomes per key. No hidden input,
regime label, oracle law, or undelivered outcome enters the warning decision.
At least eight outcomes are required; otherwise output unknown. Let n be support,
c correct outcomes and m their mean journaled nominal correctness max(p,1-p).
Warn if (c + 8*m)/(n+8) < m - .03; otherwise clear. The nominal-centered shrinkage
is a working estimate, not an ordinary posterior or a correctness certificate.
Read-only queries do not mutate recency. Deliveries follow decisions, even when
the label has zero delay. State restarts for each trajectory.

Use v20's same diagnostic gate on clustered majority shift128 confirmation Post:
warning coverage <=50%, error capture >=50%, warned error >=2*nonwarned error.
Nonwarned includes unknown. Retain every scenario including delayed/missing and
stationary groups. Record the unknown fraction to expose context sparsity.
Do not retune a failed threshold or call this a Brier/performance improvement.
Compare error targeting, not an untested counterfactual gain from extra reads.
