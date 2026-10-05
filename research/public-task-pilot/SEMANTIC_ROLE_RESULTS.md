# Paired semantic-role service replay

## Result: current temporal-priority safety falsified

The frozen12-question diagnostic completed24 actual Recall calls. Ordinary
retrieval found top1 supporting evidence in6/12 cases; temporal priority found
it in3/12. One positive-selection rescue was outweighed by four regressions.

| Role, two questions each | Ordinary correct | Priority correct |
| --- | ---: | ---: |
| Positive selection | 1/2 | 2/2 |
| Polar verification | 1/2 | 0/2 |
| Falsification | 1/2 | 0/2 |
| Contracted negation | 1/2 | 0/2 |
| Explicit negation | 1/2 | 1/2 |
| Quoted-claim explanation | 1/2 | 0/2 |

The four lost cases ask about launch evidence after2010 under non-selection
semantics. The actual launch in2004 supports answering no/refuting the claim,
or satisfies the negated condition. The parser instead demotes it below the
arrival in2014. Arrival counterparts already failed in ordinary retrieval, so
there is a baseline relation-retrieval gap in addition to the priority defect.

Forecast laws and numeric ranker output were identical across arms by record
identity; journals match. This is semantic ordering harm, not Bayesian-law drift
or persistence corruption. It falsifies a universal no-harm interpretation of
the earlier date-screen gains. The previous positive selection results remain
valid for their tested questions; they do not establish general query safety.

## Evidence boundaries

Facts reuse the previously verified public Rosetta launch and comet-arrival dates.
Questions are designed controls extending known counterexample families, not
untouched natural-user samples. The12 cases are dependent pairs over two facts;
do not attach population accuracy confidence intervals or count them as12
independent topics. All calls use fresh memory stores, local embedding-only
inference, no learned updates, no answer generation, and no oracle-fed ranking.

Artifacts: semantic-role-v1/{corpus,queries,oracle,results}.json. Protocol was
written before dispatch: SEMANTIC_ROLE_PROTOCOL.md. The runner reads the oracle
only after all predictions. The verifier checks source hashes, paired numeric
inputs, laws, journal matches, support IDs, and aggregate counts.

```sh
node research/public-task-pilot/check-semantic-role.mjs
```

## Rescue requirements

1. Distinguish selection predicates from propositions being verified/refuted.
   Unsupported intent must not trigger date-based evidence deletion or demotion.
2. Retrieve the requested event relation (launch vs arrival) independently of
   whether its date makes a proposition true. A polarity guard alone would
   restore control behavior but leave the six baseline failures unresolved.
3. Preserve the earlier selection improvements and measure entire-packet support
   recall alongside top1 and absent-answer behavior. Do not solve this by
   universally returning no candidates or disabling all interpretation.
4. Use separate frozen variants and fresh role/paraphrase families after design.
   Neither hand-annotated query plans nor a case-ID lookup counts as autonomous
   interpretation or a learned rescue.

The next candidate should separate query intent, target relation and predicate
truth in its explanation. Statistical claims still require broader outcome data.
All seven whole directions remain open; no production behavior changed.
