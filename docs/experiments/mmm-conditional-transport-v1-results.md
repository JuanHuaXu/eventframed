# Conditional gate transport v1: direct reuse fails

Frozen [protocol](mmm-conditional-transport-v1-protocol.md); complete
[16,000-row artifact](mmm-conditional-transport-v1.jsonl). This is a
two-cell projection of the MMM member observation volume and outcome rules,
not the actual MMM gate or serving path. Design and confirmation seeds were
independent and no threshold was changed after either split.

## Screen

| Case | Immediate design / confirmation | Sparse-delayed design / confirmation | Required |
| --- | ---: | ---: | ---: |
| Stable false flags | 0 / 0 of 1,000 | 0 / 0 of 1,000 | <=20 per cell |
| Common-shift false flags | 0 / 0 of 1,000 | 0 / 0 of 1,000 | <=20 per cell |
| Live shift at clock 256 | 0 / 0 of 1,000 | 0 / 0 of 1,000 | >=800 per cell |
| Live shift from clock 0 | 1 / 1 of 1,000 | 0 / 0 of 1,000 | >=800 per cell |

Overall result: **FAIL** on both splits and schedules. Immediate trials
average about 128 audit requests and 64 delivered labels per side per
context. With 20% independent missingness and 0..31 delay, about 50 per
side per context arrive by the 512-clock horizon. Audit requests and
delivered labels are accounted for separately; the gate has no access to
unrequested, missing, or future labels.

## Why the direct gate loses power

For the unchanged v2 radius,

`r(n) = sqrt(log(8*513/0.02)/(2n))`,

`r(64) = 0.30913`. With both online sides near 64 observations, the
required empirical gap is `0.10 + 2*r(64) = 0.71826`. In these outcome
rules, parity of bits 6..8 is independent of context bit 2. A live-only
switch to bit 2 changes a conditional success probability from 0.5 to
0.05 or 0.95, a population gap of only 0.45. For a clock-256 switch, the
full-history live cell is roughly half pre-change by the horizon, so its
population gap is about 0.225. A favorable *precollected* reference with
256 labels per context would still require a gap above 0.56369 when the
live count is 64, and would add 512 labels to the resource budget.

At a full 0.45 gap, this specific radius needs at least 200 labels **per
side per context** with online reference samples, or 161 live labels per
context with the extra precollected 256-reference-label bank, merely for
the radius threshold to fall below the population gap. These are
population-gap calculations, not a lower bound on every possible test:
sampling fluctuations can occasionally trigger sooner, as the two
1/1,000 immediate start-shift flags show. They explain why changing only
the numerical threshold would trade away the declared error budget.

## Audit and boundary

The independent verifier checked source/protocol hashes, all 16,000 unique
rows, terminal per-cell label counts, audit/missing/pending conservation,
flag clocks, and every summary cell. Exact deterministic replay matched
the full JSONL byte-for-byte. The source only updates the gate after
packet due time and uses outcome-independent audit and delay streams.
This does not prove the gate's error bound under nonstationary common
shifts with asynchronously arriving labels; zero observed false flags
are not a certificate. The experiment also does not evaluate downstream
forecast Brier, external target-law diameter, model-selected contexts,
multi-bucket error, or loaded serving latency.

The v2 toy gain used a larger conditional shift (0.8) and a precollected
reference bank. Directly transplanting its threshold into this smaller,
online-reference signal is rejected. A next candidate must explicitly
account for its evidence and false-split budget, for example a bounded
same-context betting process that uses available labels more efficiently;
the earlier [paired-bet v4](mmm-paired-bet-v4-results.md) remains relevant
but did not test this MMM-style asynchronous audit stream.

Reproduce:

```sh
node research/conditional-transport-v1-verify.mjs docs/experiments/mmm-conditional-transport-v1.jsonl
node research/conditional-transport-v1.mjs --replay docs/experiments/mmm-conditional-transport-v1.jsonl
```

Artifact SHA-256:
`c2fe0f36433955204c20e4ceb7ee3034bcf29bac78375860bd246ab9cb29eaa3`.
