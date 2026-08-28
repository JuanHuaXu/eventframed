import { createHash } from "node:crypto";
import type { EventField, EventFrame } from "./types.js";

export type TurnCapture = {
  tenantId: string;
  sessionId: string;
  runId?: string;
  agentId?: string;
  userText: string;
  assistantText: string;
  retrievedIds: string[];
  occurredAt?: Date;
  observedAt?: Date;
};

export function buildTurnEvent(input: TurnCapture): EventFrame {
  const occurredAt = input.occurredAt ?? input.observedAt ?? new Date();
  const observedAt = input.observedAt ?? new Date();
  const content = [`User: ${input.userText}`, `Assistant: ${input.assistantText}`].join("\n\n");
  const identity = [input.tenantId, input.sessionId, input.runId ?? "", content].join("\u0000");
  const id = `evt_${createHash("sha256").update(identity).digest("hex").slice(0, 32)}`;

  return {
    id,
    tenant_id: input.tenantId,
    session_id: input.sessionId,
    sequence: Math.max(0, Math.floor(occurredAt.getTime())),
    kind: "agent_turn",
    content,
    occurred_at: occurredAt.toISOString(),
    observed_at: observedAt.toISOString(),
    available_at: observedAt.toISOString(),
    who: field(input.agentId ? `user and agent:${input.agentId}` : "user and agent", "inferred", 0.8),
    what: field("completed conversational turn", "synthetic", 1),
    where: field(`session:${input.sessionId}`, "observed", 1),
    when: field(occurredAt.toISOString(), "observed", 1),
    why: field("agent response to the current user turn", "inferred", 0.9),
    how: field("OpenClaw agent run", "observed", 1),
    priority: 0.5,
    tags: ["conversation", "agent-turn"],
    provenance: {
      producer: "openclaw-eventframe-memory",
      retrieved_ids: [...new Set(input.retrievedIds)],
      run_id: input.runId,
    },
    attributes: {
      user_source: "observed",
      assistant_source: "synthetic",
    },
  };
}

function field(value: string, source: EventField["source"], confidence: number): EventField {
  return { value, source, confidence };
}

export function extractLatestText(messages: unknown[], role: "user" | "assistant"): string | undefined {
  for (let index = messages.length - 1; index >= 0; index--) {
    const message = asRecord(messages[index]);
    if (message?.role !== role) {
      continue;
    }
    const text = extractContent(message.content).trim();
    if (text) {
      return text;
    }
  }
  return undefined;
}

function extractContent(content: unknown): string {
  if (typeof content === "string") {
    return content;
  }
  if (!Array.isArray(content)) {
    return "";
  }
  return content
    .flatMap((block) => {
      const record = asRecord(block);
      return record?.type === "text" && typeof record.text === "string" ? [record.text] : [];
    })
    .join("\n");
}

function asRecord(value: unknown): Record<string, unknown> | undefined {
  return typeof value === "object" && value !== null ? (value as Record<string, unknown>) : undefined;
}
