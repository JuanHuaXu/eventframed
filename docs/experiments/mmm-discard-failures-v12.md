# Durable discard failure boundaries v12

Uncertain-write tests now cover discard as well as admission/feedback. An error
before append restores one pending prediction; an error after real commit
restores zero pending. Both stop the wrapper, and retry after reopen consumes
the identity exactly once with zero learned labels. Further labels reject.

The actual-wrapper child-process suite additionally exits after64 acknowledged
discards. Reopen has zero labels and zero pending, all512 forecasts match the
uninterrupted unlabeled control, repeated discards deduplicate and labels reject.
Existing pending/queued/applied exit cases remain covered.

The six-case uncertain-write suite and four-case wrapper-exit suite each pass
three race-test repetitions; vet passes. These are known API boundaries, not
physical disk failure or arbitrary commit interruption. The implementation was
not changed to make these tests pass; this turn extends failure evidence.

Service event/dependency binding, ownership enforcement, log retention and
loaded durable serving remain open. The completion of these fixture lifecycle
checks does not complete direction6 or the broader research objective.
