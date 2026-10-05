# Partial-view MMM learner v7

Frozen 2026-09-12 before evaluation. Follow v6 with fresh seeds and the original
six-coordinate scope/depth observation budget. Full-field audit training is
allowed ONLY after the live forecast; every arm receives identical available
independent Bernoulli(.25) audits, including all nine fields. This extra audit
acquisition cost is equal across arms, not included in the six-field live limit.

Four arms: fixed-count MMM; rolling-tree MMM; rolling-tree with frozen incumbent
guide; rolling-tree breadth. Fixed bundle=[incumbent,short64,long256,uniform];
other bundles=[incumbent,rolling64forest,long256,uniform]. Count/forest fits occur
at32 audits then every16, per v6. Use identical frozen mixing weights/share.

Fixed MMM uses the highest pre-update weight supported count model as guide.
Rolling MMM likewise chooses the highest weight among incumbent/tree/long; a
tree winner uses expected entropy gain per acquired coordinate. Frozen-guide
always uses incumbent MMM. Breadth uses the existing breadth policy. All experts
score the SAME observed mask within each arm; no expert reads hidden live fields.

For tree marginalization and tree expected gain, explicitly assume the synthetic
generator's independent fair binary input coordinates. Average left/right at an
unobserved split; do not substitute the actual missing value. This assumption
must not migrate to real data without validation. Count-model observation uses
its existing empirical conditional distribution. Both observer types begin at
scope0/depth0, stop at probability<=.1 or>=.9 after a read, and cap live cost6.
Forest observer requires complete chosen views; missing-view semantics are not
tested, and failure is explicit rather than inventing coordinates.

Reuse ten v5/v6 scenarios, 3 fits*8 streams*2 splits=480 streams of512 steps.
Fit base2026092801; design2026092802; confirmation2026092803. Same seed encoding
as v6 with roles0 generator,1 audit,2 missing,4 rolling forest. Frozen incumbent
fit is independent4096 samples of old law. No tuning between splits.

Success versus fixed MMM: rolling MMM must improve Brier>=.005 with paired
z=3.6 lower>0 on BOTH shift128 and shift256, preserve stable05/20 full harm
upper<=.01, and avoid >.01 mean harm in any case/window. Each fit-group primary
mean gain must be positive. Compare all three candidates to fixed with the
existing 120-comparison procedure. Report tree MMM vs frozen-guide/breadth too,
without claiming superiority or equivalence from raw means alone.

Journal chosen views, masks, probabilities BEFORE feedback, outcomes and label
deliveries. Verify hidden-value invariance, full-mask equivalence, exact uniform
marginalization, invalid-reader rejection, budget6 and replay of all artifacts.
Keep late shifts, delayed/missing labels and XOR failures visible. A pass is
synthetic partial-observation evidence, not actual agent-task success or AP
integration. Existing v3 remains the historical preserved-incumbent baseline.
