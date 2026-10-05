# SciFact citation and license provenance

2026-10-04. Source preparation only; no embedding, fitting or ranking experiment.
The [upstream license](https://github.com/allenai/scifact/blob/master/LICENSE.md)
declares CC BY4.0 for claims/annotations, ODC-By1.0 for S2ORC abstracts, and
Apache2.0 for code. Dataset terms remain separate from Eventframed's software
license. Raw public source is local research input, not newly published data.
This is recorded source metadata, not blanket redistribution/legal clearance.

The [original schema](https://github.com/allenai/scifact/blob/master/doc/data.md)
distinguishes cited documents from documents with annotated evidence. All1,109
prepared BEIR claims match original claim text. Every BEIR relevance relation
matches a unique original cited-document ID. The BEIR test300claims are the
upstream LABELED DEV set, not the upstream unlabeled300test claims. Untouched
confirmation here means unused by our prediction/fitting/selection process,
not newly annotated, upstream-unseen, or unseen in pretrained models.

| Prepared partition | Claims | Positive citation relations | With annotated evidence | Without annotated evidence | Claims with empty evidence map |
| --- | ---: | ---: | ---: | ---: | ---: |
| BEIR train | 809 | 919 | 564 | 355 | 304 |
| BEIR test = upstream dev | 300 | 339 | 209 | 130 | 112 |

Thus positive retrieval labels are CITATION RELEVANCE, not necessarily evidence
that supports or contradicts a claim. A retrieved citation without annotated
evidence cannot be credited as truth verification. Support and contradiction
remain distinct evaluator labels; neither source citation nor a relevance score
establishes causal identification. The original connected-component split uses
these citation relations, so its no-shared-labeled-document property is intact;
its earlier description as evidence-document independence needs this qualification.
Do not remove inconvenient queries or relabel the unchanged benchmark.

Preserved V2 FAIL: the first provenance validator assumed relevance equals
evidence-map keys, and rejected a genuine positive citation with no evidence.
V3 checks unique citation sets; one source dev claim repeats a citation ID.
Exact/order/duplicate-source-citation/empty-evidence controls and10corruption
rejections pass. Separate relation-first reconstruction independently validates
the actual full archive and prepared artifacts, including byte hashes.
Archive SHA256 `11c621288d41ac144d29b13b0f8503b3820b7d6e8b1f6ff24dff335c196d76be`.
See `research/public-task-pilot/scifact-provenance-v2/FAILURE.md` and the V3
`report.json`/`audit.json`. Source member extraction goes to stdout only; no
archive-controlled path can write into the filesystem. Original prepared data
and split plan unchanged. Next freeze the source-preserving frame/retrieval
contract and metrics before any prediction run. All seven WHOLE goals OPEN.
