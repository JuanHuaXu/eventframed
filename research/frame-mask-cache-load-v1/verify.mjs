import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import assert from "node:assert/strict";
import { fileURLToPath } from "node:url";

const root = path.dirname(fileURLToPath(import.meta.url));
const read = name => fs.readFileSync(path.join(root, name));
const hash = b => crypto.createHash("sha256").update(b).digest("hex");
const manifest = JSON.parse(read("manifest.json"));
const result = JSON.parse(read("results.json"));
assert.equal(hash(read("PROTOCOL.md")), manifest.protocolSHA256);
assert.equal(hash(read("run.mjs")), manifest.runnerSHA256);
assert.equal(hash(fs.readFileSync(path.join(root, "../frame-mask-cache-v1/candidate.patch"))), manifest.patchSHA256);
assert.equal(result.commands.length, 4);
const duration = value => {
  const match = value.match(/^([\d.]+)(ns|µs|us|ms|s)$/);
  assert.ok(match);
  return Number(match[1]) * { ns: .000001, "µs": .001, us: .001, ms: 1, s: 1000 }[match[2]];
};
for (const command of result.commands) {
  const text = read(command.transcript);
  assert.equal(hash(text), command.transcriptSHA256);
  assert.equal(command.status, 0);
  assert.match(text.toString(), /--- PASS: TestResearchGuardedDurableLiveFreshnessV2/);
  const lines = text.toString().split("\n").filter(line => /arm=(quiet|future-writer) calls=/.test(line));
  assert.equal(lines.length, 2);
  for (const line of lines) {
    const values = Object.fromEntries([...line.matchAll(/(\w+)=([^\s]+)/g)].map(m => [m[1],m[2]]));
    const recorded = command.measurements.find(row => row.arm === values.arm);
    for (const key of ["recall_p50", "recall_p95", "recall_p99", "feedback_p99", "live_age_p50", "live_age_p95", "live_age_p99", "live_age_max"]) assert.equal(recorded[key + "_ms"], duration(values[key]));
    assert.equal(recorded.calls, 192);
    assert.equal(recorded.labels, 192);
    assert.equal(recorded.writes, Number(values.writes));
    assert.equal(recorded.overlaps, Number(values.overlaps));
    if (values.arm === "future-writer") {
      assert.ok(recorded.recall_p99_ms < 100);
      assert.ok(recorded.live_age_p99_ms < 250);
    }
  }
}
for (const row of result.paired) {
  const get = arm => result.commands.find(c => c.pair === row.pair && c.arm === arm).measurements.find(m => m.arm === row.arm);
  const ratio = get("candidate").recall_p99_ms / get("control").recall_p99_ms;
  assert.equal(row.ratio, ratio);
  assert.equal(row.pass, ratio <= 1.10);
}
assert.equal(result.functionalAndAbsoluteGatePass, true);
assert.equal(result.pairedNonRegressionPass, result.paired.every(row => row.pass));
assert.equal(result.pairedNonRegressionPass, false);
assert.equal(result.wholeGoalValidation, false);
console.log(JSON.stringify({ verified: true, absolutePass: true, pairedNonRegressionPass: false, wholeGoalValidation: false }));
