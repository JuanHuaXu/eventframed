# Bound-origin conflicts become review nominations

[Protocol](ORIGIN_CONFLICT_PROTOCOL.md), [public replay](public-provenance-review.json),
[research helper](../originbinding/conflicts.go), [tests](../originbinding/conflicts_test.go).

Added a bounded offline nomination pass, not a change to packing or trust
authority. A pair is nominated only when both exact payload bindings verify,
their bound keys differ, and the legacy grouping signal would correlate them.
This flags inconsistent grouping evidence; it does not certify independence,
contradiction, falsity or permission to create an AP split.

## Results

The same-lineage `A > B` / `A < B` diagnostic produces one review nomination.
All52 paired public repeats produce zero nominations and preserve the preceding
packing outputs exactly: one duplicate suppressed per pair. A separate replay
checks both byte equality and all52 prior packing records.

Nine focused test functions pass under the race detector. Tests cover preserved
AP splits, unknown/tampered bindings, tenant boundaries, registry identity rules,
distinct declared occurrences and input immutability. The older diagnostic that
records the still-present packing collapse remains explicitly labeled a known
limitation, not a successful separation test.

The scan accepts at most200 candidates and returns at most32 pairs. A12-record
stress fixture checks all66 pairs, returns32 nominations and reports truncation.
Oversized frontiers, duplicate/empty IDs and invalid thresholds reject. Capped
output is not represented as complete review coverage. Full-event hashing and
pairwise comparison remain slow-path research work; no latency claim is made.

## Unfinished integration

The original packing suppression remains unchanged and unresolved. This helper
returns nominations; it does not enqueue durable work, evaluate the differing
claims, authenticate registry construction or mint AP certificates. A reviewed
decision and a legitimate integration contract are still required before packet
behavior can change. Calling the nomination layer a complete rescue would be
incorrect. The next meaningful test must connect a nomination to an externally
supported review outcome while retaining independent-evidence accounting.

No production or paper changes. All seven whole research directions remain open.
