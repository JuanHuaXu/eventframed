# V61 retention selector: component checks pass, integration remains open

The [direction and mathematical contract](../../research/tree-v61-retention-direction.md)
investigate V60's retention tradeoff. This is NOT a new complete learning-policy
experiment, recovery rescue, Anti-Pigeon certificate, untouched task result,
loaded serving test or equal-total-cost observation win. All seven WHOLE goals
remain OPEN/ACTIVE. Production and previous frozen inputs are unchanged.

## What exists

An isolated Go selector holds a uniform three-expert, member-specific switching
model. Transitions use 1/n on each member's original trial ordinal. Arriving
first/second evidence changes its original factor and replays only that member's
bounded suffix, never multiplies today's weights by an old loss. W2 requests
use smoothing at their origin; the first and second factor refer to the same Y.
Missing packets contribute no factor, and issued forecasts are immutable.

Seven component roots race-pass, including 2,000 comparisons against explicit
expert-path enumeration. Checks cover reordered delayed evidence, missingness,
joint versus independent-trial negative controls, naive-arrival negative control,
owner/epoch/time/replay/cap fences, zero support with atomic rejection, expert
permutation, member separation and immutable issue snapshots. Vet passes.
These are mathematical/lifecycle tests, not evidence of improved real answers.

The separate joint-law adapter derives both clean Y forecast and W1/W2 joint
from a single declared distribution on (Y,W1,W2), so those outputs cannot be
supplied as unrelated numbers THROUGH THAT ADAPTER. Three additional roots
race-pass and vet passes. A nonidentification negative control shows that two
latent laws can give identical observations but different clean forecasts:
internal coherence does not authenticate a model as physical truth. The lower-
level selector intentionally remains generic; future integration must use the
adapter and actually extract the base experts' issue-time latent laws.

## Measured cost and the memory blocker

Apple M4, Go 1.27.1 darwin/arm64, serial component only. Original three-repeat
benchmarks: maximum-depth origin smoothing 1.791-1.821 microseconds and 64-row
original-position replay 1.966-1.986 microseconds, zero allocations. The original
unused-output weight benchmark is not relied upon for a serving claim; a separate
public-API readback uses observable sinks and retains the first benchmark.

Public-API readback over three repetitions:

| Operation | Time | Allocated bytes | Allocations |
| --- | ---: | ---: | ---: |
| Next weights, observable sink | 4 ns (integer reporting) | 0 | 0 |
| Smoothed original W2 query | 1.860-1.879 us | 0 | 0 |
| Constructor, 150-member capacity | 36.279-36.784 us | 1,851,441 | 2 |
| Constructor plus 64 issues/first replies and original W2 request/reply | 52.523-53.044 us | 1,851,442 | 2 |

The complete row serves one member while allocating capacity for 150 members;
it is NOT 150 members' full workload. It excludes three base models, their
setup/updates, acquisition, scheduling, persistence and serving. At 200-member
cap the original constructor allocates about 2,465,841-2,465,842 bytes.

Three unshared V60 constructors plus the measured 150-member selector would
charge about 10,001,073 bytes before other integration work, above 8 MiB. Do
not add cheap selector timings to an old warm baseline and call the new full
policy compliant. Sharing immutable model templates or reducing redundant
ledger storage needs its own equivalence/lifecycle audit and new measurement.
The 400 ms full-loop and all accuracy/recovery gates remain unchanged.

## Reproducibility and repairs

[Component freeze](../../research/retention-v61-component/freeze.json) covers
three compiler-source inputs and four supports; generated Go test-main metadata
is recorded separately. [Commands](../../research/retention-v61-component/completed.json)
record race/vet/bench terminal exit 0. The [public readback](../../research/retention-v61-component/performance-audit.json)
and [joint-law completion](../../research/retention-v61-law/completed.json)
preserve independent source hashes and terminal evidence.

The first pre-freeze helper incorrectly rejected Go's generated test-main file
in the build cache. A preservation helper then hit EEXIST because the lesson
file had already created its directory. Neither failure ran scientific tests or
changed a gate. [Failure metadata](../../research/retention-v61-preflight/failure.json)
states that the first script was recovered by exact inverse of the recorded
patch, not falsely described as a pre-repair copy. Both local helper repairs
are preserved; no global instructions changed.

## Next whole-goal work

Expose actual base-expert (Y,W1,W2) laws before issue, reduce duplicated storage
without changing laws, wire the scored mixture and SAME-mixture acquisition,
freeze full controls, then run and independently audit the complete candidate.
Broader independent seeds, useful error-controlled splitting, untouched labeled
agent tasks, loaded freshness and equal TOTAL-cost observation tests are still
required. Do not open sealed confirmation just to find a passing variant.
