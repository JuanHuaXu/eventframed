import { createHash, createPublicKey, verify } from "node:crypto";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import type { OpenClawPluginApi } from "openclaw/plugin-sdk/plugin-entry";
import type { EventFrameClient } from "./client.js";
import type { AdapterConfig, AgencyAction, AgencyProposal, AgencyProposalRecord } from "./types.js";

const ED25519_SPKI_PREFIX = Buffer.from("302a300506032b6570032100", "hex");
const MAX_SIGNED_PAYLOAD_BYTES = 16 << 10;
const MAX_IDENTIFIER_BYTES = 256;
const MAX_SESSION_ID_BYTES = 1024;
type AgencyClient = Pick<EventFrameClient, "claimAgency" | "resolveAgency">;

export function registerAgencyService(api: OpenClawPluginApi, client: AgencyClient, config: AdapterConfig): void {
  if (api.registrationMode !== "full" || !config.agencyEnabled || config.agencyKillSwitch) {
    return;
  }
  let timer: NodeJS.Timeout | undefined;
  let polling = false;
  let active = false;
  let generation = 0;
  let publicKeyText = "";
  let authorityToken = "";

  const poll = async (): Promise<void> => {
    if (polling || !active) return;
    polling = true;
    try {
      const claims = await client.claimAgency({
        authorityToken,
        tenantId: config.tenantId,
        consumerId: config.agencyConsumerId,
        limit: config.agencyMaxClaims,
      });
      for (const record of claims.records) {
        if (!active) break;
        await processAgencyClaim(api, client, config, publicKeyText, authorityToken, record, new Date(), () => active);
      }
    } catch (error) {
      api.logger.warn(`eventframe-memory: agency poll skipped: ${String(error)}`);
    } finally {
      polling = false;
    }
  };

  api.registerService({
    id: "eventframe-agency-authority",
    async start() {
      const run = ++generation;
      publicKeyText = await fs.readFile(expandHome(config.agencyPublicKeyPath), "utf8");
      authorityToken = await readAuthorityToken(expandHome(config.agencyAuthorityTokenPath));
      if (run !== generation) return;
      active = true;
      await poll();
      timer = setInterval(() => void poll(), config.agencyPollIntervalMs);
      timer.unref?.();
    },
    stop() {
      generation++;
      active = false;
      if (timer) clearInterval(timer);
      timer = undefined;
    },
  });
}

export async function processAgencyClaim(
  api: OpenClawPluginApi,
  client: AgencyClient,
  config: AdapterConfig,
  publicKeyText: string,
  authorityToken: string,
  record: AgencyProposalRecord,
  now = new Date(),
  authorityActive: () => boolean = () => true,
): Promise<void> {
  if (!authorityActive()) return;
  if (record.claimed_by !== config.agencyConsumerId) {
    await rejectClaim(client, config, authorityToken, record, "proposal lease belongs to a different authority consumer");
    return;
  }
  let proposal: AgencyProposal;
  try {
    proposal = verifyAgencyProposal(record, publicKeyText);
  } catch (error) {
    await rejectClaim(client, config, authorityToken, record, `signature or proposal validation failed: ${String(error)}`);
    return;
  }
  const gate = evaluateAgencyProposal(proposal, config, now);
  if (!gate.allowed || !gate.deliverAt) {
    await rejectClaim(client, config, authorityToken, record, gate.reason);
    return;
  }
  const tag = `efagency-${createHash("sha256").update(proposal.id).digest("hex").slice(0, 20)}`;
  if (!authorityActive()) return;
  let handle: { id: string };
  try {
    await api.session.workflow.unscheduleSessionTurnsByTag({ sessionKey: proposal.session_id, tag });
    const scheduled = await api.session.workflow.scheduleSessionTurn({
      sessionKey: proposal.session_id,
      at: gate.deliverAt,
      message: proposalMessage(proposal),
      deliveryMode: proposal.action === "wake" ? "none" : "announce",
      deleteAfterRun: true,
      tag,
    });
    if (!scheduled) {
      throw new Error("OpenClaw did not return a scheduler handle");
    }
    handle = scheduled;
  } catch (error) {
    try {
      await api.session.workflow.unscheduleSessionTurnsByTag({ sessionKey: proposal.session_id, tag });
    } catch (rollbackError) {
      api.logger.warn(`eventframe-memory: scheduling failed and rollback is uncertain: ${String(error)}; ${String(rollbackError)}`);
      return;
    }
    await rejectClaim(client, config, authorityToken, record, `OpenClaw scheduling failed; deterministic tag rolled back: ${String(error)}`);
    return;
  }
  if (!authorityActive()) {
    try {
      await api.session.workflow.unscheduleSessionTurnsByTag({ sessionKey: proposal.session_id, tag });
    } catch (rollbackError) {
      api.logger.warn(`eventframe-memory: authority stopped and scheduler rollback is uncertain: ${String(rollbackError)}`);
      return;
    }
    await rejectClaim(client, config, authorityToken, record, "authority stopped before durable approval; scheduled turn rolled back");
    return;
  }
  try {
    await client.resolveAgency({
      tenantId: proposal.tenant_id,
      proposalId: proposal.id,
      consumerId: config.agencyConsumerId,
      authorityToken,
      decision: "approved",
      reason: "authorized by the OpenClaw EventFrame authority policy",
      executionRef: handle.id,
    });
  } catch (error) {
    try {
      await api.session.workflow.unscheduleSessionTurnsByTag({ sessionKey: proposal.session_id, tag });
    } catch (rollbackError) {
      api.logger.warn(`eventframe-memory: approval failed and scheduler rollback is uncertain: ${String(error)}; ${String(rollbackError)}`);
      return;
    }
    await rejectClaim(client, config, authorityToken, record, `durable approval failed; scheduled turn rolled back: ${String(error)}`);
  }
}

