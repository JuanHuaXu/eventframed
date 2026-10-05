# V63 next lead: one bank-owned observation journal

PROPOSAL ONLY. V62 connects the selector to real joint forecasts, but still
allocates9,988,752bytes at200members,1,600,144above8MiB. The150-member gate passes;
do not shrink the intended bounded frontier merely to make the200-member failure
disappear. No broad quality or loaded freshness result is established.

## Root-cause and boundary

V62 shares static model data but retains three equivalent raw receipt journals
and global-position index arrays. Their actual first/second flags, outcomes,
member/ordinal/position and timestamps always follow the same bank transaction.
Only retained support and parameter/posterior state differ by window. This is
confirmed structural duplication, not proof that every journal field is redundant:
child clean/measurement forecast metadata currently differs across windows.

A canonical journal could own common observed-event fields and the BANK-issued
forecast metadata. Each child keeps independent posterior/count/support state,
window, issue counts, pending counters and epoch. Individual expert issue kernels
remain in the selector grading ledger. Raw forecast bookkeeping must never
substitute for a child's actual conditional kernel or erase issue-time evidence.

Ownership must move explicitly to the bank, not simply alias three freely
mutable public Model objects. Child models remain unexposed by the public bank
API. Their standalone mutating APIs must reject bank-owned instances, including
Issue,Resolve,Cancel,RequestAudit and BeginEpoch; readonly hypothetical operations
must not commit. Standalone models retain their existing unshared behavior.
One bank transaction validates/prepares EVERYdependent update, then publishes
shared receipt state once. No child can evict or acknowledge it independently.

Per-window expiry removes only that child's statistical factor. A shared first
or paired receipt remains historical grading evidence, never reinserted into an
expired model merely because another window retains it. A failed child/selector
cannot modify the common journal, any posterior, clock, owner or pending flag.
Global issued-position indexes may be common because all child issue sequences
are synchronized, but member counters stay independent unless separately proved.

## Falsifiers and tests before adoption

Compare complete externally visible forecasts, joint kernels, requests, receipts,
missingness, delayed expiry, epochs and queries against V62 on identical evidence.
Keep independent top-down reconstruction and expert-path checks. Add attempts
to mutate a bank-owned child via every standalone API, cross-bank/old-epoch
tickets, failure at each child, shared-journal corruption and non-vacuous future
forks. Verify outputs are unchanged while ownership/aliasing is intentionally
different. Do not rewrite V62's frozen tests or claim its distinct-journal
ownership checks establish the new shared-journal invariant.

Measure actual construction at150AND200members versus V62 and unshared controls,
including selector state, temporary allocations and any replacement metadata.
Sharing should save duplicated trial/index storage, but byte estimates alone
are not a pass. Keep8MiB/400ms and all scientific gates unchanged. If ownership
or metadata separation requires replacement storage, charge it rather than
omitting it. No concurrent or durable atomicity theorem is inherited.

After this storage equivalence study, wire and test acquisition against the
ACTUAL scored mixture. Existing per-window prediction-value quantities cannot
be called mixture values without accounting for selector motion and delayed
grading. The working window selector is not one ordinary stationary Bayesian
posterior; a tower-identity/proper-risk reduction must be proved or explicitly
replaced by a tested heuristic. Equal requests remain unequal TOTAL cost.

The full diagnostic and independent replay, robust quality/recovery, externally
valid useful Anti-Pigeon decisions, untouched labeled agent utility, loaded
serving/freshness and cost-matched observer comparisons remain required. No
confirmatory labels should be opened to select storage or observer parameters.
