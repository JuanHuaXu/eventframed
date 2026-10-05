# Anti-Pigeon evidence gate refinement v1

Frozen 2026-09-12 before evaluation. Roadmap item 3; isolated evidence gate, not
the target-law diameter certificate or a production sharing policy.

## Mathematical contract

Let D_t in [-1,1] be reference correctness minus member correctness. Under the
null, abs(E[D_t | F_(t-1)]) <= .15 at every step. Outcomes must be genuinely
available and the pair chosen before its outcomes. Eight fixed starts j*64,
j=0..7, have two signed wealth processes each, initially one.

For sign s in {-1,1}, X_t=s D_t-.15. Update W_t=W_(t-1)*(1+lambda_t X_t).
Choose lambda_t from past data only, in [0,.8]. Factors are >=.08 and under the
null have conditional expectation <=1. Thus each W is a nonnegative
supermartingale. Average signs, and alert if any start's average reaches800.
Ville plus the union bound gives per-arm probability <=8/800=.01, for any
monitoring duration under that conditional null. No independence over time is
needed. Finite simulation is a check, not the source of this guarantee.

Four arms: fixed (.25, exactly the v3 gate); grid (equal wealth mixture of fixed
rates .05,.15,.25,.5,.8); adaptive; hedge (half fixed wealth, half adaptive).
Adaptive rate at each signed start is clamp(.05,.8,S/(1+V)), where S=sum past X
and V=sum past X^2, both initially0. Update S,V only AFTER using the chosen rate.
This second-moment adaptive heuristic is ours, informed by predictable betting,
not a reproduction of a particular optimal betting algorithm. All arms retain
the identical threshold, starts and equivalence margin. Log wealth avoids
overflow. The gate latches the first alert and never resets a monitoring budget.

Sources: Waudby-Smith & Ramdas,
[Estimating means of bounded random variables by betting](https://arxiv.org/html/2010.09686v7),
especially capital processes and predictable betting; Howard et al.,
[Time-uniform confidence sequences](https://arxiv.org/abs/1810.08240).

## Frozen matrix and comparisons

512 streams per scenario per split, 512 steps each. Five null scenarios:
symmetric (P(+1)=P(-1)=.5); sparse (both.05); boundary_low (.20,.05);
boundary_high (.575,.425); dependent (symmetric +/-1 after a zero, otherwise
zero, so conditional mean always zero). Five alternatives: moderate shift128,
moderate shift256, strong shift256, negative shift256, weak shift256.
Before changes use sparse null; afterward probabilities (+1,-1) respectively
(.5,.1), (.5,.1), (.75,.05), (.1,.5), (.25,.05).

Design base2026092501, confirmation2026092502; seed=base*1000000+scenario*1000+
stream. No tuning or selection between splits. Preserve all arms and outcomes.

Each candidate must have null alert Wilson95 upper <=.02 in EACH null scenario,
and no increase in pre-change alerts over fixed in either primary shift. Both
moderate shifts must improve restricted mean delay by >=10%, with paired
normal z=3.3 lower bound on improvement >0. Delay=first_alert-change for valid
post-change alerts; charge the full remaining horizon for misses OR premature
alerts. This prevents conditioning only on detected cases. Strong/negative/weak
are guardrails: no mean restricted-delay regression >10 steps. Report every
miss, pre-change alert and conditional detected-only delay separately.

Intervals are approximate finite-experiment comparisons, not confidence
sequences. Selecting one of four gates after these results does NOT grant a
joint 1% selection guarantee: each declared arm is valid separately; another
confirmation is required before adoption. Success here is gate-level only;
MMM forecast and actual sharing integration remain necessary.