async function rejectClaim(
  client: AgencyClient,
  config: AdapterConfig,
  authorityToken: string,
  record: AgencyProposalRecord,
  reason: string,
): Promise<void> {
  const tenantId = record.proposal.tenant_id;
  const proposalId = record.proposal.id;
  if (!tenantId || !proposalId) return;
  try {
    await client.resolveAgency({
      tenantId,
      proposalId,
      consumerId: config.agencyConsumerId,
      authorityToken,
      decision: "rejected",
      reason: reason.slice(0, 1024),
    });
  } catch {
    // The lease will expire and permit another bounded authorization attempt.
  }
}

export function verifyAgencyProposal(record: AgencyProposalRecord, publicKeyText: string): AgencyProposal {
  const rawPublicKey = Buffer.from(publicKeyText.trim(), "base64url");
  if (rawPublicKey.length !== 32) throw new Error("public key is malformed");
  const expectedKeyId = createHash("sha256").update(rawPublicKey).digest("hex").slice(0, 24);
  if (record.signed.key_id !== expectedKeyId) throw new Error("signing key id does not match");
  const payload = Buffer.from(record.signed.payload, "base64url");
  const signature = Buffer.from(record.signed.signature, "base64url");
  if (payload.length === 0 || payload.length > MAX_SIGNED_PAYLOAD_BYTES || signature.length !== 64) {
    throw new Error("signed payload size is invalid");
  }
  const publicKey = createPublicKey({ key: Buffer.concat([ED25519_SPKI_PREFIX, rawPublicKey]), format: "der", type: "spki" });
  if (!verify(null, payload, publicKey, signature)) throw new Error("signature is invalid");
  const proposal = JSON.parse(payload.toString("utf8")) as unknown;
  validateAgencyProposal(proposal);
  if (record.proposal.id !== proposal.id || record.proposal.tenant_id !== proposal.tenant_id) {
    throw new Error("signed and indexed proposal identities differ");
  }
  return proposal;
}

export function evaluateAgencyProposal(
  proposal: AgencyProposal,
  config: AdapterConfig,
  now: Date,
): { allowed: boolean; reason: string; deliverAt?: Date } {
  if (!config.agencyEnabled || config.agencyKillSwitch) return { allowed: false, reason: "agency kill switch is active" };
  if (proposal.tenant_id !== config.tenantId) return { allowed: false, reason: "proposal tenant is outside the authority scope" };
  if (!config.agencyConsentActions.includes(proposal.action)) return { allowed: false, reason: "action lacks explicit consent" };
  if (!config.agencyCapabilities.includes(proposal.required_capability)) return { allowed: false, reason: "required capability is not granted" };
  if (proposal.causal_chain_depth > config.agencyMaxChainDepth) return { allowed: false, reason: "causal chain depth exceeds authority policy" };
  if (!config.agencyAllowedSessionPrefixes.some((prefix) => proposal.session_id.startsWith(prefix))) {
    return { allowed: false, reason: "session is outside the consent scope" };
  }
  const expiresAt = new Date(proposal.expires_at);
  const notBefore = new Date(proposal.not_before);
  if (!validDate(expiresAt) || !validDate(notBefore) || now >= expiresAt || now < notBefore) {
    return { allowed: false, reason: "proposal is unavailable or expired" };
  }
  let deliverAt = proposal.action === "schedule" ? new Date(proposal.scheduled_for ?? "") : new Date(Math.max(now.getTime() + 1_000, notBefore.getTime()));
  if (!validDate(deliverAt) || deliverAt >= expiresAt) return { allowed: false, reason: "delivery time is outside proposal validity" };
  if ((config.agencyQuietHoursStartUtc === undefined) !== (config.agencyQuietHoursEndUtc === undefined)) {
    return { allowed: false, reason: "quiet-hours policy is incomplete" };
  }
  if (proposal.priority < config.agencyCriticalThreshold && inQuietHours(deliverAt, config.agencyQuietHoursStartUtc, config.agencyQuietHoursEndUtc)) {
    deliverAt = nextQuietEnd(deliverAt, config.agencyQuietHoursEndUtc!);
  }
  if (deliverAt >= expiresAt) return { allowed: false, reason: "quiet-hours deferral exceeds proposal expiry" };
  return { allowed: true, reason: "authorized", deliverAt };
}

