// Descriptive decomposition of the independently verified, consumed readback.
// This neither changes the original gates nor selects/refits a policy.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root = 'research/switch-v43-study-normal';
const hash = b => crypto.createHash('sha256').update(b).digest('hex');
const bytes = fs.readFileSync(root + '/readback.json'), r = JSON.parse(bytes);
assert.equal(r.stage, 'normal'); assert.equal(r.qualityAdoption, false);
const summaries = {};
for (const [split, v] of Object.entries(r.summaries)) {
  assert.equal(v.worlds, 448); assert.equal(Object.keys(v.cells).length, 84);
  summaries[split] = {};
  for (const mode of ['static', 'slow', 'round']) {
    const cells = Object.values(v.cells).map(c => c[mode]), badChecks = {};
    assert.equal(cells.filter(c => !c.pass).length, v.failedCells[mode]);
    for (const cell of cells) for (const [key, check] of Object.entries(cell.checks)) {
      if (check === false || check?.pass === false) badChecks[key] = (badChecks[key] ?? 0) + 1;
    }
    const qualityOnlyPassingCells = cells.filter(c => Object.entries(c.checks)
      .every(([key, check]) => ['Work', 'Allocation'].includes(key) || check.pass)).length;
    const contrasts = Object.values(v.pairedContrasts).map(c => c[mode]);
    const mean = key => contrasts.reduce((sum, c) => sum + c[key].mean / contrasts.length, 0);
    summaries[split][mode] = {failedCells: v.failedCells[mode], qualityOnlyPassingCells,
      badChecks, descriptiveMeanFullIssuedGain: mean('full/IssuedBrier'),
      descriptiveMeanAdaptiveIssuedGain: mean('adaptive/IssuedBrier')};
  }
}
const report = {time: new Date().toISOString(), readbackSHA256: hash(bytes),
  scriptSHA256: hash(fs.readFileSync('research/switch-v43-quality-breakdown.mjs')), summaries,
  sharedWorkGateUsesAllArmsInCell: true, macroMeansAreNotConfidenceCertificates: true,
  noGateChanges: true, noSelectionOrRefit: true, qualityAdoption: false,
  allSevenWholeGoals: 'OPEN', goal: 'ACTIVE'};
fs.writeFileSync(root + '/quality-breakdown.json', JSON.stringify(report, null, 2) + '\n', {flag: 'wx', mode: 0o600});
console.log(JSON.stringify(report, null, 2));
