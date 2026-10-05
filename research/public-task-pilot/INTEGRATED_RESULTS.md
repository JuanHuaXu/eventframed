# Actual pre-packing pilot integration

All16 consumed pilot queries ran through fresh CaptureTurn/Recall services for
baseline, lexical and sparse arms. Two additional full-pack baseline runs expose
all13 records per query for per-record law comparison, including promoted tails.
This is service integration with an in-memory backend and hash embedder, not
OpenClaw/LLM or libravdb validation. No fresh-generalization claim.

| Split | Baseline top1 | Lexical top1 | Sparse top1 | Positive support retained |
| --- | --- | --- | --- | --- |
| Design | 5/8 | 8/8 | 8/8 | 8/8 every arm |
| Consumed Voyager2 confirmation | 4/5 | 5/5 | 5/5 | 5/5 every arm |

Baseline has an exact top-score tie on the Saturn question with the Uranus
record (0.7317325878887497). Earlier run ranked Uranus first and scored3/5;
this run ranks Saturn first and scores4/5. Scores are identical. Do not label
tie ordering as a learned or baseline improvement. Both experimental arms have
unique top scores on these positive tasks. No production tie-breaking patch.

Every experimental callback saw all13 nominated candidates before packing10.
Sparse brought7 design and10 confirmation candidate occurrences into packs
that baseline omitted; lexical brought5 and6. These are occurrence counts across
queries, not unique records or additional correct answers. All original positive
supports already survived baseline packing, so this does not prove a rescue of
lost answers outside its top ten.

Full-pack reference verifies every served candidate, not just intersections:
backend scores and forecast bundles (excluding explicitly mutable RankScore)
are exactly equal across arms. Research deltas account for score changes and
experimental packet confidence is zero. Thus this is a ranking-only intervention,
not an improvement to the daemon's proper-scored forecast law.

Absent-task sparse mean/max scores: Mariner fourth flyby0.09949/0.22065;
Voyager2 Mercury0.09579/0.11565. A low mean hides a higher individual wrong
candidate. Neither calibrated abstention nor agent answer correctness follows.

Artifacts: integrated-*.json, integrated-summary.json and check_integrated.py.
Hash and replay tests passed. Future work must use genuinely new tasks, include
lost-support frontiers, run actual agent outcomes and measure persistent load.
