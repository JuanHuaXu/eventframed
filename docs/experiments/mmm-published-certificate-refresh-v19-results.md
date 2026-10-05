# Epoch-bound certificate refresh v19: engineering screen fails

Date: 2026-10-02. The frozen [protocol](mmm-published-certificate-refresh-v19-protocol.md)
**fails** in both normal trials. A test-only publication gate can issue
and read back fresh, epoch-bound selection and omitted-influence
certificates with fail-closed marker behavior. Under load, that makes
every Recall certificate-valid, but it does **not** make an earlier
posterior valid for the new epoch. It also adds enough contention for
outcome publication to miss the <100 ms p99 gate.

| Trial/arm | Certified Recalls | Later certified | Feedback-bearing packed candidates: certified posterior / epoch match | Learned candidates | Refreshes / p99 cost | Recall offer p99 | Outcome publish p99 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1, no refresh | 6/128 | 0/122 | 373 / **0** | 0 | 0 | 48.28 ms | 87.95 ms |
| 1, refresh | 128/128 | 122/122 | 336 / **0** | **0** | 11 / 9.29 ms | 94.78 ms | **117.09 ms** |
| 2, no refresh | 7/128 | 0/121 | 368 / **0** | 0 | 0 | 47.79 ms | 83.00 ms |
| 2, refresh | 128/128 | 123/123 | 360 / **0** | **0** | 11 / 9.57 ms | 86.88 ms | **107.74 ms** |

Each arm completed 128/128 writes, 128/128 full Recalls and 16/16
durable outcomes with zero stale, future, top-150-oracle,
journal/packet or final-marker violations. Refreshed write p99 was
135/138 ms (<250 ms); Recall call p99 was 58/52 ms (<100 ms).
Published-view maximum age was 27/26 ms (<250 ms). These timings
are finite fixture measurements, not population guarantees. A
separate focused refresh replay varied enough to put Recall offer
p99 above 100 ms as well; the frozen outcome gate already fails in
both paired trials regardless.

The model boundary is explicit: `Service.Recall` requires both active
certificates **and** `posterior.EvidenceEpoch == snapshot.EvidenceEpoch`.
Event ingestion advances the global evidence epoch; refreshing only
certificates cannot transport a posterior. The 336/360 matched
candidate observations show this is not a packing failure or a
missing-posterior failure: all referenced posteriors were marked
certified, but none matched the current epoch. Blindly retagging
them would discard the very dependency check that prevents stale
evidence from affecting the scored law.

The refresh gate's new, stale-epoch, direct-bypass and interruption/
reopen tests pass. A combined focused and loaded correctness-only
`-race` replay passes without a reported race. Under race
instrumentation outcomes can complete after all Recall offers, so
its timings and absence of scored post-feedback Recalls do not
evaluate the normal latency or learning gates.

Reproduce with:

```sh
EVENTFRAME_RUN_CERTIFICATE_REFRESH_V19=1 go test ./internal/store/libravdbstore -run '^TestResearchCertificateRefreshGateV19$' -count=1 -v -timeout 3m
EVENTFRAME_RUN_CERTIFICATE_REFRESH_LOAD_V19=1 go test ./internal/store/libravdbstore -run '^TestResearchCertificateRefreshLoadV19$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_CERTIFICATE_REFRESH_V19=1 EVENTFRAME_RUN_CERTIFICATE_REFRESH_LOAD_V19=1 EVENTFRAME_RACE_CORRECTNESS_ONLY=1 go test -race ./internal/store/libravdbstore -run '^TestResearchCertificateRefresh(Gate|Load)V19$' -count=1 -v -timeout 5m
```

The second command exits nonzero for the frozen learning and outcome
latency gates. The synthetic certificate values are assumed inputs;
this study supplies no external-law coverage or valid renewal
procedure. A credible successor needs an auditable, dependency-aware
posterior validity rule and a costed certificate source. No
production code changed; all seven whole goals remain open.
