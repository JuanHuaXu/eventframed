import http from "node:http";
import os from "node:os";
import path from "node:path";
import type { ContextPacket, EventFrame } from "./types.js";
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
