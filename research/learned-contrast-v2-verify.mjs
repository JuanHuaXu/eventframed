import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import readline from 'node:readline';
import zlib from 'node:zlib';

const root = path.resolve(import.meta.dirname, '..');
const cases = ['stable', 'bit2_shift', 'bit2_delayed', 'bit0_shift', 'interaction_shift', 'null'];
const policies = ['random', 'uncertainty', 'learned_disagreement'];
const correctAlt = {bit2_shift: 3, bit2_delayed: 3, bit0_shift: 1, interaction_shift: 10};
const report = {splits: {}, overallPass: false};

function close(a, b, label) {
  assert.ok(Number.isFinite(a) && Number.isFinite(b) && Math.abs(a - b) < 1e-10,
    `${label}: ${a} != ${b}`);
}

function paired(rows, field, control) {
  const differences = rows.map(r => r.Arms[control][field] - r.Arms[2][field]);
  const mean = differences.reduce((a, b) => a + b, 0) / differences.length;
  const variance = differences.reduce((a, b) => a + (b - mean) ** 2, 0) / (differences.length - 1);
  const radius = 3.5 * Math.sqrt(variance / differences.length);
  return {mean, low: mean - radius, high: mean + radius};
}

function fitClusterPaired(rows, field, control) {
  const clusters = Array.from({length: 16}, () => []);
  for (const row of rows) clusters[row.Fit].push(row.Arms[control][field] - row.Arms[2][field]);
  const means = clusters.map((cluster, fit) => {
    assert.equal(cluster.length, 16, `missing stream in fit ${fit}`);
    return cluster.reduce((a, b) => a + b, 0) / cluster.length;
  });
  const mean = means.reduce((a, b) => a + b, 0) / means.length;
  const variance = means.reduce((a, b) => a + (b - mean) ** 2, 0) / (means.length - 1);
  const radius = 3.5 * Math.sqrt(variance / means.length);
  return {mean, low: mean - radius, high: mean + radius};
}

