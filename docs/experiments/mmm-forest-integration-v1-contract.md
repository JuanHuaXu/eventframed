# Frozen adaptive forest integration

Fresh cohort seeds2026092221/22, ten v83 cases,16 independent4096-label base
fits per case/cohort,512 frames each;320streams. Same retained-subset control,
gate, selector,64audit-label training window and fit schedule. Candidate input
weights from frozen heldout forest (half train/half validation, no refit).
No outcome label enters input-model selection. Same outcome learner uses all64.

Control vs forest fixed to control observations vs forest own observer.
Primary own-observer changed-case gain>=.005 and paired lower(mean-3.5SE)>0;
all full/post non-harm lower gain>=-.01. Fixed-arm result cannot substitute
for primary failure. Record costs, split/fit parity, source snapshots. No
parameter sweep or claimed higher-order input solution. Preserve all failures.
Adapter must exactly match existing component before collection. Test-only
integration plus isolated researchinput package; no production wiring.
