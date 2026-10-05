# Bound-label rebuild v1: frozen component contract

2026-10-01. Goal 6 research only. The current Adapter stores bare feature/outcome
samples, and a visible source mutation invalidates its existing publication
epoch. It cannot safely copy fitted weights or infer which labels survive from
its aggregate state. Durable original forecast records do carry service source
bindings. This experiment tests a bounded reconstruction input, not a live
bridge or automatic authority decision.

Input is at most 256 ordered pairs of immutable original-forecast and terminal
feedback records, including the original service binding, issued feature bits,
query and outcome-availability times. A required caller validator receives the target snapshot and each
record. It must check current source/journal validity and feature semantics;
returning true is an assertion of external authority, not a fact the learner
can infer. Invalid shape, duplicate journal/event identity, future feedback,
out-of-order availability, mismatched prediction/feedback IDs, validation errors,
or missing authority abort the
whole rebuild without exposing a partial model. False validation drops that
record. No pending predictions transfer.

The rebuilt epoch keeps only approved samples (up to 256), fits short/long
count models and the bounded tree forest once when at least 32 survive, and
resets inner/outer forecast-mixture weights. The old mixture's pre-outcome
expert forecasts are not transferable from bare samples. Fewer than 32
approved labels leave the model cold, using the supplied baseline. The
earliest allowed score time is the last approved feedback availability.

Predeclared checks:

1. With 64 alternating source-A/source-B labels and a target that rejects A,
   only the 32 B samples influence fitted component predictions. They must
   agree with fresh direct fits on the B subset; all-A rejection stays cold.
2. Reversing B outcomes must change at least one component prediction; changing
   only rejected A outcomes must change none. The old adapted mixture must not
   survive in either reconstruction.
3. Reject duplicate identity, bad order, future/early feedback, wrong tenant,
   nil validator, and validator error without returning a model. Cross-epoch
   scores reject the old epoch.
4. Measure complete offline reconstruction time and allocations for 64 and 256
   input records. A p95 below 100 ms is a finite feasibility screen, not
   serving latency. Retain any failure rather than changing the cap afterward.

This component cannot certify its own validator, distinguish copied labels
from independent evidence, or rederive features after graph change. Those
remain service-side obligations. Full durable recovery, arbitrary mixed
mutations, crash semantics, and end-to-end agent latency remain open.
