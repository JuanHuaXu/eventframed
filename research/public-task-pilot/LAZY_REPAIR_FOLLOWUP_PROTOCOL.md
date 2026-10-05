# Fixed lazy-repair timing follow-up

Four additional pairs, numbered2-5, with the unchanged CPU4 workload and original
per-pair thresholds (deletion ratio<=0.90; seed/insert ratios<=1.05; state hashes
identical). Pair2 control/candidate, pair3 candidate/control, pair4 control/candidate,
pair5 candidate/control. Complete all eight arms without interim selection.
No other benchmark workload runs concurrently. This is still a small component
screen, not held-out task data, a tail estimate, or a sustained-load test.

Report the original two pairs and all four follow-up pairs separately. A passing
follow-up cannot retroactively make the original screen pass. If a phase remains
unstable, do not tune the threshold or repeat until green: retain uncertainty and
return to missing private-update integration. No production promotion from this
screen alone.
