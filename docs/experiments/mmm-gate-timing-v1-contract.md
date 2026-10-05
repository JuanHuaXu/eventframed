# Direct publication-gate timing

Measurement, not a semantic rescue or candidate promotion. Preserve existing
v62/v64 failures. Normal constructors remain uninstrumented. Research wrapper
records monotonic wait and held durations in a bounded recorder; no data payloads.
Release the publication gate before recording to avoid extending its held time.
Record failed acquisition separately; errors/panics must still release the gate.

Three trials, four rotated arms: idle wrapper/unmeasured, idle wrapper/measured,
active resolved source/unmeasured, active resolved source/measured. All use four
native writer slots, 192 reads at5ms,96 writes at10ms, the existing public
50-event fixture, queue64 and groups4. The idle wrapper control isolates
background work rather than conflating it with installing the wrapper.

Measured cells have recorder capacity4096; no silent loss accepted. Capture
full internal sources/hashes, module pins, protocol, runtime and timing spans.
Report ingestion/general/as-of gate wait and held distributions, cumulative
occupancy, original scheduled read/write tails, observation age, and all
completion/drop/expiry accounting. Do not add overlapping outer phases.
Use matched unmeasured cells to disclose instrumentation effect. No claimed
non-harm rescue or power-loss/durable-learning evidence from cold observations.

Run wrapper race tests, four-arm integrity race smoke and vet before serial
non-race measurement. Production, credentials, external services, dependencies,
whitepaper and remotes remain untouched. No competing benchmarks during run.
