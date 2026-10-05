import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import readline from 'node:readline';
import zlib from 'node:zlib';

const root = path.resolve(import.meta.dirname, '..');
const cases = ['stable', 'bit2_shift', 'bit2_delayed', 'bit0_shift', 'interaction_shift', 'null', 'majority_ood'];
const correctRule = {bit2_shift: 3, bit2_delayed: 3, bit0_shift: 1, interaction_shift: 10};
const isChanged = c => c !== 'stable' && c !== 'null';

function near(a, b, label) {
  assert.ok(Number.isFinite(a) && Number.isFinite(b) && Math.abs(a - b) < 1e-9,
    `${label}: ${a} != ${b}`);
}

function truth(c, x, clock) {
  if (c === 'null') return .5;
  let bit = Boolean((x >> 6) & 1) !== Boolean((x >> 7) & 1);
  bit = bit !== Boolean((x >> 8) & 1);
  if (clock >= 256) {
    if (c === 'bit2_shift' || c === 'bit2_delayed') bit = Boolean((x >> 2) & 1);
    if (c === 'bit0_shift') bit = Boolean(x & 1);
    if (c === 'interaction_shift') bit = Boolean(x & 1) !== Boolean((x >> 1) & 1) !== Boolean((x >> 2) & 1);
    if (c === 'majority_ood') bit = (Number(Boolean(x & 1)) + Number(Boolean((x >> 1) & 1)) + Number(Boolean((x >> 2) & 1))) >= 2;
  }
  return bit ? .95 : .05;
}

function interval(rows, candidate, control, field, cluster = false) {
  const diffs = rows.map(r => r.Arms[control][field] - r.Arms[candidate][field]);
  let values = diffs;
  if (cluster) {
    const groups = Array.from({length: 16}, () => []);
    rows.forEach((r, i) => groups[r.Fit].push(diffs[i]));
    values = groups.map((g, i) => {
      assert.equal(g.length, 16, `missing fit cluster ${i}`);
      return g.reduce((a, b) => a + b, 0) / 16;
    });
  }
  const mean = values.reduce((a, b) => a + b, 0) / values.length;
  const variance = values.reduce((a, b) => a + (b - mean) ** 2, 0) / (values.length - 1);
  const radius = 3.5 * Math.sqrt(variance / values.length);
  return {mean, low: mean - radius, high: mean + radius};
}

function recover(losses) {
  let consecutive = 0;
  for (let clock = 287; clock < 512; clock++) {
    let mean = 0;
    for (let j = clock - 31; j <= clock; j++) mean += losses[j] / 32;
    consecutive = mean <= .12 ? consecutive + 1 : 0;
    if (consecutive === 16) return {delay: clock - 256, missed: false};
  }
  return {delay: 256, missed: true};
}

