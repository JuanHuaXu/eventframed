import { definePluginEntry } from "openclaw/plugin-sdk/plugin-entry";
import { EventFrameClient } from "./client.js";
import { buildTurnEvent, extractLatestText } from "./event.js";
import { formatContext } from "./format.js";
import type { AdapterConfig } from "./types.js";

const DEFAULTS: AdapterConfig = {
  socketPath: "~/.eventframed/run/eventframed.sock",
  tenantId: "default",
  recallK: 50,
  packK: 10,
  tokenBudget: 2_000,
  capture: true,
};

type RecallState = {
  prompt: string;
  recalledIds: string[];
  occurredAt: Date;
};

const plugin: ReturnType<typeof definePluginEntry> = definePluginEntry({
  id: "eventframe-memory",
  name: "EventFrame Memory",
  description: "Availability-aware EventFrame recall backed by eventframed and LibraVDB.",
  kind: "memory",
  register(api) {
    const config = resolveConfig(api.pluginConfig);
    const client = new EventFrameClient({ socketPath: config.socketPath });
    const recallByRun = new Map<string, RecallState>();

    api.on("before_prompt_build", async (event, context) => {
      const prompt = normalizeText(extractLatestText(event.messages, "user") ?? event.prompt);
      if (!prompt) {
        return undefined;
      }
      try {
        const packet = await client.recall({
          tenantId: config.tenantId,
          sessionId: sessionId(context),
          query: prompt,
          recallK: config.recallK,
          packK: config.packK,
          tokenBudget: config.tokenBudget,
        });
        const key = runKey(context);
        if (key) {
          recallByRun.set(key, {
            prompt,
            recalledIds: packet.candidates.map((candidate) => candidate.event.id),
            occurredAt: new Date(),
          });
          if (recallByRun.size > 10_000) {
            const oldest = recallByRun.keys().next().value as string | undefined;
            if (oldest) recallByRun.delete(oldest);
          }
        }
        const prependContext = formatContext(packet);
        return prependContext ? { prependContext } : undefined;
      } catch (error) {
        api.logger.warn(`eventframe-memory: recall skipped: ${String(error)}`);
        return undefined;
      }
    });

    api.on("agent_end", async (event, context) => {
      if (!config.capture || !event.success) {
        return;
      }
      const state = takeRecallState(recallByRun, context, event.runId);
      const userText = normalizeText(state?.prompt ?? extractLatestText(event.messages, "user") ?? "");
      const assistantText = normalizeText(extractLatestText(event.messages, "assistant") ?? "");
      if (!userText || !assistantText) {
        return;
      }
      try {
        await client.observe(
          buildTurnEvent({
            tenantId: config.tenantId,
            sessionId: sessionId(context),
            runId: event.runId ?? context.runId,
            agentId: context.agentId,
            userText,
            assistantText,
            retrievedIds: state?.recalledIds ?? [],
            occurredAt: state?.occurredAt,
          }),
        );
      } catch (error) {
        api.logger.warn(`eventframe-memory: capture skipped: ${String(error)}`);
      }
    });
  },
});

export default plugin;

function resolveConfig(value: Record<string, unknown> | undefined): AdapterConfig {
  const recallK = readInteger(value?.recallK, DEFAULTS.recallK, 1, 1_000);
  return {
    socketPath: readString(value?.socketPath, DEFAULTS.socketPath),
    tenantId: readString(value?.tenantId, DEFAULTS.tenantId),
    recallK,
    packK: Math.min(recallK, readInteger(value?.packK, DEFAULTS.packK, 1, 100)),
    tokenBudget: readInteger(value?.tokenBudget, DEFAULTS.tokenBudget, 1, 1_000_000),
    capture: typeof value?.capture === "boolean" ? value.capture : DEFAULTS.capture,
  };
}

function sessionId(context: { sessionKey?: string; sessionId?: string }): string {
  return context.sessionKey ?? context.sessionId ?? "openclaw:unknown-session";
}

function runKey(context: { runId?: string; sessionKey?: string; sessionId?: string }, runId?: string): string | undefined {
  return runId ?? context.runId ?? context.sessionKey ?? context.sessionId;
}

function takeRecallState(
  states: Map<string, RecallState>,
  context: { runId?: string; sessionKey?: string; sessionId?: string },
  eventRunId?: string,
): RecallState | undefined {
  const keys = [eventRunId, context.runId, context.sessionKey, context.sessionId];
  for (const key of keys) {
    if (!key) continue;
    const state = states.get(key);
    if (state) {
      states.delete(key);
      return state;
    }
  }
  return undefined;
}

function normalizeText(value: string): string {
  return value.replace(/\s+/g, " ").trim().slice(0, 64_000);
}

function readString(value: unknown, fallback: string): string {
  return typeof value === "string" && value.trim() ? value : fallback;
}

function readInteger(value: unknown, fallback: number, minimum: number, maximum: number): number {
  return typeof value === "number" && Number.isInteger(value) && value >= minimum && value <= maximum
    ? value
    : fallback;
}
