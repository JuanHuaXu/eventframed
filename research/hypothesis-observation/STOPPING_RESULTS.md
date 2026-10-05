# Cost-sensitive stopping v10: saves reports, FAILED accuracy protection

Completed1,792 fresh episodes. The primary comparison has5,376 trajectories
(full two-step, one-step stop, cautious two-step stop); four earlier controls
also remain in each raw record. `STOPPING_PROTOCOL.md` froze lambda0.005 and
the new cost/accuracy gates. `stopping-v10.json.gz` and its summary contain all
results, decisions and source hashes. This does not revise any v9 criterion.

## Confirmation results

| Case | Full reports | Cautious-stop reports | Full final Brier | Cautious-stop final Brier |
| --- | ---: | ---: | ---: | ---: |
| Independent20 | 16 | 9.703 | 0.156683 | 0.182701 |
| Copied20 | 16 | 9.953 | 0.531646 | 0.542806 |
| Mixed20 | 16 | 10.414 | 0.248298 | 0.241114 |
| Matched05 | 16 | 6.172 | 0.042660 | 0.035210 |
| Matched20 | 16 | 9.922 | 0.288279 | 0.293732 |
| Matched random signal20 | 16 | 10.039 | 0.320969 | 0.325584 |
| Matched misleading signal20 | 16 | 10.008 | 0.304085 | 0.326237 |

The primary cautious-stop screen FAILED three final-Brier protection gates:
independent20 harm0.026018, copied20 harm0.011160 and misleading20 harm0.022152,
each above0.01. Their descriptive paired z3.3 intervals span zero; these are
failed mean protection tests, not certified population harm. All curve-Brier
mean harm ceilings passed. Independent20 mean final harm also exceeded0.01 in
split0 (0.026879), as did several other split0 populations.

The cost gates passed: copied20 saved37.79% of reports; matched05 saved61.43%.
Including the eight fixed signal checks, total-cost savings are25.20% and40.95%.
Penalized final-risk gains were0.019074 [0.002856,0.035292] and0.056590
[0.031283,0.081898]. Those improvements depend on the declared price conversion
and do not override the separate failed accuracy gates. Lambda is a synthetic
utility preference, not a measured database/LLM cost.

The cheaper one-step stopping control generally saved slightly more reports but
did not establish a general accuracy rescue. It was diagnostic only and was not
substituted as the primary treatment after inspecting the results.

## Tail diagnostic, not a new success gate

Inspection of the frozen confirmation artifacts found final-Brier harm>0.1 in
7/128 independent20 and12/128 misleading20 episodes. Maximum observed harms
were1.388635 and0.616888. Only2 and1 of those respective cases had stopped
confidence>=0.9. Thus a simple high-confidence-only guard would not identify
most large observed regressions. These thresholds were post-run diagnostics;
no new statistical coverage claim follows from them.

The mean savings are real in this simulator, but stopping on model-predicted
low value is not an externally certified upper bound on omitted useful evidence.
Even a coherent finite posterior can be wrong for an individual case, and the
test family also contains deliberate prior/signal misspecification.

## Verification and next work

`python3 -m unittest test_stopping_v10 -v` passed all four tests in83.723 seconds,
including complete replay of1,792 episodes, source hashes, stop inequalities,
unique paid slots, normalized/frozen forecasts and separate cost accounting.
Full-policy parity with v9, the one-report horizon and future-suffix independence
also passed. The offline batch took81.734 seconds including counterfactual
controls; that is not a serving or actual saved-CPU benchmark. Once a treatment
stops, its null report slots cost zero and its forecast is frozen; the full
control's later reports never update that treatment.

Do not tune lambda against these confirmation outcomes until a desired
cost/accuracy policy is independently justified. This bounded stopping lead
has not established a safe universal replacement. A calibrated external bound
on omitted influence or better observation families remain possible research,
not implemented guarantees. The next engineering priority is to test the still
unverified persistent shadow integration; real-agent validation also remains
open. These synthetic studies do not complete either direction.
