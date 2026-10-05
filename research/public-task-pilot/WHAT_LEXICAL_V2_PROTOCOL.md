# Lexical normalization rescue

The v1 combined screen failed by demoting HTML5.1 evidence below HTML5.2.
The tokenization discarded version numbers, a candidate upstream cause.
Freeze v2 before replay: preserve integer/decimal tokens, remove only full ISO
or English day-month-year date spans, retain the previous suffix/stopword rules,
IDF/cosine, numeric tie break and task-class precedence. No relation vocabulary,
case-ID special case, label-fitting or threshold tuning is added.

Same five consumed artifacts and the SAME no-regression plus landing>6/10 gate.
Report every output and preserve the failed v1 artifacts. Confirm that otherwise
identical documents differing in version are distinguishable, while changing an
ISO calendar date alone leaves lexical similarity unchanged. Date matching still
belongs to the task plan. This is normalization, not automatic entity resolution.
No integration or novel-relation generalization claim follows an offline pass.
