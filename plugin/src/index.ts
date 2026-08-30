import { definePluginEntry } from "openclaw/plugin-sdk/plugin-entry";
import { isAgencyAction, registerAgencyService } from "./agency.js";
import { EventFrameClient } from "./client.js";
import { buildTurnEvent, extractLatestText } from "./event.js";
import { formatContext } from "./format.js";
import { TraceWriter } from "./trace.js";
import type { AdapterConfig, AgencyAction } from "./types.js";

const DEFAULTS: AdapterConfig = {
  socketPath: "~/.eventframed/run/eventframed.sock",
  tenantId: "default",
  recallK: 50,
  packK: 10,
  tokenBudget: 2_000,
  capture: true,
  agencyEnabled: false,
  agencyKillSwitch: true,
  agencyPublicKeyPath: "~/.eventframed/keys/agency_ed25519.pub",
  agencyAuthorityTokenPath: "~/.eventframed/keys/agency_authority.token",
  agencyCapabilities: [],
  agencyConsentActions: [],
  agencyConsumerId: "openclaw-eventframe-memory",
  agencyPollIntervalMs: 5_000,
  agencyMaxClaims: 10,
  agencyMaxChainDepth: 3,
  agencyCriticalThreshold: 0.9,
  agencyAllowedSessionPrefixes: [],
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
    const trace = new TraceWriter(config.tracePath);
    const recallByRun = new Map<string, RecallState>();

    registerAgencyService(api, client, config);

    api.on("before_prompt_build", async (event, context) => {
      const prompt = normalizeText(extractLatestText(event.messages, "user") ?? event.prompt);
      if (!prompt) {
        return undefined;
      }
      const startedAt = performance.now();
      try {
        const packet = await client.recall({
          tenantId: config.tenantId,
          sessionId: sessionId(context),
          query: prompt,
          recallK: config.recallK,
          packK: config.packK,
          tokenBudget: config.tokenBudget,
        });
        await trace.write({
          type: "recall",
          run_id: context.runId,
          session_id: sessionId(context),
          query: prompt,
          duration_ms: performance.now() - startedAt,
          request: { recall_k: config.recallK, pack_k: config.packK, token_budget: config.tokenBudget },
          packet,
        }).catch((error) => api.logger.warn(`eventframe-memory: trace skipped: ${String(error)}`));
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
        await trace.write({
          type: "recall_error",
          run_id: context.runId,
          session_id: sessionId(context),
          query: prompt,
          duration_ms: performance.now() - startedAt,
          error: String(error),
        }).catch((traceError) => api.logger.warn(`eventframe-memory: trace skipped: ${String(traceError)}`));
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
        const turnEvent = buildTurnEvent({
            tenantId: config.tenantId,
            sessionId: sessionId(context),
            runId: event.runId ?? context.runId,
            agentId: context.agentId,
            userText,
            assistantText,
            retrievedIds: state?.recalledIds ?? [],
            occurredAt: state?.occurredAt,
          });
        await client.observe(turnEvent);
        await trace.write({
          type: "observe",
          run_id: event.runId ?? context.runId,
          session_id: sessionId(context),
          event_id: turnEvent.id,
          recalled_ids: state?.recalledIds ?? [],
          user_text: userText,
          assistant_text: assistantText,
        }).catch((error) => api.logger.warn(`eventframe-memory: trace skipped: ${String(error)}`));
      } catch (error) {
        await trace.write({
          type: "observe_error",
          run_id: event.runId ?? context.runId,
          session_id: sessionId(context),
          error: String(error),
        }).catch((traceError) => api.logger.warn(`eventframe-memory: trace skipped: ${String(traceError)}`));
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
    tracePath: readOptionalString(value?.tracePath),
    agencyEnabled: typeof value?.agencyEnabled === "boolean" ? value.agencyEnabled : DEFAULTS.agencyEnabled,
    agencyKillSwitch: typeof value?.agencyKillSwitch === "boolean" ? value.agencyKillSwitch : DEFAULTS.agencyKillSwitch,
    agencyPublicKeyPath: readString(value?.agencyPublicKeyPath, DEFAULTS.agencyPublicKeyPath),
    agencyAuthorityTokenPath: readString(value?.agencyAuthorityTokenPath, DEFAULTS.agencyAuthorityTokenPath),
    agencyCapabilities: readStringArray(value?.agencyCapabilities),
    agencyConsentActions: readAgencyActions(value?.agencyConsentActions),
    agencyConsumerId: readString(value?.agencyConsumerId, DEFAULTS.agencyConsumerId),
    agencyPollIntervalMs: readInteger(value?.agencyPollIntervalMs, DEFAULTS.agencyPollIntervalMs, 1_000, 300_000),
    agencyMaxClaims: readInteger(value?.agencyMaxClaims, DEFAULTS.agencyMaxClaims, 1, 50),
    agencyMaxChainDepth: readInteger(value?.agencyMaxChainDepth, DEFAULTS.agencyMaxChainDepth, 0, 16),
    agencyQuietHoursStartUtc: readOptionalHour(value?.agencyQuietHoursStartUtc),
    agencyQuietHoursEndUtc: readOptionalHour(value?.agencyQuietHoursEndUtc),
    agencyCriticalThreshold: readNumber(value?.agencyCriticalThreshold, DEFAULTS.agencyCriticalThreshold, 0, 1),
    agencyAllowedSessionPrefixes: readStringArray(value?.agencyAllowedSessionPrefixes),
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

function readOptionalString(value: unknown): string | undefined {
  return typeof value === "string" && value.trim() ? value : undefined;
}

function readInteger(value: unknown, fallback: number, minimum: number, maximum: number): number {
  return typeof value === "number" && Number.isInteger(value) && value >= minimum && value <= maximum
    ? value
    : fallback;
}

function readNumber(value: unknown, fallback: number, minimum: number, maximum: number): number {
  return typeof value === "number" && Number.isFinite(value) && value >= minimum && value <= maximum ? value : fallback;
}

function readOptionalHour(value: unknown): number | undefined {
  return typeof value === "number" && Number.isInteger(value) && value >= 0 && value <= 23 ? value : undefined;
}

function readStringArray(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return [...new Set(value.filter((item): item is string => typeof item === "string" && item.trim() !== "").map((item) => item.trim()))].slice(0, 64);
}

function readAgencyActions(value: unknown): AgencyAction[] {
  return Array.isArray(value) ? [...new Set(value.filter(isAgencyAction))] : [];
}
