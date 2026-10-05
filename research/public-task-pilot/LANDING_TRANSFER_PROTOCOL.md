# Frozen task-role transfer

The task-role-v1 implementation is frozen before this dataset is executed.
Four new fixture facts are verified against primary NASA sources in corpus.json.
Use UTC calendar date for Curiosity's landing (2012-08-06), not the previous
evening's US Pacific date. Two missions introduce entity competition; landing
and touch-down language extend the relation vocabulary. These are designed
questions, not natural-user incidence or learned unseen-knowledge acquisition.

Run24 actual service calls, ordinary focus versus task-role-v1, fixed pack1,
recall50, diversity on, budget10000, local embedding model and fresh memory per
question/arm. No algorithm/threshold/vocabulary edits after inspecting outcomes.
Oracle is separate and read after predictions. Preserve full source hashes.

Transfer screen: improve on ordinary top1 support across ten positive questions
without losing any correct baseline case. Report two absent controls separately;
do not exclude their continued packing from the limitations. Full-frontier laws,
numeric inputs and journal matches must remain equal. Any regression fails the
no-harm condition even if aggregate accuracy improves. No aggregate tuning on
this test is allowed; future rescues must be new variants with new confirmation.

This is a new fixture family after the previous design freeze, but still a tiny
researcher-designed transfer test, not population validation or full Goal5.
