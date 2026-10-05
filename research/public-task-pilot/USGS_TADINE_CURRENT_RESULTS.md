# Tadine current-contract result: failed gate, then diagnostic replay

The initially frozen current-contract run completed all 144 isolated service
calls, but **FAILED** its gate. Its saved [raw trace](usgs-tadine-transfer-v1/current-contract-raw.json)
and [summary](usgs-tadine-transfer-v1/current-contract-summary.json) show
57 paired top-1 gains, no losses and sub-24-ms p95, yet all 72 paired cases
failed pre-rank-score and scored-law equality. The runner used a fresh
`time.Now()` for each arm, changing the age of the same source records. This
was a harness confound, not evidence that ranking changed a forecast law.
The source and answer set became consumed on this run; its failure remains.

The separately declared [fixed-clock diagnostic](USGS_TADINE_CLOCK_DIAGNOSTIC.md)
used the same 24 events and 72 questions with one capture and as-of time.
Its [raw trace](usgs-tadine-transfer-v1/fixed-clock-raw.json) and
[summary](usgs-tadine-transfer-v1/fixed-clock-summary.json) restore exact
pre-rank-score and scored-law parity with zero invariant errors. On those
**consumed** cases, task-lexical top-1 is 69/72 versus focus 13/72, or 56
paired gains and zero losses. Pack survival is 72/72 versus focus 40/72.
The paired gain touches all 24 event IDs, but three wordings per event are
not independent samples. Isolated Recall p95 is 22.57 ms focus and 23.50 ms
task-lexical. This identifies the clock confound and demonstrates component
behavior; it is not untouched confirmation or generated-answer evidence.

The archived 2019 task-lexical overlay also could not compile against the
current service; [that interface gap](USGS_TADINE_REPRODUCTION_GAP.md) is a
separate failed reproduction, not an algorithm result. The current-contract
ranker is explicitly a different search-order policy. All seven whole goals
remain open; production and remotes are unchanged.
