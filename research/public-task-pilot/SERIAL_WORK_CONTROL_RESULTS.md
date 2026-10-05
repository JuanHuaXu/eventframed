# Identical-state work-probe control

The new `build-serial-work-probe.mjs` generates separate control and instrumented
test overlays from the prior harnesses. It changes initialization from BatchInsert
to sequential Insert in fixed input order, and forces the final deletion to target
the actual current entry point. Production/backend sources remain untouched.

Both runs passed: control6.507s, instrumented6.407s. These whole-test times include
setup/capture and are not operation-latency comparisons. The instrumented run
adds synchronization; no speedup claim follows from elapsed test durations.

A bounded comparison found zero differences across all 32 operations' before/after
node records and packed global state. Unlike the earlier parallel-initialization
fixtures, this pair therefore supplies an identical-state instrumentation control.
One pair does not prove universal determinism, nor isolate the exact source of
the earlier BatchInsert differences. The production initialization path is not
changed or declared defective.

| Corpus | Forced deleted entry point | Registry slots scanned | Changed records |
| --- | --- | ---: | ---: |
| 800 | seed-24 | 808 | 62 |
| 6400 | seed-2709 | 6408 | 87 |

This directly confirms that few changed records do not imply bounded discovery
work: the entry-point fallback scans the full registry. The new controlled
fixture is suitable for testing a persistent entry-point summary that preserves
highest-level/minimum-ordinal replacement. Other insertion/deletion work and
durability remain outside such a rescue.

Artifacts: `serial-work-control.json`, `serial-work-measured.json`, and
`serial-work-comparison.json`. Earlier failures and raw captures are retained.
All seven whole goals remain open.
