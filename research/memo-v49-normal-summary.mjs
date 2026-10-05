// Descriptive aggregation of the already independently reconstructed cell report.
import fs from 'node:fs';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root = 'research/memo-v49-normal';
const bytes = fs.readFileSync(root + '/readback.json'), r = JSON.parse(bytes);
assert(r.stage === 'normal' && r.sourceAndRawHashes && r.independentMetricArithmetic);
const mean = x => x.reduce((s, v) => s + v / x.length, 0);
const stationary = new Set(['aligned', 'independent', 'curved', 'baseline_matched', 'stationary_noise10']);
const summaries = {};
for (const [split, s] of Object.entries(r.summaries)) {
  assert.equal(Object.keys(s.cells).length, 84);
  const out = { maxMS: s.maximumElapsedNS / 1e6,
    workFailedCells: Object.values(s.cells).filter(v => !v.static.checks.Work).length, policies: {} };
  for (const mode of ['static', 'slow', 'round']) {
    const rows = Object.values(s.cells).map(v => v[mode]), failedChecks = {};
    for (const v of rows) for (const [key, q] of Object.entries(v.checks)) {
      if (typeof q === 'object' && !q.pass) failedChecks[key] = (failedChecks[key] ?? 0) + 1;
    }
    const pairs = Object.entries(s.pairedContrasts), shifted = pairs.filter(([key]) => !stationary.has(key.split('/')[1]));
    assert.equal(shifted.length, 54);
    out.policies[mode] = { wholeFailedCells: s.failedCells[mode],
      qualityOnlyFailedCells: rows.filter(v => Object.values(v.checks).some(q => typeof q === 'object' && !q.pass)).length,
      failedChecks, macroIssuedBrierVsFull: mean(pairs.map(([, v]) => v[mode]['full/IssuedBrier'].mean)),
      macroIssuedBrierVsAdaptive: mean(pairs.map(([, v]) => v[mode]['adaptive/IssuedBrier'].mean)),
      shiftedIssuedBrierVsFull: mean(shifted.map(([, v]) => v[mode]['full/IssuedBrier'].mean)),
      macroFinalUsefulnessVsAdaptive: mean(pairs.map(([, v]) => v[mode]['adaptive/FinalUsefulness'].mean)) };
  }
  summaries[split] = out;
}
const c = JSON.parse(fs.readFileSync(root + '/cost-report.json'));
const out = { time: new Date().toISOString(),
  readbackSHA256: crypto.createHash('sha256').update(bytes).digest('hex'), summaries,
  costScreen: { allocationBytes: c.constructorMaxBytes,
    originalMaxMS: Math.max(...c.loops.map(v => v.original.ElapsedNS)) / 1e6,
    memoMaxMS: Math.max(...c.loops.map(v => v.memo.ElapsedNS)) / 1e6 },
  macroIntervalsClaimed: false, qualityAdoption: r.qualityAdoption, allSevenWholeGoals: 'OPEN', goal: 'ACTIVE' };
fs.writeFileSync(root + '/summary.json', JSON.stringify(out, null, 2) + '\n', { flag: 'wx', mode: 0o600 });
console.log(JSON.stringify(out, null, 2));
