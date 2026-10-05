# Target-risk acquisition v28: exact objective, failed rescue

2026-10-02. Both fresh design and confirmation FAIL the frozen overall
calibration-rescue and primary modeled-cost observation screens. The exact
model-based Brier objective is implemented correctly but does not establish
better learning under the external truth regimes. Keep it isolated; all seven
whole goals remain OPEN. No production or whitepaper changes.

## Frozen Evidence

[Protocol](mmm-risk-acquisition-v28-protocol.md) retains v27's model and truth
families, replacing only the candidate observation objective and its charged
cost. Two geometries, six regimes and32 worlds per case give384 worlds and
6528 arm runs per split,768 worlds and13056 arm runs combined. Seeds are new;
all policies in each world receive the same potential-label tape.

- [Design](mmm-risk-acquisition-v28-design.jsonl), SHA256
  `283181f2480c693993ed7e0d48869079406c94f9079acc895f0cb24bd38e0ef3`.
- [Confirmation](mmm-risk-acquisition-v28-confirmation.jsonl), SHA256
  `0627dcbda5a0dfd4d357cf8f6f64a6b16cbc72370fbc27b29c402e0281d597f1`.
- [Independent verifier](../../research/risk-acquisition-v28-verify.mjs) checks
  seven frozen source hashes; all pre-label forecasts, posterior updates,
  laws, packets, expected risks, cost admission and stopping. It checks each
  chosen Gram score against a direct covariance sum and checks maximality.
- [Summary](mmm-risk-acquisition-v28-summary.json) retains all per-case gates
  and paired mean+/-3.5SE endpoints. These are not simultaneous certificates.

The risk selector includes both global-hypothesis covariance and the queried
member's individual Beta uncertainty. Exhaustive two-outcome unit controls
agree within2e-14, including already-observed future targets. A future-tape
control changes every unselected label and obtains identical adaptive traces,
forecasts and weights. No source rates enter the model or selector.

## Fixed32 Confirmation

Lower Brier is better; usefulness is mean true probability of the ten packed
members. Bias is packed mean forecast minus true usefulness.

| Case | Brier: local / information / risk | Usefulness: local / information / risk | Bias: local -> risk |
| --- | ---: | ---: | ---: |
| Tight independent | .37917 / .25144 / .25219 | .68562 / .67062 / .66875 | .26407 -> .04197 |
| Wide independent | .28923 / .25112 / .26049 | .67437 / .68750 / .67250 | .27029 -> .06484 |
| Tight reversed | .35674 / .20368 / .20328 | .23507 / .86372 / .86054 | .70139 -> .01951 |
| Wide reversed | .35384 / .20267 / .20311 | .23005 / .86396 / .86329 | .67343 -> .01656 |
| Tight curved off-model | .34901 / .34614 / .33583 | .30199 / .24991 / .29289 | .63829 -> .09265 |
| Wide curved off-model | .25672 / .35875 / .36105 | .31086 / .24308 / .27599 | .59660 -> .06080 |
| Tight matched model | .10060 / .05958 / .05978 | .91619 / .93917 / .92368 | .03277 -> .01493 |
| Wide matched model | .17730 / .17185 / .16756 | .96857 / .97697 / .97490 | -.01963 -> -.00135 |

Independent/reversed whole and priority-weighted risk recovery, and
aligned/calibrated protection comparisons pass in both splits. Overall rescue
still FAILS because packed-bias magnitude upper exceeds .10: confirmation
tight/wide independent .11649/.17105, tight/wide curved .16330/.12580 and
wide aligned .11151. Mean bias alone is not the gate. Design has the same
curved/independent/aligned warning pattern except wide independent.

The curved failure persists despite correct one-step lookahead. On wide
curved confirmation, expected Brier worsens from local .25672 to risk .36105.
Changing the observation objective cannot by itself make an unsuitable
global calibration family describe the true pattern.

## Same Total Modeled Cap

At the primary label cost100000, the random, uncertainty and information
controls can buy all150 distinct labels; risk can buy32. Their forecasts are
then order-invariant up to floating-point rounding. The computational saving
is deliberately not removed from the controls.

| Confirmation case | Brier: all150 control -> risk32 | Usefulness: all150 control -> risk32 |
| --- | ---: | ---: |
| Tight reversed | .21776 -> .20328 | .87206 -> .86054 |
| Wide reversed | .21871 -> .20311 | .87139 -> .86329 |
| Tight curved | .22457 -> .33583 | .28751 -> .29289 |
| Wide curved | .22443 -> .36105 | .31104 -> .27599 |
| Tight matched | .05364 -> .05978 | .94020 -> .92368 |
| Wide matched | .15391 -> .16756 | .97697 -> .97490 |

The primary Goal7 screen FAILS in both splits. At label cost1000000, random
uses54 labels, uncertainty/information53 and risk32; the broader advantage
still is not established. All cost sensitivities are in the summary rather
than selecting the most favorable one as a new primary result.

These are hypothetical cost units, not an empirical equal-wall-time or actual
agent-acquisition result. Also, more labels do not automatically improve
external risk under a misspecified hierarchy: noisy local corrections can
harm a good global shape. Under the matched hierarchy, the150-label control
has lower Brier than risk32 here. Do not interpret the reversed comparison
as a universal benefit from withholding evidence.

## Performance and Audit

The existing three-repetition component benchmark measures130.07-130.72us
per150-event risk selection,35328 bytes and3 allocations, versus about7.9us
for information. Maximum recorded model-arm elapsed time was7.674ms design
and7.893ms confirmation. This excludes true-law scoring, packet sorting,
label acquisition, database, queues and actual serving; it is not a p99
serving guarantee. The generator took8.160s/8.754s for the full split.

Normal two-package tests, module race tests, vet and diff checks pass.
Confirmation Go replay exactly reproduces all6528 arm records excluding only
five elapsed timing fields, including random-index traces. It reuses recorded
truth/label tapes and does not claim to independently regenerate Gamma draws.
The JavaScript verifier reconstructs all13056 arms from separate calculations.

The local generation-path prerequisite for Goal5 was rechecked read-only:
`GET http://127.0.0.1:11434/api/tags` still advertises only
`nomic-embed-text:latest` with embedding capability. No generation requests,
model installation, credential discovery or production changes were made.
This limits that specific prepared local-agent pilot, not the other research
paths or the possibility of a separately authorized generation endpoint.

## Next Lead

Test model adequacy and nonlinear/contextual structure before another
observation-only rescue. Keep a strong incumbent and off-family controls;
adding a feature suggested by the consumed curved case must be explicit and
must face new patterns. The target-risk selector may remain a mechanistic
comparator but is not adopted. Actual acquisition costs, source independence,
untouched agent outcomes, temporal recovery and loaded freshness remain open.

Reproduce: `node research/risk-acquisition-v28-verify.mjs` and
`EVENTFRAME_RISK_V28_REPLAY=docs/experiments/mmm-risk-acquisition-v28-confirmation.jsonl go test ./internal/researchcalibration -run '^TestReplayRiskV28$' -count=1`.
