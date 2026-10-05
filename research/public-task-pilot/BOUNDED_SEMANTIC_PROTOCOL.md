# Bounded semantic-preserving correction

Exploratory replay on already-consumed Nobel/NASA corpus and14 tasks. Frozen
sparse weights unchanged. Actual CaptureTurn/Recall, local nomic embedding
model (same recorded digest), in-memory backend, no generation or production.
Arms: semantic baseline, sparse direct order, and bounded additive correction.
Recall50/pack20 retains all19 corpus records for full law/rank comparison.

Bounded callback output = clamp(baseline + .01*(sparse_probability-.5),0,1).
Existing research hook translates this into rank delta, leaving forecasts
untouched. Each delta is at most.005 in absolute value; any pair with baseline
rank gap>.01 cannot reverse. This is a rank stability bound, not calibration,
semantic validity, AP certification or proof that smaller margins are unsafe.
No parameter sweep, tuning or new labels. Do not claim a fresh confirmation.

Require all12 positive supports retained, bounded top1 no worse than baseline
on literals and paraphrases separately, at least one additional positive top1,
all19 forecast laws unchanged and every delta bounded. If only non-harm passes,
call it preservation, not an intelligence gain. Two absent cases have no oracle
support; do not invent an abstention threshold or count them as solved.

All runs sequential; retrieval timings include query embedding, not ingestion,
and are not a randomized overhead benchmark. Read oracle after predictions.
Retain raw scores and hashes. Local embedding calls are public-text only.
