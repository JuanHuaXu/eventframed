# Synthetic Bidirectional Reranking Protocol

Date frozen: 2026-08-29

## Question

After receiving certified usefulness outcomes, can the implemented EventFrame
post-contract correction improve a bounded `pack_k=10` packet in both
directions: promote a useful candidate from the `recall_k=50` frontier and
demote a known-harmful candidate already inside the packet, without removing a
known-useful but superficially suspicious candidate?

## Runtime Boundary

The experiment executes the actual Go `Service` path for observation, recall,
selection certificates, Bayesian outcomes, posterior prediction, rank-delta
materialization, post-contract reranking, and packing. A deterministic synthetic
ranker substitutes only for the remote LibraVDB contract so baseline rank and
score margin can be controlled exactly. Residual application, graph scoring,
adaptive packing, and diversity packing are disabled to isolate Bayesian
reranking. The production Bayesian score weight remains `0.10`.

This is mechanism evidence. It does not measure remote contract latency,
embedding quality, language-model answer quality, or production prevalence.

## Data

- Each synthetic corpus contains 50 nominated candidates and uses
  `recall_k=50`, `pack_k=10`.
- Sixteen chronological full-stream outcomes train each designated target,
  harmful, or do-no-harm candidate before the policy is frozen for evaluation.
  Fillers remain at the declared prior so demotion tests distinguish known harm
  from merely unobserved irrelevance.
- A design block uses seed `29082901` and 50 bidirectional plus 30 retention
  cases. It may reveal implementation errors but must not tune policy weights,
  score margins, support, or confirmation thresholds.
- The untouched confirmation block uses seed `29082902` and contains:
  - 200 bidirectional cases: all 25 combinations of useful-target baseline rank
    `11, 15, 25, 40, 50` and harmful-candidate baseline rank `1, 3, 5, 7, 10`,
    repeated eight times with independently shuffled fillers.
  - 90 do-no-harm cases: a known-useful, suspiciously worded candidate at
    baseline rank `1`, `5`, or `10`, repeated 30 times.
  - 50 correction-envelope controls: a useful candidate at rank 50 with a
    top-10 score gap larger than the configured correction can bridge.

The bidirectional and retention cases use narrow but nonzero contract-score
margins inside the declared correction envelope. Envelope controls use a wide
margin and are expected not to promote; they test that the correction remains
bounded rather than manufacturing arbitrary wins.

## Arms

1. `pass_through`: the frozen contract order with no learned EventFrame
   posterior correction.
2. `eventframe_active`: the same contract order and candidates after certified
   full-stream Bayesian training, with contextual and hierarchical prediction
   enabled and Anti-Pigeon retaining authority over posterior identity.

The comparison is paired within every case. Confirmation generation occurs
after policy and thresholds are frozen.

## Endpoints

- **Promotion:** the useful target moves from baseline rank 11-50 into the
  packed top 10.
- **Demotion:** the designated harmful candidate moves from baseline top 10 out
  of the packed top 10.
- **Joint repair:** promotion and demotion both occur in the same case.
- **Do no harm:** a known-useful suspicious candidate initially in the top 10
  remains packed.
- **Packet precision:** fraction of packed candidates labeled useful.
- **Unnecessary churn:** active/pass membership changes not involving the
  designated target or harmful candidate.
- **Latency:** in-process active recall duration. This excludes remote I/O and
  is reported as mechanism cost, not production latency.

Wilson 95% intervals are reported for binary rates. Results are stratified by
baseline target rank and harmful-candidate rank.

## Frozen Criteria

The bidirectional reranking mechanism is supported on this generator only if:

- confirmation promotion rate is at least 80% and its Wilson lower 95% bound is
  at least 70%;
- confirmation harmful-demotion rate is at least 80% and its lower bound is at
  least 70%;
- joint-repair rate is at least 75%;
- do-no-harm retention is at least 99% and its Wilson lower bound is at least
  95%;
- active packet precision exceeds pass-through packet precision;
- mean unnecessary membership churn is at most 0.25 candidates per case;
- active in-process recall p99 is below 100 ms.

Failure of a criterion falsifies or narrows the corresponding implemented
mechanism claim. Failure to promote the wide-gap envelope controls is expected
and is not counted against promotion; promotion of more than 5% of those cases
fails the bounded-correction negative control.

## Non-Claims

Passing does not establish performance on natural conversations, unseen event
identities, a billion-event corpus, a live LibraVDB service, or downstream LLM
answers. Those require separate external-validity tests. Synthetic generation
code, complete case plans, raw paired ranks, and aggregate results are retained
for reproduction.
