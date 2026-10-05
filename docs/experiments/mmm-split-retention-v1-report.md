# Detection-time reset loses useful training evidence

## Finding

The actual accepted split clock is too late to reproduce the clean-fit
diagnostic simply by discarding all earlier origins. Among already-split,
delayed reverse-shift trajectories at checkpoint480, this leaves about5 audit
labels on average, versus32-37 clean labels in hindsight. A strict reset is
therefore not justified as a direct implementation of the clean-data finding.
This is an evidence-retention diagnostic, not a prediction experiment or proof
that every possible reset policy fails. All seven research goals remain open.

| Case, delayed at480 | Cohort | Split active /16 | Labels retained among active | Oracle clean labels among active |
| --- | ---: | ---: | ---: | ---: |
| Majority to parity | 1 | 16 | 9.88 | 31.81 |
| Majority to parity | 2 | 15 | 11.27 | 35.93 |
| Parity to majority | 1 | 14 | 5.21 | 32.29 |
| Parity to majority | 2 | 9 | 4.78 | 36.56 |

At checkpoint384, active delayed reverse cases retain0 and0.5 labels on average.
At480,7/16 and5/16 reverse trajectories have an empty proposed fit;6 and4 of
those respectively have not had a new fit published after the accepted split.
Unsplit cases keep their existing mixed windows, so an all-trajectory mean
sample count can misleadingly look healthy while the reset cases lack support.
The artifact separates these groups and records stale labels retained in
unsplit trajectories.

## Policy Being Diagnosed

Use the same1024 checkpoints as the clean-fit experiment. If no split has yet
occurred, retain the latest published last64-origin window. After an accepted
split, reject pre-split snapshots and use only origins strictly after the split
from a subsequently published fit, up to64 labels. Prediction at the split
clock still precedes that split; a same-clock subsequent fit can be used only
at later checkpoints. The source driver performs feedback, gate update, then
fitting, matching this timing.

No model is refitted and no forecast changes. True boundary256 is used only to
label retained/discarded evidence after computing the proposed origin set.
Outcome labels are not consulted. No stale snapshot is counted as a valid
post-split fit. All512 stationary checkpoints preserve their existing state;
none has an active split in this tape. This is not a rare false-split guarantee.

## Verification

Five explicit controls cover no split, split after old publication, same-clock
pre-split prediction, strict origin filtering, and no snapshot. Every selected
origin is checked for audit status, availability, missingness, and split order.
Actual and oracle counts agree with the prior clean-fit artifact at every
checkpoint. Parent hash is pinned; raw clean-fit identity is checked and the
clean summary's hash is recorded. All1024 checkpoint records replay exactly
using the independently regenerated clean-fit summary.

Script SHA256: `abbe6e6ca4d69c207aa59fd0e4d13e99932b3b2a2bac9556fd4f47b98bf780a6`.
Result SHA256: `5deecd3ce67fafc48b768001044de232fac009e8c69c20f16fced67fcea42de6`.

Artifacts: [script](../../research/split-retention-diagnostic.mjs),
[result](mmm-split-retention-v1.json),
[replay](mmm-split-retention-v1-replay.json).

## Next Research Boundary

Earlier local-birth and fixed-observer-birth experiments changed weights and
failed; they did not test this training-origin policy. Do not repeat them or
infer that a bigger reset weight will compensate for missing evidence.

Investigate change-onset localization distinct from detection time, or a
continuously maintained candidate that already contains recent evidence when
the external gate authorizes a split. Both require checking earlier segmentation
and window-bank failures first. An onset estimate selected from the same data
can overfit; its uncertainty, delayed labels, false localizations, and forward
predictive validation must be explicit. A hypothetical oracle boundary is not
an authorized retroactive certificate, and the existing gate cannot be weakened
to manufacture earlier detections. No method is adopted here.

No production, whitepaper, remote, commit, or push changes.
