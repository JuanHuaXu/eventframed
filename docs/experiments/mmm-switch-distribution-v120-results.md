# Switch-distribution result

Status: FAIL for both frozen priors. Research-only; no production changes.

| Candidate | Non-harm checks | Required improvement checks | Result |
| --- | ---: | ---: | --- |
| Uniform switch prior | 334/336 | 0/64 | FAIL |
| Baseline-heavy switch prior | 334/336 | 1/64 | FAIL |

These are counts of exploratory protocol gates, not independent Bernoulli trials.
Both non-harm failures occur against the matched no-switch control on terminal
parity-to-majority with schedule0, one in each consumed phase. Wide uncertainty
crosses the -.01 threshold; this is not proof of a large negative mean effect.

For phase1, delayed schedule, terminal parity-to-majority: issued Markov Brier
.102272; uniform switch .096963 versus matched no-switch .108138; baseline-heavy
switch .098247 versus matched no-switch .108182. These modest average gains do
not reach the protocol's broad recovery requirement. Stationary additive Brier
is .222702/.221747 for the two switch priors versus Markov .221755.

The oracle diagnostic shows headroom with few switches, but the admissible
working mixture does not capture it reliably from the available labels. This
rejects this declared prior/likelihood combination, not all switching methods.
Do not tune a new hazard on these consumed cases and call it confirmation.

Component verification:192 prefix comparisons against explicit latent-path
enumeration, normalization/evidence lower bound and invalid-input checks PASS.
1344 as-of checks poison unavailable labels and hidden Q without changing any
sampled prediction. All2688 trajectories retained. Full artifact replay is
byte-identical (`cmp` exit0); hashes include source, test and frozen protocol. Reference prefix
replay is quadratic in trajectory length and is not a serving benchmark.

Source: [van Erven, Grunwald and de Rooij, section2.3](https://ir.cwi.nl/pub/21169/royal.pdf).
We instantiate its basic switching prior, not its enlarged strategy families or
convergence results. Delayed/missing evidence on fixed issued forecasts is our
conditional working-model adaptation. No informative-selection guarantee.

Next useful diagnostic: separate wrong expert choice from insufficient evidence
by measuring how many revealed, independent labels actually distinguish the
oracle-preferred expert from the incumbent before recovery is needed. That may
motivate a targeted observation policy, connecting goals2/4 to goal7, rather
than another prior or forgetting grid. Any such policy needs equal-cost controls
and must not select observations using hidden Q.

Artifacts: [protocol](mmm-switch-distribution-v120-protocol.md),
[results](mmm-switch-distribution-v120.json),
[replay](mmm-switch-distribution-v120-replay.json).
