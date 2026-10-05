# Setup failure and preserved rerun

The first generation-load run stopped at800-event setup with:
"transaction commit failed: index flat transaction record seed-786: arena
exhausted: insufficient space for allocation". Process exited2. It printed two
completed200-record arms, but the final aggregate had not been written. Therefore
generation-load-results.json is an incomplete empty artifact and is not evidence
for their metrics. Retained stores remain available; do not overwrite or delete.

This is an observed backend transaction-capacity failure, not proof the host
ran out of system RAM. No backend patch is proposed from this error alone.

For v2 only change setup: seed in chunks of16, each with initial revision1,
before any reader/compactor begins. The initial corpus is one unserved logical
initial state. Measured writes, deadlines, rates, policies and corpus sizes stay
unchanged. The earlier index-scaling diagnostic already exercised bounded setup.
Also save each completed arm exclusively to a sidecar before continuing, so a
later failure does not erase its measurements.

New output: generation-load-v2-results.json, new private stores. No other tests
or builds run concurrently. Preserve original runner and failed artifact.
