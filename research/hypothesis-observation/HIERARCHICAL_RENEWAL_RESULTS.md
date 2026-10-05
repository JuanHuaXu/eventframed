# Shared-versus-local freshness: mixed evidence

Two different tests were completed. The broad consumed-trace diagnostic FAILS;
the separate exact mixed-counterfeit control PASSES its limited declared screen.
Neither is a closed-loop rescue or real-world authenticity certificate.

## Fixed acquired traces: FAIL

The [v14 protocol](HIERARCHICAL_RENEWAL_PROTOCOL.md) averages three finite
models: shared copied mechanism, shared genuine mechanism, and per-type modes,
with frozen prior weights(.25,.25,.5). Evidence updates both component posteriors
and their mixture weights. Contradiction removes an impossible component, never
the observed data. All640 v13 episodes and six original action traces are kept.

Primary uncertain-mixed traces pass8/14 gates, including0/4 required gains.
Split1 expected final Brier:

| Case | Local modes | Hierarchical | Hierarchical gain |
| --- | ---: | ---: | ---: |
| Independent20 | 0.147530 | 0.146027 | 0.001502 |
| Copied20 | 0.272410 | 0.275868 | -0.003458 |
| Mixed20 | 0.223989 | 0.224164 | -0.000175 |
| Copied05 | 0.037154 | 0.035670 | 0.001484 |
| False renewal20 | 0.378271 | 0.378295 | -0.000024 |

Both copied20 protection intervals cross the-.01 boundary, and no required gain
interval has a positive lower endpoint. Confidently-wrong fractions in these
split1 cells are unchanged. This is not useful quality improvement on the old
acquired evidence. Intervals remain descriptive paired mean +/-3.3 SE over64;
all results are consumed-data diagnosis, not fresh confirmation.

Artifacts: [full forecasts](hierarchical-renewal-v14.json),
[all gates](hierarchical-renewal-v14-summary.json),
[independent verification](hierarchical-renewal-v14-reference.json).

## Exact mixed-counterfeit control: limited PASS

The separately [predeclared control](MIXED_RENEWAL_EXACT_PROTOCOL.md) uses a
fixed16-credit schedule, enumerating1024 report vectors and16 hypotheses under
four true renewal mechanisms. It adds even-type and odd-type counterfeit cases
that were missing from v14, without changing the model or priors.

| True renewal mechanism | Certain-fresh Brier | Local Brier | Hierarchical Brier | True-law oracle floor |
| --- | ---: | ---: | ---: | ---: |
| All genuine | 0.249656 | 0.279702 | 0.268299 | 0.249656 |
| All counterfeit | 0.605689 | 0.491836 | 0.487415 | 0.482605 |
| Even types counterfeit | 0.425130 | 0.353639 | 0.362350 | 0.316807 |
| Odd types counterfeit | 0.498693 | 0.442173 | 0.452033 | 0.407727 |

All6 control gates pass. Hierarchical inference gains0.011403 Brier when
renewals really are globally genuine. In mixed-counterfeit cases it harms
Brier by0.008711 and0.009860 versus local inference. Those harms are real under
this exact finite law, though within the predeclared0.01 tolerance. The latter
passes by only0.000140; do not call that robust protection.

All-counterfeit confidently-wrong probability falls from0.16824 under certain
freshness to0 for both uncertainty models. Classification accuracy is unchanged
among the three models in each of these fixed-schedule cases. These are changes
in predictive probability quality, not demonstrated gains in target recovery.

Because this enumerates the entire declared population, its risk numbers have
no Monte Carlo sampling uncertainty. That does not establish the model family,
schedule or noise rates as representative of actual agents. The oracle is not
available to the deployed models.

Artifacts: [all exact results](mixed-renewal-exact.json),
[independent reference](mixed-renewal-exact-reference.json).

## Verification and next step

Component tests check128 six-outcome paths in two channel orders, component and
target posteriors, predictive-evidence chain rule, zero-evidence elimination,
no spurious cross-type learning without renewal, and mixed-mode recovery.
Independent batch enumeration reconstructs112,820 v14 forecasts and component
evidences; maximum forecast error2.265e-14 and log-evidence error1.066e-14.
All640 results replay exactly. The exact control has independent cross-language
verification of32 risk/mass/false-confidence quantities, maximum error2.78e-16.
Its reference does not independently check classification tie behavior.
The complete exact-control result also reproduces identically on replay.

The narrow pass justifies asking whether acquired evidence can distinguish
shared from local mechanisms before reuse. It does NOT justify forcing a larger
shared-model prior or promoting the failed v14 trace result. A prospective
closed-loop comparison still needs both mixed-counterfeit mechanisms, equal
cost accounting, and all original efficacy/protection requirements. All seven
directions remain open. Production and the whitepaper are unchanged.
