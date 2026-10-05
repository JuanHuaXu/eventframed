# Completion notification v7

Frozen before load execution. Fork v4 with the general adapter unchanged and
replace per-frontier1ms polling with WaitProcessed under the same5s timeout.
The worker now lazily broadcasts count changes to registered waiters. Preserve
original journal probabilities, label ordering, fitting cadence and publication.

Retain all18 arms and gates:64/192 requests, rotated off/queue16/queue64, three
trials, four readers, requests/2 future writes; no errors, overlap, >=80%
admission, all labels complete, zero worker failures, p99 ratio<=1.10 and age
p95<=250ms. Source/hash snapshots are exclusive-created with raw measurements.
Race tests precede execution. These are fixture load labels, not accuracy data.

Separate-run improvement is provisional; a successful screen still needs a
same-run polling comparison and replication. Do not change previous artifacts
or claim old embedded worker hashes match this updated worker.
