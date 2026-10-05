// Chained checkpoint of evidence and exact isolated sources, not a release.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync} from 'node:child_process';
const root='research/checkpoint-2026-10-04-dynvariance-v71';
const parentPath='research/checkpoint-2026-10-04-class-v69-rate-v70/manifest.json';
const usage=Number(process.argv[2]);assert(Number.isFinite(usage)&&usage>=0&&usage<=100);assert(!fs.existsSync(root));
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const json=p=>JSON.parse(fs.readFileSync(p));
assert.equal(await hash(parentPath),'da8c45a0551b895cc578b919bae78d1776b662f6dccc5d946f4ff9e17c1480a3');
const parent=json(parentPath);
for(const[p,h]of Object.entries(parent.copies))assert.equal(await hash(path.dirname(parentPath)+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.generatedCompilerCopies))assert.equal(await hash(path.dirname(parentPath)+'/'+p),h,p);
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();assert.deepEqual(tracked,Object.keys(parent.trackedHashes).sort());
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
const run='research/dynvariance-v71-diagnostic',f=json(run+'/freeze.json'),c=json(run+'/completed.json'),r=json(run+'/readback.json');
assert(c.allJobsTerminal&&c.sourceUnchanged&&c.checks.length===8&&c.checks.every(x=>x.exitCode===0));
assert(!fs.existsSync(run+'/failure.json'));
for(const[p,h]of Object.entries({...f.files,...f.protectedFiles,...f.data}))assert.equal(await hash(p),h,p);
for(const[p,h]of Object.entries(f.files))assert.equal(await hash(run+'/source/'+p),h,p);
for(const[p,h]of Object.entries(c.artifacts))assert.equal(await hash(run+'/'+p),h,p);
for(const[p,h]of Object.entries(f.generatedCompilerCopies))assert.equal(await hash(run+'/'+p),h,p);
for(const x of c.checks)assert.equal(await hash(run+'/'+x.name+'.log'),x.logSHA256,x.name);
assert.equal(r.arms,2160);assert.equal(r.modelArms,1920);assert.equal(r.independentIssuedPackets,4608000);assert.equal(r.scalarIssuedComparisons,9216000);assert.equal(r.controlsBitwiseEqual,240);assert.equal(r.populationsIdentical,40);
assert(!r.equalTotalCostSuperiorityEstablished&&!r.loadedServingEstablished);
const cost=json(run+'/cost-audit.json');assert.equal(cost.dataSHA256,await hash(run+'/diagnostic.jsonl'));assert.equal(cost.sourceSHA256,await hash('research/dynvariance-v71-cost-audit.mjs'));
for(const stage of ['initial','joint','independent','long']){
 const d='research/dynvariance-v71-'+stage,pf=json(d+'/freeze.json'),pc=json(d+'/completed.json');
 assert(pc.allJobsTerminal&&pc.allChecksPass&&pc.sourceUnchanged&&pc.checks.length===3);
 for(const[p,h]of Object.entries(pf.files))assert.equal(await hash(d+'/source/'+p.replace(/^\//,'')),h,p);
 for(const x of pc.checks){assert.equal(x.exitCode,0);assert.equal(await hash(d+'/'+x.name+'.log'),x.logSHA256)}
}
fs.mkdirSync(root,{mode:0o700});const copies={},generatedCompilerCopies={};
async function copy(src){
 if(Object.hasOwn(copies,src)){assert.equal(await hash(src),copies[src]);return}
 const out=root+'/saved/'+src;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(src,out,fs.constants.COPYFILE_EXCL|fs.constants.COPYFILE_FICLONE);fs.chmodSync(out,0o600);copies[src]=await hash(src);assert.equal(await hash(out),copies[src]);
}
async function tree(dir){for(const n of fs.readdirSync(dir)){const p=dir+'/'+n;if(fs.statSync(p).isDirectory())await tree(p);else await copy(p)}}
for(const p of Object.keys(f.files))await copy(p);
for(const p of fs.readdirSync(run).filter(p=>/\.(json|jsonl|log)$/.test(p)))await copy(run+'/'+p);
for(const[p,h]of Object.entries(f.generatedCompilerCopies)){const out='generated/'+h+'-test-main.go';fs.mkdirSync(root+'/generated',{recursive:true,mode:0o700});fs.copyFileSync(run+'/'+p,root+'/'+out,fs.constants.COPYFILE_EXCL);generatedCompilerCopies[out]=h;assert.equal(await hash(root+'/'+out),h)}
for(const stage of ['initial','joint','independent','long'])await tree('research/dynvariance-v71-'+stage);
for(const n of fs.readdirSync('research').filter(n=>/^dynvariance-v71.*\.(mjs|json|md)$/.test(n)))await copy('research/'+n);
for(const p of ['internal/researchdynvariancecheck/long_test.go','docs/experiments/mmm-dynvariance-v71-results.md','research-direction.md'])await copy(p);
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:parentPath,sha256:await hash(parentPath),copiesVerified:Object.keys(parent.copies).length},copies,generatedCompilerCopies,trackedHashes:parent.trackedHashes,goal:'ACTIVE',goals:Array(7).fill('OPEN'),weeklyUsage:{usedPercent:usage,windowDurationMins:10080},allRequiredJobsTerminal:true,completedCommands:{path:run,completedSHA256:copies[run+'/completed.json']},study:{worlds:40,arms:r.arms,modelArms:r.modelArms,independentIssuedPackets:r.independentIssuedPackets,scalarIssuedComparisons:r.scalarIssuedComparisons,controlsBitwiseEqual:r.controlsBitwiseEqual,populationsIdentical:r.populationsIdentical,totals:r.totals,gates:Object.fromEntries(Object.entries(r.summaries).map(([m,s])=>[m,s.gates])),allocation:r.allocation,equalTotalCostSuperiorityEstablished:false,loadedServingEstablished:false},preflights:['initial','joint','independent','long'].map(stage=>({stage,completedSHA256:copies['research/dynvariance-v71-'+stage+'/completed.json']})),previousTurn:'NO_RESEARCH_PROGRESS(social clarification); preceding research turn implemented a compile-only isolated prototype',thisTurn:'PROGRESS(model and independent-reference implementation, full independent-member control, unit/joint/delayed/full-journal audits, frozen complete2160-arm experiment and1920-arm independent replay, cost and immutable-control readback)',archiveIsChainedNotStandalone:true,productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,privateOrSealedLabelsOpened:false,populationConfirmationDispatched:false,allSevenEmpiricalGoalsValidated:false,scientificAdoption:false,residualLimitations:['consumed cohort/n1 cells are not untouched confirmation','static family learning is not changepoint-adaptive hyperparameter learning','coherent fitted joint law is model-conditional, not authenticated truth or Anti-Pigeon authority','equal request counts are not equal TOTAL cost','serial microbenchmarks and constructor allocations are not loaded serving, durability or freshness','valid useful splits and untouched agent outcomes remain required'],next:'See the source-backed next lead in the V71 results; preserve all seven whole criteria, failures and publication/production boundaries.'};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(copies))assert.equal(await hash(root+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await hash(root+'/manifest.json'),copies:Object.keys(copies).length,allJobsTerminal:true,goal:'ACTIVE',wholeGoals:'seven OPEN',weeklyUsage:usage},null,2));
