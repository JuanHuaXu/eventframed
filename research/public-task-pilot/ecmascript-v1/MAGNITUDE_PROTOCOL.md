# Prospective Magnitude Fusion Pilot

2026-10-03. Research-only recommendation, not a confirmed runtime repair.
The Git RRF experiment lost five paraphrase top1 answers and gained none.
Lexical mismatch, ignored score gaps, and packet occupancy remain competing
causes. We do NOT infer a cause from those losses. Falsifier: unchanged or
worse untouched top1/survival, or any change to the properly scored law.
Existing runtime hooks suffice; no upstream/production patch is proposed.

[Bruch, Gai and Ingber](https://arxiv.org/html/2210.11934v2), sections1,4,5,
motivate retaining match magnitudes and fitting a convex-combination weight.
Their empirical advantage is not assumed here. Our hook lacks raw semantic
scores: s_i=(1-lambda)/(i+1)+lambda*x_i, where x_i is frozen ASCII six-field
union coverage. This is a ONE-channel magnitude variant, not their two-score
method or a calibrated confidence estimator. Operators/punctuation are lost
by that extractor; equality distinctions explicitly test this limitation.
O(n) scoring/O(n) memory, n<=200; service sorting/extraction/packing measured
separately. No IDs, labels, full-text metadata or oracle in the scorer.

54 public program records and108questions from the sealed preparation.
New service/in-memory store/query session per question/arm; Observe imports
the exact six declared fields, not reparsed prose. This tests retrieval from
preconstructed frames, NOT automatic text conversion quality. Content/source
metadata are retained but excluded from embeddings and score features.
Actual import/query availability timestamps are supplied; the WHEN string
does not claim runtime availability. Recall50,pack10,budget10000; ordinary
occupancy/correlation packing stays ON. No OpenClaw/LLM/production/LibraVDB
connection. Only the existing loopback embedding service is used, with the
previously frozen Nomic digest and exact role-separated512-entry memo.
Warm Recall excludes embedding acquisition, which is separately recorded.

FIT ONLY:36questions from six whole clusters; collect baseline and all five
weights0,.25,.5,.75,1 in rotating order. Scorer may open a fit-only label
file, never held-out labels, after raw predictions. Select the weight with
most actual packed top1 hits, then packet survival, then lower weight.
There is no guarantee this finite fitting avoids overfit; held-out tests
exist to falsify it. Freeze that one selected model before held-out runs.
DESIGN and CONFIRMATION:each36questions/six whole other clusters; baseline,
lambda0 incumbent-order control, selected fixed model. No tuning afterward.
The corpus contains every task's answer as genuine stored knowledge; query
target labels remain oracle-only. Retrieving known answers is not forecasting
unseen program semantics. Literal-program overlap/pretraining/related families
and researcher-authored questions limit interpretation.

Technical gates: complete unique cases and54exact imported frames; correct
50-candidate nomination and stored journal; baseline/incumbent packed parity;
identical frontier IDs/laws across arms; exact callback formula; no mutation
of retrieval scores or properly scored laws; source/model freezes unchanged.
Primary quality screen on BOTH held-out splits: >=2net top1 gains, zero
literal top1 losses, zero packet-target survival losses; positive one-sided
95% cluster-paired t lower bound (df5,t2.015048373). Report two-sided95%
interval(t2.570581836), six cluster vectors and exact discordant-cluster sign
test, plus each wording's nomination/top1/survival and reciprocal rank.
These pilot intervals are assumption-dependent, not simultaneous online
certificates. A selected lambda0/no gain is a negative result, not adoption.
Benchmark three times and report warm Recall/hook/capture/acquisition costs.
Neither unladen memory-store timing nor preparation closes whole Goal6.
Preserve all raw failures; exclusive-create artifacts; all seven goals OPEN.
