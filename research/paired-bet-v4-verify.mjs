import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';

if (process.argv.length !== 3) throw Error('usage: node research/paired-bet-v4-verify.mjs DATA.jsonl');
const [header, ...tail] = readFileSync(process.argv[2], 'utf8').trim().split('\n').map(JSON.parse);
const summary = tail.pop();
const cases = ['stable', 'boundary', 'common_shift256', 'live_shift256',
  'reference_shift256', 'live_shift224', 'live_shift0'];
const schedules = ['complete', 'sparse_delayed'];
const arms = ['paired_bet', 'conditional_window', 'scalar_window'];
assert.equal(header.kind, 'header');
assert(['design', 'confirmation'].includes(header.split));
assert.equal(header.seedBase, header.split === 'design' ? 2026100601 : 2026100602);
assert.equal(header.T, 512);
assert.equal(header.trials, 1000);
assert.equal(header.width, 128);
assert.equal(header.epsilon, .1);
assert.equal(header.delta, .02);
assert.equal(header.lambda, .8);
assert.deepEqual(header.starts, [0, 64, 128, 192, 256, 320, 384, 448]);
assert.equal(header.protocolSHA256, createHash('sha256')
  .update(readFileSync('docs/experiments/mmm-paired-bet-v4-protocol.md')).digest('hex'));
assert.equal(header.sourceSHA256, createHash('sha256')
  .update(readFileSync('research/paired-bet-v4.mjs')).digest('hex'));
assert.equal(tail.length, 14000);

const seen = new Set();
for (const r of tail) {
  assert.equal(r.kind, 'trial');
  assert.equal(r.split, header.split);
  assert(schedules.includes(r.schedule));
  assert(cases.includes(r.scenario));
  assert(Number.isInteger(r.trial) && r.trial >= 0 && r.trial < 1000);
  const key = [r.schedule, r.scenario, r.trial].join(':');
  assert(!seen.has(key), `duplicate ${key}`);
  seen.add(key);
  assert.deepEqual(Object.keys(r.flags), arms);
  for (const arm of arms) {
    const flag = r.flags[arm];
    assert(flag === null || Number.isInteger(flag) && flag >= 0 && flag < 512);
  }
  for (const name of ['factorUpdates', 'nominated', 'missingPairs', 'missingReference',
    'missingLive', 'observedReference', 'observedLive', 'deliveredPairs',
    'pendingPairs', 'contextReadings', 'observedLabels', 'usableLabels']) {
    assert(Number.isInteger(r[name]) && r[name] >= 0);
  }
  assert(r.factorUpdates % 2 === 0 && r.factorUpdates <= 16 * r.deliveredPairs);
  assert(r.missingPairs >= Math.max(r.missingReference, r.missingLive));
  assert(r.missingPairs <= r.missingReference + r.missingLive);
  assert(r.nominated >= r.missingPairs + r.deliveredPairs);
  assert.equal(r.pendingPairs, r.nominated - r.missingPairs - r.deliveredPairs);
  assert(r.observedReference <= r.nominated - r.missingReference);
  assert(r.observedLive <= r.nominated - r.missingLive);
  assert(r.observedReference >= r.deliveredPairs);
  assert(r.observedLive >= r.deliveredPairs);
  assert.equal(r.contextReadings, 2 * r.nominated);
  assert.equal(r.observedLabels, r.observedReference + r.observedLive);
  assert.equal(r.usableLabels, 2 * r.deliveredPairs);
  if (r.schedule === 'complete') {
    assert.equal(r.nominated, 512);
    assert.equal(r.missingPairs, 0);
    assert.equal(r.pendingPairs, 0);
    assert.equal(r.deliveredPairs, 512);
    assert.equal(r.factorUpdates, 4608);
  }
}
assert.equal(seen.size, 14000);

function quantile(values, p) {
  if (!values.length) return null;
  const sorted = [...values].sort((a, b) => a - b);
  return sorted[Math.ceil(sorted.length * p) - 1];
}
function changeAt(scenario) {
  if (scenario === 'live_shift224') return 224;
  if (scenario.endsWith('shift256')) return 256;
  return null;
}
const cells = [];
for (const schedule of schedules) for (const scenario of cases) {
  const rows = tail.filter(r => r.schedule === schedule && r.scenario === scenario);
  assert.equal(rows.length, 1000);
  const results = {};
  for (const arm of arms) {
    const flags = rows.map(r => r.flags[arm]).filter(v => v !== null);
    const onset = changeAt(scenario);
    results[arm] = { flags: flags.length,
      prechangeFlags: onset === null ? null : flags.filter(t => t < onset).length,
      p50: quantile(flags, .5), p95: quantile(flags, .95) };
  }
  const mean = key => rows.reduce((s, r) => s + r[key], 0) / 1000;
  const cell = { schedule, scenario, trials: 1000, results,
    meanNominated: mean('nominated'), meanMissingPairs: mean('missingPairs'),
    meanPendingPairs: mean('pendingPairs'), meanDeliveredPairs: mean('deliveredPairs'),
    meanContextReadings: mean('contextReadings'), meanObservedLabels: mean('observedLabels'),
    meanUsableLabels: mean('usableLabels'), meanFactorUpdates: mean('factorUpdates') };
  const candidate = results.paired_bet, onset = changeAt(scenario);
  cell.pass = ['stable', 'boundary', 'common_shift256'].includes(scenario) ?
    candidate.flags <= 20 :
    candidate.flags >= 800 && (onset === null || candidate.prechangeFlags === 0 &&
      candidate.p50 - onset <= (schedule === 'complete' ? 192 : 224));
  cells.push(cell);
}
assert.equal(summary.kind, 'summary');
assert.deepEqual(summary.cells, cells);
assert.equal(summary.pass, cells.every(c => c.pass));
assert(Number.isFinite(summary.elapsedMs) && summary.elapsedMs > 0);
console.log(JSON.stringify({ split: header.split, verifiedTrials: tail.length, pass: summary.pass,
  cells: cells.map(c => [c.schedule, c.scenario, c.results.paired_bet, c.pass]) }, null, 2));
