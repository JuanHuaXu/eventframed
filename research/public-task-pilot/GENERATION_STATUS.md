# Generation pilot status

2026-10-01 update: a fresh `openclaw agent exec --isolated --auth-env-only`
state directory reached the installed `openai/gpt-6-luna` route but failed
before generation with `No route-compatible authentication source is
configured for openai.` The first attempt used unsupported thinking level
`off`; a second attempt with supported `low` produced the auth failure.
The local Ollama inventory still lists only `nomic-embed-text:latest` with
embedding capability. The running lab OpenClaw configuration enables its
memory plugin, so using that configured agent as a stateless generation
control would introduce a different context boundary. No provider credential
was inspected, copied or installed; no answer-generation call succeeded.
Fresh, outcome-labeled public retrieval evidence is now in
[the broad USGS result](USGS_BROAD_CONFIRMATION_RESULTS.md), but generated
answer quality remains unmeasured. The user has been asked asynchronously
whether to install a small local completion model for isolated tests.

2026-09-12. **Prepared; actual generation calls: 0.**

The local `127.0.0.1:11434/api/tags` response lists only
`nomic-embed-text:latest`, with embedding capability and no completion model.
No generation model was downloaded and no credentials were located/copied.
The user has been asked to choose a local model download or a lab endpoint.
Other roadmap work is not blocked by that pending choice.

`generation-prepared-v1.json` contains 56 stateless requests built from consumed
public Nobel/NASA retrieval selections. `GENERATION_PROTOCOL.md` freezes the
arms, ordering, output format and limitations. Model identity/digest still needs
to be recorded before execution. Each model request receives only the fixed
instruction, public question and selected text blocks with local citation IDs.

Three tests pass: deterministic preparation and oracle/identifier separation;
grading positive and negative controls; and a mocked transport verifying fresh
two-message requests without prior conversation or tools. Mocked transport is
not a language-model experiment. No accuracy, hallucination-rate, latency or
generalization result is claimed from these tests.

The no-memory control is instructed to abstain without supplied evidence.
Positive-task abstention is scored as an unsolved task; it is not misreported as
answer accuracy. Correct-but-unsupported assertions do not pass grounding.
Malformed responses and provider failures remain failures, not silently retried.

After generation, the runner output must be retained before grading loads the
oracle. This is replay-context answer testing, not an OpenClaw end-to-end run or
an untouched domain trial. All broader completion requirements remain open.
