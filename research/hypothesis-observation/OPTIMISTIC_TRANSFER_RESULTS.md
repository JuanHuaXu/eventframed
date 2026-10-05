# Guarded optimism passes the finite allocation stress screen

Under the [frozen protocol](OPTIMISTIC_TRANSFER_PROTOCOL.md), the candidate
passes all180 allocation stress gates:160 protection,10 genuine-gain and10
false-confidence requirements. This is a finite-model rescue, not completion
of the broader observation, real-data or production research directions.

## What changed

The objective now targets the prediction appropriate under the fixed all-genuine
renewal mechanism. Every supported counterfeit pattern still constrains the
forecast's conditional regret against the local model. The actual true pattern
is never supplied. Neither the risk budget(.01) nor any acceptance gate changes.

This separates a decision target from an evidence claim. The candidate does
not assert that all renewals are genuine, alter source posterior weights, or
present the output as an ordinary Bayesian mixture posterior. It pursues useful
prediction under the optimistic mechanism within the same conservative feasible
region. The full model-family adequacy premise remains load-bearing.

## Results

Across all10 allocations, exact all-genuine Brier gains versus local inference
range from0.008014 to0.009586, clearing the0.005 minimum in every allocation.
All previous seven mixed-counterfeit harm counterexamples are retained and pass
protection. All-counterfeit confidently-wrong probability decreases by0.16824
versus certain-fresh inference in each allocation.

Worst population harm versus local is approximately0.010000, essentially the
entire allowed margin: for counts[1,2,2,1] and all-counterfeit renewals, Brier
changes from0.491612304284 to0.501612304283. Protection means tolerating this
declared loss, not guaranteeing an improvement on every environment. There is
almost no spare margin for model misspecification in that cell.

For the earlier counts[2,2,1,1], all-genuine Brier is0.270116 versus local0.279702.
The unguarded hierarchical0.268299 is slightly better in that one genuine case,
but failed other masks. Guarded optimism is not claimed to dominate every
alternative cell; it meets the joint predeclared tradeoff.

## Verification and reference cost

All37,152 supported conditional-law checks pass within floating-point tolerance.
Maximum computed conditional regret is0.010000000000000342. The unchanged
projection solver needs at most3890 cycles, with maximum numerical objective
gap9.86e-13. No convergence failures or discarded vectors occur. Optimistic
targets are checked against the declared all-genuine conditional law. Original
certain/local/hierarchical controls match the independently verified artifacts.
Full result replay is byte-exact, and timing instrumentation leaves every cell
and gate unchanged.

One isolated reference run on Apple M4,16 GiB RAM, Node26.8.1 darwin/arm64,
over10,240 projection calls (including validation and solver work):

| Statistic | Milliseconds |
| --- | ---: |
| Median | 0.003625 |
| Mean | 0.175099 |
| p95 | 0.756166 |
| p99 | 2.670750 |
| Maximum | 34.479834 |

This is one cold/warm reference-call distribution, NOT loaded serving latency.
It excludes constructing the ambiguity laws, retrieval, acquisition, persistence,
queueing and integration overhead. The finite family here has only four measured
source types. Its enumeration grows exponentially with type count; these timings
do not establish a corpus-scale or robot deadline guarantee.

Artifacts: [all cells/gates](optimistic-renewal-allocation.json),
[timed identical result](optimistic-allocation-timing.json).

## Still required

Test disjoint observation structures and noise/model mismatch without changing
the guard after seeing outcomes. In particular, demonstrate what happens when
the true conditional law lies outside the declared ambiguity family. Adaptive
acquisition, uncertain likelihood estimation and empirical coverage require
separate evidence. Nothing in this exact finite test authenticates actual agent
memories or repairs a wrong measurement model. No production or whitepaper
promotion is made, and all seven directions remain open.
