# Bounded materialized batch-read comparison

Reuse envelope-batch-read-v1's6400 originals,1KiB padding,grouped/scattered
layouts,batch50/200,32 calls/cell. Three trials rotate original indexed batch
reader and materialized batch reader. Same bounded projection/shared validator,
one read transaction, caller order and8MiB request/output caps. All preparation
inside read timing except fixture/request-list construction in both arms.

Validate every source,sequence,key and payload. Capture raw timing/source text;
no exclusions or gate changes. Candidate remains admission-only/unintegrated.
Require caps,cancellation,panic cleanup tests before collection. This diagnostic
does not establish process-crash recovery, migration or full-service latency.
