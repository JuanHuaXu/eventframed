# Larger growth refutes whole-base compaction as the general rescue

The768d growth screen completed six arms:800/3200/6400 initial records, two
repetitions,1024 reads and512 writes per arm. Arrival rates and delta64/trigger32
remain unchanged. All sidecars, source hashes and9216 samples verify.

| Initial corpus | Write errors, repeats0 /1 | Read misses, repeats0 /1 | Late responses, repeats0 /1 |
| --- | --- | --- | --- |
|800|0 /0|0 /0|0 /0|
|3200|189 /191|0 /0|0 /0|
|6400|339 /352|2 /1|1 /3|

Both800-record arms pass. Both3200- and6400-record arms fail. All1071 rejected
writes report delta capacity;2001 acknowledged writes survive reopen with correct
revisions and no failed-present records. No operation-return errors occur on
reads, but three return without the self ID and four exceed100ms. Maximum read
latency146.46ms. Thus errors alone would miss both quality and latency failures.

Build durations scale from112-236ms at800 to487-629ms at3200 and1241-1442ms at6400.
The high-rate delta has320ms headroom after triggering, which the larger rebuilds
consistently exceed. The successful smaller bulk result remains valid for its
tested scope, but does not generalize to this larger growth screen.

## Next direction

Investigate a leveled/tiered immutable index: build new bounded runs and merge
selected runs, rather than rebuilding the full base after every small delta.
Do not just enlarge the buffer. Measure write amplification, query fanout,
compaction backlog, peak memory, version/tombstone shadowing and old-reader
retirement. Exact global merge correctness cannot be assumed from independent
approximate per-run queries, especially when stale versions dominate a prefix.
The observed self-query misses also require attribution; they are not proof of
an implementation bug or resolved by faster maintenance automatically.

This remains a component experiment, not full EventFrame journal/scoring or real
semantic-data validation. Streams last about5.12s before worker drain, not a
long-duration soak. All seven whole research goals remain open. No production,
dependency, paper or publication changes were made.

Verify: `node research/public-task-pilot/check-generation-growth.mjs`.
