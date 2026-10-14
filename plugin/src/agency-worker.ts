import fs from "node:fs/promises";
import { pathToFileURL } from "node:url";
import type { OpenClawPluginApi } from "openclaw/plugin-sdk/plugin-entry";
import { registerAgencyService } from "./agency.js";
import { EventFrameClient } from "./client.js";
import { operatorScheduler, type OperatorGateway } from "./operator-scheduler.js";
import type { AdapterConfig } from "./types.js";

// Some OpenClaw versions reserve the in-process scheduling surface for bundled
// plugins. The operator may run this scoped authority worker beside the daemon.
export async function startAgencyWorker(config: AdapterConfig, gateway: OperatorGateway): Promise<() => void> {
  if (!config.agencyEnabled || config.agencyKillSwitch || !config.tenantId || !Number.isInteger(config.agencyPollIntervalMs) || config.agencyPollIntervalMs < 1000 || config.agencyPollIntervalMs > 300000 || !Number.isInteger(config.agencyMaxClaims) || config.agencyMaxClaims < 1 || config.agencyMaxClaims > 50 || !config.agencyAllowedSessionPrefixes?.length) throw new Error("explicit bounded authority configuration required");
  let service: { start(): Promise<void>; stop(): void } | undefined;
  const api = { registrationMode: "full", logger: { warn: (message: string) => console.warn(message) }, session: { workflow: operatorScheduler(gateway) }, registerService(value: typeof service) { service = value; } } as unknown as OpenClawPluginApi;
  registerAgencyService(api, new EventFrameClient({ socketPath: config.socketPath }), config);
  if (!service) throw new Error("authority worker was not registered");
  await service.start();
  return () => service?.stop();
}

if (process.argv[1] && pathToFileURL(process.argv[1]).href === import.meta.url) {
  const file = process.argv[2];
  if (!file?.startsWith("/")) throw new Error("usage: agency-worker /absolute/operator-config.json");
  const input = JSON.parse(await fs.readFile(file, "utf8")) as { adapter: AdapterConfig; gateway: OperatorGateway };
  const stop = await startAgencyWorker(input.adapter, input.gateway);
  process.on("SIGINT", () => { stop(); process.exitCode = 0; });
  process.on("SIGTERM", () => { stop(); process.exitCode = 0; });
  // registerAgencyService intentionally unrefs plugin timers; a standalone
  // authority owns its own lifetime instead of relying on the gateway host.
  const alive = setInterval(() => {}, 60000);
  process.on("SIGINT", () => clearInterval(alive));
  process.on("SIGTERM", () => clearInterval(alive));
}
