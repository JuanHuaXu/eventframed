# V48 bounded global/local predictive mixture

Prospective NEW family; no production/private-data/paper change. V47 pure local
weighting was a negative diagnostic, not confirmation or a proven code defect.
The hypothesis here is that retaining pooled weight evidence while permitting
supported local specialization improves quality. Alternatives are inadequate
child experts, local evidence sparsity, or an ineffective extra aggregation
layer. The falsifier is failure of unchanged broad quality/cost requirements.

## Formulation and provenance

[Freund et al. (1997), Figure1 Bayes/SBayes](https://cseweb.ucsd.edu/~yfreund/papers/SpecializedExperts.pdf)
motivates convex expert aggregation and dormant task blocks. V48 has two
adaptive forecast strategies as outer experts. Reset-prior transitions,
delayed replay and this nested strategy are explicit adaptations; no inherited
Brier, delayed-feedback or Anti-Pigeon guarantee is asserted.
[Herbster et al. (2020), Section2.1](https://arxiv.org/pdf/2008.07055)
motivates observed task/global clocks, not a claim to its multitask algorithm.

One shared Full/Adaptive/rich-moment2 child bundle receives each revealed label.
Raw original expert advice q[t,h] is retained. Global head weights use the
global nomination clock and alpha_G=1/2400, prior(.8,.1,.1). Local head weights
use ONLY this observable member's clock and alpha_L=1/16, same prior. Both
heads retain original advice. Their current predictions are Q_G and Q_L.
The outer two-strategy mixer predicts Q=v_G Q_G+v_L Q_L, prior(.9,.1).
Three fixed outer hazards: static0, slow1/2400, round1/150. All layers use
T[g,h]=(1-alpha)1[g=h]+alpha*pi[h]; first position starts at pi.
Unknown/canceled emissions are one; delayed revealed emissions use the
ORIGINAL immutable advice at that layer's ORIGINAL nomination position.
Later head revisions never replace the outer strategy advice from issuance.
The served/returned/scored receipt contains Q, not either auxiliary head.

This is a prequential aggregation of adaptive strategies, NOT a posterior
under one generative model claiming independent child/head/scope observations.
A label is one evidence item despite arithmetic at several layers. Forecasts
are conditional on as-of visible outcomes; no true rates/hidden regimes/future
labels enter construction, routing or updating. Invalid external handles are
rejected before mutation. Unexpected partial-child/layer failure fences ALL
forecasts and preserves the global logical clock until explicit epoch rebuild.
Owners are serialized; a serial race suite is not concurrent-method safety.

## Bound and experiment

Global/pending cap2400;150observed members;16local positions, unchanged child
caps64. Both added mixers have maximum THREE storage slots: an AST-generated
fork of the audited original eight-slot filter. Only identifier renaming and
capacity8->3 are allowed; inverse syntax must match exactly. Test two/three
expert forecasts against the original on delayed/canceled/extreme-revival
streams. No label/prior/hazard change is a performance rescue.
Keep original<=8MiB constructor allocation and<=400ms complete collected-loop
requirements. Include added heads/scope and inspection in actual costs.
As with prior collectors, initial output buffers precede the loop timer;
fixture generation, reference auditing, scoring and serialization are separate.
No loaded serving/freshness/100ms guarantee follows from this component screen.

New seed bases diagnostic2026104807,n1/cell; design2026104809 and
confirmation2026104811,n16/cell. Check ACTUAL disjoint world seeds against all
V39/V40/V41/V43/V47 bases and within these splits. Retain both geometries,
all14regimes,all3delays:84cells/28diagnostic or448worlds per normal split.
No fitting/hazard/prior/policy selection from diagnostic or confirmation.
Same fixed policy must pass EVERY cell in BOTH normal cohorts; original
paired n16 mean+/-3.5SE screens, stationary/final protection>=-.01,
shifted issued Brier/priority versus Full mean>=.01 with lower>0,
Adaptive protection>=-.01, recovery lower>0 and mean>=10%Full remain.
Diagnostic is descriptive, has NOintervals/adoption. Preserve failed results.

Independent dense reference must reconstruct local/global/scope weights and
original head/served forecasts; it may not import candidate caches/shortcuts.
Bind raw advice to separately audited original control tapes, arrival sets,
owner/epoch/ordinals/served receipts, all snapshots, caps and costs. Test
future-prefix invariance, unobserved/canceled histories, incorrect original
head advice, forecasts, weights, receipts, controls, labels, and costs.
Freeze exact current test-source closure including V47's later pair helper.
All seven WHOLE goals remain OPEN; no narrower completion criterion is adopted.
