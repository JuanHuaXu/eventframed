# Guide/mixture disagreement v18 diagnostic

Consumed v17 current-mode traces only:576 streams. This is post-hoc diagnosis,
not a new policy test or a change to v17's failure. The source and dependent
known-law diagnostic code are embedded in the output, which binds the raw hash.

Clustered majority shift128 confirmation post window,4608 frames:

| Stop category | Frames | Observation-deficit sum |
| --- | --- | --- |
| Both guide and mixture confident | 3904 | 93.50608 |
| Guide confident, mixture uncertain | 427 | 29.81593 |
| Budget/other | 277 | .58843 |

Uncertain means emitted p strictly between .1 and .9, matching the existing
confidence thresholds. About9.3% of all frames (427) stop confidently at the
guide while the emitted forecast remains uncertain. They carry about24.1% of
the known-law observation deficit. Most deficit remains where both are confident.
Thus the mismatch is real, but cannot explain the whole stopping limitation.

Multiplexer has546 such disagreement frames,2663 both-confident frames and1399
budget frames. Forecast error is still substantial on budget frames, consistent
with v16/v17; a stopping-only remedy is not sufficient evidence for that task.

## Consequence

Test requiring emitted-mixture confidence before honoring a guide's confidence
stop. The gate must compute the actual current mixture from only the acquired
mask and pre-outcome model state, not consult unobserved fields or future labels.
Compare current stopping, mixture-gated stopping and full-budget control on
fresh streams, with both quality and cost gates frozen beforehand.

Do not call agreement a correctness certificate: two confident predictions can
share the same misspecification. Independent calibration or evidence of remaining
information value remains another lead if the narrow gate is insufficient.

Verification: stop-category boundary tests pass; original guide confidence is
checked against every confidence stop. All source hashes are verified, and
known-law conditional calculations reuse tested v16 functions. No online code
or previous experiment changed.

Artifact: [stopping disagreement](mmm-stopping-disagreement-v18.json).

```sh
python3 -m unittest discover -s research -p test_stopping_disagreement.py
python3 research/stopping_disagreement.py docs/experiments/mmm-fullbudget-v17.jsonl.gz NEW.json
```
