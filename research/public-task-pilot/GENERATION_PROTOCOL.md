# Stateless public answer pilot

Prepared before generation. Execution model/digest must be recorded before the
first request. Local service currently advertises only an embedding model;
generation runs remain zero until a generation endpoint/model is available.

Replay the consumed Nobel pilot's recorded semantic-baseline, lexical and sparse
top-ten selections, plus no-memory control: 14 questions x 4 arms = 56 requests.
Reconstruct each selected memory's public corpus text in recorded order. This
is not a fresh daemon retrieval run or an exact reproduction of a serialized
OpenClaw packet; it isolates generation on those selected records.

Each request consists of one fixed system message and one user JSON containing
only question and memory blocks. Local IDs m0..m9 identify citations; case IDs,
fixture IDs, gold labels, source-ranking correctness and support ranks never
enter model messages. No tools, conversation history, persistence or private
data. Fresh stateless request for every case/arm; temperature 0, fixed seed,
bounded 256 generated tokens. Request order uses fixed seed 2026112201. No retries
of model errors or malformed answers for score improvement; log failures.

Output is JSON with award, year, evidence. Absent/unsupported answers use null
award/year and empty evidence. Score exact canonical award/year, valid support
citations, unsupported assertions and abstention. Report missing/model/parse
failures, not just valid replies. No-memory abstention is expected but does not
count as solving a positive task. Category aliases are frozen in source.

The runner does not load the oracle. Grade only after recording all outputs.
Prepared prompts, actual submitted messages and raw model responses remain
auditable. Input/output files are exclusive-create and source-hashed. Existing
local endpoint only; no automatic model install or credential discovery.

This pilot is diagnostic with no population success claim. It cannot establish
research-reranking superiority: those overrides already failed against semantic
retrieval, and top-ten support is present for all positive tasks in that pilot.
It can reveal whether generation ignores available support, hallucinates absent
records, or is sensitive to the recorded ordering. Untouched domains, ambiguous
cases, repeated fitting groups, actual agent execution and persistent loaded
serving remain separate requirements.
