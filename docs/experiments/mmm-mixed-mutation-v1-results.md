# Mixed-mutation feedback boundary v1

2026-10-01. **PASS the finite fail-closed/rebind correctness screen**, not a
loaded-learning or durable-recovery gate. The
[frozen contract](mmm-mixed-mutation-v1-contract.md) preceded the combined
test in `internal/service/research_feedback_mixed_mutation_test.go`.

The same sequence passed on an in-memory store and a real temporary persistent
LibraVDB store. It used the research publication wrapper, actual service
Observe/Recall/Delete paths, a committed frontier journal and the temporal
feedback bridge. One visible event was recalled and admitted. A future-only
insert preserved its already-bound as-of feedback, which completed exactly
once; scoring where the future event would be visible was rejected. A second
frontier remained pending when the visible event was deleted. The old bridge
then rejected that pending label and score without increasing its worker's
completed count. After a new visible event and a new bridge, a fresh journal
admitted and completed one explicit label. The old bridge still rejected its
old pending label. Neither absent labels nor failed feedback became negative
training examples.

Verification:

- `go test ./internal/service -run '^TestResearchFeedbackMixedMutationRebind$' -count=1 -v`: both backends pass.
- The same focused test under `-race -count=3`: pass.
- `go vet ./internal/service`: pass.
- Focused `internal/researchpublicationstore` race tests for mutation-method
  classification, real transitions and orderly restart: pass.

The result proves only these declared event and mutation orderings. A fresh
`NewResearchTemporalFeedbackBridge` constructs a new background learner;
the earlier learned state is **not** carried into it. The test checks safe
rejection and fresh learning, not continuous-model recovery across a general
mutation or crash. No concurrent mixed-write load, tail-latency distribution,
feedback replay, arbitrary interleavings, production OpenClaw path or agent
task outcome was measured. Goal 6 remains open. The next integration problem
is to bind a validated durable learner state to a fresh publication epoch,
with unresolved labels explicitly discarded or replayed under a proven
identity/epoch contract, before a mixed-mutation load screen can claim
continuity.

No serving default, production process, whitepaper, remote or dependency was
changed.
