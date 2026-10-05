# Prior And Borrowing V40 Results

2026-10-03. Technical audits PASS. Every candidate FAILS the frozen overall
quality screen in BOTH normal cohorts. No production adoption. All seven
whole research goals remain OPEN.

## Scope And Frozen Decisions

Normal design and untouched confirmation each have448worlds,14784arms,
246400snapshots and1075200distinct outcome trials. Total896worlds,
29568arms,492800snapshots and2150400trials. The separate diagnostic adds
28worlds,924arms,15400snapshots and67200trials, with no adoption verdict.
Worlds are reused across11models and3delay schedules; those copies do not
multiply independent observations. Each of84cells has16paired worlds per
normal cohort. Two geometries and14regimes include10% feedback-noise cases.

Eight candidates vary raw/inverse nominal center, strength2/4 and private/
shared calibration family. Hazard remains1/16. Controls are full-history,
adaptive-window and legacy local16. No tuning or candidate selection from
diagnostic, design or confirmation outputs. See the
[prospective protocol](mmm-prior-v40-protocol.md).

| Candidate | Design failed cells /84 | Confirmation failed cells /84 |
|---|---:|---:|
|raw2_private|75|74|
|raw2_shared|72|74|
|raw4_private|77|76|
|raw4_shared|76|75|
|inverse2_private|76|74|
|inverse2_shared|72|74|
|inverse4_private|78|76|
|inverse4_shared|76|75|

All work/allocation checks pass. Quality, stationary protection and recovery
requirements reject adoption. Mean+/-3.5SE over paired worlds is the frozen
exploratory screen, NOT a simultaneous confidence sequence, target-law
certificate or Anti-Pigeon error-control guarantee.

## Partial Benefits And Remaining Harm

Confirmation wide/partial/immediate means:

|Model|Issued expected Brier|Final Brier|Recovery rounds|Top10 expected utility|
|---|---:|---:|---:|---:|
|Full|.229868|.215913|9|.800000|
|Adaptive|.207549|.174780|5.0625|.800000|
|Legacy local16|.233053|.196458|8.5|.800000|
|raw2_private|.217005|.185439|6.0625|.800000|
|raw2_shared|.212214|.184482|5.75|.800000|

For raw2_shared versus Full, issued-risk gain is.017653 with exploratory
interval[.016219,.019088]; recovery gain3.25rounds with[2.572228,3.927772].
That is about7.68%less issued Brier and36.11%faster penalized recovery.
Recovery includes phase-length+1 for a miss;9 here is that miss penalty,
not a measured successful detection time. It is not changepoint latency.

The SAME cell still FAILS: final Brier gain versus Adaptive is-.009702 with
[-.016365,-.003040], failing the frozen lower>=-.01 protection rule.
Identical.8top10utility does not rescue worse probability quality.

Stationary tight/aligned/immediate confirmation gives Full issued Brier
.196611 versus raw2_shared.220714, a harm of.024103. Full top10 expected
utility.874933 versus.820168 is also worse. Endpoint/partial-shift gains
do not justify replacing a stronger stationary predictor.

Across the84cells, matched shared-minus-private issued-risk contrasts have
positive means in77design/78confirmation cells and positive exploratory
lower bounds in71/72. Two design and three confirmation cells instead have
negative upper bounds. These correlated cell counts are DESCRIPTIVE, not
an omnibus significance test or universal benefit. Inverse-minus-raw has
positive lower bounds in only6cells and negative upper bounds in77in each
split. Strength2-minus4 has positive lower bounds in73cells in each split,
but negative upper bounds in9design/10confirmation cells.

Strength changes the actual finite-grid mean as well as higher moments.
For raw nominal center.925, conditional prior means are.840876(strength2)
and.886317(strength4); initial mixed marginals.738613/.770422. Thus this
contrast isolates a declared parameter, NOT variance alone or a proven
overconfidence mechanism. Shared/private comparisons DO start identically.
Legacy local16 is not the same prior as the new private three-family model.

## Correctness And Performance

Seven model race tests, explicit21^3path enumeration, independent whole-history
delayed reference, original-forecast lifecycle, atomic failure/caps/censoring,
interleaved prefix and integration future-prefix checks PASS. Thirteen
non-identity semantic corruptions plus missing-allocation rejection PASS.
Vet PASS. Independent normal audits verify448worlds/246400snapshots each.
Separate JavaScript readback verifies frozen sources/artifacts, disjoint world
seeds and all summaries/contrasts/decisions, with8altered-summary controls.
See [review and measurement caveats](mmm-prior-v40-review.md).

Apple M4,10logical CPUs. Three constructor150 repetitions:332378-340930ns,
5446162-5446171B and8allocations. Predict150:11531-11645ns,zero allocations.
Replay64with63already-known later observations:10901-10943ns,zero allocations;
checkpoint restoration is excluded. It is not the old one-label/unit-suffix
benchmark, so their numerical ratio is not an equal-work speed comparison.

|Maximum measured arm-loop elapsed /2400labels|Design ms|Confirmation ms|
|---|---:|---:|
|Full|96.402|93.532|
|Adaptive|369.596|313.789|
|Legacy local16|2.402|3.039|
|All eight new candidates, largest maximum|8.536|6.822|

All are under400ms and the constructor is under8MiB. The arm timer includes
model construction, schedule/setup, tickets, issue/arrival/drain, snapshots
and receipt-loop work, but excludes initial issued/expert-output-array
allocation. Scoring, serialization, acquisition, storage and serving are also
outside that arm boundary. No Goal6loaded100/250ms or Goal7equal-TOTAL-cost
claim follows. Epoch-reset peak memory and concurrent deployment are unmeasured.

Collector process times:386.05s design/385.40s confirmation, including scoring
and output. Independent audit times887.95s/889.10s are OFFLINE verification,
not learner or serving latency. The daemon dependency list contains neither
new prior package; no production code path is enabled.

## Preserved Evidence And Next Work

Interrupted initial session63347 is missing; fresh process inspection found no
live runner. Its completed model-race/log remained intact. Recovery verifies
the same source freeze and saved command hash, skips completed checks and
exclusive-creates remaining artifacts. See
[recovery record](../../research/prior-v40-diagnostic/RECOVERY.md).
No outcomes were resampled, files overwritten, or gates changed.

Primary artifacts: `research/prior-v40-diagnostic/` and
`research/prior-v40-normal/`: freezes, command metadata/logs, raw JSONL,
independent audit reports, completed manifests and readbacks.
Normal raw SHA256:

- Design: `8bbdfa5da048b16513f81668096ad983d77d99eee07335dee8b645d70df932ae`.
- Confirmation: `bc7e613b90b94e5d832f9e7ea6acf1eadbb35ad6a8e31f226e7abc2b21a8d80e`.

Next freeze NEW ablations of first-moment-controlled priors, richer mean
families and richer rate-shape kernels, varying these separately. Test
switching family/working-expert weights with explicit delayed-evidence
semantics, not a silently altered ordinary-Bayes claim. Preserve incumbent
quality and all issued-score/recovery gates. Publication/read ownership and
untouched-domain retrieval remain separate viable leads, not solved here.
Research sources and their limited role are recorded in the protocol/review.
