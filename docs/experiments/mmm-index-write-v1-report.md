# Source index write-cost diagnostic

## Finding

The source-identity index adds row-processing cost in all six paired cells.
This is a diagnostic, not a valid optimization: removing it permits duplicate
sources under different ledger identities. The negative control demonstrates
two committed duplicate-source rows without the index versus atomic rejection
and zero rows with it.

| Batch | Trial | Additional rows time (ms) | Additional total time (ms) | Indexed/absent total |
| --- | --- | --- | --- | --- |
| 50 | 0 | .234 | .190 | 1.330 |
| 50 | 1 | .231 | .253 | 1.544 |
| 50 | 2 | .214 | .234 | 1.493 |
| 200 | 0 | .952 | .965 | 1.575 |
| 200 | 1 | .952 | .910 | 1.529 |
| 200 | 2 | .944 | .924 | 1.536 |

Every value is the difference or ratio of two cell means, each over 32 fresh
appends. The three trials rotate arm order. Commit differences vary in sign;
do not interpret phase differences as exact deterministic decomposition or
claim population confidence from these few trials.

## Method and verification

See the precollection [contract](mmm-index-write-v1-contract.md),
[raw artifact](mmm-index-write-v1.jsonl), and
[summary](mmm-index-write-v1-summary.json). Both arms use identical fixtures,
prepared statements, WAL and synchronous FULL. Each has a new isolated ledger.
All 48,000 original records were verified by sequence, key, kind, and payload
after reopen. This uses ordinary replay reads, not an absent indexed lookup.

Raw SHA-256:
`40d4b8e1f89b101a3094fd22fddf9c1ef2bbdae03c58e82b24af2052339dd625`.
The raw artifact captures the contract, Go module files, and ledger Go sources.
The independent postcollection summarizer validates captured and live source
hashes, unique cells, counts, and nonoverlapping phase clocks. Its second output
is byte-equal to the saved summary. The summarizer itself was written after
collection and is not claimed to be part of the captured precollection sources.

Boundary and prepared-parity tests passed under the race detector (1.406s);
ledger vet passed. Experiment collection passed in .93s (package 1.140s).
These are test runtimes, not serving-latency claims. No outcome prediction,
data selection, or quality threshold changed.

## Interpretation and next step

Confirmed: maintaining this required index accompanies approximately .21-.23ms
additional row work for 50 entries and .94-.95ms for 200 in this workload.
Still unknown: the separate contributions of JSON expression evaluation,
B-tree work, page layout, and WAL effects. The ablation removes all of these
index effects together. Its speed is not a promised bound for a valid replacement.

Next investigate a typed/materialized source representation while retaining
atomic source uniqueness, source-to-payload agreement, exact retries, rollback,
and recovery. Compare against the earlier envelope study rather than repeating
its failed read strategy. A passing write microbenchmark would still require
read, concurrency, durability, and end-to-end freshness/tail-latency tests.
Do not change the publication guard or weaken synchronous durability.

All seven goals remain open. Production, the whitepaper, and remotes unchanged.
