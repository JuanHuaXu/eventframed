# V81 protected latent support, mathematical preflight

All seven WHOLE goals OPEN. Parent V80 checkpoint SHA256
`d2ae1995bfaa0caff9a2ddee5864847606fd532ec1ae7cd27e16d5a50fcd965e`.

## Research and patch reasoning

Confirmed: V80 preserves coherent law and accelerates conditioning, but64
unknown issues can prune every noisy class. Contradictory second measurements
then have zero probability in the restricted working law. Candidates: genuine
impossible evidence, top-weight support exclusion, numeric underflow. A dense
oracle and explicit noise-free witness confirm exclusion in that fixture;
underflow remains a separate long-history concern, especially with reset0.

Doucet/Johansen's tutorial (section3.5, printedp16) explains finite-particle
degeneracy and warns that a healthy-looking weight distribution need not mean
a useful approximation. Hesterberg (1995, section6.1) retains target-distribution
coverage in an importance-sampling mixture. These motivate preserving coverage,
not theorem transfer: this is deterministic constrained beam truncation, NOT
SMC or unbiased defensive importance sampling.

Sources inspected: [Doucet/Johansen](https://warwick.ac.uk/fac/sci/statistics/staff/academic-research/johansen/publications/DJ11.pdf),
[Hesterberg](https://www.stat.cmu.edu/technometrics/90-00/vol-37-02/v3702185.pdf).

Separate fork. At cap overflow and reset>0, retain ALL nine fresh `(class,t)`
reset components, followed by the best remaining carry components to cap.
Keep their actual model probabilities; do not add a likelihood floor, revive
an impossible carry path or renormalize a branch under different masks.
Conditional queries and accepted reveals retain fixed historical supports.
Cold refit uses the same protected selector and creates an explicit support
epoch. With reset0, only the original nine classes exist, so no cap overflow.

Exact-arithmetic support argument: for reset>0 each allowed mask contains a
fresh reset in a positive-noise class. The path resetting at every original
issue has positive prior mass, and either same-Y measurement pair has positive
likelihood when0<eta<1. Thus finite binary-evidence histories have positive
restricted evidence. This is not a useful lower bound over an arbitrarily long
trajectory, approximation certificate, non-harm theorem or immunity to finite
precision. With cap9 the protected mask forces reset-only trajectories; memory
loss is a required NEGATIVE control, not a clever implementation fix.

Falsifiers: missing protected keys, cap violation, changed mask during reveal,
dense/reference/receipt/tower disagreement beyond2e-11, query/update mutation,
non-positive supported reset>0 evidence, or claims hiding reset0 underflow.
Retain all V80 phase/as-of/cache/history guards. Add noisy opposite pairs at
old/latest rows, unknown gaps, reverse arrivals, cap9 degeneracy, and a long
reset0 all-one-then-opposite-pair diagnostic with an independent closed-form
log evidence. Numeric failures must be reported, not interpreted as truth0.

Freeze before unit/race/vet/serial costs; same source/compiler and protected
tracked hash guards. Runtime cap remains9..256, history scope unchanged.
No production/private/sealed/reserved seeds, whitepaper/publication or pushes.
After preflight, a separate frozen common-stream screen is required. It cannot
replace the original full120-cell/native-control/equal-TOTAL-cost protocol.
