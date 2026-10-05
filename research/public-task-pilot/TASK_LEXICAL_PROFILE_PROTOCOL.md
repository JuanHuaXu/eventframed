# Task lexical CPU diagnostic

Written before dispatch. Purpose: identify sampled CPU cost before choosing a
scheduling or computation rescue for the failed fixed-arrival screen.

Fresh memory store per arm;200 repeated public landing facts, hash32 embedder,
recall200,pack10,adaptive/diversity on. Run original then experimental packing,
one worker,128 sequential requests per arm with unique session journals. Use the
same task-lexical overlay as preceding tests. Two warmups precede profiling.

CPU profiling starts after seed/warmup, includes one explicit GC and successful
request/journal verification, and stops after128 calls. Service NS excludes
verification; sampled CPU includes it. The diagnostic intentionally has no
concurrent writers, queues or admission gate. Sampling overhead, accumulated
journals and arm ordering make this unsuitable for tail-latency qualification.

Save raw profiles exclusively beside the JSON; include profile and source
hashes. Inspect flat and cumulative CPU plus relevant annotated source. A large
profile entry is a lead, not permission to change semantics. Require behavioral
equivalence tests before any optimization and a fresh unprofiled load comparison
afterward. No expected speedup or success threshold is declared for profiling.
