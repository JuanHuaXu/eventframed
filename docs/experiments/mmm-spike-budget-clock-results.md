# Early/late activation: frozen clock screen

Protocol: `mmm-spike-budget-clock-contract.md`. Source indices0-7 are previously
consumed, not untouched confirmation. Each clock independently initializes the
two-expert weights and loss ledger, and then serves32 forecasts. The Markov
incumbent retains its source history. This is not continuous mixture learning.

## Startup (clock0)

672 records,21,504 forecasts,14,505 challenger fits; collection132.28s.

| Method | Expected Brier | Realized Brier |
| --- | ---: | ---: |
| Markov | .198447331 | .198994430 |
| Arrival challenger | .189710473 | see cadence audit |
| Unguarded feedback mixture | .185335013 | .185490982 |
| Budget guard | .189441371 | .190001244 |
| Fixed half-mixture | .188024066 | .188068500 |

Guard improves444/672 records and limits10,216/21,504 proposals. Expected gain
versus Markov is .009005960, with exploratory clustered percentile interval
[.007591116,.010854025]. Guard is worse on average than the unguarded and fixed
half-mixtures, showing a real protection cost. Six records retain expected
harm>.01, versus100 unguarded. Worst harm .013755726 is phase0/majority3/index1/
immediate. No eight-trajectory scenario mean exceeds .01; some majority/mux
means nevertheless have positive pointwise intervals. These are not simultaneous
coverage claims. The realized prefix bound passes to1.18e-16 rounding error.

Independent origin/moment audit passes, but **one fit does not converge within
the frozen cap**: phase1/dependent4/index2/delayed atclock6, final normalized
motion4.644808e-6 at1024 iterations (threshold1e-6). It serves two forecasts.
Final bound movement is about1.97e-10: small bound motion alone is insufficient.
The final capped state remains in the reported results, exactly as returned
by the declared finite-budget fitter. No retry, exclusion or cap increase was
used. The other14,504 fits converge. This numerical limitation must remain
visible even though the guard works algebraically on any bounded forecast.

Clock-specific unit tests pass under the race detector (10.082s), including
initial frozen-fit equivalence, delayed/current/future label poisoning,
positive controls, explicit zero-clock metadata and invalid-clock rejection.
Maximum moment errors1.33e-15 and1.11e-15; maximum integrand evaluations2952.

## Late activation

Atclock224:672records,21,504forecasts,16,862fits; collection172.10s.

| Method | Expected Brier | Realized Brier |
| --- | ---: | ---: |
| Markov | .141689138 | .142127557 |
| Arrival challenger | .153459417 | see cadence audit |
| Unguarded feedback mixture | .142635666 | .143101703 |
| Budget guard | .141818085 | .142363399 |
| Fixed half-mixture | .143522109 | .144036900 |

The guard limits6,286/21,504 proposals and improves265/672 records. It reduces
large expected regressions from70 to9, but has no demonstrated pooled gain:
guard-minus-Markov is+.000128947, exploratory clustered interval
[-.000149686,+.000415874]. Largest record harm is+.022293061 for
phase0/additive-abrupt/index4/immediate. Some majority/mux/parity-to-majority
scenario means show smaller systematic harm. No scenario mean exceeds .01,
but that does not erase the individual counterexamples.

All16,862 fits converge, maximum651iterations; maximum integrand evaluations
2514. Source/moment audit passes, maximum errors1.78e-15 and4.44e-16.
Realized prefix budget passes to1.39e-16 rounding error. Feedback and budget
self-tests pass before scoring; independent expert weight recomputation
agrees within3.33e-16.

## Interpretation and next step

The gain holds at startup and the prior middle window, but does not extend
to late activation as a clear average win. The method is not a universally
better replacement. The guard controls realized excess, not conditional
expected harm on each sampled trajectory. Every counterexample remains in
the machine-readable artifacts.

Next hypothesis: resetting expert weights to half at a mature clock discards
potentially useful earlier comparative evidence. Test uninterrupted forecasts,
feedback weights and budget ledger across the entire stream, with no retuning.
This could also fail: early challenger successes may retain too much weight
after a shift. Measure both possibilities rather than assume continuity fixes
the issue. This is not yet an implemented rescue or a causal attribution of
the late failure.

Full early/late fitted collections have not been replayed. Independent source
audits and initial-fit comparisons cover both; the prior clock128 full replay
is still separate evidence, not a replay of these files. Both feedback and
budget postprocessors replay exactly (excluding elapsed time) for early and
late files. Full spike race suite PASS53.968s. No running sessions remain.

SHA256:

- Early raw: `91af4bf3ff183a7cbce0d2bf4f4064b8a38a9e8c89577ca4ffb227818799cd35`
- Late raw: `80fa5959b41c0a6c1b072bc3899778ee6437c4cb777c406b5fc1b45ee078de8c`
- Early budget: `7ee68369d53e9208036746c2740b51f836d5ff9663295e1ceae7f1ee2312271c`
- Late budget: `e5469c4a880eb4449c8dbe8bb08329810dae3569dee7383dee529ae24f701d53`
- Early summary: `f2ac9bba9e64cf173eba7a0f2736bfeaf476b72733da681714365fc756534ec7`
- Late summary: `225355835ee9c4b7e63d680f446f64a125d55fae5ad1ec8e900d2adbae17dc52`

Production and the whitepaper remain untouched. All seven goals remain OPEN.