for (const split of ['design', 'confirmation']) {
  const file = path.join(root, `docs/experiments/mmm-learned-contrast-v2-${split}.jsonl.gz`);
  const reader = readline.createInterface({input: fs.createReadStream(file).pipe(zlib.createGunzip())});
  const rows = new Map(cases.map(c => [c, []]));
  const keys = new Set();
  let manifest, count = 0;
  for await (const line of reader) {
    const row = JSON.parse(line);
    if (!manifest) {
      manifest = row;
      assert.equal(row.kind, 'manifest');
      assert.equal(row.split, split);
      assert.equal(row.seedBase, split === 'design' ? 2026102201 : 2026102202);
      assert.equal(row.fitOffset, split === 'design' ? 400 : 500);
      assert.equal(row.budget, 128);
      assert.equal(row.trialsPerCase, 256);
      close(row.hypothesisPrior.reduce((a, b) => a + b, 0), 1, 'prior sum');
      for (const [source, expected] of Object.entries(row.hashes)) {
        const actual = crypto.createHash('sha256').update(fs.readFileSync(path.join(root, source))).digest('hex');
        assert.equal(actual, expected, `source changed: ${source}`);
      }
      continue;
    }
    assert.equal(row.Kind, 'trial');
    assert.equal(row.Split, split);
    assert.ok(rows.has(row.Scenario));
    assert.ok(row.Fit >= 0 && row.Fit < 16 && row.Stream >= 0 && row.Stream < 16);
    const key = `${row.Scenario}:${row.Fit}:${row.Stream}`;
    assert.ok(!keys.has(key), `duplicate ${key}`);
    keys.add(key);
    assert.equal(row.Tape.length, 512);
    assert.equal(row.Arms.length, 3);
    const delay = row.Scenario === 'bit2_delayed' ? 16 : 0;
    for (const [armIndex, arm] of row.Arms.entries()) {
      assert.equal(arm.Policy, policies[armIndex]);
      assert.equal(arm.Ticks.length, 512);
      let full = 0, post = 0, early = 0, nominated = 0, nominatedPost = 0;
      let arrived = 0, pending = 0, correctPost = 0;
      const selected = new Set(), delivered = new Set();
      for (let clock = 0; clock < 512; clock++) {
        const tick = arm.Ticks[clock], event = row.Tape[clock];
        assert.ok(event.X >= 0 && event.X < 512);
        assert.ok(Number.isFinite(tick.P) && tick.P > 0 && tick.P < 1);
        assert.ok(tick.Alternate >= 1 && tick.Alternate <= 10);
        const loss = (tick.P - Number(event.Y)) ** 2;
        full += loss / 512;
        if (clock >= 256) {
          post += loss / 256;
          if (clock < 320) early += loss / 64;
        }
        if (tick.Selected) {
          selected.add(clock);
          nominated++;
          if (clock >= 256) {
            nominatedPost++;
            if (tick.Alternate === correctAlt[row.Scenario]) correctPost++;
          }
          if (!event.Missing && clock + delay >= 512) pending++;
        }
        for (const origin of tick.Delivered ?? []) {
          assert.ok(origin <= clock, 'future label');
          assert.ok(selected.has(origin), 'unrequested label');
          assert.ok(!row.Tape[origin].Missing, 'missing label delivered');
          assert.equal(origin + delay, clock, 'incorrect delivery time');
          assert.ok(!delivered.has(origin), 'duplicate label');
          delivered.add(origin);
          arrived++;
        }
      }
      assert.equal(nominated, 128);
      assert.equal(arm.Nominated, 128);
      assert.equal(arm.Arrived, arrived);
      assert.equal(arm.Pending, pending);
      assert.equal(arm.NominatedPost, nominatedPost);
      assert.equal(arm.CorrectPost, correctPost);
      for (const origin of selected) {
        if (!row.Tape[origin].Missing && origin + delay < 512) {
          assert.ok(delivered.has(origin), 'requested available label not delivered');
        }
      }
      close(full, arm.FullBrier, `${key}/${arm.Policy}/full`);
      close(post, arm.PostBrier, `${key}/${arm.Policy}/post`);
      close(early, arm.EarlyBrier, `${key}/${arm.Policy}/early`);
    }
    rows.get(row.Scenario).push({Fit: row.Fit, Arms: row.Arms.map(arm => ({
      FullBrier: arm.FullBrier, PostBrier: arm.PostBrier,
      EarlyBrier: arm.EarlyBrier, Arrived: arm.Arrived,
      NominatedPost: arm.NominatedPost, CorrectPost: arm.CorrectPost,
    }))});
    count++;
  }
  assert.ok(manifest);
  assert.equal(count, 1536);
  const out = {trajectories: count, cases: {}, pass: true};
  for (const scenario of cases) {
    const group = rows.get(scenario);
    assert.equal(group.length, 256);
    const means = policies.map((_, arm) => ({
      full: group.reduce((sum, row) => sum + row.Arms[arm].FullBrier, 0) / 256,
      post: group.reduce((sum, row) => sum + row.Arms[arm].PostBrier, 0) / 256,
      early: group.reduce((sum, row) => sum + row.Arms[arm].EarlyBrier, 0) / 256,
      arrived: group.reduce((sum, row) => sum + row.Arms[arm].Arrived, 0) / 256,
    }));
    const comparisons = [0, 1].map(control => ({
      control: policies[control],
      full: paired(group, 'FullBrier', control),
      post: paired(group, 'PostBrier', control),
      postFitCluster: fitClusterPaired(group, 'PostBrier', control),
      early: paired(group, 'EarlyBrier', control),
      arrivedGap: Math.abs(means[control].arrived - means[2].arrived),
    }));
    const nominated = group.reduce((sum, row) => sum + row.Arms[2].NominatedPost, 0);
    const correct = group.reduce((sum, row) => sum + row.Arms[2].CorrectPost, 0);
    const identification = nominated ? correct / nominated : null;
    let pass = true;
    if (scenario === 'bit2_shift' || scenario === 'bit2_delayed') {
      pass = comparisons.every(c => c.post.mean >= .01 && c.post.low > 0 && c.early.mean >= -.005);
    } else if (scenario === 'stable' || scenario === 'null') {
      pass = comparisons.every(c => c.full.low > -.01);
    } else {
      pass = comparisons.every(c => c.post.low > -.01);
    }
    if (['bit2_shift', 'bit2_delayed', 'bit0_shift'].includes(scenario)) pass &&= identification >= .60;
    pass &&= comparisons.every(c => c.arrivedGap <= 2);
    out.cases[scenario] = {means, comparisons, identification, identificationN: nominated, pass};
    out.pass &&= pass;
  }
  report.splits[split] = out;
}
report.overallPass = report.splits.design.pass && report.splits.confirmation.pass;
if (process.argv.includes('--brief')) {
  console.log(JSON.stringify({overallPass: report.overallPass, splits: Object.fromEntries(
    Object.entries(report.splits).map(([name, split]) => [name, {
      pass: split.pass,
      cases: Object.fromEntries(Object.entries(split.cases).map(([scenario, cell]) => [scenario, {
        pass: cell.pass,
        postGain: cell.comparisons.map(c => c.post.mean),
        postFitClusterLow: cell.comparisons.map(c => c.postFitCluster.low),
        identification: cell.identification,
        identificationN: cell.identificationN,
      }]))
    }]))}, null, 2));
} else {
  console.log(JSON.stringify(report, null, 2));
}
