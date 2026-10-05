# Pilot preparation status

2026-10-01 update: The original NASA `prepared-v2` questions were consumed
by later sparse and pre-packing studies despite this old manifest's static
`model_runs=0` value. Do not treat its confirmation split as untouched.
The separate [NASA transfer v1](NASA_TRANSFER_V1_RESULTS.md) used new mission
facts and frozen prior weights. Its sparse promotion gate failed on one lost
literal top-1; it was retrieval-only, with three mission clusters and no
agent outcome run. Goal 5 remains open. The real service research-rank hook
now acts before recall truncation and packing; the older shadow-hook warning
below describes its 2026-09-12 state, not the current hook.

2026-09-12. Current export: `prepared-v2/`. The initial `prepared/` is a
superseded pre-run draft: its generic "at Earth" wording was inappropriate
for radio contact/launch. Corrected to "with Earth"/"from Earth" before any
model run. Both manifests truthfully show model_runs=0.

Public source verification, separation of corpus/query/oracle, date validation,
split-cluster disjointness and relation-wording tests passed. There are13
facts and16 tasks, but only one confirmation mission cluster. This is NOT
sufficient evidence for population improvement and no such claim is made.

Lab entrypoint and configuration paths were checked via the actual lab plist,
read-only. No production binary was invoked/changed. No credentials or private
sessions were copied. Existing harness cannot be reused unmodified: clean
states and a validated real-field research adapter are still required.

The [real-field adapter contract](ADAPTER_CONTRACT.md) is now implemented in
`internal/researchmemory`: bounded 5W1H/query features, retained learner,
pre-feedback journals, explicit labels, epoch/time checks and pending limits.
Targeted race tests and vet passed for excluded-data invariance, chronology,
duplicate rejection, fitting and concurrent access. This is an unvalidated
transfer hypothesis, not evidence of improved task output. Model runs remain0.

Next deliverables: actual post-contract capture/retrieval adapter preflight,
no-memory/control/research generation runs with packet and provider-visible
evidence, then a substantially broader untouched task sample. Reranking must
see the nominated frontier BEFORE packing; it must not operate only on an
already truncated ten-item packet. The current diagnostic shadow hook does not
satisfy that placement requirement and must not be represented as doing so.
