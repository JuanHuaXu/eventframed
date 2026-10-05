# Four-Shard Quality And Counterfactual Audit (DESIGN)

Freeze before collecting outcomes. All seven research goals remain unchanged.
The final V3 timing pilot is a prerequisite, not proof of retrieval quality.

## Sources And Boundary

Use both existing V3 isolated arms and the same repaired seven-file library.
Pin all inherited source hashes and copy the identical new test into both arms.
Use ONLY the 288 DESIGN capture requests and 96 DESIGN question strings from
the public corpus hash 20f85fc6d35ed2ec0fa5d68ec66b0cdcb48229dd85b6d1a7616731a9e7356de7.
Do not consume confirmation rows or any expected-answer/relevance labels.
This tests numerical nomination/packing quality, not real-task utility.
No production, installed cache, private corpus, learned state or paper changes.

## Experiment

1000 past CaptureTurn replicas, the same artificial as-of schedule as the load
pilot; 128D hash embeddings, SQ8, mmap and ten workers. Same backend V3 repair
in both arms; only candidate enables four-shard event collections. Record actual
nomination limit (3*RecallK), frontier 50/200, PackK10 and TokenBudget2000.
Run control/candidate then candidate/control in two fresh construction pairs.
Within each database use phases past, unchanged_repeat, future, future_repeat.
Between repeat and future, add the identical 256 future CaptureTurn replicas,
all after the fixed as-of for every query. Execute all 96 questions at both
frontiers in every phase; no concurrent writers. Separately time construction,
future captures and actual service recall. This is NOT a loaded latency trial.

Hydrate the 1000 committed past FP32 vectors from the database once. Independently
recompute each vector from its canonical EventFrame, rejecting hydration drift.
For every question compute FP64 cosine against ALL eligible past vectors.
Sort descending then ID for reproducible exact results. Ties within 1e-6 of the
exact cutoff count as equivalent (report strict ID overlap separately). For
each ordinary recall retain native nomination IDs/scores, actual full frontier,
packed IDs, every frontier forecast bundle, packet confidence and snapshot.
Emit no public conversation text. Then run the SAME service with a test-only
exact Search override returning top 3*RecallK past events with exact cosine;
all later scoring, law, journal and packing code remain ordinary. The exact
reference does not enter or train the ordinary arm, and its timing is excluded
from ordinary recall. No forecast/outcome labels are fabricated.

Both services are cold: no accepted posterior, residual, compatibility graph or
calibration rescue. Assert returned rank equals the declared baseline in this
fixture. Evaluate each packed ID with the canonical baseline using its exact
cosine, not a score guessed for a missing reference candidate. The complete law
comparison therefore covers the cold law ONLY, not learned-layer preservation.

## Frozen Finite Screens

All 4*96*2 ordinary recalls per command must succeed, full frontier size must
equal RecallK, every packed and frontier ID must be past/eligible, all returned
laws must be finite and normalized, and every common-ID law must be compared.
Declare separate screens, not a single vague pass:

1. ANN quality: mean tie-aware recall at the nominal 3K cutoff >=0.95 in EVERY
   arm/frontier/phase. Exact-score mean regret <=0.005 in every cell. Candidate
   minus control mean tie recall >=-0.02 in each paired cell. These are finite
   deterministic DESIGN screens, not confidence bounds or semantic accuracy.
2. Packet quality: for each query sum each packed event's EXACT-reference final
   rank score, divide by actual packed count; compare to the reference packet's
   mean. Mean nonnegative deficit <=0.01 in every cell; report count mismatches,
   exact overlap and score deviations separately. No claim of retrieval truth.
3. Counterfactual numerical equality (separate STRICT diagnostic): repeated
   ordinary packets/frontiers and complete common-ID forecast semantics must
   match to 1e-9; past versus future must also match, allowing opaque journal/
   runtime binding changes but NOT numerical scores, laws or packet confidence.
   Record repeat instability separately so ANN nondeterminism is not relabeled
   as causal future leakage. Future-ID exclusion alone cannot pass equality.
4. Exact reference negative control: exact packet/frontier/law semantics must
   remain stable in all four phases. If not, stop inference about ANN topology
   and investigate service/history dependence. These are as-of interventions
   on one test database, not a general future-data safety proof.

No population tail guarantee, stored RSS guarantee, untouched agent utility,
learning calibration, delayed-outcome Brier, all-arm load race, crash/migration
or million/billion scale claim. Preserve every failure and raw source pin.

## Research Motivation

Malkov and Yashunin, HNSW, https://arxiv.org/abs/1603.09320 (primary paper):
incremental random-layer graph construction motivates measuring quality and
counterfactual behavior rather than assuming a partition speedup is exact.
The tests here inspect this pinned implementation, not a theorem about HNSW.
