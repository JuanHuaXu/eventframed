import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';

if (process.argv.length !== 3) {
  throw Error('usage: node research/conditional-power-v1-verify.mjs DATA.jsonl');
}
const lines = readFileSync(process.argv[2], 'utf8').trim().split('\n').map(JSON.parse);
const [header, ...rest] = lines;
const claimed = rest.pop();
assert.equal(header.kind, 'header');
assert.equal(header.T, 512);
assert.equal(header.trials, 1000);
assert.equal(header.delta, .02);
assert.equal(header.epsilon, .10);
assert.equal(header.batchN, 256);
assert.equal(header.seedBase, header.split === 'design' ? 2026100301 : 2026100302);
assert.equal(header.protocolSHA256, createHash('sha256')
  .update(readFileSync('docs/experiments/mmm-conditional-power-v1-protocol.md')).digest('hex'));
assert.equal(header.sourceSHA256, createHash('sha256')
  .update(readFileSync('research/conditional-power-v1.mjs')).digest('hex'));
assert.equal(rest.length, 12000);

const seen = new Set();
for (const r of rest) {
  assert.equal(r.kind, 'trial');
  assert.equal(r.split, header.split);
  assert(Number.isInteger(r.trial) && r.trial >= 0 && r.trial < 1000);
  const key = [r.strategy, r.schedule, r.regime, r.trial].join(':');
  assert(!seen.has(key), `duplicate ${key}`);
  seen.add(key);
  assert(r.flagAt === null || Number.isInteger(r.flagAt) && r.flagAt >= 0 && r.flagAt < 512);
  assert.equal(r.batchLabels, r.strategy === 'batch_reference' ? 512 : 0);
  for (const side of ['reference', 'live']) {
    assert(r.nominated[side] >= r.missing[side] + r.delivered[side]);
    assert.equal(r.pending[side], r.nominated[side] - r.missing[side] - r.delivered[side]);
    const count = r[`${side}Count`], yes = r[`${side}Yes`];
    assert.equal(count.length, 2);
    assert.equal(yes.length, 2);
    assert.equal(count[0] + count[1], r.delivered[side] + (side === 'reference' ? r.batchLabels : 0));
    for (let x = 0; x < 2; x++) {
      assert(Number.isInteger(count[x]) && Number.isInteger(yes[x]));
      assert(yes[x] >= 0 && yes[x] <= count[x]);
    }
  }
}
assert.equal(seen.size, 12000);

function quantile(xs, p) {
  if (!xs.length) return null;
  const v = [...xs].sort((a, b) => a - b);
  return v[Math.ceil(v.length * p) - 1];
}
const cells = [];
for (const strategy of ['batch_reference', 'online_reference']) {
  for (const schedule of ['immediate', 'sparse_delayed']) {
    for (const regime of ['stable', 'swapped_from_start', 'swapped_at_256']) {
      const subset = rest.filter(r => r.strategy === strategy && r.schedule === schedule && r.regime === regime);
      assert.equal(subset.length, 1000);
      const flags = subset.map(r => r.flagAt).filter(v => v !== null);
      const average = f => subset.reduce((a, r) => a + f(r), 0) / 1000;
      const cell = {
        strategy, schedule, regime, trials: 1000,
        flags: flags.length,
        prechangeFlags: regime === 'swapped_at_256' ? flags.filter(t => t < 256).length : null,
        detectionP50: quantile(flags, .5), detectionP95: quantile(flags, .95),
        meanReferenceLabels: average(r => r.batchLabels + r.delivered.reference),
        meanLiveLabels: average(r => r.delivered.live),
        meanReferenceNominated: average(r => r.nominated.reference),
        meanLiveNominated: average(r => r.nominated.live),
        meanReferenceMissing: average(r => r.missing.reference),
        meanLiveMissing: average(r => r.missing.live),
        meanReferencePending: average(r => r.pending.reference),
        meanLivePending: average(r => r.pending.live),
      };
      cell.pass = regime === 'stable' ? flags.length <= 20 :
        regime === 'swapped_from_start' ? flags.length >= 800 :
        flags.length >= 800 && cell.prechangeFlags === 0;
      cells.push(cell);
    }
  }
}
assert.equal(claimed.kind, 'summary');
assert.deepEqual(claimed.cells, cells);
assert.equal(claimed.pass, cells.every(c => c.pass));
console.log(JSON.stringify({ split: header.split, verifiedTrials: rest.length, cells, pass: claimed.pass }, null, 2));
