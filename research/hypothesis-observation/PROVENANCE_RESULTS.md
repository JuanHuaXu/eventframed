# Source-aware observer v2 results

1,024 episodes: four cases, 128 per case per split, four matched-budget arms.
This is a known finite-hypothesis simulation. See PROVENANCE_PROTOCOL.md and
provenance-v2.json.gz for the frozen model and pre-outcome journals.

## Confirmation

| Case | Naive accuracy | Source-Gini accuracy | Naive final Brier | Source-Gini final Brier | Source-Gini confident wrong |
| --- | --- | --- | --- | --- | --- |
| Independent sources, 5% noise | 91.41% | 100% | 0.13072 | 0.000043 | 0/128 |
| Independent sources, 20% noise | 63.28% | 89.06% | 0.73434 | 0.18847 | 5/128 |
| One source per test | 91.41% | 95.31% | 0.12510 | 0.07826 | 0/128 |
| Falsely independent IDs | 92.97% | 94.53% | 0.10122 | 0.10098 | 7/128 |

Matched screening PASSED. Final-Brier gains over naive have paired z3.3 lower
bounds 0.01567 (5% noise) and 0.27888 (20% noise). Learning-curve Brier gains over
source-aware random / label-entropy are 0.14237 / 0.06088 at 5% noise and
0.16980 / 0.16416 at 20%; all four paired lower bounds are positive. These are
fixed-sample approximate intervals, not anytime confidence sequences.

At 20% noise, naive confident-wrong frequency was 47/128 versus 5/128 for the
source-aware observer. In the false-ID case it was 1/128 versus 7/128: the
source-aware posterior can be more confidently wrong when its source model is
violated. Do not hide that failure behind the matched pass.

One_source stops after eight actual acquisitions and repeats its unchanged
posterior for the remaining eight scored decision slots. Other cases acquire16.
All three source-aware methods reach the same final posterior after exhausting
the one-source bank. Gini changes acquisition order, not the final evidence.
Its one_source final-Brier gain over naive has an interval crossing zero, so the
point-estimate non-harm screen is not a demonstrated superiority claim.

## Scope and audit

Tests passed full deterministic replay of all 1,024 episodes, source hashes,
pre-outcome journal reconstruction, unique pair selection, source exhaustion,
and fair-coin no-update. No hidden truth or potential outcome tape is passed to
the selection function. Independent sources are supplied by the simulator;
neither this algorithm nor distinct identifiers authenticate independence.

No real-data source validation, automatic hypothesis construction, persistent
Go integration, or serving latency claim is established here. Recommendation7
remains incomplete. A next substantive lead is explicit shared-source versus
independent-source model uncertainty, rather than unconditional likelihood
multiplication for each new ID. Provenance evidence from the existing poisoning
gate must bound that model; it cannot certify itself from posterior confidence.
