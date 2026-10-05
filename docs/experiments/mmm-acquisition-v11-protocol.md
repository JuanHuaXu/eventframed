# Distribution-aware acquisition v11

Frozen before evaluation. V10 did not justify advancing integration alone. This
separate lead changes acquisition and integration together in the online policy.
Retain the original fixed, replacement, adaptive-retained and static-retained
arms in each of three modes: uniform original, empirical joint, oracle joint.
Only tree acquisition and tree partial forecasts change across modes. Count
models, mixture priors/updates, observation budget and audit cadence do not.

Joint table: v10 conditional table, estimated from last<=256 admitted audit
inputs plus one total uniform pseudo-observation. Oracle is explicitly diagnostic.
At each tree-selected view, expected posterior entropy is weighted by the joint
conditional probability of its possible observations, not uniform completions.
Choose maximum entropy reduction per new coordinate, same tie order, initial
mandatory view, confidence stopping and six-coordinate cap as before. Build
tables after each tree fit; current and future evidence cannot affect that frame.

Run fair, biased and clustered input generators with stable05, shift128,
recurring and interaction label scenarios. Two fits, four streams per fit,
two splits, three modes=576 streams of512. Fit2026105201,
design2026105202, confirmation2026105203; original seed encoding. Each split
reuses its paired fit across modes; splits reuse fitted incumbents. No tuning.

For each empirical retained arm, require all v9 finite gates versus fixed:
mean full/post harm<=.01 in every group, shift128 post gain>=.005 in every
generator/split. Additionally require no >.01 mean full/post harm versus the
same retained uniform arm and clustered shift128 post gain>=.005 versus that
arm in both splits. Oracle is reported, never an adoption candidate. Keep all
failures. These pilot means are not population confidence guarantees.

Validate mode0 exact parity, tree uniform-observer parity, budget and snapshot
guards, full replay, journal ordering, raw metric reconstruction and source
snapshots. Production remains untouched. A failed result leaves acquisition
improvement unproven, rather than relaxing thresholds or reporting a component
diagnostic as an agent-level rescue.
