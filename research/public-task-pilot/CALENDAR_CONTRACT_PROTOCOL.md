# Pre-packing calendar contract experiment

Freeze before execution. Research-only Go adapter through existing CandidateRanker
interface, no service.go or research_rank.go changes. Actual CaptureTurn -> Recall
-> calendar ranker -> packing.Select. Four records, recall50/pack1, isolated memory
service per query. Run focus control and calendar arm on W3C/ESA written/ESA ISO.
All data consumed; this is integration verification, not fresh accuracy evidence.

Both arms use the same focused query for retrieval. Calendar binds the ORIGINAL
query in request context, checks tenant and equality of the canonical focused
query, and reads historical dates only from actual candidate FrameText's what
field. No fixture corpus lookup or oracle inside the ranker. Only after results
are produced may IDs join the evaluation oracle. A caller must bind original
query before rewriting; this research command does so explicitly.

Snapshot/retry mechanics stay owned by service. The adapter has no cache, mutable
shared state, fitting, I/O or retry journal. It validates frontier<=200, K1/K2,
unique IDs, finite scores and cancellation. It retains unknown records and
preserves order on unsupported or all-contradicted inputs. It does not claim
temporal truth from ingestion timestamps or resolve multi-event attachment.

Because packing sorts/uses score, mixed partitions receive strictly ordered
ordinal scores, not probabilities. These are not calibrated certainty. Later
runtime deltas could change that order: this empty-cache experiment does NOT
prove hard precedence against learned deltas. Proper forecast laws must remain
equal to control for EVERY admitted event in the Bayesian report, not just the
single packed winner. The contract name participates in existing fingerprinting.

Record actual ranker before/after inputs, final packed winner, and full frontier
forecast laws. Success: reproduce the offline normalized top1 outcomes while
packing only1, rescue at least one control-missed winner, preserve all correct
control cases, and preserve all forecast laws. Read tests separately from the
scientific pass/fail. No performance, calibration or general learning claim.

Required before broader adoption: snapshot-race/retry and request-binding
integration tests, durable explanation journal, learned-delta interaction,
real agent outcomes and serving tail-latency tests. Passing this experiment
does not complete the original runtime-boundary contract or any whole goal.
