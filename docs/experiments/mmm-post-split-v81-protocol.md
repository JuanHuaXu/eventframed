# Post-split forecast diagnostic v81

Post-hoc analysis of all640 v80 trajectories, not fresh confirmation or a rescue.
Replay the exact v80 integration with read-only hooks before and after labels.
Require exact equality of its original per-stream metrics and prediction hashes.
No change to model fitting, observation selection, mixture weights or split gate.

For the mixture-gate arm, before the revealing label, record its observed mask,
selected guide, stored pre-share weights, four expert probabilities, split flag
and output probability. Query immutable base/short/local/pooled models on four
diagnostic masks: actual observation, all9 coordinates, bit2 only, old parity
coordinates6/7/8. Check actual-mask reconstruction of journaled experts. Bit2
and parity masks use knowledge of the experiment's generator and are NOT
implementable feature selectors. Full coordinates have extra observation cost;
these diagnostic calls cannot inform any actual prediction or learning update.

After feedback score these frozen probabilities, aggregating into eight64-step
windows. Report Brier, expert weights, guide choice, feature coverage, support
and model availability. Missing challenger models are not scored as present;
denominators are actual availability counts. Never call full conditioning an
oracle: sparse count models may be worse when conditioning on more variables.

Interpretation: a good actual-mask expert but weak emitted forecast suggests
weighting; a good reduced-mask local model but poor actual-mask model suggests
observation/conditioning mismatch; weak models on all masks suggest model/data
limitations. These are diagnostic leads, not exclusive causal attributions.
No post-hoc best mask/model is published as a validated forecast. Keep v80 FAIL.
