import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import readline from 'node:readline';
import zlib from 'node:zlib';

const root = path.resolve(import.meta.dirname, '..');
const scenarios = ['stable', 'bit2_shift', 'bit2_delayed', 'bit0_shift', 'interaction_shift', 'null', 'majority_ood'];
const file = path.join(root, 'docs/experiments/mmm-learned-contrast-v3-factorial.jsonl');
const lines = fs.readFileSync(file, 'utf8').trimEnd().split('\n').map(JSON.parse);
const manifest = lines.shift();
assert.equal(manifest.kind, 'manifest');
assert.deepEqual(manifest.cells, ['old/old', 'old/new', 'new/old', 'new/new']);
for (const [source, expected] of Object.entries(manifest.sourceHashes)) {
  const actual = crypto.createHash('sha256').update(fs.readFileSync(path.join(root, source))).digest('hex');
  assert.equal(actual, expected, `source changed: ${source}`);
}
assert.equal(lines.length, 3584);

const diagonal = new Map();
for (const [splitIndex, split] of ['design', 'confirmation'].entries()) {
  const input = path.join(root, `docs/experiments/mmm-learned-contrast-v3-${split}.jsonl.gz`);
  const hash = crypto.createHash('sha256').update(fs.readFileSync(input)).digest('hex');
  assert.equal(hash, manifest.inputHashes[splitIndex]);
  const reader = readline.createInterface({input: fs.createReadStream(input).pipe(zlib.createGunzip())});
  let header = true, count = 0;
  for await (const line of reader) {
    const row = JSON.parse(line);
    if (header) { assert.equal(row.kind, 'manifest'); header = false; continue; }
    assert.equal(row.Kind, 'trial');
    const key = `${split}:${row.Scenario}:${row.Fit}:${row.Stream}`;
    diagonal.set(key, [row.Arms[2], row.Arms[5]]);
    count++;
  }
  assert.equal(count, 1792);
}

function paired(rows, candidate, control, field) {
  const groups = Array.from({length: 16}, () => []);
  for (const row of rows) groups[row.Fit].push(row.Cells[candidate][field] - row.Cells[control][field]);
  const means = groups.map((g, fit) => {
    assert.equal(g.length, 16, `fit ${fit} lacks a stream`);
    return g.reduce((a, b) => a + b, 0) / 16;
  });
  const mean = means.reduce((a, b) => a + b, 0) / 16;
  const variance = means.reduce((a, b) => a + (b - mean) ** 2, 0) / 15;
  const radius = 3.5 * Math.sqrt(variance / 16);
  return {mean, low: mean - radius, high: mean + radius};
}

const seen = new Set();
const byCell = new Map();
for (const row of lines) {
  assert.equal(row.Kind, 'trial');
  assert.ok(['design', 'confirmation'].includes(row.Split));
  assert.ok(scenarios.includes(row.Scenario));
  assert.ok(row.Fit >= 0 && row.Fit < 16 && row.Stream >= 0 && row.Stream < 16);
  const key = `${row.Split}:${row.Scenario}:${row.Fit}:${row.Stream}`;
  assert.ok(!seen.has(key), `duplicate ${key}`);
  seen.add(key);
  assert.equal(row.Cells.length, 4);
  const originals = diagonal.get(key);
  assert.ok(originals, `missing source ${key}`);
  for (const [cellIndex, old] of [[0, originals[0]], [3, originals[1]]]) {
    const cell = row.Cells[cellIndex];
    for (const [left, right] of [['Full', 'FullBrier'], ['Post', 'PostBrier'], ['Early', 'EarlyBrier']]) {
      assert.ok(Math.abs(cell[left] - old.Data[right]) < 1e-12, `${key}/${cellIndex}/${left}`);
    }
    assert.ok(Math.abs(cell.ExpectedPost - old.PostExpected) < 1e-12);
    assert.equal(cell.RecoveryDelay, old.RecoveryDelay);
    assert.equal(cell.MissedRecovery, old.MissedRecovery);
    assert.equal(cell.Arrived, old.Data.Arrived);
  }
  for (const cell of row.Cells) {
    assert.equal(cell.Nominated, 128);
    assert.ok(cell.Arrived >= 0 && cell.Arrived <= 128);
  }
  assert.equal(row.Cells[0].Arrived, row.Cells[2].Arrived);
  assert.equal(row.Cells[1].Arrived, row.Cells[3].Arrived);
  const caseKey = `${row.Split}:${row.Scenario}`;
  if (!byCell.has(caseKey)) byCell.set(caseKey, []);
  byCell.get(caseKey).push(row);
}
assert.equal(seen.size, 3584);

const result = {kind: 'retrospective_factorial', cells: manifest.cells, splits: {}};
for (const split of ['design', 'confirmation']) {
  result.splits[split] = {};
  for (const scenario of scenarios) {
    const rows = byCell.get(`${split}:${scenario}`);
    assert.equal(rows.length, 256);
    const mean = field => [0, 1, 2, 3].map(cell => rows.reduce((sum, r) => sum + Number(r.Cells[cell][field]), 0) / 256);
    result.splits[split][scenario] = {
      post: mean('Post'), expectedPost: mean('ExpectedPost'),
      full: mean('Full'), recovery: mean('RecoveryDelay'), miss: mean('MissedRecovery'),
      contrasts: Object.fromEntries(['Post', 'ExpectedPost', 'RecoveryDelay'].map(field => [field, {
        modelOnOldSchedule: paired(rows, 2, 0, field),
        scheduleOnOldModel: paired(rows, 1, 0, field),
        scheduleOnNewModel: paired(rows, 3, 2, field),
      }]))
    };
  }
}
if (process.argv.includes('--brief')) {
  console.log(JSON.stringify({kind: result.kind, splits: Object.fromEntries(
    Object.entries(result.splits).map(([split, cells]) => [split, Object.fromEntries(
      Object.entries(cells).map(([scenario, cell]) => [scenario, {
        expectedPost: cell.expectedPost,
        recovery: cell.recovery,
        modelDelay: cell.contrasts.RecoveryDelay.modelOnOldSchedule,
        oldScheduleDelay: cell.contrasts.RecoveryDelay.scheduleOnOldModel,
        newScheduleDelay: cell.contrasts.RecoveryDelay.scheduleOnNewModel,
      }]))]))}, null, 2));
} else {
  console.log(JSON.stringify(result, null, 2));
}
