# Source-IDF Feature Experiment

2026-10-04. Freeze before execution; FIT-only exploratory screen. All seven
whole goals OPEN. The prior fresh calibration gain is sparse and failed its
promotion gate; preserve that failure and its consumed180-outcome ledger.

Hypothesis (needs investigation, not proven root cause): raw overlap treats
generic/source-structure terms and rare scientific terms equally, limiting the
source-only learner. Competing explanations include a weak hypothesis family,
bounded residual magnitude and pairwise-versus-top-ten objective mismatch.
Both previous learners improve after native-cue removal; that does not identify
this new feature bottleneck. This controlled test changes only four features.

Keep all5183actual source records, all351consumedFIT queries/221source families,
same full200frontier, five whole-family folds, labels, normalized BM25 base,
eight weights, .25tanh correction,200steps/.2rate/.001ridge/4cap. Keep the
unweighted source-only model as an immutable comparator, not a refitted control.
No CAL/confirmation prediction, target use or nomination tuning here.

For each term, df counts each source document ONCE in the union of title and
stored5W1H body. Fix w(t)=log(1+(N-df(t)+.5)/(df(t)+.5)), including df0 for absent
query terms. Features0/1 are weighted title/body matched unique-query-term mass
divided by ALL unique-query-term mass. Feature2 is ordered title bigram matched
mass divided by ALL ordered query-bigram mass, with each pair weight the average
of its two term weights. Feature3 uses the same weighted ratio for query terms
containing a Unicode decimal digit. Query repetition affects only bigrams.
Length features4/5 and zeroed native channels6/7 stay unchanged. Corpus/token
statistics use no outcomes. Reject any future/malformed source before df use;
bind features to the original verified epoch. Unknown source/epoch fails closed.

IDF/structured-field inspiration:
[Robertson & Zaragoza2009](https://www.staff.city.ac.uk/~sbrp622/papers/foundations_bm25_review.pdf).
Pairwise learning methodology:
[Joachims2002](https://www.cs.cornell.edu/people/tj/publications/joachims_02c.pdf).
The exact normalized feature ratios/union df/bigram combination are our declared
experiment, not canonical BM25F, a posterior probability or their theorem.

Primary screen: positive learned recall10 AND NDCG over BOTH common-frontier
BM25 and unweighted source-only comparator, under BOTH query and whole-family
means. Report every no-pair case, loss and unchanged frontier. No switching
aggregation or using only the weaker native baseline. This FIT mean screen is
not statistical adoption or untouched-agent validation. Any subsequent cohort
must be separately frozen and genuinely unused; consumedCAL cannot become new
confirmation and the failed prior promotion rule is not weakened.

Race original5learner/3mask plus3IDF roots x3; exact two-document union/absent-term,
numeric/bigram formulas, full-source epoch/future/cancel and concurrent native
invariance. Independent Node audit reconstructs ALL corpus dfs/70200vectors,
independently refits all6models and351held-fold full rankings, verifies original
control/prediction origins and five scientific corruption rejections. Count
source-index duplication/build, feature extraction, training, full ranking and
whole executableRSS/wall costs; not just an isolated dot-product benchmark.

Falsifier: weighted features do not improve both declared controls/aggregations,
or any other input/frontier/training/correction/epoch changes. The new isolated
module is not wired into production, continuous learning or native service code.
No live cache invalidation, loaded100ms/freshness or agent usefulness claim.
