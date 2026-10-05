# V44 load measurement preflight

2026-10-03. Terminal correctness PASS in
`research/eager-load-v44-preflight-authority-selector-repair/`.
Exact1212-source copies and SHA inventory, six command logs/metadata, and
`completed.json` preserve the executable preflight. This is NOT a load result.

All13 required root tests actually execute (no skip substitution): three new
V44 predicate/workload/owner-durability roots, four V42 wire/owner/mutation/gap
roots, three V37 generation/warm/cancellation roots, and three V33 archival
retry/lifecycle/interruption roots. The explicitly enabled native fixtures
remain temporary and isolated. Separate admission/validity race suites repeat
three times; vet passes. The normal load root deliberately skips here because
there is no normal-output environment or frozen normal manifest yet.

The independent auditor accepts a consumed V37 eager-off row, then an explicitly
SYNTHETIC logic fixture with operation-bound preparation traces and one unused
publisher. It rejects49 prefix/source/forecast/wire/queue/byte/timing/gate/core/
publication/operation-identity/acknowledgement/warm corruptions. That fixture is
NOT an eager load observation and supplies no new latency or learning evidence.
Four fake process commands and four positive command kinds also verify that a
watcher's own Node source is not mistaken for a running Go auditor.

Every real enabled journal/outcome/write will record immutable operation IDs.
All128 loaded journals bind to returned request windows and full durable batches,
all16 outcomes to recorded idempotency keys and acknowledgement windows, and
all128 writes to exactly one preparation/batch ack. ALL constructed core bytes
count, including unused publications; read-build/hit conservation is separate
from publisher work. Same-head read fields and the first immutable certificate
cannot drift, and every core replays to a full committed witness prefix.
The original prime nomination/independent returned-ack scope gap remains
explicit, not repaired by fabricating a timestamp.

Preserve the first compile failure in
`research/eager-load-v44-preflight-initial/`; its six invalid fixture selectors
are explained in [preflight repairs](mmm-eager-load-v44-preflight-repairs.md).
The one-off self-matching watcher was refused BEFORE any test by the independent
timing guard; [watcher repair](mmm-eager-load-v44-watcher-repair.md) documents it.
No earlier failure is overwritten or reclassified as a pass.

Next, after V43 collection AND offline audits terminate, run:

```sh
node research/eager-load-v44-run.mjs research/eager-load-v44-preflight-authority-selector-repair
```

The exclusive normal runner verifies this exact preflight freeze, copies its
sources, executes the complete16-trial original offered workload, and applies
the independent auditor. Original100/250ms gates remain unchanged. No loaded
interpretation during concurrent V43 CPU work. All seven WHOLE goals OPEN/ACTIVE;
production/private corpora/whitepaper remain untouched by this work.
