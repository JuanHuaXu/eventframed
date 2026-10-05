# Returning-pattern V38 preflight

2026-10-03. Isolated research only; all seven whole goals OPEN. No outcome
cohort existed when this model/test/protocol family was written. V37 and
earlier source/data/gates remain unchanged.

## Patch Reasoning Gate

Confirmed symptom: V37 fails late/recurrent recovery; delayed recurring
issued Brier is worse. That is not a proven implementation defect. Credible
causes include sparse fresh member evidence, misspecified drift and stale
expert-loss credit. This candidate tests reuse of a recurring pattern, not
lowered quality thresholds or a promised general rescue. Partial, unrelated,
asynchronous and early shifts can falsify the structural assumption.

Local source is authoritative; this is a new standalone research model, not
an upstream daemon patch. Production, AP authority and search contracts are
untouched. States distinguish nomination, original emitted law, arrived
evidence, learned anchor, current filter and independently scored outcomes.

## Declared Model

First four issue ordinals of EVERY member supply the anchor. Until all those
genuine outcomes arrive, the finite shape model learns only those labels.
Later arrived labels are retained privately, not used to fit the template.
When the last anchor label arrives, freeze its member forecasts a_i.
A cancelled/missing anchor prevents freezing; no label is fabricated.

Conditional on this plug-in template, define one hidden state S_j in{0,1}
per subsequent nomination j. S_0=0; P(S_j!=S_(j-1))=1/600. The emission
is Bernoulli(a_i) in state0 and Bernoulli(1-a_i) in state1. The anchor
control uses transition probability0. No true member rate, change time,
regime label or future outcome is exposed to this model.

For arrived evidence at its original position, normalized forward filtering
gives the current state law. Unobserved/cancelled nominations have unit
emission, NOT negative evidence. Late labels replay the affected suffix;
state advances per nomination, not wall-clock/arrival tick. Stored original
forecasts remain unchanged. The next forecast is the two-state mixture.

This is exact filtering CONDITIONAL on a fitted plug-in template. It does
not integrate template uncertainty, assume a correctly identified causal
chain, prove stationarity/coverage or authenticate source independence.
Its emission/missingness model applies to the independently delayed labels
declared in this protocol; outcome-dependent delays are not repaired here.
The four-label anchor boundary and transition rate are model choices, not
values estimated from new outcomes. First-label scarcity and early drift
are explicit negative controls.

## Research Basis

[Rabiner (1989), sectionIII-A forward procedure and sectionV scaling](https://web.mit.edu/6.435/www/Rabiner89.pdf)
provide the finite-state filtering construction. Our emission templates,
single-owner delayed ledger and event-position suffix replay are research
adaptations. The paper does not establish this candidate's quality,
calibration, acquisition benefit or daemon latency. Enumerated hidden-path
tests independently verify the two-state implementation.

## Boundaries and Cost

N<=200,64 unique issue ordinals/member/epoch. Filter history <=N*60;
single owner. Prediction/nomination O(1), late resolve O(affected suffix),
worst-case O(N*60), with preallocated replay scratch. Anchor updates use
the finite378-state model; freezing is measured in full phase costs.
No database, network or I/O in the learner. There is no constant-time
claim for arbitrarily late labels. Epoch resets are externally declared,
not learned changepoints. Original ticket ownership, cancellation and
invalid-transition atomicity remain enforced.

Precollection path enumeration, independent two-entry matrix propagation,
anchor batch-integral checks, original-receipt/failure/cancellation/epoch
tests and future-prefix flips pass. Benchmark artifacts measure only this
isolated component; loaded serving and equal-total-cost observation remain
separate requirements. The prospective protocol and23-source manifest
bind the candidate before either cohort is collected.
