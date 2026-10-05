# Published-LSN Bayesian outcome transition v17: component pass

Date: 2026-10-02. The frozen [protocol](mmm-published-outcome-v17-protocol.md)
is satisfied for the single-owner, test-only component. This is not a
Goal 6 completion or production change.

The 4D gate test confirms that a new outcome moves the durable posterior,
residual and runtime versions and publishes the matching SQLite LSN; an exact
retry does not move either LSN or snapshot; and a changed payload with the
same idempotency key is rejected. A direct outcome write committed before
the gate transition leaves the marker unready and cannot be laundered by a
subsequent gated outcome. Injected interruption after the LibraVDB commit
reopens unready; interruption after the SQLite marker commit reopens ready.
The accepted outcome is read back from LibraVDB before the marker advances.

The 256D service test uses a real committed `Service.Recall` journal and
`Service.ObserveBayesianOutcome` with full-stream feedback. The updated
posterior is visible in a later scored Recall and again after service/store
close and reopen. A Recall whose as-of precedes the feedback does not use
that posterior. The test installs synthetic, test-only selection and
omitted-influence certificates to exercise the scored path; it does not
establish either empirical certificate.

An attempted one-LSN-per-transaction guard **failed**: the first outcome
advanced LibraVDB from LSN 21 to 28. One outcome transaction contains
multiple record writes. The final component requires monotonic LSN motion,
version/readback agreement and exact marker equality, but does **not**
identify an outcome-owned LSN range against a racing out-of-band writer.
All admitted fixture writes are serialized through one owner; enforceable
multi-writer fencing remains open.

Checks on the final fixture:

```sh
EVENTFRAME_RUN_PUBLISHED_OUTCOME_V17=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedOutcome(Gate|Service)V17$' -count=1 -v -timeout 3m
EVENTFRAME_RUN_PUBLISHED_OUTCOME_V17=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedOutcome(Gate|Service)V17$' -count=1 -timeout 5m
go test ./internal/store/libravdbstore ./internal/service -count=1 -timeout 5m
go vet ./internal/store/libravdbstore ./internal/service
```

All four checks pass on the final fixture. The focused race test exercises
the same sequential cases; it is not a concurrent mixed-load screen.
No loaded concurrent outcomes, offered-label publication age,
full-Recall p99 with learning, large-corpus scaling, external-writer safety,
or real-agent benefit was measured. All seven whole research goals remain
open; production is untouched.
