// Preserve failed run/source evidence before the local test-harness repair.
import fs from 'node:fs';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/retention-v64-diagnostic',hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const freeze=JSON.parse(fs.readFileSync(root+'/freeze.json')),fail=JSON.parse(fs.readFileSync(root+'/failure.json'));assert.equal(fail.checks[0].exitCode,1);
for(const[p,h]of Object.entries(freeze.files))assert.equal(hash(fs.readFileSync(root+'/source/'+p)),h,p);
const src='internal/researchdispersion/retention_v64_test.go';assert.equal(hash(fs.readFileSync(src)),freeze.files[src]);
const record={kind:'local harness defect, not scientific model failure',symptom:'TestRetentionV64FutureAndCorruptions fails metrics on unscored run output',cause:'runRetentionV64 returns raw arm; TestExperiment scores it, preflight omitted that call',failedFixtureSHA256:freeze.files[src],failedFreezeSHA256:hash(fs.readFileSync(root+'/freeze.json')),failedLogSHA256:hash(fs.readFileSync(root+'/race-fixture.log')),repair:'call unchanged scoreWindowV37 before independent metric validation; no model/policy/gate/data modification',commands:fail.checks,wholeGoals:'OPEN',productionChanged:false};
fs.writeFileSync('research/retention-v64-preflight-failure.json',JSON.stringify(record,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[name,transform]of [['run',s=>s.replaceAll('retention-v64-diagnostic','retention-v64b-diagnostic').replaceAll('retention-v64-run.mjs','retention-v64b-run.mjs').replaceAll('retention-v64-readback.mjs','retention-v64b-readback.mjs')],['readback',s=>s.replaceAll('retention-v64-diagnostic','retention-v64b-diagnostic')]]) {
 const src='research/retention-v64-'+name+'.mjs',dest='research/retention-v64b-'+name+'.mjs';fs.writeFileSync(dest,transform(fs.readFileSync(src,'utf8')),{flag:'wx',mode:0o600});
}
console.log(JSON.stringify(record,null,2));
