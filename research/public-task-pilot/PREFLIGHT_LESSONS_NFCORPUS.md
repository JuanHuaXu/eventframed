# Local Reproduction Lessons

2026-10-04. Project-local, not a global instruction change.

- Inspect function signatures before adapting an input: frame.Decode accepts
  bounded bytes, not an io.Reader. Compile/list preflight caught the mismatch
  before any prediction; preserve v1 source copies and compiler failure.
- Mechanical version clones must distinguish output/self paths from immutable
  dependency paths. Protocolv1 remains the scientific protocol in later trials.
  Validate each rewritten path exists before dispatch. Preserve v2 failure.
- Read fixture requirements before full regression runs. Existing source/pool
  tests explicitly assert originalSciFact5183/10869counts; supply/freeze that
  fixture rather than silently relabeling them asNF3633tests. NFfullconversion
  receives its own independent complete source audit. Preserve v3 race failure.
- A negative corruption must actually change the object. Empty-array pop and
  reverse are no-ops; select a nonemptycase for row/order corruption AND test
  the legitimate empty case separately. Never remove empty queries from metrics.
  Preserve v4 original auditor failure and its frozen hash; newauditor supplement,
  unchanged model/predictions/protocol. No relevance consumed before this repair.
