import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';

const path = process.argv[2];
if (!path) throw Error('artifact path required');
const rows = fs.readFileSync(path, 'utf8').trim().split('\n').map(JSON.parse);
const header = rows.shift();
const summary = rows.pop();
assert.equal(header.kind, 'header');
assert.equal(summary.kind, 'summary');
assert.equal(rows.length, 4000);
for (const [file, digest] of [
  ['research/conditional-gate-v2.mjs', header.sourceSHA256],
  ['docs/experiments/mmm-conditional-gate-v2-protocol.md', header.protocolSHA256],
]) assert.equal(crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex'), digest);

const ids = new Set();
for (const r of rows) {
  assert.equal(r.kind, 'trial');
  const key = `${r.schedule}/${r.regime}/${r.trial}`;
  assert(!ids.has(key));
  ids.add(key);
  assert(r.trial >= 0 && r.trial < 1000);
  assert(r.nominated >= 0 && r.nominated <= 512);
  assert(r.missing >= 0 && r.missing <= r.nominated);
  assert(r.delivered >= 0 && r.delivered <= r.nominated - r.missing);
  assert.equal(r.pending, r.nominated - r.missing - r.delivered);
  assert.equal(r.observedByCell[0] + r.observedByCell[1], r.delivered);
  for (const [clock, labels] of [[r.scalarClock, r.scalarFlagLabels], [r.conditionalClock, r.conditionalFlagLabels]]) {
    if (clock === null) assert.equal(labels, null);
    else {
      assert(Number.isInteger(clock) && clock >= 0 && clock < 512);
      assert(Number.isInteger(labels) && labels > 0 && labels <= r.delivered);
    }
  }
  for (const vector of [r.brier, r.expectedBrier, r.accuracy]) {
    assert.equal(vector.length, 2);
    for (const v of vector) assert(Number.isFinite(v) && v >= 0 && v <= 1);
  }
  if (r.regime === 'stable' && r.scalarClock === null && r.conditionalClock === null) {
    assert.deepEqual(r.brier, [r.brier[0], r.brier[0]]);
    assert.deepEqual(r.expectedBrier, [r.expectedBrier[0], r.expectedBrier[0]]);
  }
}

function stats(values) {
  const n = values.length;
  assert.equal(n, 1000);
  const mean = values.reduce((a, b) => a + b, 0) / n;
  const sumsq = values.reduce((a, b) => a + (b - mean) ** 2, 0);
  const se = Math.sqrt(sumsq / (n - 1) / n);
  return { mean, lower: mean - 3.5 * se, upper: mean + 3.5 * se };
}

function quantile(values, p) {
  if (!values.length) return null;
  values = [...values].sort((a, b) => a - b);
  return values[Math.ceil(values.length * p) - 1];
}

const cells = [];
for (const schedule of header.schedules) for (const regime of ['stable', 'swapped']) {
  const rs = rows.filter(r => r.schedule === schedule.name && r.regime === regime);
  assert.equal(rs.length, 1000);
  const scalar = rs.filter(r => r.scalarClock !== null);
  const conditional = rs.filter(r => r.conditionalClock !== null);
  const gain = stats(rs.map(r => r.brier[0] - r.brier[1]));
  const expectedGain = stats(rs.map(r => r.expectedBrier[0] - r.expectedBrier[1]));
  const average = f => rs.reduce((a, r) => a + f(r), 0) / rs.length;
  const cell = {
    schedule: schedule.name, regime, trials: 1000,
    scalarFlags: scalar.length, conditionalFlags: conditional.length,
    gain, expectedGain,
    scalarBrier: average(r => r.brier[0]),
    conditionalBrier: average(r => r.brier[1]),
    scalarAccuracy: average(r => r.accuracy[0]),
    conditionalAccuracy: average(r => r.accuracy[1]),
    meanNominated: average(r => r.nominated),
    meanMissing: average(r => r.missing),
    meanDelivered: average(r => r.delivered),
    meanPending: average(r => r.pending),
    conditionalClockP50: quantile(conditional.map(r => r.conditionalClock), .5),
    conditionalClockP95: quantile(conditional.map(r => r.conditionalClock), .95),
  };
  cell.pass = regime === 'stable'
    ? scalar.length <= 20 && conditional.length <= 20 && -gain.lower <= .01
    : scalar.length <= 20 && conditional.length >= 900 && gain.mean >= .05 && gain.lower > 0;
  cells.push(cell);
}
assert.deepEqual(cells, summary.cells);
assert.equal(summary.pass, cells.every(c => c.pass));
console.log(JSON.stringify({ rows: rows.length, pass: summary.pass, cells }, null, 2));
