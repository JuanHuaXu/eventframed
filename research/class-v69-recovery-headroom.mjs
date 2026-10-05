// Post-collection oracle feasibility only; truth is never a learner input.
import fs from 'node:fs';
import readline from 'node:readline';
import crypto from 'node:crypto';
import assert from 'node:assert/strict';
const root = 'research/class-v69-diagnostic';
const source = 'research/class-v69-recovery-headroom.mjs';
async function hash(p) {
  const h = crypto.createHash('sha256');
  for await (const b of fs.createReadStream(p)) h.update(b);
  return h.digest('hex');
}
const done = JSON.parse(fs.readFileSync(root + '/completed.json'));
assert(done.allJobsTerminal && done.checks.every(x => x.exitCode === 0));
const sourceHash = await hash(source);
assert.equal(await hash(root + '/diagnostic.jsonl'), done.artifacts['diagnostic.jsonl']);
const worlds = [];
let regimes = 0, phases = 0, unreachablePhases = 0;
for await (const line of readline.createInterface({input: fs.createReadStream(root + '/diagnostic.jsonl'), crlfDelay: Infinity})) {
  const row = JSON.parse(line);
  if (!row.Arms) continue;
  const w = row.Population.World;
  const rounds = w.Rates.map((rates, index) => {
    assert.equal(rates.length, 150);
    assert(rates.every(p => Number.isFinite(p) && p >= 0 && p <= 1));
    const floor = rates.reduce((s, p) => s + p * (1 - p), 0) / 150;
    const top = [...rates].sort((a, b) => b - a).slice(0, 10);
    const maxUtility = top.reduce((s, p) => s + p, 0) / 10;
    // An independently sorted/order-free threshold check guards the packet bound.
    const cutoff = top.at(-1);
    const greater = rates.filter(p => p > cutoff);
    const bound = (greater.reduce((s, p) => s + p, 0) + (10 - greater.length) * cutoff) / 10;
    assert(Math.abs(bound - maxUtility) < 2e-12);
    return {index, floor, maxUtility, attainable: floor <= .2 && maxUtility >= .75};
  });
  const oraclePhases = [];
  for (let j = 0; j < (w.Changes ?? []).length; j++) {
    const start = w.Changes[j], end = w.Changes[j + 1] ?? 16;
    let streak = 0, delay = end - start + 1, reached = false;
    for (let n = start; n < end; n++) {
      streak = rounds[n].attainable ? streak + 1 : 0;
      if (streak === 2) {delay = n - start + 1; reached = true; break;}
    }
    phases++;
    if (!reached) unreachablePhases++;
    oraclePhases.push({start, end, delay, reached, riskImpossibleRounds: rounds.slice(start, end).filter(x => x.floor > .2).length, utilityImpossibleRounds: rounds.slice(start, end).filter(x => x.maxUtility < .75).length});
  }
  const oracleRecovery = oraclePhases.length ? oraclePhases.reduce((s, p) => s + p.delay, 0) / oraclePhases.length : 0;
  if (oraclePhases.length) regimes++;
  const oracleIssuedRisk = rounds.reduce((s, x) => s + x.floor, 0) / 16;
  const controls = row.Arms.filter(a => a.Mode === 'full' || a.Mode === 'adaptive').map(a => {
    assert(a.IssuedBrier + 2e-12 >= oracleIssuedRisk);
    assert(a.Recovery + 2e-12 >= oracleRecovery);
    return {mode: a.Mode, schedule: a.Schedule, issuedRisk: a.IssuedBrier, maximumRiskGain: a.IssuedBrier - oracleIssuedRisk, recovery: a.Recovery, maximumRecoveryGain: a.Recovery - oracleRecovery};
  });
  assert.equal(controls.length, 6);
  worlds.push({geometry: w.Geometry, regime: w.Regime, oracleIssuedRisk, oracleRecovery, rounds, oraclePhases, controls});
}
assert.equal(worlds.length, 40);
assert.equal(await hash(source), sourceHash);
assert.equal(await hash(root + '/diagnostic.jsonl'), done.artifacts['diagnostic.jsonl']);
const mean = xs => xs.reduce((s, x) => s + x, 0) / xs.length;
const result = {
  source, sourceSHA256: sourceHash, dataSHA256: done.artifacts['diagnostic.jsonl'], scope: 'postcollection consumed-cohort evaluator-only perfect-forecast feasibility, not a deployable policy or a changed gate',
  worlds: worlds.length, shiftedWorlds: regimes, phases, unreachablePhases,
  oracleMeanIssuedRisk: mean(worlds.map(x => x.oracleIssuedRisk)),
  meanMaximumRiskGainFull: mean(worlds.flatMap(x => x.controls.filter(c => c.mode === 'full').map(c => c.maximumRiskGain))),
  meanMaximumRecoveryGainAdaptiveShifted: mean(worlds.filter(x => x.oraclePhases.length).flatMap(x => x.controls.filter(c => c.mode === 'adaptive').map(c => c.maximumRecoveryGain))),
  details: worlds, originalFailedVerdictsChanged: false, externalPopulationImpossibilityEstablished: false, futureLabelsSuppliedToLearner: false, goal: 'ACTIVE', goals: Array(7).fill('OPEN'),
};
fs.writeFileSync(root + '/recovery-headroom.json', JSON.stringify(result, null, 2) + '\n', {flag: 'wx', mode: 0o600});
console.log(JSON.stringify({...result, details: undefined}, null, 2));
