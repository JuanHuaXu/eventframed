# Tadine fixed-clock diagnostic replay

The first current-contract run is retained as a **failed frozen gate**:
`current-contract-raw.json` and `current-contract-summary.json`. It completed
144 Recall calls and showed 57 paired top-1 gains, zero losses and p95 below
24 ms, but every case failed pre-rank-score and scored-law equality. The
runner used `time.Now()` separately at each arm's capture and Recall. On the
first case, the same source event's pre-rank probability was
0.8458726723181794 versus 0.8458727051221411; its scored Bernoulli law
moved by the same amount. The ranker is invoked after those laws are formed,
so arm-specific age is the leading confound, not proof of ranker influence.
The exact first runner is preserved as `runner-v1.go.txt` with SHA-256
`d49caa89bd2dbc4853c36897762c94f0e81b3d6ee1a3e70d85fb11a2b87497f5`.

Before a diagnostic replay, fix both arms' corpus capture time at
2026-10-01T20:20:00Z and their Recall as-of at 20:21:00Z. This is after the
public source snapshot was saved, and all source events precede October 1.
No question, answer, model, rank rule, packing rule or gate changes. Run the
same 144 calls into new exclusive-create raw/summary artifacts; apply the
same independent scorer. Falsifier: if pre-rank scores or scored laws still
differ, the wall-clock explanation is incomplete. Even if equality is
restored, the source and answers have now been consumed. The replay is a
diagnostic control, not a fresh confirmation or Goal 5 completion.
