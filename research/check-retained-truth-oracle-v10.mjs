import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const [inputPath, outputPath] = process.argv.slice(2);
assert.ok(inputPath && outputPath);
const raw = fs.readFileSync(inputPath);
const result = JSON.parse(raw);
assert.ok(['uniform', 'latent'].includes(result.mode));
assert.equal(result.records.length, 48);
assert.equal(result.cells.length, 4);
const keys = ['ModelBrier', 'ViewOracleBrier', 'FullOracleBrier', 'MeanObservedBits'];
let checked = 0;
for (const cell of result.cells) {
  const rows = result.records.filter(row => row.Scenario === cell.Scenario);
  assert.equal(rows.length, 24);
  assert.ok(['post', 'early64'].includes(cell.Window));
  for (let arm = 0; arm < 4; arm++) {
    const source = rows.map(row => cell.Window === 'post' ? row.Post[arm] : row.Early[arm]);
    const n = source.reduce((sum, value) => sum + value.N, 0);
    assert.equal(cell.Arms[arm].N, n);
    for (const key of keys) {
      const recomputed = source.reduce((sum, value) => sum + value[key], 0) / n;
      assert.ok(Math.abs(recomputed - cell.Arms[arm][key]) < 1e-12,
        `${cell.Scenario}/${cell.Window}/arm${arm}/${key}`);
      checked++;
    }
    for (const key of keys) assert.ok(Number.isFinite(cell.Arms[arm][key]));
    assert.ok(cell.Arms[arm].MeanObservedBits >= 0 && cell.Arms[arm].MeanObservedBits <= 6);
  }
}
const summary = {
  mode: result.mode,
  inputSHA256: hash(raw),
  checkerSHA256: hash(fs.readFileSync(new URL(import.meta.url))),
  records: result.records.length,
  aggregateFieldsChecked: checked,
  cells: result.cells.map(cell => ({
    scenario: cell.Scenario,
    window: cell.Window,
    n: cell.Arms[0].N,
    fixedModel: cell.Arms[0].ModelBrier,
    fixedViewOracle: cell.Arms[0].ViewOracleBrier,
    replacementModel: cell.Arms[1].ModelBrier,
    replacementViewOracle: cell.Arms[1].ViewOracleBrier,
    fullInputOracle: cell.Arms[0].FullOracleBrier,
  })),
};
fs.writeFileSync(outputPath, JSON.stringify(summary, null, 2) + '\n', {flag: 'wx'});
console.log(JSON.stringify({mode: summary.mode, records: summary.records, aggregateFieldsChecked: checked}));
