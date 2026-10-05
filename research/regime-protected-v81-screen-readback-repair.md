# Readback-only local repair

First readback failed at line27: `TypeError: f.Packets[at] is not iterable`.
Frozen Go fixture correctly encodes a nil/empty packet slice as JSON null;
the JS auditor wrongly required every clock slot to contain an iterable array.
Preserved original reader: `regime-protected-v81-screen-readback-failed-null.mjs`.
Only readback handling of null empty slots is repaired. No frozen learner,
harness, protocol, data, acceptance rule or results were edited or rerun.
Structural per-packet/count/arrival/uniqueness and all48-case checks remain.
This is not a scientific negative result or a learner regression. No durable
global instructions, skills or tracked .learnings file were changed.

Second readback failed on exact cross-language metadata equality: Go emitted
0.34 while JS's algebraically equal expression returned0.33999999999999997.
Preserved reader: `regime-protected-v81-screen-readback-failed-float.mjs`.
The independent formula check now records maximum metadata error and uses
2e-15 tolerance for floating arithmetic, rather than bitwise Go/JS equality.
This is distinct from the unchanged2e-14 loss audit and frozen.01 harm rule;
integer schedules, packet values, received journals, counts and hashes remain
exact. No scientific threshold or experiment was changed. The last-bit difference
is consistent with arithmetic evaluation differences; no compiler-causality
claim is established by this reader repair.
