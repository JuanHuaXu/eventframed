// Chained private checkpoint, not a release, adoption or completion claim.
import fs from 'node:fs';import path from 'node:path';import crypto from 'node:crypto';import assert from 'node:assert/strict';import {execFileSync} from 'node:child_process';
const root='research/checkpoint-2026-10-05-mean-joint-v74-anchor-v75',parentPath='research/checkpoint-2026-10-04-dynvarcache-v72-mean-v73/manifest.json',usage=Number(process.argv[2]);
assert(Number.isFinite(usage)&&usage>=0&&usage<=100);assert(!fs.existsSync(root));
async function hash(p){const h=crypto.createHash('sha256');for await(const b of fs.createReadStream(p))h.update(b);return h.digest('hex')}
const json=p=>JSON.parse(fs.readFileSync(p));
assert.equal(await hash(parentPath),'d4045ac945e8044155f54fb0871d8b2fffb30fa8a4806d431beaf028ea22cc54');const parent=json(parentPath);
for(const[p,h]of Object.entries(parent.copies))assert.equal(await hash(path.dirname(parentPath)+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.generatedCompilerCopies))assert.equal(await hash(path.dirname(parentPath)+'/'+p),h,p);
const tracked=execFileSync('git',['diff','--name-only'],{encoding:'utf8'}).trim().split('\n').filter(Boolean).sort();assert.deepEqual(tracked,Object.keys(parent.trackedHashes).sort());
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
const run='research/mean-joint-v74-diagnostic',f=json(run+'/freeze.json'),c=json(run+'/completed.json'),r=json(run+'/readback.json'),cost=json(run+'/cost-audit.json');
assert(c.allJobsTerminal&&c.sourceUnchanged&&c.checks.length===6&&c.checks.every(x=>x.exitCode===0));assert(!fs.existsSync(run+'/failure.json'));
for(const[p,h]of Object.entries({...f.files,...f.protectedFiles,...f.data}))assert.equal(await hash(p),h,p);
for(const[p,h]of Object.entries(f.files))assert.equal(await hash(run+'/source/'+p),h,p);
for(const[p,h]of Object.entries(c.artifacts))assert.equal(await hash(run+'/'+p),h,p);
for(const[p,h]of Object.entries(f.generatedCompilerCopies))assert.equal(await hash(run+'/'+p),h,p);
for(const x of c.checks)assert.equal(await hash(run+'/'+x.name+'.log'),x.logSHA256,x.name);
assert.equal(r.arms,3600);assert.equal(r.newModelArms,1440);assert.equal(r.controlsBitwiseEqual,2160);assert.equal(r.populationsIdentical,40);assert.equal(r.freshDetailedFixtureArms,3);assert(!r.fullNewArmIndependentReplay);
assert.equal(cost.dataSHA256,await hash(run+'/diagnostic.jsonl'));assert.equal(cost.sourceSHA256,await hash('research/mean-joint-v74-cost-audit.mjs'));
assert(!r.equalTotalCostSuperiorityEstablished&&!r.loadedServingEstablished);
const newModes=Object.keys(r.totals).filter(m=>m.startsWith('mean'));assert.equal(newModes.length,12);
for(const m of newModes){const s=r.summaries[m];assert(!s.gates.gain01&&!s.gates.noAdaptiveHarm01&&!s.gates.recovery&&!s.gates.completeLoop400);assert.equal(s.over400MS,120)}
const doc=fs.readFileSync('docs/experiments/mmm-mean-joint-v74-results.md','utf8');let rows=0;
for(const line of doc.split('\n')){
 const x=line.match(/^\| (mean[a-z_]+) \| ([.\d]+) \| (\d+) \| (-[.\d]+) \| ([.\d]+) \| ([.\d]+) \|$/);if(!x)continue;
 const [,m,b,h,rec,avg,worst]=x,t=r.totals[m],s=r.summaries[m];assert(t&&s);assert(Math.abs(+b-t.risk)<5.1e-10);assert.equal(+h,s.harmOver01);assert(Math.abs(+rec-s.recoveryGainAdaptive)<5.1e-5);assert(Math.abs(+avg-t.totalMS/120)<.00051);assert(Math.abs(+worst-t.maxMS)<.00051);rows++;
}assert.equal(rows,12,'all documented arm rows independently match readback');
for(const stage of ['initial','long']){
 const d='research/mean-joint-v74-'+stage,pf=json(d+'/freeze.json'),pc=json(d+'/completed.json');assert(pc.allJobsTerminal&&pc.allChecksPass&&pc.sourceUnchanged&&pc.checks.length===3);
 for(const[p,h]of Object.entries(pf.files))assert.equal(await hash(d+'/source/'+p.replace(/^\//,'')),h,p);
 for(const x of pc.checks){assert.equal(x.exitCode,0);assert.equal(await hash(d+'/'+x.name+'.log'),x.logSHA256)}
}
const v75='research/mean-anchor-v75-initial',af=json(v75+'/freeze.json'),ac=json(v75+'/completed.json'),allocation=json(v75+'/allocation.json');
assert(ac.allJobsTerminal&&ac.allChecksPass&&ac.sourceUnchanged&&ac.checks.length===4&&!ac.fullStudyEquivalenceEstablished&&!ac.scientificQualityRescueEstablished);
assert.equal(af.v74CompletedSHA256,await hash(run+'/completed.json'));
for(const[p,h]of Object.entries(af.files)){assert.equal(await hash(p),h,p);assert.equal(await hash(v75+'/source/'+p.replace(/^\//,'')),h,p)}
for(const x of ac.checks){assert.equal(x.exitCode,0);assert.equal(await hash(v75+'/'+x.name+'.log'),x.logSHA256)}
assert(Object.values(allocation.AllocatedBytes).every(x=>x<=allocation.CapBytes));
const matched='research/mean-anchor-v75-matched',mp=json(matched+'/freeze.json'),mc=json(matched+'/completed.json');
assert(mc.allJobsTerminal&&mc.allChecksPass&&mc.sourceUnchanged&&mc.checks.length===4&&!mc.wholeCohortEquivalenceEstablished&&!mc.loadedServingEstablished);
assert.equal(mp.v74CompletedSHA256,await hash(run+'/completed.json'));assert.equal(mp.v75CompletedSHA256,await hash(v75+'/completed.json'));
for(const[p,h]of Object.entries(mp.files)){assert.equal(await hash(p),h,p);assert.equal(await hash(matched+'/source/'+p.replace(/^\//,'')),h,p)}
for(const x of mc.checks){assert.equal(x.exitCode,0);assert.equal(await hash(matched+'/'+x.name+'.log'),x.logSHA256)}
fs.mkdirSync(root,{mode:0o700});const copies={},generatedCompilerCopies={};
async function copy(src){if(Object.hasOwn(copies,src)){assert.equal(await hash(src),copies[src]);return}const out=root+'/saved/'+src;fs.mkdirSync(path.dirname(out),{recursive:true,mode:0o700});fs.copyFileSync(src,out,fs.constants.COPYFILE_EXCL|fs.constants.COPYFILE_FICLONE);fs.chmodSync(out,0o600);copies[src]=await hash(src);assert.equal(await hash(out),copies[src]);}
async function tree(dir){for(const n of fs.readdirSync(dir)){const p=dir+'/'+n;if(fs.statSync(p).isDirectory())await tree(p);else await copy(p)}}
for(const p of Object.keys(f.files))await copy(p);
for(const p of fs.readdirSync(run).filter(p=>/\.(json|jsonl|log)$/.test(p)))await copy(run+'/'+p);
for(const[p,h]of Object.entries(f.generatedCompilerCopies)){const out='generated/'+h+'-test-main.go';fs.mkdirSync(root+'/generated',{recursive:true,mode:0o700});fs.copyFileSync(run+'/'+p,root+'/'+out,fs.constants.COPYFILE_EXCL);generatedCompilerCopies[out]=h;assert.equal(await hash(root+'/'+out),h)}
for(const stage of ['initial','long'])await tree('research/mean-joint-v74-'+stage);
await tree(v75);await tree('internal/researchmeananchor');await tree('internal/researchmeananchorcheck');
await tree(matched);await tree('internal/researchmeanperfcheck');
for(const n of fs.readdirSync('research').filter(n=>/^(mean-joint-v74|mean-anchor-v75).*\.(mjs|json|md)$/.test(n)))await copy('research/'+n);
for(const p of ['docs/experiments/mmm-mean-joint-v74-protocol.md','docs/experiments/mmm-mean-joint-v74-results.md','docs/experiments/mmm-mean-anchor-v75-protocol.md','docs/experiments/mmm-mean-anchor-v75-results.md','research-direction.md'])await copy(p);
const manifest={time:new Date().toISOString(),previousCheckpoint:{path:parentPath,sha256:await hash(parentPath),copiesVerified:Object.keys(parent.copies).length},copies,generatedCompilerCopies,trackedHashes:parent.trackedHashes,goal:'ACTIVE',goals:Array(7).fill('OPEN'),weeklyUsage:{usedPercent:usage,windowDurationMins:10080},allRequiredJobsTerminal:true,completedCommands:{path:run,completedSHA256:copies[run+'/completed.json']},study:{worlds:40,arms:r.arms,newModelArms:r.newModelArms,controlsBitwiseEqual:r.controlsBitwiseEqual,populationsIdentical:40,fullNewArmIndependentReplay:false,freshDetailedFixtureArms:3,totals:r.totals,gates:Object.fromEntries(Object.entries(r.summaries).map(([m,s])=>[m,s.gates])),allocation:r.allocation,equalTotalCostSuperiorityEstablished:false,loadedServingEstablished:false},anchorPreflight:{path:v75,completedSHA256:copies[v75+'/completed.json'],allocation,fullStudyEquivalenceEstablished:false,scientificQualityRescueEstablished:false},previousTurn:'NO_RESEARCH_PROGRESS(social clarification); prior research work implemented and started frozen V74 study',thisTurn:'PROGRESS(complete joint-mean controlled collection/cost readback and separate equivalent anchored-prefix integration with independent/journal/race/performance preflight)',archiveIsChainedNotStandalone:true,productionPrivateWhitepaperUntouched:true,noCommitPushInstallDeployment:true,privateOrSealedLabelsOpened:false,populationConfirmationDispatched:false,allSevenEmpiricalGoalsValidated:false,scientificAdoption:false,residualLimitations:['joint-mean screen uses consumed n1 per cell, not confirmation','full new-arm independent replay not performed; three detailed fixtures plus broad joint/full-journal references','V75 equivalence verified only at preflight scope, not all full-screen trajectories','static member-index mean maps are not adaptive ontology-derived agent features','equal requests not equal TOTAL cost','serial timings and constructor allocation not loaded serving, durability or freshness','useful valid splits and untouched labeled agent outcomes remain required'],next:'Complete full anchored-model screen and independent equivalence under unchanged gates, then choose distinct quality/observation leads from verified results; all seven WHOLE goals stay active.'};
fs.writeFileSync(root+'/manifest.json',JSON.stringify(manifest,null,2)+'\n',{flag:'wx',mode:0o600});
for(const[p,h]of Object.entries(copies))assert.equal(await hash(root+'/saved/'+p),h,p);
for(const[p,h]of Object.entries(parent.trackedHashes))assert.equal(await hash(p),h,p);
console.log(JSON.stringify({path:root+'/manifest.json',sha256:await hash(root+'/manifest.json'),copies:Object.keys(copies).length,allRequiredJobsTerminal:true,goal:'ACTIVE',wholeGoals:'seven OPEN',weeklyUsage:usage},null,2));
