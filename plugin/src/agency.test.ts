import assert from "node:assert/strict";
import { createHash, generateKeyPairSync, sign } from "node:crypto";
import test from "node:test";
import type { OpenClawPluginApi } from "openclaw/plugin-sdk/plugin-entry";
import { evaluateAgencyProposal, processAgencyClaim, verifyAgencyProposal } from "./agency.js";
import type { AdapterConfig, AgencyProposal, AgencyProposalRecord } from "./types.js";

const proposal: AgencyProposal = {
  id: "proposal-1",
  tenant_id: "tenant-a",
  session_id: "openclaw:session-a",
  action: "notify",
  reason: "A high-value follow-up became timely.",
  evidence_ids: ["event-a"],
  expected_utility: 0.8,
  priority: 0.7,
  required_capability: "eventframe.agency.notify",
  not_before: "2026-08-28T21:00:00Z",
  expires_at: "2026-08-29T12:00:00Z",
  idempotency_key: "proposal-1",
  causal_chain_id: "chain-1",
  causal_chain_depth: 0,
  created_at: "2026-08-28T20:00:00Z",
  contract_version: 9,
};

function signedRecord(value: AgencyProposal): { record: AgencyProposalRecord; publicKey: string } {
  const keys = generateKeyPairSync("ed25519");
  const publicDer = keys.publicKey.export({ format: "der", type: "spki" });
  const rawPublic = publicDer.subarray(publicDer.length - 32);
  const payload = Buffer.from(JSON.stringify(value));
  return {
    publicKey: rawPublic.toString("base64url"),
    record: {
      proposal: value,
      signed: {
        payload: payload.toString("base64url"),
        signature: sign(null, payload, keys.privateKey).toString("base64url"),
        key_id: createHash("sha256").update(rawPublic).digest("hex").slice(0, 24),
      },
      status: "claimed",
      claimed_by: "test-authority",
      lease_until: "2026-08-28T22:00:00Z",
    },
  };
}

function authorityConfig(): AdapterConfig {
  return {
    socketPath: "/tmp/eventframed.sock",
    tenantId: "tenant-a",
    recallK: 50,
    packK: 10,
    tokenBudget: 2_000,
    capture: true,
    agencyEnabled: true,
    agencyKillSwitch: false,
    agencyPublicKeyPath: "/tmp/agency.pub",
    agencyAuthorityTokenPath: "/tmp/authority.token",
    agencyCapabilities: ["eventframe.agency.notify"],
    agencyConsentActions: ["notify"],
    agencyConsumerId: "test-authority",
    agencyPollIntervalMs: 5_000,
    agencyMaxClaims: 10,
    agencyMaxChainDepth: 3,
    agencyQuietHoursStartUtc: 22,
    agencyQuietHoursEndUtc: 7,
    agencyCriticalThreshold: 0.9,
    agencyAllowedSessionPrefixes: ["openclaw:"],
  };
}

test("verifies the exact signed proposal and rejects tampering", () => {
  const fixture = signedRecord(proposal);
  assert.deepEqual(verifyAgencyProposal(fixture.record, fixture.publicKey), proposal);
  fixture.record.signed.signature = Buffer.alloc(64).toString("base64url");
  assert.throws(() => verifyAgencyProposal(fixture.record, fixture.publicKey), /signature is invalid/);
});

test("accepts a still-valid contract-6 proposal after contract-7 queue migration", () => {
  const legacy = { ...proposal, contract_version: 6 };
  const fixture = signedRecord(legacy);
  assert.equal(verifyAgencyProposal(fixture.record, fixture.publicKey).contract_version, 6);
});

test("accepts current contract 16 and 17 while rejecting unknown future contracts", () => {
  for (const version of [16, 17]) {
    const fixture = signedRecord({ ...proposal, contract_version: version });
    assert.equal(verifyAgencyProposal(fixture.record, fixture.publicKey).contract_version, version);
  }
  const future = signedRecord({ ...proposal, contract_version: 18 });
  assert.throws(() => verifyAgencyProposal(future.record, future.publicKey), /authority contract/);
});

