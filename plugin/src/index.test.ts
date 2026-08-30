import assert from "node:assert/strict";
import test from "node:test";

import plugin from "./index.js";

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
