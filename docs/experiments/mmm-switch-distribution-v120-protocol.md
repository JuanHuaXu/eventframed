# Switch-distribution diagnostic

Frozen before quality scoring. All 2688 v120 trajectories are consumed research
data, including the phase historically called confirmation. No fresh validation.

Source: van Erven, Grunwald and de Rooij, Catching Up Faster by Switching Sooner,
2011 manuscript, equations 10-11 and section 2.3:
https://ir.cwi.nl/pub/21169/royal.pdf . Use the basic finite strategy set, not the
paper's expanded frozen-strategy family or its asymptotic guarantees.

Arms [0,1,2,3,10,11]. Prior on segment count is 2^-m; switching time mass is
1/[t(t-1)] for one-based t>=2. Active (will switch again) and stopped states
each receive half the initial expert prior. After emission at one-based t,
active states switch with hazard 1/(t+1). Half the switched mass stops;
half remains active. New expert drawn from the same declared prior. Self
transitions are allowed. Stopped states never switch.

Two frozen priors: uniform, and [.95,.01,.01,.01,.01,.01] matching the earlier
Markov mixer. Each gets a no-switch Bayesian mixture control. No tuning after
scoring. Likelihoods use issued Bernoulli expert probabilities floored to
[1e-12,1-1e-12]. Retrain no experts. Replay each prefix using only origins j<t
with !Missing and j+Delay<=t; omit other emissions but retain clock transitions.
This is a conditional working mixture on a fixed issued-forecast tape. It does
not model informative arrival/nomination, nor inherit full-feedback guarantees.

Prediction is the posterior mixture, not MAP. Evaluate expected Brier, expected
accuracy and log score using Q only in the scorer. All256 and terminal64 windows.
Compare each candidate against its matched no-switch control and issued Markov.
Every cell must have paired mean +/-3.5SE gain lower >=-.01. In terminal windows
of changing cases 1,2,4,5,7,8,19,20, require mean gain >=.005 and lower >0 against
both controls. These are exploratory intervals, not simultaneous confidence
sequences. No rescue is declared unless all gates pass.

Validate the HMM against independently enumerated latent state paths; normalization,
one expert, no evidence and the no-switch evidence lower bound. Poison unavailable
labels and Q at sampled clocks. Replay artifacts byte-exact and retain failures.
Reference prefix replay is O(K*T^2) time and O(K+T*K) tape memory per trajectory;
not a hot-path performance implementation. Full-feedback recursion alone is O(K)
per frame; delayed revisions prevent claiming that cost for this reference.
