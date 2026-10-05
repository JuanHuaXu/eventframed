import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import assert from "node:assert/strict";
import { fileURLToPath } from "node:url";

const dir = path.dirname(fileURLToPath(import.meta.url));
const read = name => fs.readFileSync(path.join(dir, name));
const hash = bytes => crypto.createHash("sha256").update(bytes).digest("hex");
const result = JSON.parse(read("results.json")), manifest = JSON.parse(read("manifest.json"));
assert.equal(manifest.patchSHA256, hash(read("candidate.patch")));
assert.equal(manifest.protocolSHA256, hash(read("PROTOCOL.md")));
assert.equal(manifest.additionalTestsSHA256, hash(read("mask_reuse_test.go.txt")));
assert.equal(result.wholeGoalValidation, false);
assert.equal(result.loadedLatencyTested, false);
assert.equal(result.scientificConfirmation, false);
assert.equal(result.semanticComparisons, 8192);
assert.equal(result.parallelCallComparisons, 384);
const samples = new Map();
for (const match of read("benchmarks.txt").toString().matchAll(/^BenchmarkMaskReuse(?:Forward|Reverse)\/(\w+)\/(early|late|no-match|quoted|collective)\/(256|2048|16384)B\/(control|candidate)(?:-\d+)?\s+\d+\s+([\d.]+) ns\/op\s+[\d.]+ MB\/s\s+(\d+) B\/op\s+(\d+) allocs\/op/gm)) {
  const key = match.slice(1, 5).join("/");
  if (!samples.has(key)) samples.set(key, []);
  samples.get(key).push([Number(match[5]), Number(match[6]), Number(match[7])]);
}
assert.equal(samples.size, 90);
const median = rows => { const sorted = rows.toSorted((a, b) => a - b); return (sorted[2] + sorted[3]) / 2; };
const factors = { turn: [], text: [], query: [] };
for (const row of result.cells) {
  const control = samples.get(row.cell + "/control"), candidate = samples.get(row.cell + "/candidate");
  assert.equal(control.length, 6); assert.equal(candidate.length, 6);
  assert.equal(row.oldNS, median(control.map(v => v[0])));
  assert.equal(row.newNS, median(candidate.map(v => v[0])));
  assert.equal(row.oldBytes, median(control.map(v => v[1])));
  assert.equal(row.newBytes, median(candidate.map(v => v[1])));
  assert.equal(row.oldAllocations, median(control.map(v => v[2])));
  assert.equal(row.newAllocations, median(candidate.map(v => v[2])));
  assert.ok(Math.abs(row.changePercent - 100 * (row.newNS / row.oldNS - 1)) < 1e-12);
  factors[row.cell.split("/")[0]].push(row.newNS / row.oldNS);
}
assert.equal(result.cells.length, 45);
const passFlags = Object.entries(factors).map(([workflow, values]) => {
  assert.equal(values.length, 15);
  const reduction = 100 * (1 - Math.exp(values.reduce((sum, v) => sum + Math.log(v), 0) / values.length));
  assert.ok(Math.abs(reduction - result.workflows[workflow].geometricMeanImprovementPercent) < 1e-12);
  const pass = reduction >= 5 && values.every(v => 100 * (v - 1) <= 10);
  assert.equal(pass, result.workflows[workflow].pass);
  return pass;
});
const allPass = passFlags.every(Boolean);
assert.equal(allPass, result.preflightPass);
assert.match(read("control-tests.txt").toString(), /^ok\s/gm);
assert.match(read("candidate-tests.txt").toString(), /^PASS$/gm);
assert.match(read("candidate-race.txt").toString(), /^ok\s/gm);
const integration = JSON.parse(read("service-integration-manifest.json"));
assert.equal(integration.patchSHA256, manifest.patchSHA256);
assert.equal(integration.transcriptSHA256, hash(read("service-integration.txt")));
assert.equal(integration.loadedLatencyTested, false);
for (const [name, digest] of Object.entries(integration.sources)) {
  assert.equal(digest, manifest.sources.candidate["internal/frame/" + name]);
}
assert.equal((read("service-integration.txt").toString().match(/^--- PASS: Test/gm) ?? []).length, 5);
console.log(JSON.stringify({ verified: true, cells: 45, samplesPerArm: 6, preflightPass: allPass,
  loadedServingValidated: false, wholeGoalValidation: false }));
