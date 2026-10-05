# Same-graph search-effort diagnostic

Follow up the static recall failure without changing the graph between arms.
Two independent 6400-record builds, 768-dimensional hash vectors, no writes.
Each of 1024 self queries runs at ef=0 (default),100,200,400 on the SAME graph.
Rotate the four effort levels by query index to reduce fixed order bias. Record
every candidate, hit, duration and reference identity. Top-k remains10.
Default and explicit100 should agree; compare default against the existing
empty-delta Search adapter on every query. Record any disagreement, do not hide it.

Diagnostic success: an increased effort recovers missed self targets without
introducing misses elsewhere on these fixed queries. This is a discovery sample,
not held-out evidence of semantic recall or a selected production parameter.
No inferential claim from two identical deterministic fixture builds. Timing is
unloaded, not a replacement for growth/load experiments. All levels report costs.

No original artifact is overwritten. Use the frozen bulk-build overlay; hash the
new runner, diagnostic adapter, protocol and relevant prior adapter sources.
