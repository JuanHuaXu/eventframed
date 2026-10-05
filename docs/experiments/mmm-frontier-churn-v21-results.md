# V19 frontier churn v21: full-frontier guard has little reach

Date: 2026-10-02. This is a consumed-fixture diagnostic under the
[declared protocol](mmm-frontier-churn-v21-protocol.md), not an untouched
quality confirmation or a serving change. It reconstructs the exact
cosine order of the v18/v19 256-dimensional write stream, excluding
16 future-only rows at request as-of. The focused test passes normally
and under `-race`; package `go vet` passes.

| Exact cosine set | Visible inserts changing set | First non-changing insert index | Final k-th angle |
| --- | ---: | ---: | ---: |
| Top 10 | 6 / 128 | 6 | .01278 |
| Top 50 | 34 / 128 | 34 | .07242 |
| Top 150 | 104 / 128 | 104 | .22152 |

Indices are zero-based. The test reconstructs the source fixture's two
angle-zero seeds, 198 eligible seeds at `.005*(i+1)`, 16 future-only
seeds and 128 visible appends at `.00213*(i+1)`. Each generated vector's
measured cosine agrees with `cos(angle)` within 1e-5. The existing v18
full-Recall test separately checked every actual search's top-150
membership against this angle oracle; this diagnostic does not rerun
that loaded service.

The v20 query-scoped posterior guard requires the exact nominated
frontier to stay fixed. An initial-record frontier becomes stale at the
first visible insert and stays stale throughout this stream. Even if
the record were rebased after each change, 104 of 128 next inserts
would immediately change top-150 again, although the **cosine** top-10
set stabilizes after six. This is a significant reach limitation, not
evidence that those posteriors are safe to reuse.
Actual packet packing applies additional scoring and constraints; its
top-10 membership need not equal cosine top-10. More importantly, the
selection-conditioned likelihood and activation probability can depend
on the full nomination frontier even when an event stays in the packet.
No acceptance rule may simply replace top-150 identity with top-10
identity without a new selection/provenance proof and a scored test.

The v18/v19 loaded fixture supplies 16 selected feedback outcomes, all
`Useful=true`. It has no balanced negative labels or independent
outcome-generating law, so it cannot estimate proper-score benefit,
false-positive cost, or calibration under a relaxed posterior gate.
The next Goal 6 learning experiment needs both a durable as-of
selection witness and mixed, predeclared outcomes on the changing
frontier, while retaining full Recall and loaded latency measurements.
No posterior gate or production code was changed. All seven whole
goals remain open.

Reproduce with:

```sh
go test ./internal/store/libravdbstore -run '^TestResearchFrontierChurnV21$' -count=1 -v
go test -race ./internal/store/libravdbstore -run '^TestResearchFrontierChurnV21$' -count=1
go vet ./internal/store/libravdbstore
```