const report = {splits: {}};
for (const split of ['design', 'confirmation']) {
  const file = path.join(root, `docs/experiments/mmm-learned-contrast-v3-${split}.jsonl.gz`);
  const reader = readline.createInterface({input: fs.createReadStream(file).pipe(zlib.createGunzip())});
  const byCase = new Map(cases.map(c => [c, []]));
  const identities = new Set();
  let manifest, count = 0;
  for await (const line of reader) {
    const row = JSON.parse(line);
    if (!manifest) {
      manifest = row;
      assert.equal(row.kind, 'manifest');
      assert.equal(row.split, split);
      assert.equal(row.seedBase, split === 'design' ? 2026102301 : 2026102302);
      assert.equal(row.fitOffset, split === 'design' ? 600 : 700);
      assert.equal(row.budget, 128);
      assert.equal(row.trialsPerCase, 256);
      assert.ok(row.stateBytes > 0 && row.stateBytes < 4096);
      near(row.prior.reduce((a, b) => a + b, 0), 1, 'prior sum');
      for (const [source, expected] of Object.entries(row.hashes)) {
        const actual = crypto.createHash('sha256').update(fs.readFileSync(path.join(root, source))).digest('hex');
        assert.equal(actual, expected, `source changed: ${source}`);
      }
      continue;
    }
    assert.equal(row.Kind, 'trial');
    assert.equal(row.Split, split);
    assert.ok(byCase.has(row.Scenario));
    assert.ok(row.Fit >= 0 && row.Fit < 16 && row.Stream >= 0 && row.Stream < 16);
    const id = `${row.Scenario}:${row.Fit}:${row.Stream}`;
    assert.ok(!identities.has(id), `duplicate ${id}`);
    identities.add(id);
    assert.equal(row.Tape.length, 512);
    assert.equal(row.Arms.length, 6);
    const delay = row.Scenario === 'bit2_delayed' ? 16 : 0;
    const summary = {Fit: row.Fit, Arms: []};
    for (let clock = 0; clock < 512; clock++) {
      const event = row.Tape[clock];
      assert.ok(Number.isInteger(event.X) && event.X >= 0 && event.X < 512);
      near(event.PTrue, truth(row.Scenario, event.X, clock), `${id}/truth/${clock}`);
    }
    for (const [armIndex, wrapped] of row.Arms.entries()) {
      const arm = wrapped.Data;
      assert.equal(arm.Policy, ['random', 'uncertainty', 'learned_disagreement'][armIndex % 3]);
      assert.equal(arm.Ticks.length, 512);
      let full = 0, post = 0, early = 0, expectedPost = 0;
      let nominated = 0, nominatedPost = 0, correctPost = 0, arrived = 0, pending = 0;
      const selected = new Set(), delivered = new Set(), expected = [];
      for (let clock = 0; clock < 512; clock++) {
        const event = row.Tape[clock], tick = arm.Ticks[clock];
        assert.ok(Number.isFinite(tick.P) && tick.P > 0 && tick.P < 1);
        assert.ok(tick.Alternate >= 1 && tick.Alternate <= 10);
        const score = (tick.P - Number(event.Y)) ** 2;
        const properExpectation = event.PTrue*(1-event.PTrue) + (tick.P-event.PTrue)**2;
        expected.push(properExpectation);
        full += score / 512;
        if (clock >= 256) {
          post += score / 256;
          expectedPost += properExpectation / 256;
          if (clock < 320) early += score / 64;
        }
        if (tick.Selected) {
          selected.add(clock);
          nominated++;
          if (clock >= 256) {
            nominatedPost++;
            if (tick.Alternate === correctRule[row.Scenario]) correctPost++;
          }
          if (!event.Missing && clock + delay >= 512) pending++;
        }
        for (const origin of tick.Delivered ?? []) {
          assert.ok(origin <= clock && selected.has(origin), 'future or unrequested label');
          assert.ok(!row.Tape[origin].Missing, 'missing label delivered');
          assert.equal(origin + delay, clock, 'wrong delivery clock');
          assert.ok(!delivered.has(origin), 'duplicate label');
          delivered.add(origin);
          arrived++;
        }
      }
      for (const origin of selected) {
        if (!row.Tape[origin].Missing && origin + delay < 512) assert.ok(delivered.has(origin));
      }
      assert.equal(nominated, 128);
      assert.equal(arm.Nominated, 128);
      assert.equal(arm.Arrived, arrived);
      assert.equal(arm.Pending, pending);
      assert.equal(arm.NominatedPost, nominatedPost);
      assert.equal(arm.CorrectPost, correctPost);
      near(full, arm.FullBrier, `${id}/${armIndex}/full`);
      near(post, arm.PostBrier, `${id}/${armIndex}/post`);
      near(early, arm.EarlyBrier, `${id}/${armIndex}/early`);
      near(expectedPost, wrapped.PostExpected, `${id}/${armIndex}/expected`);
      const recovery = isChanged(row.Scenario) ? recover(expected) : {delay: -1, missed: false};
      assert.equal(wrapped.RecoveryDelay, recovery.delay);
      assert.equal(wrapped.MissedRecovery, recovery.missed);
      summary.Arms.push({FullBrier: full, PostBrier: post, EarlyBrier: early,
        PostExpected: expectedPost, RecoveryDelay: recovery.delay,
        MissedRecovery: recovery.missed, Arrived: arrived});
    }
    for (const a of [0, 1]) {
      for (let clock = 0; clock < 512; clock++) {
        assert.equal(row.Arms[a].Data.Ticks[clock].Selected,
          row.Arms[a+3].Data.Ticks[clock].Selected, 'paired ablation changed selection schedule');
      }
    }
    byCase.get(row.Scenario).push(summary);
    count++;
  }
  assert.ok(manifest);
  assert.equal(count, 1792);
  const splitReport = {trajectories: count, cases: {}, pass: true};
  for (const scenario of cases) {
    const rows = byCase.get(scenario);
    assert.equal(rows.length, 256);
    const mean = field => Array.from({length: 6}, (_, arm) => rows.reduce((sum, r) => sum + Number(r.Arms[arm][field]), 0) / 256);
    const full = mean('FullBrier'), post = mean('PostBrier'), early = mean('EarlyBrier');
    const expected = mean('PostExpected'), recovery = mean('RecoveryDelay'), missed = mean('MissedRecovery');
    const arrived = mean('Arrived');
    const gains = {
      newLearnedVsRandom: interval(rows, 5, 3, 'PostBrier', true),
      newLearnedVsUncertainty: interval(rows, 5, 4, 'PostBrier', true),
      newLearnedVsOld: interval(rows, 5, 2, 'PostBrier', true),
      newLearnedExpectedVsOld: interval(rows, 5, 2, 'PostExpected', true),
    };
    let pass = true;
    if (scenario === 'null') {
      pass = [0, 1, 2].every(i => full[i+3] <= .27 &&
        interval(rows, i+3, i, 'FullBrier', true).mean >= .05 &&
        interval(rows, i+3, i, 'FullBrier', true).low > 0);
    }
    if (scenario === 'stable') pass = interval(rows, 5, 2, 'FullBrier', true).low > -.01;
    if (scenario === 'bit2_shift' || scenario === 'bit2_delayed') {
      pass = [3, 4].every(i => {
        const gain = interval(rows, 5, i, 'PostBrier', true);
        return gain.mean >= .01 && gain.low > 0 && early[i] - early[5] >= -.005 &&
          recovery[i] - recovery[5] >= 10 && missed[5] <= missed[i] + .05;
      });
    }
    if (['bit2_shift', 'bit2_delayed', 'bit0_shift', 'interaction_shift'].includes(scenario)) {
      pass &&= gains.newLearnedVsOld.low > -.01;
    }
    if (scenario === 'majority_ood') {
      pass = gains.newLearnedExpectedVsOld.mean >= .005 && gains.newLearnedExpectedVsOld.low > 0;
    }
    pass &&= Math.abs(arrived[5] - arrived[3]) <= 2 && Math.abs(arrived[5] - arrived[4]) <= 2;
    splitReport.cases[scenario] = {full, post, early, expected, recovery, missed, arrived, gains, pass};
    splitReport.pass &&= pass;
  }
  report.splits[split] = splitReport;
}
report.efficacyPass = report.splits.design.pass && report.splits.confirmation.pass;
if (process.argv.includes('--brief')) {
  console.log(JSON.stringify({efficacyPass: report.efficacyPass,
    splits: Object.fromEntries(Object.entries(report.splits).map(([name, split]) => [name, {
      pass: split.pass, cases: Object.fromEntries(Object.entries(split.cases).map(([c, v]) => [c, {
        pass: v.pass, full: v.full, post: v.post, expected: v.expected,
        recovery: v.recovery, missed: v.missed,
        postGain: [v.gains.newLearnedVsRandom.mean, v.gains.newLearnedVsUncertainty.mean],
        fitLow: [v.gains.newLearnedVsRandom.low, v.gains.newLearnedVsUncertainty.low],
        expectedVsOld: v.gains.newLearnedExpectedVsOld,
      }]))
    }]))}, null, 2));
} else {
  console.log(JSON.stringify(report, null, 2));
}
