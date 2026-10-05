# Retained subset challenger v82 results

PASS of the frozen finite matched-input pilot in both phases. Unlike the
earlier gate-only improvement, this candidate improves the emitted forecasts.
It is not a generalization result, production deployment or completed roadmap.

[Protocol](mmm-retained-subset-v82-protocol.md),
[artifact](mmm-retained-subset-v82.jsonl),
[evaluator](../../research/retained-subset-v82-summary.mjs),
[research state](../../internal/observationgate/subset_state.go).

640 fresh trajectories,327,680 frames, six paired arms,112 frozen source/evaluator
hashes. Same audit labels and fitting cadence as the count-control workflow.
Artifact SHA256:
`48922dd2056f8c355500ced18e4194dbd9a8ea92a1ad0e83bf932ef1e3a11446`.

## Confirmation

64 trajectories per scenario. Control is v80's mixture-gate MMM workflow,
rerun on the new seeds. Metrics are post-change, or the latter half for stable/
null scenarios. Intervals use independent trajectory differences and frozen z3.5.

| Scenario | Control Brier | Retained subset Brier | Paired gain interval | Control accuracy | Candidate accuracy |
| --- | ---: | ---: | --- | ---: | ---: |
| Stable | 0.048841 | 0.048835 | [-0.000053,0.000065] | 94.891% | 94.891% |
| Member shift | 0.240256 | 0.206659 | [0.026999,0.040195] | 58.508% | 66.565% |
| Common shift | 0.240250 | 0.202992 | [0.030790,0.043727] | 59.656% | 67.389% |
| Recurring | 0.192004 | 0.172771 | [0.013489,0.024978] | 68.884% | 74.670% |
| Null | 0.251638 | 0.251577 | [-0.000087,0.000209] | 49.811% | 49.982% |

Primary member/common Brier gains0.033597/0.037259 exceed the frozen0.005 target
with positive lower bounds. Design gains0.033547/0.037046 also pass. Stable,
recurring and null full/post protection gates pass in both phases. Log loss
also improves on the changed scenarios, but was not used to select the method.

The candidate preserves the count short model and frozen incumbent. Subset
weights are learned from the past labeled sequence, not supplied bit2 or a
post-hoc best mask. Duplicate subset slots in the inner mixture represent a
combined30% prior, not independent evidence or tripled support. The outer law
is scored after this inner correction, and both weight updates use journaled
pre-label probabilities.

## Acquisition and fitting costs

Foreground coordinate costs per frame, control to candidate:

- Stable4.00031 to4.00143.
- Member shift4.73166 to4.61749.
- Common shift4.72995 to4.64499.
- Recurring4.91080 to4.89560.
- Null5.99863 to5.99918.

Every call remains capped at6; equal budgets do not imply identical realized
cost. The same Bernoulli audits provide all training labels. Confirmation
member audit cost is4.50604 coordinates/frame and monitor cost8, shared by the
paired comparison but additional to foreground acquisition. The subset model
is fitted6.531 times per member-shift trajectory on average, in addition to
the original count-model fits. Extra computation is not free label efficiency.

Apple M4, darwin/arm64, GOMAXPROCS10, three500ms microbenchmark repetitions:

| Boundary | Time | Bytes/op | Allocations/op |
| --- | --- | ---: | ---: |
| Subset fit on64 labels | 6.498-6.571 ms | about651264 | 3 |
| Ordinary count foreground fixture | 10.678-10.713 us | 11366 | 38 |
| Ordinary retained foreground fixture | 10.689-10.707 us | 11345 | 38 |
| Forced count-observer fixture | 9.466-9.503 us | 11203 | 33 |
| Forced subset-observer fixture | 2.805-2.818 us | 5585 | 28 |

The forced fixture sets selector weights only for timing, asserts the intended
observer executes, and is not an accuracy experiment. These are different
observation algorithms and may stop after different numbers of coordinates.
The ordinary fixture does not characterize worst-case subset usage. All
foreground figures exclude fitting, persistence, network and queueing. No
production p99 or write-contention guarantee follows. The compiled conditional
table is19,683 cells of two float64 values,314,928 payload bytes per model;
the fit allocation figure includes temporary objects, not just retained state.

The inherited model enumeration is bounded here by512 subsets over9 dimensions
and64 audit labels; it is exponential in dimension if naively generalized.
This successful toy integration is not a scalability proof for arbitrary text
features. Research fitting is synchronous after feedback; production background
publication and staleness contracts still need separate integration testing.

## Verification

- Disabled challenger state reproduces the existing workflow exactly.
- Adding the independent candidate preserves the control stream's metrics,
  observations, split timing and prediction tape in the parity test.
- Every candidate prediction precedes its revealing label and every refit follows
  feedback. Duplicate/out-of-order feedback is rejected without weight mutation.
- Candidate and control split times and audit-fit cadence match in every stream.
- Original full replay passes in58.689 seconds; the archive-manifest verifier
  repeats all640 trajectories successfully in57.385 seconds after adding the
  separate active-path benchmark. Recorded source hashes remain unchanged.
- Focused state/integration/adjacent race tests pass in3.171 seconds; vet passes.

For future replay use `TestSubsetV82ArchivedReplay` with
`EVENTFRAME_SUBSET_ARCHIVED_REPLAY` pointing to the artifact. It checks the
frozen112-file manifest and exact records rather than regenerating that manifest
from a directory that may contain newer experiments. The original experiment
runner remains frozen and was not rewritten to accommodate later files.

## Remaining boundary

Inputs are independently uniform in this experiment, matching the explicitly
uniform partial-input model. That is a declared model assumption, not learned
evidence of independence. The previous subset-model failures on broader/noisier
generators remain failures; this integration cannot erase them.

Next freeze a breadth test with higher-order interactions, alternative generator
fits, input dependence and delayed/missing feedback. Preserve the present model
and weights rather than tuning them on this confirmation set. Test whether
the retained design protects stationary cases while transferring the measured
learning gain. Actual chatbot outcomes, broad false-revocation coverage and
production lifecycle/performance requirements remain open. All seven roadmap
directions stay in progress.
