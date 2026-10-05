import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const pop = x => { let n = 0; for (; x; x &= x - 1) n++; return n; };
const clip = p => Math.max(1e-6, Math.min(1 - 1e-6, p));
const close = (a, b) => assert.ok(Math.abs(a - b) < 1e-12, `${a} != ${b}`);
const mean = xs => xs.reduce((s, x) => s + x, 0) / xs.length;
function oracle(target, majority, mask, values) {
  const known = pop(target & mask & values), missing = pop(target & ~mask);
  if (!majority) return missing ? .5 : .05 + .9 * (known % 2);
  let yes = 0;
  for (let x = 0; x < 1 << missing; x++) yes += Number(known + pop(x) >= 2);
  return .05 + .9 * yes / (1 << missing);
}
function projection(q, advice) {
  assert.ok(advice.length > 0 && advice.every(p => Number.isFinite(p) && p >= 0 && p <= 1));
  const lo = Math.min(...advice), hi = Math.max(...advice);
  return Math.max(lo, Math.min(hi, q));
}
close(projection(.8, [.1, .3]), .3);
close(projection(.2, [.3, .9]), .3);
close(projection(.5, [.1, .9]), .5);
close(projection(.8, [.5, .5]), .5);
// Projection is no worse than any convex combination, including endpoints.
let checks = 0;
for (let i = 0; i <= 10; i++) for (let j = 0; j <= 10; j++) {
  for (let k = 0; k <= 10; k++) for (let w = 0; w <= 10; w++) {
    const q = k / 10, a = i / 10, b = j / 10, p = (w * a + (10 - w) * b) / 10;
    assert.ok((projection(q, [a, b]) - q) ** 2 <= (p - q) ** 2 + 1e-12);
    checks++;
  }
}

const [proxyPath, coupledPath, parentPath, oraclePath, output] = process.argv.slice(2);
assert.ok(output, 'supply proxy, coupled, parent, oracle diagnostic, output');
const expected = [
  'cdd7810b488215cb6556e3912b2490bd8c11dbd69aac61267b311e2498d0be7b',
  'ab3756fc535db2a1cc47d3c97629d5e2e291452284dd1238702e7cebf840dda5',
  '4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655',
  '03669cbdd439890cf7f766eda7d92b44ae9091edd7fedd4685532cbb768ab337',
];
// Pin all tapes rather than trusting filename-based identity.
const buffers = [proxyPath, coupledPath, parentPath, oraclePath].map(p => fs.readFileSync(p));
buffers.forEach((b, i) => assert.equal(hash(b), expected[i]));
const [proxy, coupled, parent] = buffers.slice(0, 3).map(b => b.toString().trim().split('\n').slice(1).map(JSON.parse));
const previous = JSON.parse(buffers[3]);
const references = new Map(previous.records.map(r => [[r.phase, r.case, r.index, r.schedule, r.start].join('/'), r]));
for (const rows of [proxy, coupled, parent]) assert.equal(rows.length, 128);
const names = ['stable_majority3', 'stable_parity4', 'majority_to_parity', 'parity_to_majority'];
const armNames = ['fixed', 'coupled', 'proxy'], records = [];
for (let k = 0; k < 128; k++) {
  const r = proxy[k], c = coupled[k], p = parent[k];
  for (const key of ['Phase', 'Case', 'Index']) { assert.equal(r[key], c[key]); assert.equal(r[key], p[key]); }
  const scenario = names.indexOf(r.Case); assert.ok(scenario >= 0);
  for (const schedule of ['Immediate', 'Delayed']) for (const start of [0, 128, 256, 384]) {
    const metrics = armNames.map(() => ({ servedGap: 0, outerFloor: 0, leafFloor: 0, outerHeadroom: 0, innerHeadroom: 0, informativeFraction: 0, outsideLeafFraction: 0 }));
    for (let i = start; i < start + 128; i++) {
      const f = p[schedule].Frames[i], changed = scenario >= 2 && i >= 256;
      const target = p.Masks[Number(changed)], majority = (scenario === 0 || scenario === 2) !== changed;
      assert.equal(pop(target), majority ? 3 : 4);
      const arms = [r[schedule].Frames[i].Arms[0], c[schedule].Frames[i].Arms[1], r[schedule].Frames[i].Arms[1]];
      arms.forEach((a, j) => {
        assert.equal(a.Values, f.X & a.Mask);
        const q = oracle(target, majority, a.Mask, a.Values);
        const outer = a.Outer.map(clip);
        const leaves = [a.Outer[0], a.Outer[2], a.Outer[3], ...a.OldInner, ...a.NewInner.slice(1)].map(clip);
        const servedGap = (a.P - q) ** 2, outerFloor = (projection(q, outer) - q) ** 2, leafFloor = (projection(q, leaves) - q) ** 2;
        assert.ok(outerFloor <= servedGap + 1e-12 && leafFloor <= outerFloor + 1e-12);
        const values = { servedGap, outerFloor, leafFloor,
          outerHeadroom: servedGap - outerFloor, innerHeadroom: outerFloor - leafFloor,
          informativeFraction: Number(Math.abs(q - .5) > 1e-12), outsideLeafFraction: Number(leafFloor > 1e-12) };
        for (const [key, value] of Object.entries(values)) metrics[j][key] += value / 128;
      });
    }
    const ref = references.get([r.Phase, r.Case, r.Index, schedule, start].join('/')); assert.ok(ref);
    metrics.forEach((m, j) => close(m.servedGap, ref.metrics[j].forecastGap));
    records.push({ phase: r.Phase, case: r.Case, index: r.Index, schedule, start, metrics });
  }
}
const cells = [];
for (const phase of ['cohort1', 'cohort2']) for (const name of names) {
  for (const schedule of ['Immediate', 'Delayed']) for (const start of [0, 128, 256, 384]) {
    const rows = records.filter(r => r.phase === phase && r.case === name && r.schedule === schedule && r.start === start);
    assert.equal(rows.length, 16);
    const metrics = armNames.map((arm, j) => ({ arm, ...Object.fromEntries(Object.keys(rows[0].metrics[j]).map(key => [key, mean(rows.map(r => r.metrics[j][key]))])) }));
    cells.push({ phase, case: name, schedule, start, metrics });
  }
}
fs.writeFileSync(output, JSON.stringify({ inputSHA256: expected, scriptSHA256: hash(fs.readFileSync(new URL(import.meta.url))), checks, cells, records,
  limits: 'Hindsight per-frame relaxed convex hulls, not learnable policies. Ignores weight floors, delayed feedback and coupling of future observations to new weights. Same-view oracle probabilities checked against prior exhaustive diagnostic. Neither outcome labels nor oracle information enter any learner. No quality rescue or deployment claim.'
}, null, 2) + '\n', { flag: 'wx' });
console.log(JSON.stringify({ checks, cells: cells.filter(c => c.case.includes('_to_') && c.start === 384) }));