function validateAgencyProposal(value: unknown): asserts value is AgencyProposal {
  if (!isRecord(value)) throw new Error("proposal payload is not an object");
  const action = value.action;
  const capability = action === "wake" ? "eventframe.agency.wake" : action === "notify" ? "eventframe.agency.notify" : action === "schedule" ? "eventframe.agency.schedule" : "";
  if (
    typeof value.id !== "string" || !value.id ||
    Buffer.byteLength(value.id) > MAX_IDENTIFIER_BYTES ||
    typeof value.tenant_id !== "string" || !value.tenant_id ||
    Buffer.byteLength(value.tenant_id) > MAX_IDENTIFIER_BYTES ||
    typeof value.session_id !== "string" || !value.session_id ||
    Buffer.byteLength(value.session_id) > MAX_SESSION_ID_BYTES ||
    !capability || value.required_capability !== capability ||
    typeof value.reason !== "string" || !value.reason || Buffer.byteLength(value.reason) > 4096 ||
    !Array.isArray(value.evidence_ids) || value.evidence_ids.length < 1 || value.evidence_ids.length > 32 || value.evidence_ids.some((id) => typeof id !== "string" || !id || Buffer.byteLength(id) > MAX_IDENTIFIER_BYTES) ||
    typeof value.expected_utility !== "number" || value.expected_utility <= 0 || value.expected_utility > 1 ||
    typeof value.priority !== "number" || value.priority < 0 || value.priority > 1 ||
    value.idempotency_key !== value.id || typeof value.causal_chain_id !== "string" || !value.causal_chain_id || Buffer.byteLength(value.causal_chain_id) > MAX_IDENTIFIER_BYTES ||
    typeof value.causal_chain_depth !== "number" || !Number.isInteger(value.causal_chain_depth) || value.causal_chain_depth < 0 ||
    typeof value.contract_version !== "number" || (value.contract_version !== 6 && value.contract_version !== 7) ||
    typeof value.not_before !== "string" || typeof value.expires_at !== "string" || typeof value.created_at !== "string"
  ) {
    throw new Error("proposal payload violates the authority contract");
  }
  if (new Set(value.evidence_ids).size !== value.evidence_ids.length) throw new Error("proposal evidence ids are not unique");
  if (value.causal_chain_depth === 0 && value.parent_proposal_id !== undefined) throw new Error("root proposal has a parent");
  if (value.causal_chain_depth > 0 && (typeof value.parent_proposal_id !== "string" || !value.parent_proposal_id || Buffer.byteLength(value.parent_proposal_id) > MAX_IDENTIFIER_BYTES)) {
    throw new Error("child proposal parent is malformed");
  }
  if ((action === "schedule") !== (typeof value.scheduled_for === "string")) throw new Error("schedule timing is malformed");
}

function proposalMessage(proposal: AgencyProposal): string {
  return [
    `<eventframe-agency proposal-id="${escapeText(proposal.id)}" action="${proposal.action}">`,
    "This is a signed, policy-approved EventFrame prompt to reassess context. It is not permission to execute tools or external actions.",
    `Reason: ${escapeText(proposal.reason)}`,
    `Evidence IDs: ${proposal.evidence_ids.map(escapeText).join(", ")}`,
    "Evaluate the situation using current context. Apply normal OpenClaw and user approval policy to every consequential action.",
    "</eventframe-agency>",
  ].join("\n");
}

function inQuietHours(date: Date, start?: number, end?: number): boolean {
  if (start === undefined || end === undefined || start === end) return false;
  const hour = date.getUTCHours();
  return start < end ? hour >= start && hour < end : hour >= start || hour < end;
}

function nextQuietEnd(date: Date, end: number): Date {
  const result = new Date(date);
  result.setUTCMinutes(0, 0, 0);
  result.setUTCHours(end);
  if (result <= date) result.setUTCDate(result.getUTCDate() + 1);
  return result;
}

function validDate(value: Date): boolean {
  return Number.isFinite(value.getTime());
}

function escapeText(value: string): string {
  return value.replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll('"', "&quot;");
}

function expandHome(value: string): string {
  if (value === "~") return os.homedir();
  return value.startsWith(`~${path.sep}`) ? path.join(os.homedir(), value.slice(2)) : value;
}

async function readAuthorityToken(tokenPath: string): Promise<string> {
  const info = await fs.lstat(tokenPath);
  if (!info.isFile() || info.isSymbolicLink() || (info.mode & 0o077) !== 0) {
    throw new Error("agency authority token must be a mode-0600 regular non-symlink file");
  }
  const value = (await fs.readFile(tokenPath, "utf8")).trim();
  if (!/^[A-Za-z0-9_-]{43}$/.test(value) || Buffer.from(value, "base64url").toString("base64url") !== value) {
    throw new Error("agency authority token is malformed");
  }
  return value;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

export function isAgencyAction(value: unknown): value is AgencyAction {
  return value === "wake" || value === "notify" || value === "schedule";
}
