# V3 Quantizer Sample Publication Rescue

Frozen after the complete V2 candidate/200 race run FAILED, before V3 execution.
The ordinary V2 pilot passed finite timing gates but cannot be promoted.

Confirmed root: `trainQuantizer` clears `trainingVectors` while another batch
worker writes a reserved slot. The trained flag is already atomic; it is NOT
the raced field. Atomic reservation count is not a completed-sample barrier,
so training can also inspect incompletely populated slots. The authoritative
race stack is retained in repaired-candidate-load-race.txt. Upstream hnsw.go
matches the pinned source; earlier all-state issue/PR inspection remains scoped.

V3 is a NEW temporary fork. V2 sources and timing artifacts remain immutable.
Serialize the bounded, cold training-sample collection and training under one
mutex, recheck the atomic trained state after acquiring it, and publish trained
only after valid completed samples are trained. Protect related training-buffer
memory accounting/clearing with the same mutex. Warm trained insertion should
still bypass this mutex. Keep parallel graph construction, quantization and WAL
publication; do not turn off workers or SQ8 to hide the race. A missing buffer
fails closed rather than indexing nil. Training failure keeps completed samples
available for a later retry instead of overflowing the sample count.

First run a public-vector 1,000-row parallel SQ8 batch regression under race
on unpatched V3 and preserve its result. A no-race sample would NOT disprove the
already witnessed full-load race. After patch: repeat normally and under race,
plus existing HNSW quantization/regression tests. This is a component preflight;
full V3 loaded integration and its performance remain separate requirements.
No V2 pilot claim, source hash or success threshold is rewritten.
