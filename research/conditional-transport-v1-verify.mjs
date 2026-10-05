import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';

const path = process.argv[2];
assert(path, 'JSONL path required');
const lines = fs.readFileSync(path, 'utf8').trimEnd().split('\n').map(line => JSON.parse(line));
const header = lines.shift();
const summary = lines.pop();
assert.equal(header.kind, 'header');
assert.equal(summary.kind, 'summary');
assert.equal(header.version, 'conditional-transport-v1');
assert.equal(header.protocolSHA256, crypto.createHash('sha256').update(fs.readFileSync('docs/experiments/mmm-conditional-transport-v1-protocol.md')).digest('hex'));
assert.equal(header.sourceSHA256, crypto.createHash('sha256').update(fs.readFileSync('research/conditional-transport-v1.mjs')).digest('hex'));
assert.equal(lines.length, header.trials * header.splits.length * header.schedules.length * header.cases.length);

const byCell = new Map();
const seen = new Set();
for (const row of lines) {
  assert.equal(row.kind, 'trial');
  const key = [row.split, row.schedule, row.scenario].join('|');
  const identity = `${key}|${row.trial}`;
  assert(!seen.has(identity), `duplicate ${identity}`);
  seen.add(identity);
  assert(row.trial >= 0 && row.trial < header.trials);
  assert(row.flagAt === null || row.flagAt >= 0 && row.flagAt < header.T);
  assert(row.audits >= 0 && row.audits <= header.T);
  assert(row.missing >= 0 && row.pending >= 0 && row.delivered >= 0);
  assert.equal(row.delivered + row.pending + row.missing, 2 * row.audits);
  assert.equal(row.counts.flat().reduce((a, b) => a + b, 0), row.delivered);
  for (let side = 0; side < 2; side++) for (let cell = 0; cell < 2; cell++) {
    assert(Number.isInteger(row.counts[side][cell]) && row.counts[side][cell] >= 0);
    assert(Number.isInteger(row.successes[side][cell]) && row.successes[side][cell] >= 0);
    assert(row.successes[side][cell] <= row.counts[side][cell]);
  }
  if (!byCell.has(key)) byCell.set(key, []);
  byCell.get(key).push(row);
}

function mean(rows, f) {
  return rows.reduce((s, r) => s + f(r), 0) / rows.length;
}
function median(values) {
  return values.length ? values.sort((a, b) => a - b)[Math.ceil(values.length / 2) - 1] : null;
}
for (const cell of summary.cells) {
  const key = [cell.split, cell.schedule, cell.scenario].join('|');
  const rows = byCell.get(key);
  assert.equal(rows?.length, header.trials, key);
  const flags = rows.filter(r => r.flagAt !== null).map(r => r.flagAt);
  assert.equal(cell.flags, flags.length, key);
  assert.equal(cell.flagP50, median(flags), key);
  for (const [field, rowField] of [['meanAudits', 'audits'], ['meanMissing', 'missing'], ['meanDelivered', 'delivered'], ['meanPending', 'pending']]) {
    assert(Math.abs(cell[field] - mean(rows, r => r[rowField])) < 1e-10, `${key} ${field}`);
  }
  for (let side = 0; side < 2; side++) for (let context = 0; context < 2; context++) {
    assert(Math.abs(cell.meanCounts[side][context] - mean(rows, r => r.counts[side][context])) < 1e-10, `${key} counts`);
  }
  const nullCase = cell.scenario === 'stable' || cell.scenario === 'common_shift256';
  assert.equal(cell.pass, nullCase ? flags.length <= 20 : flags.length >= 800, key);
}
assert.equal(summary.cells.length, byCell.size);
assert.equal(summary.pass, summary.cells.every(c => c.pass));
console.log(`PASS: ${lines.length} unique rows, independent summary, timing and label accounting, source/protocol hashes`);
