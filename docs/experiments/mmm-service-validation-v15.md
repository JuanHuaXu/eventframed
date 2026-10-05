# Actual service admission validation v15

ValidateResearchAdmission checks a typed durable record against the actual
committed service journal: tenant/session/time/snapshot, query digest using the
same query encoding/model, exactly one candidate decision and its original
pre-residual baseline. It fetches the as-of event and recomputes research features
from the original query, then checks dependency compatibility again. A coherent
publication proof is used when available; ordinary stores require exact snapshot.

The new fixture comes from real Service.Recall and its frontier tap. Valid input
passes; altered features/baseline/event/journal/query/as-of/tenant/snapshot reject.
The original input rejects after an actual store policy change. Three targeted
race repetitions, one full service race run and service vet pass.

This validates one admission at a point in time. It does not reserve store state,
validate the complete learner history's temporal cutoff, authenticate feedback,
or atomically couple the ledger commit with service publication. A live bridge
must handle dependency changes during/after persistence and revalidate before
using restored predictions. Supplied query input is required because the journal
stores only its digest. Feature recomputation and optional embedding run on the
research path; no hot-path performance claim is made. Persistent-backend and
interleaving fixtures remain to add. No daemon configuration changed.
