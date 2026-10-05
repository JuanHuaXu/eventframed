# 2026-10-04 compiler-closure helper repair

Project-local error. The helper asserted that every GoFiles entry of a
repository-local package must itself be repository-local. Go's generated test
main legitimately resides in the build cache. The first helper stopped before
creating a component freeze, running a test, or measuring performance.

Record the generated input and its hash separately, preserve the full compiler
closure, and freeze/copy actual editable repository sources. Do not weaken the
scientific contract or treat an instrumentation failure as a model failure.
This is a local experiment-helper repair, not a global instruction change.
