import assert from "node:assert/strict";
import fs from "node:fs/promises";
import http from "node:http";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { EventFrameClient } from "./client.js";

test("recall rejects a daemon response with the wrong protocol", async (t) => {
  const fixture = await unixServer(t, (_request, response) => {
    response.setHeader("content-type", "application/json");
    response.end(JSON.stringify({ protocol_version: "eventframe.v0", candidates: [] }));
  });
  const client = new EventFrameClient({ socketPath: fixture.socketPath });
  await assert.rejects(
    client.recall({
      tenantId: "tenant",
      sessionId: "session",
      query: "query",
      recallK: 50,
      packK: 10,
      tokenBudget: 2_000,
    }),
    /unsupported or missing protocol_version/,
  );
});

test("recall rejects malformed candidates before prompt formatting", async (t) => {
  const fixture = await unixServer(t, (_request, response) => {
    response.setHeader("content-type", "application/json");
    response.end(JSON.stringify({ protocol_version: "eventframe.v1alpha1", candidates: [{ nope: true }] }));
  });
  const client = new EventFrameClient({ socketPath: fixture.socketPath });
  await assert.rejects(
    client.recall({
      tenantId: "tenant",
      sessionId: "session",
      query: "query",
      recallK: 50,
      packK: 10,
      tokenBudget: 2_000,
    }),
    /malformed candidate/,
  );
});

async function unixServer(
  t: test.TestContext,
  handler: http.RequestListener,
): Promise<{ socketPath: string }> {
  const directory = await fs.mkdtemp(path.join(os.tmpdir(), "eventframe-client-"));
  const socketPath = path.join(directory, "daemon.sock");
  const server = http.createServer(handler);
  await new Promise<void>((resolve, reject) => {
    server.once("error", reject);
    server.listen(socketPath, resolve);
  });
  t.after(async () => {
    await new Promise<void>((resolve, reject) => server.close((error) => (error ? reject(error) : resolve())));
    await fs.rm(directory, { recursive: true, force: true });
  });
  return { socketPath };
}
