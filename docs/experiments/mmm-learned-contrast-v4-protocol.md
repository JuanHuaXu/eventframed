# Learned contrast v4: evidence-gated null forecast

This is a fresh Goal 1/7 component rescue after the consumed v3 factorial.
It tests a **forecast** gate, not a new Anti-Pigeon certificate. Do not tune
the gate on v3's observed outcomes or on either v4 split.

## Fixed models and streams

- Seven cases and the 512-clock truth/noise/missing/delay schedules are as in
  the v3 protocol: stable parity, immediate and delayed bit2, bit0,
  parity(0,1,2), Bernoulli(.5) null and out-of-family majority. Design and
  confirmation use fresh seed bases 2026102401/2026102402 and independent
  baseline fit indices 800..815/900..915. Each case/split has 16 fits times
  16 streams. All arms on one trajectory share the potential-outcome tape.
- Each arm reads 512 contexts and requests exactly 128 labels. Deliver only
  selected, nonmissing labels after the corresponding tick's forecast. The
  32-arrived-label rolling likelihood, declared 0.05/0.95 known rules and
  origin-feature delivery remain unchanged. These are working predictors,
  not ordinary posteriors over the full history.
- Arms: three v2 11-rule random/uncertainty/learned controls; one v3
  always-on-null learned arm with its original null fallback; one
  always-on-null arm with the v2 learned selector; and three v4 gated-null
  random/uncertainty/learned arms. Random, uncertainty and learned v4 arms
  must select exactly the same clocks as their matched v2 arms. The
  always-on-null/v2-selector arm must match v2 learned's clocks too.

## Frozen gate

The v4 prior is `.855` on the baseline, `.0095` on each of ten known
alternatives, and `.05` on the null rule. Conditional on the known family,
those prior ratios are exactly v2's `.90`/`.01` ratios. After each arrived
label, compute the 12-rule rolling-likelihood weights. Let `w_null` be the
null weight and `w_max` the maximum over the 11 known rules. At forecast
time, output `0.5` only when at least eight requested labels have arrived
and `w_null >= 100*w_max`; otherwise divide known weights by `1-w_null`
and forecast from that conditional known-rule mixture. Re-evaluate each
clock without hysteresis. The selector is the frozen v2 11-rule working state,
updated on the same arrived labels as the 12-rule forecast state. It always
uses the top known alternative and never uses the null gate or evaluator-only
truth. This separate state avoids floating-point tie changes caused by
rescaling the known-rule prior. Include both states and both updates in the
cost and memory screen. Record gate activations.
The Bayes-factor-like ratio is a *working evidence threshold*, not a
calibrated target-law or selection-bias certificate.

## Frozen screen

Report full, first64 and post256 realized Brier, post256 expected Brier
against the generator's true conditional probability, restricted mean
recovery delay and miss fraction under v3's trailing-32/16-consecutive
expected-Brier threshold, gate activations, fit-cluster intervals and
isolated component cost. Evaluator truth never enters selection or forecast.
The gated learned arm passes only if **both** fresh splits meet all of:

1. On null, every gated policy's full realized Brier is at most `.27` and
   improves its matched 11-rule arm by at least `.05`, with a positive
   16-fit-cluster lower gain endpoint. On stable parity, gated learned's
   full harm upper cluster endpoint versus 11-rule learned is below `.01`;
   gate activation after clock100 is below 5% of stable forecasts.
2. On immediate and delayed bit2, gated learned post Brier gain over both
   gated random and uncertainty is at least `.01` with positive fit-cluster
   lower endpoints; first64 mean harm to either is at most `.005`. Its
   recovery lead over **both** gated controls is at least 10 clocks, its
   miss fraction is no more than `.05` above either, and its recovery-delay
   harm upper fit-cluster endpoint versus 11-rule learned is below 5 clocks.
3. On bit0 and parity(0,1,2), gated learned post harm upper fit-cluster
   endpoint versus 11-rule learned is below `.01`. On out-of-family
   majority, gated learned post *expected* Brier is at most `.26` and
   improves 11-rule learned by at least `.01` with a positive fit-cluster
   lower endpoint. Report majority recovery misses even if score passes.
4. All selected-clock identities specified above hold; all arms request
   exactly128 labels; gated learned arrived-label mean gaps from its two
   controls are at most two per case. No invalid state, future label or
   replay mismatch. Isolated update p99 is below 10 us, forecast-plus-select
   p99 below 1 us, and state below 4 KiB.

A pass would support only a synthetic bounded-known-family component. It
would not prove full Goal 7, external-law truth, Anti-Pigeon authority,
agent-level value or loaded request latency. Retain all failures and do not
promote any arm on a partial pass.
