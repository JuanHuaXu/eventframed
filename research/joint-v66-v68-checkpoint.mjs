// Chained checkpoint. Preserve completed negative science and new finite toy.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync} from 'node:child_process';
const root='research/checkpoint-2026-10-04-joint-v66-v68',previous='research/checkpoint-2026-10-04-retention-v63-v65/manifest.json',usage=Number(process.argv[2]);assert(!fs.existsSync(root));assert(Number.isFinite(usage)&&usage>=0&&usage<=100);
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const json=p=>JSON.parse(fs.readFileSync(p));assert.equal(await hash(previous),'8b26825dd2a90f5fc2ca50b3f4f50d9969c7ddb5b141c6ffc3091e8ff1e3a932');const parent=json(previous);
for(const[p,h]of Object.entries(parent.copies))assert.equal(await hash(path.dirname(previous)+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.generatedCompilerCopies))assert.equal(await hash(path.dirname(previous)+'/'+p),h,p);
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();assert.deepEqual(tracked,Object.keys(parent.trackedHashes).sort());assert.equal(tracked.length,14);
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
const runs=['research/joint-v66-diagnostic','research/joint-rate-v67-diagnostic'];
for(const run of runs){
 const f=json(run+'/freeze.json'),c=json(run+'/completed.json');assert(c.allJobsTerminal&&c.checks.every(x=>x.exitCode===0));
 for(const[p,h]of Object.entries({...f.files,...f.protectedFiles,...f.data}))assert.equal(await hash(p),h,p);
 for(const[p,h]of Object.entries(f.files))assert.equal(await hash(run+'/source/'+p),h,p);
 for(const[p,h]of Object.entries(c.artifacts))assert.equal(await hash(run+'/'+p),h,p);
 for(const[p,h]of Object.entries(f.generatedCompilerCopies))assert.equal(await hash(run+'/'+p),h,p);
 for(const x of c.checks)assert.equal(await hash(run+'/'+x.name+'.log'),x.logSHA256,x.name);
}
const a=json(runs[0]+'/readback.json'),b=json(runs[1]+'/readback.json'),toy=json('research/joint-shared-v68-preflight/results.json');
assert.equal(a.arms,960);assert.equal(a.modelArms,720);assert.equal(a.controlsBitwiseEqual,240);assert.equal(a.populationsIdentical,40);assert(a.allocationGate);
assert.equal(b.arms,720);assert.equal(b.nominalBitwiseEqual,240);assert.equal(b.populationsIdentical,40);
for(const r of [a,b])for(const x of Object.values(r.summaries)){assert(x.gates.completeLoop400);assert(!x.gates.gain01&&!x.gates.noAdaptiveHarm01&&!x.gates.recovery)}
assert.equal(toy.checks,5285);assert.equal(toy.queries,64);assert(toy.maxNonlocalValue>1e-5);assert(toy.towerIdentityPassed&&toy.baselineMeanPreserved&&toy.efficientVersusEnumerationPassed);
assert.equal(json('research/joint-shared-v68-preflight/completed.json').exitCode,0);
const toyFreeze=json('research/joint-shared-v68-preflight/freeze.json');assert.equal(await hash(toyFreeze.source),toyFreeze.sourceSHA256);assert.equal(await hash('research/joint-shared-v68-preflight/source.mjs'),toyFreeze.sourceSHA256);
fs.mkdirSync(root,{mode:0o700});const copies={},generatedCompilerCopies={};
async function copy(src){if(Object.hasOwn(copies,src)){assert.equal(await hash(src),copies[src]);return}const out=root+'/saved/'+src;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(src,out,fs.constants.COPYFILE_EXCL|fs.constants.COPYFILE_FICLONE);fs.chmodSync(out,0o600);copies[src]=await hash(src);assert.equal(await hash(out),copies[src])}
for(const run of runs){const f=json(run+'/freeze.json');for(const p of Object.keys(f.files))await copy(p);for(const p of fs.readdirSync(run).filter(p=>/\.(json|jsonl|log)$/.test(p)))await copy(run+'/'+p);
 for(const[p,h]of Object.entries(f.generatedCompilerCopies)){const out='generated/'+h+'-test-main.go';if(generatedCompilerCopies[out])continue;fs.mkdirSync(root+'/generated',{recursive:true,mode:0o700});fs.copyFileSync(run+'/'+p,root+'/'+out,fs.constants.COPYFILE_EXCL);assert.equal(await hash(root+'/'+out),h);generatedCompilerCopies[out]=h}
}
for(const dir of ['research/joint-v66-preflight','research/joint-shared-v68-preflight'])for(const n of fs.readdirSync(dir))if(fs.statSync(dir+'/'+n).isFile())await copy(dir+'/'+n);
for(const n of fs.readdirSync('research').filter(n=>/^joint-(v66|rate-v67|shared-v68).*\.(mjs|json|md)$/.test(n)))await copy('research/'+n);
for(const p of ['docs/experiments/mmm-joint-v66-results.md','docs/experiments/mmm-joint-rate-v67-results.md','research-direction.md'])await copy(p);
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:previous,sha256:await hash(previous),copiesVerified:Object.keys(parent.copies).length},copies,generatedCompilerCopies,trackedHashes:parent.trackedHashes,
 goal:'ACTIVE',goals:Array(7).fill('OPEN'),weeklyUsage:{usedPercent:usage,windowDurationMins:10080},allRequiredJobsTerminal:true,
 completedCommands:runs.map(p=>({path:p,completedSHA256:copies[p+'/completed.json']})),
 jointStudy:{worlds:40,arms:960,modelArms:720,issuedPacketsReconstructed:1728000,scalarIssuedComparisons:3456000,controlsBitwiseEqual:240,populationsIdentical:40,scientificGates:'FAIL',coreRuntimeGate:'PASS',allocation:a.allocation},
 rateBoundaryStudy:{worlds:40,arms:720,issuedPacketsReconstructed:1728000,scalarIssuedComparisons:3456000,nominalBitwiseEqual:240,populationsIdentical:40,scientificGates:'FAIL',coreRuntimeGate:'PASS',staticUncertaintyRisk:b.totals['static/uncertainty'].risk,nominalUncertaintyRisk:b.totals['nominal/uncertainty'].risk,memorylessUncertaintyRisk:b.totals['memoryless/uncertainty'].risk},
 sharedContextToy:{checks:toy.checks,queries:toy.queries,states:toy.latentStates,maxDefect:toy.maxDifference,maxNonlocalValue:toy.maxNonlocalValue,mathematicalPreflightOnly:true,notIntegratedOrEmpiricallyValidated:true},
 preflightFailurePreserved:{path:'research/joint-v66-preflight/failure.json',sourcePath:'research/joint-v66-preflight/reference-test.go.txt',repair:'final delayed delivery pass only; no model/reference/gate change'},
 previousGoalTurn:'PROGRESS(V63/V64/V65); intervening social clarification NO_RESEARCH_PROGRESS',thisGoalTurn:'PROGRESS(coherent joint implementation,complete960-arm quality,cost and independent replay,720-arm controlled dynamics ablation,nonlocal shared-context preflight)',
 productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,privateOrSealedLabelsOpened:false,populationConfirmationDispatched:false,allSevenEmpiricalGoalsValidated:false,scientificAdoption:false,archiveIsChainedNotStandalone:true,
 residualLimitations:['consumed worlds/n1 per cell are not fresh confirmation','race success does not establish concurrent API','64-trial member cap is not unlimited continuous serving','no loaded serving/persistence/freshness result','shared-context toy not a safe-merge certificate or online controller','shared-noise/context pooling,rate priors and source-dependence remain model alternatives,not uniquely identified causes'],
 next:'V68 coherent shared-context/independent alternatives with nonlocal acquisition cost,full dynamic independent replay and controlled replication; retain full seven-goal evidence requirements including useful certified splits,untouched tasks and durable loaded freshness'};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(copies))assert.equal(await hash(root+'/saved/'+p),h,p);for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await hash(root+'/manifest.json'),copies:Object.keys(copies).length,allJobsTerminal:true,goal:'ACTIVE',wholeGoals:'seven OPEN',weeklyUsage:usage},null,2));
