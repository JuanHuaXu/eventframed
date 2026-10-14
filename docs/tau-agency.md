# Tau-Driven Automatic Turns

Contract 17 adds `POST /v1/agency/tau:plan`. An authenticated forecast producer
publishes a bounded marked-time/no-event law and an explicit utility policy.
The daemon chooses a time or abstains, signs the decision, and persists it in
the existing agency queue. The ordinary OpenClaw authority still checks consent,
capabilities, session scope, quiet hours, and its kill switch. Agency remains
disabled by default. This path never authorizes agent tools.

## Decision

For nominal time `s`, lead `l`, window `w`, and dispatch slack `j`, useful mass
is the target event's probability on `[s+l+j, s+l+w]`. This intersects the
useful windows of every allowed dispatch time. No-event and other marks do not
contribute. With declared TV allowance `r`, use `p_lower=max(0,p-r)` and
`utility=(benefit+false_cost)*p_lower-false_cost-turn_cost-1e-12`.
Schedule only when utility strictly exceeds the positive margin and the lower
probability meets its floor. Exact utility ties choose the earlier candidate.
Explicitly empty candidates abstain; an omitted candidate list uses the frozen
atom-aligned construction. Optional UTC quiet hours require both endpoints and
exclude any intersecting dispatch interval.

This bound is conditional on the declared error allowance and utility model.
An authenticated forecast is not automatically calibrated, truthful, or causal.
Missing/null numerical declarations are rejected, including TV allowance and
interruption costs; explicit zero is permitted where the contract permits it.

## Forecast Input

Supply `protocol_version`, `issuer_token`, `tenant_id`, `session_id`, `plan_id`,
`origin_event_id`, `base`, `forecast`, `policy`, `priority`, and `reason`.
The `base` and `forecast.base` fields bind the same original snapshot from
`GET /v1/health`. Direct laws also supply `evidence_ids` including the origin.
The origin frame's session must match the destination. Every cited frame must
have become available no later than the forecast's issue time.

The forecast declares `id`, `origin`, `issued_at`, `valid_until`, `horizon_ms`,
`points` (`mark`, `delay_ms`, `probability`), `no_event`, and `tv_radius`.
At most 32 distinct, ordered atoms are accepted. All masses, including
no-event, must sum to one within `1e-12`. An all-no-event law is valid.
Forecast IDs used for durable proposals cannot be rebound to another law or
snapshot within the tenant. Exact retries of issued proposals return the original durable
decision, including its current terminal status, rather than recomputing it.
Abstentions are not permanently journaled by this endpoint. Complete-stream
timing/calibration studies require a separate pre-action journal that records them.

The policy declares `target_kind`, `benefit`, `false_cost`, `turn_cost`,
`min_probability`, `margin`, `lead_ms`, `window_ms`, `slack_ms`, `min_delay_ms`,
and `cooldown_ms`. Optional fields are `candidate_delays_ms`, `quiet_start_utc`,
and `quiet_end_utc`. The reference horizon is at most seven days; minimum
dispatch slack is two seconds and minimum planning delay is one second.

Alternatively provide `episodes` and `bin_delays_ms`, omitting `evidence_ids`.
Each episode names a distinct `anchor_id`, `outcome_id`, and `no_event` flag.
One to fifteen closed episodes must be ordered by session/time and have
disjoint intervals. A no-event outcome must be an explicit `tau_no_event`
frame closing the full horizon; a missing response is insufficient. The last
bin edge closes the horizon. The daemon derives a Dirichlet-1 histogram from
stored occurrence/availability metadata, preserving the no-event category.
It is a working forecast; the producer must separately justify any error
allowance. Fit retries preserve their original issue time.

## Delivery And Reconciliation

Any non-agency runtime mutation conservatively invalidates a pending timing
forecast. Proposal/lease/status bookkeeping does not invalidate itself. Lease
and approval validate this invariant under the backend's write lock. There is
one reserved opportunity per tenant/session/target, plus the declared cooldown.

An uncertain scheduler result retains its execution handles even after lease
or proposal expiry. It blocks admission, claiming, and approval of another
opportunity for that target. Compensation reserves this state until cancellation
is acknowledged. `uncertain` cannot be replayed into approval. Clearing it
requires `decision: rejected` plus `cancelled_execution_refs` matching the
current complete handle set; a stale or ordinary rejection cannot clear it.
The authenticated `POST /v1/agency/proposals:get` endpoint supports reconciliation.

The adapter retries an exact approval before compensating for a lost response.
It refreshes its clock after cleanup and checks scheduler cancellation results.
A separate worker's approved handle is not removed by an older attempt's tag
cleanup. Database commits and scheduler effects still are not one transaction;
already executing work cannot be recalled, and daemon outages can prevent
immediate persistence of an uncertainty report.

The current OpenClaw scheduler accepts a requested time without enforcing a
model-start deadline. The adapter checks the handoff window, while the lab
records actual start lateness separately. A hard real-time guarantee requires
an execution-time expiry facility in the host. Do not infer it from one timely
smoke test or from the utility calculation.

## External OpenClaw Authority Worker

Some OpenClaw releases restrict `api.session.workflow.scheduleSessionTurn` to
bundled plugins. For these releases, `plugin/src/agency-worker.ts` supplies an
explicit operator client to the existing loopback gateway's public cron RPCs.
It uses configured CLI/config/state paths for normal gateway authentication,
does not discover credentials, and does not impersonate a bundled plugin.

Run the built worker with an absolute JSON configuration path:

```sh
node plugin/dist/agency-worker.js /absolute/operator-config.json
```

The file has `adapter` (the normal fully specified agency adapter configuration)
and `gateway` objects. `gateway` declares absolute `cliPath`, `configPath`,
`stateDir`, and an explicit loopback `url`. Optional `silent: true` suppresses
outbound announcements; an optional explicit `model` binds the ordinary
scheduled turn's model. Consent and session prefixes remain mandatory.
Stopping a worker prevents new handoffs; inspect/reconcile any job already
acknowledged by the scheduler.

The reproducible public-only lab harness is `plugin/lab/tau-lab.ts`. It requires
explicit `EVENTFRAME_LAB_CLI`, `EVENTFRAME_LAB_CONFIG`, `EVENTFRAME_LAB_STATE`,
and `EVENTFRAME_LAB_GATEWAY` environment variables. It creates its own daemon,
database, credentials, frames, and fresh session; uses one scheduled model turn;
records controls, timings, model receipts, and configuration equality; and stops
its daemon/worker. The existing gateway is reused. Private histories and
production configuration are not experiment material.
