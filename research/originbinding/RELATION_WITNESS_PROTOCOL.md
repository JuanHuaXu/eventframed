# Research-only relation-witness packet protocol

Status: frozen before implementation and tests. This does not change serving.

## Question

Can a payload-bound, capture-supplied relation identity resolve the exact-key
ambiguity without counting formatting repeats twice or changing unbound packet
behavior? A relation identity is an externally asserted proposition identifier,
not a string hash and not proof of independent sources or truth.

## Scope and decision order

Test the default fixed-order, non-adaptive, non-diversity packet selector only.
At most 256 unique event identities may be offered. Reject other selector modes
instead of silently claiming parity. Each relation witness carries the existing
full-event binding, its origin and a nonempty declared relation ID. The registry
is immutable after construction. A witness is usable only when its event payload
and tenant still match the bound digest and the event is an external observation.

For each candidate/prior pair, apply these checks in order:

1. Different certified `ap:` buckets remain separate.
2. The same exact bound evidence key remains correlated, even if relation IDs
   conflict; a misdeclared ID cannot duplicate an identical claim.
3. Two verified witnesses with the same origin and provenance lineage, neither
   marked as an explicitly distinct observed occurrence, are correlated when
   their relation IDs agree and separate when they differ.
4. In every other case, use the existing declared-group and epistemic fallback.

"Separate" here means retain both records within the packet cap, not treat
them as independent corroboration for any posterior. The witness is supplied
by a fixture controller in this test; a real capture authority is outside scope.

## Frozen falsifiers

- A spacing-only repeat with the same relation ID occupies one slot.
- Opposite operators with distinct relation IDs occupy two slots.
- Identical bound evidence keys cannot be split by inconsistent relation IDs.
- Missing, tampered, different-origin, or different-lineage witnesses retain
  legacy fallback behavior; no unbound input gains independence.
- Certified Anti-Pigeon split and explicitly distinct observed occurrences stay
  separate regardless of relation ID.
- With no usable relation pair, selected IDs, token count and suppression count
  match `packing.Select` for fixed-order inputs, including budget and cap cases.
- Invalid registry entries and unsupported selector modes fail closed.

Run race tests, vet and a bounded allocation/time benchmark against the
existing fixed selector. The benchmark is a component cost only, not agent or
daemon serving latency. No quality claim, production promotion or whitepaper
change follows from passing this fixture.

## Post-freeze input-hardening amendment

After the first benchmark and correctness pass, the bug hunt identified empty
tenant/event identities, negative token estimates, and overflow-prone budget
addition as missing input guards. These are added as fail-closed preconditions
and tested separately. They do not change the frozen relation-decision cases or
turn the already observed component results into untouched confirmation.
