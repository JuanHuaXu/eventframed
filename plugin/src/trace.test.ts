import assert from "node:assert/strict";
import { mkdtemp, readFile, stat } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { TraceWriter } from "./trace.js";

test("TraceWriter appends ordered mode-0600 JSONL records", async () => {
  const root = await mkdtemp(path.join(os.tmpdir(), "eventframe-trace-"));
  const tracePath = path.join(root, "nested", "trace.jsonl");
  const writer = new TraceWriter(tracePath);
  await Promise.all([writer.write({ type: "first", value: 1 }), writer.write({ type: "second", value: 2 })]);

  const records = (await readFile(tracePath, "utf8")).trim().split("\n").map((line) => JSON.parse(line));
  assert.deepEqual(records.map((record) => record.type), ["first", "second"]);
  assert.equal((await stat(tracePath)).mode & 0o777, 0o600);
});

test("TraceWriter is a no-op when no path is configured", async () => {
  await new TraceWriter().write({ type: "ignored" });
});
