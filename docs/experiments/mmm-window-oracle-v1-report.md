# Observation versus forecast-error diagnostic

## Scope

This retrospective simulator diagnostic does not test a deployable rescue.
It scores all 128 consumed trajectories, both schedules, four 128-frame
windows, and three arms: fixed acquisition, original coupled acquisition, and
the failed subset-proxy intervention. All seven research goals remain open.

The exact simulator rule, including its change at frame 256, is available only
to the offline evaluator. Inputs are independent uniform nine-bit vectors,
with independent symmetric 5% outcome noise. No learner, observer, model fit,
or split gate receives the true rule or future outcomes.

## Method

For each recorded observation mask and values, enumerate all compatible full
inputs to obtain the oracle probability q. For the issued forecast p, the
conditional expected Brier score decomposes as:

```text
q(1-p)^2 + (1-q)p^2
  = 0.0475 + [q(1-q) - 0.0475] + (p-q)^2.
```

The first term is irreducible noise. The second measures information absent
from the recorded view. The third measures forecast error relative to the
oracle on that same view. It combines model and mixture-weight errors; it does
not distinguish inadequate training data from inadequate model families or
selection. The oracle conditions on the view under this simulator's input law,
not on an arbitrary real-world distribution. The adaptive observers must not
inspect hidden bits to choose their next read, as required by their existing
reader contract.

Conditional expected scores integrate fresh hidden inputs and outcome noise.
They need not equal realized Brier on these finite trajectories; both are
retained in the artifact. This is not a new confidence certificate or a claim
that oracle performance is attainable from the available training evidence.

## Results

Late window, frames 384-511, immediate schedule. Each pair below is
fixed-acquisition -> original coupled-acquisition:

| Change | Cohort | Missing-information term | Same-view forecast error |
| --- | ---: | ---: | ---: |
| Majority to parity | 1 | .12775 -> .14307 | .04909 -> .04591 |
| Majority to parity | 2 | .10165 -> .13101 | .07257 -> .05761 |
| Parity to majority | 1 | .05463 -> .03429 | .08549 -> .09609 |
| Parity to majority | 2 | .04459 -> .03349 | .07619 -> .08786 |

Forward-shift observation becomes less informative in both cohorts, more than
offsetting the reduced same-view forecast error. In the reverse direction,
observation becomes more informative, but forecasts do not fully exploit it.
The coupled reverse same-view forecast-error term is .09609/.08786 immediate
and .11512/.11652 delayed: the remaining issue is not just missing variables.

The proxy is not uniformly better. In cohort-2 immediate forward, its missing
information rises further to .15543 even though its same-view forecast-error
term falls to .04002. Its conditional expected Brier is .24296 versus .23613
for the old coupled arm. Thus a coherent substitute guide does not by itself
ensure useful observations.

These decompositions describe different recorded views, not a causal proof of
which internal transition is faulty. A smaller same-view error can accompany
less informative observations because an uninformative view is easier to
predict near one half. Do not optimize that term alone.

## Verification

The enumerator agrees with a separate completion-count formula on every ternary
assignment for majority3 and parity4: 39,366 partial-view checks. Each also
checks the risk identity at five arbitrary predictions. Controls include parity
with one hidden relevant bit (q=.5), fully revealed parity (q=.95), and majority
decided by two positive or two negative observations (q=.95/.05).

All three input tapes are SHA256 pinned; trajectory identities, target-mask
cardinalities, actual observed values, and fixed-arm equality are checked.
The replay using the independent native replay tapes is byte-identical.
No production source or experimental threshold changes were made.

- Script SHA256: `a5a81bb4df001d48594cd6725e144ae2268f124727a1a75057920b57d0645433`
- Result SHA256: `03669cbdd439890cf7f766eda7d92b44ae9091edd7fedd4685532cbb768ab337`
- [Script](../../research/window-oracle-diagnostic.mjs)
- [Result](mmm-window-oracle-v1.json)
- [Replay](mmm-window-oracle-v1-replay.json)

## Next Test

Before another family replacement, measure the oracle convex-hull headroom of
the existing constituent forecasts on exactly the same recorded views. If the
oracle probability lies outside that hull, no reweighting alone can reach it.
If it lies inside, that establishes only possible headroom, not a learnable
selector. Keep this diagnostic separate from any deployable candidate and
retain both shift directions and stationary controls. This will separate
missing forecast capacity from weight assignment more directly than another
guide or prior sweep.
