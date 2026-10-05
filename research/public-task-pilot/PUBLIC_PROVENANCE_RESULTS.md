# Public-fact provenance boundary results

[Protocol](PUBLIC_PROVENANCE_PROTOCOL.md), [grouping results](public-provenance-results.json),
[default packet-selector follow-up](public-provenance-packing.json).

Used13 existing NASA facts, with unchanged mission/event/target/date values.
These are52 paired metadata transformations, not52 independent empirical
examples. No private chats, new source fetches, LLM calls or synthetic factual
claims were used. Local origin IDs stand in for a stable upstream source mapping;
they are not production source-event records or authenticated provenance.

## Measured behavior

| Repeat condition | Correlation recognized /13 | Repeats suppressed /13 | Total packet slots used |
| --- | ---: | ---: | ---: |
| Stable origin, changed fetch/run/session IDs | 13 | 13 | 13 |
| Missing origin, changed tool-call IDs | 0 | 0 | 26 |
| Stable origin, changed producer | 0 | 0 | 26 |
| Origin IDs change per fetch | 0 | 0 | 26 |

Grouping used epistemic.Describe/Correlated directly. The follow-up used the
actual packing.Select default policy, two candidates, pack/recall caps2,100
declared tokens each and1000-token capacity. No certified split buckets were
provided. Thus token shortage or adaptive expansion did not explain suppression.
The two records differ only in the declared retrieval metadata for each mode.

This is core-component/packing integration evidence, not CaptureTurn extraction,
full Recall, authentication, posterior updating or agent outcomes. Two occupied
slots do not prove two confidence updates or a production poisoning exploit.
It does show that the default occupancy cap cannot recognize these repeats when
the required stable lineage is absent or changed.

## Interpretation and patch boundary

The behavior matches the code contract: non-conversation lineage includes
producer plus source-event IDs, falling back to tool-call identity when source
IDs are absent. Different lineages are not classified as correlated. Therefore
the result is not sufficient authorization or evidence for a broad change that
merges different producers or trusts arbitrary URLs. Such a change could erase
genuinely independent observations or cross a trust boundary.

Classification: confirmed boundary behavior; upstream origin preservation and
cross-producer equivalence need investigation. No runtime fix is proposed as
confirmed yet. The concrete next experiment is an offline origin-equivalence
adapter with explicit trusted inputs and ambiguity rejection, followed by the
actual packing path. It must preserve distinct observations and Anti-Pigeon
separation; it must not declare equal text to be equal provenance.

Unlike the previous assumed80%-accurate audit, this observation is directly
measurable from recorded metadata and component outputs. It still does not
authenticate factual truth, estimate source reliability, or establish actual
continuous-learning gains.

## Verification

Both Go runners replayed byte-for-byte. Inputs and touched implementation files
are hashed in their outputs. Existing internal/epistemic tests pass. The packing
follow-up is explicitly subsequent to the grouping diagnostic, not a separately
predeclared confirmation. Only research runners and documentation were added;
the implementation and production environment were left unchanged.

Generation work in this public pilot remains at zero actual generation calls;
no approved new model/endpoint has been selected in this continuation. Other
research work is available. All seven whole directions remain open.
