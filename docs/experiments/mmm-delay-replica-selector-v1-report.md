# Bounded replica selector: consumed-data failure

## Result

The new bounded replica selector does not rescue recovery. It improves on a
single arrival-updated selector under delay, but both are substantially worse
than the recorded original after regime changes. Immediate-feedback failures
show that fixing delay ownership alone cannot fix this base selector.

Second original cohort, now consumed; post-change Brier (lower is better):

| Case | Schedule | Original | Arrival selector | Replica selector |
| --- | --- | --- | --- | --- |
| Copied bit | Immediate | .194749 | .323497 | .323497 |
| Copied XOR2 | Immediate | .212278 | .342988 | .342988 |
| Noisy-copy bit | Immediate | .201917 | .338270 | .338270 |
| Reversing XOR2 | Immediate | .180795 | .286604 | .286604 |
| Copied bit | Delayed | .279590 | .409339 | .361767 |
| Copied XOR2 | Delayed | .279574 | .406484 | .360888 |
| Noisy-copy bit | Delayed | .280556 | .413658 | .366847 |
| Reversing XOR2 | Delayed | .273903 | .361900 | .326672 |

Both original cohorts, stationary/null controls, full and post scores remain in
the [artifact](mmm-delay-replica-selector-v1.json). Delayed stationary and null
second-cohort replica scores also worsen versus original (.048421 vs .047330,
and .256469 vs .255241). No fresh significance or pass threshold is claimed.

## Implemented diagnostic

`research/delay-replica-selector.mjs` implements a33-slot pool, one pending
origin per slot, independent four-expert weights, copied issued advice, exact
origin-bound one-use feedback, and32-tick expiry. Prior is the existing
[.7,.1,.1,.1]; Brier exponential update rate.5 follows the earlier concavity
derivation, not an outcome-tuned grid. Expiry frees a slot without treating the
missing label as observed. A single arrival-updated copy is a matched comparison.

All candidate forecasts use the original's recorded advice and observation
masks. Model fits, observation choices and gate transitions are NOT rerun under
the candidate. This is fixed-advice reweighting, not a full closed-loop effect,
a posterior calibration claim, or replacement for Anti-Pigeon's authority.

Checks include detached advice, duplicate rejection, stale token rejection
after slot reuse,33-slot cap, all-frame outcome accounting, immediate single-copy
parity, and the observed-round cumulative excess bound
`M*log(1/.7)/.5`. Passing that bound does not protect every recovery segment;
it compares a static expert over observed rounds, not the original controller's
changing mixture or unknown missing-round outcomes.

The parent SHA is frozen to
`1533a049c2146cc7839799e9a394691d96c3fdf4a1a8e8d1c504db86053bf8e6`.
384 schedule-runs replayed; repeat invocation output is byte-identical. Runtime
was about.79s for parsing and the entire diagnostic, not a Go serving benchmark.
No future label enters updates before its recorded arrival. Labels on missing
frames are used only by the evaluator to score issued predictions.

## Decision

Do not integrate this selector as a rescue or tune a new learning-rate grid on
these results. The mechanism can satisfy a cumulative static-comparator bound
while recovering poorly. Any next reduction should first preserve a strong
immediate-feedback tracking controller and its gate behavior, then test delay
handling separately. Earlier snapshot/hazard/stale-credit failures still stand.
No fresh data were consumed, original gates unchanged, all seven goals open.
Production, whitepaper and remotes untouched.
