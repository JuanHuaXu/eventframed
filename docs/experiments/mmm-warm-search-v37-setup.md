# V37 Setup Repair

2026-10-03. Initial auditor self-test exited1 before any row was validated:
the VM did not expose structuredClone, used by the sealed V36 checker. Expose
that standard function in the checker sandbox; no frozen V36 source, raw data,
runtime or gate change. No V37 normal trial had begun. This repeats the need
to test complete nested-checker globals before prospective freeze, not merely
syntax or extraction. Scope is this local isolated auditor, not global policy.
