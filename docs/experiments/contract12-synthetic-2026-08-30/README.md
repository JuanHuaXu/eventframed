# Contract-12 synthetic validation, 2026-08-30

This directory contains deterministic synthetic inputs and outputs generated
after EventFrame enrichment was moved behind the OpenClaw contract boundary.
The daemon stores and ranks the structured 5W1H EventFrame representation;
full turn text remains metadata and is not part of the semantic corpus.

## Artifacts

- `rerank-dataset.json`: seeded bidirectional reranking cases used by the
  corrected EventFrame-corpus run.
- `rerank-report.json`: promotion, demotion, retention, envelope, and latency
  results for those cases.
- `claims-design.json`: seeded design-block claim experiment output.
- `claims-confirmation.json`: independent seeded confirmation-block output.

The duplicate command stdout captures are omitted because they are byte-for-byte
copies of the retained reports. Local `.eventframed/` workspaces, databases,
tokens, account configuration, and real-session replay data are intentionally
excluded from version control.

These artifacts contain synthetic identifiers and generated event content only.
They are evidence for the tested configurations, not production guarantees.
