# Research stop checkpoint: 2026-10-02

## Stop Boundary

The live Codex usage tool reports 81% used in the 10,080-minute weekly
window. The user's stop condition is usage exceeding 80%, so research stops
here. All seven whole research goals remain OPEN. This is neither goal
completion nor a user-requested pause. No production change, commit, push,
installation, or new experiment was performed during this stop turn.

The preceding turn answered the Kalman question by inspecting existing
studies and primary literature; it did not produce a new experiment or
implementation. Treat it as an assessment, not additional validation.

## Workspace

Authoritative checkout: `<LOCAL_ROOT>`.
Git HEAD: `1a7edb62b6be4031fd01ebeab8071b17303a7815`.

Existing tracked modifications were preserved:

- `.learnings/ERRORS.md`
- `internal/model/api.go`
- `internal/productioneval/codex.go`
- `internal/productioneval/codex_test.go`
- `internal/service/service.go`
- `internal/store/libravdbstore/store.go`

Research sources and artifacts also include untracked files. HEAD alone is
not a snapshot of this research state: retain the working directory and
its local artifacts. This document records a local resumption checkpoint,
not a committed or externally backed-up snapshot. No experiment process
is being awaited in this stop turn.

## Latest Evidence: Partition V29

The current summary records 576 worlds and 11,520 independently checked
arms per split. Both design and confirmation overall screens FAIL.
The source of these statements is
`docs/experiments/mmm-partition-v29-summary.json`; the governing frozen
criteria remain in `docs/experiments/mmm-partition-v29-protocol.md`.

Verified current SHA256 values:

| Artifact | SHA256 |
| --- | --- |
| `mmm-partition-v29-design.jsonl` | `4be1a0fe4dd85c9299d18f561cf5c5675529af991f78ef0292e43e27cdd99869` |
| `mmm-partition-v29-confirmation.jsonl` | `12759648758d2b63c7254dbaf8edc645c39a10e039db265173e3c1effb7c93d3` |
| `research/partition-v29-audit.mjs` | `d80cc71f4e4d6358fad6f558b5faba02505102f9fa17874d68a1ae6143342cb8` |
| `research/partition-v29-verify.mjs` | `96ae77ae89f8fc255677d536dc439be22df66d943e9733b3b5947909f65708c5` |

The raw tapes are under `docs/experiments/`. The original frozen checker
does not accept the baseline's serialized null trace. The separate
post-collection audit checker normalizes that representation to an empty
trace; it records its own hash and leaves the frozen checker, data, and
decision gates unchanged. It also reconstructs the ordinary posterior
and forecasts, including integrated leaf likelihoods.

To rerun the independent audit after resumption, from the checkout:

```sh
node research/partition-v29-audit.mjs
```

That command rewrites the derived summary, not the raw tapes. Inspect the
checker and compare its hash before execution. The opt-in collection
entry points and frozen seeds are declared in the protocol; do not
overwrite or reuse consumed cohorts as fresh confirmation.

## Unfinished Handoff

`mmm-partition-v29-results.md` is not yet written. The top continuation
entry in `research-direction.md` still describes the historical preflight,
so its "no outcome cohort yet" sentence is stale relative to the tapes
and summary. On resumption, first reconcile these documents with the
recorded negative results, without changing the protocol or thresholds.

The previous research handoff identified deterministic bit-reversal
aliasing: the first 32 nominations are even-indexed members only. This
should be reconstructed from the raw traces before further interpretation.
The next untested lead is randomized within-stratum observation combined
with coherent Bayesian averaging of baseline, affine, and partition
models. It is a proposal, not implemented evidence or an adopted rescue.
Any successor needs a newly frozen protocol, untouched outcomes,
selection-probability accounting, stationary protection, and total-cost
comparisons against random and uncertainty controls.

Retain all original seven success criteria. A nonlinear component win,
cheap Kalman update, or passed audit is not whole-goal completion.
