# Snapshot-specialist gate and observation integration

Subsequent [fresh v115 quality comparison](../docs/experiments/mmm-snapshot-v115-results.md)
FAILS with264 failed gates. Lifecycle correctness below is not a quality pass.

Status: lifecycle integration checks PASS; fresh quality comparison UNTESTED.
No production or whitepaper promotion. This extends the
[filter/detached-law component](snapshot-specialists-component-results.md).

## Gate contract and confirmed integration repair

Tests are keyed by publication window and immutable snapshot ID. At most eight
slots across eight windows and two reserved views spend the .01 family allowance;
the boundary is12800. Each test mixes32 predeclared starts, with alternative
mass .5 on neutral and .5 divided among the other active snapshots. No
independence of alternatives is claimed. Non-rejection does not establish truth.

The first prototype reset all rejection state on publication. Inspection found
that this would resurrect an already-rejected retained snapshot. Corrected:

- Rejection persists while the same immutable snapshot remains alive.
- First-crossing credits survive only for neutral and recipients whose IDs
  still match. Retired recipients cannot transfer credit to a new slot occupant.
- Unrejected retained snapshots start their separately budgeted next window.
- The helper rejects same-window restarts, identity relocation, and changes to
  the boundary mid-run. Journal publication enforces the32-frame schedule.

New generation IDs receive fresh tests; retained rejection is not copied to a
new fitted model. This is intentionally conservative for a retained predictor
that might become useful again: quality tests must measure that tradeoff.

An explicit sum over every start and alternative agrees with the log recursion
and first-crossing credits. The four-active-role specialization at boundary6400
agrees with archived routing. This is arithmetic evidence, not empirical
confirmation of the conditional null or a new target-law diameter certificate.
The usual likelihood-ratio claim requires the null forecast to be the true
conditional law in the relevant filtration; informative missingness, delayed
selection and dependent outcomes need their own justified assumptions.

## Complete journal

The existing six-coordinate entropy observer is mechanically cloned onto the
eight-snapshot law without changing the archived observer. Hidden coordinates
still enter only through the reader. Rejected mass goes to neutral or accepted
recipients, never recursively through another rejected snapshot. Preview and
served forecasts use the same routed law.

Each issue stores origin, publication window, snapshot IDs, raw probabilities
and served probability. Advice refilters on arrival. Gate evidence drains in
origin order; a stale-window outcome remains advice-only and cannot update the
current gate. Expiry censors unknown messages without converting them to labels.
Publication, advice, gate updates and journal accounting commit atomically.

## Verification

- Gate reference, four-role compatibility, retained-rejection, recipient-ID,
  all-rejected fallback, window-budget and malformed-input checks PASS.
- Full256-frame journal check covers eight publications, delay/missing feedback,
  reverse releases, expiry and flush. Advice matches independent dense history;
  independent counters match gate-window ownership and journal accounting.
- First-bank observation paths match the archived controller's views, costs and
  probabilities. The issued probability is reconstructed from stored raw values.
- Reader callbacks cannot reenter publication, prediction, clock, delivery or
  expiry. Injected read failure, post-read filter failure and downstream gate
  failure all preserve the pre-call owner state.
- Final snapshot race suite PASS,3.465s; combined advice-component race suite
  PASS,3.654s before the additional window-budget guards; final vet and format
  checks PASS. No remaining defect found in this scoped audit.

## Measured cost

[All18 benchmark repeats](../docs/experiments/mmm-snapshot-journal-benchmarks.txt),
Apple M4,300ms/count3, per256-frame lifecycle with prebuilt models:

| Feedback | Arrival log | Fixed Markov | Snapshot journal |
| --- | ---: | ---: | ---: |
| Immediate |1.641-1.664ms|1.987-1.989ms|9.385-9.701ms|
| Delay31 |1.294-1.300ms|3.041-3.064ms|17.701-17.832ms|

Snapshot allocation is about10.24MB per run, primarily eight detached
publications; this is allocated volume, not retained memory. Controls keep their
archived four-role6400 gate, while the candidate uses eight slots and12800.
Thus these timings are a workload comparison, not a matched gate ablation.
No model fitting, storage, network, serving load or p99 measurement is included.

Gate update arithmetic is O(M^2), routing O(M^2); observer cell evaluation O(M)
times its bounded view/completion search. Prefix draining may apply up toD
buffered outcomes in one call. Transactional copies touch the full fixed
capacities in addition to live work. Do not describe issuance as O(M) overall
while ignoring those copies. Existing filter bounds and publication-copy costs
remain as recorded in the component report.

## Source hashes

```text
38506474dc76b0180d121623d81d5e7db9fa6f9cdb59c002f0ddc9ebe5d68e49 snapshot_gate.go
933f22ae476321f2c899d8c39462653dd120a1afd858a8a869851242e4bedc7d snapshot_observer.go
e94e9972dfd735fbcc8cea5afa9a708730b742ff6d93e392abfbf735974695c9 snapshot_journal.go
47ab33b9b0364a2252d1812a37b341e2f17ffe20af022986bb86a3b1d1bb6fd6 snapshot_gate_test.go
62b4c291a5fd718462b4b6fac765e1b82cf4151367ccc6d29f53a5ec9d3f64ac snapshot_journal_test.go
da23fb78fbe38d6c5a9942cfd34e725262b19f403fa3eb15d3ac7bdcf758b8f1 snapshot_journal_benchmark_test.go
```

## Next quality work

Add a four-role evolving-model control using the same12800 gate budget, without
altering archived controls. Confirm its6400 specialization against the original
policy before freezing fresh seeds. Then compare the snapshot candidate against
all existing controls plus this matched-budget control under every stationary
and recovery gate. Do not call the candidate successful from these unit tests,
from a favorable final block, or from the bounded runtime alone. All seven
research directions remain open.
