// Chained reproducible checkpoint, including failures and unchanged raw data.
// No commit/push, publication, production/private access or goal closure.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync} from 'node:child_process';
const root='research/checkpoint-2026-10-04-retention-v63-v65',previous='research/checkpoint-2026-10-04-retention-v62/manifest.json',usage=Number(process.argv[2]);assert(!fs.existsSync(root));assert(Number.isFinite(usage)&&usage>=0&&usage<=100);
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const json=p=>JSON.parse(fs.readFileSync(p));assert.equal(await hash(previous),'1ab64205ef92084495756178e1b215caa2528fc1ecbd9d6e80d3cff874ecae65');const parent=json(previous);
for(const[p,h]of Object.entries(parent.copies))assert.equal(await hash(path.dirname(previous)+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.generatedCompilerCopies))assert.equal(await hash(path.dirname(previous)+'/'+p),h,p);
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();assert.deepEqual(tracked,Object.keys(parent.trackedHashes).sort());assert.equal(tracked.length,14);
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
const good=['research/retention-v63-integration','research/retention-v64c-audit','research/retention-v65-coherence'];
const bad=['research/retention-v64-diagnostic','research/retention-v64b-diagnostic'];
for(const run of good) {
 const f=json(run+'/freeze.json'),c=json(run+'/completed.json');assert(c.allJobsTerminal);
 if(c.checks)assert(c.checks.every(x=>x.exitCode===0));else assert.equal(c.exitCode,0);
 for(const[p,h]of Object.entries({...f.files,...f.protectedFiles}))assert.equal(await hash(p),h,p);
 for(const[p,h]of Object.entries(f.files))assert.equal(await hash(run+'/source/'+p),h,p);
 for(const[p,h]of Object.entries(f.data??{}))assert.equal(await hash(p),h,p);
 for(const[p,h]of Object.entries(c.artifacts??{}))assert.equal(await hash(run+'/'+p),h,p);
 for(const[p,h]of Object.entries(f.generatedCompilerCopies??{}))assert.equal(await hash(run+'/'+p),h,p);
 for(const x of c.checks??[])assert.equal(await hash(run+'/'+x.name+'.log'),x.logSHA256);
}
for(const run of bad){const f=json(run+'/freeze.json'),c=json(run+'/failure.json');assert(c.checks.at(-1).exitCode!==0);for(const[p,h]of Object.entries(f.files))assert.equal(await hash(run+'/source/'+p),h,p);for(const x of c.checks)assert.equal(await hash(run+'/'+x.name+'.log'),x.logSHA256)}
const collected=json('research/retention-v64b-diagnostic/experiment-command.json');assert.equal(collected.exitCode,0);
const report=json('research/retention-v64c-audit/readback.json');assert.equal(report.arms,600);assert.equal(report.bankArms,360);assert.equal(report.controlsBitwiseEqual,240);assert.equal(report.populationsIdentical,40);
assert.equal(json('research/retention-v64c-audit/readback-command.json').exitCode,0);
const allocation=json('research/retention-v63-integration/allocation.json');assert(allocation.AllocatedBytes.shared150<allocation.CapBytes&&allocation.AllocatedBytes.shared200<allocation.CapBytes);
for(const m of ['no_pair','random','uncertainty']){assert(report.summaries[m].gates.completeLoop400);assert(!report.summaries[m].gates.gain01);assert(!report.summaries[m].gates.noAdaptiveHarm01);assert(!report.summaries[m].gates.recovery)}
assert.equal(json('research/retention-v65-coherence/completed.json').towerIdentityConfirmed,false);
fs.mkdirSync(root,{mode:0o700});const copies={},generatedCompilerCopies={};
async function copy(src){if(Object.hasOwn(copies,src)){assert.equal(await hash(src),copies[src]);return}const dest=root+'/saved/'+src;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.copyFileSync(src,dest,fs.constants.COPYFILE_EXCL|fs.constants.COPYFILE_FICLONE);fs.chmodSync(dest,0o600);copies[src]=await hash(src);assert.equal(await hash(dest),copies[src])}
for(const run of [...good,...bad]) {
 const f=json(run+'/freeze.json');
 for(const p of Object.keys(f.files)){if(good.includes(run))await copy(p);else await copy(run+'/source/'+p)}
 for(const p of fs.readdirSync(run).filter(p=>/\.(json|jsonl|log|mjs)$/.test(p)))await copy(run+'/'+p);
 for(const[p,h]of Object.entries(f.generatedCompilerCopies??{})){const destName='generated/'+h+'-test-main.go';if(generatedCompilerCopies[destName])continue;const dest=root+'/'+destName;fs.mkdirSync(path.dirname(dest),{recursive:true,mode:0o700});fs.copyFileSync(run+'/'+p,dest,fs.constants.COPYFILE_EXCL);assert.equal(await hash(dest),h);generatedCompilerCopies[destName]=h}
}
for(const dir of ['internal/researchwindowjournal','internal/researchwindowjournalref','internal/researchretentioncoherence'])for(const n of fs.readdirSync(dir).filter(n=>n.endsWith('.go')))await copy(dir+'/'+n);
for(const p of fs.readdirSync('research').filter(p=>/^retention-v(63|64|65|66).*\.(mjs|json|log|go|md)$/.test(p)))await copy('research/'+p);
for(const p of ['internal/researchdispersion/retention_v64_test.go','internal/researchdispersion/retention_v64_tie_audit_test.go','internal/researchdispersion/retention_v64_tie_corruptions_test.go',
 'docs/experiments/mmm-retention-v63-results.md','docs/experiments/mmm-retention-v64-results.md','docs/experiments/mmm-retention-v65-coherence-results.md','research-direction.md'])await copy(p);
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:previous,sha256:await hash(previous),copiesVerified:Object.keys(parent.copies).length},copies,generatedCompilerCopies,trackedHashes:parent.trackedHashes,
 goal:'ACTIVE',goals:Array(7).fill('OPEN'),weeklyUsage:{usedPercent:usage,windowDurationMins:10080},allRequiredJobsTerminal:true,
 completedCommands:good.map(p=>({path:p,completedSHA256:copies[p+'/completed.json']})),preservedFailedCommands:bad.map(p=>({path:p,failureSHA256:copies[p+'/failure.json']})),
 constructorAllocation:allocation,full200MemberMemoryPass:true,componentRaceRoots:26,bitwiseV62PublicComparisons:26394,independentBankForecasts:25758,
 completeFiveArmDiagnostic:{worlds:40,arms:600,bankArms:360,issuedPacketsReconstructed:864000,cleanAndObservedScalarComparisons:1728000,controlsBitwiseEqual:240,populationsIdentical:40,scienceGates:'FAIL',coreRuntimeGate:'PASS',ordering:report.ordering},
 collectedDataPath:'research/retention-v64b-diagnostic/diagnostic.jsonl',collectedDataSHA256:copies['research/retention-v64b-diagnostic/diagnostic.jsonl'],originalDataUnchanged:true,
 mixtureTowerCounterexample:json('research/retention-v65-coherence/completed.json'),allSevenEmpiricalGoalsValidated:false,scientificAdoption:false,
 previousGoalTurn:'PROGRESS(V62); intervening social clarification NO_RESEARCH_PROGRESS',thisGoalTurn:'PROGRESS(journal memory/ownership rescue,600-arm controlled quality,independent replay,numerical repair with failures retained,tower counterexample,research-grounded next lead)',
 productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,privateOrSealedLabelsOpened:false,populationConfirmationDispatched:false,archiveIsChainedNotStandalone:true,
 residualLimitations:['V65 small diagnostic did not separately archive generated test main; V63/V64 main copies preserved','race success does not establish concurrent API','no loaded serving/persistence/freshness claim','consumed worlds not fresh confirmation'],
 next:'V66 explicitly joint/calibrated observation-value and regime/prior ablations, then full seven-goal evidence including useful certified splits,untouched labeled agent tasks and loaded durable freshness'};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(copies))assert.equal(await hash(root+'/saved/'+p),h,p);for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await hash(root+'/manifest.json'),copies:Object.keys(copies).length,allJobsTerminal:true,goal:'ACTIVE',wholeGoals:'seven OPEN',weeklyUsage:usage},null,2));
