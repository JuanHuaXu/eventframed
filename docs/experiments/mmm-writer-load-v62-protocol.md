# Scheduled native writer concurrency v62

Frozen before measurement. v61's isolated concurrent gain did not pass its full
screen because one sequential p99 regressed. Keep that failure. Test whether
four per-collection writer slots improve actual mixed service load without
weakening WAL durability, index synchronization, snapshot guards or identity.

Four arms: native off/two slots, native off/four slots, resolved source
background/two slots, resolved source background/four slots. Rotate order over
three trials, twelve fresh databases. Use existing offered-arrival fixture:
192 recalls at5ms,96 future event writes at10ms, four reader lanes, one writer,
50 visible public events, recall50/pack10, queue64, groups up to4, guard entry
budget20ms. Preserve original due times when producers run late. This is cold
storage/observation work without training labels, not an accuracy experiment.

Collect per-call scheduled/inside times, writer lateness, observation age, phases,
completion/drop/expiry, full source-original readback and terminal accounting.
Capture internal Go sources and hashes, module pins, protocol and runtime in
exclusive JSONL. Run service race tests and vet before non-race measurement;
no competing task-started benchmarks. No production/configuration changes.

Primary rescue screen, in every trial: four-slot active has all192 observations
completed with no drops/expiry/errors; age p95 <=250ms; scheduled read and write
p99 each <=1.10 times the matching four-slot native-off arm. All integrity and
schedule checks must pass. Report two-slot active versus two-slot native-off
under the same rules as a control. Also report direct four/two comparisons for
both native-off and active arms, without retroactively redefining success.
Use nearest-rank percentiles. This finite diagnostic is not a statistical
non-inferiority guarantee, production throughput claim or completion of the
learning research directions. No default promotion based on this run alone.
