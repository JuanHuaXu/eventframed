# Conditional audit gate, frozen toy protocol

Date: 2026-10-01. This is a research-only identifiability and finite-sample
test for Goal 3, not an EventFrame implementation or a target-law certificate.
No existing gate, service, paper, or production configuration changes.

## World and evidence

Two observed context cells `X in {0,1}` occur independently with probability
1/2. A frozen predictor always reports class 1 (`p=0.75`). Its scalar monitor
sees only correctness `Y`; the conditional monitor sees `(X,Y)` on the *same*
externally selected audit labels. A reference sample has exactly 256 independent
labels in each cell. The reference law is `P(Y=1|X=0)=0.9` and
`P(Y=1|X=1)=0.1`. Stable live streams retain this law; shifted streams swap
the cell probabilities. Thus the contextual average total variation is 0.8,
but the scalar correctness law is Bernoulli(1/2) in BOTH regimes. The reference
sample is separate from every live stream. No monitor receives the hidden law.

Two fixed evidence schedules, each 512 live frames, 1,000 independent seeds
per regime: (a) every outcome audited immediately; (b) audit nomination
probability 1/4, independent missing probability 1/5, and integer delay
uniform on 0..31. Nomination, missingness and delay are generated independently
of context and outcome. At clock `t`, monitors can use only labels whose
arrival clock is <= `t`; unrevealed labels are never treated as negatives.
Both gates use exactly the same selected, arrived labels. Reference cost is
512 labels; live audit and usable-label counts must be reported separately.

## Frozen decision rule

Set horizon `T=512`, false-revocation budget `delta=0.02`, practical-equivalence
width `epsilon=0.10`, and seed base `2026100101`. For `K` cells let

`L_K = log(4 K (T+1) / delta)` and `r_K(n) = sqrt(L_K / (2n))` for `n>0`.

At each clock, the conditional gate flags if any cell has at least one live
label and its absolute live/reference sample-mean difference exceeds
`epsilon + r_2(256) + r_2(n_cell)`. The scalar comparator uses `K=1`, all 512
reference labels, and the same rule on pooled correctness. Stop each gate at
its first flag. No threshold, prior, schedule or seed is changed after data.

Under independent bounded labels and outcome-independent selection/arrival,
Hoeffding's two-sided bound plus a union over the `K` reference means and at
most `K*T` live prefix means gives familywise false-flag probability <=delta
under the cellwise null `|p_live(x)-p_ref(x)|<=epsilon`. The formula is deliberately
conservative; it is NOT imported from a published confidence-sequence theorem.
Its assumptions fail under informative missingness, adaptive nomination,
nonstationarity within a cell, dependent sources, or an untrusted reference.
The known synthetic target is used only by the evaluator.

## Predeclared screens

- Verify the exact scalar-law equality and contextual TV=0.8 algebraically.
- On each stable schedule, each gate flags at most 20/1,000 streams.
- On each shifted schedule, the conditional gate flags at least 900/1,000 by
  clock 511, while the scalar gate flags at most 20/1,000.
- Every observed label counted by either gate must have been nominated,
  nonmissing, and arrived by its decision clock. The two gates' observed counts
  must match exactly at every clock. Audit counts and detection-clock
  distributions must be retained even if the screens fail.
- Reject a negative control that exposes future scheduled labels at issue time;
  maintain deterministic replay and finite arithmetic checks.

Passing is only a component demonstration that context can restore information
lost by a scalar projection on this declared two-cell family. It cannot prove
the full Anti-Pigeon target-law diameter, arbitrary-context coverage, or
downstream Brier improvement. A failed screen is preserved without tuning.

Statistical background: Howard et al., [Time-uniform, nonparametric,
nonasymptotic confidence sequences](https://arxiv.org/abs/1810.08240).
The actual finite-horizon bound above follows directly from a union bound and
does not claim that paper's sharper boundary or assumptions.
