# Durable unlabeled abandonment v11

Durable.Discard writes a distinct canonical RecordedDiscard before releasing a
pending prediction. It uses the existing unique terminal slot historically named
feedback, so one admission cannot independently acquire a label and a discard.
The payload is not RecordedFeedback and cannot default to Useful=false. No schema
migration or old artifact rewrite was needed. Older strict replay code rejects
the unfamiliar payload rather than silently learning from it.

Replay validates binding/time and removes the pending identity without updating
samples, mixture weights or the evidence clock. Exact discard retries deduplicate;
different retry times conflict. Label-after-discard and discard-after-label both
reject. A write error stops the durable wrapper under its existing policy.

The test records300 admissions/discards, closes/reopens, and observes zero labels,
failures, pending records and queued work. It checks retry, terminal conflicts,
early discard and continued learning on ID301. Three full researchmemory and
researchledger race runs plus vet pass, including prior recovery fixtures.

Discard-specific crash and uncertain-write cases remain to be exercised directly.
Abandonment timestamps do not advance evidence time because they are not observed
outcomes. Service expiration policy, dependency authority, multi-process ownership,
log retention and loaded durable performance remain incomplete. This removes a
research wrapper lifecycle gap, not the service bridge's existing lifetime cap.
