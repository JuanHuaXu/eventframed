# New label families and fitting groups v14

Frozen before evaluation; subset priors/model and all online settings unchanged.
Replace the previous XOR-to-single-bit target by two new families:

- majority: at least two of bits6,7,8 before the change, bits0,1,2 afterward;
- multiplexer: bit8 selects bit7 (true) or6 (false), then bit2 selects bit1 or0.

Flip the resulting label with original scenario noise. Reuse only stable05,
shift128, recurring and delayed_missing scenarios (indices0,2,6,7), with their
original timing/delays/audit behavior. Three input distributions, two families,
six independently fitted incumbents, two streams per fit, two splits, two modes
(forest/subset)=1152 streams. Fit2026107201, design2026107202,
confirmation2026107203. Seed scenario index=30*family+10*generator+j; maximum57
keeps families separate from the next split's million-wide seed range. Each
incumbent is reused across modes and splits, so split results are not independent
fitting-set replications. Each incumbent uses4096 matching-distribution samples.

Keep v13 finite-screen rules for each retained subset arm, now in every family:
full/post mean harm<=.01 versus fixed and paired retained forest in every group;
shift128 post gain>=.005 versus fixed in every family/generator/split; clustered
shift128 gain>=.005 versus paired forest in both splits and families. No tuning.

Also report per-fit paired differences and a descriptive95% bootstrap interval
over the six fitting groups (20000 resamples, fixed seed2026107299). This is not
a simultaneous certificate, not a six-group asymptotic guarantee and not a
replacement pass criterion. A finite mean pass with weak interval evidence is
not robust population validation. Frame count is not the independent sample size.

Stream compressed JSON-lines instead of retaining all traces in RAM: first line
source/header metadata, then one full record per line. Exclusive artifact creation.
Verify every metric/observation budget, mode pairing, fitted-seed independence,
target truth tables, source snapshots and reproducibility. No production changes.
