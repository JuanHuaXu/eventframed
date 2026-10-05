# Notification comparison v8: bounded gain, one notified failure

All36 arms executed on the same current worker with rotated off/notified/polling
order. Raw artifact: mmm-notification-comparison-v8.jsonl. Test exited1; failures
are retained. Independent audit checked source hashes/current files,36 distinct
arms,192 reads/96 writes each and all queue/label/age accounting. No errors or
worker failures; every admitted label completed and writes overlapped reads.

Queue16 notified admissions were163/166/168/170/172/169 versus polling
154/155/154/148/151/150. Paired gains9/11/14/22/21/19 average16 frontiers.
Notified age p95 was92.410-98.324ms versus100.323-115.012ms polling.
However notified trial0 serving-p99 ratio was1.1975, exceeding1.10. Thus queue16
notification does NOT pass every cell, despite all its admission/delay passes.
Polling failed admission in trials3/4/5.

Queue64 notification passed all six cells, admitting192 each with age p95
172.073-191.733ms. Polling also admitted192 each but age p95 was219.420-269.606ms,
failing250ms in trials0/4. Both modes passed serving-p99 in all queue64 trials.

This supports a finite queue64 notification rescue with same-run controls and
unchanged-worker repeat trials following v7. It does not establish population
tail guarantees, nor erase the queue16 latency failure. Production deployment,
mixed mutations, durable recovery, realistic arrival patterns and actual learning
quality are not validated. No production setting was changed or pushed.

Next direction6 work should extend workload realism and lifecycle coverage, not
keep tuning this fixture. Successful posterior/snap/agency paths and durable
feedback/replay remain concrete integration gaps; the broader real-task research
directions also remain open. Preserve queue16 as a recorded partial failure.
