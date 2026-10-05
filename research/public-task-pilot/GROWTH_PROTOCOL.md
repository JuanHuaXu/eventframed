# Larger-base longer-stream saturation screen

Repeat bulk immutable-base serving at768 dimensions with initial corpora800,
3200 and6400. Two repetitions, fixedCPU4. Each arm schedules1024 reads at5ms
intervals and512 writes every10ms, all with100ms scheduled-arrival deadlines.
Thus the800-record arm is also a longer-stream check of the previous passing
small-corpus condition. Retain delta64, trigger32, leases8, retired2,5ms compactor
poll, synchronous durability and the same bulk builder.

Setup remains unserved16-record transactions at initial revision1. Increase the
outer arm timeout to120s solely for setup/build/reopen accounting; individual
operation deadlines remain100ms. No request retries. Write growth remains
observable rather than hidden by discarding failed operations. No other tests
or builds run concurrently. Source hashes and per-arm sidecars are required.

Use generation-growth-results.json and new private stores. Require all six arms
to pass the existing full operation/build/reopen gates to claim success here.
Do not enlarge the delta or lower arrival rates after seeing results. If larger
bases saturate, revisit whole-base compaction itself rather than moving the
failure beyond another finite buffer. This is still not semantic validation or
the full EventFrame scoring/journal path.
