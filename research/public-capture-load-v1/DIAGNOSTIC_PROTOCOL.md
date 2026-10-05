# Post-screen bottleneck investigation

Declared after the first measured failures and before diagnostic execution.
The frozen primary comparison is not modified or rerun. This additional run is
diagnostic, not an untouched confirmation, rescue or new finite timing verdict.

Use the same isolated control source and identical fixture at frontier 200.
Go CPU, mutex and block profiling are enabled for the whole process, including
1,000 initial captures, retrieval, concurrent writes and learner mechanics.
Retain failing test status; profiling overhead invalidates comparison to the
ordinary timing gates. Inspect cumulative ownership stacks and also CPU samples
focused on Recall. This does not measure each phase's exclusive wall duration.

Competing explanations: repeated input-mask work; per-frontier forecast and
certificate work; durable transactions or storage locks delaying unrelated
readers; ANN/metadata decoding. Source inspection shows Put holds event/write
locks across its transaction, and journal persistence takes a read lock against
mutators. That is a candidate mechanism, not proof of exclusive attribution.

A wait-heavy profile dominated by store/transaction lock stacks, with little
Recall CPU under extraction, would refute mask scans as the main rescue target
for this workload. A CPU-heavy frame/regex or per-frontier profile would instead
favor those explanations. Do not remove durable writes, snapshot validation,
availability filters or scored candidates to manufacture a passing result.

The profile uses only the public DESIGN replica timing fixture and temporary
stores. No production daemon, private corpus or real agent labels are accessed.
