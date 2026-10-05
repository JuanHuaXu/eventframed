# Abrupt-process deletion recovery

The new subprocess test uses disposable temporary databases and real libravdb
transactions. A child creates two authoritative records at revision1, builds the
tiny private writer, then signals a barrier either after staging Delete inside
the transaction or after WithTx returns success but before the writer's callback
returns. The parent waits for that exact barrier and kills only its child process.
There is no graceful DB Close or in-memory publication in the child.

Three race-instrumented runs of both cases passed in1.629s (six child terminations):

| Kill boundary | Recovered count | Target | Revision |
| --- | ---: | --- | ---: |
| Staged deletion, before commit | 2 | present with original vector | 1 |
| Commit succeeded, before acknowledgement/publication | 1 | absent | 2 |

Each parent reopens the database and verifies the survivor vector as well as
record count, target presence/content and revision metadata. The30-second timeout
is a guard, not the injection mechanism: timing out is a failure, not a pass.

## Interpretation

This extends clean-close recovery evidence to abrupt process termination at two
explicit transaction boundaries. It exercises the private writer's real callback
path, rather than deleting directly in an unrelated database test. Recovery does
not depend on the interrupted writer's unpublished graph snapshot.

It does not simulate machine power loss, torn sector writes, filesystem faults,
arbitrary instruction-point crashes or storage sync failures. OS caches remain
alive after the child is killed. The graph fixture is tiny and its adjacency is
empty; full ANN rebuild/enumeration and insertion remain unimplemented. All seven
whole research goals remain open. No production process, data, or config is used.
