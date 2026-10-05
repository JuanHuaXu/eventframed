import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const path = 'docs/experiments/mmm-online-learned-v95.json';
const sha = b => crypto.createHash('sha256').update(b).digest('hex');
const raw = fs.readFileSync(path), a = JSON.parse(raw);
assert.equal(a.Protocol, 'mmm-online-learned-v95');
assert.equal(Object.keys(a.Hashes).length, 11);
for (const [p, hash] of Object.entries(a.Hashes)) assert.equal(sha(fs.readFileSync(p)), hash, p);
const names = ['parity1','parity2','parity3','parity4','complement4','majority3','mux3','constant','null','dependent4','majority_to_parity','parity_to_majority'];
const phases = ['design','confirmation'];
assert.equal(a.Records.length, 768);
const seen = new Set(), tapes = new Set();
for (const r of a.Records) {
  assert(names.includes(r.Case) && phases.includes(r.Phase));
  assert(Number.isInteger(r.Index) && r.Index >= 0 && r.Index < 32);
  const key = `${r.Phase}/${r.Case}/${r.Index}`; assert(!seen.has(key)); seen.add(key);
  assert(/^[0-9a-f]{64}$/.test(r.Tape)); assert(!tapes.has(r.Tape)); tapes.add(r.Tape);
  assert(r.Rules.length === 2 && r.Rules.every(x => Number.isInteger(x) && x >= 0 && x < 512));
  assert.equal(r.Metrics.length, 5); assert.equal(r.Realized.length, 5);
  for (const row of r.Metrics) {assert.equal(row.length, 2); for (const view of row) {
    assert.equal(view.length, 2); for (const m of view) for (const k of ['Brier','Accuracy']) assert(Number.isFinite(m[k]) && m[k] >= 0 && m[k] <= 1);
  }}
  for (const row of r.Realized) {assert.equal(row.length, 2); assert(row.every(v => Number.isFinite(v) && v >= 0 && v <= 256));}
  assert.equal(r.MaxViolation.length, 2);
  for (const row of r.MaxViolation) {assert.equal(row.length, 2); assert(row.every(v => Number.isFinite(v) && v >= 0 && v <= 1e-8));}
}
const interval = values => {
  const n = values.length, mean = values.reduce((s,x) => s+x, 0)/n;
  const se = Math.sqrt(values.reduce((s,x) => s+(x-mean)**2, 0)/(n*(n-1)));
  return {mean, lower:mean-3.5*se, upper:mean+3.5*se};
};
const cells = [], gates = [];
for (const phase of phases) for (const scenario of names) {
  const rows = a.Records.filter(r => r.Phase === phase && r.Case === scenario); assert.equal(rows.length,32);
  const mean = f => rows.reduce((s,r) => s+f(r),0)/rows.length;
  for (let view = 0; view < 2; view++) for (let segment = 0; segment < 2; segment++) {
    const arms = Array.from({length:5}, (_,arm) => ({brier:mean(r => r.Metrics[arm][view][segment].Brier), accuracy:mean(r => r.Metrics[arm][view][segment].Accuracy),
      gain:interval(rows.map(r => r.Metrics[0][view][segment].Brier-r.Metrics[arm][view][segment].Brier))}));
    const label = {phase,scenario,view,segment};
    cells.push({...label, arms});
    const primary = arms[4].gain;
    if (segment === 0) gates.push({...label,kind:'nonharm',pass:-primary.lower <= .01,value:-primary.lower});
    if (view === 0 && segment === 0 && ['parity3','parity4','complement4'].includes(scenario)) gates.push({...label,kind:'interaction_gain',pass:primary.lower > 0 && primary.mean >= .005,...primary});
    if (view === 0 && segment === 1 && scenario.includes('_to_')) gates.push({...label,kind:'recovery_gain',pass:primary.lower > 0 && primary.mean >= .005,...primary});
  }
}
assert.equal(gates.length,58);
const result = {protocol:a.Protocol,artifactSHA256:sha(raw),evaluatorSHA256:sha(fs.readFileSync(import.meta.filename)),runtime:a.Runtime,
  streams:768,stepsPerStream:256,armOrder:['generic','parity','skeptical_BMA','online_no_share','online_share'],
  caveat:'Approximate paired trajectory z=3.5 intervals, not anytime confidence sequences; complete full-frame training feedback in both views.',
  passed:gates.filter(g => g.pass).length,total:58,verdict:gates.every(g => g.pass)?'PASS':'FAIL',
  maxBoundViolation:Math.max(...a.Records.flatMap(r => r.MaxViolation.flat())),gates,cells};
const output = JSON.stringify(result,null,2)+'\n';
if (process.argv[2]) fs.writeFileSync(process.argv[2],output,{flag:'wx',mode:0o600}); else process.stdout.write(output);