function tauProposal(): AgencyProposal {
  return { ...proposal, action: "schedule", required_capability: "eventframe.agency.schedule", contract_version: 17,
    scheduled_for: proposal.not_before, expires_at: "2026-08-28T21:00:10Z",
    tau: { plan_id: "plan", target_kind: "complete", forecast_id: "law", forecast_digest: "a".repeat(64), policy_digest: "b".repeat(64), latest_dispatch: "2026-08-28T21:00:10Z", lower_probability: .9, lower_utility: .8, cooldown_ms: 60000,
      base: { runtime_version: 1, policy_version: 1, contract_version: 17, graph_version: 1, posterior_version: 1, residual_version: 1, abstraction_version: 1, agency_version: 1, evidence_epoch: 1 } } };
}

test("tau dispatch honors signed slack, quiet hours, and binding", () => {
  const p = tauProposal();
  const config = { ...authorityConfig(), agencyCapabilities: ["eventframe.agency.schedule"], agencyConsentActions: ["schedule" as const] };
  const fixture = signedRecord(p);
  assert.deepEqual(verifyAgencyProposal(fixture.record, fixture.publicKey), p);
  const due = evaluateAgencyProposal(p, config, new Date("2026-08-28T21:00:03Z"));
  assert.equal(due.deliverAt?.toISOString(), "2026-08-28T21:00:04.000Z");
  assert.equal(evaluateAgencyProposal(p, config, new Date("2026-08-28T21:00:09Z")).allowed, false);
  assert.equal(evaluateAgencyProposal(p, { ...config, agencyQuietHoursStartUtc: 21, agencyQuietHoursEndUtc: 22 }, new Date("2026-08-28T21:00:03Z")).allowed, false);
  const malformed = signedRecord({ ...p, tau: { ...p.tau!, lower_utility: .4 } });
  assert.throws(() => verifyAgencyProposal(malformed.record, malformed.publicKey), /tau binding/);
});

test("requires consent, capability, session scope, and an inactive kill switch", () => {
  const config = authorityConfig();
  const now = new Date("2026-08-28T21:30:00Z");
  assert.equal(evaluateAgencyProposal(proposal, config, now).allowed, true);
  assert.equal(evaluateAgencyProposal(proposal, { ...config, agencyKillSwitch: true }, now).allowed, false);
  assert.equal(evaluateAgencyProposal(proposal, { ...config, agencyConsentActions: [] }, now).reason, "action lacks explicit consent");
  assert.equal(evaluateAgencyProposal(proposal, { ...config, agencyCapabilities: [] }, now).reason, "required capability is not granted");
  assert.equal(evaluateAgencyProposal(proposal, { ...config, agencyAllowedSessionPrefixes: ["different:"] }, now).reason, "session is outside the consent scope");
  assert.equal(evaluateAgencyProposal({ ...proposal, tenant_id: "tenant-b" }, config, now).reason, "proposal tenant is outside the authority scope");
  assert.equal(
    evaluateAgencyProposal({ ...proposal }, { ...config, agencyQuietHoursEndUtc: undefined }, now).reason,
    "quiet-hours policy is incomplete",
  );
});

test("rejects duplicate evidence identifiers in a signed payload", () => {
  const duplicated = { ...proposal, evidence_ids: ["event-a", "event-a"] };
  const fixture = signedRecord(duplicated);
  assert.throws(() => verifyAgencyProposal(fixture.record, fixture.publicKey), /evidence ids are not unique/);
});

test("defers non-critical delivery through UTC quiet hours", () => {
  const config = authorityConfig();
  const quietProposal = { ...proposal, not_before: "2026-08-28T22:00:00Z" };
  const result = evaluateAgencyProposal(quietProposal, config, new Date("2026-08-28T23:30:00Z"));
  assert.equal(result.allowed, true);
  assert.equal(result.deliverAt?.toISOString(), "2026-08-29T07:00:00.000Z");
  const critical = evaluateAgencyProposal({ ...quietProposal, priority: 0.95 }, config, new Date("2026-08-28T23:30:00Z"));
  assert.equal(critical.deliverAt?.toISOString(), "2026-08-28T23:30:01.000Z");
});

