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

test("rolls back a scheduled turn before rejecting a failed durable handoff", async () => {
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
  assert.deepEqual(calls.map((call) => call.kind), ["unschedule", "schedule", "resolve", "unschedule", "resolve"]);
  assert.match(JSON.stringify(calls.at(-1)?.value), /"decision":"rejected"/);
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
