# Gate tail refinement v2: all candidates FAILED

10,240 streams, ten scenarios, two splits,512 steps per stream. Fixed, grid,
90%-fixed wealth mixture and50%-fixed wealth mixture share each generated
outcome tape. Source and protocol hashes are embedded in the raw gzip artifact.

Confirmation moderate-shift results (restricted delay charges misses in full):

| Gate | Shift128 delay | Shift128 misses /512 | Shift256 delay | Shift256 misses /512 |
| --- | --- | --- | --- | --- |
| Fixed | 162.13 | 1 | 160.26 | 19 |
| Grid | 146.29 | 3 | 142.97 | 30 |
| Anchored90 | 156.54 | 1 | 154.62 | 21 |
| Anchored50 | 149.15 | 1 | 147.54 | 26 |

Anchored90 gains only3.44%/3.52%, missing the frozen10% delay requirement.
Its shift256 miss increase is2/512, within the finite screening margin. The
50/50 mixture improves delay about8% but misses7 additional late-shift cases,
failing the one-percentage-point miss ceiling. Grid misses11 additional cases;
it also falls just short of10% gain in shift128 on this fresh sample.

For negative256, misses fixed/grid/anchored90/anchored50 are17/19/17/16.
Thus the tradeoff is scenario-dependent, not uniform inferiority. No candidate
passes all conditions, and none replaces the existing gate.

All candidates preserve the per-arm theoretical null bound by fixed convex
wealth mixing, conditional on v1's null and predictable evidence. A convex
mixture cannot promise to inherit every transient threshold crossing of either
component. No coefficient sweep or threshold relaxation was performed.

The paired normal miss intervals are exactly the frozen screen's approximation,
not valid anytime certificates. In particular, zero observed discordance gives
a degenerate normal interval; it must not be interpreted as zero population
uncertainty. Any operational noninferiority claim needs a suitable rare-event
bound and independent integration data, even if this screening test had passed.

Verification: direct wealth versus log-space algebra, invalid-observation state
isolation, full generated outcome/first-alert replay and summary/source hashes.
The evidence remains isolated from serving. A possible next lead is predictable
rate adaptation to evidence variance, or an explicit speed/reliability operating
point backed by user requirements; neither is grounds to relabel this failure.
