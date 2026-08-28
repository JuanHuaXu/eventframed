import http from "node:http";
import os from "node:os";
import path from "node:path";
import type { AgencyClaimResponse, AgencyProposalRecord, ContextPacket, EventFrame } from "./types.js";
import { PROTOCOL_VERSION } from "./types.js";

const MAX_RESPONSE_BYTES = 8 << 20;

type ClientOptions = {
  socketPath: string;
  timeoutMs?: number;
};

export class EventFrameClient {
  readonly socketPath: string;
  readonly timeoutMs: number;

  constructor(options: ClientOptions) {
    this.socketPath = expandHome(options.socketPath);
    this.timeoutMs = options.timeoutMs ?? 2_000;
  }

  async health(): Promise<unknown> {
    const value = await this.request("GET", "/v1/health");
    assertProtocol(value);
    return value;
  }

  async observe(event: EventFrame): Promise<unknown> {
    const value = await this.request("POST", "/v1/events:observe", {
      protocol_version: PROTOCOL_VERSION,
      idempotency_key: event.id,
      event,
    });
    assertProtocol(value);
    return value;
  }

  async recall(input: {
    tenantId: string;
    sessionId: string;
    query: string;
    recallK: number;
    packK: number;
    tokenBudget: number;
    asOf?: Date;
  }): Promise<ContextPacket> {
    const value = await this.request("POST", "/v1/context:recall", {
      protocol_version: PROTOCOL_VERSION,
      tenant_id: input.tenantId,
      session_id: input.sessionId,
      query: input.query,
      as_of: (input.asOf ?? new Date()).toISOString(),
      recall_k: input.recallK,
      pack_k: input.packK,
      token_budget: input.tokenBudget,
    });
    return parseContextPacket(value);
  }

  async claimAgency(input: { authorityToken: string; tenantId: string; consumerId: string; limit: number }): Promise<AgencyClaimResponse> {
    const value = await this.request("POST", "/v1/agency/proposals:claim", {
      protocol_version: PROTOCOL_VERSION,
      authority_token: input.authorityToken,
      tenant_id: input.tenantId,
      consumer_id: input.consumerId,
      limit: input.limit,
    });
    return parseAgencyClaims(value, input.consumerId);
  }

  async resolveAgency(input: {
    tenantId: string;
    proposalId: string;
    consumerId: string;
    authorityToken: string;
    decision: "approved" | "rejected";
    reason: string;
    executionRef?: string;
  }): Promise<AgencyProposalRecord> {
    const value = await this.request("POST", "/v1/agency/proposals:resolve", {
      protocol_version: PROTOCOL_VERSION,
      authority_token: input.authorityToken,
      tenant_id: input.tenantId,
      proposal_id: input.proposalId,
      consumer_id: input.consumerId,
      decision: input.decision,
      reason: input.reason,
      execution_ref: input.executionRef ?? "",
    });
    assertProtocol(value);
    if (
      !isRecord(value.record) ||
      !isRecord(value.record.proposal) ||
      value.record.proposal.id !== input.proposalId ||
      value.record.proposal.tenant_id !== input.tenantId ||
      value.record.status !== input.decision ||
      (input.decision === "approved" && value.record.execution_ref !== input.executionRef)
    ) {
      throw new Error("eventframed returned a malformed agency resolution");
    }
    return value.record as AgencyProposalRecord;
  }

  private request(method: string, requestPath: string, body?: unknown): Promise<unknown> {
    const payload = body === undefined ? undefined : JSON.stringify(body);
    return new Promise((resolve, reject) => {
      const request = http.request(
        {
          method,
          path: requestPath,
          socketPath: this.socketPath,
          headers: payload
            ? {
                "content-type": "application/json",
                "content-length": Buffer.byteLength(payload),
              }
            : undefined,
          timeout: this.timeoutMs,
        },
        (response) => {
          const chunks: Buffer[] = [];
          let size = 0;
          response.on("data", (chunk: Buffer) => {
            size += chunk.length;
            if (size > MAX_RESPONSE_BYTES) {
              response.destroy(new Error("eventframed response exceeds 8 MiB"));
              return;
            }
            chunks.push(chunk);
          });
          response.on("error", reject);
          response.on("end", () => {
            const text = Buffer.concat(chunks).toString("utf8");
            let value: unknown;
            try {
              value = text === "" ? undefined : JSON.parse(text);
            } catch (error) {
              reject(new Error(`eventframed returned invalid JSON: ${String(error)}`));
              return;
            }
            if ((response.statusCode ?? 500) >= 400) {
              const message = isRecord(value) && typeof value.message === "string" ? value.message : text;
              reject(new Error(`eventframed ${response.statusCode}: ${message}`));
              return;
            }
            resolve(value);
          });
        },
      );
      request.on("timeout", () => request.destroy(new Error("eventframed request timed out")));
      request.on("error", reject);
      if (payload) {
        request.write(payload);
      }
      request.end();
    });
  }
}

function expandHome(value: string): string {
  if (value === "~") {
    return os.homedir();
  }
  if (value.startsWith(`~${path.sep}`)) {
    return path.join(os.homedir(), value.slice(2));
  }
  return value;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function assertProtocol(value: unknown): asserts value is Record<string, unknown> {
  if (!isRecord(value) || value.protocol_version !== PROTOCOL_VERSION) {
    throw new Error("eventframed returned an unsupported or missing protocol_version");
  }
}

function parseContextPacket(value: unknown): ContextPacket {
  assertProtocol(value);
  if (!Array.isArray(value.candidates)) {
    throw new Error("eventframed returned a malformed context packet");
  }
  for (const candidate of value.candidates) {
    if (!isRecord(candidate) || !isRecord(candidate.event)) {
      throw new Error("eventframed returned a malformed candidate");
    }
    const event = candidate.event;
    if (
      typeof event.id !== "string" ||
      typeof event.content !== "string" ||
      typeof event.kind !== "string" ||
      typeof event.available_at !== "string" ||
      !isRecord(event.provenance) ||
      typeof event.provenance.producer !== "string"
    ) {
      throw new Error("eventframed returned a malformed event");
    }
  }
  return value as ContextPacket;
}

function parseAgencyClaims(value: unknown, expectedConsumerId: string): AgencyClaimResponse {
  assertProtocol(value);
  if (!Array.isArray(value.records)) {
    throw new Error("eventframed returned malformed agency claims");
  }
  for (const record of value.records) {
    if (
      !isRecord(record) ||
      !isRecord(record.proposal) ||
      !isRecord(record.signed) ||
      typeof record.proposal.id !== "string" ||
      typeof record.proposal.tenant_id !== "string" ||
      typeof record.signed.payload !== "string" ||
      typeof record.signed.signature !== "string" ||
      typeof record.signed.key_id !== "string" ||
      record.status !== "claimed" ||
      record.claimed_by !== expectedConsumerId ||
      typeof record.lease_until !== "string" ||
      !Number.isFinite(new Date(record.lease_until).getTime())
    ) {
      throw new Error("eventframed returned a malformed agency proposal record");
    }
  }
  return value as AgencyClaimResponse;
}
