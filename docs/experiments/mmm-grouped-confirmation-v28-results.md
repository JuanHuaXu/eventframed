# Ready-group confirmation v28

Status: finite admission/age/read screens replicated. Not a whole-direction or
production-readiness result.

Reran the entire unchanged v27 twelve-arm experiment after full publication,
publication-store, service and research-memory race suites plus vet passed.
The v28 embedded source-hash dictionary exactly equals v27's: no cap, deadline,
validator or scheduling retuning occurred between the two experiments. Fresh
temporary persistent stores were created for every arm. This is a new execution
of the same synthetic workload, not independent real-data confirmation.

| Trial | Group4 accepted / 192 | Drop / expired | Read p99 off / group4 (ms) | Write p99 off / group4 (ms) | Group4 accepted-age p95 (ms) |
| --- | ---: | ---: | ---: | ---: | ---: |
| 0 | 192 | 0 / 0 | 34.151 / 24.700 | 19.770 / 24.859 | 122.732 |
| 1 | 192 | 0 / 0 | 54.568 / 25.698 | 30.360 / 18.820 | 84.150 |
| 2 | 192 | 0 / 0 | 32.143 / 21.134 | 18.529 / 20.765 | 118.162 |

Nearest-rank quantiles. Group4 admitted 576/576 and validated 28,800 candidate
records, with no errors, stale rejection, queue drops or deadline expiries.
Acquisitions were 50/49/49, all groups bounded by four. Partial groups were used.
Group1 admitted 175/177/175, with age p95 351.717/325.202/350.789ms. Prior batch
admitted 172/173/174, with age p95 382.611/351.417/377.944ms. Those controls still
miss the 250ms screen, supporting grouping rather than preparation order alone.

Across v27 and v28, group4 admitted 1148/1152 observations; the earlier single
deadline affecting four observations remains in the record. All six age/read
screens pass, with age p95 spanning 84-123ms. Finite failure frequencies and
latency ratios are not population guarantees. In particular v28's trial-1 off
arm has elevated tails; do not infer a universal serving-speed gain from that
ratio. Write tails are not uniformly improved and need their own workload tests.

## What this establishes

Bounded ready grouping can amortize admission ownership cost without eliminating
per-observation authority checks in this fixture. It is a justified candidate
for the next durable integration experiment, not a reason to relax validation.
The consumer remains uninstalled research code. No labels, posterior fitting,
SQLite admissions, durable terminals or actual agent answers are involved.

Next measure grouped admission with actual durable writes, preserving original
forecast/binding semantics and explicit unlabelled discard. Then connect verified
feedback/history authority before claiming loaded continuous learning. Evidence
from this cold adapter must not substitute for those remaining integration steps.
The other six research directions also retain their stated success criteria.

## Artifact verification

`mmm-grouped-confirmation-v28.jsonl`, SHA-256
`1a7e330fb93125a64ba4e255dc31495a53e4caf1d4e2700f94c97c7c5348b219`.
Independent parsing verified all twelve distinct cells, hashes (including exact
v27 source parity), request/write counts, weighted group/phase outcomes and
50 validations per accepted observation. No deployment or push was performed.
