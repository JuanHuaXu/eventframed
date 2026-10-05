// Reproducible chained research checkpoint, not a release or completion claim.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync} from 'node:child_process';
const root='research/checkpoint-2026-10-04-dynvarcache-v72-mean-v73',parentPath='research/checkpoint-2026-10-04-dynvariance-v71/manifest.json';
const usage=Number(process.argv[2]);assert(Number.isFinite(usage)&&usage>=0&&usage<=100);assert(!fs.existsSync(root));
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const json=p=>JSON.parse(fs.readFileSync(p));
assert.equal(await hash(parentPath),'9d6cb6dcb7fc8281b407ebc9cb5035c27a2e8b29bcf9230013517e2da9a2b297');const parent=json(parentPath);
for(const[p,h]of Object.entries(parent.copies))assert.equal(await hash(path.dirname(parentPath)+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.generatedCompilerCopies))assert.equal(await hash(path.dirname(parentPath)+'/'+p),h,p);
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();assert.deepEqual(tracked,Object.keys(parent.trackedHashes).sort());
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
const run='research/dynvarcache-v72-diagnostic',f=json(run+'/freeze.json'),c=json(run+'/completed.json'),r=json(run+'/readback.json'),a=json(run+'/transitive-audit.json'),cost=json(run+'/cost-audit.json');
assert(c.allJobsTerminal&&c.sourceUnchanged&&c.checks.length===8&&c.checks.every(x=>x.exitCode===0));assert(!fs.existsSync(run+'/failure.json'));
for(const[p,h]of Object.entries({...f.files,...f.protectedFiles,...f.data}))assert.equal(await hash(p),h,p);
for(const[p,h]of Object.entries(f.files))assert.equal(await hash(run+'/source/'+p),h,p);
for(const[p,h]of Object.entries(c.artifacts))assert.equal(await hash(run+'/'+p),h,p);
for(const[p,h]of Object.entries(f.generatedCompilerCopies))assert.equal(await hash(run+'/'+p),h,p);
for(const x of c.checks)assert.equal(await hash(run+'/'+x.name+'.log'),x.logSHA256,x.name);
assert.equal(r.arms,2160);assert.equal(r.modelArms,1920);assert.equal(r.controlsBitwiseEqual,240);assert.equal(r.allArmsBitwiseEqual,2160);assert.equal(r.modelArmsBitwiseEqual,1920);assert.equal(r.populationsIdentical,40);
assert.equal(a.transitiveIndependentIssuedPackets,4608000);assert.equal(a.transitiveScalarIssuedComparisons,9216000);assert(a.priorIndependentReplayVerified&&!a.directIndependentFullReplayThisStudy);
assert.equal(a.dataSHA256,await hash(run+'/diagnostic.jsonl'));assert.equal(a.priorDataSHA256,await hash('research/dynvariance-v71-diagnostic/diagnostic.jsonl'));assert.equal(a.priorCompletedSHA256,await hash('research/dynvariance-v71-diagnostic/completed.json'));
assert.equal(cost.dataSHA256,await hash(run+'/diagnostic.jsonl'));assert.equal(cost.sourceSHA256,await hash('research/dynvarcache-v72-cost-audit.mjs'));
assert(!r.equalTotalCostSuperiorityEstablished&&!r.loadedServingEstablished);
for(const s of Object.values(r.summaries))assert(!s.gates.gain01&&!s.gates.noAdaptiveHarm01&&!s.gates.recovery&&s.gates.completeLoop400);
for(const stage of ['initial','optimized']){
 const d='research/dynvarcache-v72-'+stage,pf=json(d+'/freeze.json'),pc=json(d+'/completed.json');assert(pc.allJobsTerminal&&pc.allChecksPass&&pc.sourceUnchanged&&pc.checks.length===3);
 for(const[p,h]of Object.entries(pf.files))assert.equal(await hash(d+'/source/'+p.replace(/^\//,'')),h,p);
 for(const x of pc.checks){assert.equal(x.exitCode,0);assert.equal(await hash(d+'/'+x.name+'.log'),x.logSHA256)}
}
const mean='research/mean-joint-v73-preflight',mf=json(mean+'/freeze.json'),mc=json(mean+'/completed.json'),mr=json(mean+'/results.json');
assert.equal(mc.exitCode,0);assert(mc.allJobsTerminal&&mc.sourceUnchanged);assert(!fs.existsSync(mean+'/failure.json'));assert.equal(await hash(mean+'/results.json'),mc.resultsSHA256);assert.equal(await hash(mean+'/source.mjs'),mf.sourceSHA256);assert.equal(await hash(mf.source),mf.sourceSHA256);assert.equal(await hash(mf.meanFamilySource),mf.meanFamilySourceSHA256);
assert.equal(mr.checks,217333);assert.equal(mr.histories,192);assert(mr.maxDefect<2e-11&&mr.meanFamily===27&&!mr.streamQualityRescueEstablished&&!mr.performanceEstablished);
const layout='research/mean-layout-v73-initial',lf=json(layout+'/freeze.json'),lc=json(layout+'/completed.json'),la=json(layout+'/allocation.json');
assert(lc.allJobsTerminal&&lc.allChecksPass&&lc.sourceUnchanged&&lc.checks.length===3&&!lc.completeLearnerImplemented);
for(const[p,h]of Object.entries(lf.files))assert.equal(await hash(layout+'/source/'+p.replace(/^\//,'')),h,p);
for(const x of lc.checks){assert.equal(x.exitCode,0);assert.equal(await hash(layout+'/'+x.name+'.log'),x.logSHA256)}
assert.equal(await hash(layout+'/allocation.json'),lc.allocationSHA256);assert(Object.values(la.AllocatedBytes).every(x=>x<=8388608));
fs.mkdirSync(root,{mode:0o700});const copies={},generatedCompilerCopies={};
async function copy(src){
 if(Object.hasOwn(copies,src)){assert.equal(await hash(src),copies[src]);return}
 const out=root+'/saved/'+src;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(src,out,fs.constants.COPYFILE_EXCL|fs.constants.COPYFILE_FICLONE);fs.chmodSync(out,0o600);copies[src]=await hash(src);assert.equal(await hash(out),copies[src]);
}
async function tree(dir){for(const n of fs.readdirSync(dir)){const p=dir+'/'+n;if(fs.statSync(p).isDirectory())await tree(p);else await copy(p)}}
for(const p of Object.keys(f.files))await copy(p);
for(const p of fs.readdirSync(run).filter(p=>/\.(json|jsonl|log)$/.test(p)))await copy(run+'/'+p);
for(const[p,h]of Object.entries(f.generatedCompilerCopies)){const out='generated/'+h+'-test-main.go';fs.mkdirSync(root+'/generated',{recursive:true,mode:0o700});fs.copyFileSync(run+'/'+p,root+'/'+out,fs.constants.COPYFILE_EXCL);generatedCompilerCopies[out]=h;assert.equal(await hash(root+'/'+out),h)}
for(const stage of ['initial','optimized'])await tree('research/dynvarcache-v72-'+stage);
await tree(mean);await tree(layout);await tree('internal/researchmeanlayout');
for(const n of fs.readdirSync('research').filter(n=>/^(dynvarcache-v72|mean-joint-v73|mean-layout-v73).*\.(mjs|json|md)$/.test(n)))await copy('research/'+n);
for(const p of ['docs/experiments/mmm-dynvarcache-v72-results.md','docs/experiments/mmm-mean-joint-v73-preflight-results.md','research-direction.md'])await copy(p);
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:parentPath,sha256:await hash(parentPath),copiesVerified:Object.keys(parent.copies).length},copies,generatedCompilerCopies,trackedHashes:parent.trackedHashes,goal:'ACTIVE',goals:Array(7).fill('OPEN'),weeklyUsage:{usedPercent:usage,windowDurationMins:10080},allRequiredJobsTerminal:true,completedCommands:{path:run,completedSHA256:copies[run+'/completed.json']},study:{worlds:40,arms:r.arms,modelArms:r.modelArms,allArmsBitwiseEqual:r.allArmsBitwiseEqual,populationsIdentical:40,directIndependentFullReplayThisStudy:false,transitiveIndependentIssuedPackets:a.transitiveIndependentIssuedPackets,transitiveScalarIssuedComparisons:a.transitiveScalarIssuedComparisons,totals:r.totals,gates:Object.fromEntries(Object.entries(r.summaries).map(([m,s])=>[m,s.gates])),allocation:r.allocation,equalTotalCostSuperiorityEstablished:false,loadedServingEstablished:false},meanPreflight:{checks:mr.checks,maxDefect:mr.maxDefect,histories:mr.histories,nonlocalWitness:mr.nonlocalWitness,meanFamily:27,hyperstates:243,allocation:la,completeLearnerImplemented:false,scientificQualityRescueEstablished:false},previousTurn:'NO_RESEARCH_PROGRESS(social clarification); preceding research goal turn PROGRESS(V71)',thisTurn:'PROGRESS(equivalent cached model, fresh independent/journal/race/future audits, complete2160-arm collection and full exact transitive replay link, cost readback, jointmean finite maths and measured complete proposed storage layout)',archiveIsChainedNotStandalone:true,productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,privateOrSealedLabelsOpened:false,populationConfirmationDispatched:false,allSevenEmpiricalGoalsValidated:false,scientificAdoption:false,residualLimitations:['all16 model scientific gates still fail despite every runtime gate now passing','full-study independent proof is transitive, not a fresh full replay','static mean/family preflight is not changepoint-adaptive learned hyperparameters','mean Go package initializes only; fitted updates/replay/queries still unimplemented','equal requests not equal TOTAL cost','serial costs and initial allocation not loaded serving, durability or freshness','valid useful splits and untouched agent outcomes still required'],next:'Integrate the FULL27-mean joint model into bounded posterior/replay/value machinery, audit independently and screen complete consumed cohort under unchanged gates; keep all seven WHOLE goals active.'};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(copies))assert.equal(await hash(root+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await hash(root+'/manifest.json'),copies:Object.keys(copies).length,allRequiredJobsTerminal:true,goal:'ACTIVE',wholeGoals:'seven OPEN',weeklyUsage:usage},null,2));
