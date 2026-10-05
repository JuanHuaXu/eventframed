# V39 Verification Recovery

2026-10-03. Re-freeze race/vet/13negative controls and benchmarks PASS. Design
collection completes448worlds/9408arms in393.01s, raw preserved unchanged.
Frozen audit reaches its12minute Go TEST PROCESS timeout while replaying the
adaptive control(see design-audit.log stack). No arithmetic mismatch reported
before timeout. This is incomplete verification, not a passed audit, failed
learner-quality gate or a measured daemon latency. Confirmation has not run.

One-second diagnostic sample of the SAME own auditor PID91111 is saved as
`research/continuing-v39-recovery/design-audit-sample.txt`. It reports3.1GiB
process footprint; this auditor retains full recorded world groups for summary,
unlike the2.4MB constructor benchmark. Symbols were unresolved by macOS sample;
Go addr2line maps selected addresses to independentOrientationV38,
orientationReference and auditContinuingV39. This is NOT complete CPU attribution
or evidence of model-memory cost. The timeout stack also shows control replay,
so both independent checks and reruns contribute off-line work.

Inspection also confirms a reserved-output collision in the ORIGINAL runners:
the audit writes `<split>-audit.json` for statistics, while the orchestration
would try to exclusive-create that SAME name for command metadata after a
successful audit. No successful statistical output was overwritten; the timeout
attempt's design-audit.json is clearly command metadata(code1), NOT statistics.

Local repair: keep both failed attempts/sources/raw bytes. Freeze a NEW completion
wrapper in `research/continuing-v39-completion`; symlink the original sealed
design raw rather than recollect it. Re-run the SAME frozen audit with a35minute
verification process allowance; distinguish command names from statistical
output paths. Then collect the still-untouched confirmation ONCE with the SAME
frozen model/collector/normal seeds, and run that same audit. Source hashes,
raw design hash and benchmark hash must match before and after. Record all
commands, timing failures and reports in exclusive files.

No serving/adoption/learner-cost threshold, model, rate, label, regime, prior,
parameter or auditor validation is changed. The400ms learner elapsed gate and
8MiB constructor limit remain. Longer offline verification is NOT an improved
computational result. New outcome synthesis and quality retuning on design
are prohibited; all seven whole goals remain OPEN. No production changes.

Follow-up engineering lesson remains project-local: separate command receipts
from owned result filenames, and distinguish verification-process budgets from
the runtime budgets being tested. A future STREAMING summary auditor should
retain compact per-world metrics rather than whole arm traces; that is not a
retroactive change to this frozen auditor or its proof obligations.
