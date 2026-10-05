# Immutable runs: static tradeoff screen passes

Status: all16 stage arms PASS the frozen static screen. This supports building
the durable lifecycle prototype, not adopting the design in production.

Two repeats each start with800 records and apply eight32-mutation stages:
4 updates,2 deletes,26 inserts. Final live count992. Each stage compares a newly
built small run plus retained older runs against a freshly rebuilt current-state
graph. Both use the isolated bulk/candidate-only overlay,768-dimensional vectors
and CPU cap4. Build order and query order alternate as declared.

## Results

- Small-run build plus plan:16.19-29.29ms.
- Full current-state rebuild plus plan:79.18-99.53ms.
- Every stage in both repetitions has lower construction cost for small runs.
- All512 paired probes have exact top10 recall in both variants (1024 searches).
- Run-search stage medians:0.220-0.477ms; full-rebuild medians:0.190-0.271ms.
- Worst observed query:0.648ms for runs,0.426ms for full rebuild.
- At nine runs, median read overhead is about1.77x in both repetitions.
- Final retained-run close time totals216.589ms across both repetitions,
  recorded outside query/build timers. Each full-rebuild close is also recorded.

The independent verifier reconstructs all generated vectors, mutations and
exhaustive oracles; checks every candidate ID, score, ordering and deletion;
and checks source hashes, counts and timing gates. No candidate errors occurred.

## Reproduction and artifacts

```sh
go run -modfile=research-candidate-only.mod -overlay research/public-task-pilot/candidate-only-overlay-v1/overlay-local.json ./cmd/research-run-screen NEW-results.json
node research/public-task-pilot/check-immutable-run-screen.mjs NEW-results.json
```

Preserved output: `immutable-run-screen-results.json` and its exclusive `.stores`
directory. Protocol: `IMMUTABLE_RUN_SCREEN_PROTOCOL.md`. The verifier exits
successfully for structurally valid evidence even if a performance gate fails;
its final `allPass` field must be inspected. This output has `allPass:true`.

## Limits and next step

No concurrent ingestion, authoritative durable commit, cross-run retirement or
final consolidation was measured. Initial baseline setup is excluded; full-state
enumeration before each timed rebuild is excluded, while manifest planning is
included. These are build-plus-plan timings, not complete request or flush times.
Graphs share the same machine and deterministic data; two fresh graph repeats
do not provide a population confidence interval. Exact recall on these modest
fixtures is not a guarantee at larger corpus sizes.

Run count reaches9 but never its16-run limit. Sustained growth must pay compaction
cost before that limit, and repeated updates can hit the200-candidate prefix cap
earlier. Neither may be hidden by increasing the original global delta budget.
Next implement coherent durable append/drain and consolidation publication with
old-reader leases, then run original offered-load gates through repeated full
compaction cycles. Keep update/delete controls and explicit close/drain costs.
All seven whole research goals remain open.
