# Count-matched clean-fit diagnostic

Freeze before collection. All128 consumed trajectories, both schedules, four
checkpoints128/256/384/480. Use the latest published fit with clock STRICTLY
before the checkpoint. Actual=last64 origins; clean=last min(64, available
post-boundary count) origins; matched=uniform seeded subset of actual with
exactly clean's sample count. Before the change and in stable cases, boundary0.
After a changed case reaches256, diagnostic boundary256. All selected labels
must already be arrived, audited and nonmissing in the parent fit list.

Fit existing count and subset models, without changing priors or algorithms.
Evaluate all512 full inputs against simulator truth and5% noise at checkpoint.
This excludes acquisition and mixture selection. Empty clean/matched fits give
neutral.5 and must be reported, not dropped. If no fit has been published yet,
record FitClock=-1 and neutral predictions in all arms. Actual retains its larger sample
count. Compare clean versus matched to separate sample count from composition;
one seeded mixed draw per cell is not a universal guarantee.

This uses hindsight boundaries ONLY in an offline diagnostic. It is not a
deployable detector or evidence that a current gate can obtain these fits.
Retain every case/checkpoint, origins and predictions. Pin parent/source hashes,
replay native output, independently recompute risks and count predictions, and
test origin availability/order, equal sample size, and stable controls. No
production adoption or new success thresholds.
