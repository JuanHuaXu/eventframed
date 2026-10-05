# Bound-origin conflict nomination

Research-only slow-path report, not a packing bypass or an AP certificate.
Compare verified bound keys against the actual legacy grouping signals. When
bound keys differ but the legacy group/similarity would collapse the records,
nominate the pair for review. This is disagreement between grouping mechanisms,
not proof of semantic contradiction, independence or factual error.

Require exact payload bindings for both records, same tenant, unique nonempty
event IDs and no already-separated AP keys. Reject frontiers above200 and invalid
similarity thresholds. Return at most32 nominations, report truncation, and count
all checked pairs (at most19900). Do not mutate events, probability laws, AP keys
or packets. No persistent queue is wired by this helper.

Replay the same52 public repeat pairs: they should require no conflict review
because valid bound keys agree. Focused negative/diagnostic tests cover the
same-lineage operator distinction, unbound/tampered records, cross-tenant records,
AP separation, duplicate IDs and truncated output. Passing the nomination test
does not resolve the previously observed packing suppression.
