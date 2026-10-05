import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const pop = x => { let n = 0; for (; x; x &= x - 1) n++; return n; };
const mean = xs => xs.reduce((s, x) => s + x, 0) / xs.length;
const close = (a, b) => assert.ok(Math.abs(a - b) < 1e-12, `${a} != ${b}`);
const truth = (x, target, majority) => majority ? pop(x & target) >= 2 : pop(x & target) % 2 === 1;

// Oracle information is confined to this offline scorer. Uniform independent
// inputs and symmetric 5% outcome noise are properties of this simulator only.
function oracle(target, majority, mask, values) {
  assert.equal(values & mask, values);
  let n = 0, yes = 0;
  for (let x = 0; x < 512; x++) if ((x & mask) === values) {
    n++;
    yes += Number(truth(x, target, majority));
  }
  return .05 + .9 * yes / n;
}

// Independent completion-count formula, tested over every partial assignment.
function combinatorial(target, majority, mask, values) {
  const known = pop(target & mask & values), missing = pop(target & ~mask);
  if (!majority) return missing ? .5 : .05 + .9 * (known % 2);
  let yes = 0;
  for (let x = 0; x < 1 << missing; x++) yes += Number(known + pop(x) >= 2);
  return .05 + .9 * yes / (1 << missing);
}

let checks = 0;
for (const [target, majority] of [[7, true], [15, false]]) {
  for (let mask = 0; mask < 512; mask++) {
    for (let values = mask; ; values = (values - 1) & mask) {
      const q = oracle(target, majority, mask, values);
      close(q, combinatorial(target, majority, mask, values));
      // The decomposition must hold for arbitrary predictions, not just q.
      for (const p of [0, .2, .5, .9, 1]) {
        close(q * (1 - p) ** 2 + (1 - q) * p ** 2,
          .0475 + (q * (1 - q) - .0475) + (p - q) ** 2);
      }
      checks++;
      if (values === 0) break;
    }
  }
}
close(oracle(15, false, 7, 0), .5);
close(oracle(15, false, 15, 1), .95);
close(oracle(7, true, 3, 3), .95);
close(oracle(7, true, 3, 0), .05);
close(oracle(7, true, 0, 0), .5);

const [proxyPath, coupledPath, parentPath, output] = process.argv.slice(2);
assert.ok(output, 'supply proxy, coupled, parent, output paths');
const expected = [
  'cdd7810b488215cb6556e3912b2490bd8c11dbd69aac61267b311e2498d0be7b',
  'ab3756fc535db2a1cc47d3c97629d5e2e291452284dd1238702e7cebf840dda5',
  '4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655',
];
const buffers = [proxyPath, coupledPath, parentPath].map(p => fs.readFileSync(p));
buffers.forEach((b, i) => assert.equal(hash(b), expected[i]));
const [proxy, coupled, parent] = buffers.map(b => b.toString().trim().split('\n').slice(1).map(JSON.parse));
for (const rows of [proxy, coupled, parent]) assert.equal(rows.length, 128);
const names = ['stable_majority3', 'stable_parity4', 'majority_to_parity', 'parity_to_majority'];
const armNames = ['fixed', 'coupled', 'proxy'];
const cache = new Map(), records = [];
for (let k = 0; k < 128; k++) {
  const r = proxy[k], c = coupled[k], parentRow = parent[k];
  for (const key of ['Phase', 'Case', 'Index']) {
    assert.equal(r[key], c[key]); assert.equal(r[key], parentRow[key]);
  }
  const scenario = names.indexOf(r.Case);
  assert.ok(scenario >= 0);
  for (const schedule of ['Immediate', 'Delayed']) for (const start of [0, 128, 256, 384]) {
    const metrics = armNames.map(() => ({ realized: 0, conditionalRisk: 0, missingInformation: 0, forecastGap: 0 }));
    for (let i = start; i < start + 128; i++) {
      const f = parentRow[schedule].Frames[i];
      const changed = scenario >= 2 && i >= 256;
      const target = parentRow.Masks[Number(changed)];
      const majority = (scenario === 0 || scenario === 2) !== changed;
      assert.equal(pop(target), majority ? 3 : 4);
      assert.deepEqual(r[schedule].Frames[i].Arms[0], c[schedule].Frames[i].Arms[0]);
      const arms = [r[schedule].Frames[i].Arms[0], c[schedule].Frames[i].Arms[1], r[schedule].Frames[i].Arms[1]];
      arms.forEach((a, j) => {
        assert.ok(a.Mask >= 0 && a.Mask < 512 && a.P >= 0 && a.P <= 1);
        assert.equal(a.Values, f.X & a.Mask);
        const key = `${target}/${majority}/${a.Mask}/${a.Values}`;
        if (!cache.has(key)) cache.set(key, oracle(target, majority, a.Mask, a.Values));
        const q = cache.get(key), missing = q * (1 - q) - .0475, gap = (a.P - q) ** 2;
        assert.ok(missing >= -1e-12);
        metrics[j].realized += (a.P - Number(f.Y)) ** 2 / 128;
        metrics[j].conditionalRisk += (.0475 + missing + gap) / 128;
        metrics[j].missingInformation += missing / 128;
        metrics[j].forecastGap += gap / 128;
      });
    }
    records.push({ phase: r.Phase, case: r.Case, index: r.Index, schedule, start, metrics });
  }
}
const cells = [];
for (const phase of ['cohort1', 'cohort2']) for (const name of names) {
  for (const schedule of ['Immediate', 'Delayed']) for (const start of [0, 128, 256, 384]) {
    const rows = records.filter(r => r.phase === phase && r.case === name && r.schedule === schedule && r.start === start);
    assert.equal(rows.length, 16);
    const metrics = armNames.map((arm, j) => ({ arm, ...Object.fromEntries(
      Object.keys(rows[0].metrics[j]).map(key => [key, mean(rows.map(r => r.metrics[j][key]))])) }));
    cells.push({ phase, case: name, schedule, start, metrics });
  }
}
fs.writeFileSync(output, JSON.stringify({ inputSHA256: expected, scriptSHA256: hash(fs.readFileSync(new URL(import.meta.url))),
  checks, irreducibleRisk: .0475, cells, records,
  limits: 'Consumed simulator hindsight only. Conditional risk integrates unobserved independent uniform bits and fresh noise at each recorded view; realized Brier need not equal it on finite trajectories. Forecast gap combines model and weight errors and cannot establish sample sufficiency. No policy receives oracle labels or masks; no efficacy or equal-cost claim.'
}, null, 2) + '\n', { flag: 'wx' });
console.log(JSON.stringify({ checks, cells: cells.filter(c => c.case.includes('_to_') && c.start === 384) }));
