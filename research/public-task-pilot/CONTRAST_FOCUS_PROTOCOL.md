# Contrast-clause focus rescue

Frozen after temporal-contrast-v1, before rescue predictions. Both context sets
are consumed design data. This is not fresh confirmation or continuous learning.

Intervention: for a question containing exactly one literal `, rather than `,
use the preceding clause as the retrieval query, adding its terminal question
mark. Otherwise leave the query unchanged. This is a deliberately limited query
decomposition ablation. It neither parses temporal anchors nor enforces the
excluded clause as a logical constraint. Dropping a clause can discard useful
information; do not ship this heuristic on the strength of these fixtures.

Run baseline and focus through actual CaptureTurn/Recall, fresh memory service
per question, same eight public records and embedding model. No sparse correction
in either arm: isolate query intervention. Run all context-v1 and temporal-contrast-v1
questions, including absence controls. Oracle is read only after predictions.
Record original/effective queries, candidate order and forecast laws. Changing
the query may change forecasts; do NOT claim rank/law separation. No feedback
or calibration measurement is supplied by this harness.

Design screen: improve total positive top1 with no loss of a baseline-correct
case; report each set separately and all five contrast pairs. Unchanged queries
must retain scores/laws exactly. All candidates must remain present. Report
changed-law magnitudes as side effects, not as calibrated improvement.

No sweep or post-result editing. A pass is only support for testing an explicit
positive/negative query representation on a new domain; the original goal5 needs
untouched real tasks and agent outcomes. The earlier temporal diagnostic's
ten-of-ten criterion remains unchanged and must still be reported separately.

Source motivation is the temporal query decomposition and contrast sensitivity
discussed by [Shang et al. (2022)](https://aclanthology.org/2022.acl-long.552.pdf).
The clause-deletion heuristic is ours, not their trained model or a claimed
implementation of their method.
