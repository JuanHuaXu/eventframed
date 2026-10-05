# Durable research worker wrapper v7

OpenDurable owns a ledger and resumed background worker for one research stream.
Caller-retained monotonic learner IDs allow retry after lost acknowledgment.
Admission persists the original prediction before returning. Feedback validates
pending identity/time, commits before enqueue/acknowledgment, and exact retries
look up the durable record. Conflicting retries reject. An uncertain append or
post-commit enqueue error stops the wrapper until close/reopen and full replay.

The worker is private to preserve ordering between validation and persistence.
New researchledger.Get provides point lookup rather than a lifetime RAM map.
No discard API exists yet; dropping an unlabeled admission durably still needs
an explicit operation. Service event IDs and dependency authority are not wired.

The wrapper test runs48 admissions/feedback, retries both, rejects early and
conflicting outcomes, closes/reopens, compares all512 forecasts to an original-
record control and continues at ID49. Targeted and three full package race runs
pass; vet passes. This establishes an orderly acknowledgment/replay path, not
all failure boundaries. Prior abrupt-exit tests target the component harness,
not this new wrapper, and must not be relabeled as wrapper crash tests.

Remaining: wrapper-specific crash and I/O-error injection, failed-update status,
pending/discard lifecycle, durable latency/load, multi-process ownership
enforcement, service evidence/dependency binding and real predictive quality.
Each append synchronizes storage; no hot-path latency claim is made. No daemon
constructor uses the wrapper and no production configuration was changed.
