# Supplemental Functional Checks

Frozen AFTER ordinary pilot completion and BEFORE these additional commands.
No ordinary timing or functional gate changes. These are further DESIGN audits.

Run the unchanged complete public load fixture on the repaired candidate at
frontier 200 under race detection, GOMAXPROCS=10, Go timeout 4m and the original
internal context deadline. Require all 64 recalls/worker labels, 256 captures,
128 durable ledger rows, live completions and same-epoch replay. If instrumented
timing exceeds the frozen screens, preserve the nonzero exit and allow a
FUNCTIONAL-only conclusion solely when the complete functional trace is true,
there is no race/panic/skip and ALL test errors are the exact two known timing
screen errors. Do not call that command a passing timing test. This covers one
candidate/200 fixture, not every arm/frontier or production workload.

Add an independent missing-child negative control in the temporary library copy:
create a four-shard collection in a disposable test database, close/reopen,
delete one physical child through the test's engine handle, then assert logical
discovery still nominates the parent but loading and ensuring it fail closed.
The remaining physical collection-name set must not change during those calls.
Run normally and with race detection. No change to the six frozen source files
or the original four regression fixtures; pin the extra test separately.

These checks cannot establish process-kill recovery, full retention accounting,
large-corpus latency, calibrated confidence or untouched agent task improvement.
