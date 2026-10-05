# Frozen focus rule: new-domain test

Freeze before predictions. Use the existing focusQuery function byte-for-byte,
without retuning. New public W3C historical publication facts, four records,
eight positive questions and two absence controls. This is a small domain
transfer test, not population validation or a test of agent answer generation.

Facts are paraphrases of the title/status headers of the dated W3C documents,
opened before dataset creation. They describe status at publication, not today's
status of HTML. Each source is recorded in corpus.json. Ingestion is at a common
research timestamp; historical publication dates are content, not backdated
availability. No people, private records, or invented facts.

Baseline/focus use actual CaptureTurn/Recall, local nomic embeddings, fresh
memory service each time, recall50/pack20. All four candidates must be retained.
Read answer oracle only after all predictions. Source hashes go into the artifact.

Eight positive questions: two proposal/final-status pairs, two exclusion-essential
questions sharing their positive prefix, and two temporal ordering questions.
The exclusion-essential pair has different correct records but collapses to the
same effective query under clause deletion. This is a structural falsifier of
unrestricted deletion: deterministic retrieval cannot answer both correctly from
the transformed query alone. It is not evidence against a representation that
retains and interprets the exclusion. Do not use this known limitation to alter
the frozen rule before testing.

Report positive top1, every baseline-correct regression, pair results and absence
outputs. Transfer screen: strictly better positive top1 and no baseline-correct
regression. Essential-exclusion adequacy additionally requires both members
correct; report separately even if aggregate accuracy rises. No post-hoc sweep.
Compare unchanged-query candidates/laws exactly and report changed-law magnitude.
No calibration, latency, general learning, or completion claim from this test.

Primary sources: dated HTML5 Recommendation (2014-10-28), HTML5.1 Proposed
Recommendation (2016-09-15), HTML5.1 Recommendation (2016-11-01), and HTML5.2
Recommendation (2017-12-14), linked in the corpus. Method is our existing
contrast-clause ablation; it is not a W3C or TSQA algorithm.
