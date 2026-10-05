// Chained scientific checkpoint; no git mutation or production access.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync} from 'node:child_process';
const root='research/checkpoint-2026-10-04-shared-v68-class-v69',previous='research/checkpoint-2026-10-04-joint-v66-v68/manifest.json',usage=Number(process.argv[2]);assert(!fs.existsSync(root));assert(Number.isFinite(usage)&&usage>=0&&usage<=100);
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const json=p=>JSON.parse(fs.readFileSync(p));assert.equal(await hash(previous),'aaa2f3b819d752aab0c96ca71d9d3a7916ea0ea36785f63060a4a7681bcf96e0');const parent=json(previous);
for(const[p,h]of Object.entries(parent.copies))assert.equal(await hash(path.dirname(previous)+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.generatedCompilerCopies))assert.equal(await hash(path.dirname(previous)+'/'+p),h,p);
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();assert.deepEqual(tracked,Object.keys(parent.trackedHashes).sort());assert.equal(tracked.length,14);
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
const run='research/shared-v68-diagnostic',f=json(run+'/freeze.json'),c=json(run+'/completed.json'),r=json(run+'/readback.json'),toy=json('research/class-v69-preflight/results.json');
assert(c.allJobsTerminal&&c.checks.length===8&&c.checks.every(x=>x.exitCode===0));
for(const[p,h]of Object.entries({...f.files,...f.protectedFiles,...f.data}))assert.equal(await hash(p),h,p);
for(const[p,h]of Object.entries(f.files))assert.equal(await hash(run+'/source/'+p),h,p);
for(const[p,h]of Object.entries(c.artifacts))assert.equal(await hash(run+'/'+p),h,p);
for(const[p,h]of Object.entries(f.generatedCompilerCopies))assert.equal(await hash(run+'/'+p),h,p);
for(const x of c.checks)assert.equal(await hash(run+'/'+x.name+'.log'),x.logSHA256,x.name);
assert.equal(r.arms,1680);assert.equal(r.modelArms,1440);assert.equal(r.independentIssuedPackets,3456000);assert.equal(r.scalarIssuedComparisons,6912000);assert.equal(r.controlsBitwiseEqual,240);assert.equal(r.populationsIdentical,40);assert(r.allocationGate);
const failed=json('research/shared-v68-boundary-preflight/command.json');assert.equal(failed.exitCode,1);assert(failed.allJobsTerminal);
assert.equal(await hash('research/shared-v68-boundary-preflight/command.log'),failed.logSHA256);
for(const[p,h]of Object.entries(failed.hashes))assert.equal(await hash('research/shared-v68-boundary-preflight/source/'+p),h,p);
assert.equal(toy.checks,1289);assert(toy.latentOrderingNotRefinementInvariant&&toy.classGainRefinementInvariant&&!toy.qualityRescueEstablished&&!toy.equalTotalCostEstablished);
const tf=json('research/class-v69-preflight/freeze.json');assert.equal(await hash(tf.source),tf.sourceSHA256);assert.equal(await hash('research/class-v69-preflight/source.mjs'),tf.sourceSHA256);const tc=json('research/class-v69-preflight/completed.json');assert.equal(tc.exitCode,0);assert.equal(await hash('research/class-v69-preflight/results.json'),tc.resultsSHA256);
assert.equal(json(run+'/cost-audit.json').dataSHA256,await hash(run+'/diagnostic.jsonl'));
fs.mkdirSync(root,{mode:0o700});const copies={},generatedCompilerCopies={};
async function copy(src){if(Object.hasOwn(copies,src)){assert.equal(await hash(src),copies[src]);return}const out=root+'/saved/'+src;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(src,out,fs.constants.COPYFILE_EXCL|fs.constants.COPYFILE_FICLONE);fs.chmodSync(out,0o600);copies[src]=await hash(src);assert.equal(await hash(out),copies[src])}
async function tree(dir){for(const n of fs.readdirSync(dir)){const p=dir+'/'+n;if(fs.statSync(p).isDirectory())await tree(p);else await copy(p)}}
for(const p of Object.keys(f.files))await copy(p);
for(const p of fs.readdirSync(run).filter(p=>/\.(json|jsonl|log)$/.test(p)))await copy(run+'/'+p);
for(const[p,h]of Object.entries(f.generatedCompilerCopies)){const out='generated/'+h+'-test-main.go';fs.mkdirSync(root+'/generated',{recursive:true,mode:0o700});fs.copyFileSync(run+'/'+p,root+'/'+out,fs.constants.COPYFILE_EXCL);assert.equal(await hash(root+'/'+out),h);generatedCompilerCopies[out]=h}
await tree('research/shared-v68-boundary-preflight');await tree('research/class-v69-preflight');
for(const n of fs.readdirSync('research').filter(n=>/^(shared-v68|class-v69).*\.(mjs|json|md)$/.test(n)))await copy('research/'+n);
for(const p of ['docs/experiments/mmm-shared-v68-results.md','research-direction.md'])await copy(p);
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:previous,sha256:await hash(previous),copiesVerified:Object.keys(parent.copies).length},copies,generatedCompilerCopies,trackedHashes:parent.trackedHashes,
goal:'ACTIVE',goals:Array(7).fill('OPEN'),weeklyUsage:{usedPercent:usage,windowDurationMins:10080},allRequiredJobsTerminal:true,completedCommands:{path:run,completedSHA256:copies[run+'/completed.json']},
sharedStudy:{worlds:40,arms:1680,modelArms:1440,independentIssuedPackets:3456000,scalarIssuedComparisons:6912000,controlsBitwiseEqual:240,populationsIdentical:40,allocation:r.allocation,gates:Object.fromEntries(Object.entries(r.summaries).map(([m,s])=>[m,s.gates])),equalTotalCostSuperiorityEstablished:false,loadedServingEstablished:false},
boundaryFailurePreserved:{path:'research/shared-v68-boundary-preflight/command.json',exitCode:1,repair:'retain conditional field laws per noise hypothesis with log evidence; static field rebuild in logs; no model,prior,cap or gate relaxation'},
classPreflight:{checks:toy.checks,maxDefect:toy.maxDefect,counterexample:toy.counterexample,mathematicalOnly:true,notIntegratedOrEmpiricallyValidated:true},
previousGoalTurn:'PROGRESS(V66/V67 completed science,cost,shared toy and V68 model components); intervening social clarification NO_RESEARCH_PROGRESS',thisGoalTurn:'PROGRESS(finished frozen shared/local/noise fixture,retained numerical boundary repair,full1680-arm experiment and independent replay,cost accounting and class-objective counterexample)',
productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,privateOrSealedLabelsOpened:false,populationConfirmationDispatched:false,allSevenEmpiricalGoalsValidated:false,scientificAdoption:false,archiveIsChainedNotStandalone:true,
residualLimitations:['consumed worlds/n1 per cell are not fresh confirmation','mixed sharing is model evidence,not Anti-Pigeon authority','core elapsed is not loaded serving/persistence/freshness','finite observation values do not prove external utility','class concentration does not inherit EC2 competitiveness','goals3/5/6 broader required work remains'],
next:'class/decision-focused joint observation with independent nuisance controls and equal TOTALcost testing; retain whole goals including valid useful splits,untouched agent outcomes and durable loaded freshness'};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(copies))assert.equal(await hash(root+'/saved/'+p),h,p);for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await hash(root+'/manifest.json'),copies:Object.keys(copies).length,allJobsTerminal:true,goal:'ACTIVE',wholeGoals:'seven OPEN',weeklyUsage:usage},null,2));