test("schedules an authorized proposal and records durable approval", async () => {
  const fixture = signedRecord(proposal);
  const calls: Array<{ kind: string; value: unknown }> = [];
  const api = authorityAPI(calls);
  const client = {
    async claimAgency() { throw new Error("unused"); },
    async resolveAgency(input: unknown) {
      calls.push({ kind: "resolve", value: input });
      return fixture.record;
    },
  };
  await processAgencyClaim(api, client, authorityConfig(), fixture.publicKey, "test-authority-token", fixture.record, new Date("2026-08-28T21:30:00Z"));
  assert.deepEqual(calls.map((call) => call.kind), ["unschedule", "schedule", "resolve"]);
  assert.match(JSON.stringify(calls[1]?.value), /not permission to execute tools/);
  assert.match(JSON.stringify(calls[2]?.value), /"decision":"approved"/);
});

test("retries a lost approval response without cancelling committed work", async () => {
  const fixture = signedRecord(proposal);
  const calls: Array<{ kind: string; value: unknown }> = [];
  let resolutions = 0;
  const client = {
    async claimAgency() { throw new Error("unused"); },
    async resolveAgency(input: unknown) {
      resolutions++;
      calls.push({ kind: "resolve", value: input });
      if (resolutions === 1) throw new Error("temporary daemon failure");
      return fixture.record;
    },
  };
  await processAgencyClaim(authorityAPI(calls), client, authorityConfig(), fixture.publicKey, "test-authority-token", fixture.record, new Date("2026-08-28T21:30:00Z"));
  assert.deepEqual(calls.map((call) => call.kind), ["unschedule", "schedule", "resolve", "resolve"]);
  assert.match(JSON.stringify(calls.at(-1)?.value), /"decision":"approved"/);
});

test("a slow cleanup cannot hand an expired tau turn to the scheduler", async () => {
  const fixture = signedRecord(tauProposal());
  const calls: Array<{kind: string; value: unknown}> = [];
  let clock = new Date("2026-08-28T21:00:03Z");
  const api = authorityAPI(calls);
  api.session.workflow.unscheduleSessionTurnsByTag = async (input) => { calls.push({kind:"unschedule",value:input}); clock = new Date("2026-08-28T21:00:11Z"); return {removed:0,failed:0}; };
  const config = {...authorityConfig(), agencyCapabilities:["eventframe.agency.schedule"], agencyConsentActions:["schedule" as const]};
  const client = {async claimAgency(){throw new Error("unused")},async resolveAgency(input:unknown){calls.push({kind:"resolve",value:input});return fixture.record}};
  await processAgencyClaim(api,client,config,fixture.publicKey,"authority",fixture.record,new Date("2026-08-28T21:00:03Z"),()=>true,()=>clock);
  assert.equal(calls.filter(c=>c.kind==="schedule").length,0);
  assert.match(JSON.stringify(calls.at(-1)?.value),/"decision":"rejected"/);
});

test("failed cancellation records the scheduler handle as uncertain", async () => {
  const fixture = signedRecord(proposal);
  const calls: Array<{kind:string;value:unknown}> = [];
  const api = authorityAPI(calls);
  let cleanup = 0;
  api.session.workflow.unscheduleSessionTurnsByTag = async () => { if (++cleanup>1) throw new Error("cancel failed"); return {removed:0,failed:0}; };
  const client = {async claimAgency(){throw new Error("unused")},async getAgency(){return fixture.record},async resolveAgency(input:{decision:string}){calls.push({kind:"resolve",value:input});if(input.decision==="approved")throw new Error("approval failed");return fixture.record}};
  await processAgencyClaim(api,client,authorityConfig(),fixture.publicKey,"authority",fixture.record,new Date("2026-08-28T21:30:00Z"));
  assert.match(JSON.stringify(calls.at(-1)?.value),/"decision":"uncertain"/);
  assert.match(JSON.stringify(calls.at(-1)?.value),/"executionRef":"job-1"/);
});

test("compensation holds uncertainty until cancellation acknowledges", async () => {
  const fixture=signedRecord(proposal);
  let state:AgencyProposalRecord={...fixture.record};
  const api=authorityAPI([]);
  let cleanup=0;
  api.session.workflow.unscheduleSessionTurnsByTag=async()=>{if(++cleanup>1)assert.equal(state.status,"uncertain");return {removed:1,failed:0}};
  const client={async claimAgency(){throw new Error("unused")},async getAgency(){return state},async resolveAgency(input:{decision:"approved"|"rejected"|"uncertain";executionRef?:string}){if(input.decision==="approved")throw new Error("uncommitted approval");state={...state,status:input.decision,execution_ref:input.executionRef};return state}};
  await processAgencyClaim(api,client,authorityConfig(),fixture.publicKey,"authority",fixture.record,new Date("2026-08-28T21:30:00Z"));
  assert.equal(state.status,"rejected");assert.equal(cleanup,2);
});

