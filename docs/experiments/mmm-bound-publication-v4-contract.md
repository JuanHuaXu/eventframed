# Guarded bound-model publication v4: frozen research contract

2026-10-01. Goal 6 continuation after durable source lineage and bounded
as-of-motion restoration. This is an opt-in, immutable research model slot,
not a production serving path or a replacement durable learner.

The candidate's labels come from one exclusively owned durable stream, not a
caller-supplied subset. Scan every canonical admit and terminal record in
sequence. Include every feedback with `Available <= cutoff`, exclude later
feedback and explicit discards, and reject malformed identities, duplicate
terminals, missing admissions, or more than the existing 256-label rebuild
cap. The declared cutoff is fixed before preparation. A source validator
checks each original journal, keyed witness, event, and event-continuity proof
against one exact target snapshot. No post-outcome recomputation of original
forecast features or mixture weights is allowed.

Preparation occurs outside the store mutation gate. It produces an immutable
frozen model and records its target snapshot, cutoff, epoch, retained count,
the slot revision and durable-log sequence observed at preparation. Publication
must acquire the research wrapper's exact-snapshot writer guard, then hold
the durable stream's ownership mutex while proving no log entry arrived after
that sequence and atomically comparing/swapping the slot. This lock order is
store then durable, matching guarded feedback admission. A store mutation,
new admission/label, or competing slot revision between validation and
publication must reject the candidate, not silently rebind it. The new epoch
must exceed the current one. No partial model is installed on any error.

A reader may score only through the slot's as-of guard, using the published
target snapshot and the reader's `at`. A later future-only ingestion may be
compatible for an earlier `at`; a source deletion, general mutation, or event
that is already available at `at` must reject the old model. The frozen model
itself also rejects `at` before its latest admitted feedback. A rejected
score returns an error so callers can use an explicitly declared fallback;
it must not return a stale probability.

Frozen checks before results:

1. Persistent A/B guarded forecasts and delayed labels replay from SQLite.
   Preparation reads the complete durable stream, retains only surviving
   sources, publishes a new epoch under the exact target guard, and exposes a
   score through the as-of guard. A caller cannot omit a bad labeled ID by
   passing a preferred subset.
2. Delete a source between preparation and publication: publication rejects
   and leaves the old slot unchanged. Delete a source after publication:
   scoring rejects. Future-only ingestion permits scoring only before that
   event's availability. An appended log entry, older/equal epoch, or a
   competing slot update rejects.
3. Incomplete, discarded, malformed or beyond-cutoff terminal records cannot
   be silently converted into negative or early labels. Run focused race,
   full affected-package tests, vet, and separate preparation/publication/
   read cost measurements. Preserve negative results without tuning gates.

This does not yet persist the new epoch's learner state or resume future
training from a published frozen model. It also does not measure loaded
service p99, real agent outcomes, or power-loss recovery; Goal 6 cannot be
declared complete from this component screen.
