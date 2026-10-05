# Known-law error budget v16 diagnostic

Frozen before calculation on consumed v15 traces. No learner receives oracle
information, and no new success criterion replaces v15's failure.

For a recorded forecast p and acquired mask/value, compute q=P*(Y=1|observed
values,current generator and regime) by enumerating512 inputs under the known
synthetic input law. Both target families have independent .05 label noise.
Conditional expected Brier decomposes exactly into:

q(1-q)+(p-q)^2 = .05*.95 + [q(1-q)-.05*.95] + (p-q)^2.

Report the middle term as observation deficit and the last as forecast deficit.
Forecast deficit includes finite-label estimation, model misspecification,
calibration and unknown-regime effects; it is NOT proof of a single root cause.
The .0475 floor assumes the rule/regime is known and its relevant coordinates
are read. That sensing is feasible within6 coordinates (old high3 plus mandatory
bit0 costs4; new low3 costs3), but rule/regime knowledge is unavailable to the
learner. This is a nondeployable reference, not an attainable online guarantee.

Also report realized Brier. Do not force sample realized loss to equal expected
loss; their difference is finite-sample noise. Use equal per-stream full/post
means, separate all modes, families, generators and splits, and analyze arms0/2.
Stream the artifact. Test truth-table probabilities, correlated conditioning,
the squared-loss identity and observed-mask consistency. Preserve source hashes.
