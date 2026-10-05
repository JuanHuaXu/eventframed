import assert from "node:assert/strict";
import fs from "node:fs/promises";
import http from "node:http";
import os from "node:os";
import path from "node:path";
import test from "node:test";

import plugin from "./index.js";
import type { CapturedTurn } from "./types.js";

test("capture hook forwards trusted session identity and preceding successful capture only", async (t) => {
  const directory = await fs.mkdtemp(path.join(os.tmpdir(), "eventframe-identity-"));
  const socketPath = path.join(directory, "daemon.sock");
  const captures: CapturedTurn[] = [];
  let releaseHeld: (() => void) | undefined;
  let signalHeld: (() => void) | undefined;
  const heldReceived = new Promise<void>((resolve) => { signalHeld = resolve; });
  const server = http.createServer((request, response) => {
    let body = "";
    request.on("data", (chunk: Buffer) => { body += chunk.toString("utf8"); });
    request.on("end", () => {
      assert.equal(request.url, "/v1/turns:capture");
      const { turn } = JSON.parse(body) as { turn: CapturedTurn };
      captures.push(turn);
      response.setHeader("content-type", "application/json");
      const finish = () => response.end(JSON.stringify({ protocol_version: "eventframe.v1alpha1", event_id: turn.id }));
      if (turn.user_text === "Blair will deploy.") { releaseHeld = () => { releaseHeld = undefined; finish(); }; signalHeld?.(); }
      else finish();
    });
  });
  await new Promise<void>((resolve, reject) => { server.once("error", reject); server.listen(socketPath, resolve); });
  t.after(async () => { releaseHeld?.(); await new Promise<void>((resolve) => server.close(() => resolve())); await fs.rm(directory, { recursive: true, force: true }); });
  let end: ((event: unknown, context: unknown) => Promise<void>) | undefined;
  const warnings: string[] = [];
  const api = {
    registrationMode: "full",
    pluginConfig: {
      socketPath, tenantId: "tenant", capture: true, agencyEnabled: false, agencyKillSwitch: true,
      identityBySession: { "session-a": { userId: "account:a", participantUserIds: ["account:b"] } },
    },
    logger: { warn(message: string) { warnings.push(message); } },
    registerGatewayMethod() {},
    on(name: string, handler: typeof end) { if (name === "agent_end") end = handler; },
  };
  plugin.register(api as never);
  assert.ok(end);
  for (const [run, session, text] of [["one", "session-a", "Alex will deploy."], ["two", "session-b", "I will deploy."], ["three", "session-a", "He will use his console."]]) {
    await end({ success: true, runId: run, messages: [{ role: "user", content: text }, { role: "assistant", content: "Ready." }] }, { sessionKey: session, agentId: "main" });
    await new Promise((resolve) => setTimeout(resolve, 5));
  }
  assert.deepEqual(warnings, []);
  assert.equal(captures.length, 3);
  assert.equal(captures[0]?.user_id, "account:a");
  assert.equal(captures[1]?.user_id, undefined);
  assert.equal(captures[1]?.previous_turn_id, undefined);
  assert.equal(captures[2]?.previous_turn_id, captures[0]?.id);
  assert.deepEqual(captures[2]?.participant_user_ids, ["account:b"]);
  assert.equal("who" in captures[2]!, false);
  const older = end({ success: true, runId: "four", messages: [{ role: "user", content: "Blair will deploy." }, { role: "assistant", content: "Ready." }] }, { sessionKey: "session-a", agentId: "main" });
  await heldReceived;
  await new Promise((resolve) => setTimeout(resolve, 5));
  await end({ success: true, runId: "five", messages: [{ role: "user", content: "Alex will review." }, { role: "assistant", content: "Ready." }] }, { sessionKey: "session-a", agentId: "main" });
  assert.equal(captures[4]?.previous_turn_id, undefined, "overlapping capture must not borrow an older focus");
  releaseHeld?.();
  await older;
  await new Promise((resolve) => setTimeout(resolve, 5));
  await end({ success: true, runId: "six", messages: [{ role: "user", content: "He will review." }, { role: "assistant", content: "Ready." }] }, { sessionKey: "session-a", agentId: "main" });
  assert.equal(captures[5]?.previous_turn_id, captures[4]?.id, "late completion must not rewind context");
});

test("before_prompt_build fails open when eventframed is unavailable", async () => {
  let beforePrompt: ((event: unknown, context: unknown) => Promise<unknown>) | undefined;
  const warnings: string[] = [];
  const api = {
    registrationMode: "full",
    pluginConfig: {
      socketPath: `/tmp/eventframed-intentionally-absent-${process.pid}.sock`,
      tenantId: "rc1-failopen",
      capture: false,
      agencyEnabled: false,
      agencyKillSwitch: true,
    },
    logger: { warn: (message: string) => warnings.push(message) },
    registerGatewayMethod() {},
    on(name: string, handler: (event: unknown, context: unknown) => Promise<unknown>) {
      if (name === "before_prompt_build") beforePrompt = handler;
    },
  };

  plugin.register(api as never);
  assert.ok(beforePrompt);
  const result = await beforePrompt(
    { prompt: "continue without memory", messages: [{ role: "user", content: "continue without memory" }] },
    { runId: "rc1-run", sessionKey: "agent:main:rc1-failopen" },
  );

  assert.equal(result, undefined);
  assert.ok(warnings.some((message) => message.includes("recall skipped")));
});

test("registers operator-scoped exact-journal outcome feedback", async () => {
  type Handler = (input: { params: Record<string, unknown>; respond: (...args: unknown[]) => void }) => Promise<void>;
  let registered: { handler: Handler; scope?: string } | undefined;
  const api = {
    registrationMode: "full",
    pluginConfig: { tenantId: "tenant", capture: false, agencyEnabled: false, agencyKillSwitch: true },
    logger: { warn() {} },
    on() {},
    registerGatewayMethod(_name: string, handler: Handler, options: { scope?: string }) {
      registered = { handler, scope: options.scope };
    },
  };
  plugin.register(api as never);
  assert.ok(registered);
  assert.equal(registered.scope, "operator.write");

  const responses: unknown[][] = [];
  await registered.handler({
    params: { journal_id: "journal", event_id: "event", idempotency_key: "feedback", signal: "useful" },
    respond: (...args: unknown[]) => { responses.push(args); },
  });
  assert.equal(responses[0]?.[0], false);
  assert.match(String((responses[0]?.[2] as { message?: string })?.message), /eventframed/);
});
