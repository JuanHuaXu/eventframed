import http from "node:http";
import os from "node:os";
import path from "node:path";
import type { AgencyClaimResponse, AgencyProposalRecord, CapturedTurn, ContextPacket, OutcomeObservation } from "./types.js";
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

  async captureTurn(turn: CapturedTurn): Promise<unknown> {
    const value = await this.request("POST", "/v1/turns:capture", {
      protocol_version: PROTOCOL_VERSION,
      idempotency_key: turn.id,
      turn,
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
    const payload = {
      protocol_version: PROTOCOL_VERSION,
      tenant_id: input.tenantId,
      session_id: input.sessionId,
      query: input.query,
      as_of: (input.asOf ?? new Date()).toISOString(),
      recall_k: input.recallK,
      pack_k: input.packK,
      token_budget: input.tokenBudget,
    };
    for (let attempt = 0; ; attempt++) {
      try {
        return parseContextPacket(await this.request("POST", "/v1/openclaw/context:recall", payload));
      } catch (error) {
        if (attempt >= 2 || !(error instanceof EventFrameHTTPError) || error.status !== 409 || error.code !== "snapshot_changed") {
          throw error;
        }
        await sleep(5 * (attempt + 1));
      }
    }
  }

  async observeOutcome(input: OutcomeObservation): Promise<unknown> {
    const observedAt = input.observed_at ?? new Date().toISOString();
    const signals = outcomeSignals(input.signal);
    const value = await this.request("POST", "/v1/bayesian/outcomes:observe", {
      protocol_version: PROTOCOL_VERSION,
      idempotency_key: input.idempotency_key,
      tenant_id: input.tenant_id,
      journal_id: input.journal_id,
      event_id: input.event_id,
      useful: input.signal === "useful" || input.signal === "cited" || input.signal === "successful_downstream",
      signals,
      observed_at: observedAt,
      available_at: input.available_at ?? observedAt,
      source: "full_stream",
      inclusion_probability: 1,
    });
    assertProtocol(value);
    return value;
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
              const code = isRecord(value) && typeof value.code === "string" ? value.code : "http_error";
              reject(new EventFrameHTTPError(response.statusCode ?? 500, code, message));
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

class EventFrameHTTPError extends Error {
  constructor(readonly status: number, readonly code: string, message: string) {
    super(`eventframed ${status}: ${message}`);
  }
}

function sleep(milliseconds: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, milliseconds));
}

function outcomeSignals(signal: OutcomeObservation["signal"]): Record<string, boolean> {
  switch (signal) {
    case "useful": return { explicit_useful: true };
    case "not_useful": return { explicit_useful: false };
    case "cited": return { cited: true };
    case "successful_downstream": return { successful_downstream: true };
    case "correction": return { correction: true };
    case "rejected": return { rejected: true };
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
    if (
      !isRecord(candidate) ||
      !isRecord(candidate.event) ||
      typeof candidate.retrieval_score !== "number" ||
      !Number.isFinite(candidate.retrieval_score) ||
      typeof candidate.rank_delta !== "number" ||
      !Number.isFinite(candidate.rank_delta) ||
      (candidate.rank_delta_scale !== undefined &&
        (typeof candidate.rank_delta_scale !== "number" || !Number.isFinite(candidate.rank_delta_scale))) ||
      (candidate.resolution_rank_delta !== undefined &&
        (typeof candidate.resolution_rank_delta !== "number" || !Number.isFinite(candidate.resolution_rank_delta))) ||
      (candidate.rank_delta_confidence !== undefined &&
        (typeof candidate.rank_delta_confidence !== "number" || !Number.isFinite(candidate.rank_delta_confidence))) ||
      (candidate.rank_delta_answer_certainty !== undefined &&
        (typeof candidate.rank_delta_answer_certainty !== "number" ||
          !Number.isFinite(candidate.rank_delta_answer_certainty) ||
          candidate.rank_delta_answer_certainty < 0 ||
          candidate.rank_delta_answer_certainty > 1)) ||
      (candidate.rank_delta_correction_reliability !== undefined &&
        (typeof candidate.rank_delta_correction_reliability !== "number" ||
          !Number.isFinite(candidate.rank_delta_correction_reliability) ||
          candidate.rank_delta_correction_reliability < 0 ||
          candidate.rank_delta_correction_reliability > 1)) ||
      typeof candidate.score !== "number" ||
      !Number.isFinite(candidate.score)
    ) {
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
  if (
    value.packet_confidence !== undefined &&
    (typeof value.packet_confidence !== "number" || !Number.isFinite(value.packet_confidence))
  ) {
    throw new Error("eventframed returned malformed packet confidence");
  }
  if (
    value.packet_answer_certainty !== undefined &&
    (typeof value.packet_answer_certainty !== "number" ||
      !Number.isFinite(value.packet_answer_certainty) ||
      value.packet_answer_certainty < 0 ||
      value.packet_answer_certainty > 1)
  ) {
    throw new Error("eventframed returned malformed packet answer certainty");
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
