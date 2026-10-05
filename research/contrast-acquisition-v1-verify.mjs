import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';

if (process.argv.length !== 3) throw Error('usage: node research/contrast-acquisition-v1-verify.mjs DATA.jsonl');
const [header, ...tail] = readFileSync(process.argv[2], 'utf8').trim().split('\n').map(JSON.parse);
const saved = tail.pop();
const cases = ['stable', 'member_shift', 'common_shift', 'recurring', 'null'];
const policies = ['random', 'uncertainty', 'contrast'];
const sources = ['docs/experiments/mmm-contrast-acquisition-v1-protocol.md',
  'internal/observationgate/contrast_acquisition_v1_test.go',
  'internal/observationgate/member_integration_test.go',
  'internal/observationpreserved/experiment.go',
  'internal/observation/controller.go', 'internal/observation/forecast.go'];
assert.equal(header.kind, 'header');
assert(['design', 'confirmation'].includes(header.split));
assert.equal(header.seedBase, header.split === 'design' ? 2026100801 : 2026100802);
assert.equal(header.steps, 512);
assert.equal(header.trials, 1000);
assert(Number.isInteger(header.fitNS) && header.fitNS > 0);
assert.deepEqual(Object.keys(header.hashes).sort(), [...sources].sort());
for (const name of sources) assert.equal(header.hashes[name],
  createHash('sha256').update(readFileSync(name)).digest('hex'));
assert.equal(tail.length, 5000);

const seen = new Set();
for (const r of tail) {
  assert.equal(r.Kind, 'trial');
  assert.equal(r.Split, header.split);
  assert(cases.includes(r.Scenario));
  assert(Number.isInteger(r.Trial) && r.Trial >= 0 && r.Trial < 1000);
  const key = `${r.Scenario}:${r.Trial}`;
  assert(!seen.has(key));
  seen.add(key);
  assert.equal(r.Arms.length, 3);
  assert.deepEqual(r.Arms.map(a => a.Policy), policies);
  for (const a of r.Arms) {
    for (const name of ['FullBrier', 'PostBrier', 'EarlyBrier']) {
      assert(Number.isFinite(a[name]) && a[name] >= 0 && a[name] <= 1);
    }
    assert(Number.isInteger(a.FlagAt) && a.FlagAt >= -1 && a.FlagAt < 512);
    for (const name of ['Nominated', 'ClassMatched', 'DeliveredPairs', 'ArrivedLabels',
      'LiveReads', 'RefReads', 'RequestedLabels', 'FactorUpdates']) {
      assert(Number.isInteger(a[name]) && a[name] >= 0);
    }
    assert.equal(a.Nominated, 128);
    assert.equal(a.LiveReads, 512);
    assert.equal(a.RefReads, 512);
    assert.equal(a.RequestedLabels, 256);
    assert(a.ClassMatched <= a.Nominated);
    assert(a.DeliveredPairs <= a.ClassMatched);
    assert(a.ArrivedLabels >= 2 * a.DeliveredPairs && a.ArrivedLabels <= 256);
    assert(a.FactorUpdates % 2 === 0 && a.FactorUpdates <= 16 * a.DeliveredPairs);
  }
}
assert.equal(seen.size, 5000);

function quantile(values, p) {
  if (!values.length) return -1;
  const sorted = [...values].sort((a, b) => a - b);
  return sorted[Math.ceil(values.length * p) - 1];
}
function close(actual, expected) {
  assert(Math.abs(actual - expected) <= 1e-11 * Math.max(1, Math.abs(expected)),
    `numeric mismatch ${actual} != ${expected}`);
}
function compareObject(actual, expected) {
  assert.deepEqual(Object.keys(actual).sort(), Object.keys(expected).sort());
  for (const key of Object.keys(expected)) {
    if (typeof expected[key] === 'number' && !Number.isInteger(expected[key])) close(actual[key], expected[key]);
    else assert.deepEqual(actual[key], expected[key]);
  }
}

const cells = [];
let pass = true;
for (const Scenario of cases) for (const Policy of policies) {
  const arms = tail.filter(r => r.Scenario === Scenario)
    .map(r => r.Arms.find(a => a.Policy === Policy));
  assert.equal(arms.length, 1000);
  const flags = arms.filter(a => a.FlagAt >= 0).map(a => a.FlagAt);
  const mean = field => arms.reduce((s, a) => s + a[field], 0) / arms.length;
  const c = { Scenario, Policy, Trials: 1000, Flags: flags.length,
    Prechange: Scenario === 'member_shift' ? flags.filter(t => t < 256).length : 0,
    MedianFlag: quantile(flags, .5), P95Flag: quantile(flags, .95),
    MeanFull: mean('FullBrier'), MeanPost: mean('PostBrier'),
    MeanEarly: mean('EarlyBrier'), MeanMatched: mean('ClassMatched'),
    MeanPairs: mean('DeliveredPairs'), MeanArrived: mean('ArrivedLabels'),
    MeanFactors: mean('FactorUpdates') };
  if (Policy === 'contrast') {
    if (['stable', 'common_shift', 'null'].includes(Scenario) && c.Flags > 20) pass = false;
    if (Scenario === 'member_shift' && (c.Flags < 800 || c.Prechange !== 0)) pass = false;
  }
  cells.push(c);
}

function gain(rows, control, field) {
  const ds = rows.map(r => r.Arms.find(a => a.Policy === control)[field] -
    r.Arms.find(a => a.Policy === 'contrast')[field]);
  const n = ds.length, sum = ds.reduce((a, b) => a + b, 0),
    sq = ds.reduce((a, b) => a + b * b, 0), mean = sum / n;
  const variance = Math.max(0, (sq - n * mean * mean) / (n - 1));
  return [mean, mean - 3.5 * Math.sqrt(variance / n)];
}
const comparisons = [];
for (const Scenario of cases) for (const Control of ['random', 'uncertainty']) {
  const rows = tail.filter(r => r.Scenario === Scenario);
  const [PostGain, PostLower] = gain(rows, Control, 'PostBrier');
  const [EarlyGain, EarlyLower] = gain(rows, Control, 'EarlyBrier');
  const [full] = gain(rows, Control, 'FullBrier');
  const [arrived] = gain(rows, Control, 'ArrivedLabels');
  const c = { Scenario, Control, PostGain, PostLower, EarlyGain, EarlyLower,
    StableHarm: -full, ArrivedGap: Math.abs(arrived), Pass: true };
  if (Scenario === 'member_shift') {
    c.Pass = PostGain >= .02 && PostLower > 0 && EarlyGain >= .02 && EarlyLower > 0;
  }
  if (Scenario === 'stable' && c.StableHarm > .01) c.Pass = false;
  if (c.ArrivedGap > 2) c.Pass = false;
  pass &&= c.Pass;
  comparisons.push(c);
}
assert.equal(saved.kind, 'summary');
assert.equal(saved.cells.length, cells.length);
assert.equal(saved.comparisons.length, comparisons.length);
for (let i = 0; i < cells.length; i++) compareObject(saved.cells[i], cells[i]);
for (let i = 0; i < comparisons.length; i++) compareObject(saved.comparisons[i], comparisons[i]);
assert.equal(saved.pass, pass);
assert(Number.isFinite(saved.elapsedMs) && saved.elapsedMs > 0);
console.log(JSON.stringify({ split: header.split, trials: tail.length, pass,
  member: cells.filter(c => c.Scenario === 'member_shift'),
  comparisons: comparisons.filter(c => c.Scenario === 'member_shift' || !c.Pass) }, null, 2));
