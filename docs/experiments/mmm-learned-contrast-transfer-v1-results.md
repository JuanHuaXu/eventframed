# Known-hypothesis acquisition: stress transfer

**Frozen overall screen: FAIL on both fresh splits.** The unchanged v2
11-rule learned-disagreement selector shows robust post-change score gains
across several altered generators, but it does not meet the predeclared
late-change recovery lead. This is a Goal 1/7 synthetic component study,
not a whole-goal or serving result.

## Design and audit

The [frozen protocol](mmm-learned-contrast-transfer-v1-protocol.md) used
seven cases, 16 independently refitted baselines x 16 streams per case and
split, 512 clocks and exactly 128 label requests per arm. Random, uncertainty
and learned arms share the same context/outcome/missingness tape. Only due,
requested labels update the learner. Evaluator truth is not passed to the
selector. The [independent verifier](../../research/learned-contrast-transfer-v1-verify.mjs)
reconstructed all 1,792 trajectories and 5,376 arm trajectories per split,
including generator probabilities, every per-tick score, delivery chronology,
requests, fit-cluster intervals, recovery and frozen gates. Source hashes match;
confirmation recollection, including a full race-detector replay, is
byte-identical.

Archive SHA-256: design
`8515e4fdcb199aea581e4607285abbded89ae302e8732f34cc25257d652b797a`;
confirmation
`5920cf76618fb82805570a224e4739a748eabe14c23635d7732ce8fe52b6dec5`.

## Results

Lower expected Brier and restricted mean recovery clocks are better. Columns
show random / uncertainty / learned. Fit-cluster intervals use 16 fit means
and `mean +/- 3.5 SE`.

| Case | Design post expected Brier | Confirmation post expected Brier | Frozen case result |
| --- | --- | --- | --- |
| Early bit2 | .1368 / .1414 / **.1236** | .1348 / .1402 / **.1226** | Score and recovery pass |
| Late bit2 | .3069 / .3098 / **.2756** | .3097 / .3152 / **.2797** | **Recovery fails** |
| 20%-noise bit2 | .2668 / .2691 / **.2541** | .2705 / .2714 / **.2547** | Strong score pass |
| 32-clock delayed bit2 | .2349 / .2332 / **.2245** | .2388 / .2341 / **.2204** | Design strong-gain interval inconclusive; confirmation passes |
| 10%-bit2 context | .1841 / .1844 / **.1607** | .1892 / .1824 / **.1663** | Score and recovery pass |
| Stable 20%-noise | .1785 / .1786 / .1785 | .1783 / .1782 / .1782 | Non-harm passes |
| Majority out-of-family | .3069 / .3093 / **.2958** | .3040 / .3078 / **.2940** | Relative non-harm passes; no recovery claim |

The learned arm exceeds the `.01` post expected-Brier gain with positive
fit-cluster lower endpoints against both controls in 4/5 known cases on
design and 5/5 on confirmation. All known-case non-harm, early-64 non-harm,
stable and majority relative non-harm, budget and arrived-label parity gates
pass. The measured skewed-case bit2 frequency is .0983/.1007, matching its
declared 10% input distribution. With 20% label noise, the rule likelihood
is deliberately misspecified and still passes the strong score gate; this
does not establish robustness to arbitrary noise models.

The decisive failure is `late_bit2`, which changes at clock 384. Restricted
mean recovery is 123.8/124.3/**119.1** clocks on design and
124.9/125.0/**120.1** on confirmation. Learned therefore gains only about
4.7-5.2 clocks, short of the predeclared 10-clock lead. It reduces recovery
miss fractions from .727/.703 to .473 on design and .746/.777 to .535 on
confirmation, but many trajectories are censored at the 128-clock horizon.
The score gain and miss reduction are useful evidence; neither makes the
frozen recovery gate pass. Delayed feedback also has no first-32-clock
post-change labels by construction, so its near-identical first-64 forecasts
must not be described as instant adaptation.

## Cost and limits

The unchanged learner state is 232 bytes. Four 20,000-sample isolated runs
measured update p99 at 1.54-5.21 us and forecast-plus-select p99 at
84-167 ns. `go vet`, the package suite and a targeted race test pass. These
measurements exclude fitting, retrieval, storage, queues and request latency.

This transfer supports a bounded, declared known-rule acquisition advantage
under several shifts, not automatic hypothesis invention or real-agent
answer improvement. The majority result is only relative non-harm against
weak controls: its post expected Brier remains about .294-.296. No arm is
promoted and no threshold is retuned on these consumed cohorts. Production
and the whitepaper are unchanged; all seven whole research goals remain open.
