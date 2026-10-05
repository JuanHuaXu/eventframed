# Decision/publication support audit

The [read-only protocol](mmm-query-support-protocol.md) compares actual stored
supports, not merely prose. [Raw results](mmm-query-support-v1.json) retain all
2688 records and84 cells. There are38845 support checks against original sources
and both publication artifacts.
The [replay](mmm-query-support-v1-replay.json) is byte-identical; input and
script hashes are recorded in the result artifact.

Across1344 delayed records:

- Every actual noquery support differs from its63-label decision support.
- In595 records, publication restores a known label omitted by the reserved slot.
- In749 records, publication instead includes newly eligible evidence. This
  category includes frame160's label after the decision, not only arrivals at161.
- Of9277 candidate branches,4182 preserve the hypothetical decision support plus
  the queried label;5095 change it. There are278 naturally redundant purchases.
- Mean restored-known/newly-eligible/retired counts per record are
  .442708/.795387/.238095. These describe support, not forecast differences.

All1344 complete-delivery records have no paid candidates. Their publication
supports add one newly eligible label relative to the reserved decision support.

This proves an information-set difference, not that it causes the failed query
scores. The original protocol already disclosed publication refitting and the
reserved evidence slot. Classify this as a research-design distinction, not a
confirmed implementation bug. The next controlled comparison must separate:

1. Restoration of previously known evidence omitted by selection's reserved slot.
2. Newly eligible natural evidence, including the current frame's later label.
3. The hypothetical conditional model versus the actual capped update operation.

Keep the64-label noquery baseline for efficacy claims. A63-label frozen-support
diagnostic may isolate a mechanism but cannot replace that baseline and count
as a rescue. Action-aware expected utility must evaluate the actual update and
its information budget; a nonnegative hypothetical information gain need not
be an improvement under a different publication operator.

No fitter, selector, production configuration or paper changed in this audit.
All seven whole goals remain open.
