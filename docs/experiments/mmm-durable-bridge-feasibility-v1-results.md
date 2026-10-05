# Fixed-time durable feedback screen v1

**FAILED before the second guarded admission.** The frozen
[v1 contract](mmm-durable-bridge-feasibility-v1-contract.md) requested 32
successive feedback updates while all later forecasts remained at the same
earlier query time. After the first label became available at `now+1s`, the
durable learner correctly refused the next prediction at `now` with
`invalid or stale frozen prediction`. The failing test source SHA-256 was
`73059597236a3f84388515ccf7c47f9fd2b360be56ca1a8090218d241919d82b`.

This is a protocol contradiction, not evidence of a learner bug. The frozen
model's as-of time advances after feedback; a later forecast must not be
backdated before that feedback. No completion, replay, mutation or performance
claim follows from v1. The failed run is retained rather than changing its
time contract in place. V2 will predeclare monotonically advancing query and
feedback times, then rerun the same cross-layer assertions.
