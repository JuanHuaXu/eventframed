# Entry summary in the backend overlay

`build-summary-backend.mjs` layers the persistent summary into the serial work
probe. Single-node metadata insertion adds its ordinal/level, retirement removes
it, and entry-point replacement queries a private summary with the deleted
ordinal excluded instead of scanning the registry. The generated files are Go
build overlays, not edits to the pinned backend or normal daemon dependencies.

## Observed result

The real backend32-operation test passed in6.387s. A bounded record-by-record
comparison found zero before/after node or packed global-state differences from
the identical-state serial control. Both forced entry-point deletions performed
one summary lookup and zero registry-scan iterations:

| Corpus | Deleted entry | Previous scan iterations | Summary lookups |
| --- | --- | ---: | ---: |
| 800 | seed-24 | 808 | 1 |
| 6400 | seed-2709 | 6408 | 1 |

Results are in `summary-backend-results.json` and
`summary-backend-comparison.json`. Whole-test elapsed time includes initialization
and captures and is NOT evidence of an end-to-end latency improvement.

A second explicit run under the race detector passed in35.087s; its captured
states also have zero differences from the serial control. The separate output
is `summary-backend-race-results.json`. This tests the serial fixture under race
instrumentation, not concurrent-writer correctness.

## Scope and cost

This is an actual backend-algorithm substitution in a controlled serial fixture,
not merely replaying a precomputed replacement. Graph reconnection is unchanged.
Each inserted/retired ordinal also maintains the persistent summary. Entry-point
deletion currently prepares an exclusion summary for selection and later updates
the retained summary during retirement, so it performs duplicate path work in
that case. Maintenance allocates and has not been included in a paired benchmark.

The fixture intentionally uses serial Insert initialization. BatchInsert,
snapshot load, reset, concurrent-writer atomicity, and durable publication are
NOT wired to this summary. A mutex protects the summary object but does not make
summary and mutable graph publication transactional. Using this overlay as a
general backend replacement would be incorrect. No production adoption is made.

Next, cover lifecycle integration and measure total maintenance plus deletion
cost with interleaved controls. This removes one scan mechanism only; it does not
resolve private neighbor repair, byte budgets, or the sustained-load failures.
All seven whole research goals remain open.
