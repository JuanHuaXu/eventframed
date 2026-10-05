# Durable wrapper acknowledged-exit recovery v8

TestDurableWrapperAcknowledgedRecovery uses the actual OpenDurable/Admit/Feedback
API in child processes, each on a temporary private database. It terminates with
os.Exit(25) after64 acknowledged admissions without labels, after64 acknowledged
feedback calls while fitting is blocked, or after all64 updates have completed.
The parent verifies the expected child exit, reopens through OpenDurable and
compares against an uninterrupted fixture with the same issued forecasts.

All three cases pass three race-test repetitions; vet passes. Pending recovery
has zero learned labels and64 original pending records; queued/applied recovery
has64 learned labels and zero pending. All512 feature forecasts match control.
Admission retries return the original .6 forecast, feedback retries deduplicate,
and a recovered unlabeled prediction accepts later feedback and completes.

This is wrapper-specific abrupt-exit evidence, unlike earlier component-only
tests. It covers known boundaries after successful acknowledgment, not a crash
inside commit, write errors reported after commit, arbitrary scheduler positions,
checkpointing or multi-process ownership. No learner checkpoint is persisted;
reopen rebuilds the complete ordered log. Durable latency/load, discard semantics,
service event/dependency binding and real predictive quality remain open.
