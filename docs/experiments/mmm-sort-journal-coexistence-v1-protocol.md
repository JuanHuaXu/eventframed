# Sort publication and Recall journal coexistence v1: frozen probe

Start one private, verified incremental sort-key publication gate with its
three EventFrames. Capture its published LSN and runtime snapshot. Call the
real `libravdbstore.Store.PutBayesianJournal` with an exact current-snapshot
frontier entry, as a normal Recall does. Then measure the database's latest
LSN, runtime snapshot, gate READY status, and whether one later authorized
EventFrame batch can publish. Check the journal can be read back and no
future-available event enters an exact-LSN search.

The coexistence condition is that a valid metadata-only frontier journal
must not permanently invalidate EventFrame publication or prevent later
authorized appends. It must not be silently counted as an EventFrame. A
failure is an integration blocker, not a reason to skip the journal in the
service or weaken as-of checks. This is a test-only, single-owner diagnostic
with no latency gate and no production change.
