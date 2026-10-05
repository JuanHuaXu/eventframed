import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';

if (process.argv.length !== 3) throw Error('usage: node research/member-pair-feasibility-v1-verify.mjs DATA.jsonl');
const [header, ...tail] = readFileSync(process.argv[2], 'utf8').trim().split('\n').map(JSON.parse);
const summary = tail.pop();
const cases = ['stable', 'member_shift', 'common_shift', 'recurring', 'null'];
const modes = ['passive', 'targeted'];
assert.equal(header.kind, 'header');
assert(['design', 'confirmation'].includes(header.split));
assert.equal(header.seedBase, header.split === 'design' ? 2026100701 : 2026100702);
assert.equal(header.steps, 512);
assert.equal(header.trials, 1000);
for (const name of ['docs/experiments/mmm-member-pair-feasibility-v1-protocol.md',
  'internal/observationgate/member_pair_feasibility_v1_test.go',
  'internal/observationgate/member_integration_test.go',
  'internal/observationpreserved/experiment.go']) {
  assert.equal(header.hashes[name], createHash('sha256').update(readFileSync(name)).digest('hex'));
}
assert.equal(Object.keys(header.hashes).length, 4);
assert.equal(tail.length, 10000);
const seen = new Set();
for (const r of tail) {
  assert.equal(r.Kind, 'trial');
  assert.equal(r.Split, header.split);
  assert(modes.includes(r.Mode));
  assert(cases.includes(r.Scenario));
  assert(Number.isInteger(r.Trial) && r.Trial >= 0 && r.Trial < 1000);
  const key = [r.Mode, r.Scenario, r.Trial].join(':');
  assert(!seen.has(key), `duplicate ${key}`);
  seen.add(key);
  assert(Number.isInteger(r.FlagAt) && r.FlagAt >= -1 && r.FlagAt < 512);
  for (const name of ['Nominated', 'Matched', 'Delivered', 'ObservedLabels',
    'LiveContexts', 'RefContexts', 'FactorUpdates']) {
    assert(Number.isInteger(r[name]) && r[name] >= 0);
  }
  assert.equal(r.LiveContexts, 512);
  assert(r.Nominated <= 512);
  assert(r.Matched <= r.Nominated);
  assert(r.Delivered <= r.Matched);
  assert(r.ObservedLabels >= 2 * r.Delivered);
  assert(r.ObservedLabels <= 2 * r.Nominated);
  assert(r.FactorUpdates % 2 === 0 && r.FactorUpdates <= 16 * r.Delivered);
  if (r.Mode === 'passive') assert.equal(r.RefContexts, r.Nominated);
  else {
    assert.equal(r.Matched, r.Nominated);
    assert(r.RefContexts >= r.Nominated);
  }
}
assert.equal(seen.size, 10000);

function quantile(a, p) {
  if (!a.length) return -1;
  const sorted = [...a].sort((x, y) => x - y);
  return sorted[Math.ceil(a.length * p) - 1];
}
const cells = [];
let pass = true;
for (const Mode of modes) for (const Scenario of cases) {
  const rows = tail.filter(r => r.Mode === Mode && r.Scenario === Scenario);
  assert.equal(rows.length, 1000);
  const flags = rows.filter(r => r.FlagAt >= 0).map(r => r.FlagAt);
  const mean = key => rows.reduce((s, r) => s + r[key], 0) / rows.length;
  const c = { Mode, Scenario, Trials: 1000,
    Flags: flags.length,
    PrechangeFlags: Scenario === 'member_shift' ? flags.filter(t => t < 256).length : 0,
    MedianFlag: quantile(flags, .5), P95Flag: quantile(flags, .95),
    MeanNominated: mean('Nominated'), MeanMatched: mean('Matched'),
    MeanDelivered: mean('Delivered'), MeanLabels: mean('ObservedLabels'),
    MeanLiveReads: mean('LiveContexts'), MeanRefReads: mean('RefContexts'),
    MeanFactors: mean('FactorUpdates') };
  c.Pass = true;
  if (['stable', 'common_shift', 'null'].includes(Scenario)) c.Pass = c.Flags <= 20;
  if (Scenario === 'member_shift') {
    c.Pass = c.PrechangeFlags === 0 && (Mode !== 'targeted' || c.Flags >= 800);
  }
  if (Mode === 'targeted' && c.MeanRefReads > 2.1 * c.MeanNominated) c.Pass = false;
  pass &&= c.Pass;
  cells.push(c);
}
assert.equal(summary.kind, 'summary');
assert.deepEqual(summary.cells, cells);
assert.equal(summary.pass, pass);
assert(Number.isFinite(summary.elapsedMs) && summary.elapsedMs > 0);
console.log(JSON.stringify({ split: header.split, trials: tail.length, pass,
  cells: cells.map(c => [c.Mode, c.Scenario, c.Flags, c.MedianFlag, c.MeanMatched,
    c.MeanDelivered, c.MeanRefReads, c.Pass]) }, null, 2));
