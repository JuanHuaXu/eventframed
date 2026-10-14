import { execFile } from "node:child_process";
import path from "node:path";
import { promisify } from "node:util";
import type { OpenClawPluginApi } from "openclaw/plugin-sdk/plugin-entry";

const execute = promisify(execFile);
export type OperatorGateway = { cliPath: string; configPath: string; stateDir: string; url: string; silent?: boolean; model?: string };

// This is an explicitly configured operator client. It uses the existing
// gateway's public cron RPCs and never pretends to be a bundled plugin.
export function operatorScheduler(options: OperatorGateway): OpenClawPluginApi["session"]["workflow"] {
  const url = new URL(options.url);
  if (url.protocol !== "ws:" || !["127.0.0.1", "[::1]"].includes(url.hostname) || url.username || url.password || url.search || url.hash || (url.pathname !== "/" && url.pathname !== "")) throw new Error("operator scheduler requires an explicit loopback gateway");
  for (const p of [options.cliPath, options.configPath, options.stateDir]) if (!path.isAbsolute(p)) throw new Error("operator gateway paths must be absolute");
  const call = async (method: "cron.add" | "cron.list" | "cron.remove", params: unknown): Promise<any> => {
    const result = await execute(process.execPath, [options.cliPath, "gateway", "call", method, "--expect-url", options.url, "--json", "--params", JSON.stringify(params), "--timeout", "10000"], {
      env: { ...process.env, OPENCLAW_STATE_DIR: options.stateDir, OPENCLAW_CONFIG_PATH: options.configPath }, timeout: 15000, maxBuffer: 2 << 20,
    });
    return JSON.parse(result.stdout);
  };
  return {
    async scheduleSessionTurn(input) {
      if (!input.tag || !/^efagency-[a-f0-9]{20}$/.test(input.tag) || !("at" in input) || !(input.at instanceof Date) || !Number.isFinite(input.at.getTime())) throw new Error("bounded agency tag and one-shot time required");
      const job = await call("cron.add", {
        name: `eventframe:${input.tag}`, enabled: true, schedule: { kind: "at", at: input.at.toISOString() }, sessionTarget: `session:${input.sessionKey}`,
        payload: { kind: "agentTurn", message: input.message, ...(options.model ? {model: options.model} : {}), timeoutSeconds: 60 }, deleteAfterRun: true, wakeMode: "now", delivery: { mode: options.silent ? "none" : input.deliveryMode ?? "none" },
      });
      if (typeof job.id !== "string" || !job.id) throw new Error("gateway did not acknowledge a job ID");
      return { id: job.id };
    },
    async unscheduleSessionTurnsByTag(input) {
      if (!/^efagency-[a-f0-9]{20}$/.test(input.tag)) throw new Error("invalid agency cleanup tag");
      const name = `eventframe:${input.tag}`;
      let offset = 0;
      let revision: unknown;
      const matches: string[] = [];
      for (let page = 0; ; page++) {
        if (page >= 6) throw new Error("agency cleanup inventory exceeds bound");
        const result = await call("cron.list", { includeDisabled: true, query: name, limit: 200, offset, sortBy: "name", sortDir: "asc" });
        if (!Array.isArray(result.jobs) || result.jobs.length > 200 || (revision !== undefined && revision !== result.snapshotRevision)) throw new Error("invalid or changing cleanup inventory");
        revision = result.snapshotRevision;
        for (const job of result.jobs) if (job.name === name && job.sessionTarget === `session:${input.sessionKey}` && typeof job.id === "string") matches.push(job.id);
        if (!result.hasMore) break;
        if (!Number.isInteger(result.nextOffset) || result.nextOffset <= offset) throw new Error("invalid cleanup pagination");
        offset = result.nextOffset;
      }
      for (const id of matches) { const result = await call("cron.remove", { id }); if (!result.ok || !result.removed) throw new Error("agency scheduler cancellation was not acknowledged"); }
      return { removed: matches.length, failed: 0 };
    },
  } as OpenClawPluginApi["session"]["workflow"];
}
