# Confirmed partial-transition time-fence defect

The new regression fails on the original isolated pool: child Full/Adaptive
consume a label at logical2 before the deliberately corrupted third ticket
rejects. The pool poisons, but its own clock remains1, so BeginEpoch(2,1)
incorrectly succeeds. No served forecast is emitted from the poisoned state.

The repair advances the pool's clock to the already validated operation time
when ANY unexpected child issue/resolve/cancel or advice-consistency failure
poisons it. Invalid caller input still rejects without mutating time/state.
This is not a rollback or shared transaction claim. No production code changes.
Reverse the four `p.poison, p.clock = true, at` assignments to the old
`p.poison = true` to reconstruct this frozen candidate source.

Adjacent tests retain valid explicit replacement, stale ticket rejection,
future-prefix equality, cancel conservation, private scored laws and all
independent matrix/path comparisons. Failed freeze/log/command remain intact.
