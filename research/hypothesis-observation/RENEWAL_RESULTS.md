# Renewed measurements: strong conditional gains, unqualified overall

The [frozen renewal pilot](RENEWAL_PROTOCOL.md) completed640 episodes across
two splits, five cases and four arms. Every arm spends16 acquisition credits;
ordinary reports cost1 and genuinely fresh draws cost2. All trajectories and
failures are retained. Overall screen: FAIL, with14/16 required gates passing.

## Confirmation split

| Case | Regular-only Brier | Mixed planner Brier | Mixed accuracy | Mixed fresh calls |
| --- | ---: | ---: | ---: | ---: |
| Independent20 | 0.226098 | 0.129272 | 93.75% | 5.469 |
| Copied20 | 0.496218 | 0.210300 | 85.94% | 6.000 |
| Mixed20 | 0.324156 | 0.067544 | 96.88% | 5.875 |
| Copied05 | 0.078088 | 0.00000216 | 100% | 5.000 |
| False renewal20 | 0.492846 | 0.744449 | 62.50% | 6.000 |

Gains against regular-only are0.285918 [0.049975,0.521861] for copied20 and
0.256612 [0.029029,0.484195] for mixed20 in split1. Genuine new evidence can
beat the copied-report information floor because it changes the observation
family; that does not contradict the earlier bound. The simulator knows the
fresh measurement mechanism, not the selected latent target.

The screen still fails. In split0, mixed20 gain versus regular-only is0.087166
[-0.133763,0.308095], and versus entropy is0.092305 [-0.120889,0.305500]. Both
fail the required positive lower bound. Do not discard this split or promote
the candidate based only on the better split1. Intervals are descriptive paired
mean +/-3.3 SE over64 episodes, not simultaneous or anytime-valid guarantees.
The nonharm gates use the predeclared mean ceiling, not certified noninferiority.

## Freshness is a load-bearing assumption

The false-renewal case deliberately labels copied first reports as fresh.
Compared with regular-only, mixed-planner Brier worsens by0.251602
[0.025694,0.477511] in split1. Confidently wrong outcomes (maximum forecast
probability>=.9 but incorrect argmax) rise from0/64 to24/64, or37.5%.
The negative control is excluded from the conditional efficacy screen but
categorically prevents a robustness or operational-authentication claim.

Freshness here is conditional independence given the hypothesis. A renamed
source, retrieval from another cache, or repeated model answer does not meet
that contract. No real measurement adapter or proof of independence was built.

## Verification and audit trail

Five component tests pass: fresh/ordinary state separation, clone ownership,
one-credit feasibility, parity with the original ordinary-report model, and
complete episode prefix reconstruction. Every actual action is selected before
the simulator accesses its target or tape. All2,560 arms have unique paid slots
and exactly16 credits. Credit-area Brier holds the pre-action forecast during
the entire paid observation cost, so expensive calls do not get free time.

The [independent verifier](verify-renewal.mjs) computes batch posterior weights
by marginalizing independent/copied source modes, rather than importing the
online update model. It reconstructs35,894 forecasts and all Brier scores;
maximum forecast discrepancy is7.78e-16. It verifies source hashes and costs.
All640 episodes were rerun; every trace and computed summary reproduced exactly.
Counterfactual planning CPU is not acquisition cost; no loaded runtime claim
is inferred from this offline experiment.

Artifacts: [full traces and hashes](renewal-v11.json),
[all summary cells and gates](renewal-v11-summary.json),
[independent reconstruction](renewal-v11-reference.json).

## Next research boundary

This is the strongest new conditional observation result in this continuation,
but not a qualified rescue. A follow-up must preserve every case and compare
under a fixed untouched sampling plan; do not add samples until a desired
interval passes. More importantly, admitting renewed evidence requires a model
or external mechanism that distinguishes actual independent measurement from
replay. Test that mechanism against the false-renewal control before operational
integration. Its own uncertainty and acquisition cost cannot be hidden.
All seven research directions remain open, with no production or paper changes.
