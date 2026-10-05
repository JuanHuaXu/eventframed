# Delayed expert-state switching V43: research contract

2026-10-03. New isolated package; not production, and not a quality result.
No V41 frozen candidate/source or original success criterion is changed.

## Research basis and precise model

[Korotin, V'yugin and Burnaev (2019), Algorithm4 and section4.2](https://arxiv.org/pdf/1902.10433)
gives a delayed fixed-share recursion which replays from the earliest newly
revealed original position. Unknown losses do not emit evidence. The paper's
bounded-loss regret comparison is not an absolute Brier, retrieval, Anti-Pigeon
or serving guarantee. This prototype uses Bernoulli likelihood, not an arbitrary
power posterior, with fixed rather than adaptively selected share probability.

For original nomination j, retain all expert advice q[j,h] from visible history.
Z[1]~pi; P(Z[j]=h|Z[j-1]=g)=(1-alpha)1[h=g]+alpha*pi[h]. This reset includes
self transitions; it is NOT the off-diagonal alpha/(N-1) version in
[Herbster and Warmuth (1998)](https://mwarmuth.bitbucket.io/pubs/J39.pdf).
Y[j]|Z[j]=h,visible history ~ Bernoulli(q[j,h]). The SAME likelihood defines
posterior updates and the scored forecast sum_h predictiveWeight[h]*q[j,h].
First issue uses pi without an extra transition. An arrival reveals its original
Y[j], not a new trial; replay retains every original q and issued mixture law.

All advice must be predictable under the actual arrival filtration. Delayed
missingness/cancellation must be ignorable; this core cannot certify either
condition from caller input. Future expert forecasts may change after arrivals;
past ones may NOT be recomputed from that newer history. Cancellation has unit
likelihood and still consumes a state transition. No guessed evidence.

[Koolen and van Erven (2010), section1.2](https://arxiv.org/pdf/1008.4532)
distinguishes black-box experts trained on all visible data from local/reset
expert semantics. The proposed wrapper uses ALL-visible-data experts; switching
mixture weights does not reset their training or inherit local-learning results.

## Bounded implementation and falsifiers

2..8 experts,1..4096 nominations per explicit epoch, bounded pending capacity,
strictly positive declared simplex prior, share probability in[0,1], expert
probabilities in[1e-6,1-1e-6]. Invalid inputs reject rather than silently floor.
Latent state messages remain in log space. Output weights may round below
Float64 range, but that rounding is never fed back into the filter: an expert
may recover after extreme contrary evidence, even with zero share probability.
Advice and issued mixture forecasts are privately retained. Tickets bind owner,
epoch and original slot. Epoch reset is explicit and invalidates old tickets;
no autonomous detector/abstraction/source authority is granted.

The unit-emission span rescue permits issue/predict O(N+log T) and resolution
O(T+N*K), K revealed anchors at/after arrival, with bounded O(T) index movement;
worst-case O(N*T) remains. Ledger/scratch/index O(N*T). Single-owner code, no
unbounded history or truncation of
unresolved prefixes. The global4096 cap is NOT continuous-learning completion.
Generating and updating every expert costs EXTRA. A cheap mixer cannot conceal
Full/Adaptive/model construction, scheduling, replay, observation or storage.

Required preflight: independent path enumeration, independent full-history
matrix filtering, immediate/delayed/cancelled/duplicate/foreign/stale-epoch
checks, as-of prefix invariance, immutable caller advice, atomic numerical
failure, probability/cap checks, and explicit zero/full-share controls.
No performance benchmark while V41 timed collection is live. Offline audit
concurrency is permitted for technical tests, NOT calibrated latency claims.

The test-only pool now uses existing Full, Adaptive and rich-moment2 learners;
each observes ALL the same revealed labels. All three original child forecasts
are retained by the mixer before any label is consumed. This prequential
likelihood does NOT multiply three supposedly independent copies of one label.
An unexpected cross-child failure fences the whole pool until explicit epoch
replacement, because child APIs do not expose a shared transactional prepare.
No partially updated bundle may emit. The failure's logical time also becomes
the reset lower bound: a new epoch cannot move before a child's partial commit.
Invalid input rejects before mutation.

Next quality study must freeze all variants BEFORE NEW cohorts. The pool
uses existing Full, Adaptive and the declared rich moment learner; each observes
the same revealed labels. Static and two declared fixed-share controls isolate
switching. This proposal is not selected by V41 normal results. The NEW
prospective outcome study is specified in `mmm-switch-v43-study-protocol.md`;
no quality result is assumed. All original stationarity/recovery/Adaptive
protection400ms/8MiB criteria stay; all seven WHOLE goals remain OPEN.
