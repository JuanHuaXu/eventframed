# Log-score advice after the v108 failure

Subsequently implemented and tested in [v109](../docs/experiments/mmm-log-v109-results.md).
Both candidates FAIL overall, with substantial progress outside delayed
parity-to-majority. This proposal's original plan remains below for provenance.
Next is the [model handoff diagnosis](model-handoff-diagnosis-proposal.md), not
a retrospective adjustment of priors or acceptance criteria.

Original pre-implementation plan (retained for provenance). Preserve the complete
[v108 negative result](../docs/experiments/mmm-advice-v108-results.md), including
stationary harm from unconditional discounting and all failed gain gates.

## Why this is a different mechanism

The origin-age candidate bounds how much accumulated evidence can overcome
the strong generic prior, even when the environment stays stable. Neutral-only
avoids that forgetting problem but barely changes delayed recovery. Rather than
tune their half-life or prior on consumed data, test the loss used by the advice
selector while keeping the issued forecast and external Brier tests explicit.

An illustrative confidently wrong prediction p=.95 with outcome0 gives the
neutral .5 forecast an odds multiplier of exp(.5*(.95^2-.5^2))=about1.386 under
the existing Brier update. A likelihood update gives .5/.05=10 before sharing.
This is elementary score arithmetic, not a measured detection speed or proof
that log weighting improves Brier. A rare honest error can also trigger a much
stronger reaction, so stable-case protection remains essential.

## Proposed controlled ablation

For the same raw advice and optional neutral role, use stable log weights:

```text
log w'_j = log w_j + y log(p_j) + (1-y) log(1-p_j)
normalize w'
w_next = .999 w' + .001 pi
```

Preserve existing priors, role identities, publication cadence, arrival-time
delivery, captured issue-time forecasts and the unchanged evidence gate. No
unconditional age discount. Compare log/no-neutral and log/neutral separately
against Brier/arrival and Brier/neutral controls. The changed mechanism is the
proper selector score, not an arbitrary fitted learning-rate multiplier.

Use log1p(-p), require strict full support and validate finite values. Never
silently floor a zero probability. Keep coherent acquired/served mixtures and
duplicate/version/censoring invariants. The existing Brier regret statement
does not transfer to a log-score selector, and the evidence-gated served law
is not simply the ungated mixture analyzed by expert-advice theory.

First verify the immediate ungated recurrence against direct enumeration of
finite hidden expert paths with transition (1-rho)I + rho*pi. With zero share,
reordered likelihood updates commute up to roundoff. With share, arrival-order
updates are not event-clock Bayesian filtering of a delayed hidden Markov chain;
describe them as role-advice updates, not exact delayed posteriors. Do not use
that interpretation to claim nominal coverage under informative missingness.

Then freeze fresh quality seeds and all prior protection/recovery gates, with
Brier remaining the primary empirical decision criterion. Report log loss as
a secondary metric, not a substitute for failed Brier. Keep a full-support
noise/stationarity screen and the original two delayed switch directions.

## Primary research grounding and limits

[Mourtada and Maillard, Efficient Tracking of a Growing Number of Experts,
ALT2017](https://proceedings.mlr.press/v76/mourtada17a/mourtada17a.pdf), Section2,
gives logarithmic loss with eta1 and bounded square loss with its exp-concavity
constant. AppendixC derives MarkovHedge from expert paths; Corollary11 recovers
fixed share for uniform transitions. These motivate score and transition choices,
not our gated delayed/missing-feedback guarantee. The nonuniform-prior recurrence
above must be checked directly rather than attributed to the uniform corollary.

The stronger surprise response may accelerate recovery or merely make noisy
switching worse. Only the planned experiment can distinguish those outcomes.
