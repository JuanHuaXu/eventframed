import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';

if (process.argv.length !== 3) throw Error('usage: node research/conditional-window-v2-verify.mjs DATA.jsonl');
const rows = readFileSync(process.argv[2], 'utf8').trim().split('\n').map(JSON.parse);
const [header, ...tail] = rows;
const claimed = tail.pop();
assert.equal(header.kind, 'header');
assert.equal(header.T, 512);
assert.equal(header.trials, 1000);
assert.equal(header.delta, .02);
assert.equal(header.epsilon, .10);
assert.equal(header.batchN, 256);
assert.equal(header.seedBase, header.split === 'design' ? 2026100401 : 2026100402);
assert.equal(header.protocolSHA256, createHash('sha256')
  .update(readFileSync('docs/experiments/mmm-conditional-window-v2-protocol.md')).digest('hex'));
assert.equal(header.sourceSHA256, createHash('sha256')
  .update(readFileSync('research/conditional-window-v2.mjs')).digest('hex'));
assert.equal(tail.length, 12000);

const seen = new Set();
for (const r of tail) {
  assert.equal(r.kind, 'trial');
  assert.equal(r.split, header.split);
  assert(Number.isInteger(r.trial) && r.trial >= 0 && r.trial < 1000);
  const key = [r.strategy, r.schedule, r.regime, r.trial].join(':');
  assert(!seen.has(key), `duplicate ${key}`);
  seen.add(key);
  const arms = r.regime === 'swapped_at_256' ?
    ['cumulative', 'window256', 'oracle_reset'] : ['cumulative', 'window256'];
  assert.deepEqual(Object.keys(r.flags), arms);
  assert.deepEqual(Object.keys(r.liveActive), arms);
  for (const arm of arms) {
    const flag = r.flags[arm];
    assert(flag === null || Number.isInteger(flag) && flag >= 0 && flag < 512);
    const active = r.liveActive[arm];
    assert.equal(active.length, 2);
    assert(active.every(n => Number.isInteger(n) && n >= 0));
    assert(active[0] + active[1] <= r.delivered.live);
  }
  assert.equal(r.liveActive.cumulative[0] + r.liveActive.cumulative[1], r.delivered.live);
  assert.equal(r.referenceCount[0] + r.referenceCount[1],
    r.delivered.reference + r.batchLabels);
  assert.equal(r.batchLabels, r.strategy === 'batch_reference' ? 512 : 0);
  for (const side of ['reference', 'live']) {
    assert(r.nominated[side] >= r.missing[side] + r.delivered[side]);
    assert.equal(r.pending[side], r.nominated[side] - r.missing[side] - r.delivered[side]);
  }
}
assert.equal(seen.size, 12000);

function quantile(values, p) {
  if (!values.length) return null;
  const v = [...values].sort((a, b) => a - b);
  return v[Math.ceil(v.length * p) - 1];
}
const cells = [];
for (const strategy of ['batch_reference', 'online_reference']) {
  for (const schedule of ['immediate', 'sparse_delayed']) {
    for (const regime of ['stable', 'swapped_from_start', 'swapped_at_256']) {
      const subset = tail.filter(r => r.strategy === strategy && r.schedule === schedule && r.regime === regime);
      assert.equal(subset.length, 1000);
      const arms = regime === 'swapped_at_256' ?
        ['cumulative', 'window256', 'oracle_reset'] : ['cumulative', 'window256'];
      const results = {};
      for (const arm of arms) {
        const flags = subset.map(r => r.flags[arm]).filter(v => v !== null);
        results[arm] = {
          flags: flags.length,
          prechangeFlags: regime === 'swapped_at_256' ? flags.filter(t => t < 256).length : null,
          p50: quantile(flags, .5), p95: quantile(flags, .95),
        };
      }
      const avg = f => subset.reduce((a, r) => a + f(r), 0) / 1000;
      const cell = {
        strategy, schedule, regime, trials: 1000, results,
        meanReferenceLabels: avg(r => r.batchLabels + r.delivered.reference),
        meanLiveLabels: avg(r => r.delivered.live),
        meanReferencePending: avg(r => r.pending.reference),
        meanLivePending: avg(r => r.pending.live),
      };
      const w = results.window256;
      cell.pass = regime === 'stable' ? w.flags <= 20 :
        regime === 'swapped_from_start' ? w.flags >= 800 :
        w.flags >= 800 && w.prechangeFlags === 0;
      cells.push(cell);
    }
  }
}
assert.equal(claimed.kind, 'summary');
assert.deepEqual(claimed.cells, cells);
assert.equal(claimed.pass, cells.every(c => c.pass));
console.log(JSON.stringify({ split: header.split, verifiedTrials: tail.length, pass: claimed.pass,
  cells: cells.map(c => [c.strategy, c.schedule, c.regime, c.results]) }, null, 2));
