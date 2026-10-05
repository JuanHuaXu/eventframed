# Bounded guarded group scheduling v64

Frozen before measurement. v63 identifies guarded admission as an ingestion
waiting source. The lifecycle audit rules out simply moving Admit past the
guard: SourceOwner resolves/allocates identities; Durable stages actual Predict
and Record, advances pending state, marks itself stopped on uncertainty, then
atomically appends originals before acknowledging. Staging is not pure encoding.
Existing APIs require service validity through admission. A historical
certificate/two-phase alternative would need explicit replay and publication
semantics; it is not established by the existing batch/prepared-write contracts.

Test scheduling without weakening that invariant: four native writer slots in
all arms; native-off, resolved-source group cap4, cap2 and cap1. Three rotated
trials, twelve fresh cells. Same v62 fixture:192 reads at5ms,96 writes at10ms,
four reader lanes, one writer,50 visible events, recall50/pack10,queue64,
20ms guard-entry deadline. Drain only already-ready observations, never wait
to fill a batch. Lower caps increase transaction overhead but shorten each
guarded admission. Every observation remains eligible; no subsampling shortcut.

Record scheduled and inside-call tails, observation ages, completion/drop/expiry,
actual group sizes, admission/cleanup phases, all original/terminal integrity,
source snapshots and hashes in exclusive JSONL. Native default stays two slots;
no runtime, production, dependency or durability change. Test race/accounting
and vet before measurement; no concurrent task-started performance tests.

For each active cap separately, require all192 observations complete without
drop/expiry/error, age p95<=250ms, scheduled read and write p99<=1.10 times
same-trial native-four control in all three trials. Compare caps without
retuning these criteria. This is a finite exploratory screen, not simultaneous
statistical validation or permission to select and deploy a winning arm.
