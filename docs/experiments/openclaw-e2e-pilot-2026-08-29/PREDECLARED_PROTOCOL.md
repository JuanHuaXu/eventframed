# Contained OpenClaw E2E Pilot

Date frozen: 2026-08-29

## Boundary

This pilot runs a disposable loopback-only OpenClaw gateway using the production
`openai/gpt-5.6-luna` model connection from the existing lab profile. It does
not use the production OpenClaw state, workspace, sessions, tenant, delivery
channels, or memory collections. Agency is disabled and its kill switch remains
active.

The exercised EventFrame path is:

`OpenClaw agent -> before_prompt_build -> EventFrame adapter -> Unix socket ->
eventframed -> LibraVDB SearchTextCollections/RankCandidates -> packing ->
prompt injection -> production model -> agent_end -> event observation`.

## Arms

1. `libravdb-control`: the unmodified LibraVDB memory plugin selected as both
   memory and context-engine slots, with a unique test tenant.
2. `eventframe-pass`: EventFrame selected as the memory plugin, contextual and
   hierarchical scoring disabled, and residuals disabled.
3. `eventframe-active`: EventFrame selected with contextual scoring,
   hierarchical shrinkage, and applicable certified residuals enabled.

Each arm receives a different tenant identifier but the same ordered seed and
query turns. The gateway and daemon are restarted between arms. `recall_k=50`,
`pack_k=10`, and the memory token budget is 2,000 for both EventFrame arms.

## Fixtures

The model is instructed to acknowledge each seed record exactly. Queries are
issued from a separate session key to test cross-session recall. Expected
answers are exact strings. The fixture set includes direct lookup, paraphrase,
a superseded value, untrusted recalled instructions, and an absent fact.

## Recorded Evidence

- OpenClaw JSON result for every turn and wall-clock duration.
- OpenClaw cache trace containing the provider-visible prompt boundary.
- EventFrame recall packet, candidate scores/deltas, snapshot, hook duration,
  observed event identifier, and recalled lineage.
- Daemon Prometheus metrics and process resource samples.
- Gateway and daemon logs.

Raw traces stay in the ignored `.eventframed/e2e` directory because they contain
complete prompts and model responses. The checked-in report contains aggregate
results and redacted failure excerpts only.

## Pilot Gates

- All paths and tenant identifiers must point into the disposable test boundary.
- No channel delivery and no agency execution may occur.
- Every successful EventFrame model turn must have a matching recall trace and
  every successful captured turn must have a matching observe trace.
- Daemon-unavailable recall must fail open without preventing a model response.
- The EventFrame packet must report the real LibraVDB nomination and ranking
  contracts, not embedded substitutes.
- Exact answer accuracy and harmful stale/instruction-following failures are
  reported per arm. This small pilot is diagnostic and does not establish
  population superiority.
- EventFrame recall p99 must remain below 100 ms for local adapter/daemon work,
  excluding remote model generation. Remote contract latency is reported
  separately as part of observed recall duration.

## Non-Claims

This pilot does not have enough independent trajectories for confidence
intervals over production workloads. It does not test a target outside the
50-candidate frontier, autonomous agency, robotics, corpus-scale storage, or
long-run convergence. A later untouched confirmation run must freeze its own
fixtures and thresholds after this pilot is complete.
