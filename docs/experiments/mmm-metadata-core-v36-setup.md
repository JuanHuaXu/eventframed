# V36 Setup Repairs

2026-10-03. First compile failed because the new differential test referenced
the load wrapper's nonexistent `authority` field. It belongs to its embedded
metadata store. Fixed the three test references; no runtime or frozen source
change. No normal trial had begun. Subsequent failures, if any, stay recorded
separately rather than overwriting consumed normal evidence.

First auditor self-test failed duplicate-core-root: its consumed V35 fixture
has one copy per request, not one per core. Deduplicated that IN-MEMORY fixture
by root with earliest recorded capture time and recomputed fixture bytes/hits.
Actual V35 raw data and V36 runtime remain unchanged. This is an auditor fixture,
not a V36 normal result or a performance counter repair. Manual race controls
passed both storage arms before prospective preflight.
