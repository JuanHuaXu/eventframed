# New-subject transfer: literal success, paraphrase weakness

The NASA-trained model was frozen before these14 queries. Actual CaptureTurn /
Recall,19 public records, recall50/pack10, three arms. Six new facts are Nobel
award motivations from2022/2023; each has literal and paraphrased questions.
Two additional absent-record questions. No fitting, private data or LLM calls.

| Wording | Baseline top1 / survival | Lexical top1 / survival | Sparse top1 / survival |
| --- | --- | --- | --- |
| Literal,6 facts | 6/6 /6/6 | 6/6 /6/6 | 6/6 /6/6 |
| Paraphrase, same6 facts | 2/6 /5/6 | 2/6 /6/6 | 2/6 /6/6 |

The prospective non-harm screen PASSED, but this does not establish a useful
semantic rescue: all methods rank only2/6 paraphrases correctly. Paraphrase MRR
is0.4143 baseline,0.5694 lexical,0.5417 sparse. Lexical has three top ties and
sparse one. Sparse is not uniquely better than lexical.

Both experimental methods recover the quantum-dot support that baseline dropped
from packing. Unlike the earlier NASA pilot, this is a concrete lost-support
case, but just one case. No superiority interval is justified from six paired
facts, and wording variants are not independent observations.

Sparse maximum scores for the six paraphrases range0.0356-0.0522 despite there
being support for every question. Its absent-task maxima are0.0259 and0.0462,
overlapping supported queries. Thus a threshold chosen to abstain on absent
questions can also reject valid paraphrases. The old low-mean absent screen
cannot establish useful abstention across this domain shift.

## Sources and limits

Facts were checked against search-indexed primary-source excerpts. Direct site
and PDF requests returned403; no successful full-page verification is claimed.
Historical motivations only, paraphrased without private identities or contact
details. Primary sources:

- [Physics2023](https://www.nobelprize.org/prizes/physics/2023/press-release/)
- [Physics2022](https://www.nobelprize.org/prizes/physics/2022/press-release/)
- [Chemistry2023](https://www.nobelprize.org/prizes/chemistry/2023/press-release/)
- [Chemistry2022](https://www.nobelprize.org/prizes/chemistry/2022/press-release/)
- [Medicine2023](https://www.nobelprize.org/prizes/medicine/2023/press-release/)
- [Medicine2022](https://www.nobelprize.org/prizes/medicine/2022/press-release/)

No production embeddings were used; a hash embedder cannot establish what a
semantic embedding backend would recover. Model and fixture hashes, all outputs,
ties and score maxima are retained under nobel-v1. Audit tests passed. This set
is now consumed for any subsequent rescue tuning.

Next meaningful lead: semantic retrieval/representation with verified original
contracts, plus abstention testing on supported paraphrases and absent queries
together. More tuning of lexical coefficients would not establish semantic
transfer. Agent execution and real persistence/load remain outstanding.
