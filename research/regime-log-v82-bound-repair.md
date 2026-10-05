# Diagnosed bound housekeeping, not a relaxed forecast check

The frozen INITIAL run failed two cache checks on the old numeric envelope
values. The separate frozen33-case read-only diagnosis found forecast defects
at most3.33e-16 and evidence defects at most1.94e-12, while bound differences
reached1. It reproduced old phantom discarded mass~1e-14 with NO omitted paths.
The old loose conditioning bound amplified that roundoff over unavailable
likelihood-one rows. Thus copying its envelope is not an independent bound oracle.

Small upstream housekeeping repair: if every valid unique next path is retained,
mathematical removed mass is EXACTLY0; do not turn floating normalization error
into a fictitious removal. A missing factor is exactly1, so it cannot amplify TV;
only the true issue transition acts on that previous envelope. Otherwise preserve
the existing conservative amplification and removed-mass formulas/vacuity.

Initial failed sources/logs and all diagnostic sources/logs stay immutable.
Do NOT claim V81 envelope numeric parity: record comparison count and maximum
difference. Cold/resumed log bounds now compare within unchanged2e-11, with a
NEW independent full (UNRESTRICTED) dense-TV check at issue, old reveal and pair
phases. All law/evidence/joint/point/dense/reference tolerances stay2e-11; no
quality, support, numeric-log or resource acceptance rule was loosened.

The returned envelope is an analytical model-truncation calculation, NOT an
interval-arithmetic numerical guarantee or an external target-law/AP authority
certificate. Floating comparisons allow the declared2e-11 budget in these finite
fixtures; untested larger-domain/certified numerical coverage remains open.
