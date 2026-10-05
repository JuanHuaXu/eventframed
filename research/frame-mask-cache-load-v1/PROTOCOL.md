# Repaired-control mask-cache loaded regression

Frozen before execution, 2026-10-05. Follow-up to the mask-reuse component
preflight, not a new definition of research goal 6 success.

Control: published `f3231fab2244c5c6bca8f7f822e5669ac58d13cd`.
Candidate: exactly `research/frame-mask-cache-v1/candidate.patch`, whose SHA-256
must match the frozen preflight manifest. No subsequent code change is allowed
inside this comparison. Production, private data and live extractors are untouched.

Run the unchanged `TestResearchGuardedDurableLiveFreshnessV2` fixture in full
isolated checkouts, with `EVENTFRAME_RESEARCH_PERF_GATE=1`, ordinary builds,
`-count=1`, and `-mod=readonly`. Use control-candidate then candidate-control
order, with no competing local test/benchmark started by this runner. Each
command runs quiet and future-writer arms, three independent ephemeral-store
trials each, 64 guarded labels per trial. Retain failed command output and
continue the planned comparison; a nonzero exit is not converted to a pass.

The fixture retains its own validation: no future event enters the as-of
frontier; bound admission/feedback and complete 128-row per-trial journal replay
agree; all 64 labels complete before Close; concurrent future writes overlap.
Timing gates remain writer recall p99 <100 ms and offered-label to observed
published completion p99 <250 ms. No -race timings are compared to ordinary
build gates. Prior race tests remain functional checks only.

Additional paired regression screen: candidate/control recall p99 ratio <=1.10
for quiet and future-writer arms in BOTH paired repetitions. A failure means
loaded non-regression is not established here, regardless of absolute gates.
These paired order statistics are descriptive screens, not population-tail
confidence guarantees. Record the full command, raw output, source identities,
case counts, writes/overlap, label counts, p50/p95/p99 and maximum live age.

Important boundary: this old fixture has only one eligible candidate and uses
structured Observe for future writes. It exercises QueryText but NOT CaptureTurn
for the writer. It does not cover real outcomes, visible mutations, cross-epoch
learning continuity, sustained backlog, 50-200 frontier scale or a large corpus.
Do not attribute a timing difference to quote caching or declare goal 6 complete.
Afterward, build a separately frozen capture/large-frontier mixed fixture or
pursue state-transfer/freshness issues exposed by authoritative results.
