# Bounded snapshot specialists: next research lead

Subsequent [v115 comparison](../docs/experiments/mmm-snapshot-v115-results.md)
FAILS:754/1,018 gates, with739/960 non-harm and15/58 gains. Preserve the
implemented component and original proposal below; no quality promotion.

Status: [arithmetic and detached-law component implemented and tested](snapshot-specialists-component-results.md);
[gate/controller integration](snapshot-specialists-integration-results.md) also
passes scoped lifecycle tests; quality remains untested. Motivated by
[v114](../docs/experiments/mmm-decomposition-v114-results.md), not by a new
confirmation claim. V111 compatibility transfer and v113 rate uncertainty failed.

## Evidence and competing explanations

Late reverse-switch loss remains mostly at matched observations. Generic64
becomes better than generic32, but the served mixture remains worse. Early
after changes, all current models can be poor. The other direction also has
model-specific observation loss. These are different failure surfaces.

Possible explanations include delayed evidence; role-history carry across
changing snapshots; and model or acquisition misspecification. V112 addressed
event-time delayed filtering but did not meet the complete quality criteria.
V114 does not by itself prove role-history carry is the cause.

The present slot is a valid evolving expert: scoring its past forecasts is not
a mathematical bug. The proposed alternative changes the comparison family.
Its hypothesis is that frozen snapshot identities offer more useful recovery
behavior than advice attached to four continually changing roles.

## Research basis

Mourtada and Maillard's [Efficient tracking of a growing number of experts
(2017)](https://proceedings.mlr.press/v76/mourtada17a/mourtada17a.pdf), sections
5.1-5.3, develops Markov-weight updates for incoming experts, with prior mass
assigned when new experts enter. Equations21,22 and29 provide a starting
recursion. Their immediate-feedback growing-ensemble result does not establish
our bounded-retirement, delayed/censored, gated-policy Brier guarantees.

## Intended intervention

Give a fitted predictor the identity (publication version, model family,
training-window contract). Keep its forecast table immutable. Distinguish this
from keeping an old probability but evaluating it under a newer fitted model.

Initially test two retained publication banks, each with four models: a fixed
eight-predictor cap. Admit a new bank only at the existing32-frame publication
boundary. Retain the previous bank, retire the oldest before admitting the new
one, and explicitly account for its removed posterior mass. Incoming weight
and retirement redistribution must be frozen before quality generation, not
chosen from v114's favorable blocks. Exact duplicate-law coalescing is optional
and must not double-count evidence or reset tests.

Every feedback record binds origin, snapshot identities and immutable issued
probabilities. A retired slot cannot receive a late outcome intended for an old
identity. Retain enough bounded tombstone/message state to resolve pending
feedback, or define censoring at retirement explicitly and test its cost.
No historical outcome may be scored under a snapshot trained using that outcome.

The current four-role gate cannot silently certify eight snapshot predictors.
Define identity-scoped tests and a total error budget across admissions and
repeated decisions. Use matched gate accounting in controls to isolate the
snapshot intervention from a stricter or weaker gate. Recompute the coherent
forecast/observation mixture over the actually accepted snapshots; a rejected
predictor must not re-enter through a bank-level shortcut.

## Required component evidence before quality

1. Literal path enumeration and dense reference agreement through admission,
   delayed delivery and retirement; zero prior before admission.
2. Predict-before-update timing, no training leakage, duplicate handling,
   generation-safe identifiers and atomic rollback after failed publication.
3. Known immediate-feedback limits against the chosen source recursion;
   clearly identify any bounded modification rather than inheriting a theorem.
4. Detached coherent snapshots, bounded memory and measured publication,
   prediction, delayed-delivery and retirement cost. No unbounded history.
5. Frozen fresh quality comparison against all existing controls, retaining
   every stationary and recovery gate, not only final-block reverse switching.

A failed complete comparison rejects this candidate under that protocol. Do
not count a better oracle, a faster component, or a passing admission test as
meeting the roadmap. If the snapshot intervention fails, model recovery and
observation allocation remain distinct leads; independent generators, actual
prospective tasks and durable serving remain separate requirements.
