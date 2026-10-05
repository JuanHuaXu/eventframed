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
assert.equal(hash(read("public_capture_load_test.go.txt")), manifest.fixtureSHA256);
assert.equal(hash(fs.readFileSync(path.join(root, "../frame-mask-cache-v1/candidate.patch"))), manifest.patchSHA256);
assert.equal(result.commands.length, 8);
assert.deepEqual(result.commands.map(c => [c.frontier, c.pair, c.arm]), [50,200].flatMap(k => [[0,["control","candidate"]],[1,["candidate","control"]]].flatMap(([pair,order]) => order.map(arm => [k,pair,arm]))));
const quantile = (values, fraction) => [...values].sort((a,b) => a-b)[Math.ceil(fraction*values.length)-1]/1e6;
for (const command of result.commands) {
  const bytes = read(command.transcript);
  assert.equal(hash(bytes), command.transcriptSHA256);
  const match = bytes.toString().match(/PUBLIC_CAPTURE_TRACE=(\{[^\n]+\})/);
  assert.ok(match, "missing complete trace");
  const trace = JSON.parse(match[1]);
  assert.deepEqual(trace, command.trace);
  assert.equal(trace.functional, true);
  for (const key of ["recall_ns", "recall_start_ns", "recall_end_ns", "feedback_ns", "live_age_ns", "candidate_counts", "packed_counts"]) assert.equal(trace[key].length, 64);
  for (const key of ["capture_ns", "capture_start_ns", "capture_end_ns"]) assert.equal(trace[key].length, 256);
  assert.ok(trace.candidate_counts.every(n => n === command.frontier));
  assert.ok(trace.packed_counts.every(n => n >= 0 && n <= 10));
  assert.equal(trace.completed, 64);
  assert.equal(trace.replay_completed, 64);
  assert.equal(trace.ledger_rows, 128);
  assert.ok(trace.init_ns > 0 && trace.init_bytes > 0 && trace.capture_bytes > 0);
  let overlaps = 0;
  for (let j=0;j<256;j++) {
    assert.equal(trace.capture_ns[j], trace.capture_end_ns[j]-trace.capture_start_ns[j]);
    if (trace.recall_start_ns.some((start,i) => trace.capture_start_ns[j] < trace.recall_end_ns[i] && trace.capture_end_ns[j] > start)) overlaps++;
  }
  for (let i=0;i<64;i++) assert.equal(trace.recall_ns[i], trace.recall_end_ns[i]-trace.recall_start_ns[i]);
  assert.equal(trace.overlaps, overlaps);
  assert.ok(overlaps > 0);
  for (const name of ["recall", "feedback", "live_age", "capture"]) {
    assert.ok(trace[name + "_ns"].every(n => Number.isSafeInteger(n) && n >= 0));
    for (const [key,fraction] of [["p50",.5],["p95",.95],["p99",.99],["max",1]]) assert.equal(command.measurements[name][key + "_ms"], quantile(trace[name + "_ns"],fraction));
  }
  const absolute = command.measurements.recall.p99_ms < 100 && command.measurements.live_age.p99_ms < 250;
  assert.equal(command.status === 0, absolute);
}
for (const paired of result.paired) {
  const get = arm => result.commands.find(c => c.frontier === paired.frontier && c.pair === paired.pair && c.arm === arm).measurements.recall.p99_ms;
  assert.equal(paired.ratio, get("candidate")/get("control"));
  assert.equal(paired.pass, paired.ratio <= 1.10);
}
assert.equal(result.functionalPass, result.commands.every(c => c.trace.functional));
assert.equal(result.absoluteTimingPass, result.commands.every(c => c.status === 0));
assert.equal(result.pairedNonRegressionPass, result.paired.every(p => p.pass));
assert.equal(result.wholeGoalValidation, false);
console.log(JSON.stringify({ verified:true, commands:8, functionalPass:result.functionalPass, absoluteTimingPass:result.absoluteTimingPass, pairedNonRegressionPass:result.pairedNonRegressionPass, wholeGoalValidation:false }));
