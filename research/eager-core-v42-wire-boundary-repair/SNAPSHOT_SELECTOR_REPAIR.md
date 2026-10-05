# Technical compile correction

This run failed before executing fixtures: the new corruption control used
Snapshot.Epoch, but the declared field is Snapshot.EvidenceEpoch. The one-field
repair changes only the negative control; no model, native operation or gate.
The old freeze/log/failure remain. Reverse this one selector to recover its
candidate SHA256. The next invocation has a new label and freeze.