test("an older attempt cannot cancel another worker's approved handle", async () => {
  const fixture=signedRecord(proposal);const calls:Array<{kind:string;value:unknown}>=[];
  const client={async claimAgency(){throw new Error("unused")},async getAgency(){return {...fixture.record,status:"approved" as const,execution_ref:"new-job"}},async resolveAgency(input:{decision:string}){calls.push({kind:"resolve",value:input});if(input.decision==="approved")throw new Error("old lease");return {...fixture.record,status:"uncertain" as const}}};
  await processAgencyClaim(authorityAPI(calls),client,authorityConfig(),fixture.publicKey,"authority",fixture.record,new Date("2026-08-28T21:30:00Z"));
  assert.equal(calls.filter(c=>c.kind==="unschedule").length,1);assert.match(JSON.stringify(calls.at(-1)?.value),/"decision":"uncertain"/);
});

test("rolls back an ambiguous scheduler failure before recording rejection", async () => {
  const fixture = signedRecord(proposal);
  const calls: Array<{ kind: string; value: unknown }> = [];
  let unschedules = 0;
  const api = {
    logger: { warn(message: unknown) { calls.push({ kind: "warn", value: message }); } },
    session: {
      workflow: {
        async unscheduleSessionTurnsByTag(input: unknown) {
          unschedules++;
          calls.push({ kind: "unschedule", value: input });
        },
        async scheduleSessionTurn(input: unknown) {
          calls.push({ kind: "schedule", value: input });
          throw new Error("scheduler response was lost");
        },
      },
    },
  } as unknown as OpenClawPluginApi;
  const client = {
    async claimAgency() { throw new Error("unused"); },
    async resolveAgency(input: unknown) {
      calls.push({ kind: "resolve", value: input });
      return fixture.record;
    },
  };
  await processAgencyClaim(api, client, authorityConfig(), fixture.publicKey, "test-authority-token", fixture.record, new Date("2026-08-28T21:30:00Z"));
  assert.equal(unschedules, 2);
  assert.match(JSON.stringify(calls.at(-1)?.value), /"decision":"rejected"/);
});

test("rolls back a turn when authority stops during scheduling", async () => {
  const fixture = signedRecord(proposal);
  const calls: Array<{ kind: string; value: unknown }> = [];
  let active = true;
  const api = {
    logger: { warn(message: unknown) { calls.push({ kind: "warn", value: message }); } },
    session: {
      workflow: {
        async unscheduleSessionTurnsByTag(input: unknown) { calls.push({ kind: "unschedule", value: input }); },
        async scheduleSessionTurn(input: unknown) {
          calls.push({ kind: "schedule", value: input });
          active = false;
          return { id: "job-1" };
        },
      },
    },
  } as unknown as OpenClawPluginApi;
  const client = {
    async claimAgency() { throw new Error("unused"); },
    async resolveAgency(input: unknown) {
      calls.push({ kind: "resolve", value: input });
      return fixture.record;
    },
  };
  await processAgencyClaim(api, client, authorityConfig(), fixture.publicKey, "test-authority-token", fixture.record, new Date("2026-08-28T21:30:00Z"), () => active);
  assert.deepEqual(calls.map((call) => call.kind), ["unschedule", "schedule", "unschedule", "resolve"]);
  assert.match(JSON.stringify(calls.at(-1)?.value), /"decision":"rejected"/);
});

function authorityAPI(calls: Array<{ kind: string; value: unknown }>): OpenClawPluginApi {
  return {
    logger: { warn(message: unknown) { calls.push({ kind: "warn", value: message }); } },
    session: {
      workflow: {
        async unscheduleSessionTurnsByTag(input: unknown) { calls.push({ kind: "unschedule", value: input }); },
        async scheduleSessionTurn(input: unknown) {
          calls.push({ kind: "schedule", value: input });
          return { id: "job-1" };
        },
      },
    },
  } as unknown as OpenClawPluginApi;
}
