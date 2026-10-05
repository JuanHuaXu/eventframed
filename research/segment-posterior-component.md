# Bounded segment-posterior component

Status: implemented and numerically checked in the research package; fresh
predictive-quality and serving integration remain untested. All seven directions
remain open. This is not a replacement for Anti-Pigeon authority.

## Model and evidence boundary

The [source-derived lead](segment-posterior-lead.md) now has a Go component in
internal/observationlearners/segment_posterior.go. A snapshot accepts a consecutive
history with implicit unique origins left+i, left>=-16 and0<=clock<=256. History
ends strictly before the query clock. Label arrivals must be valid, at most287;
-1 marks unavailable labels. Select only arrived labels, then retain the latest
declared1..64 eligible origins. This is an explicit snapshot contract, not a
duplicate-message or provenance-authentication journal.

The model computes interval evidence under a declared mixture of the existing
generic conditional-cell model and Boolean agreement model. Both have the same
subset prior inclusion1/3 and Beta(.5,.5) rates as their original implementations.
Their marginal likelihoods are integrated sequence likelihoods, not fitted
training scores or variational bounds. Family weights and tail predictions come
from this same mixture. Empty observed intervals have likelihood1.

For constant hazard h in(0,1), enumerate interval boundaries by the log-space
prefix recurrence in the lead. The posterior distribution over the last segment
then determines the forecast, including the prior predictive with probability h
for a new boundary before the next observation. The prior predictive is.5.

Rebuild from the same as-of label set instead of repeatedly updating a cached
posterior with old labels. Missing and capped-out labels cannot enter fitting.
Excluded older labels remain outside the retained probabilistic model; this is
not exact inference conditioned on the entire historical stream. Selection that
depends on labels would require another likelihood model. Returned snapshots
own their forecasts, origin list and boundary weights.

## Checks

- Independent half-integer beta-integral references check all interval evidences
  of a six-label fixture and appended-label predictive ratios, including pure
  generic, pure Boolean and equal-mixture priors.
- Full512-output forecasts and marginal evidence agree with existing batch
  fitters on a16-label fixture.
- Every cut pattern and available-label mask of a four-frame history agrees
  with independent partition enumeration at two hazards and three queries.
- Delayed eligibility and the latest-label cap are explicit. Changing excluded
  labels leaves the entire snapshot identical; changing eligible evidence changes
  forecasts. Rebuilding is idempotent and prior snapshots remain unchanged.
- Empty and maximum272-frame histories are exercised. Under the declared
  constant hazard, removing an entirely unobserved prefix preserves predictions.
- Invalid horizons, lengths, inputs, arrivals, hazards, mixture masses and caps
  reject. Concurrent first use and independent fits pass the race detector.

Focused race checks pass in3.451s; a separate fresh-process concurrent-first-use
check passes in1.230s. Scoped go vet passes. These checks are not evidence of
better recovery, a calibrated population certificate or operational readiness.

## Cost and initialization audit

An eager512 KiB index-table initialization was identified before quality tests.
The service can import this package through researchmemory, so research-disabled
imports should not pay that construction cost. The table and subset-prior cache
now initialize lazily on nonempty use through sync.OnceValue. No production
configuration or running process was changed.

The implementation builds each distinct eligible-label interval once, rather
than refitting identical intervals across unobserved wall-clock gaps. For N<=64,
d=9,H<=272, cost is O(N^2*2^d + N*4^d + H^2 + H*2^d), after lazy indexing.
The first-use index construction costs O(d*4^d); its separate cold latency is
not measured. Working likelihood tables are300040 bytes, with a303104-byte
allocation on this runtime. Shared index/prior arrays have528384 bytes of data.
Returned model storage excludes those temporary interval tables.

[All39 benchmark rows](../docs/experiments/mmm-segment-component-benchmarks.json)
retain nine source hashes, command and machine details. Apple M4, Go1.27.1,
200ms benchmarks repeated three times, with no other research computation live:

| Boundary | Observed time |
| --- | ---: |
| Full16-label segment fit | 7.476-7.484ms |
| Full32-label segment fit | 20.438-20.524ms |
| Full64-label segment fit | 62.623-63.218ms |
|272-frame history, latest64 labels | 65.418-65.677ms |
| Generic64 batch control | 6.865-6.934ms |
| Boolean64 batch control | .435-.455ms |

Most fit cost is interval likelihood/prediction construction, not the boundary
recurrence. These are warm-component measurements, not worst-case or loaded
serving latency; no quality-matched performance advantage is established.

## Next work

Integrate this snapshot with the frozen immediate/delayed full-input experiment,
preserving all21 cases and all incumbent controls. Freeze hazard, family prior,
label caps and success criteria before fresh outcomes. Record last-segment
weights and eligible origins for independent replay. Keep the old fixed-window
and expert-routing failures visible. Faster or more elaborate inference is not
a rescue unless it improves the scored predictions under the full criteria.
