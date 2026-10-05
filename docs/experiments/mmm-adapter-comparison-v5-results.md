# Same-run adapter comparison v5: incomplete rescue

Executed all18 frozen arms: six rotated trials of off, ingestion-only and
general adapters, each192 requests. Raw artifact: mmm-adapter-comparison-v5.jsonl.
The test exited1 because general-adapter trials1 and5 admitted153 frontiers,
below the required154. Ingestion-only passed all six finite cells.

| Trial | Ingestion admissions | General admissions | General minus ingestion |
| --- | --- | --- | --- |
| 0 | 155 | 156 | 1 |
| 1 | 157 | 153 | -4 |
| 2 | 159 | 157 | -2 |
| 3 | 155 | 157 | 2 |
| 4 | 155 | 157 | 2 |
| 5 | 155 | 153 | -2 |

Mean paired admission difference is -0.5 frontier. This descriptive comparison
does not establish statistical equivalence, nor prove an adapter regression.
Both are close to the admission threshold, with little throughput headroom.
All enabled serving-p99 ratios passed (general .7172-.9294; ingestion
.7308-1.0265). Completion-age p95 was102.378-110.959ms across enabled arms,
below250ms. All arms had write/read overlap and zero errors; every admitted
frontier produced50 completed labels, with no worker failures.

Independent artifact audit verified18 unique arms, all embedded source hashes
against current files,192 read and96 write measurements per arm, queue accounting
and label/age conservation. Prior failed artifacts remain unchanged.

Next: instrument the general adapter consumer's admission, feedback submission,
model work and completion wait. The harness sleeps1ms while polling completion
once per frontier; measure this contribution rather than hiding it or assuming
it dominates. Event-driven completion notification is a possible engineering
lead, not a demonstrated fix. Do not loosen admission or delay gates.

These repeated fixture labels test load only. Real-task learning, mixed mutation
safety and durable recovery remain unvalidated. No deployment or push occurred.
