import { createHash } from "node:crypto";
import type { CapturedTurn, CaptureIdentity } from "./types.js";

export type TurnCapture = {
  tenantId: string;
  sessionId: string;
  runId?: string;
  agentId?: string;
  userId?: string;
  participantUserIds?: string[];
  previousTurnId?: string;
  userText: string;
  assistantText: string;
  retrievedIds: string[];
  occurredAt?: Date;
  observedAt?: Date;
};

export function buildTurnCapture(input: TurnCapture): CapturedTurn {
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
    run_id: input.runId,
    agent_id: input.agentId,
    user_id: input.userId,
    participant_user_ids: input.participantUserIds ? [...new Set(input.participantUserIds)] : undefined,
    previous_turn_id: input.previousTurnId,
    user_text: input.userText,
    assistant_text: input.assistantText,
    retrieved_ids: [...new Set(input.retrievedIds)],
    occurred_at: occurredAt.toISOString(),
    observed_at: observedAt.toISOString(),
    available_at: observedAt.toISOString(),
  };
}

// Operator-managed, exact session bindings only: no names or account IDs are
// inferred from a prompt, a recalled document, or an assistant response.
export function parseIdentityBindings(value: unknown): Record<string, CaptureIdentity> {
  const raw = asRecord(value);
  const result: Record<string, CaptureIdentity> = Object.create(null);
  if (!raw) return result;
  for (const [session, entry] of Object.entries(raw).slice(0, 1000)) {
    const identity = asRecord(entry);
    if (!identity || !session.trim() || session.length > 1024) continue;
    const userId = identity.userId;
    const participants = identity.participantUserIds;
    const validID = (id: unknown): id is string => typeof id === "string" && !!id && id.trim() === id && id.length <= 1024;
    if (userId !== undefined && !validID(userId)) continue;
    if (participants !== undefined && (!Array.isArray(participants) || participants.length > 7 || !participants.every(validID))) continue;
    result[session] = {
      userId: validID(userId) ? userId : undefined,
      participantUserIds: Array.isArray(participants) ? [...new Set(participants as string[])] : undefined,
    };
  }
  return result;
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
