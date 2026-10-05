import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';

const root = path.resolve(import.meta.dirname, '..');
const cases = ['stable05', 'shift128', 'gradual', 'delayed_missing', 'interaction'];
const change = {stable05: 0, shift128: 128, gradual: 128, delayed_missing: 256, interaction: 256};
const splits = ['design', 'confirmation'];
const names = ['base', 'last64', 'q0005', 'q002', 'q01'];
const records = new Map();

function near(a, b, where) {
  assert.ok(Number.isFinite(a) && Number.isFinite(b) && Math.abs(a - b) < 1e-10, `${where}: ${a} != ${b}`);
}

function paired(rows, candidate, control, segment) {
  const differences = rows.map(row => row.Metric[segment][control] - row.Metric[segment][candidate]);
  const mean = differences.reduce((a, b) => a + b, 0) / differences.length;
  const variance = differences.reduce((a, b) => a + (b - mean) ** 2, 0) / (differences.length - 1);
  const radius = 3.5 * Math.sqrt(variance / differences.length);
  return {mean, low: mean - radius, high: mean + radius};
}

for (const split of splits) {
  const file = path.join(root, `docs/experiments/mmm-kalman-early-v2-${split}.jsonl`);
  const lines = fs.readFileSync(file, 'utf8').trimEnd().split('\n').map(JSON.parse);
  const manifest = lines.shift();
  assert.equal(manifest.kind, 'manifest');
  assert.equal(manifest.split, split);
  assert.equal(manifest.seedBase, split === 'design' ? 2026102101 : 2026102102);
  assert.equal(manifest.fitOffset, split === 'design' ? 200 : 300);
  assert.deepEqual(manifest.processNoise, [0.0005, 0.002, 0.01]);
  assert.deepEqual(manifest.scenarios, [0, 2, 5, 7, 8]);
  for (const [source, hash] of Object.entries(manifest.hashes)) {
    const actual = crypto.createHash('sha256').update(fs.readFileSync(path.join(root, source))).digest('hex');
    assert.equal(actual, hash, `${source} changed after collection`);
  }
  assert.equal(lines.length, 160);
  const keys = new Set();
  for (const row of lines) {
    assert.equal(row.Kind, 'trial');
    assert.equal(row.Split, split);
    assert.ok(cases.includes(row.Scenario));
    assert.ok(row.Fit >= 0 && row.Fit < 4 && row.Stream >= 0 && row.Stream < 8);
    const key = `${row.Scenario}:${row.Fit}:${row.Stream}`;
    assert.ok(!keys.has(key), `duplicate ${key}`);
    keys.add(key);
    assert.equal(row.Ticks.length, 512);
    assert.equal(row.StateBytes, 1056);
    assert.equal(row.Updates, row.Audits);
    const due = new Map();
    const loss = {Full: Array(5).fill(0), Early: Array(5).fill(0), Tail: Array(5).fill(0)};
    let available = 0, audits = 0;
    for (let clock = 0; clock < 512; clock++) {
      const tick = row.Ticks[clock];
      assert.ok(tick.X >= 0 && tick.X < 512);
      assert.equal(tick.P.length, 5);
      if (!tick.Missing) {
        const deliveryClock = clock + (row.Scenario === 'delayed_missing' ? 16 : 0);
        due.set(clock, deliveryClock);
      }
      for (const origin of tick.Delivered ?? []) {
        assert.ok(origin <= clock, 'future outcome');
        assert.equal(due.get(origin), clock, 'delivery mismatch');
        due.delete(origin);
        available++;
        if (row.Ticks[origin].Audit) audits++;
      }
      for (const [arm, p] of tick.P.entries()) {
        assert.ok(Number.isFinite(p) && p >= 0 && p <= 1);
        const z = tick.Y ? 1 : 0;
        const score = (p - z) ** 2;
        loss.Full[arm] += score / 512;
        if (clock >= change[row.Scenario] && clock < change[row.Scenario] + 64) loss.Early[arm] += score / 64;
        if (clock >= 384) loss.Tail[arm] += score / 128;
      }
    }
    assert.equal(available, row.Available);
    assert.equal(audits, row.Audits);
    assert.equal(due.size, row.Pending);
    for (const segment of ['Full', 'Early', 'Tail']) {
      for (let arm = 0; arm < 5; arm++) near(loss[segment][arm], row.Metric[segment][arm], `${key}/${segment}/${arm}`);
    }
    if (row.Scenario === 'delayed_missing') {
      for (let clock = 0; clock <= 16; clock++) {
        for (let arm = 2; arm < 5; arm++) near(row.Ticks[clock].P[arm], row.Ticks[clock].P[0], `${key}/unarrived/${clock}/${arm}`);
      }
    }
  }
  records.set(split, lines);
}

const report = {status: 'audited', names, cohort: {}, decision: {}};
for (const split of splits) {
  report.cohort[split] = {};
  report.decision[split] = {};
  for (const scenario of cases) {
    const rows = records.get(split).filter(row => row.Scenario === scenario);
    assert.equal(rows.length, 32);
    const cell = {mean: {}, paired: {}};
    for (const segment of ['Full', 'Early', 'Tail']) {
      cell.mean[segment] = names.map((_, arm) => rows.reduce((sum, row) => sum + row.Metric[segment][arm], 0) / 32);
    }
    for (const candidate of [3, 4]) {
      cell.paired[names[candidate]] = {
        earlyVsDefault: paired(rows, candidate, 2, 'Early'),
        earlyVsLast64: paired(rows, candidate, 1, 'Early'),
        tailVsLast64: paired(rows, candidate, 1, 'Tail'),
        fullVsBase: paired(rows, candidate, 0, 'Full'),
      };
    }
    cell.meanAudits = rows.reduce((sum, row) => sum + row.Audits, 0) / 32;
    cell.maxUpdateP99NS = [0, 1, 2].map(i => Math.max(...rows.map(row => row.UpdateP99NS[i])));
    report.cohort[split][scenario] = cell;
  }
  for (const candidate of ['q002', 'q01']) {
    const c = scenario => report.cohort[split][scenario].paired[candidate];
    const early = ['shift128', 'delayed_missing'].every(s => c(s).earlyVsDefault.mean >= .01 &&
      c(s).earlyVsDefault.low > 0 && c(s).earlyVsLast64.mean >= -.005);
    const tail = ['shift128', 'gradual', 'delayed_missing'].every(s => c(s).tailVsLast64.mean >= -.005);
    const stable = c('stable05').fullVsBase.low > -.01;
    const interaction = c('interaction').tailVsLast64.low > -.01;
    const cost = cases.every(s => report.cohort[split][s].maxUpdateP99NS[names.indexOf(candidate) - 2] < 50000);
    report.decision[split][candidate] = {early, tail, stable, interaction, cost,
      pass: early && tail && stable && interaction && cost};
  }
}
report.overallPass = ['q002', 'q01'].some(q => splits.every(s => report.decision[s][q].pass));
console.log(JSON.stringify(report, null, 2));
