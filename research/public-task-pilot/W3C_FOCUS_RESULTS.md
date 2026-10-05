# W3C domain transfer: frozen clause deletion

Completed20 actual CaptureTurn/Recall runs on four new public W3C publication
records, using the previously frozen query rule without changes. Sources are
dated W3C document headers linked in [corpus](w3c-focus-v1/corpus.json).

Baseline positive top1:2/8. Focus:4/8. No baseline-correct loss; the finite
transfer improvement screen PASSES. Both-correct pairs:0/4 to1/4. This is a
small new-domain result, not population statistical evidence. The questions
were constructed after the rule was frozen, and include deliberate challenges.

Necessary-exclusion adequacy FAILS. Two questions with the same requested prefix
but excluding different dates become exactly the same effective query and
produce identical candidate orders, scores and laws. Their supports differ.
Clause deletion cannot retain this distinction; do not promote it as a general
replacement for an explicit target/anchor/exclusion representation.

Six queries change; unchanged-query candidate scores/laws are identical.
Maximum absolute forecast movement0.004970143488998269. All four records retained.
The absent-answer controls remain unanswered by the corpus and no abstention
mechanism is tested. No generation, feedback, production or runtime changes.

Raw artifact: [results](w3c-focus-v1/results.json).
`node research/public-task-pilot/check-w3c-focus.mjs` verifies hashes, frozen
function identity, cases, oracle ranks, source coverage, unchanged-query
invariance and the exclusion-pair collision. W3C fixtures are now consumed.

Next action: retain date constraints from the ORIGINAL question, using the focus
only for candidate retrieval; test explicit contradiction demotion separately.
