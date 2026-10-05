import assert from 'node:assert/strict';
import fs from 'node:fs';

const audit = JSON.parse(fs.readFileSync(process.argv[2], 'utf8'));
assert.equal(audit.pass, true);
assert.equal(audit.comparisons.length, 84);
assert.equal(audit.descriptive.length, 4);
const summary = xs => {
  assert(xs.every(Number.isFinite));
  const sorted = xs.slice().sort((a, b) => a - b);
  return {min: sorted[0], median: (sorted[41] + sorted[42]) / 2,
    max: sorted.at(-1), mean: xs.reduce((s, x) => s + x, 0) / xs.length};
};
const groups = [];
for (const phase of [0, 1]) for (const schedule of [0, 1]) {
  const rows = audit.comparisons.filter(r => r.phase === phase && r.schedule === schedule);
  assert.equal(rows.length, 21);
  assert.equal(new Set(rows.map(r => r.scenario)).size, 21);
  const cell = audit.descriptive.find(c => c.key === `${phase}/${schedule}/64/128`);
  assert(cell);
  const mean = key => rows.reduce((s, r) => s + r[key], 0) / rows.length;
  const extended = mean('extendedBrier');
  assert(Math.abs(extended - cell.brier[0]) < 1e-12);
  groups.push({phase, schedule, cappedBrier: mean('cappedBrier'), extendedBrier: extended,
    pluginBrier: cell.brier[1], genericBrier: cell.brier[2], booleanBrier: cell.brier[3], markovBrier: cell.brier[14],
    improvedCases: rows.filter(r => r.extendedBrier < r.cappedBrier).length,
    worsenedCases: rows.filter(r => r.extendedBrier > r.cappedBrier).length});
}
const result = {fits: 84, forecasts: 2688, stops: audit.stops,
  iterations: summary(audit.comparisons.map(r => r.iterations)),
  boundGain: summary(audit.comparisons.map(r => r.boundGain)),
  lastIncrement: summary(audit.comparisons.map(r => r.lastIncrement)), groups,
  interpretation: 'Matched descriptive consumed pilot; no trajectory confidence intervals, general quality validation, or stationary-point guarantee.'};
if (process.argv[3]) fs.writeFileSync(process.argv[3], JSON.stringify(result, null, 2) + '\n', {flag: 'wx', mode: 0o600});
console.log(JSON.stringify(result, null, 2));
