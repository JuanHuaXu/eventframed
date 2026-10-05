# V58: rolling paired-evidence partition mixture

RECOMMENDATION, not a confirmed production bug. V57 target-value changes did
not resolve the shared fixed-shape model's recovery/protection failures.
Alternatives remain source dependence, delayed information, and acquisition
cost. No assertion that the proposed model is the uniquely identified cause.

## Primary Research And Distinction

Kontoyiannis, Mertzanis, Panotopoulou, Papageorgiou and Skoularidou (2022),
[Bayesian context trees: Modelling and exact inference for discrete time series](https://rss.onlinelibrary.wiley.com/doi/10.1111/rssb.12511).
Sections 2.2-2.3 and 3.1 describe model priors and exact recursive marginal
likelihood averaging; section 6 connects these to predictive likelihood ratios.
The earlier [CTW paper](https://www.cs.cmu.edu/~aarti/Class/10704_Fall16/CTW.pdf)
motivates bounded tree mixtures. The adaptation below is NOT their temporal
suffix model, Dirichlet implementation, coding guarantee, or convergence theorem.

V70 already tested static depth-three Boolean CART averaging and failed broad
protection. Preserve that rejection; this is a different conditional model:
a fixed hierarchy over PUBLIC baseline ranks, paired noisy observers of the
same latent outcome, and a declared rolling issued-event suffix. It pools
member evidence at multiple resolutions rather than using only a fixed global
family with separate scarce member-rate states. No label-dependent hierarchy,
source authentication or automatic causal interpretation is asserted.

## Declared Model

Depth 7 binary rank hierarchy, maximum 255 nodes. Public base scores determine
ranks; labels never do. At each internal node stop with probability 1/2 or
split with probability 1/2. Each terminal rate has a 21-atom baseline-mean
moment prior of strength 2. Shared noise eta has prior (.8,.1,.1) on (0,.1,.2).
Evidence consists of first-only or paired observations with their joint latent
Y marginalized once. Node marginal evidence is a finite rate sum; subtree
evidence is 0.5*leaf evidence + 0.5*left*right evidence. Bottom-up evaluation
exactly sums this DECLARED finite tree/rate/noise model.

Retain the last 600 issued event positions globally, not the last arrivals.
Expired evidence is removed exactly. A late expired response is acknowledged
but cannot displace newer evidence. This is a truncated-evidence working
posterior, not ordinary full-history Bayes or an inherited change-point model.
Forecasts and observation conditionals come from the SAME current suffix model.
Second observers still measure the original Y, not an independent fresh Y.

Pre-cohort protocol correction: a 300-position suffix cannot retain all 150
origins until first-round nomination with first delays up to 299 ticks. The
600-position contract retains them at nomination; some second responses can
still expire before arrival and must be charged and acknowledged without
learning. This analytical reachability issue was found before cohort collection,
not used to retune a failed empirical result. Equal public scores share one
feature rank; sorted summation also prevents ID order changing fallback priors.

## Invariants And Tests Before A Cohort

Identity/epoch/cap/time checks precede mutation. Prepare updated node paths and
root weights in scratch; commit only after finite/support checks. Pending,
canceled, missing and expired labels are distinct. Queries do not commit
hypothetical observations. Own original forecasts accompany delayed receipts.

Compare every forecast and second-measurement conditional with an independent
whole-suffix rebuild; enumerate all small-depth tree models independently.
Test pairing, zero support, eviction, out-of-order arrival, expired responses,
foreign tickets, epochs, faults, and public-rank permutation invariance. Keep
future generator rates, labels and source availability outside all APIs.

Approximate costs: path update O(depth*3*21*6), forecast O(depth*3), storage
O(255*21 + member_count*64). Predictive-value acquisition adds all-target
conditional prediction work and must be timed, not called constant for free.
No likelihood or source-independence assumption is a real-world certificate.

## Prospective Evaluation

Only after correctness: all 40 consumed V54/V57 diagnostic worlds and all three
schedules, with no-pair, random, uncertainty, noise-class concentration,
information and predictive-value policies. Anchor earlier baseline artifacts;
do not relabel consumed diagnostics as fresh confirmation. Preserve all original
quality, recovery, usefulness, 8 MiB / 400 ms and total-cost requirements.
New pooling remains a forecast proposal, NOT final Anti-Pigeon authority.
No Goal 5 agent or Goal 6 serving claim from these synthetic loops.

Research boundary: new isolated packages/fixtures only; production, private data,
sealed task outcomes, whitepaper and publication untouched. No upstream bug fix,
installation, deployment, deletion, or gate relaxation. Previous goal turn
PROGRESS (complete V56/V57 tests and checkpoint); all seven whole goals OPEN.

Local discovery errors: guessed package/file paths did not exist; actual scoped
`rg --files` identified the prior implementations. No instruction-tree edits or
cleanup follow from these command errors. Checkpoint/source copies preserve
the already dirty tracked files instead of modifying their learning log.
Two attempted multi-file patches failed context verification and made no edits;
readback confirmed this before applying small exact-context patches separately.
