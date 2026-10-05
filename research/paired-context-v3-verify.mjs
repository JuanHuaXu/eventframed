import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';

if (process.argv.length !== 3) throw Error('usage: node research/paired-context-v3-verify.mjs DATA.jsonl');
const [header, ...tail] = readFileSync(process.argv[2], 'utf8').trim().split('\n').map(JSON.parse);
const summary = tail.pop();
const cases = ['stable', 'live_shift256', 'reference_shift256', 'common_shift256', 'live_shift0'];
const schedules = ['complete', 'sparse_delayed'];
const arms = ['paired_window', 'paired_cumulative', 'scalar_window'];
assert.equal(header.kind, 'header');
assert(['design', 'confirmation'].includes(header.split));
assert.equal(header.seedBase, header.split === 'design' ? 2026100501 : 2026100502);
assert.equal(header.T, 512);
assert.equal(header.trials, 1000);
assert.equal(header.delta, .02);
assert.equal(header.epsilon, .10);
assert.equal(header.width, 128);
assert.equal(header.protocolSHA256, createHash('sha256')
  .update(readFileSync('docs/experiments/mmm-paired-context-v3-protocol.md')).digest('hex'));
assert.equal(header.sourceSHA256, createHash('sha256')
  .update(readFileSync('research/paired-context-v3.mjs')).digest('hex'));
assert.equal(tail.length, 10000);

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
  assert.deepEqual(Object.keys(r.counts), arms);
  for (const arm of arms) {
    const f = r.flags[arm];
    assert(f === null || Number.isInteger(f) && f >= 0 && f < 512);
    assert.equal(r.counts[arm].length, 2);
    assert(r.counts[arm].every(n => Number.isInteger(n) && n >= 0));
    assert(r.counts[arm][0] + r.counts[arm][1] <= r.deliveredPairs);
  }
  assert.equal(r.counts.paired_cumulative[0] + r.counts.paired_cumulative[1], r.deliveredPairs);
  assert.equal(r.counts.scalar_window[1], 0);
  for (const name of ['nominated', 'missingPairs', 'missingReference', 'missingLive',
    'observedReference', 'observedLive', 'deliveredPairs', 'pendingPairs',
    'contextReadings', 'observedLabels', 'usableLabels']) {
    assert(Number.isInteger(r[name]) && r[name] >= 0);
  }
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
  }
}
assert.equal(seen.size, 10000);

function quantile(v, p) {
  if (!v.length) return null;
  const sorted = [...v].sort((a, b) => a - b);
  return sorted[Math.ceil(v.length * p) - 1];
}
const cells = [];
for (const schedule of schedules) for (const scenario of cases) {
  const rows = tail.filter(r => r.schedule === schedule && r.scenario === scenario);
  assert.equal(rows.length, 1000);
  const results = {};
  for (const arm of arms) {
    const flags = rows.map(r => r.flags[arm]).filter(v => v !== null);
    results[arm] = { flags: flags.length,
      prechangeFlags: scenario.includes('shift256') ? flags.filter(t => t < 256).length : null,
      p50: quantile(flags, .5), p95: quantile(flags, .95) };
  }
  const avg = name => rows.reduce((total, r) => total + r[name], 0) / rows.length;
  const cell = { schedule, scenario, trials: 1000, results,
    meanNominated: avg('nominated'), meanMissingPairs: avg('missingPairs'),
    meanPendingPairs: avg('pendingPairs'), meanDeliveredPairs: avg('deliveredPairs'),
    meanContextReadings: avg('contextReadings'), meanObservedLabels: avg('observedLabels'),
    meanUsableLabels: avg('usableLabels') };
  const candidate = results.paired_window;
  cell.pass = scenario === 'stable' || scenario === 'common_shift256' ? candidate.flags <= 20 :
    scenario === 'live_shift0' ? candidate.flags >= 800 :
    candidate.flags >= 800 && candidate.prechangeFlags === 0 &&
    (schedule !== 'complete' || candidate.p50 - 256 <= 192);
  cells.push(cell);
}
assert.equal(summary.kind, 'summary');
assert.deepEqual(summary.cells, cells);
assert.equal(summary.pass, cells.every(c => c.pass));
assert(Number.isFinite(summary.elapsedMs) && summary.elapsedMs > 0);
console.log(JSON.stringify({ split: header.split, verifiedTrials: tail.length, pass: summary.pass,
  cells: cells.map(c => [c.schedule, c.scenario, c.results, c.pass]) }, null, 2));
