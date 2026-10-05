# V54 preflight repairs

No main diagnostic outcome was collected before these repairs.

- The initial paired-model and fixture preflight tests passed. During source
  audit, the proposed additive RNG salts54001/54003/etc were found to alias
  across nearby diagnostic/design/confirmation bases, which differ by2.
  Example: diagnostic world seed+54003 equals design world seed+54001.
  This is a CONFIRMED upstream simulation-independence bug, not quality tuning.
- Replaced those proposed additive stream seeds with SHA256 domain-separated
  encodings of(namespace,world seed,channel). Checked ALL declared world/channel
  combinations for collisions. No true rate/noise parameter/threshold changed.
  Original V53 artifacts and existing global instructions untouched.
- Reading the full Golovin et al. paper showed its warning about myopic
  information gain and supplied Algorithm1's posterior-concentration criterion.
  Added that as a DISTINCT falsification policy before any cohort evaluation;
  retained information gain as its own arm. No inherited guarantee claimed.
- An apply_patch verification failed because a protocol hunk lacked its leading
  text. Atomic patch made no edits. Read the exact lines and reapplied; this
  tooling repair changed no scientific parameter.

Fixture preflight seed740541 is developmental and may be consumed in tests.
Main diagnostic2026105407 and reserved design/confirmation use the subsequently
frozen domain-separated stream specification. Do not reuse additive-salt results.

- `go test ./internal/researchdispersion -run '^TestPairedV54ControlMetricGuard$'
  -count=1 -v` failed with `control aggregate corruption accepted full 0`.
  Confirmed auditor early-return bug: controls had independent forecast and
  snapshot checks but skipped issued-loss/priority/recovery checks. Factored the
  same independent aggregate checks across BOTH controls and candidates; the
  failing regression must pass before the main freeze. No experiment data or
  model behavior changed.
