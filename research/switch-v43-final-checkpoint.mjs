// Evidence-only ledger. Exclusive outputs and exact documentation copies.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';
const root='research/switch-v43-final-checkpoint';assert(!fs.existsSync(root));
const usage=Number(process.env.EVENTFRAME_WEEKLY_USAGE);assert(Number.isFinite(usage)&&usage>=0&&usage<=80);
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const study='research/switch-v43-study-diagnostic',done=JSON.parse(fs.readFileSync(study+'/completed.json')),frozen=JSON.parse(fs.readFileSync(study+'/freeze.json'));
assert(done.sourceUnchanged&&done.checks.length===6&&done.checks.every(c=>c.code===0));
for(const[p,h]of Object.entries(frozen.files))assert.equal(hash(fs.readFileSync(p)),h);
const docs=['research-direction.md','docs/experiments/mmm-moment-v41-results.md','docs/experiments/mmm-switch-v43-cost-results.md','docs/experiments/mmm-switch-v43-diagnostic-results.md','docs/experiments/research-checkpoint-2026-10-03-switch.md','research/switch-v43-policy-verdict.mjs','research/switch-v43-final-checkpoint.mjs'];
fs.mkdirSync(root,{mode:0o700});
function write(p,b){fs.mkdirSync(path.dirname(p),{recursive:true,mode:0o700});const f=fs.openSync(p,'wx',0o600);try{fs.writeFileSync(f,b);fs.fsyncSync(f)}finally{fs.closeSync(f)}}
const docHashes={};for(const p of docs){const b=fs.readFileSync(p);docHashes[p]=hash(b);write(root+'/source/'+p,b);assert.equal(hash(fs.readFileSync(root+'/source/'+p)),docHashes[p])}
const evidencePaths=['research/moment-v41-normal/completed.json','research/moment-v41-normal/readback.json','research/moment-v41-normal/oracle-ceilings.json',...['initial','cached-constants','log-only','unit-spans'].map(s=>'research/switch-v43-cost-'+s+'/cost-report.json'),study+'/completed.json',study+'/readback.json',study+'/policy-verdict.json',study+'/production-dependencies.json','research/moment-switch-checkpoint-span-study-diagnostic/manifest.json'];
const evidenceHashes={};for(const p of evidencePaths){const b=fs.readFileSync(p);evidenceHashes[p]=hash(b);write(root+'/evidence/'+p,b)}
write(root+'/manifest.json',Buffer.from(JSON.stringify({time:new Date().toISOString(),docHashes,evidenceHashes,learnerCheckpoint:'research/moment-switch-checkpoint-span-study-diagnostic',sources:77,previousGoalTurn:'PROGRESS',thisGoalTurn:'PROGRESS',requiredProcessesTerminal:true,normalStudyRun:false,weeklyUsage:usage,allSevenWholeGoals:'OPEN',goal:'ACTIVE',productionChangedByThisResearch:false,whitepaperChanged:false,commitOrPush:false,next:'SAME frozen normal design/confirmation plus independent fixed-policy verdict; no diagnostic tuning.'},null,2)+'\n'));
console.log(JSON.stringify({root,sourceCopies:docs.length,evidenceCopies:evidencePaths.length,learnerSources:77,allSevenWholeGoals:'OPEN',goal:'ACTIVE'}));
