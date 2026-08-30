# Contained OpenClaw E2E Clean-State Confirmation

Date frozen: 2026-08-29

## Question

Can the frozen LibraVDB plugin and the EventFrame pass-through and active paths
retrieve seeded facts for the production model when no query can reuse session,
workspace, or previously generated answer context?

## Arms

1. `libravdb-control`: the unmodified LibraVDB memory plugin.
2. `eventframe-pass`: EventFrame with contextual and hierarchical scoring and
   residual application disabled.
3. `eventframe-active`: EventFrame with contextual and hierarchical scoring and
   applicable certified residuals enabled.

Each arm uses a fresh remote tenant. The EventFrame arms also use separate fresh
daemon databases. `recall_k=50`, `pack_k=10`, and the memory token budget is
2,000 in both EventFrame arms.

## Isolation Controls

- Seed and query phases use different OpenClaw state directories and different
  workspaces in every arm.
- Every seed and query uses a unique session key. Query states contain no seed
  sessions before the query batch starts.
- Internal `session-memory` is disabled. Workspaces contain no `MEMORY.md` or
  `memory/` tree and only frozen neutral harness instructions.
- EventFrame query configurations set `capture=false`, so query prompts and
  answers cannot enter EventFrame or LibraVDB.
- The frozen LibraVDB plugin has no query-capture switch. Its nine queries are
  therefore dispatched concurrently. The arm is valid only if all nine
  provider prompts were compiled before the earliest query model completion.
- Query prompt text is scanned before execution. No positive query may contain
  its expected answer. The negative control's declared `UNKNOWN` fallback is
  the sole exception.
- Tenants, paths, state directories, workspace directories, session keys, and
  daemon databases are audited for uniqueness before scoring.

## Fixtures And Scoring

Fourteen records are seeded using one isolated session per record. Eight
positive queries and one absent-fact negative control are then run. A response
passes only after trimming whitespace and comparing case-insensitively to the
predeclared exact answer. The instruction-bearing Quartz record also fails if
the response follows its quoted `MALACHITE` instruction.

The primary endpoint is exact-answer accuracy. Secondary endpoints are target
survival in the packed context, target rank, active-versus-pass ordering,
provider-visible context, adapter recall latency, and daemon resources. This
small synthetic confirmation can validate wiring and isolation but cannot
establish population superiority.

## Validity Gates

- All query trajectories begin with zero prior messages.
- No query workspace or state contains seed text or persistent markdown memory.
- All provider-visible positive answers must be attributable to retrieved
  memory, not the query, workspace, or session history.
- Every successful EventFrame query has a matching recall trace and no observe
  trace during the query phase.
- EventFrame packets name the real LibraVDB nomination and ranking contracts.
- EventFrame adapter recall p99 is below 100 ms, including remote LibraVDB
  contract calls but excluding model generation.
- Failure of any isolation or concurrency gate invalidates the affected arm;
  its accuracy is reported descriptively but not used as confirmation evidence.

## Non-Claims

This run does not test corpus-scale storage, targets outside the 50-candidate
frontier, long-run Bayesian calibration, autonomous agency, robotics, or
population-level accuracy. It also does not estimate production tail latency
from nine concurrent synthetic queries.
