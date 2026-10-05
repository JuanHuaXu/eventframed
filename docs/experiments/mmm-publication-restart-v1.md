# Orderly publication adapter restart

TestOrderlyRestartDoesNotInventPublicationHistory uses a temporary persistent
database and three adapter lifetimes. Three package race repetitions and vet
passed. No running agent, production database or executor is involved.

Verified: future ingestion supports an earlier-as-of proof within one lifetime;
Close revokes the old adapter's authority; orderly reopen restores the committed
snapshot, event and vector. The new adapter accepts its current snapshot but
rejects older snapshots requiring absent motion history. New lifetime ingestion
can build new proofs, and a policy change invalidates its earlier snapshot.

The test also exercises a known duplicate: the current adapter quarantines
instead of resolving it. After orderly Close/Open, backend state is intact and
the new adapter can operate, without granting old motion proofs. This controlled
case is not crash recovery or a general prescription to clear quarantine after
an ambiguous transaction. Duplicate handling remains an availability limitation.

Not tested: power loss, interrupted writes, corrupted storage, mixed mutations
during restart, persistent worker queues or feedback replay. The publication
adapter deliberately has no durable motion ledger. A restarted learner must not
assume this test restores its observations, mixture state or duplicate history.
