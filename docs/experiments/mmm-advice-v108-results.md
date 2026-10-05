# v108: neutral competition and origin-age ablation

## Verdict: all three candidates FAIL

This was implemented and run, not only proposed. All1,536 schedule-runs replay
exactly. No candidate satisfies its complete418-gate rule; none is promoted.
The best passing cells from different arms cannot be combined into a success.

| Candidate | Non-harm pass | Gain pass | Total pass | Verdict |
| --- | ---: | ---: | ---: | --- |
| Neutral-only | 384/384 | 14/34 | 398/418 | FAIL: little additional recovery |
| Age-only | 349/384 | 8/34 | 357/418 | FAIL: recovery insufficient; stable-case harm |
| Combined | 348/384 | 12/34 | 360/418 | FAIL: partial recovery; stable-case harm |

The [frozen protocol](mmm-advice-v108-protocol.md) preserves768 independent
latent trajectories, each under immediate and delayed/missing feedback,
12 cases, two phase-disjoint rule pools and32 trajectories per cell. Seven arms
share the same as-of training evidence. No model receives future outcomes.
All1,254 gates are retained in the [summary](mmm-advice-v108-summary.json).
Intervals are the predeclared paired mean +/-3.5 standard errors, not confidence
sequences, exact finite-sample coverage or history-wide multiple-testing control.

## Confirmation delayed late-half results

Expected binary Brier, lower is better. These are synthetic forecast scores,
not real-agent answer accuracy or a general MMM94.7% reproduction.

| Case | Generic | Conservative | Full carry | Arrival | Neutral-only | Age-only | Combined |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Parity4 | .072352 | .070292 | .060294 | .059513 | .059518 | .070132 | .070467 |
| Majority to parity | .261559 | .257326 | .259128 | .258020 | .258053 | .255014 | .252858 |
| Parity to majority | .222223 | .221759 | .226849 | .225533 | .225407 | .221184 | .220000 |

The combined candidate improves both confirmation switches over arrival:

- Majority to parity: gain .005162, interval[.003876,.006448].
- Parity to majority: gain .005533, interval[.001305,.009762].

Those two gates pass, but design majority-to-parity gain is only .003933,
below the .005 floor. Against the stronger conservative control, confirmation
gains are only .004468 and .001759; both fail the floor. This is a partial
recovery result, not a validated rescue.

Stationary delayed late parity4 illustrates the cost: combined Brier increases
by .010955 versus arrival, interval[.007560,.014349]. Expected accuracy decreases
from94.6484% to94.2529%. Neutral-only preserves94.6484%, but does not achieve
the required extra recovery. The previous policy remains unchanged.

## Mechanism interpretation

The v107 oracle showed room for neutral fallback, but giving neutrality a small
persistent Brier-selector prior does not make that room operationally available
quickly enough. Origin-age discounting helps switches but continually returns
weights toward the original strongly generic prior, including in stable streams.

There is a concrete selector limitation, separate from the observed scores.
For any raw challenger j and the generic64 role G, the age selector has:

```text
log(w_j / w_G) = -log(57) + .5 * sum_received lambda^(t-i) (loss_G,i-loss_j,i)
lambda = 2^(-1/32)
```

To outweigh G, a challenger needs weighted Brier advantage exceeding
2log(57)=8.086103. Before prediction, even infinitely many immediately available
past labels have total discounted mass at most lambda/(1-lambda)=45.668046.
That implies a weighted-average advantage greater than .177063 is necessary
in this most favorable mass limit. Delays and missingness reduce the mass.
This is an exact raw-advice constraint of the declared formula, not a proof
that every observed regression has one cause. The downstream evidence gate
can independently redistribute weights, so it is not a bound on the served law.

Do not silently lengthen the half-life or change the prior on consumed outcomes.
The next distinct lead is [log-score advice](../../research/log-score-advice-proposal.md):
preserve long-lived evidence but make confident errors more consequential,
while retaining Brier as the actual external quality criterion. Conditional
activation of a fast challenger remains another lead; neither is validated.

## Integrity and bug hunt

- Literal age-weighted batch reconstruction across289 clock values, missing labels,
  delayed arrivals and reversed within-clock order passes. Clock advancement
  discounts losses even without new labels; late arrivals retain origin age.
- Disabling both changes matches the frozen arrival policy within numerical
  tolerance, with identical acquired masks, gate tests and lifecycle counts
  across immediate, delay8 and delay31 schedules with missingness.
- All four original controls match frozen v106 exactly on consumed compatibility
  streams. Gate rejection, version isolation, duplicate/expired delivery,
  reentrant operations and failed-reader atomicity tests pass.
- One confirmed bug was found before fresh generation: all-rejected neutral
  composition could produce1.0000000000000002. Normalize the composed weights;
  a dedicated regression verifies exactly neutral output. No threshold changed.
- Combined race/compatibility/seed tests passed in6.293s; package vet passed.
- Full generation179.06s; deterministic replay181.62s. Independent JS verifies
  all393,216 step scores/costs, exact as-of fit lists and both update clocks.
  All29 source hashes match; evaluator output regenerates byte-identically.

## Performance

After generation and replay stopped, fixed-model256-frame delay8 lifetimes:

| Arm | Time range | Allocated bytes per lifetime |
| --- | ---: | ---: |
| Arrival | 1.610-1.620ms | about148,016 |
| Neutral-only | 1.629-1.638ms | about151,473 |
| Age-only | 1.560-1.565ms | about146,673 |
| Combined | 1.616-1.636ms | about151,473 |

These are policy-dependent acquisition lifetimes, not pure selector overhead,
per-request latency or a significant speedup claim. The seven-arm whole fixture,
including model fitting, took115.54-132.11ms immediate and112.96-114.01ms delayed.
No database, network, concurrent serving tail or production OpenClaw is measured.

The first component batch inadvertently overlapped independent verification.
It is retained and labeled contended, not used for the table. Both handles
were terminal before a complete isolated repeat; no individual sample was
selectively discarded. See [all raw timings](mmm-advice-v108-benchmarks.txt).

Artifact304,541,794 bytes, mode0600, SHA256:
`715b9776799c5931f7ee18e0124168f6537637bbedea3f453cbb2b55c3e691fc`.
All seven research directions remain open. No deployment, commit, push,
whitepaper promotion or new private-data collection occurred.
