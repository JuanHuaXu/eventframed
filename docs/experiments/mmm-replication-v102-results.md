# v102: unchanged routing passes the larger finite screen

## Verdict: PASS106/106, within the declared scope

All48 whole-stream and48 late-half non-harm gates,6 stationary interaction-gain
gates and4 recovery-gain gates pass. All four finite null screens also pass.
This advances the bounded immediate-feedback research candidate. It does NOT
complete any whole roadmap direction or establish production MMM performance.

The [frozen replication protocol](mmm-replication-v102-protocol.md) specifies
128 trajectories per phase/case,3,072 streams total,256 scored steps each and
the same twelve arms. All19 shared model/kernel/dependency hashes match v101;
only the experiment driver, protocol, seeds and sample count changed. Routing,
bank priors, model windows, refit cadence, rejection budget and all106 acceptance
criteria are unchanged. Both views of a trajectory share its inputs/outcomes;
they are not separate independent trajectories.

All3,072 streams and null families were generated before any quality summary
was inspected. There was no interim quality look, early stop, pooled v101 data,
algorithm change or post-outcome sample-size increase. Fresh samples do not make
the repeatedly studied scenario family an independently invented test domain.
All previous failed results, including v101, remain failures under their own
predeclared protocols.

Evidence: [raw artifact](mmm-replication-v102.json),
[complete summary](mmm-replication-v102-summary.json),
[evaluator](../../research/replication-v102-summary.mjs).

## Matched confirmation results

Expected Brier is lower-is-better. Hidden simulator truth is used for scoring
only after every forecast exists; it does not enter fitting or routing.

| Full-view scenario and segment | Generic64 | Raw bank | Neutral gate | Comparative gate | Routed |
| --- | ---: | ---: | ---: | ---: | ---: |
| Parity4, all256 | 0.097948 | 0.078000 | 0.078001 | 0.076928 | 0.076933 |
| Parity4, late128 | 0.073204 | 0.053960 | 0.053960 | 0.053960 | 0.053960 |
| Majority to parity, late128 | 0.200315 | 0.191684 | 0.182148 | 0.175153 | 0.169594 |
| Parity to majority, late128 | 0.172970 | 0.179595 | 0.172960 | 0.173467 | 0.166177 |

Recovery improvements over generic64, in absolute Brier units:

| Phase | Direction | Mean gain | Paired interval |
| --- | --- | ---: | --- |
| Design | Majority to parity | 0.028939 | [0.023246,0.034632] |
| Design | Parity to majority | 0.008225 | [0.004172,0.012278] |
| Confirmation | Majority to parity | 0.030721 | [0.024830,0.036611] |
| Confirmation | Parity to majority | 0.006793 | [0.003398,0.010187] |

All means exceed0.005 and all lower bounds exceed zero. The criterion does NOT
require a lower bound above0.005. These are approximate paired z=3.5 intervals
over128 trajectories, not exact nonparametric coverage, anytime confidence
sequences or a family-wide guarantee across all historical research iterations.

Parity4 late expected accuracy is94.87%. The two switched cases have routed
late expected accuracy76.89% and77.82%, respectively. Passing recovery does not
mean retaining95% accuracy through a change or learning arbitrary new knowledge.

## Mechanism remains limited

The confirmation parity-to-majority block128..159 has routed neutral weight
0.276153 on average. Brier falls from comparative0.370293 to routed0.342311,
but both classification accuracies remain about51%. This is mostly attenuation
of stale confidence, not successful prediction of the new regime.

In160..191, generic32 has Brier0.103598 versus generic64's0.209255. Routing raises
its mean weight from0.020821 to0.030361, still small. Routed Brier is0.198361
versus comparative0.199538. At192 and224 generic32 is worse than generic64;
permanent promotion would be poorly motivated. The larger replication supports
the net finite benefit, not a claim that the remaining underweighting is solved.

## Null checks and verification

| Mode | Monitor | Rejected families / total | Wilson95% upper |
| --- | --- | ---: | ---: |
| Independent target tests | Neutral | 5/4096 | 0.2855% |
| Independent target tests | Comparative | 5/4096 | 0.2855% |
| Shared within version | Neutral | 1/4096 | 0.1382% |
| Shared within version | Comparative | 0/4096 | 0.0937% |

These finite upper bounds are below1%. The target-law calibration null remains
a substantive assumption, not a property established for the learned models.
Routing does not add a rejection test: its complete comparative state equals
the paired comparative control after EVERY outcome. The raw bank's cumulative
loss guarantee is not inherited by either the gated or routed law.

- Fresh generation passes:564.066s package time; complete learned/null replay
  passes:573.906s. Every record and tape reproduces against its manifest.
- Learned race smoke passes:3.582s. Routing/comparative/neutral/bank reference
  and lifecycle race checks pass:1.331s. Vet passes.
- Independent summary reproduction verifies21 source/protocol hashes, all3,072
  records, publication origins, block-to-whole/late score reconstruction,
  effective-weight normalization, rejection counts and raw-bank loss bounds.
- Preflight verifies9,216 unique learned RNG seeds and disjoint effective
  learned/null seed ranges.
- Artifact SHA256:b92fc9cb8cb78a86bc59f563b426a13ca70e696949ad115a314d120deaa97836.

## Performance and next boundary

The [full twelve-arm fixture](mmm-replication-v102-benchmarks.txt), measured
after replay and race tests finished, takes177.58-202.51ms per256-step stream
and allocates21.580-21.603MB in2058-2188 allocations. Apple M4,Go1.27.1,
darwin/arm64,GOMAXPROCS10; three single-iteration repetitions. It includes
eight refits, both views, all controls, scoring and hashing, not serving I/O.

The unchanged routing kernel retains its [v101 paired arithmetic measurement](mmm-routing-v101-results.md):
approximately1.22-1.24us per update, zero allocations, about8% more monitor
arithmetic than comparative-only in that fixture. No new production-latency
or cross-run speedup claim is made from this replication.

Next [return the routed law to adaptive observation](../../research/routed-observation-proposal.md),
with coherent acquisition/forecast/stopping semantics and explicit total costs.
Then revisit delayed-feedback and member-sharing integration. The
[scope audit](../../research/scope-audit-v102.md) preserves all seven requirements,
including real-agent tasks, external evidence validity and persistent serving.
No production, OpenClaw, database, commit, push or whitepaper change was made.
