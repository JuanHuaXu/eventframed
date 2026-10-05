import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';

const path = process.argv[2];
if (!path) throw Error('artifact path required');
const lines = fs.readFileSync(path, 'utf8').trim().split('\n').map(JSON.parse);
const header = lines.shift();
const summary = lines.pop();
assert.equal(header.kind, 'header');
assert.equal(summary.kind, 'summary');
for (const [name, hash] of [
  ['research/conditional-gate-v1.mjs', header.sourceSHA256],
  ['docs/experiments/mmm-conditional-gate-v1-protocol.md', header.protocolSHA256],
]) {
  assert.equal(crypto.createHash('sha256').update(fs.readFileSync(name)).digest('hex'), hash);
}
assert.equal(lines.length, 4000);
const seen = new Set();
for (const r of lines) {
  assert.equal(r.kind, 'trial');
  const key = `${r.schedule}/${r.regime}/${r.trial}`;
  assert(!seen.has(key));
  seen.add(key);
  assert(r.trial >= 0 && r.trial < 1000);
  assert(r.nominated >= 0 && r.nominated <= 512);
  assert(r.missing >= 0 && r.missing <= r.nominated);
  assert(r.delivered >= 0 && r.delivered <= r.nominated - r.missing);
  assert.equal(r.pending, r.nominated - r.missing - r.delivered);
  assert.equal(r.observedByCell[0] + r.observedByCell[1], r.delivered);
  for (const [clock, labels] of [[r.scalarClock, r.scalarLabels], [r.conditionalClock, r.conditionalLabels]]) {
    if (clock === null) assert.equal(labels, null);
    else {
      assert(Number.isInteger(clock) && clock >= 0 && clock < 512);
      assert(Number.isInteger(labels) && labels > 0 && labels <= r.delivered);
    }
  }
}
function q(xs, p) {
  if (!xs.length) return null;
  return xs.sort((a, b) => a - b)[Math.ceil(xs.length * p) - 1];
}
const cells = [];
for (const schedule of header.schedules) for (const regime of ['stable', 'swapped']) {
  const rs = lines.filter(r => r.schedule === schedule.name && r.regime === regime);
  assert.equal(rs.length, 1000);
  const scalar = rs.filter(r => r.scalarClock !== null);
  const conditional = rs.filter(r => r.conditionalClock !== null);
  const mean = k => rs.reduce((v, r) => v + r[k], 0) / rs.length;
  const cell = {
    schedule: schedule.name, regime, trials: rs.length,
    scalarFlags: scalar.length, conditionalFlags: conditional.length,
    meanNominated: mean('nominated'), meanMissing: mean('missing'),
    meanDelivered: mean('delivered'), meanPending: mean('pending'),
    scalarClockP50: q(scalar.map(r => r.scalarClock), .5),
    conditionalClockP50: q(conditional.map(r => r.conditionalClock), .5),
    conditionalClockP95: q(conditional.map(r => r.conditionalClock), .95),
    conditionalLabelsP50: q(conditional.map(r => r.conditionalLabels), .5),
  };
  cell.pass = regime === 'stable'
    ? scalar.length <= 20 && conditional.length <= 20
    : scalar.length <= 20 && conditional.length >= 900;
  cells.push(cell);
}
assert.deepEqual(cells, summary.cells);
assert.equal(summary.pass, cells.every(c => c.pass));
console.log(JSON.stringify({ rows: lines.length, pass: summary.pass, cells }, null, 2));
