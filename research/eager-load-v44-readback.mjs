// Supplementary descriptive readback of the fully audited owner-time repeat.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
import {readEnvelope} from './eager-load-v44-stream.mjs';
const root = 'research/eager-load-v44-owner-time', hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const doneBytes = fs.readFileSync(root + '/completed.json'), done = JSON.parse(doneBytes);
const auditBytes = fs.readFileSync(root + '/audit.json'), r = JSON.parse(auditBytes);
assert(done.sourceUnchanged && done.checks.every(c => c.code === 0));
assert.equal(done.passedFiniteTrials, 0); assert.equal(r.controls.length, 52); assert.equal(r.trials.length, 16);
const tick = value => { const m = value.match(/^(.*?)(?:\.(\d+))?Z$/); assert(m); return BigInt(Date.parse(m[1] + 'Z')) * 1000000n + BigInt((m[2] ?? '').padEnd(9, '0')); };
let duringOwnerWait = 0, reuse = 0, preparation = 0;
const envelope = await readEnvelope(root + '/raw.ndjson', row => {
  const builds = new Map(row.CoreBuilds.map(b => [b.Root, b]));
  for (const trace of row.Preparations ?? []) {
    preparation++;
    if (trace.Owner) assert(tick(trace.Begin) <= tick(trace.OwnerAt) && tick(trace.OwnerAt) <= tick(trace.End));
    if (trace.Reused) {
      const build = builds.get(trace.Root); assert(build); reuse++;
      assert(tick(build.At) <= tick(trace.OwnerAt), 'no future core at actual owner acquisition');
      if (tick(trace.Begin) < tick(build.At)) duringOwnerWait++;
    }
  }
});
assert.equal(envelope.rawSHA256, done.rawSHA256); assert.equal(envelope.rawSHA256, r.rawSHA256);
const cells = [];
for (const visible of [false, true]) for (const archived of [false, true]) for (const eager of [false, true]) {
  const trials = r.trials.filter(t => t.visible === visible && t.archived === archived && t.eager === eager);
  assert.equal(trials.length, 2);
  const range = key => { const values = trials.map(t => t.metrics[key] / 1e6); return {minimum: Math.min(...values), maximum: Math.max(...values)}; };
  cells.push({visible, archived, eager, passed: trials.filter(t => t.pass).length,
    callP99MS: range('call_p99_ns'), offerP99MS: range('offer_p99_ns'), outcomeP99MS: range('outcome_p99_ns'),
    writeP99MS: range('write_p99_ns'), physicalBytes: trials.map(t => t.core.PhysicalBytes),
    unusedPublisherBuilds: trials.map(t => t.publication.unused)});
}
const failedGates = Object.fromEntries(Object.entries({call: ['call_p99_ns', 100e6], offer: ['offer_p99_ns', 100e6],
  write: ['write_p99_ns', 250e6], outcome: ['outcome_p99_ns', 100e6], outcomeMax: ['outcome_max_ns', 250e6],
  view: ['view_max_ns', 250e6]}).map(([key, [metric, bound]]) => [key, r.trials.filter(t => !(t.metrics[metric] < bound)).length]));
const report = {time: new Date().toISOString(), completedSHA256: hash(doneBytes), auditSHA256: hash(auditBytes),
  scriptSHA256: hash(fs.readFileSync('research/eager-load-v44-readback.mjs')), rawSHA256: envelope.rawSHA256,
  cells, failedGates, preparations: preparation, reusedCores: reuse, actualBuildsDuringOwnerWait: duringOwnerWait,
  allReuseAvailableAtOwnerAcquisition: true, rowsRetained: envelope.count,
  fixedOrderTwoRepeatsNotPopulationCertification: true, noGateChanges: true,
  passedFiniteTrials: 0, allSevenWholeGoals: 'OPEN', goal: 'ACTIVE'};
fs.writeFileSync(root + '/readback.json', JSON.stringify(report, null, 2) + '\n', {flag: 'wx', mode: 0o600});
console.log(JSON.stringify(report, null, 2));
