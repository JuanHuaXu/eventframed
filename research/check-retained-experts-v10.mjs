import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';

const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const [diagnosticPath, journalPath, summaryPath, outputPath] = process.argv.slice(2);
assert.ok(diagnosticPath && journalPath && summaryPath && outputPath);
const diagnosticRaw = fs.readFileSync(diagnosticPath);
const diagnostic = JSON.parse(diagnosticRaw);
const summary = JSON.parse(fs.readFileSync(summaryPath));
assert.ok(['uniform', 'latent'].includes(diagnostic.mode));
assert.equal(diagnostic.inputSHA256, hash(fs.readFileSync(journalPath)));
assert.equal(diagnostic.protocolSHA256,
  hash(fs.readFileSync('docs/experiments/mmm-retained-truth-v10-experts-protocol.md')));
assert.equal(diagnostic.sourceSHA256,
  hash(fs.readFileSync('cmd/research-retained-experts/main.go')));
assert.equal(diagnostic.records.length, 48);
assert.equal(diagnostic.cells.length, 4);
const fields = ['Final', 'Base', 'Challenger', 'Long', 'Neutral', 'Short', 'Tree', 'ObservedBits'];
let checks = 0;
for (const cell of diagnostic.cells) {
  assert.ok(['shift128', 'shift256'].includes(cell.Scenario));
  assert.ok(['post', 'early64'].includes(cell.Window));
  const records = diagnostic.records.filter(row => row.Scenario === cell.Scenario);
  assert.equal(records.length, 24);
  for (let arm = 0; arm < 4; arm++) {
    const values = records.map(row => cell.Window === 'post' ? row.Post[arm] : row.Early[arm]);
    const n = values.reduce((sum, row) => sum + row.N, 0);
    assert.equal(cell.Arms[arm].N, n);
    for (const field of fields) {
      const independent = values.reduce((sum, row) => sum + row[field], 0) / n;
      assert.ok(Math.abs(independent - cell.Arms[arm][field]) < 1e-12,
        `${cell.Scenario}/${cell.Window}/arm${arm}/${field}`);
      checks++;
    }
    if (cell.Window === 'post') {
      const source = summary.Records.filter(row => row.Split === 'confirmation' && row.Scenario === cell.Scenario);
      assert.equal(source.length, 24);
      const original = source.reduce((sum, row) => sum + row.Post[arm].Brier, 0) / 24;
      assert.ok(Math.abs(original - cell.Arms[arm].Final) < 1e-12);
      checks++;
    }
  }
}
const output = {
  mode: diagnostic.mode,
  diagnosticSHA256: hash(diagnosticRaw),
  checkerSHA256: hash(fs.readFileSync(new URL(import.meta.url))),
  records: diagnostic.records.length,
  aggregateFieldsChecked: checks,
  cells: diagnostic.cells.map(cell => ({
    scenario: cell.Scenario,
    window: cell.Window,
    n: cell.Arms[0].N,
    fixedFinal: cell.Arms[0].Final,
    shortOnFixedView: cell.Arms[0].Short,
    replacementFinal: cell.Arms[1].Final,
    treeOnReplacementView: cell.Arms[1].Tree,
    neutralOnReplacementView: cell.Arms[1].Neutral,
  })),
};
fs.writeFileSync(outputPath, JSON.stringify(output, null, 2) + '\n', {flag: 'wx'});
console.log(JSON.stringify({mode: output.mode, records: output.records, aggregateFieldsChecked: checks}));
