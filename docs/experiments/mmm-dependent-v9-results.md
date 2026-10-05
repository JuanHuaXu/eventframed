# Dependent-input v9: transfer screen FAILED

192 streams, 98,304 frames, four arms; eight streams per generator/scenario/split.
Protocol and exact research source text are embedded with hashes in the raw
artifact. No tuning or seed replacement occurred between design and confirmation.

Confirmation shift128 post-change Brier (lower is better):

| Input distribution | Fixed MMM | Adaptive retained | Static retained |
| --- | --- | --- | --- |
| Independent fair | 0.206168 | 0.173925 | 0.184307 |
| Independent biased | 0.187330 | 0.176152 | 0.178162 |
| Correlated latent cluster | 0.146933 | 0.147104 | 0.150541 |

Both retained variants failed the >=.005 shift-gain requirement for clustered
inputs in both splits. The adaptive variant additionally exceeded the .01
harm allowance on fair-input recurring design streams: post Brier
0.193443 -> 0.204726 (harm .011283). Static retention had no >.01 mean-harm
failure, but insufficient gain is still a failed screen. Replacement results,
all windows and both splits remain in the summary, not just selected successes.

This weakens the robustness of v8's finite synthetic pass. It does not establish
that every retained method fails, nor prove uniform marginalization caused the
correlated-input failure: the baseline is also stronger on clustered inputs,
and limited fitting samples/observation choices can contribute. The forest's
uniform integration is explicitly misspecified here; a controlled conditional-
integration ablation is needed to isolate its impact. Do not relax the original
criterion or adopt a candidate based on the two successful distributions.

## Verification

- Independent-fair arm behavior matches the original v8 runner exactly on a
  shared stream, excluding timing instrumentation.
- Generator frequency checks distinguish biased marginals and correlation.
- All192 records fully replay, including choices, weights, inputs and labels.
- Raw-artifact audit recomputes every full/post Brier and accuracy, checks
  forecast bounds, <=6-coordinate budgets, observed values against hidden input,
  label-delivery ordering, unique stream keys, and embedded source hashes.
- Targeted race tests and command vet pass. The initial Python audit used Go
  field capitalization instead of the existing lowercase JSON tags; corrected
  before summary creation. This changed no experiment or raw result.

Reproduction from the repository root:

```sh
go run ./cmd/eventframe-observation-dependent NEW.json.gz
python3 research/dependent_summary.py NEW.json.gz NEW-summary.json
EVENTFRAME_DEPENDENT_ARTIFACT=NEW.json.gz go test ./internal/observationlearners -run '^TestDependentArtifactReplay$' -count=1
```

Artifacts: [raw](mmm-dependent-v9.json.gz),
[summary](mmm-dependent-v9-summary.json), [protocol](mmm-dependent-v9-protocol.md).
These are pilot mean screens, not population confidence guarantees. The two
splits reuse fitted incumbents, and still share the original label family.
Directions1,2,4 remain open; production behavior is unchanged.

Next discriminating test: keep the tree, data, masks and labels fixed while
replacing uniform integration by an estimate of the joint input distribution
from past admitted audits only. An oracle joint distribution can be a diagnostic
upper reference, never a deployable result. Separately test whether a smaller
challenger weight merely trades away the required shift gain for protection.
