# Learned-delta interaction and bounded-margin rescue

The earlier empty-cache integration result did not imply that calendar order
survives cached corrections. A deterministic counterexample now exercises the
actual service applyRankDeltas and packing.Select implementations.

At frontiers2,50 and200, a valid calendar winner is initially first. Injected
reliable correction records subtract0.25 from that winner and add0.25 to
contradicted candidates. These values respect the configured cap. In ALL three
cases, the original ordinal adapter loses the packed winner. Proper forecast
laws remain unchanged. This is a genuine composition failure, not an error in
the individual date comparison.

The correction records are injected test fixtures, not claimed to have been
learned from actual feedback here. The result proves an allowed failure mode;
it does not estimate its production frequency or invalidate unrelated MMM claims.

## Research Rescue

New optional MarginRanker reserves score bands against a DECLARED later additive
bound b. It requires0<=b<=0.49 and complete-frontier return K1=K2. Let
a=(1-2b)/2 and o_i=(n-i)/(n+1) for zero-based ordered position i. For a mixed
partition assign:

```text
non-contradicted: score_i = 1-a + a*o_i
contradicted:     score_i = a*o_i
```

Every ordinal lies strictly between0 and1. Consequently every cross-partition
gap is strictly greater than1-2a=2b. Opposing additive perturbations bounded
by b cannot reverse that order, including monotone clipping to[0,1]. The allowed
bound range also keeps the band width away from machine-zero at the tested
frontier cap200. Unknown evidence belongs to the non-contradicted partition;
this is NOT a declaration that it is correct.

If a partition is empty, existing behavior is preserved. Partial-frontier output
is rejected because the service can append omitted candidates at their old
scores, which would defeat a band guarantee. The configured bound is encoded
in the adapter's contract name. Caller configuration MUST honor that bound;
the interface does not expose the service cap for independent verification.

## Checks

Race-enabled tests pass:

-Original ordinal counterexample reproduced at2/50/200.
-New margin adapter preserves the winner through actual applyRankDeltas and
 packing.Select at2/50/200 under the same injected0.25 corrections.
-Gap grid: sizes2/3/50/200, small/middle/large non-contradicted groups, and
 bounds0/0.1/0.25/0.49. All worst opposing cross-partition gaps remain positive.
-Invalid/nonfinite bounds and partial-frontier requests rejected.

```sh
go test -race ./internal/researchcalendar ./internal/service -run 'TestMargin|TestCalendarOrdinalDeltaCounterexample|TestCalendarMarginSurvivesBoundedDeltas' -v
```

This is a conditional ranking-resilience result, not calibrated certainty or
new factual evidence. The large artificial retrieval margin can affect elastic
modulation and packet-certainty interpretation; those effects are unvalidated.
Later arbitrary rerankers, resolution preferences, diversity logic or larger
corrections are outside the theorem. The new adapter reparses dates and has NOT
inherited the original adapter's217us benchmark. Full-service public replay,
new-variant performance and durable reason records remain required. Production
and frozen earlier source/artifacts remain unchanged.
