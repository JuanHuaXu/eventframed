# Provenance factorial v5: scoring is the stronger diagnostic lead

Completed1,792 new episodes and7,168 paired scored trajectories under the frozen
`FACTORIAL_PROTOCOL.md`. Raw data, calibration and source hashes are in
`factorial-v5.json.gz`; all summaries and contrasts are in
`factorial-v5-summary.json`. This was a diagnostic, not a production adoption
screen. It does not convert v4's failed rescue into a pass.

## Confirmation contrasts

F means fixed50/50 mode prior, I means calibrated signal-informed prior. First
letter is acquisition, second is scoring. Positive gain is lower Brier after
the change. Every arm costs24 units, including eight signal checks.

| Case | Acquisition-only final gain FF-IF | Scoring-only final gain FF-FI | Total final gain FF-II |
| --- | ---: | ---: | ---: |
| Independent20 | 0.003213 | 0.034976 | 0.037159 |
| Copied20 | 0.000147 | 0.014382 | 0.015186 |
| Mixed20 | -0.001626 | 0.032703 | 0.038365 |
| Matched05 | -0.000000153 | 0.015817 | 0.015817 |
| Matched20 | 0.000096 | 0.039480 | 0.040951 |
| Matched random signal20 | -0.000149 | -0.010197 | -0.013879 |
| Matched misleading signal20 | -0.001786 | -0.090652 | -0.102895 |

The independent20 scoring-only final gain has paired z=3.3 descriptive interval
[0.014446,0.055506]. Its curve-Brier gain is also positive under the declared
interval criterion in both splits. The acquisition-only final contrast is
inconclusive in every confirmation case. This is not proof that acquisition
does nothing, especially earlier in the trajectory or with a different budget.

For misleading signals, scoring alone increases final Brier by0.090652 with
interval[0.018893,0.162412]. Holding informed acquisition fixed gives the same
direction: informed scoring harms final Brier by0.101110. Misleading-signal
acquisition additionally harms curve Brier when scoring is informed, although its
final contrast remains inconclusive. Thus signal trust is not solely a search
ordering problem.

Copied20 mean harm from v4 did not recur here: total final gain is0.015186, but
its interval[-0.029358,0.059729] is inconclusive. This neither reproduces harm nor
establishes a rescue. The matched20 total gain likewise remains inconclusive
[-0.012351,0.094254]. Do not aggregate these chosen contrasts into a new success
criterion or interpret the descriptive intervals as simultaneous coverage.

## Design checks and limitations

Both acquisition policies visit five distinct tests per episode in this finite
family. They frequently reach the same16-slot set in different orders: among
128 confirmation episodes, same-set counts are110 independent20,118 copied20,
63 mixed20,127 matched05,93 matched20,91 random-signal and84 misleading-signal.
This post-run diagnostic comes from the stored pair traces. Shared final evidence
sets can limit final acquisition effects; ordering can still affect curve Brier.
The benchmark's finite budget and small test catalogue therefore limit transfer
to open-world acquisition.

All five tests in
`python3 -m unittest test_factorial_v5 test_informed_v4 -v` passed, including
complete replay of both experiments, hashes, normalized forecasts, equal paired
observations, unique source slots and cost accounting. FF/II match frozen v4
diagonal implementations at common seeds for the three retained populations.
Replacing calibrated priors by0.5 leaves FF unchanged and makes all diagonals
agree. The unchanged inference update retains unequal-prior exact-enumeration
coverage from v4.

## Next lead

Retain the fixed observer as a control. Test a declared mixture over reliable,
uninformative and misleading signal models, updating their weights from observed
source consistency rather than treating calibration as permanently valid. This
requires a joint predictive model and marginal-likelihood accounting; it must not
use the hidden target or true source mode for online trust updates. Evaluate on
fresh populations/seeds with independent-stream protection and explicit signal
failure stress. It is still only a candidate: consistency can be uninformative
when all available sources share one mistake, and inference cannot authenticate
external provenance by itself. Cost-sensitive stopping remains another lead.
