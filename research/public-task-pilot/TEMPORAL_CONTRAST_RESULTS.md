# Temporal contrast diagnostic results

Executed 24 isolated CaptureTurn/Recall runs with local nomic embeddings. No
generation, feedback updates, private data, LibraVDB backend or production changes.

| Arm | Positive top1 | Both members correct | Same top for opposite targets |
| --- | --- | --- | --- |
| Baseline | 6/10 | 1/5 pairs | 4/5 pairs |
| Frozen bounded sparse correction | 7/10 | 2/5 pairs | 3/5 pairs |

Both FAIL the predeclared ten-of-ten temporal adequacy diagnostic. The bounded
correction rescues Ceres's explicitly worded subsequent-classification query,
but still misses pre-decision Pluto and two proposal/adoption questions.
The Ceres before/after-asteroid pair succeeds in both arms: do not infer that
all temporal relations are lost or that the embedding representation has no
ordering information.

All supporting records remain in the eight-record pack. Maximum score delta
0.0047588454743361375; all saved forecast laws are exactly equal across arms.
Both absent questions still receive a top record; abstention is not implemented
here and these are not scored as successful answers.

Artifact: [raw results](temporal-contrast-v1/results.json).
Verification: `node research/public-task-pilot/check-temporal-contrast.mjs` checks
hashes, case coverage, record identity, oracle ranks, correction bound and laws.
The test evidence is within consumed public facts and newly tested phrasings,
not independent domain confirmation. All phrasings are now consumed.

Next lead: isolate the effect of including an explicitly contrasted alternative
in the query. This will not address the separate temporal-anchor failure and
must not be marketed as general temporal reasoning.
