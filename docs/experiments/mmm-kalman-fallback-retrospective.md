# Kalman fallback probes on consumed v1 streams

This is an exploratory, **post-hoc diagnostic**, not an independent
confirmation. Both v1 design and v1 confirmation records were already
inspected before these policies were specified. No result here can validate
a new policy or count toward completion of Goal 2 or Goal 4.

The [replay script](../../research/kalman-fallback-retrospective.mjs) consumes
the exact journaled v1 forecasts and delivered outcomes. It validates the v1
source manifest and issues every new forecast before processing that tick's
delivered outcomes. It tests two cheap policies:

1. Existing four-expert `ForecastMix` update rule with frozen base, last-64,
   Kalman, and flat .5 experts; only arrived audited labels update weights.
2. A deliberately simple fallback: hedge the frozen base 80/20 with .5 until
   base-loss CUSUM `max(0, S + loss_base - .28)` reaches 1; afterward choose
   last-64 unless the last 8-16 available outcomes favor Kalman by a total
   Brier margin .08. This policy uses all arrived labels for its gate, like
   the v1 loss-window detector, while the underlying Kalman/last-64 models
   still train on audits only. Parameters were selected after seeing v1 and
   have no preregistered error or performance guarantee.

Mean confirmation Brier across 32 trajectories per case:

| Case | Mix early64 / tail128 | Fallback early64 / tail128 | Fallback alarms |
| --- | ---: | ---: | ---: |
| Stable .05 | .05957 / .06155 | .07504 / .07607 | 13/32 |
| Shift at 128 | .33424 / .06073 | .24939 / .06010 | 32/32 |
| Gradual | .09422 / .09739 | .11187 / .09075 | 32/32 |
| Delay 16, 25% missing | .37211 / .14411 | .29652 / .09715 | 32/32 |
| Nonlinear interaction | .34684 / .25063 | .27122 / .25170 | 32/32 |

The v1 confirmation last-64 control's early64 Brier was .25150 after the
abrupt shift and .25151 after delayed shift. The simple mixture is much worse
on both. The fallback nearly matches the abrupt control but remains worse
with delayed evidence. More importantly, its 13/32 stationary alarms and
full-stream stable Brier .07779 versus frozen base .06224 reject it as a
stationary-safe rescue. The design split shows the same qualitative failures:
12/32 stationary alarms, mix early64 .33340 after the abrupt shift, and
fallback early64 .28499 with delayed evidence.

These probes do not justify tuning the CUSUM threshold on the same records.
The next candidate needs a distinct frozen onset-aligned recovery gate,
explicit stationary false-alarm accounting, equal observed-label budgets,
and genuinely new fitting and trajectory seeds. No production code changed.

Method background only: [Page's original CUSUM paper](https://academic.oup.com/biomet/article-abstract/41/1-2/100/456627)
and [Herbster and Warmuth's best-expert tracking paper](https://mlanthology.org/mlj/1998/herbster1998mlj-tracking/).
This heuristic is not a faithful or guaranteed implementation of either.
