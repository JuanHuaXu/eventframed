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

test("recall rejects a candidate without a finite rank delta", async (t) => {
  const fixture = await unixServer(t, (_request, response) => {
    response.setHeader("content-type", "application/json");
    response.end(JSON.stringify({
      protocol_version: "eventframe.v1alpha1",
      candidates: [{ event: { id: "event", content: "text", kind: "turn", available_at: "2026-01-01T00:00:00Z", provenance: { producer: "test" } }, retrieval_score: .5, score: .5 }],
    }));
  });
  const client = new EventFrameClient({ socketPath: fixture.socketPath });
  await assert.rejects(client.recall({
    tenantId: "tenant", sessionId: "session", query: "query", recallK: 50, packK: 10, tokenBudget: 2_000,
  }), /malformed candidate/);
});

test("agency claims reject a lease assigned to another consumer", async (t) => {
  const fixture = await unixServer(t, (_request, response) => {
    response.setHeader("content-type", "application/json");
    response.end(JSON.stringify({
      protocol_version: "eventframe.v1alpha1",
      records: [{
        proposal: { id: "proposal-1", tenant_id: "tenant" },
        signed: { payload: "payload", signature: "signature", key_id: "key" },
        status: "claimed",
        claimed_by: "different-authority",
        lease_until: "2026-08-28T22:00:00Z",
      }],
      snapshot: {},
    }));
  });
  const client = new EventFrameClient({ socketPath: fixture.socketPath });
  await assert.rejects(
    client.claimAgency({ authorityToken: "test-token", tenantId: "tenant", consumerId: "expected-authority", limit: 1 }),
    /malformed agency proposal record/,
  );
});

test("agency claim sends the private authority credential", async (t) => {
  let requestBody = "";
  const fixture = await unixServer(t, (request, response) => {
    request.on("data", (chunk: Buffer) => { requestBody += chunk.toString("utf8"); });
    request.on("end", () => {
      response.setHeader("content-type", "application/json");
      response.end(JSON.stringify({ protocol_version: "eventframe.v1alpha1", records: [], snapshot: {} }));
    });
  });
  const client = new EventFrameClient({ socketPath: fixture.socketPath });
  await client.claimAgency({ authorityToken: "authority-secret", tenantId: "tenant", consumerId: "authority", limit: 1 });
  assert.equal(JSON.parse(requestBody).authority_token, "authority-secret");
});

test("agency resolution rejects a terminal state that differs from the request", async (t) => {
  const fixture = await unixServer(t, (_request, response) => {
    response.setHeader("content-type", "application/json");
    response.end(JSON.stringify({
      protocol_version: "eventframe.v1alpha1",
      record: { proposal: { id: "proposal-1", tenant_id: "tenant" }, status: "expired" },
      snapshot: {},
    }));
  });
  const client = new EventFrameClient({ socketPath: fixture.socketPath });
  await assert.rejects(client.resolveAgency({
    authorityToken: "authority-secret",
    tenantId: "tenant",
    proposalId: "proposal-1",
    consumerId: "authority",
    decision: "approved",
    reason: "approved",
    executionRef: "job-1",
  }), /malformed agency resolution/);
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
