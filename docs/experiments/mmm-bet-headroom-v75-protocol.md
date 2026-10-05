# Betting headroom v75

Deterministic mechanism diagnostic before another runtime candidate. No fresh
outcome/agent experiment and no modification of v74's failure. Use exact laws
for homogeneous/sparse/negative/weak alternatives, a heterogeneous boundary null
and zero-evidence control. Oracle moments define q/m, with v74's floor and eta.

For each sign compute the worst permitted X=sZ-.15 over every channel and
D in [-1,1], including endpoints that have zero probability under the oracle.
Choose a cap min(.8,.92/(-min X)) when min X<0, otherwise.8. This preserves
factor positivity >=.08 without relying on the oracle's narrower support.
Maximize E[log(1+lambda X)] on [0,cap] by its decreasing derivative, with
boundary optima handled explicitly. The true law is used only in this offline
diagnostic. A prospective runtime would need predictable learned probabilities,
would suffer learning/transient costs, and could choose poor bets.

Check probability mass, mean-zero control variate, all-channel factor bounds
and optimization against a10001-point rate grid for both signs in all six
cases. These numerical checks support the algebra, not universal inference.
Source basis for predictable evidence betting:
[Waudby-Smith & Ramdas, Estimating means of bounded random variables by betting](https://arxiv.org/abs/2010.09686).
Our particular cap/optimizer is not attributed verbatim to that paper.

A per-step expected log-growth advantage is NOT a demonstrated reduction in
restricted mean stopping time, a global oracle performance ceiling, or evidence
that a learned algorithm passes the research gates. Preserve that distinction.
