# V1 Preserved Partial Failure

Five standalone candidate mechanics tests pass ordinary/race after repairing
the checkpoint lock cycle. Both arms pass five selected database durability,
atomicity, rollback and reopen tests. These are component capabilities only.

Frozen cost runner stopped at its FOURTH command. Three commands completed all
128 durable insert/search operations: control/candidate32D initial256, and
control32D initial1024. Candidate32D initial1024 FAILED during bulk initial
transaction at e-00896 with arena exhausted. Do not claim the24-command study
completed, average only successful cells into a result, or skip initialization.

Source diagnosis: newFlatTxCollectionState sums raw ID lengths into KeyBytes,
but memory.IDMap uses a key arena aligned to8bytes.1024seven-byte keys require
8192bytes, while7168 were reserved; the896th put hits exactly that boundary.
The new DeltaIndex exposes this existing transaction scratch defect. Add a
separate sibling-path regression over ordinary/rename and aligned/unaligned
keys before repair. Preserve all original traces, source pins and witnesses.

Create fresh V2 copies for the aligned-key repair and rerun unchanged cells;
do not edit V1's frozen sources or overwrite V1's partial timing history.
All seven whole goals OPEN; production, private/sealed data and paper untouched.
