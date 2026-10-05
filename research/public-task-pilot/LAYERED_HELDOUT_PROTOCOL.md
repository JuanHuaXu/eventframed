# Held-out synthetic traversal probes

Use32 newly generated queries named `heldout-traversal-v1-0` through31, generated
by the existing deterministic768-dimensional vector helper. None is a corpus ID.
Use initial/final captured graphs at N800/6400, k10, ef100 and ef200, metric budget
20000. Reuse the exact same queries for both breadths and snapshots. Compare every
query to exhaustive scalar-cosine ranking. Report per-query hits and evaluations.

Diagnostic target: at least95% aggregate recall@10 in EACH graph snapshot for a
configuration. Do not average away one failed corpus/snapshot. No task/semantic
claim follows from random-vector recall. No latency comparison from this mixed
oracle/test workload; no threshold/query changes after execution. These queries
become exploratory data once results are inspected, not reusable untouched tests.
