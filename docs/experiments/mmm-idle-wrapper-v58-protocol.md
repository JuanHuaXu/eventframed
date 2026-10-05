# Idle publication wrapper diagnostic v58

Frozen before measurement. v57's5ms read/10ms write offered schedule shifted
scheduled read p99 from roughly600ms off to30ms in every active arm, while
scheduled writer latency increased. Wrapper effects and active guarded work
were confounded. Test the wrapper alone before changing lock ownership.

No runtime changes. At the5/10ms schedule only, rotate five arms over three
trials: unwrapped off; idle wrapped off; raw durable; combined source; resolved
source. Fifteen cells. The idle arm wraps the same backend after fixture setup
but leaves ResearchFrontier disabled, starts no consumer, opens no learner log,
and must have zero attempts/admissions/terminals/phases/drops. All served recall
and observe work remains real. The other four arms retain v57 behavior.

Keep192 offered reads,four reader lanes,96 offered writes,one writer lane,
due/start/end timing,50-event overlap,K50/pack10,queue64,group<=4,20ms guard entry,
FULL durability,fixed as-of,future writes and parent context. Drain each fresh
isolated public cell before starting another. No labels/fitting/private data or
production. Test idle accounting and scheduled timing under race before timing.
Capture source/hash snapshots and all raw timings in exclusive JSONL.

Primary mechanism check, not a deployment screen: does idle wrapping alone
halve scheduled read p99 relative to same-trial unwrapped off in all three
trials? Report writer changes and absolute timings regardless. If idle remains
within10% of off while active reads still differ greatly, evidence rejects the
wrapper-alone explanation. Other results are partial/inconclusive; do not infer
that this one contrast proves all interactions or general scheduler behavior.

Keep the original resolved-cell checks visible: age p95<=250ms,192 completions,
scheduled read and write p99<=1.10x unwrapped off. The idle control cannot replace
the user-facing off baseline to manufacture a non-harm pass. Focusing on the
largest unexplained effect is diagnostic, not a reduction of the full rate or
seven-direction goals. Retain v55/v57 failures. No defaults,push or deployment.
