# Guarded source-owner load v50

FAIL: source-owner completion age exceeds the unchanged 250ms screen in every
trial. Serving-read p99 passes the same-trial 1.10x-off screen, but this does not
offset incomplete observation processing or writer overhead. The isolated v49
improvement was real; it did not establish loaded capacity for the new owner.

## Method and integrity

The frozen protocol runs off, prepared raw durable/union validation, and source
owner/batch discard in rotated order over three trials. Every cell uses 192 real
Recall requests and 96 concurrent future-dated Observe writes in isolated
LibraVDB, with fixed as-of and 50 overlapping visible events. No labels, fitting,
production instance, private data, or served forecast influence.

The source path passes no learner ID to admission. The cold preview is only a
guard-validation envelope; the owner emits the stored original, whose ID is
checked for expected uniqueness/order. All other fields must match the cold
envelope. Actual returned originals, not adjusted previews, are used for complete
post-guard readback and source-key atomic discard. This does not validate warm
preview remapping or confer evidence authority.

Command: `EVENTFRAME_SOURCE_LOAD_ARTIFACT=.../mmm-source-load-v50.jsonl go test ./internal/service -run '^TestResearchSourceLoadExperiment$' -count=1 -v`.
Completed in 15.62s on Go1.27.1 darwin/arm64, Apple M4, GOMAXPROCS10.
All nine cells passed request/phase/original/terminal accounting. Source hashes
and aggregate counters were independently recomputed from the artifact.
Targeted source/control/guard race tests passed three repetitions; the full
ledger, learner and service race suites and their `go vet` also passed.

## Results

| Trial | Raw/source completed | Source drops | Raw/source age p95 ms | Off/source read p99 ms | Off/source write p99 ms |
|---|---|---|---|---|---|
| 0 | 192 / 134 | 58 | 205.247 / 538.502 | 32.003 / 23.596 | 18.490 / 29.901 |
| 1 | 192 / 134 | 58 | 192.124 / 561.924 | 32.990 / 29.343 | 18.865 / 25.840 |
| 2 | 192 / 127 | 65 | 178.987 / 591.132 | 34.603 / 30.093 | 17.847 / 36.730 |

Source completes 395/576 observations (68.58%), dropping 181 at the bounded
queue; no entry expiry, unexpected error or record mismatch occurred. The raw
prepared control completes all 576 and passes both finite screens in every
trial. Completion ages describe accepted observations only, not dropped work.
All 19,750 accepted source originals and 28,800 control originals were read back
and durably discarded. Dropped observations do not become negative labels.

Source write tails exceed off in every trial. Relative to the raw durable control
they improve in trials0/1 and worsen in trial2: raw write p99 is 31.056, 30.378,
23.436ms. No broad serving non-harm or throughput improvement is claimed from
the passing read-tail screen alone.

Mean accepted-group callback time is 10.85-12.37ms source versus 6.75-7.43ms raw;
admission is 9.20-10.29ms versus 4.45-5.13ms. Post-guard work is 17.62-17.86ms
versus 5.97-6.59ms, comprising source readback 6.91-7.19ms and discard
10.63-10.87ms. The raw counterparts are 1.96-2.01ms and 4.00-4.62ms. Group sizes
are recorded rather than assumed equal; these are observed group means, not
matched per-record causal effect estimates.

## Next lead

Repeated per-source indexed resolution occurs in admission, original readback
and batch discard. v49 removed repeated terminal transactions, not these read
calls. Query setup, row decoding and index/data access are credible remaining
costs; phase timings alone do not distinguish them. Profile the unchanged load
before selecting a bounded transactional source-read batch implementation. Keep
the five-field identity index, canonical originals, missing/error distinction,
byte limits, atomic discard and all service authority checks.

This failure does not justify relaxing staleness, dropping durable records, using
a lifetime in-memory identity map, or bypassing source checks. Feedback/history
authority remains independently open. All research directions stay in progress.

Artifact: `mmm-source-load-v50.jsonl`.
SHA-256: `c956e0b015b292a89dc19f3aebadb31f82b99417a71ea59566d11f199c004392`.
