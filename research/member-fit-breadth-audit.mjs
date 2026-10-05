import fs from 'node:fs';
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import {execFileSync} from 'node:child_process';

const [raw, summary] = process.argv.slice(2);
// Keep the separate metric/CDF implementation; add independent-fit provenance.
const metricAudit = JSON.parse(execFileSync(process.execPath,
  ['research/member-coverage-audit.mjs', raw, summary], {encoding:'utf8'}));
const [header, ...rows] = fs.readFileSync(raw,'utf8').trim().split('\n').map(JSON.parse);
assert.equal(header.Version,'member-fit-breadth-v1');
assert.deepEqual(header.Seeds,[2026091511,2026091512]);
const fits = new Set(), digests = new Set(), allSeeds = new Set();
for (const r of rows) {
  const phase = ['design','confirmation'].indexOf(r.Split);
  const scenario = ['stable','member_shift','common_shift','recurring','null'].indexOf(r.Scenario);
  assert(phase >= 0 && scenario >= 0 && Number.isInteger(r.Index) && r.Index >= 0 && r.Index < 512);
  const prefix = BigInt(header.Seeds[phase])*1000000n + BigInt(scenario)*100000n + BigInt(r.Index)*100n;
  assert.equal(BigInt(r.FitSeed),prefix+4n);
  assert(/^[a-f0-9]{64}$/.test(r.FitHash));
  fits.add(r.FitSeed); digests.add(r.FitHash);
  for (let role=0n; role<5n; role++) {
    const seed=String(prefix+role);
    assert(!allSeeds.has(seed)); allSeeds.add(seed);
  }
}
assert.equal(fits.size,5120); assert.equal(digests.size,5120); assert.equal(allSeeds.size,25600);
console.log(JSON.stringify({...metricAudit, freshFits:fits.size, uniqueFitDigests:digests.size,
  distinctRoleSeeds:allSeeds.size,
  fitAuditSHA256:crypto.createHash('sha256').update(fs.readFileSync(import.meta.filename)).digest('hex'),
  fitScope:'Independent seed arithmetic and uniqueness; sample generation and model parity checked by Go contracts/full replay'}));
