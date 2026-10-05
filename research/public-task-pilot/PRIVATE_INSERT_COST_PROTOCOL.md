# Private insertion preparation cost

Diagnostic, not a sustained-load acceptance gate. Freeze four existing captured
cases: new-0 and new-7 at N=800 and N=6400. Three repetitions, one-second Go
benchmark duration each, CPU=4. No other benchmark runs concurrently.

Timed boundary includes traversal, selection, connection/backlink edits, all
intermediate immutable roots, and final summary preparation. Source parsing,
vector generation and initial graph construction are excluded. Each iteration
uses the same original snapshot; it does not grow the corpus or hold old readers.
Persistence, queueing, publication, ID allocation and concurrent retrieval are
not measured. Report ns/op, bytes/op, allocations, query evaluations and pairs.

Do not compare this component time with end-to-end service latency or claim the
long partial-tail load failure is rescued. Use the measured cost to decide how
to integrate and what to instrument in the durable serving test.
