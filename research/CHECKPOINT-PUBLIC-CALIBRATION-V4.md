# Public Calibration Research Checkpoint

The actual manifest is
`research/checkpoint-2026-10-04-public-calibration-v4/manifest.json`.
It chains the previous pairrank checkpoint, preserves final sources and command
records, and hashes all explicitly linked large public artifacts/installed models.
This is reproducible in the recorded local dependency environment, not a
self-contained portable export. No private transcripts or production endpoints
are part of the new study.

## Verify Before Continuing

1. Verify manifest hashes for saved copies, explicit large-file links and the
   previous checkpoint. Do not compare historical saved research-direction text
   to a later edited live copy; the checkpoint's own copy is authoritative.
2. Confirm three independent study audits still match their study manifests;
   native-calibration readback matches audit, trace, means, counts and gate.
3. Keep source-only FIT and calibration data consumption separate. Calibration
   180 is now consumed; confirmation 300 remains unused. Never relabel repeated
   calibration runs as fresh confirmation or choose a winner using those labels.
4. All 14 core commands and all owned predictor/daemon children are terminal.
   A future rerun needs a NEW owned root/config/script with recorded explicit
   path transformations; existing roots intentionally refuse overwrite.
5. Native launch must copy ONLY the four explicitly frozen closed-store files
   into a new research root. Do not copy sockets, locks, PID files, HOME, TMP or
   credentials. Keep original failed stores and traces immutable.

## Repeat in a New Owned Study

The native-calibration runner is
`research/public-task-pilot/scifact-native-calibration-v4-run.mjs`; its frozen
copy is retained. It generates a calibration-only query projection, freezes the
two ALL-FIT models and complete local test source closure, then records actual
race (x3), vet, build and owned native invocation. The owned native launcher
freezes installed model/binary assets, runs a direct priority-10 PID under the
existing sampled 2GiB ceiling and joins both children before manifesting output.
Do not probe serve-help/default configuration or discover other endpoints.

Evaluate only after complete predictions and termination, using the independently
frozen `scifact-native-calibration-v4-audit.mjs` logic. It reconstructs full raw
source/epoch, BM25/RRF, payloads, features and rankings before outcome access.
Later repeated experiments use these outcomes as consumed diagnostic data only.
Changed models/features require a new protocol and an appropriately untouched
task cohort; this failed promotion gate cannot be relaxed retroactively.

## Findings and Next Work

Source-only ablation improves FIT recall by 0.18993 points. Fresh calibration
primary recall improves 77.77778% -> 78.33333%, but only one source family moves;
its paired interval reaches zero. NDCG improves with a positive descriptive
interval. NO PROMOTION; all seven whole goals remain OPEN and ACTIVE.

Observed serial stage p99 is 173.816ms, mostly native search (170.302ms);
non-native stages average 2.984ms. Full setup, trace and memory costs are recorded.
This is not loaded serving or live snapshot-invalidation evidence.

Next leads are more discriminative FIT-only source features plus genuinely new
outcome families, and a source-bound lexical-first serving path tested with live
version rejection, background freshness and loaded request latency. The full
goal still includes broad delayed/noisy recovery, stationary protection, valid
useful Anti-Pigeon splits, bounded shifted challengers, feedback authority and
equal-total-cost falsification observation. Production remains untouched.
