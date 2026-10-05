# Published-LSN cosine conversion v2: nonunit counterexample

Date: 2026-10-02. Frozen [protocol](mmm-published-cosine-v2-protocol.md).
Switching SQL to `<=>` and using `1-distance` fixes the operator family,
but the frozen unrestricted score gate still **fails**. The stored angled
vector `(3,4,0,0)` produced SQL cosine distance 0 and converted similarity
1; ordinary Store Search returned similarity 0.600000024. Aligned,
orthogonal and opposite controls matched at 1, 0 and -1 respectively.
All five visible bodies were returned in the expected order, and the
future event remained excluded. The opt-in test retains this failure.

The local LibraVDB v1.6.13 SQL cosine utility computes `1-dot` assuming
pre-normalized input, clamping a negative distance to zero. Ordinary
collection Search normalizes vectors for its cosine index, so the same
nonunit stored vector can legitimately receive a different score. Both
built-in EventFrame embedders normalize their outputs, but the Store and
pre-embedded Observe path do not enforce that invariant. The
[unit-restricted test](mmm-published-unit-cosine-v3-results.md) does not erase
this counterexample. No unrestricted adapter or production change follows.

Reproduce the expected failure:

```sh
EVENTFRAME_RUN_PUBLISHED_COSINE_V2=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedCosineV2$' -count=1 -v -timeout 5m
```

SHA-256: test `cd17a4378b5cae2d0b6d1920fc0fd9c2330ec08890a8137b44f65967bded309b`;
protocol `3547dcddda2526528897478f714a1db381778e8eb8f43ec12e03a311c280950f`.
