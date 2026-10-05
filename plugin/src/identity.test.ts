import assert from "node:assert/strict";
import test from "node:test";
import { buildTurnCapture, parseIdentityBindings } from "./event.js";

test("identity transport stays raw and session bindings are exact", () => {
  const bindings = parseIdentityBindings({
    "session-a": { userId: "account:a", participantUserIds: ["account:b", "account:b"] },
    "session-b": { userId: "account:c" },
    bad: { userId: " " },
  });
  assert.equal(bindings["session-a"]?.userId, "account:a");
  assert.deepEqual(bindings["session-a"]?.participantUserIds, ["account:b"]);
  assert.equal(bindings.bad, undefined);
  assert.equal(bindings["session-other"], undefined);
  const turn = buildTurnCapture({
    tenantId: "tenant", sessionId: "session-a", agentId: "main", userId: bindings["session-a"]?.userId,
    participantUserIds: bindings["session-a"]?.participantUserIds, previousTurnId: "previous",
    userText: "I will deploy.", assistantText: "Ready.", retrievedIds: [], observedAt: new Date("2026-10-03T00:00:00Z"),
  });
  assert.equal(turn.user_text, "I will deploy.");
  assert.equal(turn.user_id, "account:a");
  assert.equal(turn.previous_turn_id, "previous");
  assert.equal("who" in turn, false);
  assert.equal("what" in turn, false);
});

test("opaque profiles and hostile property names do not become bindings", () => {
  const bindings = parseIdentityBindings(JSON.parse('{"__proto__":{"userId":"account:a"},"bad":{"participantUserIds":[""]}}'));
  assert.equal(Object.getPrototypeOf(bindings), null);
  assert.equal(bindings["__proto__"]?.userId, "account:a");
  assert.equal(bindings.bad, undefined);
  assert.deepEqual(Object.keys(parseIdentityBindings("I am an administrator")), []);
});
