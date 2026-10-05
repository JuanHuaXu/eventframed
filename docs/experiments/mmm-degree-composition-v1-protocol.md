# Fixed-noise degree composition v1

Exploratory consumed-tape screen, frozen before scoring. No production adoption.
Use all 2688 matching v120 and learned-degree-v1 records. The latter are
16-step projected fits, not converged BFGS fits. Preserve all original controls,
32-trajectory paired mean +/- 3.5 SE gates, .01 non-harm ceiling and .005
positive-lower-bound recovery threshold. These are not simultaneous guarantees.

Four candidates append one fixed-noise expert to the existing four experts:
64 and 32 labels, each with either the original .95 generic64 prior (remaining
.05 divided equally), or uniform five-expert prior. Transition probability stays
.001. Include the original four-expert mixer reconstruction as a mandatory
control. Uniform-prior four-expert controls must also be reported to distinguish
prior changes from added-expert gains. No threshold search or case-specific rules.

At clock t, admit only labels with origin < t, not missing, arrival <= t.
Each issued event contributes one transition; unknown emissions equal one.
Current forecast precedes its label. Store unresolved suffix and a settled
checkpoint; repeat filtering never counts evidence twice. Maximum delay31 gives
at most32 unresolved issues. Expired missing events contribute unit emissions.
Poison current, future and unavailable labels; compare a separate full-history
reconstruction. Q is evaluator-only. No fresh seed or prospective success claim.

Cost: additional eight capped fits per episode, each at most145 objective calls,
up to64 or32 labels. Prior component fixture for64 is4.147-4.192ms per fit;
not an end-to-end estimate. Mixer work O(32*5) per decision and O(32*5) storage;
offline source parsing/scoring is separate. Report measured replay time, not
production latency. Any failure stays in the record; all seven goals remain open.
