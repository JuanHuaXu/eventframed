# Top-Ten-Aware Rescue: Smaller Regression, Still Not Adoptable

2026-10-04. Authoritative evidence:
`research/public-task-pilot/scifact-lambdarank-v1/audit-results.json` and the
15-source local test-dependency freeze/five terminal code-zero commands.
Seven actual race roots x3, all70200features, six independently refit models,
351full-ranking reconstructions and five corruption rejections passed.

Scientific inputs and FIT labels are byte-identical to pairrank-v1. The only
training change is dynamic, per-query-normalized binary NDCG10swap weighting;
all source features, epochs, family folds,200frontiers and residual/optimizer
caps remain fixed. No calibration/confirmation predictions, no hyperparameter
search, no new private/agent data or production changes.

| Query-Macro Metric | BM25 Control | All-Pair Residual | Top-Ten Rescue |
| --- | ---: | ---: | ---: |
| Recall10 |79.81956%|79.08357%|79.51092%|
| MRR10 |0.631353|0.627632|0.630939|
| NDCG10 |0.667185|0.662051|0.666090|

The rescue reduces the previous harm but still loses0.309percentage points
query-macro recall10 and0.001095NDCG versus BM25. It FAILED the joint improvement
screen. Do not reinterpret reduced harm as a successful learned challenger.
Family-average recall10 does rise79.78381% ->79.94218% (+0.158points), while
family MRR0.650197 ->0.645662 and NDCG0.681731 ->0.678653 worsen. Four families
improve recall10, fiveworsen,212tie; NDCG28improve/32worsen. Preserve this mixed
aggregation result rather than selecting the one favorable summary.

The full frontier recall remains92.92498% for both common-frontier arms. Thus
loss alignment alone is insufficient to establish a useful learned correction
over this strong baseline. The next credible controlled lead is native-rank
feature ablation/source-only or more discriminative IDF-weighted source features,
not bigger corrections without outcome evidence. Any design uses FIT only until
a clearly frozen prospective comparison; do not spend confirmation on a model
already failing the stronger control.

Actual source feature-index construction98.09ms; feature extraction
median0.363ms/p990.718ms/max1.154ms. Two rank arms togethermedian33.75us/p9949.04us.
Full-fit training846.92ms, fivefold fits636.54-700.94ms: more training than the
all-pair variant. Whole executable4.81s, maximumRSS189661184bytes (180.88MiB).
Repeated sort/derivative work is counted; no constant-time training or live
serving/freshness guarantee. Original3.12sfailed executable cost remains evidence.

Generation records describe the command before gofmt; final frozen formatted
source is authoritative. Only an additional Fit import/call plus formatting
changes the command. The independent evaluator uses separately written dynamic
swap-weight refitting, not Go model output as its own reference.
[Burges2010](https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/MSR-TR-2010-82.pdf)
motivates metric-sensitive gradients; our normalization/bounded residual is a
declared variant, not a convergence guarantee. All seven WHOLE goals OPEN/ACTIVE.
