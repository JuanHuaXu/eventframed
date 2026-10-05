# Actual Prefix Recovery: Import Success, Native Rank Failure

2026-10-04. Authoritative evidence:
`research/public-task-pilot/scifact-native-resume-v2/failure-audit.json`, its
36-source freeze, command logs and `nativeresume-v2` raw trace/native logs.
This is not an untouched quality result. All seven whole goals remain OPEN.

A new isolated read-only ListByMeta adapter verified EVERY one of the4250
acknowledged source records in an owned COPY of the failed v1 store BEFORE
inserting any further document. Exact ID/text/provenance/as-of metadata,
one-record cardinality and the canonical leading prefix are required. Original
failed stores remain hash-identical. The933remaining original documents were
then acknowledged:5183total, with no corpus reduction or raised RSS ceiling.
Full post-restart durable verification was not yet established at v2; the later
frontier-v3 run independently verifies ALL5183 stored rows after reopening.

The first FIT query returned200 SearchTextCollections candidates, but native
RankCandidates returned0 (`raw_union=200`, `top_k_output=0` in the native log).
The original complete-frontier gate rejected this. Client exit1/daemon exit0,
all owned processes terminal; zero completed FIT queries, zero calibration or
confirmation predictions, no evaluator label access. The old client rejected
before journaling the response bytes, so the native log and rejection are the
evidence, not a reconstructed full response trace. This failure stays failed.

Actual prefix RPC22.15s; remaining import RPC77.17s; owned process120.96s;
sampled daemon peak1460320KiB below the unchanged2GiB stop ceiling. This offline
restart bounds the observed import run, not a repair of the native memory leak,
dirty-journal panic or long-run durability. The original failed395.541s process
remains in cumulative acquisition accounting.

Race/vet/build passed; native command failed as above. Six actual race roots
ran3times. Independent source/command/copy/actual-prefix/terminal failure audit
passed. Preserve the initial Darwin fixture socket-path-length failure in
`scifact-native-resume-v2-preflight`; only the test's owned path was shortened.
Production sockets/configuration and native source code were not changed.
